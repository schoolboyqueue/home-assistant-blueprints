package handlers

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFilterByPattern(t *testing.T) {
	t.Parallel()

	type testItem struct {
		ID   string
		Name string
	}

	matchFields := func(item testItem) []string {
		return []string{item.ID, item.Name}
	}

	tests := []struct {
		name     string
		items    []testItem
		pattern  string
		expected []testItem
	}{
		{
			name: "filters matching items",
			items: []testItem{
				{ID: "light.kitchen", Name: "Kitchen Light"},
				{ID: "sensor.temp", Name: "Temperature"},
				{ID: "light.bedroom", Name: "Bedroom Light"},
			},
			pattern: "light.*",
			expected: []testItem{
				{ID: "light.kitchen", Name: "Kitchen Light"},
				{ID: "light.bedroom", Name: "Bedroom Light"},
			},
		},
		{
			name: "matches by name field",
			items: []testItem{
				{ID: "sensor.1", Name: "Kitchen Temperature"},
				{ID: "sensor.2", Name: "Bedroom Temperature"},
				{ID: "sensor.3", Name: "Living Room Humidity"},
			},
			pattern: "*Temperature*",
			expected: []testItem{
				{ID: "sensor.1", Name: "Kitchen Temperature"},
				{ID: "sensor.2", Name: "Bedroom Temperature"},
			},
		},
		{
			name: "no matches returns empty slice",
			items: []testItem{
				{ID: "light.kitchen", Name: "Kitchen Light"},
				{ID: "light.bedroom", Name: "Bedroom Light"},
			},
			pattern:  "sensor.*",
			expected: nil,
		},
		{
			name:     "empty items returns nil",
			items:    []testItem{},
			pattern:  ".*",
			expected: nil,
		},
		{
			name: "case insensitive matching",
			items: []testItem{
				{ID: "LIGHT.KITCHEN", Name: "Kitchen Light"},
				{ID: "light.bedroom", Name: "Bedroom Light"},
			},
			pattern: "light.*",
			expected: []testItem{
				{ID: "LIGHT.KITCHEN", Name: "Kitchen Light"},
				{ID: "light.bedroom", Name: "Bedroom Light"},
			},
		},
		{
			name: "matches only once per item",
			items: []testItem{
				{ID: "light.light", Name: "light"},
			},
			pattern: "light*",
			expected: []testItem{
				{ID: "light.light", Name: "light"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Compile pattern (convert glob to regex like WithPattern does)
			regexPattern := regexp.QuoteMeta(tt.pattern)
			regexPattern = regexp.MustCompile(`\\\*`).ReplaceAllString(regexPattern, ".*")
			re := regexp.MustCompile("(?i)" + regexPattern)

			result := filterByPattern(tt.items, re, matchFields)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFilterByPattern_MultipleFields(t *testing.T) {
	t.Parallel()

	type entity struct {
		ID           string
		Name         string
		OriginalName string
	}

	// Match against three fields (like EntityEntry)
	matchFields := func(e entity) []string {
		return []string{e.ID, e.Name, e.OriginalName}
	}

	items := []entity{
		{ID: "sensor.temp1", Name: "Temperature", OriginalName: "Temp Sensor 1"},
		{ID: "sensor.temp2", Name: "Humidity", OriginalName: "Temp Sensor 2"},
		{ID: "light.kitchen", Name: "Kitchen", OriginalName: "Kitchen Light"},
	}

	// Pattern matching on OriginalName field
	re := regexp.MustCompile("(?i).*Temp Sensor.*")
	result := filterByPattern(items, re, matchFields)

	assert.Len(t, result, 2)
	assert.Equal(t, "sensor.temp1", result[0].ID)
	assert.Equal(t, "sensor.temp2", result[1].ID)
}

func TestRegistryConfig(t *testing.T) {
	t.Parallel()

	// Test that RegistryConfig can be properly constructed with type parameters
	t.Run("constructs with string type", func(t *testing.T) {
		t.Parallel()

		cfg := RegistryConfig[string]{
			MessageType: "test/message",
			Title:       "Test Title",
			Command:     "test",
			MatchFields: func(s string) []string { return []string{s} },
			Formatter:   func(s string, _ int) string { return s },
		}

		assert.Equal(t, "test/message", cfg.MessageType)
		assert.Equal(t, "Test Title", cfg.Title)
		assert.Equal(t, "test", cfg.Command)
		assert.Equal(t, []string{"hello"}, cfg.MatchFields("hello"))
		assert.Equal(t, "world", cfg.Formatter("world", 0))
	})

	t.Run("constructs with struct type", func(t *testing.T) {
		t.Parallel()

		type item struct {
			ID   string
			Name string
		}

		cfg := RegistryConfig[item]{
			MessageType: "config/registry/list",
			Title:       "Registry",
			Command:     "registry",
			MatchFields: func(i item) []string { return []string{i.ID, i.Name} },
			Formatter:   func(i item, _ int) string { return i.ID + ": " + i.Name },
		}

		// Verify all fields are correctly set
		assert.Equal(t, "config/registry/list", cfg.MessageType)
		assert.Equal(t, "Registry", cfg.Title)
		assert.Equal(t, "registry", cfg.Command)

		testItem := item{ID: "1", Name: "Test"}
		assert.Equal(t, []string{"1", "Test"}, cfg.MatchFields(testItem))
		assert.Equal(t, "1: Test", cfg.Formatter(testItem, 0))
	})
}

// Benchmark for filterByPattern
func BenchmarkFilterByPattern(b *testing.B) {
	type item struct {
		ID   string
		Name string
	}

	items := make([]item, 1000)
	for i := range items {
		if i%3 == 0 {
			items[i] = item{ID: "light.entity_" + string(rune('a'+i%26)), Name: "Light"}
		} else {
			items[i] = item{ID: "sensor.entity_" + string(rune('a'+i%26)), Name: "Sensor"}
		}
	}

	matchFields := func(it item) []string {
		return []string{it.ID, it.Name}
	}

	re := regexp.MustCompile("(?i)light.*")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = filterByPattern(items, re, matchFields)
	}
}
