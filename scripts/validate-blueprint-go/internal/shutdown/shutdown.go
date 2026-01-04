// Package shutdown provides graceful shutdown coordination for CLI tools.
// This package re-exports the shared shutdown package for validate-blueprint-go.
package shutdown

import (
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
	New              = shutdown.New
	WithGracePeriod  = shutdown.WithGracePeriod
	WithOnShutdown   = shutdown.WithOnShutdown
	NewPartialResult = shutdown.NewPartialResult
)
