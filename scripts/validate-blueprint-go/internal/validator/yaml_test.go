package validator

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadYAML(t *testing.T) {
	t.Parallel()

	t.Run("valid YAML file", func(t *testing.T) {
		t.Parallel()
		// Create temp file with valid YAML (note: no leading newline to avoid YAML issues)
		content := `blueprint:
  name: Test Blueprint
  description: A test blueprint
  domain: automation
  input:
    my_entity:
      name: My Entity
      selector:
        entity: {}
trigger:
  - platform: state
    entity_id: light.test
action:
  - service: light.turn_on
`
		tmpFile := createTempYAMLFile(t, content)
		defer os.Remove(tmpFile)

		v := New(tmpFile)
		result := v.LoadYAML()

		assert.True(t, result)
		assert.Empty(t, v.Errors, "Errors: %v", v.Errors)
		assert.NotEmpty(t, v.Data)
		assert.Contains(t, v.Data, "blueprint")
		assert.Contains(t, v.Data, "trigger")
		assert.Contains(t, v.Data, "action")
	})

	t.Run("YAML with !input tags", func(t *testing.T) {
		t.Parallel()
		content := `blueprint:
  name: Test
  input:
    target:
      selector:
        entity: {}
action:
  - service: light.turn_on
    target:
      entity_id: !input target
`
		tmpFile := createTempYAMLFile(t, content)
		defer os.Remove(tmpFile)

		v := New(tmpFile)
		result := v.LoadYAML()

		assert.True(t, result)
		assert.Empty(t, v.Errors, "Errors: %v", v.Errors)
		// Check that !input was converted to string
		if actions, ok := v.Data["action"].([]interface{}); ok && len(actions) > 0 {
			if action, ok := actions[0].(map[string]interface{}); ok {
				if target, ok := action["target"].(map[string]interface{}); ok {
					if entityID, ok := target["entity_id"].(string); ok {
						assert.Contains(t, entityID, "!input target")
					}
				}
			}
		}
	})

	t.Run("file not found", func(t *testing.T) {
		t.Parallel()
		v := New("/nonexistent/path/to/file.yaml")
		result := v.LoadYAML()

		assert.False(t, result)
		require.Len(t, v.Errors, 1)
		assert.Contains(t, v.Errors[0], "failed to read file")
	})

	t.Run("invalid YAML syntax", func(t *testing.T) {
		t.Parallel()
		content := `blueprint:
  name: Test
  invalid: [unclosed bracket
`
		tmpFile := createTempYAMLFile(t, content)
		defer os.Remove(tmpFile)

		v := New(tmpFile)
		result := v.LoadYAML()

		assert.False(t, result)
		require.Len(t, v.Errors, 1)
		assert.Contains(t, v.Errors[0], "YAML syntax error")
	})

	t.Run("empty YAML file", func(t *testing.T) {
		t.Parallel()
		content := ""
		tmpFile := createTempYAMLFile(t, content)
		defer os.Remove(tmpFile)

		v := New(tmpFile)
		result := v.LoadYAML()

		// Empty YAML is technically valid, but will result in nil data
		assert.True(t, result)
		assert.Empty(t, v.Errors)
	})

	t.Run("YAML with nested input references", func(t *testing.T) {
		t.Parallel()
		content := `variables:
  entity: !input target_entity
  brightness: !input brightness_level
`
		tmpFile := createTempYAMLFile(t, content)
		defer os.Remove(tmpFile)

		v := New(tmpFile)
		result := v.LoadYAML()

		assert.True(t, result)
		assert.Empty(t, v.Errors, "Errors: %v", v.Errors)
	})
}

// createTempYAMLFile creates a temporary YAML file for testing
func createTempYAMLFile(t *testing.T, content string) string {
	t.Helper()
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.yaml")
	err := os.WriteFile(tmpFile, []byte(content), 0o644)
	require.NoError(t, err)
	return tmpFile
}
