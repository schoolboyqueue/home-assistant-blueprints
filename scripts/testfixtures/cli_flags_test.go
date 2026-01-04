// Package testfixtures provides CLI flag unit tests.
// This file tests CLI flag parsing logic using a mock CLI app.
// These are NOT integration tests - they test flag parsing in isolation.
//
// For true integration tests that connect to a real Home Assistant instance,
// see scripts/ha-ws-client-go/internal/handlers/handlers_integration_test.go
package testfixtures

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v3"
)

// =============================================================================
// Test Helpers
// =============================================================================

// OutputConfig represents the parsed output configuration for testing.
type OutputConfig struct {
	Format       string
	NoHeaders    bool
	NoTimestamps bool
	ShowAge      bool
	MaxItems     int
	FromTime     string
	ToTime       string
}

// createTestApp creates a minimal CLI app that captures flag values for testing.
func createTestApp(configOut *OutputConfig) *cli.Command {
	return &cli.Command{
		Name: "ha-ws-client",
		Flags: []cli.Flag{
			// Time filtering flags
			&cli.StringFlag{
				Name:  "from",
				Usage: "Start time for filtering (YYYY-MM-DD or YYYY-MM-DD HH:MM)",
			},
			&cli.StringFlag{
				Name:  "to",
				Usage: "End time for filtering (YYYY-MM-DD or YYYY-MM-DD HH:MM)",
			},
			// Output format flags
			&cli.StringFlag{
				Name:    "output",
				Aliases: []string{"format"},
				Value:   "default",
				Usage:   "Output format: json, compact, or default",
			},
			&cli.BoolFlag{
				Name:  "compact",
				Usage: "Use compact output format (single-line entries)",
			},
			&cli.BoolFlag{
				Name:  "json",
				Usage: "Use JSON output format (machine-readable)",
			},
			&cli.BoolFlag{
				Name:  "no-headers",
				Usage: "Hide section headers and titles",
			},
			&cli.BoolFlag{
				Name:  "no-timestamps",
				Usage: "Hide timestamps in output",
			},
			&cli.BoolFlag{
				Name:  "show-age",
				Usage: "Show last_updated age for states-filter command",
			},
			&cli.IntFlag{
				Name:  "max-items",
				Usage: "Limit output to N items (0 = unlimited)",
			},
		},
		Commands: []*cli.Command{
			{
				Name:  "test",
				Usage: "Test command for flag validation",
				Action: func(_ context.Context, cmd *cli.Command) error {
					// Determine output format
					outputFormat := cmd.String("output")
					if cmd.Bool("json") {
						outputFormat = "json"
					} else if cmd.Bool("compact") {
						outputFormat = "compact"
					}

					// Capture parsed values
					configOut.Format = outputFormat
					configOut.NoHeaders = cmd.Bool("no-headers")
					configOut.NoTimestamps = cmd.Bool("no-timestamps")
					configOut.ShowAge = cmd.Bool("show-age")
					configOut.MaxItems = cmd.Int("max-items")
					configOut.FromTime = cmd.String("from")
					configOut.ToTime = cmd.String("to")

					return nil
				},
			},
		},
	}
}

// runCLI runs the CLI app with the given arguments and returns the parsed config.
func runCLI(t *testing.T, args ...string) OutputConfig {
	t.Helper()
	var config OutputConfig
	app := createTestApp(&config)

	// Prepend program name
	fullArgs := append([]string{"ha-ws-client"}, args...)
	err := app.Run(context.Background(), fullArgs)
	require.NoError(t, err)

	return config
}

// =============================================================================
// Output Format Flag Tests
// =============================================================================

func TestCLIFlags_OutputFormatDefault(t *testing.T) {
	t.Parallel()

	config := runCLI(t, "test")

	assert.Equal(t, "default", config.Format)
}

