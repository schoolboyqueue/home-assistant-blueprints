package handlers

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/home-assistant-blueprints/ha-ws-client-go/internal/types"
)

func TestNewRequest(t *testing.T) {
	tests := []struct {
		name    string
		msgType string
		data    map[string]any
	}{
		{
			name:    "simple request without data",
			msgType: "get_states",
			data:    nil,
		},
		{
			name:    "request with data",
			msgType: "trace/get",
			data: map[string]any{
				"domain":  "automation",
				"item_id": "test_automation",
				"run_id":  "12345",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := NewRequest[map[string]any](tt.msgType, tt.data)
			if req.Type != tt.msgType {
				t.Errorf("NewRequest().Type = %v, want %v", req.Type, tt.msgType)
			}
			if tt.data != nil {
				for k, v := range tt.data {
					if req.Data[k] != v {
						t.Errorf("NewRequest().Data[%s] = %v, want %v", k, req.Data[k], v)
					}
				}
			}
		})
	}
}

func TestOutputConfig(t *testing.T) {
	tests := []struct {
		name        string
		opts        []OutputOption
		wantCommand string
		wantSummary string
		wantCount   int
	}{
		{
			name:        "with command",
			opts:        []OutputOption{WithOutputCommand("test-cmd")},
			wantCommand: "test-cmd",
		},
		{
			name:        "with summary",
			opts:        []OutputOption{WithOutputSummary("test summary")},
			wantSummary: "test summary",
		},
		{
			name:      "with count",
			opts:      []OutputOption{WithOutputCount(42)},
			wantCount: 42,
		},
		{
			name: "combined options",
			opts: []OutputOption{
				WithOutputCommand("combined-cmd"),
				WithOutputSummary("combined summary"),
				WithOutputCount(100),
			},
			wantCommand: "combined-cmd",
			wantSummary: "combined summary",
			wantCount:   100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := OutputConfig{}
			for _, opt := range tt.opts {
				opt(&cfg)
			}
			if cfg.Command != tt.wantCommand {
				t.Errorf("OutputConfig.Command = %v, want %v", cfg.Command, tt.wantCommand)
			}
			if cfg.Summary != tt.wantSummary {
				t.Errorf("OutputConfig.Summary = %v, want %v", cfg.Summary, tt.wantSummary)
			}
			if cfg.Count != tt.wantCount {
				t.Errorf("OutputConfig.Count = %v, want %v", cfg.Count, tt.wantCount)
			}
		})
	}
}

func TestDispatchCreation(t *testing.T) {
	d := Dispatch[[]string]("get_states", nil)
	if d.request.Type != "get_states" {
		t.Errorf("Dispatch request.Type = %v, want get_states", d.request.Type)
	}
}

func TestDispatchTransform(t *testing.T) {
	called := false
	d := Dispatch[[]string]("test", nil).Transform(func(s []string) ([]string, error) {
		called = true
		return append(s, "transformed"), nil
	})

	// Transform function should be set
	if d.transform == nil {
		t.Fatal("Transform function not set")
	}

	// Verify transform works by calling it directly
	result, err := d.transform([]string{"original"})
	if err != nil {
		t.Errorf("Transform error: %v", err)
	}
	if !called {
		t.Error("Transform function was not called")
	}
	if len(result) != 2 || result[1] != "transformed" {
		t.Errorf("Transform result = %v, want [original transformed]", result)
	}
}

func TestDispatchOutput(t *testing.T) {
	called := false
	d := Dispatch[string]("test", nil).Output(func(_ string) error {
		called = true
		return nil
	})

	// Output function should be set
	if d.outputFn == nil {
		t.Fatal("Output function not set")
	}

	// Verify output works by calling it directly
	err := d.outputFn("test")
	if err != nil {
		t.Errorf("Output error: %v", err)
	}
	if !called {
		t.Error("Output function was not called")
	}
}

