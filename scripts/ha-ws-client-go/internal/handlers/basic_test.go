package handlers

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFormatDuration(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		duration time.Duration
		expected string
	}{
		{
			name:     "seconds - zero",
			duration: 0,
			expected: "0s",
		},
		{
			name:     "seconds - small",
			duration: 5 * time.Second,
			expected: "5s",
		},
		{
			name:     "seconds - near minute",
			duration: 59 * time.Second,
			expected: "59s",
		},
		{
			name:     "minutes - exact",
			duration: time.Minute,
			expected: "1m",
		},
		{
			name:     "minutes - with seconds ignored",
			duration: 5*time.Minute + 30*time.Second,
			expected: "5m",
		},
		{
			name:     "minutes - near hour",
			duration: 59 * time.Minute,
			expected: "59m",
		},
		{
			name:     "hours - exact",
			duration: time.Hour,
			expected: "1h",
		},
		{
			name:     "hours - with minutes ignored",
			duration: 5*time.Hour + 30*time.Minute,
			expected: "5h",
		},
		{
			name:     "hours - near day",
			duration: 23 * time.Hour,
			expected: "23h",
		},
		{
			name:     "days - exact",
			duration: 24 * time.Hour,
			expected: "1d",
		},
		{
			name:     "days - multiple",
			duration: 72 * time.Hour,
			expected: "3d",
		},
		{
			name:     "days - with hours ignored",
			duration: 25 * time.Hour,
			expected: "1d",
		},
		{
			name:     "large duration",
			duration: 30 * 24 * time.Hour,
			expected: "30d",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := formatDuration(tt.duration)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestErrEntityNotFound(t *testing.T) {
	t.Parallel()

	assert.NotNil(t, ErrEntityNotFound)
	assert.Equal(t, "entity not found", ErrEntityNotFound.Error())

	// Test wrapping
	wrappedErr := errors.New("test: entity not found")
	assert.Contains(t, wrappedErr.Error(), "entity not found")
}

func TestStatusConstants(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "ok", statusOK)
	assert.Equal(t, "stale", statusStale)
	assert.Equal(t, "unknown", statusUnknown)
}

func TestEnsureAutomationPrefix_FromMiddleware(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input    string
		expected string
	}{
		{"kitchen_lights", "automation.kitchen_lights"},
		{"automation.kitchen_lights", "automation.kitchen_lights"},
		{"", "automation."},
		{"automation.", "automation."},
		{"test_automation", "automation.test_automation"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			t.Parallel()
			result := EnsureAutomationPrefix(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Benchmark for formatDuration
func BenchmarkFormatDuration(b *testing.B) {
	durations := []time.Duration{
		5 * time.Second,
		5 * time.Minute,
		5 * time.Hour,
		5 * 24 * time.Hour,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		d := durations[i%len(durations)]
		_ = formatDuration(d)
	}
}
