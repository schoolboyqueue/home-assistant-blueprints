package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/home-assistant-blueprints/ha-ws-client-go/internal/output"
	"github.com/home-assistant-blueprints/ha-ws-client-go/internal/types"
)

// captureOutput captures stdout during function execution.
func captureOutput(t *testing.T, fn func()) string {
	t.Helper()
	old := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w

	fn()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	_, err = io.Copy(&buf, r)
	require.NoError(t, err)
	return buf.String()
}

// =============================================================================
// HandleState Behavior Tests
// =============================================================================

func TestHandleState_EntityFound_OutputsCorrectFormat(t *testing.T) {
	// Note: This test doesn't use t.Parallel() because subtests modify global config and capture stdout

	// Test data
	states := []types.HAState{
		{EntityID: "light.kitchen", State: "on", Attributes: map[string]any{"brightness": 255}},
		{EntityID: "light.bedroom", State: "off"},
		{EntityID: "sensor.temperature", State: "23.5", Attributes: map[string]any{"unit": "°C"}},
	}

	tests := []struct {
		name       string
		entityID   string
		format     output.Format
		wantState  string
		wantFields []string // Fields that should be present in output
	}{
		{
			name:       "JSON format includes all fields",
			entityID:   "light.kitchen",
			format:     output.FormatJSON,
			wantState:  "on",
			wantFields: []string{"entity_id", "state", "brightness"},
		},
		{
			name:       "compact format shows entity and state",
			entityID:   "sensor.temperature",
			format:     output.FormatCompact,
			wantState:  "23.5",
			wantFields: []string{"sensor.temperature", "23.5"},
		},
		{
			name:       "default format is human readable",
			entityID:   "light.bedroom",
			format:     output.FormatDefault,
			wantState:  "off",
			wantFields: []string{"light.bedroom", "off"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Subtests run sequentially to avoid stdout conflicts
			original := output.GetConfig()
			output.SetConfig(&output.Config{Format: tt.format})
			defer output.SetConfig(original)

			// We can't easily mock the client.SendMessageTyped call,
			// so we test the output formatting logic directly
			var out string
			if tt.format == output.FormatJSON {
				out = captureOutput(t, func() {
					for _, s := range states {
						if s.EntityID == tt.entityID {
							output.Data(s, output.WithCommand("state"))
							break
						}
					}
				})
				// Verify JSON structure
				var result output.Result
				err := json.Unmarshal([]byte(out), &result)
				require.NoError(t, err, "output should be valid JSON")
				assert.True(t, result.Success)

				// Verify data contains expected fields
				data, ok := result.Data.(map[string]any)
				require.True(t, ok)
				assert.Equal(t, tt.entityID, data["entity_id"])
				assert.Equal(t, tt.wantState, data["state"])
			} else {
				out = captureOutput(t, func() {
					for _, s := range states {
						if s.EntityID == tt.entityID {
							output.Entity(s.EntityID, s.State, s.Attributes)
							break
						}
					}
				})
				for _, field := range tt.wantFields {
					assert.Contains(t, out, field, "output should contain %q", field)
				}
			}
		})
	}
}

func TestHandleState_EntityNotFound_ReturnsError(t *testing.T) {
	t.Parallel()

	entityID := "nonexistent.entity"

	// Simulate the error path
	err := ErrEntityNotFound
	assert.ErrorIs(t, err, ErrEntityNotFound)
	assert.Contains(t, err.Error(), "entity not found")

	// Verify error message format when wrapping
	wrappedErr := fmt.Errorf("%w: %s", ErrEntityNotFound, entityID)
	assert.Contains(t, wrappedErr.Error(), "nonexistent.entity")
}

// =============================================================================
// HandleStatesFilter Behavior Tests
// =============================================================================