func TestListRequest(t *testing.T) {
	lr := &ListRequest[string]{
		MessageType: "test/list",
		Title:       "Test Title",
		Command:     "test-cmd",
		Formatter: func(s string, _ int) string {
			return s
		},
		Filter: func(s string) bool {
			return s != ""
		},
	}

	if lr.MessageType != "test/list" {
		t.Errorf("ListRequest.MessageType = %v, want test/list", lr.MessageType)
	}
	if lr.Title != "Test Title" {
		t.Errorf("ListRequest.Title = %v, want Test Title", lr.Title)
	}
	if lr.Command != "test-cmd" {
		t.Errorf("ListRequest.Command = %v, want test-cmd", lr.Command)
	}
	if lr.Formatter == nil {
		t.Error("ListRequest.Formatter should not be nil")
	}
	if lr.Filter == nil {
		t.Error("ListRequest.Filter should not be nil")
	}
	// Test filter
	if !lr.Filter("hello") {
		t.Error("Filter should return true for non-empty string")
	}
	if lr.Filter("") {
		t.Error("Filter should return false for empty string")
	}
}

func TestTimelineRequest(t *testing.T) {
	tr := &TimelineRequest[string]{
		MessageType: "test/timeline",
		Title:       "Timeline Title",
		Command:     "timeline-cmd",
		Formatter: func(s string) string {
			return "formatted: " + s
		},
	}

	if tr.MessageType != "test/timeline" {
		t.Errorf("TimelineRequest.MessageType = %v, want test/timeline", tr.MessageType)
	}
	if tr.Title != "Timeline Title" {
		t.Errorf("TimelineRequest.Title = %v, want Timeline Title", tr.Title)
	}
	if tr.Command != "timeline-cmd" {
		t.Errorf("TimelineRequest.Command = %v, want timeline-cmd", tr.Command)
	}
	if tr.Formatter == nil {
		t.Error("TimelineRequest.Formatter should not be nil")
	}
	// Test formatter
	result := tr.Formatter("test")
	if result != "formatted: test" {
		t.Errorf("Formatter result = %v, want 'formatted: test'", result)
	}
}

func TestMapRequest(t *testing.T) {
	mr := &MapRequest[string]{
		MessageType:  "test/map",
		Key:          "testkey",
		EmptyMessage: "No data found",
	}

	if mr.MessageType != "test/map" {
		t.Errorf("MapRequest.MessageType = %v, want test/map", mr.MessageType)
	}
	if mr.Key != "testkey" {
		t.Errorf("MapRequest.Key = %v, want testkey", mr.Key)
	}
	if mr.EmptyMessage != "No data found" {
		t.Errorf("MapRequest.EmptyMessage = %v, want 'No data found'", mr.EmptyMessage)
	}
}

// =====================================
// Integration Tests with Mock Server
// =====================================

func TestMessageRequest_Execute_Success(t *testing.T) {
	t.Parallel()

	states := StateListResult(
		MockState("light.kitchen", "on"),
		MockState("sensor.temperature", "22.5"),
	)

	router := NewMessageRouter(t).
		OnSuccess("get_states", states)

	ctx, cleanup := NewTestContext(t, router)
	defer cleanup()

	req := NewRequest[[]types.HAState]("get_states", nil)
	result, err := req.Execute(ctx)

	require.NoError(t, err)
	require.Len(t, result, 2)
	assert.Equal(t, "light.kitchen", result[0].EntityID)
	assert.Equal(t, "on", result[0].State)
	assert.Equal(t, "sensor.temperature", result[1].EntityID)
	assert.Equal(t, "22.5", result[1].State)
}

func TestMessageRequest_Execute_Error(t *testing.T) {
	t.Parallel()

	router := NewMessageRouter(t).
		OnError("get_states", "not_authorized", "Authentication required")

	ctx, cleanup := NewTestContext(t, router)
	defer cleanup()

	req := NewRequest[[]types.HAState]("get_states", nil)
	_, err := req.Execute(ctx)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "Authentication required")
}

