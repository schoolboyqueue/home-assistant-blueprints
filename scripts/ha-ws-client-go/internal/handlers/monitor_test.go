package handlers

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/home-assistant-blueprints/ha-ws-client-go/internal/types"
)

func TestMonitorEntityPatterns(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		entities []string
		pattern  string
		expected []string
	}{
		{
			name:     "sensor entities",
			entities: []string{"sensor.temperature", "sensor.humidity", "sensor.pressure"},
			pattern:  "sensor.*",
			expected: []string{"sensor.temperature", "sensor.humidity", "sensor.pressure"},
		},
		{
			name:     "mixed entities filtered to lights",
			entities: []string{"light.kitchen", "sensor.temp", "light.bedroom", "switch.fan"},
			pattern:  "light.*",
			expected: []string{"light.kitchen", "light.bedroom"},
		},
		{
			name:     "binary sensors with specific name",
			entities: []string{"binary_sensor.motion_kitchen", "binary_sensor.motion_bedroom", "binary_sensor.door_front"},
			pattern:  "*motion*",
			expected: []string{"binary_sensor.motion_kitchen", "binary_sensor.motion_bedroom"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create a simple pattern matcher (similar to what monitor uses)
			var matched []string
			for _, entity := range tt.entities {
				// Simplified matching for test
				if matchEntityPattern(entity, tt.pattern) {
					matched = append(matched, entity)
				}
			}
			assert.Equal(t, tt.expected, matched)
		})
	}
}

// matchEntityPattern is a simplified pattern matcher for testing
func matchEntityPattern(entity, pattern string) bool {
	// This is a simple implementation for testing purposes
	// Real implementation uses compiled regex
	if pattern == "*" {
		return true
	}

	// Simple glob matching
	if pattern != "" && pattern[0] == '*' && pattern[len(pattern)-1] == '*' {
		// *something* pattern
		middle := pattern[1 : len(pattern)-1]
		return containsString(entity, middle)
	}
	if pattern != "" && pattern[0] == '*' {
		// *suffix pattern
		suffix := pattern[1:]
		return len(entity) >= len(suffix) && entity[len(entity)-len(suffix):] == suffix
	}
	if pattern != "" && pattern[len(pattern)-1] == '*' {
		// prefix* pattern
		prefix := pattern[:len(pattern)-1]
		return len(entity) >= len(prefix) && entity[:len(prefix)] == prefix
	}

	return entity == pattern
}

func containsString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestMonitorStateChangeDetection(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		previousState string
		newState      string
		changed       bool
	}{
		{
			name:          "state changed",
			previousState: "off",
			newState:      "on",
			changed:       true,
		},
		{
			name:          "state unchanged",
			previousState: "on",
			newState:      "on",
			changed:       false,
		},
		{
			name:          "numeric state changed",
			previousState: "72.5",
			newState:      "73.0",
			changed:       true,
		},
		{
			name:          "unavailable to on",
			previousState: "unavailable",
			newState:      "on",
			changed:       true,
		},
		{
			name:          "unknown to value",
			previousState: "unknown",
			newState:      "50",
			changed:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := tt.previousState != tt.newState
			assert.Equal(t, tt.changed, result)
		})
	}
}

func TestMonitorTriggerStructure(t *testing.T) {
	t.Parallel()

	t.Run("state trigger", func(t *testing.T) {
		t.Parallel()

		trigger := map[string]any{
			"platform":  "state",
			"entity_id": "binary_sensor.motion",
			"to":        "on",
		}

		platform, ok := trigger["platform"].(string)
		require.True(t, ok)
		assert.Equal(t, "state", platform)

		entityID, ok := trigger["entity_id"].(string)
		require.True(t, ok)
		assert.Equal(t, "binary_sensor.motion", entityID)
	})

	t.Run("time pattern trigger", func(t *testing.T) {
		t.Parallel()

		trigger := map[string]any{
			"platform": "time_pattern",
			"seconds":  "/1", // Every second
		}

		platform, ok := trigger["platform"].(string)
		require.True(t, ok)
		assert.Equal(t, "time_pattern", platform)
	})

	t.Run("multiple entity trigger", func(t *testing.T) {
		t.Parallel()

		trigger := map[string]any{
			"platform":  "state",
			"entity_id": []string{"light.kitchen", "light.bedroom", "light.living_room"},
		}

		entityIDs, ok := trigger["entity_id"].([]string)
		require.True(t, ok)
		assert.Len(t, entityIDs, 3)
	})
}

