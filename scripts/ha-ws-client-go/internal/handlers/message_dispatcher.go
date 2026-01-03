// Package handlers provides command handlers for the CLI.
package handlers

import (
	"fmt"

	"github.com/home-assistant-blueprints/ha-ws-client-go/internal/client"
	"github.com/home-assistant-blueprints/ha-ws-client-go/internal/output"
	"github.com/home-assistant-blueprints/ha-ws-client-go/internal/types"
)

// MessageRequest defines the parameters for a WebSocket message request.
// This provides a unified way to define message requests with type-safe responses.
type MessageRequest[T any] struct {
	// Type is the WebSocket message type (e.g., "get_states", "trace/list").
	Type string

	// Data contains the message payload. Can be nil for messages without data.
	Data map[string]any
}

// NewRequest creates a new MessageRequest with the specified type and optional data.
func NewRequest[T any](msgType string, data map[string]any) MessageRequest[T] {
	return MessageRequest[T]{
		Type: msgType,
		Data: data,
	}
}

// Execute sends the message request and returns the typed response.
// This is the core dispatch method that handles error propagation.
func (r MessageRequest[T]) Execute(ctx *Context) (T, error) {
	return client.SendMessageTyped[T](ctx.Client, r.Type, r.Data)
}

// ExecuteRaw sends the message request and returns the raw HAMessage response.
// Use this when you need access to the full response, not just the typed result.
func (r MessageRequest[T]) ExecuteRaw(ctx *Context) (*types.HAMessage, error) {
	return ctx.Client.SendMessage(r.Type, r.Data)
}

// OutputConfig defines how to format and display message response data.
type OutputConfig struct {
	// Command is the command name for JSON output metadata.
	Command string

	// Summary provides an optional summary message for output.
	Summary string

	// Count provides an optional item count for output metadata.
	Count int
}

// OutputOption is a functional option for configuring OutputConfig.
type OutputOption func(*OutputConfig)

// WithCommand sets the command name for output metadata.
func WithOutputCommand(cmd string) OutputOption {
	return func(c *OutputConfig) {
		c.Command = cmd
	}
}

// WithOutputSummary sets the summary message for output.
func WithOutputSummary(summary string) OutputOption {
	return func(c *OutputConfig) {
		c.Summary = summary
	}
}

// WithOutputCount sets the item count for output metadata.
func WithOutputCount(count int) OutputOption {
	return func(c *OutputConfig) {
		c.Count = count
	}
}

// ExecuteAndOutput executes the request and outputs the result using standard output formatting.
// This combines the common pattern of execute + output.Data in a single call.
func (r MessageRequest[T]) ExecuteAndOutput(ctx *Context, opts ...OutputOption) error {
	result, err := r.Execute(ctx)
	if err != nil {
		return err
	}

	cfg := OutputConfig{}
	for _, opt := range opts {
		opt(&cfg)
	}

	// Build output options
	var outputOpts []output.Option
	if cfg.Command != "" {
		outputOpts = append(outputOpts, output.WithCommand(cfg.Command))
	}
	if cfg.Summary != "" {
		outputOpts = append(outputOpts, output.WithSummary(cfg.Summary))
	}
	if cfg.Count > 0 {
		outputOpts = append(outputOpts, output.WithCount(cfg.Count))
	}

	output.Data(result, outputOpts...)
	return nil
}

// MessageDispatcher provides a fluent API for executing message requests
// with optional transformations and output formatting.
type MessageDispatcher[T any] struct {
	request   MessageRequest[T]
	transform func(T) (T, error)
	outputFn  func(T) error
}

// Dispatch creates a new MessageDispatcher for the given request type.
// The type parameter T is the response type from the WebSocket message.
func Dispatch[T any](msgType string, data map[string]any) *MessageDispatcher[T] {
	return &MessageDispatcher[T]{
		request: MessageRequest[T]{
			Type: msgType,
			Data: data,
		},
		transform: func(t T) (T, error) { return t, nil },
	}
}

// Transform adds a transformation step to the dispatcher.
// This allows you to filter, modify, or convert the response data.
func (d *MessageDispatcher[T]) Transform(fn func(T) (T, error)) *MessageDispatcher[T] {
	d.transform = fn
	return d
}

// Output sets a custom output function for the result.
func (d *MessageDispatcher[T]) Output(fn func(T) error) *MessageDispatcher[T] {
	d.outputFn = fn
	return d
}

// Execute runs the full dispatch pipeline: request -> transform -> output.
func (d *MessageDispatcher[T]) Execute(ctx *Context) error {
	result, err := d.request.Execute(ctx)
	if err != nil {
		return err
	}

	transformed, err := d.transform(result)
	if err != nil {
		return err
	}

	if d.outputFn != nil {
		return d.outputFn(transformed)
	}

	// Default: output as JSON data
	output.Data(transformed)
	return nil
}

// Result executes only the request and transform, returning the result.
// This is useful when you need the data for further processing.
func (d *MessageDispatcher[T]) Result(ctx *Context) (T, error) {
	result, err := d.request.Execute(ctx)
	if err != nil {
		var zero T
		return zero, err
	}
	return d.transform(result)
}

