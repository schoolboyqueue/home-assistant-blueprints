package handlers

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/gorilla/websocket"

	"github.com/home-assistant-blueprints/testfixtures"

	"github.com/home-assistant-blueprints/ha-ws-client-go/internal/client"
	"github.com/home-assistant-blueprints/ha-ws-client-go/internal/output"
)

// =====================================
// Mock WebSocket Handlers
// =====================================

// MockHandler defines a function that handles WebSocket requests in tests.
// It receives the request type, data, and should return the result to send back.
type MockHandler func(msgType string, data map[string]any) any

// MessageRouter routes WebSocket requests to the appropriate mock handler based on message type.
type MessageRouter struct {
	handlers map[string]MockHandler
	t        *testing.T
}

// NewMessageRouter creates a new message router for testing.
func NewMessageRouter(t *testing.T) *MessageRouter {
	t.Helper()
	return &MessageRouter{
		handlers: make(map[string]MockHandler),
		t:        t,
	}
}

// On registers a handler for a specific message type.
func (r *MessageRouter) On(msgType string, handler MockHandler) *MessageRouter {
	r.handlers[msgType] = handler
	return r
}

// OnSuccess registers a handler that returns a static success result.
func (r *MessageRouter) OnSuccess(msgType string, result any) *MessageRouter {
	r.handlers[msgType] = func(_ string, _ map[string]any) any {
		return result
	}
	return r
}

// OnError registers a handler that returns an error result.
func (r *MessageRouter) OnError(msgType, code, message string) *MessageRouter {
	r.handlers[msgType] = func(_ string, _ map[string]any) any {
		return &testfixtures.HAError{Code: code, Message: message}
	}
	return r
}

// Handler returns a testfixtures.WSHandler that routes messages to registered handlers.
func (r *MessageRouter) Handler() testfixtures.WSHandler {
	return func(conn *websocket.Conn) {
		for {
			_, data, err := conn.ReadMessage()
			if err != nil {
				return
			}

			var req map[string]any
			if err := json.Unmarshal(data, &req); err != nil {
				return
			}

			msgType, ok := req["type"].(string)
			if !ok {
				return
			}
			reqID, ok := req["id"].(float64)
			if !ok {
				return
			}
			id := int(reqID)

			// Get data from request if present
			msgData, ok := req["data"].(map[string]any)
			if !ok {
				// Some messages put data at the top level (not nested under "data")
				msgData = req
			}

			// Find handler for this message type
			handler, ok := r.handlers[msgType]
			if !ok {
				// Default: return empty success
				resp := testfixtures.NewSuccessMessage(id, nil)
				if err := conn.WriteJSON(resp); err != nil {
					return
				}
				continue
			}

			result := handler(msgType, msgData)

			// Check if handler returned an error
			if haErr, ok := result.(*testfixtures.HAError); ok {
				resp := testfixtures.NewErrorMessage(id, haErr.Code, haErr.Message)
				if err := conn.WriteJSON(resp); err != nil {
					return
				}
				continue
			}

			resp := testfixtures.NewSuccessMessage(id, result)
			if err := conn.WriteJSON(resp); err != nil {
				return
			}
		}
	}
}

// =====================================
// Test Context Helpers
// =====================================

// TestContextOptions configures a test context.
type TestContextOptions struct {
	Args     []string
	FromTime *time.Time
	ToTime   *time.Time
	Config   *HandlerConfig
	Ctx      context.Context
}

// TestContextOption is a functional option for configuring test contexts.
type TestContextOption func(*TestContextOptions)

// WithArgs sets the command arguments.
func WithArgs(args ...string) TestContextOption {
	return func(o *TestContextOptions) {
		o.Args = args
	}
}

// WithFromTime sets the from time for history queries.
func WithFromTime(t time.Time) TestContextOption {
	return func(o *TestContextOptions) {
		o.FromTime = &t
	}
}

