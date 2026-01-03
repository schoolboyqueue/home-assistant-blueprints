package output

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// captureOutput captures stdout during function execution
func captureOutput(t *testing.T, f func()) string {
	t.Helper()

	old := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)

	os.Stdout = w

	f()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	_, err = io.Copy(&buf, r)
	require.NoError(t, err)

	return buf.String()
}

// captureStderr captures stderr during function execution
func captureStderr(t *testing.T, f func()) string {
	t.Helper()

	old := os.Stderr
	r, w, err := os.Pipe()
	require.NoError(t, err)

	os.Stderr = w

	f()

	w.Close()
	os.Stderr = old

	var buf bytes.Buffer
	_, err = io.Copy(&buf, r)
	require.NoError(t, err)

	return buf.String()
}

// resetGlobalConfig resets the global config to default before each test
func resetGlobalConfig() {
	globalConfig = DefaultConfig()
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	assert.Equal(t, FormatDefault, cfg.Format)
	assert.True(t, cfg.ShowTimestamps)
	assert.True(t, cfg.ShowHeaders)
	assert.False(t, cfg.ShowAge)
	assert.Equal(t, 0, cfg.MaxItems)
}

func TestGetSetConfig(t *testing.T) {
	// Save and restore global state
	original := globalConfig
	defer func() { globalConfig = original }()

	cfg := &Config{
		Format:         FormatJSON,
		ShowTimestamps: false,
		ShowHeaders:    false,
		ShowAge:        true,
		MaxItems:       100,
	}

	SetConfig(cfg)
	got := GetConfig()

	assert.Equal(t, FormatJSON, got.Format)
	assert.False(t, got.ShowTimestamps)
	assert.False(t, got.ShowHeaders)
	assert.True(t, got.ShowAge)
	assert.Equal(t, 100, got.MaxItems)
}

func TestConfigureFromFlags(t *testing.T) {
	// Save and restore global state
	original := globalConfig
	defer func() { globalConfig = original }()

	tests := []struct {
		name         string
		format       string
		noHeaders    bool
		noTimestamps bool
		showAge      bool
		maxItems     int
		expected     *Config
	}{
		{
			name:         "json format",
			format:       "json",
			noHeaders:    false,
			noTimestamps: false,
			showAge:      false,
			maxItems:     0,
			expected: &Config{
				Format:         FormatJSON,
				ShowHeaders:    true,
				ShowTimestamps: true,
				ShowAge:        false,
				MaxItems:       0,
			},
		},
		{
			name:         "compact format",
			format:       "compact",
			noHeaders:    true,
			noTimestamps: true,
			showAge:      true,
			maxItems:     50,
			expected: &Config{
				Format:         FormatCompact,
				ShowHeaders:    false,
				ShowTimestamps: false,
				ShowAge:        true,
				MaxItems:       50,
			},
		},
		{
			name:         "default format",
			format:       "default",
			noHeaders:    false,
			noTimestamps: false,
			showAge:      false,
			maxItems:     0,
			expected: &Config{
				Format:         FormatDefault,
				ShowHeaders:    true,
				ShowTimestamps: true,
				ShowAge:        false,
				MaxItems:       0,
			},
		},
		{
			name:         "unknown format defaults to default",
			format:       "unknown",
			noHeaders:    false,
			noTimestamps: false,
			showAge:      false,
			maxItems:     0,
			expected: &Config{
				Format:         FormatDefault,
				ShowHeaders:    true,
				ShowTimestamps: true,
				ShowAge:        false,
				MaxItems:       0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetGlobalConfig()
			ConfigureFromFlags(tt.format, tt.noHeaders, tt.noTimestamps, tt.showAge, tt.maxItems)

			cfg := GetConfig()
			assert.Equal(t, tt.expected.Format, cfg.Format)
			assert.Equal(t, tt.expected.ShowHeaders, cfg.ShowHeaders)
			assert.Equal(t, tt.expected.ShowTimestamps, cfg.ShowTimestamps)
			assert.Equal(t, tt.expected.ShowAge, cfg.ShowAge)
			assert.Equal(t, tt.expected.MaxItems, cfg.MaxItems)
		})
	}
}

