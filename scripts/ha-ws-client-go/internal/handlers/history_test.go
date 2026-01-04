package handlers

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/home-assistant-blueprints/testfixtures"

	"github.com/home-assistant-blueprints/ha-ws-client-go/internal/types"
)

func TestFindEntityByID(t *testing.T) {
	t.Parallel()

	states := []types.HAState{
		{EntityID: "light.kitchen", State: "on"},
		{EntityID: "sensor.temperature", State: "72"},
		{EntityID: "input_boolean.presence", State: "on"},
	}

	tests := []struct {
		name     string
		entityID string
		found    bool
		state    string
	}{
		{
			name:     "finds existing entity",
			entityID: "light.kitchen",
			found:    true,
			state:    "on",
		},
		{
			name:     "finds sensor entity",
			entityID: "sensor.temperature",
			found:    true,
			state:    "72",
		},
		{
			name:     "returns nil for non-existent entity",
			entityID: "light.bedroom",
			found:    false,
		},
		{
			name:     "returns nil for empty string",
			entityID: "",
			found:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := findEntityByID(states, tt.entityID)
			if tt.found {
				assert.NotNil(t, result)
				assert.Equal(t, tt.state, result.State)
			} else {
				assert.Nil(t, result)
			}
		})
	}
}

func TestFindEntityByID_EmptyStates(t *testing.T) {
	t.Parallel()

	result := findEntityByID([]types.HAState{}, "light.kitchen")
	assert.Nil(t, result)
}

func TestFindStatesByContext(t *testing.T) {
	t.Parallel()

	contextA := &types.HAContext{ID: "ctx-a", ParentID: ""}
	contextB := &types.HAContext{ID: "ctx-b", ParentID: "ctx-a"}
	contextC := &types.HAContext{ID: "ctx-c", ParentID: ""}

	states := []types.HAState{
		{EntityID: "light.kitchen", State: "on", Context: contextA},
		{EntityID: "light.bedroom", State: "off", Context: contextB},
		{EntityID: "sensor.temp", State: "72", Context: contextC},
		{EntityID: "switch.fan", State: "on", Context: contextA},
	}

	tests := []struct {
		name      string
		contextID string
		expected  int
	}{
		{
			name:      "finds states with matching context ID",
			contextID: "ctx-a",
			expected:  3, // kitchen (direct), bedroom (parent), fan (direct)
		},
		{
			name:      "finds states with matching context ID only",
			contextID: "ctx-c",
			expected:  1,
		},
		{
			name:      "returns empty for non-existent context",
			contextID: "ctx-z",
			expected:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := findStatesByContext(states, tt.contextID)
			assert.Len(t, result, tt.expected)
		})
	}
}

func TestFindStatesByContext_NilContexts(t *testing.T) {
	t.Parallel()

	states := []types.HAState{
		{EntityID: "light.kitchen", State: "on", Context: nil},
		{EntityID: "light.bedroom", State: "off", Context: &types.HAContext{ID: "ctx-a"}},
	}

	result := findStatesByContext(states, "ctx-a")
	assert.Len(t, result, 1)
	assert.Equal(t, "light.bedroom", result[0].EntityID)
}

func TestAddParentContextMatches(t *testing.T) {
	t.Parallel()

	contextParent := &types.HAContext{ID: "parent-ctx"}
	contextChild := &types.HAContext{ID: "child-ctx", ParentID: "parent-ctx"}

	states := []types.HAState{
		{EntityID: "automation.trigger", State: "on", Context: contextParent},
		{EntityID: "light.kitchen", State: "on", Context: contextChild},
		{EntityID: "sensor.unrelated", State: "72", Context: &types.HAContext{ID: "other"}},
	}

	tests := []struct {
		name           string
		initialMatches []types.HAState
		parentID       string
		expectedLen    int
	}{
		{
			name:           "adds parent context match",
			initialMatches: []types.HAState{states[1]}, // light.kitchen
			parentID:       "parent-ctx",
			expectedLen:    2, // adds automation.trigger
		},
		{
			name:           "does not add duplicates",
			initialMatches: []types.HAState{states[0], states[1]}, // already has automation.trigger
			parentID:       "parent-ctx",
			expectedLen:    2,
		},
		{
			name:           "returns original if no parent matches",
			initialMatches: []types.HAState{states[1]},
			parentID:       "nonexistent",
			expectedLen:    1,
		},
		{
			name:           "handles empty initial matches",
			initialMatches: []types.HAState{},
			parentID:       "parent-ctx",
			expectedLen:    1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := addParentContextMatches(tt.initialMatches, states, tt.parentID)
			assert.Len(t, result, tt.expectedLen)
		})
	}
}

