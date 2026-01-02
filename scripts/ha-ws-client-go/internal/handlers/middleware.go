// Package handlers provides command handlers for the CLI.
package handlers

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/home-assistant-blueprints/ha-ws-client-go/internal/types"
)

// Handler is the standard handler function signature.
type Handler func(ctx *Context) error

// Middleware is a function that wraps a handler.
type Middleware func(Handler) Handler

// HandlerConfig contains configuration extracted by middleware.
// Handlers can access this via ctx.Config after middleware runs.
type HandlerConfig struct {
	// Required arguments extracted by RequireArgs middleware
	Args []string

	// Time range extracted by WithTimeRange middleware
	TimeRange *types.TimeRange

	// Optional numeric argument (e.g., hours, seconds)
	OptionalInt int

	// Pattern compiled by WithPattern middleware
	Pattern *regexp.Regexp

	// Automation ID cleaned by WithAutomationID middleware
	AutomationID string
}

// WithConfig adds a HandlerConfig to the context if not present.
func WithConfig(ctx *Context) *HandlerConfig {
	if ctx.Config == nil {
		ctx.Config = &HandlerConfig{}
	}
	return ctx.Config
}

// RequireArgs creates middleware that validates required arguments.
// It extracts arguments at the specified indices and stores them in ctx.Config.Args.
//
// Example:
//
//	handler := RequireArgs(1, "Usage: state <entity_id>")(func(ctx *Context) error {
//	    entityID := ctx.Config.Args[0]
//	    // ...
//	})
func RequireArgs(usage string, indices ...int) Middleware {
	return func(next Handler) Handler {
		return func(ctx *Context) error {
			config := WithConfig(ctx)
			config.Args = make([]string, len(indices))

			for i, idx := range indices {
				if idx >= len(ctx.Args) {
					return fmt.Errorf("missing argument: %s", usage)
				}
				config.Args[i] = ctx.Args[idx]
			}

			return next(ctx)
		}
	}
}

// RequireArg1 is a convenience wrapper for requiring a single argument at index 1.
func RequireArg1(usage string) Middleware {
	return RequireArgs(usage, 1)
}

// RequireArg2 is a convenience wrapper for requiring arguments at indices 1 and 2.
func RequireArg2(usage string) Middleware {
	return RequireArgs(usage, 1, 2)
}

// WithTimeRange creates middleware that parses time range from context.
// It uses ctx.FromTime/ToTime if set, otherwise calculates from defaultHours.
// The optional hours argument can be provided at argIndex (0 = disabled).
//
// Example:
//
//	handler := WithTimeRange(24, 2)(func(ctx *Context) error {
//	    timeRange := ctx.Config.TimeRange
//	    // ...
//	})
func WithTimeRange(defaultHours, hoursArgIndex int) Middleware {
	return func(next Handler) Handler {
		return func(ctx *Context) error {
			config := WithConfig(ctx)

			endTime := time.Now()
			if ctx.ToTime != nil {
				endTime = *ctx.ToTime
			}

			var startTime time.Time
			if ctx.FromTime != nil {
				startTime = *ctx.FromTime
			} else {
				hours := defaultHours
				if hoursArgIndex > 0 && hoursArgIndex < len(ctx.Args) {
					if h, err := strconv.Atoi(ctx.Args[hoursArgIndex]); err == nil {
						hours = h
					}
				}
				startTime = endTime.Add(-time.Duration(hours) * time.Hour)
			}

			config.TimeRange = &types.TimeRange{StartTime: startTime, EndTime: endTime}
			return next(ctx)
		}
	}
}

// WithOptionalInt creates middleware that parses an optional integer argument.
// It stores the parsed value (or default) in ctx.Config.OptionalInt.
//
// Example:
//
//	handler := WithOptionalInt(60, 2)(func(ctx *Context) error {
//	    seconds := ctx.Config.OptionalInt
//	    // ...
//	})
func WithOptionalInt(defaultValue, argIndex int) Middleware {
	return func(next Handler) Handler {
		return func(ctx *Context) error {
			config := WithConfig(ctx)
			config.OptionalInt = defaultValue

			if argIndex > 0 && argIndex < len(ctx.Args) {
				if parsed, err := strconv.Atoi(ctx.Args[argIndex]); err == nil {
					config.OptionalInt = parsed
				}
			}

			return next(ctx)
		}
	}
}