func TestIsJSON(t *testing.T) {
	original := globalConfig
	defer func() { globalConfig = original }()

	tests := []struct {
		name     string
		format   Format
		expected bool
	}{
		{"JSON format", FormatJSON, true},
		{"Compact format", FormatCompact, false},
		{"Default format", FormatDefault, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			globalConfig = &Config{Format: tt.format}
			assert.Equal(t, tt.expected, IsJSON())
		})
	}
}

func TestIsCompact(t *testing.T) {
	original := globalConfig
	defer func() { globalConfig = original }()

	tests := []struct {
		name     string
		format   Format
		expected bool
	}{
		{"JSON format", FormatJSON, false},
		{"Compact format", FormatCompact, true},
		{"Default format", FormatDefault, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			globalConfig = &Config{Format: tt.format}
			assert.Equal(t, tt.expected, IsCompact())
		})
	}
}

func TestShowAge(t *testing.T) {
	original := globalConfig
	defer func() { globalConfig = original }()

	globalConfig = &Config{ShowAge: true}
	assert.True(t, ShowAge())

	globalConfig = &Config{ShowAge: false}
	assert.False(t, ShowAge())
}

func TestData_JSONFormat(t *testing.T) {
	original := globalConfig
	defer func() { globalConfig = original }()

	globalConfig = &Config{Format: FormatJSON}

	data := map[string]string{"key": "value"}
	output := captureOutput(t, func() {
		Data(data, WithCommand("test"), WithCount(1), WithSummary("test summary"))
	})

	var result Result
	err := json.Unmarshal([]byte(output), &result)
	require.NoError(t, err)

	assert.True(t, result.Success)
	assert.Equal(t, "test", result.Command)
	assert.Equal(t, 1, result.Count)
	assert.Equal(t, "test summary", result.Summary)
}

func TestData_CompactFormat(t *testing.T) {
	original := globalConfig
	defer func() { globalConfig = original }()

	globalConfig = &Config{Format: FormatCompact}

	output := captureOutput(t, func() {
		Data("simple data", WithSummary("Summary line"))
	})

	assert.Contains(t, output, "Summary line")
	assert.Contains(t, output, "simple data")
}

func TestData_DefaultFormat(t *testing.T) {
	original := globalConfig
	defer func() { globalConfig = original }()

	globalConfig = &Config{Format: FormatDefault, ShowHeaders: true}

	data := map[string]int{"count": 42}
	output := captureOutput(t, func() {
		Data(data, WithSummary("Header summary"))
	})

	assert.Contains(t, output, "Header summary")
	assert.Contains(t, output, "42")
}

func TestData_DefaultFormatNoHeaders(t *testing.T) {
	original := globalConfig
	defer func() { globalConfig = original }()

	globalConfig = &Config{Format: FormatDefault, ShowHeaders: false}

	output := captureOutput(t, func() {
		Data("data", WithSummary("Should not show"))
	})

	assert.NotContains(t, output, "Should not show")
}

func TestMessage_JSONFormat(t *testing.T) {
	original := globalConfig
	defer func() { globalConfig = original }()

	globalConfig = &Config{Format: FormatJSON}

	output := captureOutput(t, func() {
		Message("pong")
	})

	var result Result
	err := json.Unmarshal([]byte(output), &result)
	require.NoError(t, err)

	assert.True(t, result.Success)
	assert.Equal(t, "pong", result.Message)
}

func TestMessage_DefaultFormat(t *testing.T) {
	original := globalConfig
	defer func() { globalConfig = original }()

	globalConfig = &Config{Format: FormatDefault}

	output := captureOutput(t, func() {
		Message("Hello World")
	})

	assert.Equal(t, "Hello World\n", output)
}

func TestError_JSONFormat(t *testing.T) {
	original := globalConfig
	defer func() { globalConfig = original }()

	globalConfig = &Config{Format: FormatJSON}

	output := captureOutput(t, func() {
		Error(assert.AnError, "TEST_CODE")
	})

	var result Result
	err := json.Unmarshal([]byte(output), &result)
	require.NoError(t, err)

	assert.False(t, result.Success)
	assert.Contains(t, result.Error, "assert.AnError")
}