func TestCLIFlags_OutputJSON_LongFlag(t *testing.T) {
	t.Parallel()

	config := runCLI(t, "--output", "json", "test")

	assert.Equal(t, "json", config.Format)
}

func TestCLIFlags_OutputJSON_ShortFlag(t *testing.T) {
	t.Parallel()

	config := runCLI(t, "--json", "test")

	assert.Equal(t, "json", config.Format)
}

func TestCLIFlags_OutputCompact_LongFlag(t *testing.T) {
	t.Parallel()

	config := runCLI(t, "--output", "compact", "test")

	assert.Equal(t, "compact", config.Format)
}

func TestCLIFlags_OutputCompact_ShortFlag(t *testing.T) {
	t.Parallel()

	config := runCLI(t, "--compact", "test")

	assert.Equal(t, "compact", config.Format)
}

func TestCLIFlags_FormatAlias(t *testing.T) {
	t.Parallel()

	config := runCLI(t, "--format", "json", "test")

	assert.Equal(t, "json", config.Format)
}

// =============================================================================
// Output Format Combination Tests
// =============================================================================

func TestCLIFlags_JSON_NoHeaders(t *testing.T) {
	t.Parallel()

	config := runCLI(t, "--output", "json", "--no-headers", "test")

	assert.Equal(t, "json", config.Format)
	assert.True(t, config.NoHeaders)
}

func TestCLIFlags_JSON_NoTimestamps(t *testing.T) {
	t.Parallel()

	config := runCLI(t, "--output", "json", "--no-timestamps", "test")

	assert.Equal(t, "json", config.Format)
	assert.True(t, config.NoTimestamps)
}

func TestCLIFlags_JSON_NoHeaders_NoTimestamps(t *testing.T) {
	t.Parallel()

	config := runCLI(t, "--output", "json", "--no-headers", "--no-timestamps", "test")

	assert.Equal(t, "json", config.Format)
	assert.True(t, config.NoHeaders)
	assert.True(t, config.NoTimestamps)
}

func TestCLIFlags_Compact_NoHeaders(t *testing.T) {
	t.Parallel()

	config := runCLI(t, "--compact", "--no-headers", "test")

	assert.Equal(t, "compact", config.Format)
	assert.True(t, config.NoHeaders)
}

func TestCLIFlags_Compact_ShowAge(t *testing.T) {
	t.Parallel()

	config := runCLI(t, "--compact", "--show-age", "test")

	assert.Equal(t, "compact", config.Format)
	assert.True(t, config.ShowAge)
}

func TestCLIFlags_Compact_MaxItems(t *testing.T) {
	t.Parallel()

	config := runCLI(t, "--compact", "--max-items", "5", "test")

	assert.Equal(t, "compact", config.Format)
	assert.Equal(t, 5, config.MaxItems)
}

func TestCLIFlags_Default_AllDisplayOptions(t *testing.T) {
	t.Parallel()

	config := runCLI(t, "--no-headers", "--no-timestamps", "--show-age", "--max-items", "10", "test")

	assert.Equal(t, "default", config.Format)
	assert.True(t, config.NoHeaders)
	assert.True(t, config.NoTimestamps)
	assert.True(t, config.ShowAge)
	assert.Equal(t, 10, config.MaxItems)
}

// =============================================================================
// Time Filtering Flag Tests
// =============================================================================

func TestCLIFlags_FromTime(t *testing.T) {
	t.Parallel()

	config := runCLI(t, "--from", "2024-01-01", "test")

	assert.Equal(t, "2024-01-01", config.FromTime)
}

func TestCLIFlags_ToTime(t *testing.T) {
	t.Parallel()

	config := runCLI(t, "--to", "2024-12-31", "test")

	assert.Equal(t, "2024-12-31", config.ToTime)
}

func TestCLIFlags_FromTo_DateOnly(t *testing.T) {
	t.Parallel()

	config := runCLI(t, "--from", "2024-01-01", "--to", "2024-01-31", "test")

	assert.Equal(t, "2024-01-01", config.FromTime)
	assert.Equal(t, "2024-01-31", config.ToTime)
}