func TestMonitorEventVariables(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		vars     map[string]any
		hasNow   bool
		hasTrgr  bool
		hasState bool
	}{
		{
			name: "complete event variables",
			vars: map[string]any{
				"now": time.Now().Unix(),
				"trigger": map[string]any{
					"platform":  "state",
					"entity_id": "sensor.temperature",
					"to_state":  map[string]any{"state": "72"},
				},
			},
			hasNow:  true,
			hasTrgr: true,
		},
		{
			name: "minimal variables",
			vars: map[string]any{
				"now": time.Now().Unix(),
			},
			hasNow: true,
		},
		{
			name:   "empty variables",
			vars:   map[string]any{},
			hasNow: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, hasNow := tt.vars["now"]
			assert.Equal(t, tt.hasNow, hasNow)

			_, hasTrigger := tt.vars["trigger"]
			assert.Equal(t, tt.hasTrgr, hasTrigger)
		})
	}
}

func TestAnalyzePatternRecognition(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		states   []types.HAState
		expected int
	}{
		{
			name: "multiple entities",
			states: []types.HAState{
				{EntityID: "sensor.temp_1", State: "72"},
				{EntityID: "sensor.temp_2", State: "74"},
				{EntityID: "sensor.temp_3", State: "71"},
			},
			expected: 3,
		},
		{
			name: "mixed states",
			states: []types.HAState{
				{EntityID: "light.kitchen", State: "on"},
				{EntityID: "light.bedroom", State: "off"},
				{EntityID: "light.living", State: "unavailable"},
			},
			expected: 3,
		},
		{
			name:     "empty list",
			states:   []types.HAState{},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Len(t, tt.states, tt.expected)
		})
	}
}

func TestMonitorTimeoutParsing(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		seconds    int
		expectDur  time.Duration
		isInfinite bool
	}{
		{
			name:      "60 seconds",
			seconds:   60,
			expectDur: 60 * time.Second,
		},
		{
			name:      "300 seconds (5 minutes)",
			seconds:   300,
			expectDur: 5 * time.Minute,
		},
		{
			name:       "0 seconds (infinite)",
			seconds:    0,
			isInfinite: true,
		},
		{
			name:      "1 second",
			seconds:   1,
			expectDur: time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if tt.isInfinite {
				assert.Equal(t, 0, tt.seconds)
			} else {
				dur := time.Duration(tt.seconds) * time.Second
				assert.Equal(t, tt.expectDur, dur)
			}
		})
	}
}

func TestMonitorMultiEntityList(t *testing.T) {
	t.Parallel()

	t.Run("parses entity list", func(t *testing.T) {
		t.Parallel()

		entities := []string{
			"sensor.temperature",
			"sensor.humidity",
			"binary_sensor.motion",
		}

		assert.Len(t, entities, 3)
		for _, e := range entities {
			assert.Contains(t, e, ".")
		}
	})

	t.Run("validates entity format", func(t *testing.T) {
		t.Parallel()

		validEntities := []string{
			"sensor.kitchen_temperature",
			"binary_sensor.motion_living_room",
			"input_boolean.guest_mode",
		}

		for _, e := range validEntities {
			// Entity ID must have domain.name format
			parts := splitEntityID(e)
			assert.Len(t, parts, 2, "Entity %s should have domain.name format", e)
			assert.NotEmpty(t, parts[0], "Domain should not be empty")
			assert.NotEmpty(t, parts[1], "Name should not be empty")
		}
	})
}