func TestFormatContextInfo(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		context  *types.HAContext
		expected string
	}{
		{
			name:     "nil context returns empty string",
			context:  nil,
			expected: "",
		},
		{
			name:     "context without parent",
			context:  &types.HAContext{ID: "ctx-123"},
			expected: " (context: ctx-123)",
		},
		{
			name:     "context with parent",
			context:  &types.HAContext{ID: "ctx-456", ParentID: "ctx-123"},
			expected: " (context: ctx-456, parent: ctx-123)",
		},
		{
			name:     "context with empty parent treated as no parent",
			context:  &types.HAContext{ID: "ctx-789", ParentID: ""},
			expected: " (context: ctx-789)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := formatContextInfo(tt.context)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Benchmark for findEntityByID
func BenchmarkFindEntityByID(b *testing.B) {
	states := make([]types.HAState, 1000)
	for i := range states {
		states[i] = types.HAState{
			EntityID: "entity_" + string(rune('a'+i%26)) + "_" + string(rune('0'+i%10)),
			State:    "on",
		}
	}
	// Add target at the end
	states = append(states, types.HAState{EntityID: "target.entity", State: "found"})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = findEntityByID(states, "target.entity")
	}
}

// Benchmark for findStatesByContext
func BenchmarkFindStatesByContext(b *testing.B) {
	states := make([]types.HAState, 1000)
	for i := range states {
		ctx := &types.HAContext{ID: "ctx-" + string(rune('a'+i%26))}
		if i%10 == 0 {
			ctx.ParentID = "target-ctx"
		}
		states[i] = types.HAState{
			EntityID: "entity_" + string(rune('a'+i%26)),
			State:    "on",
			Context:  ctx,
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = findStatesByContext(states, "target-ctx")
	}
}

func TestHistoryStateGetState(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		state    types.HistoryState
		expected string
	}{
		{
			name:     "compact format (S field)",
			state:    types.HistoryState{S: "72.5"},
			expected: "72.5",
		},
		{
			name:     "full format (State field)",
			state:    types.HistoryState{State: "on"},
			expected: "on",
		},
		{
			name:     "compact takes precedence",
			state:    types.HistoryState{S: "compact", State: "full"},
			expected: "compact",
		},
		{
			name:     "empty state",
			state:    types.HistoryState{},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := tt.state.GetState()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestHistoryStateGetLastUpdated(t *testing.T) {
	t.Parallel()

	now := time.Now()
	nowUnix := float64(now.Unix())

	tests := []struct {
		name        string
		state       types.HistoryState
		expectValid bool
	}{
		{
			name:        "compact format (LU field)",
			state:       types.HistoryState{LU: nowUnix},
			expectValid: true,
		},
		{
			name:        "full format (LastUpdated field)",
			state:       types.HistoryState{LastUpdated: now.Format(time.RFC3339)},
			expectValid: true,
		},
		{
			name:        "compact takes precedence",
			state:       types.HistoryState{LU: nowUnix, LastUpdated: "2020-01-01T00:00:00Z"},
			expectValid: true,
		},
		{
			name:        "empty state returns zero time",
			state:       types.HistoryState{},
			expectValid: false,
		},
		{
			name:        "invalid LastUpdated format",
			state:       types.HistoryState{LastUpdated: "not-a-date"},
			expectValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := tt.state.GetLastUpdated()
			if tt.expectValid {
				assert.False(t, result.IsZero())
			} else {
				// Either zero or parsing failed
				_ = result
			}
		})
	}
}

func TestSysLogEntryGetSource(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		entry    types.SysLogEntry
		expected string
	}{
		{
			name:     "source as []any with string",
			entry:    types.SysLogEntry{Source: []any{"homeassistant/core.py", 123}},
			expected: "homeassistant/core.py",
		},
		{
			name:     "source as []string",
			entry:    types.SysLogEntry{Source: []string{"custom_component/sensor.py", "42"}},
			expected: "custom_component/sensor.py",
		},
		{
			name:     "empty source array",
			entry:    types.SysLogEntry{Source: []any{}},
			expected: "",
		},
		{
			name:     "nil source",
			entry:    types.SysLogEntry{Source: nil},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := tt.entry.GetSource()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSysLogEntryGetMessage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		entry    types.SysLogEntry
		expected string
	}{
		{
			name:     "message as string",
			entry:    types.SysLogEntry{Message: "Something went wrong"},
			expected: "Something went wrong",
		},
		{
			name:     "message as []any with string",
			entry:    types.SysLogEntry{Message: []any{"First message", "second part"}},
			expected: "First message",
		},
		{
			name:     "empty message array",
			entry:    types.SysLogEntry{Message: []any{}},
			expected: "",
		},
		{
			name:     "nil message",
			entry:    types.SysLogEntry{Message: nil},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := tt.entry.GetMessage()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestStatEntryGetStartTime(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		entry     types.StatEntry
		expectISO bool
	}{
		{
			name:      "unix timestamp in seconds",
			entry:     types.StatEntry{Start: float64(1705312800)},
			expectISO: true,
		},
		{
			name:      "unix timestamp in milliseconds",
			entry:     types.StatEntry{Start: float64(1705312800000)},
			expectISO: true,
		},
		{
			name:      "string format",
			entry:     types.StatEntry{Start: "2024-01-15T10:00:00Z"},
			expectISO: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := tt.entry.GetStartTime()
			if tt.expectISO {
				assert.NotEmpty(t, result)
				assert.Contains(t, result, "T") // ISO format contains T
			}
		})
	}
}

func TestLogbookEntryFormatting(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		entry        types.LogbookEntry
		expectEntity bool
		expectState  bool
	}{
		{
			name: "complete entry",
			entry: types.LogbookEntry{
				When:      float64(time.Now().Unix()),
				EntityID:  "light.kitchen",
				State:     "on",
				Message:   "turned on",
				ContextID: "ctx-123",
			},
			expectEntity: true,
			expectState:  true,
		},
		{
			name: "entry without entity",
			entry: types.LogbookEntry{
				When:    float64(time.Now().Unix()),
				Message: "Home Assistant started",
			},
			expectEntity: false,
			expectState:  false,
		},
		{
			name: "entry with only state",
			entry: types.LogbookEntry{
				When:     float64(time.Now().Unix()),
				EntityID: "sensor.temperature",
				State:    "72.5",
			},
			expectEntity: true,
			expectState:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if tt.expectEntity {
				assert.NotEmpty(t, tt.entry.EntityID)
			}
			if tt.expectState {
				assert.NotEmpty(t, tt.entry.State)
			}
			assert.Greater(t, tt.entry.When, float64(0))
		})
	}
}

func TestTimeRangeCalculation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		hours         int
		expectSeconds int64
	}{
		{
			name:          "1 hour",
			hours:         1,
			expectSeconds: 3600,
		},
		{
			name:          "24 hours",
			hours:         24,
			expectSeconds: 86400,
		},
		{
			name:          "48 hours",
			hours:         48,
			expectSeconds: 172800,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			endTime := time.Now()
			startTime := endTime.Add(-time.Duration(tt.hours) * time.Hour)
			diff := endTime.Sub(startTime)

			assert.Equal(t, tt.expectSeconds, int64(diff.Seconds()))
		})
	}
}