func TestCLIFlags_FromTo_DateTimeWithSpace(t *testing.T) {
	t.Parallel()

	// Note: Shell would quote this, but we pass directly
	config := runCLI(t, "--from", "2024-01-01 10:00", "--to", "2024-01-01 18:00", "test")

	assert.Equal(t, "2024-01-01 10:00", config.FromTime)
	assert.Equal(t, "2024-01-01 18:00", config.ToTime)
}

func TestCLIFlags_FromTo_RFC3339(t *testing.T) {
	t.Parallel()

	config := runCLI(t, "--from", "2024-01-01T10:00:00Z", "--to", "2024-01-01T18:00:00Z", "test")

	assert.Equal(t, "2024-01-01T10:00:00Z", config.FromTime)
	assert.Equal(t, "2024-01-01T18:00:00Z", config.ToTime)
}

// =============================================================================
// Time + Output Format Combination Tests
// =============================================================================

func TestCLIFlags_FromTo_OutputCompact(t *testing.T) {
	t.Parallel()

	config := runCLI(t, "--from", "2024-01-01", "--to", "2024-01-31", "--output", "compact", "test")

	assert.Equal(t, "2024-01-01", config.FromTime)
	assert.Equal(t, "2024-01-31", config.ToTime)
	assert.Equal(t, "compact", config.Format)
}

func TestCLIFlags_FromTo_OutputJSON(t *testing.T) {
	t.Parallel()

	config := runCLI(t, "--from", "2024-01-01", "--to", "2024-01-31", "--output", "json", "test")

	assert.Equal(t, "2024-01-01", config.FromTime)
	assert.Equal(t, "2024-01-31", config.ToTime)
	assert.Equal(t, "json", config.Format)
}

func TestCLIFlags_FromTo_Compact_NoHeaders(t *testing.T) {
	t.Parallel()

	config := runCLI(t, "--from", "2024-01-01", "--to", "2024-01-31", "--compact", "--no-headers", "test")

	assert.Equal(t, "2024-01-01", config.FromTime)
	assert.Equal(t, "2024-01-31", config.ToTime)
	assert.Equal(t, "compact", config.Format)
	assert.True(t, config.NoHeaders)
}

func TestCLIFlags_FullCombination_HistoryUseCase(t *testing.T) {
	t.Parallel()

	// Simulates: ha-ws-client --from 2024-01-01 --to 2024-01-31 --output compact --no-headers --max-items 100 history sensor.temperature
	config := runCLI(t,
		"--from", "2024-01-01",
		"--to", "2024-01-31",
		"--output", "compact",
		"--no-headers",
		"--max-items", "100",
		"test",
	)

	assert.Equal(t, "2024-01-01", config.FromTime)
	assert.Equal(t, "2024-01-31", config.ToTime)
	assert.Equal(t, "compact", config.Format)
	assert.True(t, config.NoHeaders)
	assert.Equal(t, 100, config.MaxItems)
}

func TestCLIFlags_FullCombination_StatesFilterUseCase(t *testing.T) {
	t.Parallel()

	// Simulates: ha-ws-client --output json --show-age --max-items 50 states-filter "sensor.*"
	config := runCLI(t,
		"--output", "json",
		"--show-age",
		"--max-items", "50",
		"test",
	)

	assert.Equal(t, "json", config.Format)
	assert.True(t, config.ShowAge)
	assert.Equal(t, 50, config.MaxItems)
}

func TestCLIFlags_FullCombination_LogbookUseCase(t *testing.T) {
	t.Parallel()

	// Simulates: ha-ws-client --from 2024-01-01 --to 2024-01-02 --output json --no-timestamps logbook
	config := runCLI(t,
		"--from", "2024-01-01",
		"--to", "2024-01-02",
		"--output", "json",
		"--no-timestamps",
		"test",
	)

	assert.Equal(t, "2024-01-01", config.FromTime)
	assert.Equal(t, "2024-01-02", config.ToTime)
	assert.Equal(t, "json", config.Format)
	assert.True(t, config.NoTimestamps)
}