func TestMessageRequest_Execute_WithData(t *testing.T) {
	t.Parallel()

	traceDetail := MockTraceDetail("automation.test", "1234567890")

	router := NewMessageRouter(t).
		On("trace/get", func(msgType string, data map[string]any) any {
			// Verify the data was passed correctly
			assert.Equal(t, "automation", data["domain"])
			assert.Equal(t, "test", data["item_id"])
			assert.Equal(t, "1234567890", data["run_id"])
			return traceDetail
		})

	ctx, cleanup := NewTestContext(t, router)
	defer cleanup()

	req := NewRequest[types.TraceDetail]("trace/get", map[string]any{
		"domain":  "automation",
		"item_id": "test",
		"run_id":  "1234567890",
	})

	result, err := req.Execute(ctx)
	require.NoError(t, err)
	assert.Equal(t, "automation.test", result.ItemID)
	assert.Equal(t, "1234567890", result.RunID)
}

func TestMessageRequest_ExecuteRaw_Success(t *testing.T) {
	t.Parallel()

	config := MockHAConfig()
	router := NewMessageRouter(t).
		OnSuccess("get_config", config)

	ctx, cleanup := NewTestContext(t, router)
	defer cleanup()

	req := NewRequest[types.HAConfig]("get_config", nil)
	msg, err := req.ExecuteRaw(ctx)

	require.NoError(t, err)
	require.NotNil(t, msg)
	require.NotNil(t, msg.Success)
	assert.True(t, *msg.Success)
	assert.NotNil(t, msg.Result)
}

func TestMessageRequest_ExecuteRaw_Error(t *testing.T) {
	t.Parallel()

	router := NewMessageRouter(t).
		OnError("get_config", "unavailable", "Service unavailable")

	ctx, cleanup := NewTestContext(t, router)
	defer cleanup()

	req := NewRequest[types.HAConfig]("get_config", nil)
	_, err := req.ExecuteRaw(ctx)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "Service unavailable")
}

func TestMessageRequest_ExecuteAndOutput_Success(t *testing.T) {
	t.Parallel()

	config := MockHAConfig()
	router := NewMessageRouter(t).
		OnSuccess("get_config", config)

	ctx, cleanup := NewTestContext(t, router)
	defer cleanup()

	// Capture output
	_, restoreOutput := CaptureOutput()
	defer restoreOutput()

	req := NewRequest[types.HAConfig]("get_config", nil)
	err := req.ExecuteAndOutput(ctx,
		WithOutputCommand("config"),
		WithOutputSummary("Home Assistant configuration"),
		WithOutputCount(1),
	)

	require.NoError(t, err)
}

func TestMessageRequest_ExecuteAndOutput_Error(t *testing.T) {
	t.Parallel()

	router := NewMessageRouter(t).
		OnError("get_config", "not_found", "Config not found")

	ctx, cleanup := NewTestContext(t, router)
	defer cleanup()

	req := NewRequest[types.HAConfig]("get_config", nil)
	err := req.ExecuteAndOutput(ctx, WithOutputCommand("config"))

	require.Error(t, err)
	assert.Contains(t, err.Error(), "Config not found")
}

func TestMessageRequest_ExecuteAndOutput_NoOptions(t *testing.T) {
	t.Parallel()

	states := StateListResult(
		MockState("light.kitchen", "on"),
	)
	router := NewMessageRouter(t).
		OnSuccess("get_states", states)

	ctx, cleanup := NewTestContext(t, router)
	defer cleanup()

	// Capture output
	_, restoreOutput := CaptureOutput()
	defer restoreOutput()

	req := NewRequest[[]types.HAState]("get_states", nil)
	err := req.ExecuteAndOutput(ctx) // No options

	require.NoError(t, err)
}