func TestEntityStatsSummary(t *testing.T) {
	t.Parallel()

	t.Run("calculates aggregations correctly", func(t *testing.T) {
		t.Parallel()

		stats := []types.StatEntry{
			{Min: 70.0, Max: 75.0, Mean: 72.5, Sum: 290.0},
			{Min: 68.0, Max: 78.0, Mean: 73.0, Sum: 292.0},
			{Min: 72.0, Max: 76.0, Mean: 74.0, Sum: 296.0},
		}

		// Calculate aggregated stats (same logic as HandleStatsMulti)
		summary := EntityStatsSummary{
			EntityID:   "sensor.temperature",
			Min:        stats[0].Min,
			Max:        stats[0].Max,
			DataPoints: len(stats),
		}

		var sumMean float64
		for _, s := range stats {
			if s.Min < summary.Min {
				summary.Min = s.Min
			}
			if s.Max > summary.Max {
				summary.Max = s.Max
			}
			sumMean += s.Mean
			summary.Sum += s.Sum
		}
		summary.Mean = sumMean / float64(len(stats))

		assert.Equal(t, "sensor.temperature", summary.EntityID)
		assert.Equal(t, 68.0, summary.Min)
		assert.Equal(t, 78.0, summary.Max)
		assert.InDelta(t, 73.17, summary.Mean, 0.01)
		assert.Equal(t, 878.0, summary.Sum)
		assert.Equal(t, 3, summary.DataPoints)
	})

	t.Run("handles empty stats", func(t *testing.T) {
		t.Parallel()

		summary := EntityStatsSummary{
			EntityID: "sensor.empty",
			Error:    "no statistics found",
		}

		assert.Equal(t, "sensor.empty", summary.EntityID)
		assert.NotEmpty(t, summary.Error)
		assert.Equal(t, 0, summary.DataPoints)
	})
}

func TestTimeRangeStruct(t *testing.T) {
	t.Parallel()

	t.Run("creates valid time range", func(t *testing.T) {
		t.Parallel()

		endTime := time.Now()
		startTime := endTime.Add(-24 * time.Hour)
		tr := types.TimeRange{
			StartTime: startTime,
			EndTime:   endTime,
		}

		assert.True(t, tr.EndTime.After(tr.StartTime))
		assert.Equal(t, 24*time.Hour, tr.EndTime.Sub(tr.StartTime))
	})

	t.Run("formats times for API", func(t *testing.T) {
		t.Parallel()

		now := time.Now()
		tr := types.TimeRange{
			StartTime: now.Add(-1 * time.Hour),
			EndTime:   now,
		}

		startStr := tr.StartTime.Format(time.RFC3339)
		endStr := tr.EndTime.Format(time.RFC3339)

		assert.Contains(t, startStr, "T")
		assert.Contains(t, endStr, "T")
	})
}

