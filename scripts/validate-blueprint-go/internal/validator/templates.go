package validator

import (
	"fmt"
	"regexp"
	"strings"
)

// ValidateTemplates validates Jinja2 templates throughout the blueprint
func (v *BlueprintValidator) ValidateTemplates() {
	v.validateTemplatesInValue(v.Data, "")
}

// validateTemplatesInValue recursively validates templates in a value
func (v *BlueprintValidator) validateTemplatesInValue(value interface{}, path string) {
	switch val := value.(type) {
	case string:
		v.validateTemplateString(val, path)
	case map[string]interface{}:
		for k, v2 := range val {
			newPath := path
			if newPath == "" {
				newPath = k
			} else {
				newPath = fmt.Sprintf("%s.%s", path, k)
			}
			v.validateTemplatesInValue(v2, newPath)
		}
	case []interface{}:
		for i, v2 := range val {
			v.validateTemplatesInValue(v2, fmt.Sprintf("%s[%d]", path, i))
		}
	}
}

// validateTemplateString validates a template string
func (v *BlueprintValidator) validateTemplateString(template, path string) {
	// Check for !input inside {{ }} blocks
	inputInTemplatePattern := regexp.MustCompile(`\{\{[^}]*!input[^}]*\}\}`)
	if inputInTemplatePattern.MatchString(template) {
		v.AddErrorf(
			"%s: Cannot use !input tags inside {{ }} blocks. Assign the input to a variable first.",
			path)
	}

	// Check for balanced Jinja2 delimiters
	if strings.Count(template, "{{") != strings.Count(template, "}}") {
		v.AddErrorf("%s: Unbalanced {{ }} delimiters", path)
	}
	if strings.Count(template, "{%") != strings.Count(template, "%}") {
		v.AddErrorf("%s: Unbalanced {%% %%} delimiters", path)
	}
}