// FetchStates is a convenience function to fetch all entity states.
// This wraps the common get_states call pattern.
func FetchStates[T any](ctx *Context) ([]T, error) {
	return client.SendMessageTyped[[]T](ctx.Client, "get_states", nil)
}

// FetchAndFindEntity fetches all states and finds a specific entity by ID.
// Returns the entity and nil error if found, or an error if not found.
func FetchAndFindEntity[T interface{ GetEntityID() string }](ctx *Context, entityID string) (*T, error) {
	states, err := FetchStates[T](ctx)
	if err != nil {
		return nil, err
	}

	for i := range states {
		if states[i].GetEntityID() == entityID {
			return &states[i], nil
		}
	}

	return nil, fmt.Errorf("%w: %s", ErrEntityNotFound, entityID)
}

// ListRequest provides a simplified interface for commands that fetch and list data.
type ListRequest[T any] struct {
	// MessageType is the WebSocket message type.
	MessageType string
	// Data is the optional message payload.
	Data map[string]any
	// Title is the output list title.
	Title string
	// Command is the command name for output metadata.
	Command string
	// Formatter formats each item for display.
	Formatter func(T, int) string
	// Filter optionally filters the results.
	Filter func(T) bool
}

// Execute runs the list request and outputs the results.
func (r *ListRequest[T]) Execute(ctx *Context) error {
	items, err := client.SendMessageTyped[[]T](ctx.Client, r.MessageType, r.Data)
	if err != nil {
		return err
	}

	// Apply filter if specified
	if r.Filter != nil {
		var filtered []T
		for _, item := range items {
			if r.Filter(item) {
				filtered = append(filtered, item)
			}
		}
		items = filtered
	}

	// Build list options
	var listOpts []output.ListOption[T]
	if r.Title != "" {
		listOpts = append(listOpts, output.ListTitle[T](r.Title))
	}
	if r.Command != "" {
		listOpts = append(listOpts, output.ListCommand[T](r.Command))
	}
	if r.Formatter != nil {
		listOpts = append(listOpts, output.ListFormatter(r.Formatter))
	}

	output.List(items, listOpts...)
	return nil
}

// TimelineRequest provides a simplified interface for commands that fetch and display timeline data.
type TimelineRequest[T any] struct {
	// MessageType is the WebSocket message type.
	MessageType string
	// Data is the optional message payload.
	Data map[string]any
	// Title is the output timeline title.
	Title string
	// Command is the command name for output metadata.
	Command string
	// Formatter formats each item for display.
	Formatter func(T) string
}

// Execute runs the timeline request and outputs the results.
func (r *TimelineRequest[T]) Execute(ctx *Context) error {
	items, err := client.SendMessageTyped[[]T](ctx.Client, r.MessageType, r.Data)
	if err != nil {
		return err
	}

	output.Timeline(items,
		output.TimelineTitle[T](r.Title),
		output.TimelineCommand[T](r.Command),
		output.TimelineFormatter(r.Formatter),
	)
	return nil
}

// MapRequest provides a simplified interface for commands that fetch map data and extract a specific key.
type MapRequest[T any] struct {
	// MessageType is the WebSocket message type.
	MessageType string
	// Data is the optional message payload.
	Data map[string]any
	// Key is the map key to extract.
	Key string
	// EmptyMessage is shown when no data is found.
	EmptyMessage string
}

// Execute runs the map request and returns the value for the specified key.
func (r *MapRequest[T]) Execute(ctx *Context) ([]T, error) {
	result, err := client.SendMessageTyped[map[string][]T](ctx.Client, r.MessageType, r.Data)
	if err != nil {
		return nil, err
	}

	data, ok := result[r.Key]
	if !ok || len(data) == 0 {
		if r.EmptyMessage != "" {
			output.Message(r.EmptyMessage)
		}
		return nil, nil
	}

	return data, nil
}

// SimpleHandler creates a handler that executes a message request and outputs the result.
// This is the simplest pattern for commands that just fetch and display data.
func SimpleHandler[T any](msgType string, getData func(*Context) map[string]any, command string) Handler {
	return func(ctx *Context) error {
		var data map[string]any
		if getData != nil {
			data = getData(ctx)
		}

		req := NewRequest[T](msgType, data)
		return req.ExecuteAndOutput(ctx, WithOutputCommand(command))
	}
}

// TransformHandler creates a handler that fetches data, transforms it, and outputs the result.
// Use this when you need to process the response before outputting.
func TransformHandler[T any, R any](
	msgType string,
	getData func(*Context) map[string]any,
	transform func(T) (R, error),
	command string,
) Handler {
	return func(ctx *Context) error {
		var data map[string]any
		if getData != nil {
			data = getData(ctx)
		}

		result, err := client.SendMessageTyped[T](ctx.Client, msgType, data)
		if err != nil {
			return err
		}

		transformed, err := transform(result)
		if err != nil {
			return err
		}

		output.Data(transformed, output.WithCommand(command))
		return nil
	}
}
