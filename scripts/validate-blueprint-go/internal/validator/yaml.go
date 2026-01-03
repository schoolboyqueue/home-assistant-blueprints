package validator

import (
	"os"
	"regexp"

	"gopkg.in/yaml.v3"

	errs "github.com/home-assistant-blueprints/validate-blueprint-go/internal/errors"
)

// LoadYAML loads and parses the YAML file
func (v *BlueprintValidator) LoadYAML() bool {
	content, err := os.ReadFile(v.FilePath)
	if err != nil {
		v.AddTypedError(errs.ErrFileReadError(v.FilePath, err))
		return false
	}

	// Custom handling for !input tags - replace with placeholder
	contentStr := string(content)
	inputRegex := regexp.MustCompile(`!input\s+(\S+)`)
	contentStr = inputRegex.ReplaceAllString(contentStr, `"!input $1"`)

	if err := yaml.Unmarshal([]byte(contentStr), &v.Data); err != nil {
		v.AddTypedError(errs.ErrYAMLSyntax(err))
		return false
	}

	return true
}