// WithToTime sets the to time for history queries.
func WithToTime(t time.Time) TestContextOption {
	return func(o *TestContextOptions) {
		o.ToTime = &t
	}
}

// WithMockTimeRange sets both from and to times for history queries.
func WithMockTimeRange(from, to time.Time) TestContextOption {
	return func(o *TestContextOptions) {
		o.FromTime = &from
		o.ToTime = &to
	}
}

// WithHandlerConfig sets a pre-configured handler config.
func WithHandlerConfig(cfg *HandlerConfig) TestContextOption {
	return func(o *TestContextOptions) {
		o.Config = cfg
	}
}

// WithContext sets the Go context for cancellation.
func WithContext(ctx context.Context) TestContextOption {
	return func(o *TestContextOptions) {
		o.Ctx = ctx
	}
}

// NewTestContext creates a Context for testing with a mock WebSocket server.
// The router should be configured with On/OnSuccess/OnError before calling this.
// Returns the context and a cleanup function that should be deferred.
func NewTestContext(t *testing.T, router *MessageRouter, opts ...TestContextOption) (*Context, func()) {
	t.Helper()

	options := &TestContextOptions{}
	for _, opt := range opts {
		opt(options)
	}

	// Create test server with the router's handler
	server := testfixtures.TestServer(t, router.Handler())

	// Dial the server using testfixtures helper
	conn := testfixtures.DialServer(t, server)

	// Create client
	var c *client.Client
	if options.Ctx != nil {
		c = client.NewWithContext(options.Ctx, conn)
	} else {
		c = client.New(conn)
	}

	ctx := &Context{
		Client:   c,
		Args:     options.Args,
		FromTime: options.FromTime,
		ToTime:   options.ToTime,
		Config:   options.Config,
		Ctx:      options.Ctx,
	}

	cleanup := func() {
		c.Close()
		server.Close()
	}

	return ctx, cleanup
}

// =====================================
// Quick Test Context (for simple tests)
// =====================================

// QuickTestContext creates a simple test context for handlers that return a static result.
// This is a convenience wrapper for simple test cases.
func QuickTestContext(t *testing.T, msgType string, result any, args ...string) (*Context, func()) {
	t.Helper()
	router := NewMessageRouter(t).OnSuccess(msgType, result)
	return NewTestContext(t, router, WithArgs(args...))
}

// =====================================
// Response Builders
// =====================================

// StateListResult creates a result containing a list of entity states.
func StateListResult(states ...testfixtures.HAState) []testfixtures.HAState {
	return states
}

// EntityRegistryResult creates an entity registry result.
func EntityRegistryResult(entries ...testfixtures.EntityEntry) []testfixtures.EntityEntry {
	return entries
}

// DeviceRegistryResult creates a device registry result.
func DeviceRegistryResult(entries ...testfixtures.DeviceEntry) []testfixtures.DeviceEntry {
	return entries
}

// AreaRegistryResult creates an area registry result.
func AreaRegistryResult(entries ...testfixtures.AreaEntry) []testfixtures.AreaEntry {
	return entries
}

// HistoryResult creates a history query result.
// The result is a map from entity_id to list of history states.
func HistoryResult(entityID string, states ...testfixtures.HistoryState) map[string][]testfixtures.HistoryState {
	return map[string][]testfixtures.HistoryState{
		entityID: states,
	}
}

// LogbookResult creates a logbook query result.
func LogbookResult(entries ...testfixtures.LogbookEntry) []testfixtures.LogbookEntry {
	return entries
}

// =====================================
// Output Capture (for testing output)
// =====================================

// CaptureOutput captures output written during handler execution.
// Returns the original config so it can be restored.
func CaptureOutput() (*output.Config, func()) {
	original := output.GetConfig()
	testConfig := &output.Config{
		Format:         output.FormatJSON,
		ShowTimestamps: true,
		ShowHeaders:    false,
	}
	output.SetConfig(testConfig)
	return original, func() {
		output.SetConfig(original)
	}
}
