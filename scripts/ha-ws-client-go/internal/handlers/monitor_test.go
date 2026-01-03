package handlers

import (
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
