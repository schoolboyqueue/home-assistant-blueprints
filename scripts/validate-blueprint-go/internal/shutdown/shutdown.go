// Package shutdown provides graceful shutdown coordination for CLI tools.
// It handles OS signals (SIGINT, SIGTERM) and coordinates cleanup of resources
// with configurable timeouts and partial result reporting.
package shutdown

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// DefaultGracePeriod is the default time allowed for cleanup operations.
const DefaultGracePeriod = 5 * time.Second

// Coordinator manages graceful shutdown for CLI applications.
// It handles signal interception, context cancellation, and cleanup coordination.
type Coordinator struct {
	mu sync.RWMutex

	// ctx is the base context that gets canceled on shutdown
	ctx    context.Context
	cancel context.CancelFunc

	// gracePeriod is how long to wait for cleanup before force exit
	gracePeriod time.Duration

	// shutdownOnce ensures shutdown only happens once
	shutdownOnce sync.Once

	// shutdownChan signals that shutdown has started
	shutdownChan chan struct{}

	// shutdownReason records why shutdown was triggered
	shutdownReason string

	// onShutdown is called when shutdown is initiated (for logging/reporting)
	onShutdown func(reason string)
}

// Option configures a Coordinator.
type Option func(*Coordinator)

// WithGracePeriod sets the time allowed for cleanup before force exit.
func WithGracePeriod(d time.Duration) Option {
	return func(c *Coordinator) {
		c.gracePeriod = d
	}
}

// WithOnShutdown sets a callback for when shutdown is initiated.
func WithOnShutdown(fn func(reason string)) Option {
	return func(c *Coordinator) {
		c.onShutdown = fn
	}
}

// New creates a new shutdown Coordinator.
// The returned context is canceled when shutdown is triggered.
func New(opts ...Option) (*Coordinator, context.Context) {
	ctx, cancel := context.WithCancel(context.Background())

	c := &Coordinator{
		ctx:          ctx,
		cancel:       cancel,
		gracePeriod:  DefaultGracePeriod,
		shutdownChan: make(chan struct{}),
	}

	for _, opt := range opts {
		opt(c)
	}

	return c, ctx
}

// HandleSignals starts listening for SIGINT and SIGTERM signals.
// When a signal is received, it triggers graceful shutdown.
// A second signal forces immediate exit.
func (c *Coordinator) HandleSignals() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		c.Shutdown(fmt.Sprintf("received signal %v", sig))

		// Wait for another signal for force exit
		<-sigChan
		os.Exit(1)
	}()
}

// Shutdown initiates graceful shutdown with the given reason.
// It cancels the context to signal all operations to stop.
// This is safe to call multiple times; only the first call has effect.
func (c *Coordinator) Shutdown(reason string) {
	c.shutdownOnce.Do(func() {
		c.mu.Lock()
		c.shutdownReason = reason
		c.mu.Unlock()

		close(c.shutdownChan)

		if c.onShutdown != nil {
			c.onShutdown(reason)
		}

		// Cancel the main context to signal all operations to stop
		c.cancel()
	})
}

// ShutdownChan returns a channel that's closed when shutdown begins.
// This is useful for select statements in long-running operations.
func (c *Coordinator) ShutdownChan() <-chan struct{} {
	return c.shutdownChan
}

// IsShuttingDown returns true if shutdown has been initiated.
func (c *Coordinator) IsShuttingDown() bool {
	select {
	case <-c.shutdownChan:
		return true
	default:
		return false
	}
}

// Context returns the coordinator's context.
func (c *Coordinator) Context() context.Context {
	return c.ctx
}

// ShutdownReason returns the reason for shutdown, or empty string if not shut down.
func (c *Coordinator) ShutdownReason() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.shutdownReason
}

// PartialResult represents a partial result that can be reported on interruption.
type PartialResult struct {
	mu        sync.Mutex
	completed int
	total     int
	passed    int
	failed    int
	errors    []string
}

// NewPartialResult creates a new partial result tracker.
func NewPartialResult(total int) *PartialResult {
	return &PartialResult{total: total}
}

// RecordPass records a successful validation.
func (p *PartialResult) RecordPass(_ string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.completed++
	p.passed++
}

// RecordFail records a failed validation.
func (p *PartialResult) RecordFail(name, reason string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.completed++
	p.failed++
	p.errors = append(p.errors, fmt.Sprintf("%s: %s", name, reason))
}

// Summary returns a summary of the partial results.
func (p *PartialResult) Summary() (completed, total, passed, failed int, errors []string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.completed, p.total, p.passed, p.failed, p.errors
}

// IsComplete returns true if all operations have been processed.
func (p *PartialResult) IsComplete() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.completed >= p.total
}