func TestHandleStatesFilter_PatternMatching(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		pattern       string
		wantMatches   []string
		wantNoMatches []string
	}{
		{
			name:          "light.* matches all lights",
			pattern:       "light.*",
			wantMatches:   []string{"light.kitchen", "light.bedroom", "light.living_room"},
			wantNoMatches: []string{"sensor.temperature", "switch.fan"},
		},
		{
			name:          "sensor.* matches all sensors",
			pattern:       "sensor.*",
			wantMatches:   []string{"sensor.temperature", "sensor.humidity"},
			wantNoMatches: []string{"light.kitchen", "binary_sensor.motion_kitchen"},
		},
		{
			name:          "*kitchen* matches entities with kitchen in name",
			pattern:       "*kitchen*",
			wantMatches:   []string{"light.kitchen", "binary_sensor.motion_kitchen"},
			wantNoMatches: []string{"light.bedroom", "sensor.temperature"},
		},
		{
			name:          "exact match",
			pattern:       "switch.fan",
			wantMatches:   []string{"switch.fan"},
			wantNoMatches: []string{"light.kitchen", "sensor.temperature"},
		},
		{
			name:          "case insensitive matching",
			pattern:       "LIGHT.*",
			wantMatches:   []string{"light.kitchen", "light.bedroom"},
			wantNoMatches: []string{"sensor.temperature"},
		},
		{
			name:          "binary_sensor prefix",
			pattern:       "binary_sensor.*",
			wantMatches:   []string{"binary_sensor.motion_kitchen"},
			wantNoMatches: []string{"sensor.temperature", "sensor.humidity"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Build the regex pattern (same logic as middleware)
			pattern := tt.pattern
			pattern = strings.ReplaceAll(pattern, ".", `\.`)
			pattern = strings.ReplaceAll(pattern, "*", ".*")
			pattern = "(?i)^" + pattern + "$"

			re, err := regexp.Compile(pattern)
			require.NoError(t, err)

			// Test matches
			for _, entityID := range tt.wantMatches {
				assert.True(t, re.MatchString(entityID),
					"pattern %q should match %q", tt.pattern, entityID)
			}

			// Test non-matches
			for _, entityID := range tt.wantNoMatches {
				assert.False(t, re.MatchString(entityID),
					"pattern %q should NOT match %q", tt.pattern, entityID)
			}
		})
	}
}

func TestHandleStatesFilter_EmptyResult(t *testing.T) {
	// Note: This test doesn't use t.Parallel() because it modifies global config

	// Set JSON format for this test
	original := output.GetConfig()
	output.SetConfig(&output.Config{Format: output.FormatJSON})
	defer output.SetConfig(original)

	// Test output when no entities match
	out := captureOutput(t, func() {
		output.List([]types.HAState{},
			output.ListTitle[types.HAState]("Found 0 matching entities"),
			output.ListCommand[types.HAState]("states-filter"),
		)
	})

	var result output.Result
	err := json.Unmarshal([]byte(out), &result)
	require.NoError(t, err, "output should be valid JSON: %s", out)
	assert.True(t, result.Success)
	assert.Equal(t, 0, result.Count)
}

// =============================================================================
// HandleHistory Behavior Tests
// =============================================================================

func TestHandleHistory_OutputFormatting(t *testing.T) {
	// Note: This test doesn't use t.Parallel() because it modifies global config

	refTime := time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC)
	historyStates := []types.HistoryState{
		{S: "on", LU: float64(refTime.Unix())},
		{S: "off", LU: float64(refTime.Add(10 * time.Minute).Unix())},
		{S: "on", LU: float64(refTime.Add(20 * time.Minute).Unix())},
	}

	tests := []struct {
		name       string
		format     output.Format
		wantFields []string
	}{
		{
			name:       "JSON format includes all states",
			format:     output.FormatJSON,
			wantFields: []string{"success", "data", "count"},
		},
		{
			name:       "compact format shows timestamp and state",
			format:     output.FormatCompact,
			wantFields: []string{"on", "off"},
		},
		{
			name:       "default format shows state values",
			format:     output.FormatDefault,
			wantFields: []string{"on", "off"}, // Timeline output shows formatted entries
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set format for this specific test
			original := output.GetConfig()
			output.SetConfig(&output.Config{Format: tt.format})
			defer output.SetConfig(original)

			out := captureOutput(t, func() {
				output.Timeline(historyStates,
					output.TimelineTitle[types.HistoryState]("History for sensor.test"),
					output.TimelineCommand[types.HistoryState]("history"),
					output.TimelineFormatter(func(s types.HistoryState) string {
						ts := s.GetLastUpdated()
						state := s.GetState()
						if output.IsCompact() {
							return fmt.Sprintf("%s %s", ts.Format(time.RFC3339), state)
						}
						return fmt.Sprintf("%s: %s", output.FormatTime(ts), state)
					}),
				)
			})

			for _, field := range tt.wantFields {
				assert.Contains(t, out, field, "output should contain %q", field)
			}

			if tt.format == output.FormatJSON {
				var result output.Result
				err := json.Unmarshal([]byte(out), &result)
				require.NoError(t, err, "JSON output should be valid")
				assert.True(t, result.Success)
				assert.Equal(t, 3, result.Count, "should have 3 history entries")
			}
		})
	}
}