func TestMessageDispatcher_Execute_FullPipeline(t *testing.T) {
	t.Parallel()

	states := StateListResult(
		MockState("light.kitchen", "on"),
		MockState("light.bedroom", "off"),
		MockState("sensor.temperature", "22.5"),
	)

	router := NewMessageRouter(t).
		OnSuccess("get_states", states)

	ctx, cleanup := NewTestContext(t, router)
	defer cleanup()

	// Capture output
	_, restoreOutput := CaptureOutput()
	defer restoreOutput()

	// Use dispatcher with transform to filter only lights
	var outputCalled bool
	err := Dispatch[[]types.HAState]("get_states", nil).
		Transform(func(states []types.HAState) ([]types.HAState, error) {
			var lights []types.HAState
			for _, s := range states {
				if len(s.EntityID) > 5 && s.EntityID[:5] == "light" {
					lights = append(lights, s)
				}
			}
			return lights, nil
		}).
		Output(func(states []types.HAState) error {
			outputCalled = true
			assert.Len(t, states, 2)
			return nil
		}).
		Execute(ctx)

	require.NoError(t, err)
	assert.True(t, outputCalled)
}

func TestMessageDispatcher_Execute_TransformError(t *testing.T) {
	t.Parallel()

	states := StateListResult(MockState("light.kitchen", "on"))

	router := NewMessageRouter(t).
		OnSuccess("get_states", states)

	ctx, cleanup := NewTestContext(t, router)
	defer cleanup()

	err := Dispatch[[]types.HAState]("get_states", nil).
		Transform(func(_ []types.HAState) ([]types.HAState, error) {
			return nil, assert.AnError
		}).
		Execute(ctx)

	require.Error(t, err)
	assert.Equal(t, assert.AnError, err)
}

func TestMessageDispatcher_Execute_DefaultOutput(t *testing.T) {
	t.Parallel()

	states := StateListResult(MockState("light.kitchen", "on"))

	router := NewMessageRouter(t).
		OnSuccess("get_states", states)

	ctx, cleanup := NewTestContext(t, router)
	defer cleanup()

	// Capture output
	_, restoreOutput := CaptureOutput()
	defer restoreOutput()

	// No custom output function - should use default JSON output
	err := Dispatch[[]types.HAState]("get_states", nil).Execute(ctx)

	require.NoError(t, err)
}

func TestMessageDispatcher_Execute_ServerError(t *testing.T) {
	t.Parallel()

	router := NewMessageRouter(t).
		OnError("get_states", "server_error", "Internal error")

	ctx, cleanup := NewTestContext(t, router)
	defer cleanup()

	err := Dispatch[[]types.HAState]("get_states", nil).Execute(ctx)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "Internal error")
}

func TestMessageDispatcher_Result_Success(t *testing.T) {
	t.Parallel()

	states := StateListResult(
		MockState("light.kitchen", "on"),
		MockState("light.bedroom", "off"),
	)

	router := NewMessageRouter(t).
		OnSuccess("get_states", states)

	ctx, cleanup := NewTestContext(t, router)
	defer cleanup()

	result, err := Dispatch[[]types.HAState]("get_states", nil).
		Transform(func(states []types.HAState) ([]types.HAState, error) {
			// Filter to only "on" states
			var on []types.HAState
			for _, s := range states {
				if s.State == "on" {
					on = append(on, s)
				}
			}
			return on, nil
		}).
		Result(ctx)

	require.NoError(t, err)
	require.Len(t, result, 1)
	assert.Equal(t, "light.kitchen", result[0].EntityID)
}

func TestMessageDispatcher_Result_Error(t *testing.T) {
	t.Parallel()

	router := NewMessageRouter(t).
		OnError("get_states", "timeout", "Request timeout")

	ctx, cleanup := NewTestContext(t, router)
	defer cleanup()

	result, err := Dispatch[[]types.HAState]("get_states", nil).Result(ctx)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "Request timeout")
	assert.Nil(t, result)
}

func TestMessageDispatcher_Result_TransformError(t *testing.T) {
	t.Parallel()

	states := StateListResult(MockState("light.kitchen", "on"))

	router := NewMessageRouter(t).
		OnSuccess("get_states", states)

	ctx, cleanup := NewTestContext(t, router)
	defer cleanup()

	result, err := Dispatch[[]types.HAState]("get_states", nil).
		Transform(func(_ []types.HAState) ([]types.HAState, error) {
			return nil, assert.AnError
		}).
		Result(ctx)

	require.Error(t, err)
	assert.Nil(t, result)
}

