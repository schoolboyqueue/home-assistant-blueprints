package handlers

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/home-assistant-blueprints/ha-ws-client-go/internal/types"
)

func TestFindEntityByID(t *testing.T) {
	t.Parallel()

	states := []types.HAState{
		{EntityID: "light.kitchen", State: "on"},
		{EntityID: "sensor.temperature", State: "72"},
		{EntityID: "input_boolean.presence", State: "on"},
	}

	tests := []struct {
		name     string
		entityID string
		found    bool
		state    string
	}{
		{
			name:     "finds existing entity",
			entityID: "light.kitchen",
			found:    true,
			state:    "on",
		},
		{
			name:     "finds sensor entity",
			entityID: "sensor.temperature",
			found:    true,
			state:    "72",
		},
		{
			name:     "returns nil for non-existent entity",
			entityID: "light.bedroom",
			found:    false,
		},
		{
			name:     "returns nil for empty string",
			entityID: "",
			found:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := findEntityByID(states, tt.entityID)
			if tt.found {
				assert.NotNil(t, result)
				assert.Equal(t, tt.state, result.State)
			} else {
				assert.Nil(t, result)
			}
		})
	}
}

func TestFindEntityByID_EmptyStates(t *testing.T) {
	t.Parallel()

	result := findEntityByID([]types.HAState{}, "light.kitchen")
	assert.Nil(t, result)
}

func TestFindStatesByContext(t *testing.T) {
	t.Parallel()

	contextA := &types.HAContext{ID: "ctx-a", ParentID: ""}
	contextB := &types.HAContext{ID: "ctx-b", ParentID: "ctx-a"}
	contextC := &types.HAContext{ID: "ctx-c", ParentID: ""}

	states := []types.HAState{
		{EntityID: "light.kitchen", State: "on", Context: contextA},
		{EntityID: "light.bedroom", State: "off", Context: contextB},
		{EntityID: "sensor.temp", State: "72", Context: contextC},
		{EntityID: "switch.fan", State: "on", Context: contextA},
	}

	tests := []struct {
		name      string
		contextID string
		expected  int
	}{
		{
			name:      "finds states with matching context ID",
			contextID: "ctx-a",
			expected:  3, // kitchen (direct), bedroom (parent), fan (direct)
		},
		{
			name:      "finds states with matching context ID only",
			contextID: "ctx-c",
			expected:  1,
		},
		{
			name:      "returns empty for non-existent context",
			contextID: "ctx-z",
			expected:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := findStatesByContext(states, tt.contextID)
			assert.Len(t, result, tt.expected)
		})
	}
}

func TestFindStatesByContext_NilContexts(t *testing.T) {
	t.Parallel()

	states := []types.HAState{
		{EntityID: "light.kitchen", State: "on", Context: nil},
		{EntityID: "light.bedroom", State: "off", Context: &types.HAContext{ID: "ctx-a"}},
	}

	result := findStatesByContext(states, "ctx-a")
	assert.Len(t, result, 1)
	assert.Equal(t, "light.bedroom", result[0].EntityID)
}

func TestAddParentContextMatches(t *testing.T) {
	t.Parallel()

	contextParent := &types.HAContext{ID: "parent-ctx"}
	contextChild := &types.HAContext{ID: "child-ctx", ParentID: "parent-ctx"}

	states := []types.HAState{
		{EntityID: "automation.trigger", State: "on", Context: contextParent},
		{EntityID: "light.kitchen", State: "on", Context: contextChild},
		{EntityID: "sensor.unrelated", State: "72", Context: &types.HAContext{ID: "other"}},
	}

	tests := []struct {
		name           string
		initialMatches []types.HAState
		parentID       string
		expectedLen    int
	}{
		{
			name:           "adds parent context match",
			initialMatches: []types.HAState{states[1]}, // light.kitchen
			parentID:       "parent-ctx",
			expectedLen:    2, // adds automation.trigger
		},
		{
			name:           "does not add duplicates",
			initialMatches: []types.HAState{states[0], states[1]}, // already has automation.trigger
			parentID:       "parent-ctx",
			expectedLen:    2,
		},
		{
			name:           "returns original if no parent matches",
			initialMatches: []types.HAState{states[1]},
			parentID:       "nonexistent",
			expectedLen:    1,
		},
		{
			name:           "handles empty initial matches",
			initialMatches: []types.HAState{},
			parentID:       "parent-ctx",
			expectedLen:    1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := addParentContextMatches(tt.initialMatches, states, tt.parentID)
			assert.Len(t, result, tt.expectedLen)
		})
	}
}

func TestFormatContextInfo(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		context  *types.HAContext
		expected string
	}{
		{
			name:     "nil context returns empty string",
			context:  nil,
			expected: "",
		},
		{
			name:     "context without parent",
			context:  &types.HAContext{ID: "ctx-123"},
			expected: " (context: ctx-123)",
		},
		{
			name:     "context with parent",
			context:  &types.HAContext{ID: "ctx-456", ParentID: "ctx-123"},
			expected: " (context: ctx-456, parent: ctx-123)",
		},
		{
			name:     "context with empty parent treated as no parent",
			context:  &types.HAContext{ID: "ctx-789", ParentID: ""},
			expected: " (context: ctx-789)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := formatContextInfo(tt.context)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Benchmark for findEntityByID
func BenchmarkFindEntityByID(b *testing.B) {
	states := make([]types.HAState, 1000)
	for i := range states {
		states[i] = types.HAState{
			EntityID: "entity_" + string(rune('a'+i%26)) + "_" + string(rune('0'+i%10)),
			State:    "on",
		}
	}
	// Add target at the end
	states = append(states, types.HAState{EntityID: "target.entity", State: "found"})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = findEntityByID(states, "target.entity")
	}
}

// Benchmark for findStatesByContext
func BenchmarkFindStatesByContext(b *testing.B) {
	states := make([]types.HAState, 1000)
	for i := range states {
		ctx := &types.HAContext{ID: "ctx-" + string(rune('a'+i%26))}
		if i%10 == 0 {
			ctx.ParentID = "target-ctx"
		}
		states[i] = types.HAState{
			EntityID: "entity_" + string(rune('a'+i%26)),
			State:    "on",
			Context:  ctx,
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = findStatesByContext(states, "target-ctx")
	}
}