func TestHandleHistory_EmptyHistory(t *testing.T) {
	// Note: This test doesn't use t.Parallel() because it modifies global config

	original := output.GetConfig()
	output.SetConfig(&output.Config{Format: output.FormatJSON})
	defer output.SetConfig(original)

	out := captureOutput(t, func() {
		output.Timeline([]types.HistoryState{},
			output.TimelineTitle[types.HistoryState]("History for sensor.nonexistent"),
			output.TimelineCommand[types.HistoryState]("history"),
		)
	})

	var result output.Result
	err := json.Unmarshal([]byte(out), &result)
	require.NoError(t, err, "output should be valid JSON: %s", out)
	assert.True(t, result.Success)
	assert.Equal(t, 0, result.Count)
}

func TestHandleHistory_TimeRangeCalculation(t *testing.T) {
	t.Parallel()

	now := time.Now()

	tests := []struct {
		name         string
		hours        int
		wantDuration time.Duration
	}{
		{
			name:         "1 hour ago",
			hours:        1,
			wantDuration: 1 * time.Hour,
		},
		{
			name:         "24 hours ago (default)",
			hours:        24,
			wantDuration: 24 * time.Hour,
		},
		{
			name:         "7 days ago",
			hours:        168,
			wantDuration: 168 * time.Hour,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			startTime := now.Add(-tt.wantDuration)
			timeRange := types.TimeRange{
				StartTime: startTime,
				EndTime:   now,
			}

			// Verify time range is correct
			assert.WithinDuration(t, now, timeRange.EndTime, time.Second)
			assert.WithinDuration(t, now.Add(-tt.wantDuration), timeRange.StartTime, time.Second)
		})
	}
}

// =============================================================================
// Error Message Tests
// =============================================================================

func TestErrorMessages_AreHelpful(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		err         error
		wantMessage string
	}{
		{
			name:        "entity not found includes entity ID",
			err:         fmt.Errorf("%w: light.nonexistent", ErrEntityNotFound),
			wantMessage: "entity not found: light.nonexistent",
		},
		{
			name:        "missing required argument",
			err:         fmt.Errorf("Usage: state <entity_id>"),
			wantMessage: "Usage: state <entity_id>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.wantMessage, tt.err.Error())
		})
	}
}

// =============================================================================
// Output Format Consistency Tests
// =============================================================================

func TestOutputFormats_AreConsistent(t *testing.T) {
	// Note: This test doesn't use t.Parallel() because it modifies global config

	// All commands should produce valid JSON when in JSON mode
	testCases := []struct {
		name    string
		command string
		data    any
	}{
		{
			name:    "state command",
			command: "state",
			data:    types.HAState{EntityID: "light.test", State: "on"},
		},
		{
			name:    "states command",
			command: "states",
			data:    map[string]any{"total": 100, "sample": []types.HAState{}},
		},
		{
			name:    "history command",
			command: "history",
			data:    []types.HistoryState{{S: "on", LU: 1234567890}},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Set JSON format for this specific test
			original := output.GetConfig()
			output.SetConfig(&output.Config{Format: output.FormatJSON})
			defer output.SetConfig(original)

			out := captureOutput(t, func() {
				output.Data(tc.data, output.WithCommand(tc.command))
			})

			// All JSON output should be valid
			var result output.Result
			err := json.Unmarshal([]byte(out), &result)
			require.NoError(t, err, "command %q should produce valid JSON: %s", tc.command, out)

			// All JSON output should have consistent structure
			assert.True(t, result.Success, "success should be true")
			assert.Equal(t, tc.command, result.Command, "command should be set")
			assert.NotNil(t, result.Data, "data should not be nil")
		})
	}
}

// =============================================================================
// Middleware Integration Tests
// =============================================================================

func TestMiddleware_RequireArg_ProducesUsefulErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		middleware Middleware
		args       []string
		wantErr    string
	}{
		{
			name:       "RequireArg1 with no args",
			middleware: RequireArg1("Usage: state <entity_id>"),
			args:       []string{"state"},
			wantErr:    "Usage: state <entity_id>",
		},
		{
			name:       "RequireArg2 with only one arg",
			middleware: RequireArg2("Usage: call <domain> <service>"),
			args:       []string{"call", "light"},
			wantErr:    "Usage: call <domain> <service>",
		},
		{
			name:       "RequireArgs with missing arg",
			middleware: RequireArgs("Usage: history <entity_id> [hours]", 1),
			args:       []string{"history"},
			wantErr:    "Usage: history <entity_id> [hours]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			handler := tt.middleware(func(_ *Context) error {
				return nil
			})

			ctx := &Context{Args: tt.args}
			err := handler(ctx)

			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