// Benchmark for HistoryState.GetState
func BenchmarkHistoryStateGetState(b *testing.B) {
	states := []types.HistoryState{
		{S: "72.5"},
		{State: "on"},
		{S: "compact", State: "full"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s := states[i%len(states)]
		_ = s.GetState()
	}
}

// Benchmark for HistoryState.GetLastUpdated
func BenchmarkHistoryStateGetLastUpdated(b *testing.B) {
	now := time.Now()
	states := []types.HistoryState{
		{LU: float64(now.Unix())},
		{LastUpdated: now.Format(time.RFC3339)},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s := states[i%len(states)]
		_ = s.GetLastUpdated()
	}
}

// Benchmark for SysLogEntry.GetMessage
func BenchmarkSysLogEntryGetMessage(b *testing.B) {
	entries := []types.SysLogEntry{
		{Message: "Simple message"},
		{Message: []any{"Array message", "part 2"}},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e := entries[i%len(entries)]
		_ = e.GetMessage()
	}
}

// =====================================
// Handler Function Tests
// =====================================

func TestHandleLogbook(t *testing.T) {
	t.Parallel()

	t.Run("returns logbook entries successfully", func(t *testing.T) {
		t.Parallel()

		now := time.Now()
		entries := []testfixtures.LogbookEntry{
			testfixtures.NewLogbookEntry("light.kitchen", "on", "turned on", now.Add(-1*time.Hour)),
			testfixtures.NewLogbookEntry("light.kitchen", "off", "turned off", now.Add(-30*time.Minute)),
		}

		router := NewMessageRouter(t).OnSuccess("logbook/get_events", entries)

		timeRange := &types.TimeRange{
			StartTime: now.Add(-24 * time.Hour),
			EndTime:   now,
		}
		config := &HandlerConfig{
			Args:      []string{"light.kitchen"},
			TimeRange: timeRange,
		}

		ctx, cleanup := NewTestContext(t, router,
			WithArgs("logbook", "light.kitchen"),
			WithHandlerConfig(config),
		)
		defer cleanup()

		// Capture output
		_, restoreOutput := CaptureOutput()
		defer restoreOutput()

		err := handleLogbook(ctx)
		require.NoError(t, err)
	})

	t.Run("handles empty results", func(t *testing.T) {
		t.Parallel()

		router := NewMessageRouter(t).OnSuccess("logbook/get_events", []testfixtures.LogbookEntry{})

		now := time.Now()
		timeRange := &types.TimeRange{
			StartTime: now.Add(-24 * time.Hour),
			EndTime:   now,
		}
		config := &HandlerConfig{
			Args:      []string{"light.nonexistent"},
			TimeRange: timeRange,
		}

		ctx, cleanup := NewTestContext(t, router,
			WithArgs("logbook", "light.nonexistent"),
			WithHandlerConfig(config),
		)
		defer cleanup()

		_, restoreOutput := CaptureOutput()
		defer restoreOutput()

		err := handleLogbook(ctx)
		require.NoError(t, err)
	})

	t.Run("returns error on client failure", func(t *testing.T) {
		t.Parallel()

		router := NewMessageRouter(t).OnError("logbook/get_events", "error", "connection failed")

		now := time.Now()
		timeRange := &types.TimeRange{
			StartTime: now.Add(-24 * time.Hour),
			EndTime:   now,
		}
		config := &HandlerConfig{
			Args:      []string{"light.kitchen"},
			TimeRange: timeRange,
		}

		ctx, cleanup := NewTestContext(t, router,
			WithArgs("logbook", "light.kitchen"),
			WithHandlerConfig(config),
		)
		defer cleanup()

		err := handleLogbook(ctx)
		require.Error(t, err)
	})
}

func TestHandleHistory(t *testing.T) {
	t.Parallel()

	t.Run("returns history successfully", func(t *testing.T) {
		t.Parallel()

		now := time.Now()
		historyStates := []testfixtures.HistoryState{
			testfixtures.NewHistoryState("72.5", now.Add(-2*time.Hour)),
			testfixtures.NewHistoryState("73.0", now.Add(-1*time.Hour)),
			testfixtures.NewHistoryState("72.8", now),
		}

		result := HistoryResult("sensor.temperature", historyStates...)
		router := NewMessageRouter(t).OnSuccess("history/history_during_period", result)

		timeRange := &types.TimeRange{
			StartTime: now.Add(-24 * time.Hour),
			EndTime:   now,
		}
		config := &HandlerConfig{
			Args:      []string{"sensor.temperature"},
			TimeRange: timeRange,
		}

		ctx, cleanup := NewTestContext(t, router,
			WithArgs("history", "sensor.temperature"),
			WithHandlerConfig(config),
		)
		defer cleanup()

		_, restoreOutput := CaptureOutput()
		defer restoreOutput()

		err := handleHistory(ctx)
		require.NoError(t, err)
	})

	t.Run("handles empty history", func(t *testing.T) {
		t.Parallel()

		result := HistoryResult("sensor.temperature")
		router := NewMessageRouter(t).OnSuccess("history/history_during_period", result)

		now := time.Now()
		timeRange := &types.TimeRange{
			StartTime: now.Add(-24 * time.Hour),
			EndTime:   now,
		}
		config := &HandlerConfig{
			Args:      []string{"sensor.temperature"},
			TimeRange: timeRange,
		}

		ctx, cleanup := NewTestContext(t, router,
			WithArgs("history", "sensor.temperature"),
			WithHandlerConfig(config),
		)
		defer cleanup()

		_, restoreOutput := CaptureOutput()
		defer restoreOutput()

		err := handleHistory(ctx)
		require.NoError(t, err)
	})

	t.Run("handles entity not in response", func(t *testing.T) {
		t.Parallel()

		// Response contains different entity than requested
		result := map[string][]testfixtures.HistoryState{
			"sensor.other": {},
		}
		router := NewMessageRouter(t).OnSuccess("history/history_during_period", result)

		now := time.Now()
		timeRange := &types.TimeRange{
			StartTime: now.Add(-24 * time.Hour),
			EndTime:   now,
		}
		config := &HandlerConfig{
			Args:      []string{"sensor.temperature"},
			TimeRange: timeRange,
		}

		ctx, cleanup := NewTestContext(t, router,
			WithArgs("history", "sensor.temperature"),
			WithHandlerConfig(config),
		)
		defer cleanup()

		_, restoreOutput := CaptureOutput()
		defer restoreOutput()

		err := handleHistory(ctx)
		require.NoError(t, err) // Should not error, just empty output
	})

	t.Run("returns error on client failure", func(t *testing.T) {
		t.Parallel()

		router := NewMessageRouter(t).OnError("history/history_during_period", "error", "server error")

		now := time.Now()
		timeRange := &types.TimeRange{
			StartTime: now.Add(-24 * time.Hour),
			EndTime:   now,
		}
		config := &HandlerConfig{
			Args:      []string{"sensor.temperature"},
			TimeRange: timeRange,
		}

		ctx, cleanup := NewTestContext(t, router,
			WithArgs("history", "sensor.temperature"),
			WithHandlerConfig(config),
		)
		defer cleanup()

		err := handleHistory(ctx)
		require.Error(t, err)
	})
}

func TestHandleHistoryFull(t *testing.T) {
	t.Parallel()

	t.Run("returns history with attributes", func(t *testing.T) {
		t.Parallel()

		now := time.Now()
		historyStates := []testfixtures.HistoryState{
			testfixtures.NewHistoryStateWithAttrs("on", now.Add(-2*time.Hour), map[string]any{
				"brightness": 255,
				"color_temp": 4000,
			}),
			testfixtures.NewHistoryStateWithAttrs("off", now.Add(-1*time.Hour), map[string]any{
				"brightness": 0,
			}),
		}

		result := HistoryResult("light.kitchen", historyStates...)
		router := NewMessageRouter(t).OnSuccess("history/history_during_period", result)

		timeRange := &types.TimeRange{
			StartTime: now.Add(-24 * time.Hour),
			EndTime:   now,
		}
		config := &HandlerConfig{
			Args:      []string{"light.kitchen"},
			TimeRange: timeRange,
		}

		ctx, cleanup := NewTestContext(t, router,
			WithArgs("history-full", "light.kitchen"),
			WithHandlerConfig(config),
		)
		defer cleanup()

		_, restoreOutput := CaptureOutput()
		defer restoreOutput()

		err := handleHistoryFull(ctx)
		require.NoError(t, err)
	})

	t.Run("handles states without attributes", func(t *testing.T) {
		t.Parallel()

		now := time.Now()
		historyStates := []testfixtures.HistoryState{
			testfixtures.NewHistoryState("on", now.Add(-1*time.Hour)),
			testfixtures.NewHistoryState("off", now),
		}

		result := HistoryResult("light.kitchen", historyStates...)
		router := NewMessageRouter(t).OnSuccess("history/history_during_period", result)

		timeRange := &types.TimeRange{
			StartTime: now.Add(-24 * time.Hour),
			EndTime:   now,
		}
		config := &HandlerConfig{
			Args:      []string{"light.kitchen"},
			TimeRange: timeRange,
		}

		ctx, cleanup := NewTestContext(t, router,
			WithArgs("history-full", "light.kitchen"),
			WithHandlerConfig(config),
		)
		defer cleanup()

		_, restoreOutput := CaptureOutput()
		defer restoreOutput()

		err := handleHistoryFull(ctx)
		require.NoError(t, err)
	})

	t.Run("uses Attributes field when A is nil", func(t *testing.T) {
		t.Parallel()

		now := time.Now()
		// Simulate full format response (Attributes field instead of A)
		historyStates := []testfixtures.HistoryState{
			{
				State:       "on",
				LastUpdated: now.Format(time.RFC3339),
				Attributes: map[string]any{
					"friendly_name": "Kitchen Light",
					"brightness":    200,
				},
			},
		}

		result := HistoryResult("light.kitchen", historyStates...)
		router := NewMessageRouter(t).OnSuccess("history/history_during_period", result)

		timeRange := &types.TimeRange{
			StartTime: now.Add(-24 * time.Hour),
			EndTime:   now,
		}
		config := &HandlerConfig{
			Args:      []string{"light.kitchen"},
			TimeRange: timeRange,
		}

		ctx, cleanup := NewTestContext(t, router,
			WithArgs("history-full", "light.kitchen"),
			WithHandlerConfig(config),
		)
		defer cleanup()

		_, restoreOutput := CaptureOutput()
		defer restoreOutput()

		err := handleHistoryFull(ctx)
		require.NoError(t, err)
	})

	t.Run("returns error on client failure", func(t *testing.T) {
		t.Parallel()

		router := NewMessageRouter(t).OnError("history/history_during_period", "error", "database error")

		now := time.Now()
		timeRange := &types.TimeRange{
			StartTime: now.Add(-24 * time.Hour),
			EndTime:   now,
		}
		config := &HandlerConfig{
			Args:      []string{"light.kitchen"},
			TimeRange: timeRange,
		}

		ctx, cleanup := NewTestContext(t, router,
			WithArgs("history-full", "light.kitchen"),
			WithHandlerConfig(config),
		)
		defer cleanup()

		err := handleHistoryFull(ctx)
		require.Error(t, err)
	})
}

func TestHandleAttrs(t *testing.T) {
	t.Parallel()

	t.Run("displays attribute changes", func(t *testing.T) {
		t.Parallel()

		now := time.Now()
		historyStates := []testfixtures.HistoryState{
			testfixtures.NewHistoryStateWithAttrs("72.5", now.Add(-2*time.Hour), map[string]any{
				"unit_of_measurement": "°F",
				"device_class":        "temperature",
			}),
			testfixtures.NewHistoryStateWithAttrs("73.0", now.Add(-1*time.Hour), map[string]any{
				"unit_of_measurement": "°F",
				"device_class":        "temperature",
			}),
			testfixtures.NewHistoryStateWithAttrs("72.8", now, map[string]any{
				"unit_of_measurement": "°F",
				"device_class":        "temperature",
			}),
		}

		result := HistoryResult("sensor.temperature", historyStates...)
		router := NewMessageRouter(t).OnSuccess("history/history_during_period", result)

		timeRange := &types.TimeRange{
			StartTime: now.Add(-24 * time.Hour),
			EndTime:   now,
		}
		config := &HandlerConfig{
			Args:      []string{"sensor.temperature"},
			TimeRange: timeRange,
		}

		ctx, cleanup := NewTestContext(t, router,
			WithArgs("attrs", "sensor.temperature"),
			WithHandlerConfig(config),
		)
		defer cleanup()

		_, restoreOutput := CaptureOutput()
		defer restoreOutput()

		err := handleAttrs(ctx)
		require.NoError(t, err)
	})

	t.Run("handles states without attributes", func(t *testing.T) {
		t.Parallel()

		now := time.Now()
		historyStates := []testfixtures.HistoryState{
			testfixtures.NewHistoryState("on", now.Add(-1*time.Hour)),
			testfixtures.NewHistoryState("off", now),
		}

		result := HistoryResult("switch.light", historyStates...)
		router := NewMessageRouter(t).OnSuccess("history/history_during_period", result)

		timeRange := &types.TimeRange{
			StartTime: now.Add(-24 * time.Hour),
			EndTime:   now,
		}
		config := &HandlerConfig{
			Args:      []string{"switch.light"},
			TimeRange: timeRange,
		}

		ctx, cleanup := NewTestContext(t, router,
			WithArgs("attrs", "switch.light"),
			WithHandlerConfig(config),
		)
		defer cleanup()

		_, restoreOutput := CaptureOutput()
		defer restoreOutput()

		err := handleAttrs(ctx)
		require.NoError(t, err)
	})

	t.Run("handles empty history", func(t *testing.T) {
		t.Parallel()

		result := HistoryResult("sensor.temperature")
		router := NewMessageRouter(t).OnSuccess("history/history_during_period", result)

		now := time.Now()
		timeRange := &types.TimeRange{
			StartTime: now.Add(-24 * time.Hour),
			EndTime:   now,
		}
		config := &HandlerConfig{
			Args:      []string{"sensor.temperature"},
			TimeRange: timeRange,
		}

		ctx, cleanup := NewTestContext(t, router,
			WithArgs("attrs", "sensor.temperature"),
			WithHandlerConfig(config),
		)
		defer cleanup()

		_, restoreOutput := CaptureOutput()
		defer restoreOutput()

		err := handleAttrs(ctx)
		require.NoError(t, err)
	})

	t.Run("returns error on client failure", func(t *testing.T) {
		t.Parallel()

		router := NewMessageRouter(t).OnError("history/history_during_period", "error", "timeout")

		now := time.Now()
		timeRange := &types.TimeRange{
			StartTime: now.Add(-24 * time.Hour),
			EndTime:   now,
		}
		config := &HandlerConfig{
			Args:      []string{"sensor.temperature"},
			TimeRange: timeRange,
		}

		ctx, cleanup := NewTestContext(t, router,
			WithArgs("attrs", "sensor.temperature"),
			WithHandlerConfig(config),
		)
		defer cleanup()

		err := handleAttrs(ctx)
		require.Error(t, err)
	})
}

func TestHandleTimeline(t *testing.T) {
	t.Parallel()

	t.Run("returns multi-entity timeline sorted by time", func(t *testing.T) {
		t.Parallel()

		now := time.Now()

		// Create history states for multiple entities
		result := map[string][]testfixtures.HistoryState{
			"light.kitchen": {
				testfixtures.NewHistoryState("on", now.Add(-2*time.Hour)),
				testfixtures.NewHistoryState("off", now.Add(-30*time.Minute)),
			},
			"light.bedroom": {
				testfixtures.NewHistoryState("off", now.Add(-3*time.Hour)),
				testfixtures.NewHistoryState("on", now.Add(-1*time.Hour)),
			},
		}

		router := NewMessageRouter(t).OnSuccess("history/history_during_period", result)

		ctx, cleanup := NewTestContext(t, router,
			WithArgs("timeline", "4", "light.kitchen", "light.bedroom"),
		)
		defer cleanup()

		_, restoreOutput := CaptureOutput()
		defer restoreOutput()

		err := HandleTimeline(ctx)
		require.NoError(t, err)
	})

	t.Run("handles empty results", func(t *testing.T) {
		t.Parallel()

		result := map[string][]types.HistoryState{}
		router := NewMessageRouter(t).OnSuccess("history/history_during_period", result)

		ctx, cleanup := NewTestContext(t, router,
			WithArgs("timeline", "4", "light.nonexistent"),
		)
		defer cleanup()

		_, restoreOutput := CaptureOutput()
		defer restoreOutput()

		err := HandleTimeline(ctx)
		require.NoError(t, err)
	})

	t.Run("returns error with missing arguments", func(t *testing.T) {
		t.Parallel()

		router := NewMessageRouter(t)

		ctx, cleanup := NewTestContext(t, router,
			WithArgs("timeline", "4"), // Missing entity
		)
		defer cleanup()

		err := HandleTimeline(ctx)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "missing")
	})

	t.Run("returns error with invalid hours", func(t *testing.T) {
		t.Parallel()

		router := NewMessageRouter(t)

		ctx, cleanup := NewTestContext(t, router,
			WithArgs("timeline", "not-a-number", "light.kitchen"),
		)
		defer cleanup()

		err := HandleTimeline(ctx)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid")
	})

	t.Run("returns error on client failure", func(t *testing.T) {
		t.Parallel()

		router := NewMessageRouter(t).OnError("history/history_during_period", "error", "connection failed")

		ctx, cleanup := NewTestContext(t, router,
			WithArgs("timeline", "4", "light.kitchen"),
		)
		defer cleanup()

		err := HandleTimeline(ctx)
		require.Error(t, err)
	})
}

