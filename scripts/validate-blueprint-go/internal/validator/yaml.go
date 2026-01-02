package validator

import (
	"os"
	"regexp"

	"gopkg.in/yaml.v3"
)

// LoadYAML loads and parses the YAML file
func (v *BlueprintValidator) LoadYAML() bool {
	content, err := os.ReadFile(v.FilePath)
	if err != nil {
		v.AddErrorf("Failed to load file: %v", err)
		return false
	}

	// Custom handling for !input tags - replace with placeholder
	contentStr := string(content)
	inputRegex := regexp.MustCompile(`!input\s+(\S+)`)
	contentStr = inputRegex.ReplaceAllString(contentStr, `"!input $1"`)

	if err := yaml.Unmarshal([]byte(contentStr), &v.Data); err != nil {
		v.AddErrorf("YAML syntax error: %v", err)
		return false
	}

	return true
}
