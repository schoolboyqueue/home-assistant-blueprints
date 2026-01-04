package handlers

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/gorilla/websocket"

	"github.com/home-assistant-blueprints/ha-ws-client-go/internal/client"
	"github.com/home-assistant-blueprints/ha-ws-client-go/internal/output"
	"github.com/home-assistant-blueprints/ha-ws-client-go/internal/types"
	"github.com/home-assistant-blueprints/testfixtures"
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
func (r *MessageRouter) OnError(msgType string, code, message string) *MessageRouter {
	r.handlers[msgType] = func(_ string, _ map[string]any) any {
		return &types.HAError{Code: code, Message: message}
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

			msgType, _ := req["type"].(string)
			reqID, _ := req["id"].(float64)
			id := int(reqID)

			// Get data from request if present
			msgData, _ := req["data"].(map[string]any)
			if msgData == nil {
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
			if haErr, ok := result.(*types.HAError); ok {
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
func StateListResult(states ...types.HAState) []types.HAState {
	return states
}

// TraceListResult creates a result containing a list of trace info entries.
func TraceListResult(traces ...types.TraceInfo) map[string][]types.TraceInfo {
	if len(traces) == 0 {
		return map[string][]types.TraceInfo{}
	}
	// Group by item_id
	result := make(map[string][]types.TraceInfo)
	for _, trace := range traces {
		result[trace.ItemID] = append(result[trace.ItemID], trace)
	}
	return result
}

// TraceDetailResult creates a trace detail response.
func TraceDetailResult(detail types.TraceDetail) types.TraceDetail {
	return detail
}

// EntityRegistryResult creates an entity registry result.
func EntityRegistryResult(entries ...types.EntityEntry) []types.EntityEntry {
	return entries
}

// DeviceRegistryResult creates a device registry result.
func DeviceRegistryResult(entries ...types.DeviceEntry) []types.DeviceEntry {
	return entries
}

// AreaRegistryResult creates an area registry result.
func AreaRegistryResult(entries ...types.AreaEntry) []types.AreaEntry {
	return entries
}

// HistoryResult creates a history query result.
// The result is a map from entity_id to list of history states.
func HistoryResult(entityID string, states ...types.HistoryState) map[string][]types.HistoryState {
	return map[string][]types.HistoryState{
		entityID: states,
	}
}

// LogbookResult creates a logbook query result.
func LogbookResult(entries ...types.LogbookEntry) []types.LogbookEntry {
	return entries
}

// ConfigResult creates a config result.
func ConfigResult(cfg types.HAConfig) types.HAConfig {
	return cfg
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

// =====================================
// Mock State Builders
// =====================================

// MockState creates a HAState for testing.
func MockState(entityID, state string) types.HAState {
	return types.HAState{
		EntityID: entityID,
		State:    state,
	}
}

// MockStateWithAttrs creates a HAState with attributes.
func MockStateWithAttrs(entityID, state string, attrs map[string]any) types.HAState {
	return types.HAState{
		EntityID:   entityID,
		State:      state,
		Attributes: attrs,
	}
}

// MockStateWithTimestamps creates a HAState with timestamps.
func MockStateWithTimestamps(entityID, state string) types.HAState {
	now := time.Now().UTC().Format(time.RFC3339)
	return types.HAState{
		EntityID:    entityID,
		State:       state,
		LastChanged: now,
		LastUpdated: now,
	}
}

// =====================================
// Mock Trace Builders
// =====================================

// MockTraceInfo creates a TraceInfo for testing.
func MockTraceInfo(itemID, runID, state string) types.TraceInfo {
	return types.TraceInfo{
		ItemID: itemID,
		RunID:  runID,
		State:  state,
	}
}

// MockTraceInfoFull creates a fully populated TraceInfo.
func MockTraceInfoFull(itemID, runID, state, scriptExecution string) types.TraceInfo {
	now := time.Now().UTC().Format(time.RFC3339)
	return types.TraceInfo{
		ItemID:          itemID,
		RunID:           runID,
		State:           state,
		ScriptExecution: scriptExecution,
		Timestamp: &types.Timestamp{
			Start:  now,
			Finish: now,
		},
		Context: &types.HAContext{
			ID: "test-context-id",
		},
	}
}

// MockTraceDetail creates a TraceDetail for testing.
func MockTraceDetail(itemID, runID string) types.TraceDetail {
	now := time.Now().UTC().Format(time.RFC3339)
	return types.TraceDetail{
		ItemID:          itemID,
		RunID:           runID,
		Domain:          "automation",
		ScriptExecution: "finished",
		Timestamp: &types.Timestamp{
			Start:  now,
			Finish: now,
		},
		Context: &types.HAContext{
			ID: "test-context-id",
		},
		Trace: map[string][]types.TraceStep{
			"action/0": {
				{
					Path:      "action/0",
					Timestamp: now,
					Result: &types.TraceResult{
						Enabled: true,
					},
				},
			},
		},
	}
}

// MockTraceDetailWithTrigger creates a TraceDetail with trigger information.
func MockTraceDetailWithTrigger(itemID, runID string, trigger map[string]any) types.TraceDetail {
	detail := MockTraceDetail(itemID, runID)
	detail.Trigger = trigger
	return detail
}

// MockTraceDetailWithConfig creates a TraceDetail with automation config.
func MockTraceDetailWithConfig(itemID, runID string, config *types.AutomationConfig) types.TraceDetail {
	detail := MockTraceDetail(itemID, runID)
	detail.Config = config
	return detail
}

// =====================================
// Mock History Builders
// =====================================

// MockHistoryState creates a HistoryState for testing.
func MockHistoryState(state string, timestamp time.Time) types.HistoryState {
	return types.HistoryState{
		S:  state,
		LU: float64(timestamp.Unix()),
		LC: float64(timestamp.Unix()),
	}
}

// MockHistoryStateWithAttrs creates a HistoryState with attributes.
func MockHistoryStateWithAttrs(state string, timestamp time.Time, attrs map[string]any) types.HistoryState {
	return types.HistoryState{
		S:  state,
		LU: float64(timestamp.Unix()),
		LC: float64(timestamp.Unix()),
		A:  attrs,
	}
}

// =====================================
// Mock Logbook Builders
// =====================================

// MockLogbookEntry creates a LogbookEntry for testing.
func MockLogbookEntry(entityID, state, message string, when time.Time) types.LogbookEntry {
	return types.LogbookEntry{
		EntityID: entityID,
		State:    state,
		Message:  message,
		When:     float64(when.Unix()),
	}
}

// =====================================
// Mock Registry Builders
// =====================================

// MockEntityEntry creates an EntityEntry for testing.
func MockEntityEntry(entityID, name, platform string) types.EntityEntry {
	return types.EntityEntry{
		EntityID:     entityID,
		Name:         name,
		OriginalName: name,
		Platform:     platform,
	}
}

// MockDeviceEntry creates a DeviceEntry for testing.
func MockDeviceEntry(id, name, manufacturer, model string) types.DeviceEntry {
	return types.DeviceEntry{
		ID:           id,
		Name:         name,
		Manufacturer: manufacturer,
		Model:        model,
	}
}

// MockAreaEntry creates an AreaEntry for testing.
func MockAreaEntry(areaID, name string) types.AreaEntry {
	return types.AreaEntry{
		AreaID: areaID,
		Name:   name,
	}
}

// =====================================
// Mock Config Builders
// =====================================

// MockHAConfig creates an HAConfig for testing.
func MockHAConfig() types.HAConfig {
	return types.HAConfig{
		Version:      "2024.1.0",
		LocationName: "Test Home",
		TimeZone:     "America/New_York",
		UnitSystem: map[string]string{
			"length":      "mi",
			"temperature": "°F",
		},
		State:      "RUNNING",
		Components: []string{"homeassistant", "automation", "script"},
	}
}