func TestFetchStates_Success(t *testing.T) {
	t.Parallel()

	states := StateListResult(
		MockState("light.kitchen", "on"),
		MockState("sensor.temperature", "22.5"),
		MockState("binary_sensor.motion", "on"),
	)

	router := NewMessageRouter(t).
		OnSuccess("get_states", states)

	ctx, cleanup := NewTestContext(t, router)
	defer cleanup()

	result, err := FetchStates[types.HAState](ctx)

	require.NoError(t, err)
	require.Len(t, result, 3)
	assert.Equal(t, "light.kitchen", result[0].EntityID)
}

func TestFetchStates_Empty(t *testing.T) {
	t.Parallel()

	router := NewMessageRouter(t).
		OnSuccess("get_states", []types.HAState{})

	ctx, cleanup := NewTestContext(t, router)
	defer cleanup()

	result, err := FetchStates[types.HAState](ctx)

	require.NoError(t, err)
	assert.Empty(t, result)
}

func TestFetchStates_Error(t *testing.T) {
	t.Parallel()

	router := NewMessageRouter(t).
		OnError("get_states", "unavailable", "Home Assistant unavailable")

	ctx, cleanup := NewTestContext(t, router)
	defer cleanup()

	result, err := FetchStates[types.HAState](ctx)

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "Home Assistant unavailable")
}

// testStateWithGetter wraps HAState and implements GetEntityID interface for testing.
type testStateWithGetter struct {
	types.HAState
}

func (s testStateWithGetter) GetEntityID() string {
	return s.EntityID
}

func TestFetchAndFindEntity_Found(t *testing.T) {
	t.Parallel()

	// Create test states that will be serialized/deserialized correctly
	states := []testStateWithGetter{
		{MockState("light.kitchen", "on")},
		{MockState("light.bedroom", "off")},
		{MockState("sensor.temperature", "22.5")},
	}

	router := NewMessageRouter(t).
		OnSuccess("get_states", states)

	ctx, cleanup := NewTestContext(t, router)
	defer cleanup()

	result, err := FetchAndFindEntity[testStateWithGetter](ctx, "light.bedroom")

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "light.bedroom", result.EntityID)
	assert.Equal(t, "off", result.State)
}

func TestFetchAndFindEntity_NotFound(t *testing.T) {
	t.Parallel()

	states := []testStateWithGetter{
		{MockState("light.kitchen", "on")},
		{MockState("sensor.temperature", "22.5")},
	}

	router := NewMessageRouter(t).
		OnSuccess("get_states", states)

	ctx, cleanup := NewTestContext(t, router)
	defer cleanup()

	result, err := FetchAndFindEntity[testStateWithGetter](ctx, "light.nonexistent")

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "not found")
}

func TestFetchAndFindEntity_FetchError(t *testing.T) {
	t.Parallel()

	router := NewMessageRouter(t).
		OnError("get_states", "server_error", "Server error")

	ctx, cleanup := NewTestContext(t, router)
	defer cleanup()

	result, err := FetchAndFindEntity[testStateWithGetter](ctx, "light.kitchen")

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "Server error")
}

func TestListRequest_Execute_Success(t *testing.T) {
	t.Parallel()

	areas := AreaRegistryResult(
		MockAreaEntry("living_room", "Living Room"),
		MockAreaEntry("bedroom", "Master Bedroom"),
	)

	router := NewMessageRouter(t).
		OnSuccess("config/area_registry/list", areas)

	ctx, cleanup := NewTestContext(t, router)
	defer cleanup()

	// Capture output
	_, restoreOutput := CaptureOutput()
	defer restoreOutput()

	lr := &ListRequest[types.AreaEntry]{
		MessageType: "config/area_registry/list",
		Title:       "Areas",
		Command:     "areas",
		Formatter: func(a types.AreaEntry, _ int) string {
			return a.Name
		},
	}

	err := lr.Execute(ctx)
	require.NoError(t, err)
}