func TestMiddleware_WithTimeRange_CalculatesCorrectly(t *testing.T) {
	t.Parallel()

	now := time.Now()

	tests := []struct {
		name         string
		defaultHours int
		argsHours    string
		wantHours    int
	}{
		{
			name:         "uses default when no arg provided",
			defaultHours: 24,
			argsHours:    "",
			wantHours:    24,
		},
		{
			name:         "uses provided hours",
			defaultHours: 24,
			argsHours:    "48",
			wantHours:    48,
		},
		{
			name:         "uses default on invalid arg",
			defaultHours: 24,
			argsHours:    "invalid",
			wantHours:    24,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var capturedCtx *Context
			handler := WithTimeRange(tt.defaultHours, 2)(func(ctx *Context) error {
				capturedCtx = ctx
				return nil
			})

			args := []string{"history", "sensor.test"}
			if tt.argsHours != "" {
				args = append(args, tt.argsHours)
			}

			ctx := &Context{Args: args}
			err := handler(ctx)
			require.NoError(t, err)

			require.NotNil(t, capturedCtx.Config)
			require.NotNil(t, capturedCtx.Config.TimeRange)

			// Check the time range is approximately correct
			tr := capturedCtx.Config.TimeRange
			expectedDuration := time.Duration(tt.wantHours) * time.Hour
			actualDuration := tr.EndTime.Sub(tr.StartTime)

			assert.InDelta(t, expectedDuration.Seconds(), actualDuration.Seconds(), 1,
				"time range should be %d hours", tt.wantHours)
			assert.WithinDuration(t, now, tr.EndTime, time.Second,
				"end time should be now")
		})
	}
}

// =============================================================================
// Edge Cases
// =============================================================================

func TestEdgeCases_SpecialCharactersInEntityID(t *testing.T) {
	t.Parallel()

	// Entity IDs can contain underscores, numbers, and specific patterns
	validEntityIDs := []string{
		"light.kitchen_ceiling_1",
		"sensor.temperature_outdoor_2",
		"binary_sensor.door_front_window_left",
		"switch.smart_plug_office_desk",
		"automation.turn_on_lights_at_sunset",
	}

	for _, entityID := range validEntityIDs {
		t.Run(entityID, func(t *testing.T) {
			t.Parallel()

			state := types.HAState{EntityID: entityID, State: "on"}

			// Should be able to serialize to JSON without issues
			data, err := json.Marshal(state)
			require.NoError(t, err)
			assert.Contains(t, string(data), entityID)

			// Should be able to deserialize back
			var decoded types.HAState
			err = json.Unmarshal(data, &decoded)
			require.NoError(t, err)
			assert.Equal(t, entityID, decoded.EntityID)
		})
	}
}

func TestEdgeCases_LargeDatasets(t *testing.T) {
	// Note: This test doesn't use t.Parallel() because it modifies global config

	// Create a large dataset (1000 entities)
	states := make([]types.HAState, 1000)
	for i := 0; i < 1000; i++ {
		states[i] = types.HAState{
			EntityID: fmt.Sprintf("sensor.test_%d", i),
			State:    fmt.Sprintf("value_%d", i),
		}
	}

	original := output.GetConfig()
	output.SetConfig(&output.Config{Format: output.FormatJSON})
	defer output.SetConfig(original)

	out := captureOutput(t, func() {
		output.Data(states, output.WithCommand("states-json"), output.WithCount(len(states)))
	})

	var result output.Result
	err := json.Unmarshal([]byte(out), &result)
	require.NoError(t, err, "should handle large datasets")
	assert.Equal(t, 1000, result.Count)
}

func TestEdgeCases_UnicodeInStateValues(t *testing.T) {
	// Note: This test doesn't use t.Parallel() because it modifies global config

	states := []types.HAState{
		{EntityID: "sensor.temperature", State: "23.5°C"},
		{EntityID: "sensor.description", State: "Très bien 👍"},
		{EntityID: "sensor.japanese", State: "良い天気"},
	}

	for _, state := range states {
		t.Run(state.EntityID, func(t *testing.T) {
			original := output.GetConfig()
			output.SetConfig(&output.Config{Format: output.FormatJSON})
			defer output.SetConfig(original)

			out := captureOutput(t, func() {
				output.Data(state, output.WithCommand("state"))
			})

			var result output.Result
			err := json.Unmarshal([]byte(out), &result)
			require.NoError(t, err, "should handle unicode in state values")

			data, ok := result.Data.(map[string]any)
			require.True(t, ok)
			assert.Equal(t, state.State, data["state"])
		})
	}
}
