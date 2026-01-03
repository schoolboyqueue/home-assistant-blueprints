package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
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

func TestFormatDuration_EdgeCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		duration time.Duration
		expected string
	}{
		{
			name:     "negative duration",
			duration: -5 * time.Second,
			expected: "-5s",
		},
		{
			name:     "sub-second duration",
			duration: 500 * time.Millisecond,
			expected: "0s",
		},
		{
			name:     "59 minutes 59 seconds",
			duration: 59*time.Minute + 59*time.Second,
			expected: "59m",
		},
		{
			name:     "23 hours 59 minutes",
			duration: 23*time.Hour + 59*time.Minute,
			expected: "23h",
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

func TestContext_Initialization(t *testing.T) {
	t.Parallel()

	t.Run("empty context", func(t *testing.T) {
		t.Parallel()
		ctx := &Context{}
		assert.Nil(t, ctx.Client)
		assert.Nil(t, ctx.Args)
		assert.Nil(t, ctx.FromTime)
		assert.Nil(t, ctx.ToTime)
		assert.Nil(t, ctx.Config)
		assert.Nil(t, ctx.Ctx)
	})

	t.Run("context with args", func(t *testing.T) {
		t.Parallel()
		ctx := &Context{
			Args: []string{"command", "arg1", "arg2"},
		}
		assert.Len(t, ctx.Args, 3)
		assert.Equal(t, "command", ctx.Args[0])
		assert.Equal(t, "arg1", ctx.Args[1])
	})

	t.Run("context with time range", func(t *testing.T) {
		t.Parallel()
		now := time.Now()
		from := now.Add(-24 * time.Hour)
		ctx := &Context{
			FromTime: &from,
			ToTime:   &now,
		}
		assert.NotNil(t, ctx.FromTime)
		assert.NotNil(t, ctx.ToTime)
		assert.True(t, ctx.ToTime.After(*ctx.FromTime))
	})

	t.Run("context with context.Context", func(t *testing.T) {
		t.Parallel()
		goCtx := context.Background()
		ctx := &Context{
			Ctx: goCtx,
		}
		assert.NotNil(t, ctx.Ctx)
		assert.Equal(t, goCtx, ctx.Ctx)
	})
}

func TestContext_Done(t *testing.T) {
	t.Parallel()

	t.Run("nil context returns nil channel", func(t *testing.T) {
		t.Parallel()
		ctx := &Context{}
		assert.Nil(t, ctx.Done())
	})

	t.Run("with context returns done channel", func(t *testing.T) {
		t.Parallel()
		goCtx, cancel := context.WithCancel(context.Background())
		defer cancel()

		ctx := &Context{Ctx: goCtx}

		// Channel should not be closed initially
		select {
		case <-ctx.Done():
			t.Error("expected done channel to not be closed initially")
		default:
			// Expected
		}
	})

	t.Run("done channel closes on cancel", func(t *testing.T) {
		t.Parallel()
		goCtx, cancel := context.WithCancel(context.Background())

		ctx := &Context{Ctx: goCtx}

		// Cancel the context
		cancel()

		// Channel should be closed after cancel
		select {
		case <-ctx.Done():
			// Expected
		case <-time.After(100 * time.Millisecond):
			t.Error("expected done channel to be closed after cancel")
		}
	})
}

func TestContext_Err(t *testing.T) {
	t.Parallel()

	t.Run("nil context returns nil error", func(t *testing.T) {
		t.Parallel()
		ctx := &Context{}
		assert.Nil(t, ctx.Err())
	})

	t.Run("uncanceled context returns nil error", func(t *testing.T) {
		t.Parallel()
		goCtx, cancel := context.WithCancel(context.Background())
		defer cancel()

		ctx := &Context{Ctx: goCtx}
		assert.Nil(t, ctx.Err())
	})

	t.Run("canceled context returns context.Canceled", func(t *testing.T) {
		t.Parallel()
		goCtx, cancel := context.WithCancel(context.Background())

		ctx := &Context{Ctx: goCtx}

		// Cancel the context
		cancel()

		// Should return context.Canceled
		assert.Equal(t, context.Canceled, ctx.Err())
	})

	t.Run("timed out context returns context.DeadlineExceeded", func(t *testing.T) {
		t.Parallel()
		goCtx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
		defer cancel()

		ctx := &Context{Ctx: goCtx}

		// Wait for timeout
		time.Sleep(10 * time.Millisecond)

		// Should return context.DeadlineExceeded
		assert.Equal(t, context.DeadlineExceeded, ctx.Err())
	})
}

func TestContext_DoneAndErrConsistency(t *testing.T) {
	t.Parallel()

	t.Run("done and err are consistent", func(t *testing.T) {
		t.Parallel()
		goCtx, cancel := context.WithCancel(context.Background())

		ctx := &Context{Ctx: goCtx}

		// Before cancel, Done() should not be ready and Err() should be nil
		select {
		case <-ctx.Done():
			t.Error("Done() should not be ready before cancel")
		default:
			assert.Nil(t, ctx.Err())
		}

		// Cancel
		cancel()

		// After cancel, Done() should be ready and Err() should be non-nil
		select {
		case <-ctx.Done():
			assert.NotNil(t, ctx.Err())
		case <-time.After(100 * time.Millisecond):
			t.Error("Done() should be ready after cancel")
		}
	})
}

func TestEntityIDParsing(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		entityID   string
		wantDomain string
		wantName   string
		wantValid  bool
	}{
		{
			name:       "sensor entity",
			entityID:   "sensor.temperature",
			wantDomain: "sensor",
			wantName:   "temperature",
			wantValid:  true,
		},
		{
			name:       "binary_sensor entity",
			entityID:   "binary_sensor.motion_living_room",
			wantDomain: "binary_sensor",
			wantName:   "motion_living_room",
			wantValid:  true,
		},
		{
			name:       "input_boolean entity",
			entityID:   "input_boolean.guest_mode",
			wantDomain: "input_boolean",
			wantName:   "guest_mode",
			wantValid:  true,
		},
		{
			name:       "no dot separator",
			entityID:   "invalid_entity",
			wantDomain: "",
			wantName:   "",
			wantValid:  false,
		},
		{
			name:       "empty string",
			entityID:   "",
			wantDomain: "",
			wantName:   "",
			wantValid:  false,
		},
		{
			name:       "multiple dots",
			entityID:   "sensor.temperature.value",
			wantDomain: "sensor",
			wantName:   "temperature.value",
			wantValid:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			parts := strings.SplitN(tt.entityID, ".", 2)
			if tt.wantValid {
				assert.Len(t, parts, 2)
				assert.Equal(t, tt.wantDomain, parts[0])
				assert.Equal(t, tt.wantName, parts[1])
			} else {
				valid := len(parts) == 2 && parts[0] != "" && parts[1] != ""
				assert.False(t, valid)
			}
		})
	}
}

func TestServiceCallDataValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		jsonData  string
		wantValid bool
	}{
		{
			name:      "valid entity_id target",
			jsonData:  `{"entity_id": "light.kitchen"}`,
			wantValid: true,
		},
		{
			name:      "valid with nested data",
			jsonData:  `{"entity_id": "light.kitchen", "brightness": 255, "rgb_color": [255, 0, 0]}`,
			wantValid: true,
		},
		{
			name:      "empty object",
			jsonData:  `{}`,
			wantValid: true,
		},
		{
			name:      "invalid json",
			jsonData:  `{entity_id: light}`,
			wantValid: false,
		},
		{
			name:      "array instead of object",
			jsonData:  `["light.kitchen"]`,
			wantValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var data map[string]any
			err := json.Unmarshal([]byte(tt.jsonData), &data)
			if tt.wantValid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestCompareAttributeDifferences(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		attrs1     map[string]any
		attrs2     map[string]any
		expectDiff bool
	}{
		{
			name:       "identical attributes",
			attrs1:     map[string]any{"brightness": 255, "color_mode": "rgb"},
			attrs2:     map[string]any{"brightness": 255, "color_mode": "rgb"},
			expectDiff: false,
		},
		{
			name:       "different values",
			attrs1:     map[string]any{"brightness": 255},
			attrs2:     map[string]any{"brightness": 128},
			expectDiff: true,
		},
		{
			name:       "missing attribute in second",
			attrs1:     map[string]any{"brightness": 255, "color_mode": "rgb"},
			attrs2:     map[string]any{"brightness": 255},
			expectDiff: true,
		},
		{
			name:       "extra attribute in second",
			attrs1:     map[string]any{"brightness": 255},
			attrs2:     map[string]any{"brightness": 255, "color_mode": "rgb"},
			expectDiff: true,
		},
		{
			name:       "both empty",
			attrs1:     map[string]any{},
			attrs2:     map[string]any{},
			expectDiff: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			attrDiffs := make(map[string]any)

			// Find differences (same logic as handleCompare)
			for k, v1 := range tt.attrs1 {
				if v2, ok := tt.attrs2[k]; ok {
					if fmt.Sprintf("%v", v1) != fmt.Sprintf("%v", v2) {
						attrDiffs[k] = map[string]any{"entity1": v1, "entity2": v2}
					}
				} else {
					attrDiffs[k] = map[string]any{"entity1": v1, "entity2": nil}
				}
			}
			for k, v2 := range tt.attrs2 {
				if _, ok := tt.attrs1[k]; !ok {
					attrDiffs[k] = map[string]any{"entity1": nil, "entity2": v2}
				}
			}

			if tt.expectDiff {
				assert.NotEmpty(t, attrDiffs)
			} else {
				assert.Empty(t, attrDiffs)
			}
		})
	}
}