func TestListRequest_Execute_WithFilter(t *testing.T) {
	t.Parallel()

	entities := EntityRegistryResult(
		MockEntityEntry("light.kitchen", "Kitchen Light", "hue"),
		types.EntityEntry{
			EntityID:     "switch.disabled",
			Name:         "Disabled Switch",
			OriginalName: "Old Switch",
			Platform:     "zwave",
			DisabledBy:   "user",
		},
		MockEntityEntry("light.bedroom", "Bedroom Light", "hue"),
	)

	router := NewMessageRouter(t).
		OnSuccess("config/entity_registry/list", entities)

	ctx, cleanup := NewTestContext(t, router)
	defer cleanup()

	// Capture output
	_, restoreOutput := CaptureOutput()
	defer restoreOutput()

	lr := &ListRequest[types.EntityEntry]{
		MessageType: "config/entity_registry/list",
		Title:       "Entities",
		Command:     "entities",
		Formatter: func(e types.EntityEntry, _ int) string {
			return e.EntityID
		},
		Filter: func(e types.EntityEntry) bool {
			return e.DisabledBy == "" // Only show enabled entities
		},
	}

	err := lr.Execute(ctx)
	require.NoError(t, err)
}

func TestListRequest_Execute_Error(t *testing.T) {
	t.Parallel()

	router := NewMessageRouter(t).
		OnError("config/area_registry/list", "not_found", "Areas not found")

	ctx, cleanup := NewTestContext(t, router)
	defer cleanup()

	lr := &ListRequest[types.AreaEntry]{
		MessageType: "config/area_registry/list",
		Title:       "Areas",
		Command:     "areas",
	}

	err := lr.Execute(ctx)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Areas not found")
}

func TestListRequest_Execute_EmptyResult(t *testing.T) {
	t.Parallel()

	router := NewMessageRouter(t).
		OnSuccess("config/area_registry/list", []types.AreaEntry{})

	ctx, cleanup := NewTestContext(t, router)
	defer cleanup()

	// Capture output
	_, restoreOutput := CaptureOutput()
	defer restoreOutput()

	lr := &ListRequest[types.AreaEntry]{
		MessageType: "config/area_registry/list",
		Title:       "Areas",
		Command:     "areas",
	}

	err := lr.Execute(ctx)
	require.NoError(t, err)
}

func TestTimelineRequest_Execute_Success(t *testing.T) {
	t.Parallel()

	logbook := LogbookResult(
		MockLogbookEntry("light.kitchen", "on", "Light turned on", time.Now().Add(-10*time.Minute)),
		MockLogbookEntry("light.kitchen", "off", "Light turned off", time.Now().Add(-5*time.Minute)),
	)

	router := NewMessageRouter(t).
		OnSuccess("logbook/period", logbook)

	ctx, cleanup := NewTestContext(t, router)
	defer cleanup()

	// Capture output
	_, restoreOutput := CaptureOutput()
	defer restoreOutput()

	tr := &TimelineRequest[types.LogbookEntry]{
		MessageType: "logbook/period",
		Title:       "Logbook",
		Command:     "logbook",
		Formatter: func(e types.LogbookEntry) string {
			return e.Message
		},
	}

	err := tr.Execute(ctx)
	require.NoError(t, err)
}

func TestTimelineRequest_Execute_Error(t *testing.T) {
	t.Parallel()

	router := NewMessageRouter(t).
		OnError("logbook/period", "timeout", "Request timeout")

	ctx, cleanup := NewTestContext(t, router)
	defer cleanup()

	tr := &TimelineRequest[types.LogbookEntry]{
		MessageType: "logbook/period",
		Title:       "Logbook",
		Command:     "logbook",
	}

	err := tr.Execute(ctx)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Request timeout")
}