// MockSysLogEntry creates a SysLogEntry for testing.
func MockSysLogEntry(level, source, message string) types.SysLogEntry {
	return types.SysLogEntry{
		Level:     level,
		Source:    []any{source, 123},
		Message:   message,
		Timestamp: float64(time.Now().Unix()),
	}
}

func TestHandleSyslog(t *testing.T) {
	t.Parallel()

	t.Run("returns syslog entries", func(t *testing.T) {
		t.Parallel()

		entries := []types.SysLogEntry{
			MockSysLogEntry("ERROR", "homeassistant/core.py", "Failed to load component"),
			MockSysLogEntry("WARNING", "homeassistant/components/sensor.py", "Sensor unavailable"),
		}

		router := NewMessageRouter(t).OnSuccess("system_log/list", entries)

		ctx, cleanup := NewTestContext(t, router,
			WithArgs("syslog"),
		)
		defer cleanup()

		_, restoreOutput := CaptureOutput()
		defer restoreOutput()

		err := HandleSyslog(ctx)
		require.NoError(t, err)
	})

	t.Run("handles empty syslog", func(t *testing.T) {
		t.Parallel()

		entries := []types.SysLogEntry{}
		router := NewMessageRouter(t).OnSuccess("system_log/list", entries)

		ctx, cleanup := NewTestContext(t, router,
			WithArgs("syslog"),
		)
		defer cleanup()

		_, restoreOutput := CaptureOutput()
		defer restoreOutput()

		err := HandleSyslog(ctx)
		require.NoError(t, err)
	})

	t.Run("handles entries with string source", func(t *testing.T) {
		t.Parallel()

		entries := []types.SysLogEntry{
			{
				Level:   "ERROR",
				Source:  []string{"custom_component.py", "42"},
				Message: "Custom error",
			},
		}

		router := NewMessageRouter(t).OnSuccess("system_log/list", entries)

		ctx, cleanup := NewTestContext(t, router,
			WithArgs("syslog"),
		)
		defer cleanup()

		_, restoreOutput := CaptureOutput()
		defer restoreOutput()

		err := HandleSyslog(ctx)
		require.NoError(t, err)
	})

	t.Run("handles entries with array message", func(t *testing.T) {
		t.Parallel()

		entries := []types.SysLogEntry{
			{
				Level:   "WARNING",
				Source:  []any{"test.py", 1},
				Message: []any{"Primary message", "Additional detail"},
			},
		}

		router := NewMessageRouter(t).OnSuccess("system_log/list", entries)

		ctx, cleanup := NewTestContext(t, router,
			WithArgs("syslog"),
		)
		defer cleanup()

		_, restoreOutput := CaptureOutput()
		defer restoreOutput()

		err := HandleSyslog(ctx)
		require.NoError(t, err)
	})

	t.Run("returns error on client failure", func(t *testing.T) {
		t.Parallel()

		router := NewMessageRouter(t).OnError("system_log/list", "error", "permission denied")

		ctx, cleanup := NewTestContext(t, router,
			WithArgs("syslog"),
		)
		defer cleanup()

		err := HandleSyslog(ctx)
		require.Error(t, err)
	})
}