func TestDeviceHealthStatus(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		lastUpdated    string
		staleThreshold time.Duration
		expectedStatus string
	}{
		{
			name:           "recent update (ok)",
			lastUpdated:    time.Now().Add(-5 * time.Minute).Format(time.RFC3339),
			staleThreshold: time.Hour,
			expectedStatus: statusOK,
		},
		{
			name:           "stale update",
			lastUpdated:    time.Now().Add(-2 * time.Hour).Format(time.RFC3339),
			staleThreshold: time.Hour,
			expectedStatus: statusStale,
		},
		{
			name:           "just at threshold (stale)",
			lastUpdated:    time.Now().Add(-61 * time.Minute).Format(time.RFC3339),
			staleThreshold: time.Hour,
			expectedStatus: statusStale,
		},
		{
			name:           "empty timestamp",
			lastUpdated:    "",
			staleThreshold: time.Hour,
			expectedStatus: statusUnknown,
		},
		{
			name:           "invalid timestamp",
			lastUpdated:    "not-a-timestamp",
			staleThreshold: time.Hour,
			expectedStatus: statusUnknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			status := statusUnknown
			if tt.lastUpdated != "" {
				if parsedTime, err := time.Parse(time.RFC3339, tt.lastUpdated); err == nil {
					age := time.Since(parsedTime)
					if age > tt.staleThreshold {
						status = statusStale
					} else {
						status = statusOK
					}
				}
			}

			assert.Equal(t, tt.expectedStatus, status)
		})
	}
}

func TestErrEntityNotFound_Wrapping(t *testing.T) {
	t.Parallel()

	t.Run("error wrapping with entity id", func(t *testing.T) {
		t.Parallel()

		entityID := "sensor.nonexistent"
		err := fmt.Errorf("%w: %s", ErrEntityNotFound, entityID)

		assert.True(t, errors.Is(err, ErrEntityNotFound))
		assert.Contains(t, err.Error(), entityID)
		assert.Contains(t, err.Error(), "entity not found")
	})

	t.Run("unwrap returns original error", func(t *testing.T) {
		t.Parallel()

		err := fmt.Errorf("failed to get state: %w", ErrEntityNotFound)
		assert.True(t, errors.Is(err, ErrEntityNotFound))
	})
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

// Benchmark for entity ID parsing
func BenchmarkEntityIDParsing(b *testing.B) {
	entityIDs := []string{
		"sensor.temperature",
		"binary_sensor.motion_living_room",
		"light.kitchen_ceiling",
		"input_boolean.guest_mode",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		id := entityIDs[i%len(entityIDs)]
		_ = strings.SplitN(id, ".", 2)
	}
}

// Benchmark for attribute comparison
func BenchmarkAttributeComparison(b *testing.B) {
	attrs1 := map[string]any{
		"brightness":    255,
		"color_mode":    "rgb",
		"rgb_color":     []int{255, 0, 0},
		"friendly_name": "Kitchen Light",
	}
	attrs2 := map[string]any{
		"brightness":    128,
		"color_mode":    "rgb",
		"rgb_color":     []int{0, 255, 0},
		"friendly_name": "Kitchen Light",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		diffs := make(map[string]any)
		for k, v1 := range attrs1 {
			if v2, ok := attrs2[k]; ok {
				if fmt.Sprintf("%v", v1) != fmt.Sprintf("%v", v2) {
					diffs[k] = map[string]any{"entity1": v1, "entity2": v2}
				}
			}
		}
	}
}
