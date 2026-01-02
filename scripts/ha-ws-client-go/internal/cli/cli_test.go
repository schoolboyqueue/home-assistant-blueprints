package cli

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/home-assistant-blueprints/ha-ws-client-go/internal/types"
)

func TestParseFlexibleDate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		input       string
		expected    time.Time
		expectError bool
	}{
		{
			name:     "RFC3339 with timezone",
			input:    "2024-06-15T10:30:45Z",
			expected: time.Date(2024, 6, 15, 10, 30, 45, 0, time.UTC),
		},
		{
			name:     "RFC3339 with offset",
			input:    "2024-06-15T10:30:45+05:00",
			expected: time.Date(2024, 6, 15, 10, 30, 45, 0, time.FixedZone("", 5*3600)),
		},
		{
			name:     "ISO format without timezone",
			input:    "2024-06-15T10:30:45",
			expected: time.Date(2024, 6, 15, 10, 30, 45, 0, time.UTC),
		},
		{
			name:     "Date and time with space",
			input:    "2024-06-15 10:30:45",
			expected: time.Date(2024, 6, 15, 10, 30, 45, 0, time.UTC),
		},
		{
			name:     "Date and time without seconds",
			input:    "2024-06-15 10:30",
			expected: time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC),
		},
		{
			name:     "Date only",
			input:    "2024-06-15",
			expected: time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC),
		},
		{
			name:        "Invalid format",
			input:       "not-a-date",
			expectError: true,
		},
		{
			name:        "Empty string",
			input:       "",
			expectError: true,
		},
		{
			name:        "US format not supported",
			input:       "06/15/2024",
			expectError: true,
		},
		{
			name:        "Partial date",
			input:       "2024-06",
			expectError: true,
		},
		{
			name:     "Midnight",
			input:    "2024-01-01T00:00:00Z",
			expected: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "End of day",
			input:    "2024-12-31T23:59:59Z",
			expected: time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result, err := ParseFlexibleDate(tt.input)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "unable to parse date")
			} else {
				require.NoError(t, err)
				assert.True(t, tt.expected.Equal(result),
					"expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestParse(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		args      []string
		checkFunc func(t *testing.T, f *types.CLIFlags, err error)
	}{
		{
			name: "bare help command",
			args: []string{"help"},
			checkFunc: func(t *testing.T, f *types.CLIFlags, err error) {
				t.Helper()
				require.NoError(t, err)
				assert.True(t, f.Help)
				assert.Empty(t, f.Args)
			},
		},
		{
			name: "bare version command",
			args: []string{"version"},
			checkFunc: func(t *testing.T, f *types.CLIFlags, err error) {
				t.Helper()
				require.NoError(t, err)
				assert.True(t, f.Version)
				assert.Empty(t, f.Args)
			},
		},
		{
			name: "help flag",
			args: []string{"--help"},
			checkFunc: func(t *testing.T, f *types.CLIFlags, err error) {
				t.Helper()
				require.NoError(t, err)
				assert.True(t, f.Help)
			},
		},
		{
			name: "help shorthand flag",
			args: []string{"-h"},
			checkFunc: func(t *testing.T, f *types.CLIFlags, err error) {
				t.Helper()
				require.NoError(t, err)
				assert.True(t, f.Help)
			},
		},
		{
			name: "version flag",
			args: []string{"--version"},
			checkFunc: func(t *testing.T, f *types.CLIFlags, err error) {
				t.Helper()
				require.NoError(t, err)
				assert.True(t, f.Version)
			},
		},
		{
			name: "version shorthand flag",
			args: []string{"-v"},
			checkFunc: func(t *testing.T, f *types.CLIFlags, err error) {
				t.Helper()
				require.NoError(t, err)
				assert.True(t, f.Version)
			},
		},
		{
			name: "json output flag",
			args: []string{"--json", "state", "light.kitchen"},
			checkFunc: func(t *testing.T, f *types.CLIFlags, err error) {
				t.Helper()
				require.NoError(t, err)
				assert.True(t, f.JSON)
				assert.Equal(t, []string{"state", "light.kitchen"}, f.Args)
			},
		},
		{
			name: "compact output flag",
			args: []string{"--compact", "states"},
			checkFunc: func(t *testing.T, f *types.CLIFlags, err error) {
				t.Helper()
				require.NoError(t, err)
				assert.True(t, f.Compact)
				assert.Equal(t, []string{"states"}, f.Args)
			},
		},
		{
			name: "output format flag",
			args: []string{"--output", "json", "ping"},
			checkFunc: func(t *testing.T, f *types.CLIFlags, err error) {
				t.Helper()
				require.NoError(t, err)
				assert.Equal(t, "json", f.Output)
			},
		},
		{
			name: "format alias flag",
			args: []string{"--format", "compact", "states"},
			checkFunc: func(t *testing.T, f *types.CLIFlags, err error) {
				t.Helper()
				require.NoError(t, err)
				assert.Equal(t, "compact", f.Output)
			},
		},
		{
			name: "no-headers flag",
			args: []string{"--no-headers", "states"},
			checkFunc: func(t *testing.T, f *types.CLIFlags, err error) {
				t.Helper()
				require.NoError(t, err)
				assert.True(t, f.NoHeaders)
			},
		},
		{
			name: "no-timestamps flag",
			args: []string{"--no-timestamps", "history", "sensor.temp"},
			checkFunc: func(t *testing.T, f *types.CLIFlags, err error) {
				t.Helper()
				require.NoError(t, err)
				assert.True(t, f.NoTimestamps)
			},
		},
		{
			name: "show-age flag",
			args: []string{"--show-age", "states"},
			checkFunc: func(t *testing.T, f *types.CLIFlags, err error) {
				t.Helper()
				require.NoError(t, err)
				assert.True(t, f.ShowAge)
			},
		},
		{
			name: "max-items flag",
			args: []string{"--max-items", "10", "states"},
			checkFunc: func(t *testing.T, f *types.CLIFlags, err error) {
				t.Helper()
				require.NoError(t, err)
				assert.Equal(t, 10, f.MaxItems)
			},
		},
		{
			name: "from time flag with valid date",
			args: []string{"--from", "2024-06-15", "history", "sensor.temp"},
			checkFunc: func(t *testing.T, f *types.CLIFlags, err error) {
				t.Helper()
				require.NoError(t, err)
				assert.Equal(t, "2024-06-15", f.From)
				require.NotNil(t, f.FromTime)
				assert.Equal(t, 2024, f.FromTime.Year())
				assert.Equal(t, time.June, f.FromTime.Month())
				assert.Equal(t, 15, f.FromTime.Day())
			},
		},
		{
			name: "to time flag with valid datetime",
			args: []string{"--to", "2024-06-15T23:59:59Z", "history", "sensor.temp"},
			checkFunc: func(t *testing.T, f *types.CLIFlags, err error) {
				t.Helper()
				require.NoError(t, err)
				assert.Equal(t, "2024-06-15T23:59:59Z", f.To)
				require.NotNil(t, f.ToTime)
				assert.Equal(t, 23, f.ToTime.Hour())
			},
		},
		{
			name: "from and to flags together",
			args: []string{"--from", "2024-06-01", "--to", "2024-06-15", "history", "sensor.temp"},
			checkFunc: func(t *testing.T, f *types.CLIFlags, err error) {
				t.Helper()
				require.NoError(t, err)
				require.NotNil(t, f.FromTime)
				require.NotNil(t, f.ToTime)
				assert.True(t, f.FromTime.Before(*f.ToTime))
			},
		},
		{
			name: "invalid from date",
			args: []string{"--from", "not-a-date", "history", "sensor.temp"},
			checkFunc: func(t *testing.T, _ *types.CLIFlags, err error) {
				t.Helper()
				require.Error(t, err)
				assert.Contains(t, err.Error(), "invalid --from value")
			},
		},
		{
			name: "invalid to date",
			args: []string{"--to", "bad-date", "history", "sensor.temp"},
			checkFunc: func(t *testing.T, _ *types.CLIFlags, err error) {
				t.Helper()
				require.Error(t, err)
				assert.Contains(t, err.Error(), "invalid --to value")
			},
		},
		{
			name: "positional arguments only",
			args: []string{"state", "light.living_room"},
			checkFunc: func(t *testing.T, f *types.CLIFlags, err error) {
				t.Helper()
				require.NoError(t, err)
				assert.Equal(t, []string{"state", "light.living_room"}, f.Args)
				assert.False(t, f.Help)
				assert.False(t, f.Version)
				assert.False(t, f.JSON)
				assert.False(t, f.Compact)
			},
		},
		{
			name: "empty args",
			args: []string{},
			checkFunc: func(t *testing.T, f *types.CLIFlags, err error) {
				t.Helper()
				require.NoError(t, err)
				assert.Empty(t, f.Args)
			},
		},
		{
			name: "combined flags and args",
			args: []string{"--json", "--max-items", "5", "--no-headers", "traces", "automation.test"},
			checkFunc: func(t *testing.T, f *types.CLIFlags, err error) {
				t.Helper()
				require.NoError(t, err)
				assert.True(t, f.JSON)
				assert.Equal(t, 5, f.MaxItems)
				assert.True(t, f.NoHeaders)
				assert.Equal(t, []string{"traces", "automation.test"}, f.Args)
			},
		},
		{
			name: "help command with extra args",
			args: []string{"help", "state"},
			checkFunc: func(t *testing.T, f *types.CLIFlags, err error) {
				t.Helper()
				require.NoError(t, err)
				assert.True(t, f.Help)
				assert.Equal(t, []string{"state"}, f.Args)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			f, err := Parse(tt.args)
			tt.checkFunc(t, f, err)
		})
	}
}