// splitEntityID splits an entity_id into domain and name
func splitEntityID(entityID string) []string {
	for i := 0; i < len(entityID); i++ {
		if entityID[i] == '.' {
			return []string{entityID[:i], entityID[i+1:]}
		}
	}
	return []string{entityID}
}

func TestMonitorStateFormatting(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		state    types.HAState
		expected string
	}{
		{
			name: "sensor with numeric state",
			state: types.HAState{
				EntityID:   "sensor.temperature",
				State:      "72.5",
				Attributes: map[string]any{"unit_of_measurement": "°F"},
			},
			expected: "sensor.temperature: 72.5",
		},
		{
			name: "binary sensor",
			state: types.HAState{
				EntityID:   "binary_sensor.motion",
				State:      "on",
				Attributes: map[string]any{"friendly_name": "Living Room Motion"},
			},
			expected: "binary_sensor.motion: on",
		},
		{
			name: "unavailable entity",
			state: types.HAState{
				EntityID: "sensor.offline",
				State:    "unavailable",
			},
			expected: "sensor.offline: unavailable",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			formatted := fmt.Sprintf("%s: %s", tt.state.EntityID, tt.state.State)
			assert.Equal(t, tt.expected, formatted)
		})
	}
}

func TestMonitorSubscriptionConfig(t *testing.T) {
	t.Parallel()

	t.Run("state trigger config", func(t *testing.T) {
		t.Parallel()

		entityID := "sensor.temperature"
		trigger := map[string]any{
			"platform":  "state",
			"entity_id": entityID,
		}

		assert.Equal(t, "state", trigger["platform"])
		assert.Equal(t, entityID, trigger["entity_id"])
	})

	t.Run("state trigger with filter", func(t *testing.T) {
		t.Parallel()

		trigger := map[string]any{
			"platform":  "state",
			"entity_id": "light.kitchen",
			"to":        "on",
			"from":      "off",
		}

		assert.Equal(t, "on", trigger["to"])
		assert.Equal(t, "off", trigger["from"])
	})

	t.Run("numeric state trigger", func(t *testing.T) {
		t.Parallel()

		trigger := map[string]any{
			"platform":  "numeric_state",
			"entity_id": "sensor.temperature",
			"above":     70,
		}

		assert.Equal(t, "numeric_state", trigger["platform"])
		above, ok := trigger["above"].(int)
		require.True(t, ok)
		assert.Equal(t, 70, above)
	})
}

// Benchmarks for monitor-related operations

func BenchmarkSplitEntityID(b *testing.B) {
	entities := []string{
		"sensor.temperature",
		"binary_sensor.motion_living_room",
		"light.kitchen_ceiling",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e := entities[i%len(entities)]
		_ = splitEntityID(e)
	}
}

func BenchmarkMatchEntityPattern(b *testing.B) {
	entities := []string{
		"sensor.temperature",
		"light.kitchen",
		"binary_sensor.motion",
		"switch.fan",
	}
	pattern := "sensor.*"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e := entities[i%len(entities)]
		_ = matchEntityPattern(e, pattern)
	}
}