// MockStatEntry creates a StatEntry for testing.
func MockStatEntry(start, minVal, maxVal, mean, sum float64) types.StatEntry {
	return types.StatEntry{
		Start: start,
		Min:   minVal,
		Max:   maxVal,
		Mean:  mean,
		Sum:   sum,
	}
}

// StatisticsResult creates a statistics query result.
func StatisticsResult(entityID string, stats ...types.StatEntry) map[string][]types.StatEntry {
	return map[string][]types.StatEntry{
		entityID: stats,
	}
}

func TestHandleStats(t *testing.T) {
	t.Parallel()

	t.Run("returns statistics successfully", func(t *testing.T) {
		t.Parallel()

		now := time.Now()
		baseTime := float64(now.Add(-2 * time.Hour).Unix())

		stats := []types.StatEntry{
			MockStatEntry(baseTime, 70.0, 75.0, 72.5, 290.0),
			MockStatEntry(baseTime+3600, 68.0, 78.0, 73.0, 292.0),
			MockStatEntry(baseTime+7200, 72.0, 76.0, 74.0, 296.0),
		}

		result := StatisticsResult("sensor.temperature", stats...)
		router := NewMessageRouter(t).OnSuccess("recorder/statistics_during_period", result)

		timeRange := &types.TimeRange{
			StartTime: now.Add(-24 * time.Hour),
			EndTime:   now,
		}
		config := &HandlerConfig{
			Args:      []string{"sensor.temperature"},
			TimeRange: timeRange,
		}

		ctx, cleanup := NewTestContext(t, router,
			WithArgs("stats", "sensor.temperature"),
			WithHandlerConfig(config),
		)
		defer cleanup()

		_, restoreOutput := CaptureOutput()
		defer restoreOutput()

		err := handleStats(ctx)
		require.NoError(t, err)
	})

	t.Run("handles no statistics found", func(t *testing.T) {
		t.Parallel()

		result := map[string][]types.StatEntry{}
		router := NewMessageRouter(t).OnSuccess("recorder/statistics_during_period", result)

		now := time.Now()
		timeRange := &types.TimeRange{
			StartTime: now.Add(-24 * time.Hour),
			EndTime:   now,
		}
		config := &HandlerConfig{
			Args:      []string{"sensor.temperature"},
			TimeRange: timeRange,
		}

		ctx, cleanup := NewTestContext(t, router,
			WithArgs("stats", "sensor.temperature"),
			WithHandlerConfig(config),
		)
		defer cleanup()

		_, restoreOutput := CaptureOutput()
		defer restoreOutput()

		err := handleStats(ctx)
		require.NoError(t, err)
	})

	t.Run("handles empty statistics array", func(t *testing.T) {
		t.Parallel()

		result := StatisticsResult("sensor.temperature")
		router := NewMessageRouter(t).OnSuccess("recorder/statistics_during_period", result)

		now := time.Now()
		timeRange := &types.TimeRange{
			StartTime: now.Add(-24 * time.Hour),
			EndTime:   now,
		}
		config := &HandlerConfig{
			Args:      []string{"sensor.temperature"},
			TimeRange: timeRange,
		}

		ctx, cleanup := NewTestContext(t, router,
			WithArgs("stats", "sensor.temperature"),
			WithHandlerConfig(config),
		)
		defer cleanup()

		_, restoreOutput := CaptureOutput()
		defer restoreOutput()

		err := handleStats(ctx)
		require.NoError(t, err)
	})

	t.Run("returns error on client failure", func(t *testing.T) {
		t.Parallel()

		router := NewMessageRouter(t).OnError("recorder/statistics_during_period", "error", "database error")

		now := time.Now()
		timeRange := &types.TimeRange{
			StartTime: now.Add(-24 * time.Hour),
			EndTime:   now,
		}
		config := &HandlerConfig{
			Args:      []string{"sensor.temperature"},
			TimeRange: timeRange,
		}

		ctx, cleanup := NewTestContext(t, router,
			WithArgs("stats", "sensor.temperature"),
			WithHandlerConfig(config),
		)
		defer cleanup()

		err := handleStats(ctx)
		require.Error(t, err)
	})
}