func TestError_CompactFormatWithCode(t *testing.T) {
	original := globalConfig
	defer func() { globalConfig = original }()

	globalConfig = &Config{Format: FormatCompact}

	output := captureStderr(t, func() {
		Error(assert.AnError, "ERR001")
	})

	assert.Contains(t, output, "[ERR001]")
}

func TestError_CompactFormatWithoutCode(t *testing.T) {
	original := globalConfig
	defer func() { globalConfig = original }()

	globalConfig = &Config{Format: FormatCompact}

	output := captureStderr(t, func() {
		Error(assert.AnError, "")
	})

	assert.NotContains(t, output, "[")
	assert.Contains(t, output, "assert.AnError")
}

func TestError_DefaultFormatWithCode(t *testing.T) {
	original := globalConfig
	defer func() { globalConfig = original }()

	globalConfig = &Config{Format: FormatDefault}

	output := captureStderr(t, func() {
		Error(assert.AnError, "ERR_TEST")
	})

	assert.Contains(t, output, "Error [ERR_TEST]:")
}

func TestError_DefaultFormatWithoutCode(t *testing.T) {
	original := globalConfig
	defer func() { globalConfig = original }()

	globalConfig = &Config{Format: FormatDefault}

	output := captureStderr(t, func() {
		Error(assert.AnError, "")
	})

	assert.Contains(t, output, "Error:")
	assert.NotContains(t, output, "[]")
}

func TestList_JSONFormat(t *testing.T) {
	original := globalConfig
	defer func() { globalConfig = original }()

	globalConfig = &Config{Format: FormatJSON}

	items := []string{"a", "b", "c"}
	output := captureOutput(t, func() {
		List(items, ListCommand[string]("test-list"))
	})

	var result Result
	err := json.Unmarshal([]byte(output), &result)
	require.NoError(t, err)

	assert.True(t, result.Success)
	assert.Equal(t, "test-list", result.Command)
	assert.Equal(t, 3, result.Count)
}

func TestList_CompactFormatWithFormatter(t *testing.T) {
	original := globalConfig
	defer func() { globalConfig = original }()

	globalConfig = &Config{Format: FormatCompact, ShowHeaders: true}

	items := []int{1, 2, 3}
	output := captureOutput(t, func() {
		List(items,
			ListTitle[int]("Numbers"),
			ListFormatter(func(n int, _ int) string {
				return strings.Repeat("*", n)
			}),
		)
	})

	assert.Contains(t, output, "Numbers: 3")
	assert.Contains(t, output, "*")
	assert.Contains(t, output, "**")
	assert.Contains(t, output, "***")
}

func TestList_MaxItems(t *testing.T) {
	original := globalConfig
	defer func() { globalConfig = original }()

	globalConfig = &Config{Format: FormatCompact, MaxItems: 2, ShowHeaders: true}

	items := []string{"a", "b", "c", "d", "e"}
	output := captureOutput(t, func() {
		List(items, ListTitle[string]("Items"))
	})

	assert.Contains(t, output, "+3 more")
}

func TestList_DefaultFormatMaxItems(t *testing.T) {
	original := globalConfig
	defer func() { globalConfig = original }()

	globalConfig = &Config{Format: FormatDefault, MaxItems: 2, ShowHeaders: true}

	items := []string{"a", "b", "c", "d"}
	output := captureOutput(t, func() {
		List(items, ListTitle[string]("Test"))
	})

	assert.Contains(t, output, "... and 2 more")
}

func TestEntity_JSONFormat(t *testing.T) {
	original := globalConfig
	defer func() { globalConfig = original }()

	globalConfig = &Config{Format: FormatJSON}

	output := captureOutput(t, func() {
		Entity("light.kitchen", "on", map[string]any{"brightness": 255})
	})

	var result Result
	err := json.Unmarshal([]byte(output), &result)
	require.NoError(t, err)

	assert.True(t, result.Success)
	data, ok := result.Data.(map[string]any)
	require.True(t, ok, "expected map[string]any")
	assert.Equal(t, "light.kitchen", data["entity_id"])
	assert.Equal(t, "on", data["state"])
}

func TestEntity_CompactFormat(t *testing.T) {
	original := globalConfig
	defer func() { globalConfig = original }()

	globalConfig = &Config{Format: FormatCompact}

	output := captureOutput(t, func() {
		Entity("sensor.temp", "23.5", nil)
	})

	assert.Equal(t, "sensor.temp=23.5\n", output)
}