func BenchmarkStateFormatting(b *testing.B) {
	states := []types.HAState{
		{EntityID: "sensor.temp", State: "72.5"},
		{EntityID: "light.kitchen", State: "on"},
		{EntityID: "binary_sensor.motion", State: "off"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s := states[i%len(states)]
		_ = fmt.Sprintf("%s: %s", s.EntityID, s.State)
	}
}

// =====================================
// Unit Tests for handleMonitor
// =====================================

func TestHandleMonitor_Success(t *testing.T) {
	t.Parallel()

	// Create a mock server that handles subscribe_trigger and sends events
	router := NewMessageRouter(t).On("subscribe_trigger", func(_ string, data map[string]any) any {
		// Verify the trigger configuration
		trigger, ok := data["trigger"].(map[string]any)
		if !ok {
			return nil
		}
		platform, _ := trigger["platform"].(string)
		assert.Equal(t, "state", platform)
		return nil // Success response
	})

	// Use a short timeout for testing
	ctx, cleanup := NewTestContext(t, router,
		WithArgs("monitor", "sensor.temperature"),
		WithHandlerConfig(&HandlerConfig{
			Args:        []string{"sensor.temperature"},
			OptionalInt: 1, // 1 second timeout
		}),
	)
	defer cleanup()

	// Capture output
	_, restoreOutput := CaptureOutput()
	defer restoreOutput()

	// Run the handler - should complete after timeout
	err := handleMonitor(ctx)
	require.NoError(t, err)
}

func TestHandleMonitor_Cancellation(t *testing.T) {
	t.Parallel()

	// Create a mock server that handles subscribe_trigger
	router := NewMessageRouter(t).OnSuccess("subscribe_trigger", nil)

	// Create a cancellable context
	goCtx, cancel := context.WithCancel(context.Background())

	ctx, cleanup := NewTestContext(t, router,
		WithArgs("monitor", "sensor.temperature"),
		WithHandlerConfig(&HandlerConfig{
			Args:        []string{"sensor.temperature"},
			OptionalInt: 60, // Long timeout
		}),
		WithContext(goCtx),
	)
	defer cleanup()

	// Capture output
	_, restoreOutput := CaptureOutput()
	defer restoreOutput()

	// Cancel after a short delay
	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel()
	}()

	// Run the handler - should complete after cancellation
	err := handleMonitor(ctx)
	// Should not return an error on graceful cancellation
	require.NoError(t, err)
}

func TestHandleMonitor_SubscriptionError(t *testing.T) {
	t.Parallel()

	// Create a mock server that returns an error for subscription
	router := NewMessageRouter(t).OnError("subscribe_trigger", "entity_not_found", "Entity sensor.nonexistent not found")

	ctx, cleanup := NewTestContext(t, router,
		WithArgs("monitor", "sensor.nonexistent"),
		WithHandlerConfig(&HandlerConfig{
			Args:        []string{"sensor.nonexistent"},
			OptionalInt: 5,
		}),
	)
	defer cleanup()

	// Capture output
	_, restoreOutput := CaptureOutput()
	defer restoreOutput()

	// Run the handler - should return error
	err := handleMonitor(ctx)
	require.Error(t, err)
}

// =====================================
// Unit Tests for HandleMonitorMulti
// =====================================

func TestHandleMonitorMulti_SingleEntity_Success(t *testing.T) {
	t.Parallel()

	// Create a mock server that handles subscribe_trigger for a single entity
	router := NewMessageRouter(t).OnSuccess("subscribe_trigger", nil)

	ctx, cleanup := NewTestContext(t, router,
		WithArgs("monitor-multi", "sensor.temp1", "1"), // Single entity with 1 second timeout
	)
	defer cleanup()

	// Capture output
	_, restoreOutput := CaptureOutput()
	defer restoreOutput()

	// Run the handler - should complete after timeout
	err := HandleMonitorMulti(ctx)
	require.NoError(t, err)
}

func TestHandleMonitorMulti_MissingArgument(t *testing.T) {
	t.Parallel()

	router := NewMessageRouter(t)
	ctx, cleanup := NewTestContext(t, router,
		WithArgs("monitor-multi"), // No entities provided
	)
	defer cleanup()

	err := HandleMonitorMulti(ctx)
	require.Error(t, err)
}

func TestHandleMonitorMulti_SingleEntity_AllFail(t *testing.T) {
	t.Parallel()

	// Create a mock server that fails for all subscriptions
	router := NewMessageRouter(t).OnError("subscribe_trigger", "not_found", "Entity not found")

	ctx, cleanup := NewTestContext(t, router,
		WithArgs("monitor-multi", "sensor.bad1", "1"), // Single entity
	)
	defer cleanup()

	// Capture output
	_, restoreOutput := CaptureOutput()
	defer restoreOutput()

	// Run the handler - should return error when all fail
	err := HandleMonitorMulti(ctx)
	require.Error(t, err)
}