func TestFlagSet(t *testing.T) {
	t.Parallel()

	fs := FlagSet("test-app")
	require.NotNil(t, fs)
	assert.Equal(t, "test-app", fs.Name())
}

func TestAddStandardFlags(t *testing.T) {
	t.Parallel()

	fs := FlagSet("test")
	f := &types.CLIFlags{}
	AddStandardFlags(fs, f)

	// Parse --help
	err := fs.Parse([]string{"--help"})
	require.NoError(t, err)
	assert.True(t, f.Help)
}

func TestAddTimeFlags(t *testing.T) {
	t.Parallel()

	fs := FlagSet("test")
	f := &types.CLIFlags{}
	AddTimeFlags(fs, f)

	err := fs.Parse([]string{"--from", "2024-01-01", "--to", "2024-12-31"})
	require.NoError(t, err)
	assert.Equal(t, "2024-01-01", f.From)
	assert.Equal(t, "2024-12-31", f.To)
}

func TestAddOutputFlags(t *testing.T) {
	t.Parallel()

	fs := FlagSet("test")
	f := &types.CLIFlags{}
	AddOutputFlags(fs, f)

	err := fs.Parse([]string{"--json", "--compact", "--no-headers", "--max-items", "20"})
	require.NoError(t, err)
	assert.True(t, f.JSON)
	assert.True(t, f.Compact)
	assert.True(t, f.NoHeaders)
	assert.Equal(t, 20, f.MaxItems)
}

func TestAddAllFlags(t *testing.T) {
	t.Parallel()

	fs := FlagSet("test")
	f := &types.CLIFlags{}
	AddAllFlags(fs, f)

	err := fs.Parse([]string{
		"--help",
		"--json",
		"--from", "2024-01-01",
		"--max-items", "50",
	})
	require.NoError(t, err)
	assert.True(t, f.Help)
	assert.True(t, f.JSON)
	assert.Equal(t, "2024-01-01", f.From)
	assert.Equal(t, 50, f.MaxItems)
}

// Benchmark for ParseFlexibleDate
func BenchmarkParseFlexibleDate(b *testing.B) {
	inputs := []string{
		"2024-06-15T10:30:45Z",
		"2024-06-15T10:30:45",
		"2024-06-15 10:30:45",
		"2024-06-15 10:30",
		"2024-06-15",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		input := inputs[i%len(inputs)]
		_, _ = ParseFlexibleDate(input) //nolint:errcheck // benchmark intentionally ignores error
	}
}