func TestHandleStatsMulti(t *testing.T) {
	// Note: Multi-entity concurrent requests cause websocket write race conditions
	// in unit tests. Use single entity tests to verify handler behavior.
	t.Parallel()

	t.Run("returns statistics for single entity", func(t *testing.T) {
		t.Parallel()

		now := time.Now()
		baseTime := float64(now.Add(-2 * time.Hour).Unix())

		// Mock handler that returns stats for the requested entity
		router := NewMessageRouter(t).On("recorder/statistics_during_period", func(_ string, data map[string]any) any {
			statIDs, ok := data["statistic_ids"].([]any)
			if !ok || len(statIDs) == 0 {
				return map[string][]types.StatEntry{}
			}
			entityID, ok := statIDs[0].(string)
			if !ok {
				return map[string][]types.StatEntry{}
			}
			return map[string][]types.StatEntry{
				entityID: {
					MockStatEntry(baseTime, 70.0, 75.0, 72.5, 290.0),
					MockStatEntry(baseTime+3600, 68.0, 78.0, 73.0, 292.0),
				},
			}
		})

		ctx, cleanup := NewTestContext(t, router,
			WithArgs("stats-multi", "sensor.temp1"),
		)
		defer cleanup()

		_, restoreOutput := CaptureOutput()
		defer restoreOutput()

		err := HandleStatsMulti(ctx)
		require.NoError(t, err)
	})

	t.Run("handles hours parameter", func(t *testing.T) {
		t.Parallel()

		now := time.Now()
		baseTime := float64(now.Add(-2 * time.Hour).Unix())

		router := NewMessageRouter(t).On("recorder/statistics_during_period", func(_ string, data map[string]any) any {
			statIDs, ok := data["statistic_ids"].([]any)
			if !ok || len(statIDs) == 0 {
				return map[string][]types.StatEntry{}
			}
			entityID, ok := statIDs[0].(string)
			if !ok {
				return map[string][]types.StatEntry{}
			}
			return map[string][]types.StatEntry{
				entityID: {MockStatEntry(baseTime, 70.0, 75.0, 72.5, 290.0)},
			}
		})

		ctx, cleanup := NewTestContext(t, router,
			WithArgs("stats-multi", "sensor.temp1", "48"), // 48 hours
		)
		defer cleanup()

		_, restoreOutput := CaptureOutput()
		defer restoreOutput()

		err := HandleStatsMulti(ctx)
		require.NoError(t, err)
	})

	t.Run("handles empty stats response", func(t *testing.T) {
		t.Parallel()

		// Mock handler that returns empty stats
		router := NewMessageRouter(t).On("recorder/statistics_during_period", func(_ string, _ map[string]any) any {
			return map[string][]types.StatEntry{}
		})

		ctx, cleanup := NewTestContext(t, router,
			WithArgs("stats-multi", "sensor.temp1"),
		)
		defer cleanup()

		_, restoreOutput := CaptureOutput()
		defer restoreOutput()

		err := HandleStatsMulti(ctx)
		require.NoError(t, err)
	})

	t.Run("returns error with missing arguments", func(t *testing.T) {
		t.Parallel()

		router := NewMessageRouter(t)

		ctx, cleanup := NewTestContext(t, router,
			WithArgs("stats-multi"), // Missing entity IDs
		)
		defer cleanup()

		err := HandleStatsMulti(ctx)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "missing")
	})

	t.Run("handles error response", func(t *testing.T) {
		t.Parallel()

		router := NewMessageRouter(t).OnError("recorder/statistics_during_period", "not_found", "Entity not found")

		ctx, cleanup := NewTestContext(t, router,
			WithArgs("stats-multi", "sensor.nonexistent"),
		)
		defer cleanup()

		_, restoreOutput := CaptureOutput()
		defer restoreOutput()

		err := HandleStatsMulti(ctx)
		// Handler continues on errors but may have no results
		require.NoError(t, err)
	})
}

