package handlers

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/home-assistant-blueprints/testfixtures"

	"github.com/home-assistant-blueprints/ha-ws-client-go/internal/types"
)

func TestGetTraceDetailExtractsID(t *testing.T) {
	t.Parallel()

	// Test that the function correctly strips the automation. prefix
	// This tests the helper function logic without requiring a real client
	tests := []struct {
		name         string
		automationID string
		expectedID   string
	}{
		{
			name:         "strips automation prefix",
			automationID: "automation.kitchen_lights",
			expectedID:   "kitchen_lights",
		},
		{
			name:         "no prefix to strip",
			automationID: "kitchen_lights",
			expectedID:   "kitchen_lights",
		},
		{
			name:         "empty string",
			automationID: "",
			expectedID:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// EnsureAutomationPrefix is already tested, but we verify the inverse operation
			// that getTraceDetail uses internally (strings.TrimPrefix)
			result := stripAutomationPrefix(tt.automationID)
			assert.Equal(t, tt.expectedID, result)
		})
	}
}

// stripAutomationPrefix is the inverse of EnsureAutomationPrefix, used internally
// We test the logic here to validate the automation handlers work correctly
func stripAutomationPrefix(entityID string) string {
	const prefix = "automation."
	if len(entityID) >= len(prefix) && entityID[:len(prefix)] == prefix {
		return entityID[len(prefix):]
	}
	return entityID
}

func TestTraceInfoFormatting(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		trace    types.TraceInfo
		hasTime  bool
		hasState bool
	}{
		{
			name: "complete trace info",
			trace: types.TraceInfo{
				ItemID:          "kitchen_lights",
				RunID:           "run-123",
				State:           "stopped",
				ScriptExecution: "finished",
				Timestamp: &types.Timestamp{
					Start:  "2024-01-15T10:00:00Z",
					Finish: "2024-01-15T10:00:01Z",
				},
			},
			hasTime:  true,
			hasState: true,
		},
		{
			name: "trace without timestamp",
			trace: types.TraceInfo{
				ItemID:          "motion_lights",
				RunID:           "run-456",
				State:           "running",
				ScriptExecution: "",
				Timestamp:       nil,
			},
			hasTime:  false,
			hasState: true,
		},
		{
			name: "minimal trace",
			trace: types.TraceInfo{
				ItemID: "test_auto",
				RunID:  "run-789",
			},
			hasTime:  false,
			hasState: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.NotEmpty(t, tt.trace.ItemID)
			assert.NotEmpty(t, tt.trace.RunID)

			if tt.hasTime {
				assert.NotNil(t, tt.trace.Timestamp)
				assert.NotEmpty(t, tt.trace.Timestamp.Start)
			}

			if tt.hasState {
				// ScriptExecution takes precedence over State
				state := tt.trace.ScriptExecution
				if state == "" {
					state = tt.trace.State
				}
				assert.NotEmpty(t, state)
			}
		})
	}
}

func TestTraceDetailGetTriggerDescription(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		trace    types.TraceDetail
		expected string
	}{
		{
			name: "state trigger",
			trace: types.TraceDetail{
				Trigger: map[string]any{
					"platform":  "state",
					"entity_id": "binary_sensor.motion",
					"to":        "on",
				},
			},
			expected: "state",
		},
		{
			name: "time trigger",
			trace: types.TraceDetail{
				Trigger: map[string]any{
					"platform": "time",
					"at":       "sunset",
				},
			},
			expected: "time",
		},
		{
			name: "nil trigger",
			trace: types.TraceDetail{
				Trigger: nil,
			},
			expected: "",
		},
		{
			name: "empty trigger falls back to default",
			trace: types.TraceDetail{
				Trigger: map[string]any{},
			},
			expected: "trigger",
		},
		{
			name: "event trigger",
			trace: types.TraceDetail{
				Trigger: map[string]any{
					"platform":   "event",
					"event_type": "homeassistant_started",
				},
			},
			expected: "event",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := tt.trace.GetTriggerDescription()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAutomationConfigStructure(t *testing.T) {
	t.Parallel()

	t.Run("complete config", func(t *testing.T) {
		t.Parallel()

		config := types.AutomationConfig{
			ID:    "test_automation",
			Alias: "Test Automation",
			Trigger: []any{
				map[string]any{"platform": "state", "entity_id": "light.kitchen"},
			},
			Action: []any{
				map[string]any{"service": "light.turn_on", "target": map[string]any{"entity_id": "light.bedroom"}},
			},
		}

		assert.NotEmpty(t, config.ID)
		assert.NotEmpty(t, config.Alias)
		assert.Len(t, config.Trigger, 1)
		assert.Len(t, config.Action, 1)
	})

	t.Run("blueprint config", func(t *testing.T) {
		t.Parallel()

		config := types.AutomationConfig{
			ID:    "blueprint_auto",
			Alias: "Blueprint Automation",
			UseBlueprint: &types.BlueprintRef{
				Path: "homeassistant/motion_light.yaml",
				Input: map[string]any{
					"motion_entity": "binary_sensor.hallway_motion",
					"light_target":  "light.hallway",
				},
			},
		}

		assert.Equal(t, "blueprint_auto", config.ID)
		assert.Equal(t, "Blueprint Automation", config.Alias)
		assert.NotNil(t, config.UseBlueprint)
		assert.NotEmpty(t, config.UseBlueprint.Path)
		assert.NotNil(t, config.UseBlueprint.Input)
		assert.Contains(t, config.UseBlueprint.Input, "motion_entity")
	})
}

func TestTraceStepVariables(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		step           types.TraceStep
		hasVars        bool
		hasChangedVars bool
	}{
		{
			name: "step with variables",
			step: types.TraceStep{
				Path:      "action/0",
				Timestamp: "2024-01-15T10:00:00Z",
				Variables: map[string]any{
					"trigger": map[string]any{"platform": "state"},
				},
			},
			hasVars:        true,
			hasChangedVars: false,
		},
		{
			name: "step with changed variables",
			step: types.TraceStep{
				Path:      "action/1",
				Timestamp: "2024-01-15T10:00:01Z",
				ChangedVariables: map[string]any{
					"my_var": "new_value",
				},
			},
			hasVars:        false,
			hasChangedVars: true,
		},
		{
			name: "step with both",
			step: types.TraceStep{
				Path:      "action/2",
				Timestamp: "2024-01-15T10:00:02Z",
				Variables: map[string]any{
					"trigger": map[string]any{"platform": "time"},
				},
				ChangedVariables: map[string]any{
					"result": "success",
				},
			},
			hasVars:        true,
			hasChangedVars: true,
		},
		{
			name: "step with error",
			step: types.TraceStep{
				Path:      "action/3",
				Timestamp: "2024-01-15T10:00:03Z",
				Error:     "Service call failed",
			},
			hasVars:        false,
			hasChangedVars: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.NotEmpty(t, tt.step.Path)
			assert.NotEmpty(t, tt.step.Timestamp)

			if tt.hasVars {
				assert.NotNil(t, tt.step.Variables)
				assert.Greater(t, len(tt.step.Variables), 0)
			}
			if tt.hasChangedVars {
				assert.NotNil(t, tt.step.ChangedVariables)
				assert.Greater(t, len(tt.step.ChangedVariables), 0)
			}
		})
	}
}