func TestTimelineRequest_Execute_Empty(t *testing.T) {
	t.Parallel()

	router := NewMessageRouter(t).
		OnSuccess("logbook/period", []types.LogbookEntry{})

	ctx, cleanup := NewTestContext(t, router)
	defer cleanup()

	// Capture output
	_, restoreOutput := CaptureOutput()
	defer restoreOutput()

	tr := &TimelineRequest[types.LogbookEntry]{
		MessageType: "logbook/period",
		Title:       "Logbook",
		Command:     "logbook",
		Formatter: func(e types.LogbookEntry) string {
			return e.Message
		},
	}

	err := tr.Execute(ctx)
	require.NoError(t, err)
}

func TestMapRequest_Execute_KeyFound(t *testing.T) {
	t.Parallel()

	result := map[string][]string{
		"events": {"event1", "event2", "event3"},
		"states": {"state1", "state2"},
	}

	router := NewMessageRouter(t).
		OnSuccess("trace/contexts", result)

	ctx, cleanup := NewTestContext(t, router)
	defer cleanup()

	mr := &MapRequest[string]{
		MessageType:  "trace/contexts",
		Key:          "events",
		EmptyMessage: "No events found",
	}

	data, err := mr.Execute(ctx)
	require.NoError(t, err)
	require.Len(t, data, 3)
	assert.Equal(t, "event1", data[0])
}

func TestMapRequest_Execute_KeyNotFound(t *testing.T) {
	t.Parallel()

	result := map[string][]string{
		"events": {"event1", "event2"},
	}

	router := NewMessageRouter(t).
		OnSuccess("trace/contexts", result)

	ctx, cleanup := NewTestContext(t, router)
	defer cleanup()

	// Capture output to verify empty message
	_, restoreOutput := CaptureOutput()
	defer restoreOutput()

	mr := &MapRequest[string]{
		MessageType:  "trace/contexts",
		Key:          "nonexistent",
		EmptyMessage: "No data found for this key",
	}

	data, err := mr.Execute(ctx)
	require.NoError(t, err)
	assert.Nil(t, data)
}

func TestMapRequest_Execute_EmptyArray(t *testing.T) {
	t.Parallel()

	result := map[string][]string{
		"events": {},
	}

	router := NewMessageRouter(t).
		OnSuccess("trace/contexts", result)

	ctx, cleanup := NewTestContext(t, router)
	defer cleanup()

	// Capture output
	_, restoreOutput := CaptureOutput()
	defer restoreOutput()

	mr := &MapRequest[string]{
		MessageType:  "trace/contexts",
		Key:          "events",
		EmptyMessage: "No events found",
	}

	data, err := mr.Execute(ctx)
	require.NoError(t, err)
	assert.Nil(t, data)
}

func TestMapRequest_Execute_Error(t *testing.T) {
	t.Parallel()

	router := NewMessageRouter(t).
		OnError("trace/contexts", "not_found", "Contexts not found")

	ctx, cleanup := NewTestContext(t, router)
	defer cleanup()

	mr := &MapRequest[string]{
		MessageType: "trace/contexts",
		Key:         "events",
	}

	data, err := mr.Execute(ctx)
	require.Error(t, err)
	assert.Nil(t, data)
	assert.Contains(t, err.Error(), "Contexts not found")
}

func TestSimpleHandler_Success(t *testing.T) {
	t.Parallel()

	config := MockHAConfig()
	router := NewMessageRouter(t).
		OnSuccess("get_config", config)

	ctx, cleanup := NewTestContext(t, router)
	defer cleanup()

	// Capture output
	_, restoreOutput := CaptureOutput()
	defer restoreOutput()

	handler := SimpleHandler[types.HAConfig]("get_config", nil, "config")
	err := handler(ctx)

	require.NoError(t, err)
}

func TestSimpleHandler_WithData(t *testing.T) {
	t.Parallel()

	states := StateListResult(
		MockState("light.kitchen", "on"),
	)

	router := NewMessageRouter(t).
		On("get_states", func(_ string, data map[string]any) any {
			// Verify data is passed
			if limit, ok := data["limit"]; ok {
				assert.Equal(t, 10, int(limit.(float64)))
			}
			return states
		})

	ctx, cleanup := NewTestContext(t, router)
	defer cleanup()

	// Capture output
	_, restoreOutput := CaptureOutput()
	defer restoreOutput()

	handler := SimpleHandler[[]types.HAState]("get_states",
		func(_ *Context) map[string]any {
			return map[string]any{"limit": 10}
		},
		"states",
	)

	err := handler(ctx)
	require.NoError(t, err)
}