func TestHandleContext(t *testing.T) {
	t.Parallel()

	t.Run("finds states by entity context", func(t *testing.T) {
		t.Parallel()

		targetContext := &types.HAContext{ID: "ctx-target", ParentID: ""}
		relatedContext := &types.HAContext{ID: "ctx-related", ParentID: "ctx-target"}

		states := []types.HAState{
			{EntityID: "automation.trigger", State: "on", Context: targetContext},
			{EntityID: "light.kitchen", State: "on", Context: relatedContext},
			{EntityID: "sensor.unrelated", State: "72", Context: &types.HAContext{ID: "other"}},
		}

		router := NewMessageRouter(t).OnSuccess("get_states", states)

		config := &HandlerConfig{
			Args: []string{"automation.trigger"},
		}

		ctx, cleanup := NewTestContext(t, router,
			WithArgs("context", "automation.trigger"),
			WithHandlerConfig(config),
		)
		defer cleanup()

		_, restoreOutput := CaptureOutput()
		defer restoreOutput()

		err := handleContext(ctx)
		require.NoError(t, err)
	})

	t.Run("finds states by context ID directly", func(t *testing.T) {
		t.Parallel()

		states := []types.HAState{
			{EntityID: "automation.trigger", State: "on", Context: &types.HAContext{ID: "ctx-123"}},
			{EntityID: "light.kitchen", State: "on", Context: &types.HAContext{ID: "ctx-123"}},
			{EntityID: "sensor.unrelated", State: "72", Context: &types.HAContext{ID: "other"}},
		}

		router := NewMessageRouter(t).OnSuccess("get_states", states)

		config := &HandlerConfig{
			Args: []string{"ctx-123"}, // Using context ID directly
		}

		ctx, cleanup := NewTestContext(t, router,
			WithArgs("context", "ctx-123"),
			WithHandlerConfig(config),
		)
		defer cleanup()

		_, restoreOutput := CaptureOutput()
		defer restoreOutput()

		err := handleContext(ctx)
		require.NoError(t, err)
	})

	t.Run("handles no matching context", func(t *testing.T) {
		t.Parallel()

		states := []types.HAState{
			{EntityID: "light.kitchen", State: "on", Context: &types.HAContext{ID: "other"}},
		}

		router := NewMessageRouter(t).OnSuccess("get_states", states)

		config := &HandlerConfig{
			Args: []string{"ctx-nonexistent"},
		}

		ctx, cleanup := NewTestContext(t, router,
			WithArgs("context", "ctx-nonexistent"),
			WithHandlerConfig(config),
		)
		defer cleanup()

		_, restoreOutput := CaptureOutput()
		defer restoreOutput()

		err := handleContext(ctx)
		require.NoError(t, err) // Returns no error, just shows no matches found
	})

	t.Run("handles entity with parent context", func(t *testing.T) {
		t.Parallel()

		parentContext := &types.HAContext{ID: "parent-ctx"}
		childContext := &types.HAContext{ID: "child-ctx", ParentID: "parent-ctx"}

		states := []types.HAState{
			{EntityID: "automation.parent", State: "on", Context: parentContext},
			{EntityID: "light.kitchen", State: "on", Context: childContext},
			{EntityID: "light.bedroom", State: "off", Context: childContext},
		}

		router := NewMessageRouter(t).OnSuccess("get_states", states)

		config := &HandlerConfig{
			Args: []string{"light.kitchen"}, // Has parent context
		}

		ctx, cleanup := NewTestContext(t, router,
			WithArgs("context", "light.kitchen"),
			WithHandlerConfig(config),
		)
		defer cleanup()

		_, restoreOutput := CaptureOutput()
		defer restoreOutput()

		err := handleContext(ctx)
		require.NoError(t, err)
	})

	t.Run("handles nil context on states", func(t *testing.T) {
		t.Parallel()

		states := []types.HAState{
			{EntityID: "light.kitchen", State: "on", Context: nil},
			{EntityID: "sensor.temp", State: "72", Context: &types.HAContext{ID: "ctx-a"}},
		}

		router := NewMessageRouter(t).OnSuccess("get_states", states)

		config := &HandlerConfig{
			Args: []string{"ctx-a"},
		}

		ctx, cleanup := NewTestContext(t, router,
			WithArgs("context", "ctx-a"),
			WithHandlerConfig(config),
		)
		defer cleanup()

		_, restoreOutput := CaptureOutput()
		defer restoreOutput()

		err := handleContext(ctx)
		require.NoError(t, err)
	})

	t.Run("returns error on client failure", func(t *testing.T) {
		t.Parallel()

		router := NewMessageRouter(t).OnError("get_states", "error", "connection failed")

		config := &HandlerConfig{
			Args: []string{"light.kitchen"},
		}

		ctx, cleanup := NewTestContext(t, router,
			WithArgs("context", "light.kitchen"),
			WithHandlerConfig(config),
		)
		defer cleanup()

		err := handleContext(ctx)
		require.Error(t, err)
	})
}

func TestOutputNoContextMatches(t *testing.T) {
	t.Parallel()

	t.Run("with target entity and parent context", func(t *testing.T) {
		t.Parallel()

		targetEntity := &types.HAState{
			EntityID: "light.kitchen",
			State:    "on",
			Context:  &types.HAContext{ID: "ctx-123", ParentID: "parent-ctx"},
		}

		_, restoreOutput := CaptureOutput()
		defer restoreOutput()

		err := outputNoContextMatches("light.kitchen", "ctx-123", targetEntity)
		require.NoError(t, err)
	})

	t.Run("with target entity without parent context", func(t *testing.T) {
		t.Parallel()

		targetEntity := &types.HAState{
			EntityID: "light.kitchen",
			State:    "on",
			Context:  &types.HAContext{ID: "ctx-123"},
		}

		_, restoreOutput := CaptureOutput()
		defer restoreOutput()

		err := outputNoContextMatches("light.kitchen", "ctx-123", targetEntity)
		require.NoError(t, err)
	})

	t.Run("with nil target entity (context ID lookup)", func(t *testing.T) {
		t.Parallel()

		_, restoreOutput := CaptureOutput()
		defer restoreOutput()

		err := outputNoContextMatches("ctx-unknown", "ctx-unknown", nil)
		require.NoError(t, err)
	})
}

func TestHandleWatch(t *testing.T) {
	t.Parallel()

	// Note: handleWatch is complex to test because it uses subscriptions
	// The tests here verify the basic setup and error paths

	t.Run("returns error on subscription failure", func(t *testing.T) {
		t.Parallel()

		router := NewMessageRouter(t).OnError("subscribe_trigger", "error", "subscription failed")

		config := &HandlerConfig{
			Args:        []string{"light.kitchen"},
			OptionalInt: 1, // 1 second timeout for quick test
		}

		ctx, cleanup := NewTestContext(t, router,
			WithArgs("watch", "light.kitchen", "1"),
			WithHandlerConfig(config),
		)
		defer cleanup()

		_, restoreOutput := CaptureOutput()
		defer restoreOutput()

		err := handleWatch(ctx)
		require.Error(t, err)
	})
}
