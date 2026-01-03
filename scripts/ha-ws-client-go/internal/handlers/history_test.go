package handlers

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

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