func TestSimpleHandler_Error(t *testing.T) {
	t.Parallel()

	router := NewMessageRouter(t).
		OnError("get_config", "unavailable", "Service unavailable")

	ctx, cleanup := NewTestContext(t, router)
	defer cleanup()

	handler := SimpleHandler[types.HAConfig]("get_config", nil, "config")
	err := handler(ctx)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "Service unavailable")
}

func TestTransformHandler_Success(t *testing.T) {
	t.Parallel()

	states := StateListResult(
		MockState("light.kitchen", "on"),
		MockState("light.bedroom", "off"),
		MockState("sensor.temperature", "22.5"),
	)

	router := NewMessageRouter(t).
		OnSuccess("get_states", states)

	ctx, cleanup := NewTestContext(t, router)
	defer cleanup()

	// Capture output
	_, restoreOutput := CaptureOutput()
	defer restoreOutput()

	// Transform to count states by domain
	type countResult struct {
		Domain string `json:"domain"`
		Count  int    `json:"count"`
	}

	handler := TransformHandler[[]types.HAState, []countResult](
		"get_states",
		nil,
		func(states []types.HAState) ([]countResult, error) {
			counts := make(map[string]int)
			for _, s := range states {
				// Split entity ID at first dot
				for i, c := range s.EntityID {
					if c == '.' {
						domain := s.EntityID[:i]
						counts[domain]++
						break
					}
				}
			}
			var result []countResult
			for domain, count := range counts {
				result = append(result, countResult{Domain: domain, Count: count})
			}
			return result, nil
		},
		"domain_counts",
	)

	err := handler(ctx)
	require.NoError(t, err)
}

func TestTransformHandler_TransformError(t *testing.T) {
	t.Parallel()

	states := StateListResult(MockState("light.kitchen", "on"))

	router := NewMessageRouter(t).
		OnSuccess("get_states", states)

	ctx, cleanup := NewTestContext(t, router)
	defer cleanup()

	handler := TransformHandler[[]types.HAState, string](
		"get_states",
		nil,
		func(_ []types.HAState) (string, error) {
			return "", assert.AnError
		},
		"transformed",
	)

	err := handler(ctx)
	require.Error(t, err)
	assert.Equal(t, assert.AnError, err)
}

func TestTransformHandler_FetchError(t *testing.T) {
	t.Parallel()

	router := NewMessageRouter(t).
		OnError("get_states", "unavailable", "Service unavailable")

	ctx, cleanup := NewTestContext(t, router)
	defer cleanup()

	handler := TransformHandler[[]types.HAState, string](
		"get_states",
		nil,
		func(_ []types.HAState) (string, error) {
			return "should not reach", nil
		},
		"transformed",
	)

	err := handler(ctx)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Service unavailable")
}

func TestTransformHandler_WithData(t *testing.T) {
	t.Parallel()

	areas := AreaRegistryResult(
		MockAreaEntry("living_room", "Living Room"),
		MockAreaEntry("bedroom", "Master Bedroom"),
	)

	router := NewMessageRouter(t).
		OnSuccess("config/area_registry/list", areas)

	ctx, cleanup := NewTestContext(t, router)
	defer cleanup()

	// Capture output
	_, restoreOutput := CaptureOutput()
	defer restoreOutput()

	handler := TransformHandler[[]types.AreaEntry, []string](
		"config/area_registry/list",
		func(_ *Context) map[string]any {
			return map[string]any{"include_disabled": false}
		},
		func(areas []types.AreaEntry) ([]string, error) {
			var names []string
			for _, a := range areas {
				names = append(names, a.Name)
			}
			return names, nil
		},
		"area_names",
	)

	err := handler(ctx)
	require.NoError(t, err)
}