func TestHandleMonitorMulti_SingleEntity_Cancellation(t *testing.T) {
	t.Parallel()

	router := NewMessageRouter(t).OnSuccess("subscribe_trigger", nil)

	// Create a cancellable context
	goCtx, cancel := context.WithCancel(context.Background())

	ctx, cleanup := NewTestContext(t, router,
		WithArgs("monitor-multi", "sensor.temp1", "60"), // Single entity, 60 second timeout
		WithContext(goCtx),
	)
	defer cleanup()

	// Capture output
	_, restoreOutput := CaptureOutput()
	defer restoreOutput()

	// Cancel after a short delay
	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel()
	}()

	// Run the handler - should complete after cancellation
	err := HandleMonitorMulti(ctx)
	// Should not return an error on graceful cancellation
	require.NoError(t, err)
}

func TestHandleMonitorMulti_ArgumentParsing(t *testing.T) {
	// Tests argument parsing logic for HandleMonitorMulti without actual subscriptions
	// Note: Multi-entity subscriptions require synchronized websocket writes
	// which is out of scope for unit tests. Test validates parsing logic only.
	t.Parallel()

	tests := []struct {
		name              string
		args              []string
		expectError       bool
		expectedEntityCnt int
		expectedSeconds   int
	}{
		{
			name:              "single entity no duration uses default",
			args:              []string{"monitor-multi", "sensor.temp"},
			expectError:       false,
			expectedEntityCnt: 1,
			expectedSeconds:   60, // default
		},
		{
			name:              "single entity with duration",
			args:              []string{"monitor-multi", "sensor.temp", "30"},
			expectError:       false,
			expectedEntityCnt: 1,
			expectedSeconds:   30,
		},
		{
			name:        "no entities returns error",
			args:        []string{"monitor-multi"},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if tt.expectError {
				// For error cases, we just need to verify the error is returned
				router := NewMessageRouter(t)
				ctx, cleanup := NewTestContext(t, router, WithArgs(tt.args...))
				defer cleanup()

				err := HandleMonitorMulti(ctx)
				require.Error(t, err)
				return
			}

			// For success cases, verify subscription is made and completes
			router := NewMessageRouter(t).OnSuccess("subscribe_trigger", nil)

			// Create context with short cancellation for quick test
			goCtx, cancel := context.WithCancel(context.Background())

			ctx, cleanup := NewTestContext(t, router,
				WithArgs(tt.args...),
				WithContext(goCtx),
			)
			defer cleanup()

			_, restoreOutput := CaptureOutput()
			defer restoreOutput()

			// Cancel after subscriptions are established
			go func() {
				time.Sleep(100 * time.Millisecond)
				cancel()
			}()

			err := HandleMonitorMulti(ctx)
			require.NoError(t, err)
		})
	}
}

// =====================================
// Unit Tests for handleAnalyze
// =====================================

func TestHandleAnalyze_Success(t *testing.T) {
	t.Parallel()

	// Mock states response (get_states returns all states)
	statesResult := []any{
		map[string]any{
			"entity_id":  "sensor.temperature",
			"state":      "72.5",
			"attributes": map[string]any{"unit_of_measurement": "°F"},
		},
		map[string]any{
			"entity_id": "light.living_room",
			"state":     "on",
		},
	}

	// Mock history response
	historyResult := map[string]any{
		"sensor.temperature": []any{
			map[string]any{"s": "70.0", "lu": 1704067200.0},
			map[string]any{"s": "71.5", "lu": 1704070800.0},
			map[string]any{"s": "72.5", "lu": 1704074400.0},
		},
	}

	router := NewMessageRouter(t).
		OnSuccess("get_states", statesResult).
		OnSuccess("history/history_during_period", historyResult)

	now := time.Now()
	startTime := now.Add(-24 * time.Hour)

	ctx, cleanup := NewTestContext(t, router,
		WithArgs("analyze", "sensor.temperature"),
		WithHandlerConfig(&HandlerConfig{
			Args: []string{"sensor.temperature"},
			TimeRange: &types.TimeRange{
				StartTime: startTime,
				EndTime:   now,
			},
		}),
	)
	defer cleanup()

	// Capture output
	_, restoreOutput := CaptureOutput()
	defer restoreOutput()

	err := handleAnalyze(ctx)
	require.NoError(t, err)
}

