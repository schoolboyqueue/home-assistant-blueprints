package types

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHistoryState_GetState(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		state    HistoryState
		expected string
	}{
		{
			name:     "compact format with S field",
			state:    HistoryState{S: "on"},
			expected: "on",
		},
		{
			name:     "full format with State field",
			state:    HistoryState{State: "off"},
			expected: "off",
		},
		{
			name:     "compact takes precedence over full",
			state:    HistoryState{S: "on", State: "off"},
			expected: "on",
		},
		{
			name:     "empty state returns empty string",
			state:    HistoryState{},
			expected: "",
		},
		{
			name:     "numeric state as string",
			state:    HistoryState{S: "23.5"},
			expected: "23.5",
		},
		{
			name:     "unavailable state",
			state:    HistoryState{State: "unavailable"},
			expected: "unavailable",
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

func TestHistoryState_GetLastUpdated(t *testing.T) {
	t.Parallel()

	// Fixed reference time for tests
	refTime := time.Date(2024, 6, 15, 10, 30, 45, 0, time.UTC)
	refTimestamp := float64(refTime.Unix()) + 0.123456789

	tests := []struct {
		name      string
		state     HistoryState
		checkFunc func(t *testing.T, result time.Time)
	}{
		{
			name:  "compact format with LU Unix timestamp",
			state: HistoryState{LU: refTimestamp},
			checkFunc: func(t *testing.T, result time.Time) {
				t.Helper()
				assert.Equal(t, refTime.Unix(), result.Unix())
				// Check nanoseconds are preserved (approximately)
				assert.InDelta(t, 123456789, result.Nanosecond(), 1000)
			},
		},
		{
			name:  "full format with RFC3339 LastUpdated",
			state: HistoryState{LastUpdated: "2024-06-15T10:30:45Z"},
			checkFunc: func(t *testing.T, result time.Time) {
				t.Helper()
				assert.Equal(t, refTime.Unix(), result.Unix())
			},
		},
		{
			name:  "LU takes precedence over LastUpdated",
			state: HistoryState{LU: refTimestamp, LastUpdated: "2020-01-01T00:00:00Z"},
			checkFunc: func(t *testing.T, result time.Time) {
				t.Helper()
				assert.Equal(t, refTime.Unix(), result.Unix())
			},
		},
		{
			name:  "empty state returns zero time",
			state: HistoryState{},
			checkFunc: func(t *testing.T, result time.Time) {
				t.Helper()
				assert.True(t, result.IsZero())
			},
		},
		{
			name:  "invalid RFC3339 returns zero time",
			state: HistoryState{LastUpdated: "not-a-date"},
			checkFunc: func(t *testing.T, result time.Time) {
				t.Helper()
				assert.True(t, result.IsZero())
			},
		},
		{
			name:  "zero LU with valid LastUpdated uses LastUpdated",
			state: HistoryState{LU: 0, LastUpdated: "2024-06-15T10:30:45Z"},
			checkFunc: func(t *testing.T, result time.Time) {
				t.Helper()
				assert.Equal(t, refTime.Unix(), result.Unix())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := tt.state.GetLastUpdated()
			tt.checkFunc(t, result)
		})
	}
}

func TestTraceDetail_GetTrigger(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		detail   TraceDetail
		expected *TraceTrigger
	}{
		{
			name:     "nil trigger returns nil",
			detail:   TraceDetail{Trigger: nil},
			expected: nil,
		},
		{
			name: "map with all fields",
			detail: TraceDetail{Trigger: map[string]any{
				"id":          "trigger_1",
				"idx":         "0",
				"alias":       "Motion Detected",
				"platform":    "state",
				"entity_id":   "binary_sensor.motion",
				"description": "Motion sensor triggered",
			}},
			expected: &TraceTrigger{
				ID:          "trigger_1",
				Idx:         "0",
				Alias:       "Motion Detected",
				Platform:    "state",
				EntityID:    "binary_sensor.motion",
				Description: "Motion sensor triggered",
			},
		},
		{
			name: "map with partial fields",
			detail: TraceDetail{Trigger: map[string]any{
				"platform":  "time",
				"entity_id": "sensor.time",
			}},
			expected: &TraceTrigger{
				Platform: "time",
				EntityID: "sensor.time",
			},
		},
		{
			name:     "string trigger returns nil",
			detail:   TraceDetail{Trigger: "some string trigger"},
			expected: nil,
		},
		{
			name:     "integer trigger returns nil",
			detail:   TraceDetail{Trigger: 42},
			expected: nil,
		},
		{
			name:     "empty map returns empty trigger",
			detail:   TraceDetail{Trigger: map[string]any{}},
			expected: &TraceTrigger{},
		},
		{
			name: "map with wrong types ignores fields",
			detail: TraceDetail{Trigger: map[string]any{
				"id":       123, // int instead of string
				"platform": "state",
			}},
			expected: &TraceTrigger{
				Platform: "state",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := tt.detail.GetTrigger()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTraceDetail_GetTriggerDescription(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		detail   TraceDetail
		expected string
	}{
		{
			name:     "nil trigger returns empty",
			detail:   TraceDetail{Trigger: nil},
			expected: "",
		},
		{
			name:     "string trigger returns string",
			detail:   TraceDetail{Trigger: "Manual trigger"},
			expected: "Manual trigger",
		},
		{
			name: "map with description returns description",
			detail: TraceDetail{Trigger: map[string]any{
				"description": "Motion sensor activated",
				"platform":    "state",
			}},
			expected: "Motion sensor activated",
		},
		{
			name: "map without description falls back to platform",
			detail: TraceDetail{Trigger: map[string]any{
				"platform": "time",
			}},
			expected: "time",
		},
		{
			name:     "map without description or platform returns trigger",
			detail:   TraceDetail{Trigger: map[string]any{"id": "1"}},
			expected: "trigger",
		},
		{
			name:     "empty map returns trigger",
			detail:   TraceDetail{Trigger: map[string]any{}},
			expected: "trigger",
		},
		{
			name:     "integer trigger returns empty",
			detail:   TraceDetail{Trigger: 123},
			expected: "",
		},
		{
			name:     "slice trigger returns empty",
			detail:   TraceDetail{Trigger: []string{"a", "b"}},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := tt.detail.GetTriggerDescription()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSysLogEntry_GetSource(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		entry    SysLogEntry
		expected string
	}{
		{
			name:     "nil source returns empty",
			entry:    SysLogEntry{Source: nil},
			expected: "",
		},
		{
			name:     "[]any with string first element",
			entry:    SysLogEntry{Source: []any{"components/switch.py", 123}},
			expected: "components/switch.py",
		},
		{
			name:     "[]any with non-string first element",
			entry:    SysLogEntry{Source: []any{123, "file.py"}},
			expected: "",
		},
		{
			name:     "[]string with elements",
			entry:    SysLogEntry{Source: []string{"homeassistant/core.py", "42"}},
			expected: "homeassistant/core.py",
		},
		{
			name:     "empty []any returns empty",
			entry:    SysLogEntry{Source: []any{}},
			expected: "",
		},
		{
			name:     "empty []string returns empty",
			entry:    SysLogEntry{Source: []string{}},
			expected: "",
		},
		{
			name:     "string source returns empty (not supported)",
			entry:    SysLogEntry{Source: "direct-string"},
			expected: "",
		},
		{
			name:     "integer source returns empty",
			entry:    SysLogEntry{Source: 42},
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

func TestSysLogEntry_GetMessage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		entry    SysLogEntry
		expected string
	}{
		{
			name:     "nil message returns empty",
			entry:    SysLogEntry{Message: nil},
			expected: "",
		},
		{
			name:     "string message",
			entry:    SysLogEntry{Message: "Error loading component"},
			expected: "Error loading component",
		},
		{
			name:     "[]any with string first element",
			entry:    SysLogEntry{Message: []any{"Primary message", "secondary"}},
			expected: "Primary message",
		},
		{
			name:     "[]any with non-string first element",
			entry:    SysLogEntry{Message: []any{123, "message"}},
			expected: "",
		},
		{
			name:     "empty []any returns empty",
			entry:    SysLogEntry{Message: []any{}},
			expected: "",
		},
		{
			name:     "integer message returns empty",
			entry:    SysLogEntry{Message: 42},
			expected: "",
		},
		{
			name:     "empty string returns empty",
			entry:    SysLogEntry{Message: ""},
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

func TestStatEntry_GetStartTime(t *testing.T) {
	t.Parallel()

	// Reference time
	refTime := time.Date(2024, 6, 15, 10, 0, 0, 0, time.UTC)

	tests := []struct {
		name      string
		entry     StatEntry
		checkFunc func(t *testing.T, result string)
	}{
		{
			name:  "string ISO format passthrough",
			entry: StatEntry{Start: "2024-06-15T10:00:00+00:00"},
			checkFunc: func(t *testing.T, result string) {
				t.Helper()
				assert.Equal(t, "2024-06-15T10:00:00+00:00", result)
			},
		},
		{
			name:  "float64 Unix seconds",
			entry: StatEntry{Start: float64(refTime.Unix())},
			checkFunc: func(t *testing.T, result string) {
				t.Helper()
				// Parse the result and compare Unix timestamps
				parsed, err := time.Parse(time.RFC3339, result)
				require.NoError(t, err)
				assert.Equal(t, refTime.Unix(), parsed.Unix())
			},
		},
		{
			name:  "float64 Unix milliseconds (large value)",
			entry: StatEntry{Start: float64(refTime.UnixMilli())},
			checkFunc: func(t *testing.T, result string) {
				t.Helper()
				parsed, err := time.Parse(time.RFC3339, result)
				require.NoError(t, err)
				assert.Equal(t, refTime.Unix(), parsed.Unix())
			},
		},
		{
			name:  "nil Start returns empty",
			entry: StatEntry{Start: nil},
			checkFunc: func(t *testing.T, result string) {
				t.Helper()
				assert.Empty(t, result)
			},
		},
		{
			name:  "integer Start returns empty (not supported)",
			entry: StatEntry{Start: refTime.Unix()},
			checkFunc: func(t *testing.T, result string) {
				t.Helper()
				assert.Empty(t, result)
			},
		},
		{
			name:  "zero float64 returns epoch",
			entry: StatEntry{Start: float64(0)},
			checkFunc: func(t *testing.T, result string) {
				t.Helper()
				parsed, err := time.Parse(time.RFC3339, result)
				require.NoError(t, err)
				assert.Equal(t, int64(0), parsed.Unix())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := tt.entry.GetStartTime()
			tt.checkFunc(t, result)
		})
	}
}

func TestCLIFlags_GetOutputFormat(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		flags    CLIFlags
		expected string
	}{
		{
			name:     "default with empty Output",
			flags:    CLIFlags{},
			expected: "",
		},
		{
			name:     "Output field only",
			flags:    CLIFlags{Output: "compact"},
			expected: "compact",
		},
		{
			name:     "Compact flag takes precedence over Output",
			flags:    CLIFlags{Compact: true, Output: "json"},
			expected: "compact",
		},
		{
			name:     "JSON flag takes precedence over Compact",
			flags:    CLIFlags{JSON: true, Compact: true, Output: "default"},
			expected: "json",
		},
		{
			name:     "JSON flag only",
			flags:    CLIFlags{JSON: true},
			expected: "json",
		},
		{
			name:     "Compact flag only",
			flags:    CLIFlags{Compact: true},
			expected: "compact",
		},
		{
			name:     "Output set to default",
			flags:    CLIFlags{Output: "default"},
			expected: "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := tt.flags.GetOutputFormat()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCLIFlags_ShowHelp(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		flags    CLIFlags
		expected bool
	}{
		{
			name:     "Help flag true",
			flags:    CLIFlags{Help: true},
			expected: true,
		},
		{
			name:     "Help flag false",
			flags:    CLIFlags{Help: false},
			expected: false,
		},
		{
			name:     "Default zero value",
			flags:    CLIFlags{},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := tt.flags.ShowHelp()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCLIFlags_ShowVersion(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		flags    CLIFlags
		expected bool
	}{
		{
			name:     "Version flag true",
			flags:    CLIFlags{Version: true},
			expected: true,
		},
		{
			name:     "Version flag false",
			flags:    CLIFlags{Version: false},
			expected: false,
		},
		{
			name:     "Default zero value",
			flags:    CLIFlags{},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := tt.flags.ShowVersion()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTimeRange(t *testing.T) {
	t.Parallel()

	start := time.Now().Add(-1 * time.Hour)
	end := time.Now()

	tr := TimeRange{
		StartTime: start,
		EndTime:   end,
	}

	require.Equal(t, start, tr.StartTime)
	require.Equal(t, end, tr.EndTime)
}
