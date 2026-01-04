// Package shutdown provides graceful shutdown coordination for CLI tools.
// This package re-exports the shared shutdown package for ha-ws-client-go.
package shutdown

import (
	"context"

	"github.com/home-assistant-blueprints/shared/shutdown"
)

// Re-export types from shared shutdown package
type (
	// Coordinator manages graceful shutdown for CLI applications.
	Coordinator = shutdown.Coordinator
	// CleanupFunc is a function that performs cleanup during shutdown.
	CleanupFunc = shutdown.CleanupFunc
	// Option configures a Coordinator.
	Option = shutdown.Option
	// PartialResult represents a partial result that can be reported on interruption.
	PartialResult = shutdown.PartialResult
)

// Re-export constants
const DefaultGracePeriod = shutdown.DefaultGracePeriod

// Re-export functions
var (
	New                  = shutdown.New
	WithGracePeriod      = shutdown.WithGracePeriod
	WithOnShutdown       = shutdown.WithOnShutdown
	WithOnCleanupTimeout = shutdown.WithOnCleanupTimeout
	WrapContext          = shutdown.WrapContext
	NewPartialResult     = shutdown.NewPartialResult
)

// WrapContextCompat is an alias for WrapContext for backwards compatibility.
func WrapContextCompat(parent context.Context, coord *Coordinator) (context.Context, context.CancelFunc) {
	return shutdown.WrapContext(parent, coord)
}
