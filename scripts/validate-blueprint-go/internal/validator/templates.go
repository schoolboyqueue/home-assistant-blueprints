package validator

import (
	"github.com/home-assistant-blueprints/validate-blueprint-go/internal/common"
	errs "github.com/home-assistant-blueprints/validate-blueprint-go/internal/errors"
)

// ValidateTemplates validates Jinja2 templates throughout the blueprint
func (v *BlueprintValidator) ValidateTemplates() {
	v.validateTemplatesInValue(v.Data, "")
}

// validateTemplatesInValue recursively validates templates in a value
// Uses common.TraverseValue for consistent traversal and common path building.
func (v *BlueprintValidator) validateTemplatesInValue(value interface{}, path string) {
	common.TraverseValue(value, path, func(val interface{}, currentPath string) bool {
		if s, ok := val.(string); ok {
			v.validateTemplateString(s, currentPath)
		}
		return true
	})
}

// validateTemplateString validates a template string
// Uses common validation functions for consistency.
func (v *BlueprintValidator) validateTemplateString(template, path string) {
	// Check for !input inside {{ }} blocks
	if err := common.ValidateNoInputInTemplate(template, path); err != "" {
		v.AddTypedError(errs.ErrInvalidTemplate(path, err))
	}

	// Check for balanced Jinja2 delimiters
	for _, err := range common.ValidateBalancedDelimiters(template, path) {
		v.AddTypedError(errs.Create(errs.CodeUnclosedTemplateTag).WithPath(path).WithMessage(err))
	}
}