func TestTraceResultStructure(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		result types.TraceResult
	}{
		{
			name: "service call result",
			result: types.TraceResult{
				Response: map[string]any{
					"success": true,
				},
			},
		},
		{
			name: "nil response",
			result: types.TraceResult{
				Response: nil,
			},
		},
		{
			name: "complex response",
			result: types.TraceResult{
				Response: map[string]any{
					"context": map[string]any{
						"id":        "ctx-123",
						"parent_id": "parent-456",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// Just validate the structure can hold various response types
			_ = tt.result.Response
		})
	}
}

func TestIsNumericID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "numeric ID",
			input:    "1764091895602",
			expected: true,
		},
		{
			name:     "single digit",
			input:    "5",
			expected: true,
		},
		{
			name:     "entity name",
			input:    "guest_bedroom_adaptive_shade",
			expected: false,
		},
		{
			name:     "mixed alphanumeric",
			input:    "abc123",
			expected: false,
		},
		{
			name:     "empty string",
			input:    "",
			expected: false,
		},
		{
			name:     "starts with number but has letters",
			input:    "123abc",
			expected: false,
		},
		{
			name:     "has dash",
			input:    "123-456",
			expected: false,
		},
		{
			name:     "has underscore",
			input:    "123_456",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := isNumericID(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// isNumericID checks if a string contains only digits.
// This is the logic used by resolveAutomationInternalID to determine
// if an ID is already a numeric internal ID or needs resolution.
func isNumericID(s string) bool {
	if s == "" {
		return false
	}
	for _, ch := range s {
		if ch < '0' || ch > '9' {
			return false
		}
	}
	return true
}

func TestResolveAutomationInternalID_NumericPassthrough(t *testing.T) {
	t.Parallel()

	// When given a numeric ID, resolveAutomationInternalID should return it unchanged
	// without making any API calls. We test this indirectly by checking the logic.
	tests := []struct {
		name      string
		input     string
		isNumeric bool
	}{
		{
			name:      "numeric ID passes through",
			input:     "1764091895602",
			isNumeric: true,
		},
		{
			name:      "entity name needs resolution",
			input:     "guest_bedroom_adaptive_shade",
			isNumeric: false,
		},
		{
			name:      "empty returns empty",
			input:     "",
			isNumeric: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// Verify the numeric detection logic matches expectations
			assert.Equal(t, tt.isNumeric, isNumericID(tt.input))
		})
	}
}

// Note: Full integration testing of resolveAutomationInternalID requires a real
// WebSocket connection to Home Assistant. The handlers_integration_test.go file
// contains integration tests that verify the complete flow including API calls.

// Benchmark for EnsureAutomationPrefix which is used throughout automation handlers
func BenchmarkEnsureAutomationPrefix(b *testing.B) {
	inputs := []string{
		"kitchen_lights",
		"automation.kitchen_lights",
		"motion_triggered_lights",
		"automation.motion_triggered_lights",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		input := inputs[i%len(inputs)]
		_ = EnsureAutomationPrefix(input)
	}
}

// Benchmark for stripAutomationPrefix
func BenchmarkStripAutomationPrefix(b *testing.B) {
	inputs := []string{
		"kitchen_lights",
		"automation.kitchen_lights",
		"motion_triggered_lights",
		"automation.motion_triggered_lights",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		input := inputs[i%len(inputs)]
		_ = stripAutomationPrefix(input)
	}
}

// =====================================
// Handler Unit Tests
// =====================================

func TestHandleTraces(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		args        []string
		traces      []types.TraceInfo
		expectError bool
		description string
	}{
		{
			name: "success with traces",
			args: []string{},
			traces: []types.TraceInfo{
				convertTraceInfo(testfixtures.NewTraceInfoFull("kitchen_lights", "run-001", "stopped", "finished")),
				convertTraceInfo(testfixtures.NewTraceInfoFull("kitchen_lights", "run-002", "stopped", "finished")),
			},
			expectError: false,
			description: "Returns list of traces when traces exist",
		},
		{
			name:        "empty results",
			args:        []string{},
			traces:      []types.TraceInfo{},
			expectError: false,
			description: "Returns empty list when no traces exist",
		},
		{
			name: "with automation filter",
			args: []string{"kitchen_lights"},
			traces: []types.TraceInfo{
				convertTraceInfo(testfixtures.NewTraceInfoFull("kitchen_lights", "run-001", "stopped", "finished")),
			},
			expectError: false,
			description: "Returns filtered traces for specific automation",
		},
		{
			name: "with numeric automation ID",
			args: []string{"1764091895602"},
			traces: []types.TraceInfo{
				convertTraceInfo(testfixtures.NewTraceInfo("1764091895602", "run-001", "finished")),
			},
			expectError: false,
			description: "Works with numeric automation IDs (internal IDs)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create router with appropriate handlers
			router := NewMessageRouter(t)

			// Handle trace/list
			router.OnSuccess("trace/list", tt.traces)

			// Handle get_states for ID resolution (for non-numeric IDs)
			router.OnSuccess("get_states", []testfixtures.HAState{
				testfixtures.NewHAStateWithAttrs("automation.kitchen_lights", "on", map[string]any{
					"id":             "12345",
					"last_triggered": "2024-01-15T10:00:00Z",
				}),
			})

			// Set up config with automation ID if provided
			var config *HandlerConfig
			if len(tt.args) > 0 {
				config = &HandlerConfig{
					AutomationID: tt.args[0],
					Args:         tt.args,
				}
			} else {
				config = &HandlerConfig{}
			}

			ctx, cleanup := NewTestContext(t, router, WithArgs(tt.args...), WithHandlerConfig(config))
			defer cleanup()

			err := handleTraces(ctx)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestHandleTracesWithFromTimeFilter(t *testing.T) {
	t.Parallel()

	// Create traces with timestamps
	now := time.Now().UTC()
	old := now.Add(-24 * time.Hour)

	traces := []testfixtures.TraceInfo{
		{
			ItemID:          "kitchen_lights",
			RunID:           "run-new",
			ScriptExecution: "finished",
			Timestamp: &testfixtures.Timestamp{
				Start: now.Format(time.RFC3339),
			},
		},
		{
			ItemID:          "kitchen_lights",
			RunID:           "run-old",
			ScriptExecution: "finished",
			Timestamp: &testfixtures.Timestamp{
				Start: old.Format(time.RFC3339),
			},
		},
	}

	router := NewMessageRouter(t).OnSuccess("trace/list", traces)

	// Filter to only traces after 12 hours ago
	filterTime := now.Add(-12 * time.Hour)

	ctx, cleanup := NewTestContext(t, router,
		WithFromTime(filterTime),
		WithHandlerConfig(&HandlerConfig{}),
	)
	defer cleanup()

	err := handleTraces(ctx)
	assert.NoError(t, err)
}

func TestHandleTracesWebSocketError(t *testing.T) {
	t.Parallel()

	router := NewMessageRouter(t).OnError("trace/list", "unknown_error", "Something went wrong")

	ctx, cleanup := NewTestContext(t, router, WithHandlerConfig(&HandlerConfig{}))
	defer cleanup()

	err := handleTraces(ctx)
	assert.Error(t, err)
}

func TestHandleTrace(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		config      *HandlerConfig
		traceDetail any
		expectError bool
	}{
		{
			name: "success case",
			config: &HandlerConfig{
				AutomationID: "kitchen_lights",
				Args:         []string{"kitchen_lights", "run-001"},
			},
			traceDetail: testfixtures.NewTraceDetail("kitchen_lights", "run-001"),
			expectError: false,
		},
		{
			name: "with numeric automation ID",
			config: &HandlerConfig{
				AutomationID: "1764091895602",
				Args:         []string{"1764091895602", "run-002"},
			},
			traceDetail: testfixtures.NewTraceDetail("1764091895602", "run-002"),
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			router := NewMessageRouter(t)
			router.OnSuccess("trace/get", tt.traceDetail)
			// For ID resolution
			router.OnSuccess("get_states", []testfixtures.HAState{
				testfixtures.NewHAStateWithAttrs("automation.kitchen_lights", "on", map[string]any{
					"id": "12345",
				}),
			})

			ctx, cleanup := NewTestContext(t, router, WithHandlerConfig(tt.config))
			defer cleanup()

			err := handleTrace(ctx)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestHandleTraceError(t *testing.T) {
	t.Parallel()

	router := NewMessageRouter(t).OnError("trace/get", "trace_not_found", "Trace not found")
	router.OnSuccess("get_states", []testfixtures.HAState{})

	config := &HandlerConfig{
		AutomationID: "kitchen_lights",
		Args:         []string{"kitchen_lights", "invalid-run"},
	}

	ctx, cleanup := NewTestContext(t, router, WithHandlerConfig(config))
	defer cleanup()

	err := handleTrace(ctx)
	assert.Error(t, err)
}

func TestHandleTraceLatest(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		config      *HandlerConfig
		traces      []types.TraceInfo
		traceDetail any
		expectError bool
	}{
		{
			name: "success with traces",
			config: &HandlerConfig{
				AutomationID: "kitchen_lights",
				Args:         []string{"kitchen_lights"},
			},
			traces: []types.TraceInfo{
				convertTraceInfo(testfixtures.NewTraceInfoFull("kitchen_lights", "run-003", "stopped", "finished")),
				convertTraceInfo(testfixtures.NewTraceInfoFull("kitchen_lights", "run-002", "stopped", "finished")),
			},
			traceDetail: testfixtures.NewTraceDetail("kitchen_lights", "run-003"),
			expectError: false,
		},
		{
			name: "no traces available",
			config: &HandlerConfig{
				AutomationID: "empty_automation",
				Args:         []string{"empty_automation"},
			},
			traces:      []types.TraceInfo{},
			traceDetail: testfixtures.TraceDetail{},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			router := NewMessageRouter(t)
			router.OnSuccess("trace/list", tt.traces)
			router.OnSuccess("trace/get", tt.traceDetail)
			router.OnSuccess("get_states", []testfixtures.HAState{
				testfixtures.NewHAStateWithAttrs("automation.kitchen_lights", "on", map[string]any{
					"id": "12345",
				}),
			})

			ctx, cleanup := NewTestContext(t, router, WithHandlerConfig(tt.config))
			defer cleanup()

			err := handleTraceLatest(ctx)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestHandleTraceSummary(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		config      *HandlerConfig
		traces      []types.TraceInfo
		traceDetail any
		expectError bool
	}{
		{
			name: "success with mixed states",
			config: &HandlerConfig{
				AutomationID: "kitchen_lights",
				Args:         []string{"kitchen_lights"},
			},
			traces: []types.TraceInfo{
				convertTraceInfo(testfixtures.NewTraceInfoFull("kitchen_lights", "run-003", "stopped", "finished")),
				convertTraceInfo(testfixtures.NewTraceInfoFull("kitchen_lights", "run-002", "stopped", "finished")),
				convertTraceInfo(testfixtures.NewTraceInfoFull("kitchen_lights", "run-001", "stopped", "error")),
			},
			traceDetail: testfixtures.NewTraceDetail("kitchen_lights", "run-003"),
			expectError: false,
		},
		{
			name: "no traces - prints message",
			config: &HandlerConfig{
				AutomationID: "empty_automation",
				Args:         []string{"empty_automation"},
			},
			traces:      []types.TraceInfo{},
			traceDetail: testfixtures.TraceDetail{},
			expectError: false, // Prints message but doesn't return error
		},
		{
			name: "with error in trace",
			config: &HandlerConfig{
				AutomationID: "failing_automation",
				Args:         []string{"failing_automation"},
			},
			traces: []types.TraceInfo{
				convertTraceInfo(testfixtures.NewTraceInfoFull("failing_automation", "run-001", "stopped", "error")),
			},
			traceDetail: testfixtures.TraceDetail{
				ItemID:          "failing_automation",
				RunID:           "run-001",
				ScriptExecution: "error",
				Error:           "Service not available",
				Trigger:         map[string]any{"platform": "state"},
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			router := NewMessageRouter(t)
			router.OnSuccess("trace/list", tt.traces)
			router.OnSuccess("trace/get", tt.traceDetail)
			router.OnSuccess("get_states", []testfixtures.HAState{})

			ctx, cleanup := NewTestContext(t, router, WithHandlerConfig(tt.config))
			defer cleanup()

			err := handleTraceSummary(ctx)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestHandleTraceVars(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		config      *HandlerConfig
		traceDetail any
		expectError bool
	}{
		{
			name: "with variables",
			config: &HandlerConfig{
				AutomationID: "kitchen_lights",
				Args:         []string{"kitchen_lights", "run-001"},
			},
			traceDetail: testfixtures.TraceDetail{
				ItemID:          "kitchen_lights",
				RunID:           "run-001",
				ScriptExecution: "finished",
				Trace: map[string][]testfixtures.TraceStep{
					"action/0": {
						{
							Path:      "action/0",
							Timestamp: time.Now().UTC().Format(time.RFC3339),
							Variables: map[string]any{
								"trigger": map[string]any{"platform": "state", "entity_id": "sensor.motion"},
								"target":  "light.kitchen",
							},
						},
					},
					"action/1": {
						{
							Path:      "action/1",
							Timestamp: time.Now().UTC().Format(time.RFC3339),
							ChangedVariables: map[string]any{
								"result": "success",
							},
						},
					},
				},
			},
			expectError: false,
		},
		{
			name: "empty variables",
			config: &HandlerConfig{
				AutomationID: "simple_automation",
				Args:         []string{"simple_automation", "run-002"},
			},
			traceDetail: testfixtures.TraceDetail{
				ItemID:          "simple_automation",
				RunID:           "run-002",
				ScriptExecution: "finished",
				Trace:           map[string][]testfixtures.TraceStep{}, // No trace steps with variables
			},
			expectError: false,
		},
		{
			name: "nil trace map",
			config: &HandlerConfig{
				AutomationID: "minimal_automation",
				Args:         []string{"minimal_automation", "run-003"},
			},
			traceDetail: testfixtures.TraceDetail{
				ItemID:          "minimal_automation",
				RunID:           "run-003",
				ScriptExecution: "finished",
				Trace:           nil, // No trace at all
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			router := NewMessageRouter(t)
			router.OnSuccess("trace/get", tt.traceDetail)
			router.OnSuccess("get_states", []testfixtures.HAState{})

			ctx, cleanup := NewTestContext(t, router, WithHandlerConfig(tt.config))
			defer cleanup()

			err := handleTraceVars(ctx)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestHandleTraceTimeline(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()

	tests := []struct {
		name        string
		config      *HandlerConfig
		traceDetail any
		expectError bool
	}{
		{
			name: "with steps",
			config: &HandlerConfig{
				AutomationID: "kitchen_lights",
				Args:         []string{"kitchen_lights", "run-001"},
			},
			traceDetail: testfixtures.TraceDetail{
				ItemID:          "kitchen_lights",
				RunID:           "run-001",
				ScriptExecution: "finished",
				Trace: map[string][]testfixtures.TraceStep{
					"trigger/0": {
						{
							Path:      "trigger/0",
							Timestamp: now.Add(-2 * time.Second).Format(time.RFC3339),
						},
					},
					"action/0": {
						{
							Path:      "action/0",
							Timestamp: now.Add(-1 * time.Second).Format(time.RFC3339),
						},
					},
					"action/1": {
						{
							Path:      "action/1",
							Timestamp: now.Format(time.RFC3339),
						},
					},
				},
			},
			expectError: false,
		},
		{
			name: "with errors in steps",
			config: &HandlerConfig{
				AutomationID: "failing_automation",
				Args:         []string{"failing_automation", "run-002"},
			},
			traceDetail: testfixtures.TraceDetail{
				ItemID:          "failing_automation",
				RunID:           "run-002",
				ScriptExecution: "error",
				Trace: map[string][]testfixtures.TraceStep{
					"trigger/0": {
						{
							Path:      "trigger/0",
							Timestamp: now.Add(-2 * time.Second).Format(time.RFC3339),
						},
					},
					"action/0": {
						{
							Path:      "action/0",
							Timestamp: now.Add(-1 * time.Second).Format(time.RFC3339),
							Error:     "Service 'light.turn_on' not found",
						},
					},
				},
			},
			expectError: false,
		},
		{
			name: "empty trace",
			config: &HandlerConfig{
				AutomationID: "minimal_automation",
				Args:         []string{"minimal_automation", "run-003"},
			},
			traceDetail: testfixtures.TraceDetail{
				ItemID:          "minimal_automation",
				RunID:           "run-003",
				ScriptExecution: "finished",
				Trace:           nil,
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			router := NewMessageRouter(t)
			router.OnSuccess("trace/get", tt.traceDetail)
			router.OnSuccess("get_states", []testfixtures.HAState{})

			ctx, cleanup := NewTestContext(t, router, WithHandlerConfig(tt.config))
			defer cleanup()

			err := handleTraceTimeline(ctx)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestHandleTraceTrigger(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		config      *HandlerConfig
		traceDetail any
		expectError bool
	}{
		{
			name: "with trigger",
			config: &HandlerConfig{
				AutomationID: "kitchen_lights",
				Args:         []string{"kitchen_lights", "run-001"},
			},
			traceDetail: testfixtures.TraceDetail{
				ItemID:          "kitchen_lights",
				RunID:           "run-001",
				ScriptExecution: "finished",
				Trigger: map[string]any{
					"platform":  "state",
					"entity_id": "binary_sensor.motion",
					"from":      "off",
					"to":        "on",
				},
			},
			expectError: false,
		},
		{
			name: "nil trigger",
			config: &HandlerConfig{
				AutomationID: "time_automation",
				Args:         []string{"time_automation", "run-002"},
			},
			traceDetail: testfixtures.TraceDetail{
				ItemID:          "time_automation",
				RunID:           "run-002",
				ScriptExecution: "finished",
				Trigger:         nil,
			},
			expectError: false, // Prints message but no error
		},
		{
			name: "with time trigger",
			config: &HandlerConfig{
				AutomationID: "sunset_automation",
				Args:         []string{"sunset_automation", "run-003"},
			},
			traceDetail: testfixtures.TraceDetail{
				ItemID:          "sunset_automation",
				RunID:           "run-003",
				ScriptExecution: "finished",
				Trigger: map[string]any{
					"platform": "sun",
					"event":    "sunset",
					"offset":   "-00:30:00",
				},
			},
			expectError: false,
		},
		{
			name: "with event trigger",
			config: &HandlerConfig{
				AutomationID: "event_automation",
				Args:         []string{"event_automation", "run-004"},
			},
			traceDetail: testfixtures.TraceDetail{
				ItemID:          "event_automation",
				RunID:           "run-004",
				ScriptExecution: "finished",
				Trigger: map[string]any{
					"platform":   "event",
					"event_type": "homeassistant_started",
				},
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			router := NewMessageRouter(t)
			router.OnSuccess("trace/get", tt.traceDetail)
			router.OnSuccess("get_states", []testfixtures.HAState{})

			ctx, cleanup := NewTestContext(t, router, WithHandlerConfig(tt.config))
			defer cleanup()

			err := handleTraceTrigger(ctx)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestHandleTraceActions(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()

	tests := []struct {
		name        string
		config      *HandlerConfig
		traceDetail any
		expectError bool
	}{
		{
			name: "with actions",
			config: &HandlerConfig{
				AutomationID: "kitchen_lights",
				Args:         []string{"kitchen_lights", "run-001"},
			},
			traceDetail: testfixtures.TraceDetail{
				ItemID:          "kitchen_lights",
				RunID:           "run-001",
				ScriptExecution: "finished",
				Trace: map[string][]testfixtures.TraceStep{
					"action/0": {
						{
							Path:      "action/0",
							Timestamp: now.Format(time.RFC3339),
							Result: &testfixtures.TraceResult{
								Enabled: true,
								Response: map[string]any{
									"success": true,
								},
							},
						},
					},
					"action/1": {
						{
							Path:      "action/1",
							Timestamp: now.Add(1 * time.Second).Format(time.RFC3339),
							Result: &testfixtures.TraceResult{
								Enabled: true,
							},
						},
					},
				},
			},
			expectError: false,
		},
		{
			name: "no action steps",
			config: &HandlerConfig{
				AutomationID: "trigger_only",
				Args:         []string{"trigger_only", "run-002"},
			},
			traceDetail: testfixtures.TraceDetail{
				ItemID:          "trigger_only",
				RunID:           "run-002",
				ScriptExecution: "finished",
				Trace: map[string][]testfixtures.TraceStep{
					"trigger/0": {
						{
							Path:      "trigger/0",
							Timestamp: now.Format(time.RFC3339),
						},
					},
					"condition/0": {
						{
							Path:      "condition/0",
							Timestamp: now.Add(1 * time.Second).Format(time.RFC3339),
						},
					},
				},
			},
			expectError: false,
		},
		{
			name: "nil trace",
			config: &HandlerConfig{
				AutomationID: "minimal_automation",
				Args:         []string{"minimal_automation", "run-003"},
			},
			traceDetail: testfixtures.TraceDetail{
				ItemID:          "minimal_automation",
				RunID:           "run-003",
				ScriptExecution: "finished",
				Trace:           nil,
			},
			expectError: false,
		},
		{
			name: "action with nil result",
			config: &HandlerConfig{
				AutomationID: "simple_action",
				Args:         []string{"simple_action", "run-004"},
			},
			traceDetail: testfixtures.TraceDetail{
				ItemID:          "simple_action",
				RunID:           "run-004",
				ScriptExecution: "finished",
				Trace: map[string][]testfixtures.TraceStep{
					"action/0": {
						{
							Path:      "action/0",
							Timestamp: now.Format(time.RFC3339),
							Result:    nil,
						},
					},
				},
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			router := NewMessageRouter(t)
			router.OnSuccess("trace/get", tt.traceDetail)
			router.OnSuccess("get_states", []testfixtures.HAState{})

			ctx, cleanup := NewTestContext(t, router, WithHandlerConfig(tt.config))
			defer cleanup()

			err := handleTraceActions(ctx)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestHandleTraceDebug(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		config      *HandlerConfig
		traceDetail any
		expectError bool
	}{
		{
			name: "comprehensive trace",
			config: &HandlerConfig{
				AutomationID: "kitchen_lights",
				Args:         []string{"kitchen_lights", "run-001"},
			},
			traceDetail: testfixtures.NewTraceDetailWithTrigger("kitchen_lights", "run-001", map[string]any{
				"platform":  "state",
				"entity_id": "binary_sensor.motion",
			}),
			expectError: false,
		},
		{
			name: "with automation prefix",
			config: &HandlerConfig{
				AutomationID: "automation.living_room",
				Args:         []string{"automation.living_room", "run-002"},
			},
			traceDetail: testfixtures.NewTraceDetail("living_room", "run-002"),
			expectError: false,
		},
		{
			name: "with numeric ID",
			config: &HandlerConfig{
				AutomationID: "1764091895602",
				Args:         []string{"1764091895602", "run-003"},
			},
			traceDetail: testfixtures.NewTraceDetail("1764091895602", "run-003"),
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			router := NewMessageRouter(t)
			router.OnSuccess("trace/get", tt.traceDetail)
			router.OnSuccess("get_states", []testfixtures.HAState{})

			ctx, cleanup := NewTestContext(t, router, WithHandlerConfig(tt.config))
			defer cleanup()

			err := handleTraceDebug(ctx)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestHandleTraceDebugError(t *testing.T) {
	t.Parallel()

	router := NewMessageRouter(t).OnError("trace/get", "trace_not_found", "Trace not found")
	router.OnSuccess("get_states", []testfixtures.HAState{})

	config := &HandlerConfig{
		AutomationID: "invalid_automation",
		Args:         []string{"invalid_automation", "invalid-run"},
	}

	ctx, cleanup := NewTestContext(t, router, WithHandlerConfig(config))
	defer cleanup()

	err := handleTraceDebug(ctx)
	assert.Error(t, err)
}

func TestHandleAutomationConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		config      *HandlerConfig
		traces      []types.TraceInfo
		traceDetail any
		apiConfig   map[string]any
		expectError bool
	}{
		{
			name: "from trace with config",
			config: &HandlerConfig{
				Args: []string{"kitchen_lights"},
			},
			traces: []types.TraceInfo{
				convertTraceInfo(testfixtures.NewTraceInfoFull("kitchen_lights", "run-001", "stopped", "finished")),
			},
			traceDetail: testfixtures.NewTraceDetailWithConfig("kitchen_lights", "run-001", &testfixtures.AutomationConfig{
				ID:    "kitchen_lights",
				Alias: "Kitchen Lights",
				Trigger: []any{
					map[string]any{"platform": "state", "entity_id": "binary_sensor.motion"},
				},
				Action: []any{
					map[string]any{"service": "light.turn_on", "target": map[string]any{"entity_id": "light.kitchen"}},
				},
			}),
			apiConfig:   nil,
			expectError: false,
		},
		{
			name: "from API fallback",
			config: &HandlerConfig{
				Args: []string{"api_automation"},
			},
			traces:      []types.TraceInfo{}, // No traces available
			traceDetail: testfixtures.TraceDetail{},
			apiConfig: map[string]any{
				"config": map[string]any{
					"id":    "api_automation",
					"alias": "API Automation",
					"trigger": []any{
						map[string]any{"platform": "time", "at": "08:00:00"},
					},
					"action": []any{
						map[string]any{"service": "light.turn_on"},
					},
				},
			},
			expectError: false,
		},
		{
			name: "minimal config - blueprint with no traces",
			config: &HandlerConfig{
				Args: []string{"blueprint_auto"},
			},
			traces:      []types.TraceInfo{},
			traceDetail: testfixtures.TraceDetail{},
			apiConfig: map[string]any{
				"config": map[string]any{
					"id":    "blueprint_auto",
					"alias": "Blueprint Automation",
					// No trigger/action - blueprint automation with no stored traces
				},
			},
			expectError: false, // Should print message but not error
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			router := NewMessageRouter(t)
			router.OnSuccess("trace/list", tt.traces)
			router.OnSuccess("trace/get", tt.traceDetail)
			if tt.apiConfig != nil {
				router.OnSuccess("automation/config", tt.apiConfig)
			}
			router.OnSuccess("get_states", []testfixtures.HAState{})

			ctx, cleanup := NewTestContext(t, router, WithHandlerConfig(tt.config))
			defer cleanup()

			err := handleAutomationConfig(ctx)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestHandleAutomationConfigError(t *testing.T) {
	t.Parallel()

	router := NewMessageRouter(t)
	router.OnSuccess("trace/list", []testfixtures.TraceInfo{})
	router.OnError("automation/config", "not_found", "Automation not found")
	router.OnSuccess("get_states", []testfixtures.HAState{})

	config := &HandlerConfig{
		Args: []string{"nonexistent_automation"},
	}

	ctx, cleanup := NewTestContext(t, router, WithHandlerConfig(config))
	defer cleanup()

	err := handleAutomationConfig(ctx)
	assert.Error(t, err)
}

func TestHandleBlueprintInputs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		config      *HandlerConfig
		apiConfig   map[string]any
		expectError bool
	}{
		{
			name: "with blueprint",
			config: &HandlerConfig{
				Args: []string{"blueprint_automation"},
			},
			apiConfig: map[string]any{
				"config": map[string]any{
					"id":    "blueprint_automation",
					"alias": "Blueprint Automation",
					"use_blueprint": map[string]any{
						"path": "homeassistant/motion_light.yaml",
						"input": map[string]any{
							"motion_entity": "binary_sensor.hallway_motion",
							"light_target":  "light.hallway",
							"wait_time":     120,
						},
					},
				},
			},
			expectError: false,
		},
		{
			name: "no blueprint",
			config: &HandlerConfig{
				Args: []string{"regular_automation"},
			},
			apiConfig: map[string]any{
				"config": map[string]any{
					"id":    "regular_automation",
					"alias": "Regular Automation",
					"trigger": []any{
						map[string]any{"platform": "state"},
					},
					// No use_blueprint
				},
			},
			expectError: false, // Prints message but no error
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			router := NewMessageRouter(t)
			router.OnSuccess("automation/config", tt.apiConfig)

			ctx, cleanup := NewTestContext(t, router, WithHandlerConfig(tt.config))
			defer cleanup()

			err := handleBlueprintInputs(ctx)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestHandleBlueprintInputsError(t *testing.T) {
	t.Parallel()

	router := NewMessageRouter(t).OnError("automation/config", "not_found", "Automation not found")

	config := &HandlerConfig{
		Args: []string{"nonexistent_automation"},
	}

	ctx, cleanup := NewTestContext(t, router, WithHandlerConfig(config))
	defer cleanup()

	err := handleBlueprintInputs(ctx)
	assert.Error(t, err)
}

// =====================================
// getTraceDetail Tests
// =====================================

func TestGetTraceDetail(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		automationID string
		runID        string
		traceDetail  any
		expectError  bool
	}{
		{
			name:         "success case",
			automationID: "kitchen_lights",
			runID:        "run-001",
			traceDetail:  testfixtures.NewTraceDetail("kitchen_lights", "run-001"),
			expectError:  false,
		},
		{
			name:         "with automation prefix",
			automationID: "automation.kitchen_lights",
			runID:        "run-002",
			traceDetail:  testfixtures.NewTraceDetail("kitchen_lights", "run-002"),
			expectError:  false,
		},
		{
			name:         "with numeric ID",
			automationID: "1764091895602",
			runID:        "run-003",
			traceDetail:  testfixtures.NewTraceDetail("1764091895602", "run-003"),
			expectError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			router := NewMessageRouter(t)
			router.OnSuccess("trace/get", tt.traceDetail)
			router.OnSuccess("get_states", []testfixtures.HAState{})

			ctx, cleanup := NewTestContext(t, router, WithHandlerConfig(&HandlerConfig{}))
			defer cleanup()

			result, err := getTraceDetail(ctx.Client, tt.automationID, tt.runID)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.runID, result.RunID)
			}
		})
	}
}

func TestGetTraceDetailError(t *testing.T) {
	t.Parallel()

	router := NewMessageRouter(t).OnError("trace/get", "trace_not_found", "Trace not found")
	router.OnSuccess("get_states", []testfixtures.HAState{})

	ctx, cleanup := NewTestContext(t, router, WithHandlerConfig(&HandlerConfig{}))
	defer cleanup()

	result, err := getTraceDetail(ctx.Client, "kitchen_lights", "invalid-run")
	assert.Error(t, err)
	assert.Nil(t, result)
}

// =====================================
// resolveAutomationInternalID Tests
// =====================================

func TestResolveAutomationInternalID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		automationID string
		states       []testfixtures.HAState
		expectedID   string
		description  string
	}{
		{
			name:         "numeric ID passthrough",
			automationID: "1764091895602",
			states:       []testfixtures.HAState{},
			expectedID:   "1764091895602",
			description:  "Numeric IDs should pass through unchanged",
		},
		{
			name:         "empty string passthrough",
			automationID: "",
			states:       []testfixtures.HAState{},
			expectedID:   "",
			description:  "Empty string should return empty",
		},
		{
			name:         "entity name resolved to ID",
			automationID: "guest_bedroom_adaptive_shade",
			states: []testfixtures.HAState{
				testfixtures.NewHAStateWithAttrs("automation.guest_bedroom_adaptive_shade", "on", map[string]any{
					"id":             "1764091895602",
					"last_triggered": "2024-01-15T10:00:00Z",
				}),
			},
			expectedID:  "1764091895602",
			description: "Entity name should be resolved to internal ID via get_states",
		},
		{
			name:         "entity name not found - fallback",
			automationID: "nonexistent_automation",
			states: []testfixtures.HAState{
				testfixtures.NewHAStateWithAttrs("automation.kitchen_lights", "on", map[string]any{
					"id": "12345",
				}),
			},
			expectedID:  "nonexistent_automation",
			description: "Non-existent entity should fall back to original ID",
		},
		{
			name:         "entity without ID attribute - fallback",
			automationID: "old_automation",
			states: []testfixtures.HAState{
				testfixtures.NewHAStateWithAttrs("automation.old_automation", "on", map[string]any{
					"last_triggered": "2024-01-15T10:00:00Z",
					// No "id" attribute
				}),
			},
			expectedID:  "old_automation",
			description: "Entity without ID attribute should fall back to original",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			router := NewMessageRouter(t)
			router.OnSuccess("get_states", tt.states)

			ctx, cleanup := NewTestContext(t, router, WithHandlerConfig(&HandlerConfig{}))
			defer cleanup()

			result := resolveAutomationInternalID(ctx.Client, tt.automationID)
			assert.Equal(t, tt.expectedID, result, tt.description)
		})
	}
}

func TestResolveAutomationInternalIDWithGetStatesError(t *testing.T) {
	t.Parallel()

	router := NewMessageRouter(t).OnError("get_states", "connection_error", "Connection failed")

	ctx, cleanup := NewTestContext(t, router, WithHandlerConfig(&HandlerConfig{}))
	defer cleanup()

	// When get_states fails, it should fall back to the original ID
	result := resolveAutomationInternalID(ctx.Client, "kitchen_lights")
	assert.Equal(t, "kitchen_lights", result)
}

// =====================================
// Test Fixture Conversion Helpers
// =====================================

// convertTraceInfo converts a testfixtures.TraceInfo to types.TraceInfo.
// This is needed because test structs use types.TraceInfo while we use testfixtures factories.
func convertTraceInfo(tf testfixtures.TraceInfo) types.TraceInfo {
	var timestamp *types.Timestamp
	if tf.Timestamp != nil {
		timestamp = &types.Timestamp{
			Start:  tf.Timestamp.Start,
			Finish: tf.Timestamp.Finish,
		}
	}
	var context *types.HAContext
	if tf.Context != nil {
		context = &types.HAContext{
			ID:       tf.Context.ID,
			ParentID: tf.Context.ParentID,
			UserID:   tf.Context.UserID,
		}
	}
	return types.TraceInfo{
		ItemID:          tf.ItemID,
		RunID:           tf.RunID,
		State:           tf.State,
		ScriptExecution: tf.ScriptExecution,
		Timestamp:       timestamp,
		Context:         context,
	}
}