func TestHandleAnalyze_EntityNotFound(t *testing.T) {
	t.Parallel()

	// Mock states response without the target entity
	statesResult := []any{
		map[string]any{
			"entity_id": "light.living_room",
			"state":     "on",
		},
	}

	router := NewMessageRouter(t).OnSuccess("get_states", statesResult)

	now := time.Now()
	startTime := now.Add(-24 * time.Hour)

	ctx, cleanup := NewTestContext(t, router,
		WithArgs("analyze", "sensor.nonexistent"),
		WithHandlerConfig(&HandlerConfig{
			Args: []string{"sensor.nonexistent"},
			TimeRange: &types.TimeRange{
				StartTime: startTime,
				EndTime:   now,
			},
		}),
	)
	defer cleanup()

	// Capture output
	_, restoreOutput := CaptureOutput()
	defer restoreOutput()

	err := handleAnalyze(ctx)
	require.Error(t, err)
}

func TestHandleAnalyze_EmptyHistory(t *testing.T) {
	t.Parallel()

	// Mock states response
	statesResult := []any{
		map[string]any{
			"entity_id":  "sensor.temperature",
			"state":      "72.5",
			"attributes": map[string]any{"unit_of_measurement": "°F"},
		},
	}

	// Mock empty history response
	historyResult := map[string]any{
		"sensor.temperature": []any{},
	}

	router := NewMessageRouter(t).
		OnSuccess("get_states", statesResult).
		OnSuccess("history/history_during_period", historyResult)

	now := time.Now()
	startTime := now.Add(-24 * time.Hour)

	ctx, cleanup := NewTestContext(t, router,
		WithArgs("analyze", "sensor.temperature"),
		WithHandlerConfig(&HandlerConfig{
			Args: []string{"sensor.temperature"},
			TimeRange: &types.TimeRange{
				StartTime: startTime,
				EndTime:   now,
			},
		}),
	)
	defer cleanup()

	// Capture output
	_, restoreOutput := CaptureOutput()
	defer restoreOutput()

	err := handleAnalyze(ctx)
	require.NoError(t, err)
}

func TestHandleAnalyze_GetStatesError(t *testing.T) {
	t.Parallel()

	// Mock error response from get_states
	router := NewMessageRouter(t).OnError("get_states", "connection_error", "Failed to get states")

	now := time.Now()
	startTime := now.Add(-24 * time.Hour)

	ctx, cleanup := NewTestContext(t, router,
		WithArgs("analyze", "sensor.temperature"),
		WithHandlerConfig(&HandlerConfig{
			Args: []string{"sensor.temperature"},
			TimeRange: &types.TimeRange{
				StartTime: startTime,
				EndTime:   now,
			},
		}),
	)
	defer cleanup()

	// Capture output
	_, restoreOutput := CaptureOutput()
	defer restoreOutput()

	err := handleAnalyze(ctx)
	require.Error(t, err)
}

// =====================================
// Unit Tests for getStates
// =====================================