// =============================================================================
// Flag Priority Tests
// =============================================================================

func TestCLIFlags_BoolFlagOverridesOutputFlag_JSON(t *testing.T) {
	t.Parallel()

	// --json should override --output default
	config := runCLI(t, "--output", "default", "--json", "test")

	assert.Equal(t, "json", config.Format)
}

func TestCLIFlags_BoolFlagOverridesOutputFlag_Compact(t *testing.T) {
	t.Parallel()

	// --compact should override --output default
	config := runCLI(t, "--output", "default", "--compact", "test")

	assert.Equal(t, "compact", config.Format)
}

func TestCLIFlags_LastBoolFlagWins_JSONThenCompact(t *testing.T) {
	t.Parallel()

	// When both are specified, the code checks --json first, then --compact
	// So --json wins if both are set
	config := runCLI(t, "--json", "--compact", "test")

	// Based on the implementation in main.go: json is checked first
	assert.Equal(t, "json", config.Format)
}

// =============================================================================
// Edge Cases
// =============================================================================

func TestCLIFlags_MaxItemsZero(t *testing.T) {
	t.Parallel()

	config := runCLI(t, "--max-items", "0", "test")

	assert.Equal(t, 0, config.MaxItems)
}

func TestCLIFlags_MaxItemsLarge(t *testing.T) {
	t.Parallel()

	config := runCLI(t, "--max-items", "10000", "test")

	assert.Equal(t, 10000, config.MaxItems)
}

func TestCLIFlags_EmptyFromTo(t *testing.T) {
	t.Parallel()

	config := runCLI(t, "test")

	assert.Empty(t, config.FromTime)
	assert.Empty(t, config.ToTime)
}

func TestCLIFlags_OnlyFrom(t *testing.T) {
	t.Parallel()

	config := runCLI(t, "--from", "2024-01-01", "test")

	assert.Equal(t, "2024-01-01", config.FromTime)
	assert.Empty(t, config.ToTime)
}

func TestCLIFlags_OnlyTo(t *testing.T) {
	t.Parallel()

	config := runCLI(t, "--to", "2024-12-31", "test")

	assert.Empty(t, config.FromTime)
	assert.Equal(t, "2024-12-31", config.ToTime)
}

// =============================================================================
// Table-Driven Comprehensive Flag Combination Tests
// =============================================================================

