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

	// cleanupFuncs holds registered cleanup functions
	cleanupFuncs []CleanupFunc

	// shutdownOnce ensures shutdown only happens once
	shutdownOnce sync.Once

	// shutdownChan signals that shutdown has started
	shutdownChan chan struct{}

	// shutdownReason records why shutdown was triggered
	shutdownReason string

	// onShutdown is called when shutdown is initiated (for logging/reporting)
	onShutdown func(reason string)

	// onCleanupTimeout is called if cleanup times out
	onCleanupTimeout func()
}

// CleanupFunc is a function that performs cleanup during shutdown.
// It receives a context that may be canceled if cleanup is taking too long.
// The name is used for logging/debugging purposes.
type CleanupFunc struct {
	Name string
	Func func(ctx context.Context) error
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

// WithOnCleanupTimeout sets a callback for when cleanup times out.
func WithOnCleanupTimeout(fn func()) Option {
	return func(c *Coordinator) {
		c.onCleanupTimeout = fn
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

// RegisterCleanup adds a cleanup function to be called during shutdown.
// Cleanup functions are called in LIFO order (last registered, first called).
// This is safe to call from multiple goroutines.
func (c *Coordinator) RegisterCleanup(name string, fn func(ctx context.Context) error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cleanupFuncs = append(c.cleanupFuncs, CleanupFunc{Name: name, Func: fn})
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
// It cancels the context, runs cleanup functions, and respects the grace period.
// This is safe to call multiple times; only the first call has effect.
func (c *Coordinator) Shutdown(reason string) {
	c.shutdownOnce.Do(func() {
		c.mu.Lock()
		c.shutdownReason = reason
		cleanups := make([]CleanupFunc, len(c.cleanupFuncs))
		copy(cleanups, c.cleanupFuncs)
		c.mu.Unlock()

		close(c.shutdownChan)

		if c.onShutdown != nil {
			c.onShutdown(reason)
		}

		// Cancel the main context to signal all operations to stop
		c.cancel()

		// Run cleanup with a timeout
		c.runCleanups(cleanups)
	})
}

// runCleanups executes all cleanup functions with a timeout.
func (c *Coordinator) runCleanups(cleanups []CleanupFunc) {
	if len(cleanups) == 0 {
		return
	}

	cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), c.gracePeriod)
	defer cleanupCancel()

	done := make(chan struct{})

	go func() {
		defer close(done)
		// Run in LIFO order
		for i := len(cleanups) - 1; i >= 0; i-- {
			cleanup := cleanups[i]
			if err := cleanup.Func(cleanupCtx); err != nil {
				// Log cleanup errors but continue with other cleanups
				fmt.Fprintf(os.Stderr, "cleanup %q error: %v\n", cleanup.Name, err)
			}
		}
	}()

	select {
	case <-done:
		// All cleanups completed
	case <-cleanupCtx.Done():
		if c.onCleanupTimeout != nil {
			c.onCleanupTimeout()
		}
	}
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

// Wait blocks until shutdown is complete or the grace period expires.
// This is typically called at the end of main() to ensure cleanup completes.
func (c *Coordinator) Wait() {
	<-c.shutdownChan
	// Give cleanups time to complete
	time.Sleep(100 * time.Millisecond)
}

// ShutdownReason returns the reason for shutdown, or empty string if not shut down.
func (c *Coordinator) ShutdownReason() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.shutdownReason
}

// WrapContext creates a child context that is canceled when either
// the parent context or the coordinator's context is canceled.
// This is useful for operations that should respect both a timeout
// and graceful shutdown.
func WrapContext(parent context.Context, coord *Coordinator) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(parent)

	go func() {
		select {
		case <-coord.shutdownChan:
			cancel()
		case <-parent.Done():
			// Parent already canceled
		case <-ctx.Done():
			// Context already canceled
		}
	}()

	return ctx, cancel
}

// PartialResult represents a partial result that can be reported on interruption.
// This is useful for batch operations that should report progress on interrupt.
type PartialResult struct {
	mu         sync.Mutex
	completed  int
	total      int
	passed     int
	failed     int
	lastResult any
	errors     []error
	messages   []string
}

// NewPartialResult creates a new partial result tracker.
func NewPartialResult(total int) *PartialResult {
	return &PartialResult{total: total}
}

// RecordSuccess records a successful operation.
func (p *PartialResult) RecordSuccess(result any) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.completed++
	p.passed++
	p.lastResult = result
}

// RecordError records a failed operation with an error.
func (p *PartialResult) RecordError(err error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.completed++
	p.failed++
	p.errors = append(p.errors, err)
}

// RecordPass records a successful validation (named variant).
func (p *PartialResult) RecordPass(name string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.completed++
	p.passed++
}

// RecordFail records a failed validation with a reason message.
func (p *PartialResult) RecordFail(name, reason string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.completed++
	p.failed++
	p.messages = append(p.messages, fmt.Sprintf("%s: %s", name, reason))
}

// Summary returns a summary of the partial results.
func (p *PartialResult) Summary() (completed, total int, errors []error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.completed, p.total, p.errors
}

// SummaryWithCounts returns an extended summary including pass/fail counts.
func (p *PartialResult) SummaryWithCounts() (completed, total, passed, failed int, messages []string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.completed, p.total, p.passed, p.failed, p.messages
}

// IsComplete returns true if all operations have been processed.
func (p *PartialResult) IsComplete() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.completed >= p.total
}

// LastResult returns the last recorded successful result.
func (p *PartialResult) LastResult() any {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.lastResult
}