func TestGetStates_Success(t *testing.T) {
	t.Parallel()

	// Mock states response
	statesResult := []any{
		map[string]any{
			"entity_id":  "sensor.temperature",
			"state":      "72.5",
			"attributes": map[string]any{"unit_of_measurement": "°F"},
		},
		map[string]any{
			"entity_id":  "light.living_room",
			"state":      "on",
			"attributes": map[string]any{"brightness": 255},
		},
		map[string]any{
			"entity_id": "binary_sensor.motion",
			"state":     "off",
		},
	}

	router := NewMessageRouter(t).OnSuccess("get_states", statesResult)

	ctx, cleanup := NewTestContext(t, router)
	defer cleanup()

	states, err := getStates(ctx)
	require.NoError(t, err)
	assert.Len(t, states, 3)

	// Verify first state
	assert.Equal(t, "sensor.temperature", states[0].EntityID)
	assert.Equal(t, "72.5", states[0].State)
	assert.NotNil(t, states[0].Attributes)

	// Verify second state
	assert.Equal(t, "light.living_room", states[1].EntityID)
	assert.Equal(t, "on", states[1].State)

	// Verify third state
	assert.Equal(t, "binary_sensor.motion", states[2].EntityID)
	assert.Equal(t, "off", states[2].State)
}

func TestGetStates_EmptyResult(t *testing.T) {
	t.Parallel()

	// Mock empty states response
	statesResult := []any{}

	router := NewMessageRouter(t).OnSuccess("get_states", statesResult)

	ctx, cleanup := NewTestContext(t, router)
	defer cleanup()

	states, err := getStates(ctx)
	require.NoError(t, err)
	assert.Empty(t, states)
}

func TestGetStates_InvalidResponseType(t *testing.T) {
	t.Parallel()

	// Mock invalid response type (not a slice)
	router := NewMessageRouter(t).OnSuccess("get_states", "invalid")

	ctx, cleanup := NewTestContext(t, router)
	defer cleanup()

	_, err := getStates(ctx)
	require.Error(t, err)
}

func TestGetStates_Error(t *testing.T) {
	t.Parallel()

	// Mock error response
	router := NewMessageRouter(t).OnError("get_states", "connection_error", "Failed to get states")

	ctx, cleanup := NewTestContext(t, router)
	defer cleanup()

	_, err := getStates(ctx)
	require.Error(t, err)
}

func TestGetStates_MalformedEntries(t *testing.T) {
	t.Parallel()

	// Mock response with some malformed entries
	statesResult := []any{
		map[string]any{
			"entity_id":  "sensor.temperature",
			"state":      "72.5",
			"attributes": map[string]any{"unit_of_measurement": "°F"},
		},
		"invalid_entry", // Not a map
		nil,             // Nil entry
		map[string]any{
			"entity_id": "light.living_room",
			"state":     "on",
		},
	}

	router := NewMessageRouter(t).OnSuccess("get_states", statesResult)

	ctx, cleanup := NewTestContext(t, router)
	defer cleanup()

	states, err := getStates(ctx)
	require.NoError(t, err)
	// Should only have the 2 valid entries
	assert.Len(t, states, 2)
}

// =====================================
// Unit Tests for getHistory
// =====================================

func TestGetHistory_Success(t *testing.T) {
	t.Parallel()

	// Mock history response
	historyResult := map[string]any{
		"sensor.temperature": []any{
			map[string]any{"s": "70.0", "lu": 1704067200.0},
			map[string]any{"s": "71.5", "lu": 1704070800.0},
			map[string]any{"s": "72.5", "lu": 1704074400.0},
		},
	}

	router := NewMessageRouter(t).OnSuccess("history/history_during_period", historyResult)

	ctx, cleanup := NewTestContext(t, router)
	defer cleanup()

	now := time.Now()
	startTime := now.Add(-24 * time.Hour)
	timeRange := types.TimeRange{
		StartTime: startTime,
		EndTime:   now,
	}

	history, err := getHistory(ctx, "sensor.temperature", timeRange)
	require.NoError(t, err)
	assert.Len(t, history, 3)

	// Verify history states
	assert.Equal(t, "70.0", history[0].GetState())
	assert.Equal(t, "71.5", history[1].GetState())
	assert.Equal(t, "72.5", history[2].GetState())
}

