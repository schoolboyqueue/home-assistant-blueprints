package cli

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
		_, _ = ParseFlexibleDate(input)
	}
}