func TestTimeline_JSONFormat(t *testing.T) {
	original := globalConfig
	defer func() { globalConfig = original }()

	globalConfig = &Config{Format: FormatJSON}

	entries := []map[string]string{
		{"time": "10:00", "event": "start"},
		{"time": "10:30", "event": "end"},
	}
	output := captureOutput(t, func() {
		Timeline(entries, TimelineCommand[map[string]string]("timeline-test"))
	})

	var result Result
	err := json.Unmarshal([]byte(output), &result)
	require.NoError(t, err)

	assert.True(t, result.Success)
	assert.Equal(t, "timeline-test", result.Command)
	assert.Equal(t, 2, result.Count)
}

func TestTimeline_WithFormatter(t *testing.T) {
	original := globalConfig
	defer func() { globalConfig = original }()

	globalConfig = &Config{Format: FormatCompact, ShowHeaders: true}

	type Event struct {
		Time  string
		Value int
	}
	entries := []Event{{Time: "10:00", Value: 1}, {Time: "11:00", Value: 2}}

	output := captureOutput(t, func() {
		Timeline(entries,
			TimelineTitle[Event]("Events"),
			TimelineFormatter(func(e Event) string {
				return e.Time + ": " + string(rune('0'+e.Value))
			}),
		)
	})

	assert.Contains(t, output, "Events: 2")
	assert.Contains(t, output, "10:00: 1")
	assert.Contains(t, output, "11:00: 2")
}

func TestTimeline_MaxItems(t *testing.T) {
	original := globalConfig
	defer func() { globalConfig = original }()

	globalConfig = &Config{Format: FormatCompact, MaxItems: 1, ShowHeaders: true}

	entries := []string{"a", "b", "c"}
	output := captureOutput(t, func() {
		Timeline(entries, TimelineTitle[string]("Log"))
	})

	assert.Contains(t, output, "+2 more")
}

func TestTimeline_DefaultFormatShowsTotal(t *testing.T) {
	original := globalConfig
	defer func() { globalConfig = original }()

	globalConfig = &Config{Format: FormatDefault, ShowHeaders: true}

	entries := []string{"entry1", "entry2"}
	output := captureOutput(t, func() {
		Timeline(entries, TimelineTitle[string]("Log"))
	})

	assert.Contains(t, output, "Total: 2 entries")
}

func TestFormatTime(t *testing.T) {
	original := globalConfig
	defer func() { globalConfig = original }()

	refTime := time.Date(2024, 6, 15, 10, 30, 45, 0, time.UTC)

	tests := []struct {
		name    string
		format  Format
		checkFn func(t *testing.T, result string)
	}{
		{
			name:   "compact format uses RFC3339",
			format: FormatCompact,
			checkFn: func(t *testing.T, result string) {
				t.Helper()
				assert.Equal(t, refTime.Format(time.RFC3339), result)
			},
		},
		{
			name:   "default format uses local readable",
			format: FormatDefault,
			checkFn: func(t *testing.T, result string) {
				t.Helper()
				// The format should be "2006-01-02 15:04:05"
				assert.Contains(t, result, "2024")
				assert.Contains(t, result, "15")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			globalConfig = &Config{Format: tt.format}
			result := FormatTime(refTime)
			tt.checkFn(t, result)
		})
	}
}

func TestOptions(t *testing.T) {
	t.Parallel()

	t.Run("WithCommand", func(t *testing.T) {
		t.Parallel()
		o := &options{}
		WithCommand("test-cmd")(o)
		assert.Equal(t, "test-cmd", o.command)
	})

	t.Run("WithCount", func(t *testing.T) {
		t.Parallel()
		o := &options{}
		WithCount(42)(o)
		assert.Equal(t, 42, o.count)
	})

	t.Run("WithSummary", func(t *testing.T) {
		t.Parallel()
		o := &options{}
		WithSummary("test summary")(o)
		assert.Equal(t, "test summary", o.summary)
	})
}

