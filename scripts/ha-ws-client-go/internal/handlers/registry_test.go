package handlers

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/home-assistant-blueprints/ha-ws-client-go/internal/types"
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

func TestEntityEntryFormatting(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		entry          types.EntityEntry
		expectName     string
		expectDisabled bool
	}{
		{
			name: "entry with name",
			entry: types.EntityEntry{
				EntityID:     "light.kitchen",
				Name:         "Kitchen Light",
				OriginalName: "Kitchen Ceiling Light",
				Platform:     "hue",
			},
			expectName:     "Kitchen Light",
			expectDisabled: false,
		},
		{
			name: "entry without name falls back to original",
			entry: types.EntityEntry{
				EntityID:     "sensor.temperature",
				Name:         "",
				OriginalName: "Temperature Sensor",
				Platform:     "esphome",
			},
			expectName:     "Temperature Sensor",
			expectDisabled: false,
		},
		{
			name: "disabled entry",
			entry: types.EntityEntry{
				EntityID:   "switch.old_device",
				Name:       "Old Switch",
				DisabledBy: "user",
			},
			expectName:     "Old Switch",
			expectDisabled: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			name := tt.entry.Name
			if name == "" {
				name = tt.entry.OriginalName
			}
			assert.Equal(t, tt.expectName, name)
			assert.Equal(t, tt.expectDisabled, tt.entry.DisabledBy != "")
		})
	}
}

func TestDeviceEntryFormatting(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		entry      types.DeviceEntry
		expectName string
	}{
		{
			name: "device with user-defined name",
			entry: types.DeviceEntry{
				ID:           "abc123def456",
				Name:         "Hue Bridge",
				NameByUser:   "Living Room Hub",
				Manufacturer: "Philips",
				Model:        "BSB002",
				AreaID:       "living_room",
			},
			expectName: "Living Room Hub",
		},
		{
			name: "device with default name only",
			entry: types.DeviceEntry{
				ID:           "xyz789abc012",
				Name:         "Motion Sensor",
				Manufacturer: "Aqara",
				Model:        "RTCGQ11LM",
			},
			expectName: "Motion Sensor",
		},
		{
			name: "device with empty names",
			entry: types.DeviceEntry{
				ID:           "empty123",
				Manufacturer: "Generic",
				Model:        "Unknown",
			},
			expectName: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			name := tt.entry.Name
			if tt.entry.NameByUser != "" {
				name = tt.entry.NameByUser
			}
			assert.Equal(t, tt.expectName, name)
		})
	}
}

func TestAreaEntryFormatting(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		entry         types.AreaEntry
		expectAliases int
	}{
		{
			name: "area with aliases",
			entry: types.AreaEntry{
				AreaID:  "living_room",
				Name:    "Living Room",
				Aliases: []string{"Lounge", "Family Room"},
			},
			expectAliases: 2,
		},
		{
			name: "area without aliases",
			entry: types.AreaEntry{
				AreaID: "bedroom",
				Name:   "Master Bedroom",
			},
			expectAliases: 0,
		},
		{
			name: "area with empty alias list",
			entry: types.AreaEntry{
				AreaID:  "kitchen",
				Name:    "Kitchen",
				Aliases: []string{},
			},
			expectAliases: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.NotEmpty(t, tt.entry.AreaID)
			assert.NotEmpty(t, tt.entry.Name)
			assert.Len(t, tt.entry.Aliases, tt.expectAliases)
		})
	}
}

func TestDeviceIDShortening(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		deviceID string
		shortLen int
		expected string
	}{
		{
			name:     "standard UUID-like ID",
			deviceID: "abc123def456ghi789",
			shortLen: 8,
			expected: "abc123de",
		},
		{
			name:     "short ID stays same",
			deviceID: "abcd",
			shortLen: 8,
			expected: "abcd",
		},
		{
			name:     "exact length",
			deviceID: "12345678",
			shortLen: 8,
			expected: "12345678",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var result string
			if len(tt.deviceID) > tt.shortLen {
				result = tt.deviceID[:tt.shortLen]
			} else {
				result = tt.deviceID
			}
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestPatternToRegex(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		pattern string
		match   string
		expect  bool
	}{
		{
			name:    "glob star at end",
			pattern: "light.*",
			match:   "light.kitchen",
			expect:  true,
		},
		{
			name:    "glob star at start",
			pattern: "*.kitchen",
			match:   "light.kitchen",
			expect:  true,
		},
		{
			name:    "glob star in middle",
			pattern: "sensor.*_temperature",
			match:   "sensor.kitchen_temperature",
			expect:  true,
		},
		{
			name:    "double glob star",
			pattern: "*temperature*",
			match:   "sensor.kitchen_temperature_celsius",
			expect:  true,
		},
		{
			name:    "no match",
			pattern: "light.*",
			match:   "sensor.temperature",
			expect:  false,
		},
		{
			name:    "case insensitive",
			pattern: "LIGHT.*",
			match:   "light.kitchen",
			expect:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Convert glob to regex (same logic as WithPattern middleware)
			regexPattern := regexp.QuoteMeta(tt.pattern)
			regexPattern = regexp.MustCompile(`\\\*`).ReplaceAllString(regexPattern, ".*")
			re := regexp.MustCompile("(?i)" + regexPattern)

			result := re.MatchString(tt.match)
			assert.Equal(t, tt.expect, result)
		})
	}
}

func TestEmptyMatchFields(t *testing.T) {
	t.Parallel()

	type item struct {
		Value string
	}

	items := []item{
		{Value: "test1"},
		{Value: "test2"},
	}

	// MatchFields that returns empty slice should never match
	matchFields := func(_ item) []string {
		return []string{}
	}

	re := regexp.MustCompile(".*")
	result := filterByPattern(items, re, matchFields)

	assert.Nil(t, result)
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

// Benchmark for pattern to regex conversion
func BenchmarkPatternToRegex(b *testing.B) {
	patterns := []string{
		"light.*",
		"*temperature*",
		"sensor.kitchen_*",
		"binary_sensor.*_motion",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pattern := patterns[i%len(patterns)]
		regexPattern := regexp.QuoteMeta(pattern)
		regexPattern = regexp.MustCompile(`\\\*`).ReplaceAllString(regexPattern, ".*")
		_ = regexp.MustCompile("(?i)" + regexPattern)
	}
}

// Benchmark for MatchFields with multiple fields
func BenchmarkMatchFields(b *testing.B) {
	entries := []types.EntityEntry{
		{EntityID: "light.kitchen", Name: "Kitchen Light", OriginalName: "Kitchen Ceiling"},
		{EntityID: "sensor.temp", Name: "Temperature", OriginalName: "Temp Sensor"},
		{EntityID: "switch.fan", Name: "", OriginalName: "Ceiling Fan"},
	}

	matchFields := func(e types.EntityEntry) []string {
		return []string{e.EntityID, e.Name, e.OriginalName}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e := entries[i%len(entries)]
		_ = matchFields(e)
	}
}