func TestCLIFlags_TableDriven_FlagCombinations(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		args     []string
		expected OutputConfig
	}{
		{
			name: "default_everything",
			args: []string{"test"},
			expected: OutputConfig{
				Format: "default",
			},
		},
		{
			name: "json_only",
			args: []string{"--json", "test"},
			expected: OutputConfig{
				Format: "json",
			},
		},
		{
			name: "compact_only",
			args: []string{"--compact", "test"},
			expected: OutputConfig{
				Format: "compact",
			},
		},
		{
			name: "json_with_no_headers",
			args: []string{"--json", "--no-headers", "test"},
			expected: OutputConfig{
				Format:    "json",
				NoHeaders: true,
			},
		},
		{
			name: "compact_with_no_timestamps",
			args: []string{"--compact", "--no-timestamps", "test"},
			expected: OutputConfig{
				Format:       "compact",
				NoTimestamps: true,
			},
		},
		{
			name: "json_no_headers_no_timestamps",
			args: []string{"--json", "--no-headers", "--no-timestamps", "test"},
			expected: OutputConfig{
				Format:       "json",
				NoHeaders:    true,
				NoTimestamps: true,
			},
		},
		{
			name: "compact_show_age",
			args: []string{"--compact", "--show-age", "test"},
			expected: OutputConfig{
				Format:  "compact",
				ShowAge: true,
			},
		},
		{
			name: "default_max_items_5",
			args: []string{"--max-items", "5", "test"},
			expected: OutputConfig{
				Format:   "default",
				MaxItems: 5,
			},
		},
		{
			name: "from_to_json",
			args: []string{"--from", "2024-01-01", "--to", "2024-12-31", "--json", "test"},
			expected: OutputConfig{
				Format:   "json",
				FromTime: "2024-01-01",
				ToTime:   "2024-12-31",
			},
		},
		{
			name: "from_to_compact_no_headers_max_items",
			args: []string{"--from", "2024-06-01", "--to", "2024-06-30", "--compact", "--no-headers", "--max-items", "25", "test"},
			expected: OutputConfig{
				Format:    "compact",
				FromTime:  "2024-06-01",
				ToTime:    "2024-06-30",
				NoHeaders: true,
				MaxItems:  25,
			},
		},
		{
			name: "all_flags_combined",
			args: []string{
				"--from", "2024-01-01",
				"--to", "2024-12-31",
				"--output", "compact",
				"--no-headers",
				"--no-timestamps",
				"--show-age",
				"--max-items", "100",
				"test",
			},
			expected: OutputConfig{
				Format:       "compact",
				FromTime:     "2024-01-01",
				ToTime:       "2024-12-31",
				NoHeaders:    true,
				NoTimestamps: true,
				ShowAge:      true,
				MaxItems:     100,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			config := runCLI(t, tt.args...)

			assert.Equal(t, tt.expected.Format, config.Format, "Format mismatch")
			assert.Equal(t, tt.expected.NoHeaders, config.NoHeaders, "NoHeaders mismatch")
			assert.Equal(t, tt.expected.NoTimestamps, config.NoTimestamps, "NoTimestamps mismatch")
			assert.Equal(t, tt.expected.ShowAge, config.ShowAge, "ShowAge mismatch")
			assert.Equal(t, tt.expected.MaxItems, config.MaxItems, "MaxItems mismatch")
			assert.Equal(t, tt.expected.FromTime, config.FromTime, "FromTime mismatch")
			assert.Equal(t, tt.expected.ToTime, config.ToTime, "ToTime mismatch")
		})
	}
}

// =============================================================================
// ConfigureFromFlags Unit Tests
// =============================================================================

// CLIOutputConfig mirrors the output.Config from ha-ws-client-go
// This allows us to test flag parsing without importing the actual package
type CLIOutputConfig struct {
	Format         string
	ShowTimestamps bool
	ShowHeaders    bool
	ShowAge        bool
	MaxItems       int
}

// configureFromFlags mirrors output.ConfigureFromFlags behavior
func configureFromFlags(format string, noHeaders, noTimestamps, showAge bool, maxItems int) CLIOutputConfig {
	cfg := CLIOutputConfig{
		ShowTimestamps: true,
		ShowHeaders:    true,
	}

	switch format {
	case "json":
		cfg.Format = "json"
	case "compact":
		cfg.Format = "compact"
	default:
		cfg.Format = "default"
	}

	cfg.ShowHeaders = !noHeaders
	cfg.ShowTimestamps = !noTimestamps
	cfg.ShowAge = showAge
	cfg.MaxItems = maxItems

	return cfg
}