// WithPattern creates middleware that compiles a glob pattern to regex.
// If no pattern argument is provided at argIndex, the pattern is set to nil (match all).
// The pattern supports * as wildcard and is case-insensitive.
//
// Example:
//
//	handler := WithPattern(1)(func(ctx *Context) error {
//	    if ctx.Config.Pattern != nil {
//	        matched := ctx.Config.Pattern.MatchString(entityID)
//	    }
//	    // ...
//	})
func WithPattern(argIndex int) Middleware {
	return func(next Handler) Handler {
		return func(ctx *Context) error {
			config := WithConfig(ctx)

			if argIndex > 0 && argIndex < len(ctx.Args) {
				pattern := ctx.Args[argIndex]
				regexPattern := regexp.QuoteMeta(pattern)
				regexPattern = regexp.MustCompile(`\\\*`).ReplaceAllString(regexPattern, ".*")
				re, err := regexp.Compile("(?i)" + regexPattern)
				if err != nil {
					return fmt.Errorf("invalid pattern: %w", err)
				}
				config.Pattern = re
			}

			return next(ctx)
		}
	}
}

// WithRequiredPattern creates middleware that requires a pattern argument.
// Unlike WithPattern, this returns an error if the pattern is not provided.
//
// Example:
//
//	handler := WithRequiredPattern(1, "Usage: states-filter <pattern>")(func(ctx *Context) error {
//	    matched := ctx.Config.Pattern.MatchString(entityID)
//	    // ...
//	})
func WithRequiredPattern(argIndex int, usage string) Middleware {
	return func(next Handler) Handler {
		return func(ctx *Context) error {
			if argIndex >= len(ctx.Args) {
				return fmt.Errorf("missing argument: %s", usage)
			}

			config := WithConfig(ctx)
			pattern := ctx.Args[argIndex]
			regexPattern := regexp.QuoteMeta(pattern)
			regexPattern = regexp.MustCompile(`\\\*`).ReplaceAllString(regexPattern, ".*")
			re, err := regexp.Compile("(?i)" + regexPattern)
			if err != nil {
				return fmt.Errorf("invalid pattern: %w", err)
			}
			config.Pattern = re

			return next(ctx)
		}
	}
}

// WithAutomationID creates middleware that cleans up an automation ID.
// It extracts the automation ID from the specified argIndex, removes the
// "automation." prefix if present, and stores it in ctx.Config.AutomationID.
//
// Example:
//
//	handler := WithAutomationID(1, "Usage: traces <automation_id>")(func(ctx *Context) error {
//	    id := ctx.Config.AutomationID
//	    // ...
//	})
func WithAutomationID(argIndex int, usage string) Middleware {
	return func(next Handler) Handler {
		return func(ctx *Context) error {
			if argIndex >= len(ctx.Args) {
				return fmt.Errorf("missing argument: %s", usage)
			}

			config := WithConfig(ctx)
			automationID := ctx.Args[argIndex]
			config.AutomationID = strings.TrimPrefix(automationID, "automation.")

			return next(ctx)
		}
	}
}

// WithOptionalAutomationID creates middleware that optionally extracts an automation ID.
// If the argument is not provided, AutomationID will be empty.
func WithOptionalAutomationID(argIndex int) Middleware {
	return func(next Handler) Handler {
		return func(ctx *Context) error {
			config := WithConfig(ctx)

			if argIndex > 0 && argIndex < len(ctx.Args) {
				config.AutomationID = strings.TrimPrefix(ctx.Args[argIndex], "automation.")
			}

			return next(ctx)
		}
	}
}

// EnsureAutomationPrefix ensures an entity ID has the "automation." prefix.
// This is a helper function, not middleware.
func EnsureAutomationPrefix(entityID string) string {
	if !strings.HasPrefix(entityID, "automation.") {
		return "automation." + entityID
	}
	return entityID
}

// Chain combines multiple middleware into a single middleware.
// Middleware is applied in order (first middleware wraps the outermost layer).
//
// Example:
//
//	handler := Chain(
//	    RequireArg1("Usage: history <entity_id> [hours]"),
//	    WithTimeRange(24, 2),
//	)(func(ctx *Context) error {
//	    entityID := ctx.Config.Args[0]
//	    timeRange := ctx.Config.TimeRange
//	    // ...
//	})
func Chain(middlewares ...Middleware) Middleware {
	return func(handler Handler) Handler {
		for i := len(middlewares) - 1; i >= 0; i-- {
			handler = middlewares[i](handler)
		}
		return handler
	}
}

// Apply applies middleware to a handler and returns the wrapped handler.
// This is a convenience function for use in command registration.
//
// Example:
//
//	var commandRegistry = map[string]func(*handlers.Context) error{
//	    "history": Apply(
//	        Chain(RequireArg1(...), WithTimeRange(...)),
//	        handleHistory,
//	    ),
//	}
func Apply(m Middleware, h Handler) Handler {
	return m(h)
}