func TestListOptions(t *testing.T) {
	t.Parallel()

	t.Run("ListTitle", func(t *testing.T) {
		t.Parallel()
		o := &listOptions[string]{}
		ListTitle[string]("My Title")(o)
		assert.Equal(t, "My Title", o.title)
	})

	t.Run("ListCommand", func(t *testing.T) {
		t.Parallel()
		o := &listOptions[string]{}
		ListCommand[string]("list-cmd")(o)
		assert.Equal(t, "list-cmd", o.command)
	})

	t.Run("ListFormatter", func(t *testing.T) {
		t.Parallel()
		o := &listOptions[int]{}
		formatter := func(_ int, _ int) string { return "formatted" }
		ListFormatter(formatter)(o)
		assert.NotNil(t, o.formatter)
		assert.Equal(t, "formatted", o.formatter(1, 0))
	})
}

func TestTimelineOptions(t *testing.T) {
	t.Parallel()

	t.Run("TimelineTitle", func(t *testing.T) {
		t.Parallel()
		o := &timelineOptions[string]{}
		TimelineTitle[string]("Timeline Title")(o)
		assert.Equal(t, "Timeline Title", o.title)
	})

	t.Run("TimelineCommand", func(t *testing.T) {
		t.Parallel()
		o := &timelineOptions[string]{}
		TimelineCommand[string]("tl-cmd")(o)
		assert.Equal(t, "tl-cmd", o.command)
	})

	t.Run("TimelineFormatter", func(t *testing.T) {
		t.Parallel()
		o := &timelineOptions[string]{}
		formatter := func(s string) string { return "formatted: " + s }
		TimelineFormatter(formatter)(o)
		assert.NotNil(t, o.formatter)
		assert.Equal(t, "formatted: test", o.formatter("test"))
	})
}

func TestPrintCompact_EntityState(t *testing.T) {
	original := globalConfig
	defer func() { globalConfig = original }()

	globalConfig = &Config{Format: FormatCompact}

	output := captureOutput(t, func() {
		printCompact(map[string]any{
			"entity_id": "light.test",
			"state":     "on",
		})
	})

	assert.Equal(t, "light.test=on\n", output)
}

func TestPrintCompact_TraceInfo(t *testing.T) {
	original := globalConfig
	defer func() { globalConfig = original }()

	globalConfig = &Config{Format: FormatCompact}

	output := captureOutput(t, func() {
		printCompact(map[string]any{
			"item_id": "automation.test",
			"run_id":  "abc123",
			"timestamp": map[string]any{
				"start": "2024-06-15T10:00:00Z",
			},
		})
	})

	assert.Contains(t, output, "automation.test")
	assert.Contains(t, output, "abc123")
	assert.Contains(t, output, "2024-06-15T10:00:00Z")
}

func TestPrintCompact_Slice(t *testing.T) {
	original := globalConfig
	defer func() { globalConfig = original }()

	globalConfig = &Config{Format: FormatCompact}

	output := captureOutput(t, func() {
		printCompact([]any{"item1", "item2"})
	})

	assert.Contains(t, output, "item1")
	assert.Contains(t, output, "item2")
}

func TestResult_JSONMarshal(t *testing.T) {
	t.Parallel()

	result := Result{
		Success: true,
		Command: "test",
		Data:    map[string]int{"count": 5},
		Count:   5,
		Summary: "5 items found",
		Message: "OK",
	}

	data, err := json.Marshal(result)
	require.NoError(t, err)

	var decoded Result
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.True(t, decoded.Success)
	assert.Equal(t, "test", decoded.Command)
	assert.Equal(t, 5, decoded.Count)
	assert.Equal(t, "5 items found", decoded.Summary)
	assert.Equal(t, "OK", decoded.Message)
}

func TestResult_ErrorMarshal(t *testing.T) {
	t.Parallel()

	result := Result{
		Success: false,
		Error:   "Entity not found",
	}

	data, err := json.Marshal(result)
	require.NoError(t, err)

	assert.Contains(t, string(data), `"success":false`)
	assert.Contains(t, string(data), `"error":"Entity not found"`)
}