func TestConfigureFromFlags_Unit(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		format       string
		noHeaders    bool
		noTimestamps bool
		showAge      bool
		maxItems     int
		expected     CLIOutputConfig
	}{
		{
			name:         "json_no_headers",
			format:       "json",
			noHeaders:    true,
			noTimestamps: false,
			showAge:      false,
			maxItems:     0,
			expected: CLIOutputConfig{
				Format:         "json",
				ShowHeaders:    false,
				ShowTimestamps: true,
				ShowAge:        false,
				MaxItems:       0,
			},
		},
		{
			name:         "compact_no_timestamps",
			format:       "compact",
			noHeaders:    false,
			noTimestamps: true,
			showAge:      false,
			maxItems:     0,
			expected: CLIOutputConfig{
				Format:         "compact",
				ShowHeaders:    true,
				ShowTimestamps: false,
				ShowAge:        false,
				MaxItems:       0,
			},
		},
		{
			name:         "default_show_age_max_items",
			format:       "default",
			noHeaders:    false,
			noTimestamps: false,
			showAge:      true,
			maxItems:     50,
			expected: CLIOutputConfig{
				Format:         "default",
				ShowHeaders:    true,
				ShowTimestamps: true,
				ShowAge:        true,
				MaxItems:       50,
			},
		},
		{
			name:         "compact_all_display_options",
			format:       "compact",
			noHeaders:    true,
			noTimestamps: true,
			showAge:      true,
			maxItems:     100,
			expected: CLIOutputConfig{
				Format:         "compact",
				ShowHeaders:    false,
				ShowTimestamps: false,
				ShowAge:        true,
				MaxItems:       100,
			},
		},
		{
			name:         "unknown_format_defaults",
			format:       "unknown",
			noHeaders:    false,
			noTimestamps: false,
			showAge:      false,
			maxItems:     0,
			expected: CLIOutputConfig{
				Format:         "default",
				ShowHeaders:    true,
				ShowTimestamps: true,
				ShowAge:        false,
				MaxItems:       0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cfg := configureFromFlags(tt.format, tt.noHeaders, tt.noTimestamps, tt.showAge, tt.maxItems)

			assert.Equal(t, tt.expected.Format, cfg.Format)
			assert.Equal(t, tt.expected.ShowHeaders, cfg.ShowHeaders)
			assert.Equal(t, tt.expected.ShowTimestamps, cfg.ShowTimestamps)
			assert.Equal(t, tt.expected.ShowAge, cfg.ShowAge)
			assert.Equal(t, tt.expected.MaxItems, cfg.MaxItems)
		})
	}
}

// =============================================================================
// Output Format Behavior Unit Tests
// =============================================================================

func TestOutputBehavior_JSONFormat(t *testing.T) {
	t.Parallel()

	// Simulate JSON output structure
	type JSONResult struct {
		Success bool   `json:"success"`
		Command string `json:"command,omitempty"`
		Data    any    `json:"data,omitempty"`
		Count   int    `json:"count,omitempty"`
	}

	// Test that JSON output is always valid JSON regardless of other flags
	testCases := []struct {
		name string
		data any
	}{
		{"simple_string", "test"},
		{"map", map[string]any{"key": "value"}},
		{"slice", []string{"a", "b", "c"}},
		{"nested", map[string]any{"outer": map[string]any{"inner": "value"}}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := JSONResult{
				Success: true,
				Command: "test",
				Data:    tc.data,
			}

			jsonBytes, err := json.Marshal(result)
			require.NoError(t, err)

			// Verify it's valid JSON
			var parsed JSONResult
			err = json.Unmarshal(jsonBytes, &parsed)
			require.NoError(t, err)

			assert.True(t, parsed.Success)
			assert.Equal(t, "test", parsed.Command)
		})
	}
}

