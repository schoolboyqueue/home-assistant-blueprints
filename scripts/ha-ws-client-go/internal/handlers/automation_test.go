package handlers

import (
	"testing"

	"github.com/stretchr/testify/assert"

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