// TestPrintCompact_Struct tests compact output for typed structs (like HAState)
func TestPrintCompact_Struct(t *testing.T) {
	original := globalConfig
	defer func() { globalConfig = original }()

	globalConfig = &Config{Format: FormatCompact}

	// Define a struct similar to HAState
	type TestState struct {
		EntityID   string         `json:"entity_id"`
		State      string         `json:"state"`
		Attributes map[string]any `json:"attributes,omitempty"`
	}

	state := TestState{
		EntityID:   "light.kitchen",
		State:      "on",
		Attributes: map[string]any{"brightness": 255},
	}

	output := captureOutput(t, func() {
		printCompact(state)
	})

	assert.Equal(t, "light.kitchen=on\n", output)
}

func TestStructToMap(t *testing.T) {
	t.Parallel()

	t.Run("valid struct", func(t *testing.T) {
		t.Parallel()
		type TestStruct struct {
			Name  string `json:"name"`
			Value int    `json:"value"`
		}
		s := TestStruct{Name: "test", Value: 42}
		m := structToMap(s)

		require.NotNil(t, m)
		assert.Equal(t, "test", m["name"])
		assert.Equal(t, float64(42), m["value"]) // JSON numbers are float64
	})

	t.Run("nil value", func(t *testing.T) {
		t.Parallel()
		m := structToMap(nil)
		assert.Nil(t, m)
	})

	t.Run("non-struct primitive", func(t *testing.T) {
		t.Parallel()
		m := structToMap("string value")
		assert.Nil(t, m)
	})

	t.Run("slice returns nil", func(t *testing.T) {
		t.Parallel()
		m := structToMap([]string{"a", "b"})
		assert.Nil(t, m)
	})
}

func TestPrintCompactMap(t *testing.T) {
	original := globalConfig
	defer func() { globalConfig = original }()

	globalConfig = &Config{Format: FormatCompact}

	t.Run("entity state format", func(t *testing.T) {
		output := captureOutput(t, func() {
			printCompactMap(map[string]any{
				"entity_id": "sensor.temp",
				"state":     "23.5",
			})
		})
		assert.Equal(t, "sensor.temp=23.5\n", output)
	})

	t.Run("trace info format", func(t *testing.T) {
		output := captureOutput(t, func() {
			printCompactMap(map[string]any{
				"item_id": "test_automation",
				"run_id":  "run123",
				"timestamp": map[string]any{
					"start": "2024-01-01T00:00:00Z",
				},
			})
		})
		assert.Contains(t, output, "test_automation")
		assert.Contains(t, output, "run123")
		assert.Contains(t, output, "2024-01-01T00:00:00Z")
	})

	t.Run("generic map format", func(t *testing.T) {
		output := captureOutput(t, func() {
			printCompactMap(map[string]any{
				"key1": "value1",
				"key2": 42,
			})
		})
		assert.Contains(t, output, "key1=value1")
		assert.Contains(t, output, "key2=42")
	})
}

func TestData_CompactFormat_WithStruct(t *testing.T) {
	original := globalConfig
	defer func() { globalConfig = original }()

	globalConfig = &Config{Format: FormatCompact}

	// Simulate HAState struct behavior
	type EntityState struct {
		EntityID    string         `json:"entity_id"`
		State       string         `json:"state"`
		Attributes  map[string]any `json:"attributes,omitempty"`
		LastChanged string         `json:"last_changed,omitempty"`
	}

	state := EntityState{
		EntityID:    "switch.garage",
		State:       "off",
		LastChanged: "2024-01-01T12:00:00Z",
	}

	output := captureOutput(t, func() {
		Data(state, WithCommand("state"))
	})

	assert.Equal(t, "switch.garage=off\n", output)
}

// Note: Concurrent config access is not tested because the CLI runs
// single-threaded - commands execute sequentially, not in parallel.
// The global config pattern is appropriate for this use case.

// Benchmark for Data output
func BenchmarkData_JSON(b *testing.B) {
	original := globalConfig
	defer func() { globalConfig = original }()

	globalConfig = &Config{Format: FormatJSON}

	// Redirect stdout to discard
	old := os.Stdout
	devNull, err := os.Open(os.DevNull)
	if err != nil {
		b.Fatal(err)
	}
	os.Stdout = devNull
	defer func() { os.Stdout = old }()

	data := map[string]any{
		"entity_id":  "sensor.temperature",
		"state":      "23.5",
		"attributes": map[string]any{"unit": "°C"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Data(data, WithCommand("state"), WithCount(1))
	}
}