func TestOutputBehavior_CompactFormat(t *testing.T) {
	t.Parallel()

	// Simulate compact output formatting
	formatCompactEntity := func(entityID, state string) string {
		return fmt.Sprintf("%s=%s", entityID, state)
	}

	tests := []struct {
		entityID string
		state    string
		expected string
	}{
		{"light.kitchen", "on", "light.kitchen=on"},
		{"sensor.temperature", "23.5", "sensor.temperature=23.5"},
		{"binary_sensor.motion", "off", "binary_sensor.motion=off"},
	}

	for _, tc := range tests {
		t.Run(tc.entityID, func(t *testing.T) {
			t.Parallel()
			result := formatCompactEntity(tc.entityID, tc.state)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestOutputBehavior_MaxItemsTruncation(t *testing.T) {
	t.Parallel()

	items := make([]string, 100)
	for i := 0; i < 100; i++ {
		items[i] = fmt.Sprintf("item_%d", i)
	}

	testCases := []struct {
		maxItems      int
		expectedCount int
		hasMore       bool
	}{
		{0, 100, false},   // 0 means unlimited
		{10, 10, true},    // truncate to 10
		{50, 50, true},    // truncate to 50
		{100, 100, false}, // exactly matches
		{150, 100, false}, // more than available
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("max_%d", tc.maxItems), func(t *testing.T) {
			t.Parallel()

			displayItems := items
			if tc.maxItems > 0 && len(items) > tc.maxItems {
				displayItems = items[:tc.maxItems]
			}

			assert.Equal(t, tc.expectedCount, len(displayItems))

			hasMore := tc.maxItems > 0 && len(items) > tc.maxItems
			assert.Equal(t, tc.hasMore, hasMore)
		})
	}
}

// =============================================================================
// Time Parsing Unit Tests
// =============================================================================

func TestTimeParsing_FlexibleFormats(t *testing.T) {
	t.Parallel()

	// Test that various time formats can be parsed
	// This mirrors the ParseFlexibleDate function behavior
	formats := []string{
		time.RFC3339,
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
		"2006-01-02 15:04",
		"2006-01-02",
	}

	parseFlexibleDate := func(s string) (time.Time, error) {
		for _, format := range formats {
			if t, err := time.Parse(format, s); err == nil {
				return t, nil
			}
		}
		return time.Time{}, fmt.Errorf("unable to parse date: %s", s)
	}

	tests := []struct {
		input       string
		expectError bool
	}{
		// Valid formats
		{"2024-01-15T10:30:45Z", false},
		{"2024-01-15T10:30:45+05:00", false},
		{"2024-01-15T10:30:45", false},
		{"2024-01-15 10:30:45", false},
		{"2024-01-15 10:30", false},
		{"2024-01-15", false},

		// Invalid formats
		{"not-a-date", true},
		{"01/15/2024", true},
		{"2024/01/15", true},
		{"", true},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			t.Parallel()

			_, err := parseFlexibleDate(tc.input)
			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestTimeRange_FromToValidation(t *testing.T) {
	t.Parallel()

	// Test that from/to time ranges are logically valid
	tests := []struct {
		name  string
		from  string
		to    string
		valid bool
	}{
		{"same_day", "2024-01-15", "2024-01-15", true},
		{"normal_range", "2024-01-01", "2024-01-31", true},
		{"reversed_invalid", "2024-12-31", "2024-01-01", false}, // from > to
		{"only_from", "2024-01-01", "", true},
		{"only_to", "", "2024-12-31", true},
		{"neither", "", "", true},
	}

	parseDate := func(s string) time.Time {
		if s == "" {
			return time.Time{}
		}
		t, _ := time.Parse("2006-01-02", s)
		return t
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			from := parseDate(tc.from)
			to := parseDate(tc.to)

			// A range is valid if:
			// - Both are empty
			// - Only one is set
			// - from <= to
			isValid := from.IsZero() || to.IsZero() || !from.After(to)
			assert.Equal(t, tc.valid, isValid)
		})
	}
}

// =============================================================================
// Benchmark Tests
// =============================================================================

func BenchmarkFlagParsing(b *testing.B) {
	var config OutputConfig
	app := createTestApp(&config)

	args := []string{
		"ha-ws-client",
		"--from", "2024-01-01",
		"--to", "2024-12-31",
		"--output", "json",
		"--no-headers",
		"--max-items", "100",
		"test",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = app.Run(context.Background(), args)
	}
}

func BenchmarkConfigureFromFlags(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = configureFromFlags("json", true, false, true, 100)
	}
}
