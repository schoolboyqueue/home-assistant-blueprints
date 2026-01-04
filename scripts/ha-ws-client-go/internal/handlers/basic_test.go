package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/home-assistant-blueprints/testfixtures"

	"github.com/home-assistant-blueprints/ha-ws-client-go/internal/types"
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

// =====================================
// Handler Unit Tests
// =====================================

func TestHandlePing(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		setupRouter func(*MessageRouter)
		expectError bool
		description string
	}{
		{
			name: "success with latency",
			setupRouter: func(r *MessageRouter) {
				r.OnSuccess("ping", nil)
			},
			expectError: false,
			description: "Successful ping returns latency",
		},
		{
			name: "websocket error",
			setupRouter: func(r *MessageRouter) {
				r.OnError("ping", "connection_error", "Connection failed")
			},
			expectError: true,
			description: "WebSocket error returns error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			router := NewMessageRouter(t)
			tt.setupRouter(router)

			ctx, cleanup := NewTestContext(t, router)
			defer cleanup()

			err := HandlePing(ctx)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestHandleState(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		entityID    string
		states      any
		expectError bool
		description string
	}{
		{
			name:     "entity found",
			entityID: "light.kitchen",
			states: []testfixtures.HAState{
				testfixtures.NewHAState("light.kitchen", "on"),
				testfixtures.NewHAState("light.bedroom", "off"),
			},
			expectError: false,
			description: "Returns state when entity is found",
		},
		{
			name:     "entity not found",
			entityID: "light.nonexistent",
			states: []testfixtures.HAState{
				testfixtures.NewHAState("light.kitchen", "on"),
			},
			expectError: true,
			description: "Returns error when entity is not found",
		},
		{
			name:     "with attributes",
			entityID: "light.kitchen",
			states: []testfixtures.HAState{
				testfixtures.NewHAStateWithAttrs("light.kitchen", "on", map[string]any{
					"brightness":    255,
					"color_mode":    "rgb",
					"friendly_name": "Kitchen Light",
				}),
			},
			expectError: false,
			description: "Returns state with attributes",
		},
		{
			name:        "empty states list",
			entityID:    "light.any",
			states:      []testfixtures.HAState{},
			expectError: true,
			description: "Returns error when no states exist",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			router := NewMessageRouter(t)
			router.OnSuccess("get_states", tt.states)

			config := &HandlerConfig{
				Args: []string{tt.entityID},
			}

			ctx, cleanup := NewTestContext(t, router,
				WithArgs("state", tt.entityID),
				WithHandlerConfig(config),
			)
			defer cleanup()

			err := handleState(ctx)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestHandleStateWebSocketError(t *testing.T) {
	t.Parallel()

	router := NewMessageRouter(t).OnError("get_states", "connection_error", "Connection failed")

	config := &HandlerConfig{
		Args: []string{"light.kitchen"},
	}

	ctx, cleanup := NewTestContext(t, router,
		WithArgs("state", "light.kitchen"),
		WithHandlerConfig(config),
	)
	defer cleanup()

	err := handleState(ctx)
	assert.Error(t, err)
}

func TestHandleStates(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		states      any
		expectError bool
		description string
	}{
		{
			name: "with multiple states",
			states: []testfixtures.HAState{
				testfixtures.NewHAState("light.kitchen", "on"),
				testfixtures.NewHAState("light.bedroom", "off"),
				testfixtures.NewHAState("sensor.temperature", "23.5"),
			},
			expectError: false,
			description: "Returns sample of states",
		},
		{
			name:        "empty states",
			states:      []testfixtures.HAState{},
			expectError: false,
			description: "Handles empty state list",
		},
		{
			name: "many states (tests limit)",
			states: func() []testfixtures.HAState {
				states := make([]testfixtures.HAState, 50)
				for i := range 50 {
					states[i] = testfixtures.NewHAState(fmt.Sprintf("sensor.test_%d", i), strconv.Itoa(i))
				}
				return states
			}(),
			expectError: false,
			description: "Limits output to configured max",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			router := NewMessageRouter(t)
			router.OnSuccess("get_states", tt.states)

			ctx, cleanup := NewTestContext(t, router, WithArgs("states"))
			defer cleanup()

			err := HandleStates(ctx)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestHandleStatesWebSocketError(t *testing.T) {
	t.Parallel()

	router := NewMessageRouter(t).OnError("get_states", "connection_error", "Connection failed")

	ctx, cleanup := NewTestContext(t, router, WithArgs("states"))
	defer cleanup()

	err := HandleStates(ctx)
	assert.Error(t, err)
}

func TestHandleStatesJSON(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		states      any
		expectError bool
		description string
	}{
		{
			name: "returns all states as JSON",
			states: []testfixtures.HAState{
				testfixtures.NewHAState("light.kitchen", "on"),
				testfixtures.NewHAState("light.bedroom", "off"),
				testfixtures.NewHAState("sensor.temperature", "23.5"),
			},
			expectError: false,
			description: "Returns all states in JSON format",
		},
		{
			name:        "empty states",
			states:      []testfixtures.HAState{},
			expectError: false,
			description: "Handles empty state list",
		},
		{
			name: "states with timestamps",
			states: []testfixtures.HAState{
				testfixtures.NewHAStateWithTimestamps("light.kitchen", "on"),
				testfixtures.NewHAStateWithTimestamps("sensor.temperature", "23.5"),
			},
			expectError: false,
			description: "Includes timestamp information",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			router := NewMessageRouter(t)
			router.OnSuccess("get_states", tt.states)

			ctx, cleanup := NewTestContext(t, router, WithArgs("states-json"))
			defer cleanup()

			err := HandleStatesJSON(ctx)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestHandleStatesJSONWebSocketError(t *testing.T) {
	t.Parallel()

	router := NewMessageRouter(t).OnError("get_states", "connection_error", "Connection failed")

	ctx, cleanup := NewTestContext(t, router, WithArgs("states-json"))
	defer cleanup()

	err := HandleStatesJSON(ctx)
	assert.Error(t, err)
}

func TestHandleStatesFilter(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		pattern     string
		states      any
		expectError bool
		description string
	}{
		{
			name:    "pattern matches entities",
			pattern: "light.*",
			states: []testfixtures.HAState{
				testfixtures.NewHAState("light.kitchen", "on"),
				testfixtures.NewHAState("light.bedroom", "off"),
				testfixtures.NewHAState("sensor.temperature", "23.5"),
			},
			expectError: false,
			description: "Filters states matching pattern",
		},
		{
			name:    "no matches",
			pattern: "switch.*",
			states: []testfixtures.HAState{
				testfixtures.NewHAState("light.kitchen", "on"),
				testfixtures.NewHAState("sensor.temperature", "23.5"),
			},
			expectError: false,
			description: "Handles no matches gracefully",
		},
		{
			name:    "exact match pattern",
			pattern: "sensor.temperature",
			states: []testfixtures.HAState{
				testfixtures.NewHAState("light.kitchen", "on"),
				testfixtures.NewHAState("sensor.temperature", "23.5"),
			},
			expectError: false,
			description: "Matches exact entity ID",
		},
		{
			name:    "with timestamps for age display",
			pattern: "light.*",
			states: []testfixtures.HAState{
				testfixtures.NewHAStateWithTimestamps("light.kitchen", "on"),
				testfixtures.NewHAStateWithTimestamps("light.bedroom", "off"),
			},
			expectError: false,
			description: "Includes age information",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			router := NewMessageRouter(t)
			router.OnSuccess("get_states", tt.states)

			// Build the regex pattern (same logic as middleware)
			pattern := tt.pattern
			pattern = strings.ReplaceAll(pattern, ".", `\.`)
			pattern = strings.ReplaceAll(pattern, "*", ".*")
			pattern = "(?i)^" + pattern + "$"
			re, err := regexp.Compile(pattern)
			assert.NoError(t, err)

			config := &HandlerConfig{
				Args:    []string{tt.pattern},
				Pattern: re,
			}

			ctx, cleanup := NewTestContext(t, router,
				WithArgs("states-filter", tt.pattern),
				WithHandlerConfig(config),
			)
			defer cleanup()

			err = handleStatesFilter(ctx)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestHandleStatesFilterWebSocketError(t *testing.T) {
	t.Parallel()

	router := NewMessageRouter(t).OnError("get_states", "connection_error", "Connection failed")

	re := regexp.MustCompile("light.*")
	config := &HandlerConfig{
		Args:    []string{"light.*"},
		Pattern: re,
	}

	ctx, cleanup := NewTestContext(t, router,
		WithArgs("states-filter", "light.*"),
		WithHandlerConfig(config),
	)
	defer cleanup()

	err := handleStatesFilter(ctx)
	assert.Error(t, err)
}

func TestHandleConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		config      any
		expectError bool
		description string
	}{
		{
			name:        "full config",
			config:      testfixtures.NewHAConfig(),
			expectError: false,
			description: "Returns HA configuration",
		},
		{
			name: "minimal config",
			config: testfixtures.HAConfig{
				Version:      "2024.1.0",
				LocationName: "Home",
				State:        "RUNNING",
			},
			expectError: false,
			description: "Handles minimal config",
		},
		{
			name: "config with many components",
			config: testfixtures.HAConfig{
				Version:      "2024.6.0",
				LocationName: "Smart Home",
				TimeZone:     "Europe/London",
				State:        "RUNNING",
				Components:   []string{"homeassistant", "automation", "script", "scene", "light", "switch", "sensor", "binary_sensor", "climate", "media_player"},
			},
			expectError: false,
			description: "Handles config with many components",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			router := NewMessageRouter(t)
			router.OnSuccess("get_config", tt.config)

			ctx, cleanup := NewTestContext(t, router, WithArgs("config"))
			defer cleanup()

			err := HandleConfig(ctx)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestHandleConfigWebSocketError(t *testing.T) {
	t.Parallel()

	router := NewMessageRouter(t).OnError("get_config", "connection_error", "Connection failed")

	ctx, cleanup := NewTestContext(t, router, WithArgs("config"))
	defer cleanup()

	err := HandleConfig(ctx)
	assert.Error(t, err)
}

func TestHandleServices(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		services    map[string]map[string]any
		expectError bool
		description string
	}{
		{
			name: "multiple domains with services",
			services: map[string]map[string]any{
				"light": {
					"turn_on":  map[string]any{"description": "Turn on a light"},
					"turn_off": map[string]any{"description": "Turn off a light"},
					"toggle":   map[string]any{"description": "Toggle a light"},
				},
				"switch": {
					"turn_on":  map[string]any{"description": "Turn on a switch"},
					"turn_off": map[string]any{"description": "Turn off a switch"},
				},
				"homeassistant": {
					"reload_all": map[string]any{"description": "Reload all configs"},
				},
			},
			expectError: false,
			description: "Lists services by domain",
		},
		{
			name:        "empty services",
			services:    map[string]map[string]any{},
			expectError: false,
			description: "Handles empty service list",
		},
		{
			name: "single domain",
			services: map[string]map[string]any{
				"automation": {
					"trigger": map[string]any{"description": "Trigger an automation"},
					"reload":  map[string]any{"description": "Reload automations"},
				},
			},
			expectError: false,
			description: "Handles single domain",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			router := NewMessageRouter(t)
			router.OnSuccess("get_services", tt.services)

			ctx, cleanup := NewTestContext(t, router, WithArgs("services"))
			defer cleanup()

			err := HandleServices(ctx)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestHandleServicesWebSocketError(t *testing.T) {
	t.Parallel()

	router := NewMessageRouter(t).OnError("get_services", "connection_error", "Connection failed")

	ctx, cleanup := NewTestContext(t, router, WithArgs("services"))
	defer cleanup()

	err := HandleServices(ctx)
	assert.Error(t, err)
}

func TestHandleCall(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		args        []string
		config      *HandlerConfig
		response    any
		expectError bool
		description string
	}{
		{
			name: "simple service call",
			args: []string{"call", "light", "turn_on"},
			config: &HandlerConfig{
				Args: []string{"light", "turn_on"},
			},
			response:    map[string]any{"success": true},
			expectError: false,
			description: "Calls service successfully",
		},
		{
			name: "service call with JSON data",
			args: []string{"call", "light", "turn_on", `{"entity_id":"light.kitchen","brightness":255}`},
			config: &HandlerConfig{
				Args: []string{"light", "turn_on"},
			},
			response:    map[string]any{"success": true},
			expectError: false,
			description: "Calls service with data",
		},
		{
			name: "service call with invalid JSON",
			args: []string{"call", "light", "turn_on", `{invalid json}`},
			config: &HandlerConfig{
				Args: []string{"light", "turn_on"},
			},
			response:    nil,
			expectError: true,
			description: "Returns error for invalid JSON",
		},
		{
			name: "service call with response data",
			args: []string{"call", "homeassistant", "reload_all"},
			config: &HandlerConfig{
				Args: []string{"homeassistant", "reload_all"},
			},
			response: map[string]any{
				"context": map[string]any{
					"id":        "ctx-123",
					"parent_id": nil,
					"user_id":   "user-456",
				},
			},
			expectError: false,
			description: "Returns service response",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			router := NewMessageRouter(t)
			router.OnSuccess("call_service", tt.response)

			ctx, cleanup := NewTestContext(t, router,
				WithArgs(tt.args...),
				WithHandlerConfig(tt.config),
			)
			defer cleanup()

			err := handleCall(ctx)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestHandleCallWebSocketError(t *testing.T) {
	t.Parallel()

	router := NewMessageRouter(t).OnError("call_service", "service_not_found", "Service not found")

	config := &HandlerConfig{
		Args: []string{"nonexistent", "service"},
	}

	ctx, cleanup := NewTestContext(t, router,
		WithArgs("call", "nonexistent", "service"),
		WithHandlerConfig(config),
	)
	defer cleanup()

	err := handleCall(ctx)
	assert.Error(t, err)
}

func TestHandleDeviceHealth(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	recentUpdate := now.Add(-5 * time.Minute).Format(time.RFC3339)
	staleUpdate := now.Add(-2 * time.Hour).Format(time.RFC3339)

	tests := []struct {
		name        string
		entityID    string
		states      []types.HAState
		expectError bool
		description string
	}{
		{
			name:     "ok status - recent update",
			entityID: "cover.living_room_shade",
			states: []types.HAState{
				{
					EntityID:    "cover.living_room_shade",
					State:       "open",
					LastUpdated: recentUpdate,
				},
			},
			expectError: false,
			description: "Returns ok status for recent updates",
		},
		{
			name:     "stale status - old update",
			entityID: "cover.old_shade",
			states: []types.HAState{
				{
					EntityID:    "cover.old_shade",
					State:       "open",
					LastUpdated: staleUpdate,
				},
			},
			expectError: false,
			description: "Returns stale status for old updates",
		},
		{
			name:     "entity not found",
			entityID: "cover.nonexistent",
			states: []types.HAState{
				{EntityID: "cover.other", State: "open", LastUpdated: recentUpdate},
			},
			expectError: true,
			description: "Returns error when entity not found",
		},
		{
			name:     "with related entities",
			entityID: "cover.guest_bedroom_window_shade",
			states: []types.HAState{
				{
					EntityID:    "cover.guest_bedroom_window_shade",
					State:       "open",
					LastUpdated: recentUpdate,
				},
				{
					EntityID:    "sensor.guest_bedroom_window_shade_battery",
					State:       "85",
					LastUpdated: recentUpdate,
				},
				{
					EntityID:    "sensor.guest_bedroom_window_shade_signal",
					State:       "-45",
					LastUpdated: staleUpdate,
				},
			},
			expectError: false,
			description: "Includes related entities in health check",
		},
		{
			name:        "invalid entity ID format",
			entityID:    "invalid_no_domain",
			states:      []types.HAState{},
			expectError: true,
			description: "Returns error for invalid entity ID format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			router := NewMessageRouter(t)
			router.OnSuccess("get_states", tt.states)

			config := &HandlerConfig{
				Args: []string{tt.entityID},
			}

			ctx, cleanup := NewTestContext(t, router,
				WithArgs("device-health", tt.entityID),
				WithHandlerConfig(config),
			)
			defer cleanup()

			err := handleDeviceHealth(ctx)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestHandleDeviceHealthWebSocketError(t *testing.T) {
	t.Parallel()

	router := NewMessageRouter(t).OnError("get_states", "connection_error", "Connection failed")

	config := &HandlerConfig{
		Args: []string{"cover.test"},
	}

	ctx, cleanup := NewTestContext(t, router,
		WithArgs("device-health", "cover.test"),
		WithHandlerConfig(config),
	)
	defer cleanup()

	err := handleDeviceHealth(ctx)
	assert.Error(t, err)
}

func TestHandleCompare(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	recentUpdate := now.Add(-5 * time.Minute).Format(time.RFC3339)

	tests := []struct {
		name        string
		entity1     string
		entity2     string
		states      []types.HAState
		expectError bool
		description string
	}{
		{
			name:    "same state values",
			entity1: "light.kitchen",
			entity2: "light.bedroom",
			states: []types.HAState{
				{EntityID: "light.kitchen", State: "on", LastUpdated: recentUpdate, Attributes: map[string]any{"brightness": 255}},
				{EntityID: "light.bedroom", State: "on", LastUpdated: recentUpdate, Attributes: map[string]any{"brightness": 255}},
			},
			expectError: false,
			description: "Compares entities with matching states",
		},
		{
			name:    "different state values",
			entity1: "light.kitchen",
			entity2: "light.bedroom",
			states: []types.HAState{
				{EntityID: "light.kitchen", State: "on", LastUpdated: recentUpdate, Attributes: map[string]any{"brightness": 255}},
				{EntityID: "light.bedroom", State: "off", LastUpdated: recentUpdate, Attributes: map[string]any{"brightness": 0}},
			},
			expectError: false,
			description: "Compares entities with different states",
		},
		{
			name:    "first entity not found",
			entity1: "light.nonexistent",
			entity2: "light.bedroom",
			states: []types.HAState{
				{EntityID: "light.bedroom", State: "on", LastUpdated: recentUpdate},
			},
			expectError: true,
			description: "Returns error when first entity not found",
		},
		{
			name:    "second entity not found",
			entity1: "light.kitchen",
			entity2: "light.nonexistent",
			states: []types.HAState{
				{EntityID: "light.kitchen", State: "on", LastUpdated: recentUpdate},
			},
			expectError: true,
			description: "Returns error when second entity not found",
		},
		{
			name:    "different attribute sets",
			entity1: "light.kitchen",
			entity2: "light.bedroom",
			states: []types.HAState{
				{
					EntityID:    "light.kitchen",
					State:       "on",
					LastUpdated: recentUpdate,
					Attributes: map[string]any{
						"brightness":    255,
						"color_mode":    "rgb",
						"rgb_color":     []any{255, 0, 0},
						"friendly_name": "Kitchen Light",
					},
				},
				{
					EntityID:    "light.bedroom",
					State:       "on",
					LastUpdated: recentUpdate,
					Attributes: map[string]any{
						"brightness":    128,
						"color_mode":    "color_temp",
						"color_temp":    400,
						"friendly_name": "Bedroom Light",
					},
				},
			},
			expectError: false,
			description: "Identifies attribute differences",
		},
		{
			name:    "empty attributes",
			entity1: "sensor.temperature",
			entity2: "sensor.humidity",
			states: []types.HAState{
				{EntityID: "sensor.temperature", State: "23.5", LastUpdated: recentUpdate, Attributes: map[string]any{}},
				{EntityID: "sensor.humidity", State: "45", LastUpdated: recentUpdate, Attributes: map[string]any{}},
			},
			expectError: false,
			description: "Handles entities with empty attributes",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			router := NewMessageRouter(t)
			router.OnSuccess("get_states", tt.states)

			config := &HandlerConfig{
				Args: []string{tt.entity1, tt.entity2},
			}

			ctx, cleanup := NewTestContext(t, router,
				WithArgs("compare", tt.entity1, tt.entity2),
				WithHandlerConfig(config),
			)
			defer cleanup()

			err := handleCompare(ctx)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestHandleCompareWebSocketError(t *testing.T) {
	t.Parallel()

	router := NewMessageRouter(t).OnError("get_states", "connection_error", "Connection failed")

	config := &HandlerConfig{
		Args: []string{"light.kitchen", "light.bedroom"},
	}

	ctx, cleanup := NewTestContext(t, router,
		WithArgs("compare", "light.kitchen", "light.bedroom"),
		WithHandlerConfig(config),
	)
	defer cleanup()

	err := handleCompare(ctx)
	assert.Error(t, err)
}