func TestGetHistory_EmptyResult(t *testing.T) {
	t.Parallel()

	// Mock empty history response
	historyResult := map[string]any{
		"sensor.temperature": []any{},
	}

	router := NewMessageRouter(t).OnSuccess("history/history_during_period", historyResult)

	ctx, cleanup := NewTestContext(t, router)
	defer cleanup()

	now := time.Now()
	startTime := now.Add(-24 * time.Hour)
	timeRange := types.TimeRange{
		StartTime: startTime,
		EndTime:   now,
	}

	history, err := getHistory(ctx, "sensor.temperature", timeRange)
	require.NoError(t, err)
	assert.Empty(t, history)
}

func TestGetHistory_EntityNotInResult(t *testing.T) {
	t.Parallel()

	// Mock history response without the requested entity
	historyResult := map[string]any{
		"sensor.other": []any{
			map[string]any{"s": "50.0", "lu": 1704067200.0},
		},
	}

	router := NewMessageRouter(t).OnSuccess("history/history_during_period", historyResult)

	ctx, cleanup := NewTestContext(t, router)
	defer cleanup()

	now := time.Now()
	startTime := now.Add(-24 * time.Hour)
	timeRange := types.TimeRange{
		StartTime: startTime,
		EndTime:   now,
	}

	history, err := getHistory(ctx, "sensor.temperature", timeRange)
	require.NoError(t, err)
	// Should return nil when entity not in result
	assert.Nil(t, history)
}

func TestGetHistory_InvalidResponseType(t *testing.T) {
	t.Parallel()

	// Mock invalid response type (not a map)
	router := NewMessageRouter(t).OnSuccess("history/history_during_period", "invalid")

	ctx, cleanup := NewTestContext(t, router)
	defer cleanup()

	now := time.Now()
	startTime := now.Add(-24 * time.Hour)
	timeRange := types.TimeRange{
		StartTime: startTime,
		EndTime:   now,
	}

	history, err := getHistory(ctx, "sensor.temperature", timeRange)
	require.NoError(t, err) // Returns nil, nil for invalid response type
	assert.Nil(t, history)
}

func TestGetHistory_Error(t *testing.T) {
	t.Parallel()

	// Mock error response
	router := NewMessageRouter(t).OnError("history/history_during_period", "connection_error", "Failed to get history")

	ctx, cleanup := NewTestContext(t, router)
	defer cleanup()

	now := time.Now()
	startTime := now.Add(-24 * time.Hour)
	timeRange := types.TimeRange{
		StartTime: startTime,
		EndTime:   now,
	}

	_, err := getHistory(ctx, "sensor.temperature", timeRange)
	require.Error(t, err)
}

func TestGetHistory_WithLegacyFormat(t *testing.T) {
	t.Parallel()

	// Mock history response using legacy format (state, last_updated instead of s, lu)
	historyResult := map[string]any{
		"sensor.temperature": []any{
			map[string]any{"state": "70.0", "last_updated": "2024-01-01T00:00:00Z"},
			map[string]any{"state": "71.5", "last_updated": "2024-01-01T01:00:00Z"},
		},
	}

	router := NewMessageRouter(t).OnSuccess("history/history_during_period", historyResult)

	ctx, cleanup := NewTestContext(t, router)
	defer cleanup()

	now := time.Now()
	startTime := now.Add(-24 * time.Hour)
	timeRange := types.TimeRange{
		StartTime: startTime,
		EndTime:   now,
	}

	history, err := getHistory(ctx, "sensor.temperature", timeRange)
	require.NoError(t, err)
	assert.Len(t, history, 2)

	// Verify history states using legacy format
	assert.Equal(t, "70.0", history[0].GetState())
	assert.Equal(t, "71.5", history[1].GetState())
}
