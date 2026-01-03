package validator

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestErrorCategoryString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		category ErrorCategory
		expected string
	}{
		{CategorySyntax, "Syntax"},
		{CategorySchema, "Schema"},
		{CategoryReferences, "References"},
		{CategoryTemplates, "Templates"},
		{CategoryInputs, "Inputs"},
		{CategoryTriggers, "Triggers"},
		{CategoryConditions, "Conditions"},
		{CategoryActions, "Actions"},
		{CategoryDocumentation, "Documentation"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expected, tt.category.String())
		})
	}
}

func TestErrorCategoryDescription(t *testing.T) {
	t.Parallel()

	// All categories should have non-empty descriptions
	for _, cat := range AllCategories() {
		t.Run(cat.String(), func(t *testing.T) {
			t.Parallel()
			desc := cat.Description()
			assert.NotEmpty(t, desc)
			assert.NotEqual(t, "Unknown category", desc)
		})
	}
}

func TestAllCategories(t *testing.T) {
	t.Parallel()

	categories := AllCategories()
	assert.Len(t, categories, 9) // Should have 9 categories
	assert.Contains(t, categories, CategorySyntax)
	assert.Contains(t, categories, CategorySchema)
	assert.Contains(t, categories, CategoryReferences)
	assert.Contains(t, categories, CategoryTemplates)
	assert.Contains(t, categories, CategoryInputs)
	assert.Contains(t, categories, CategoryTriggers)
	assert.Contains(t, categories, CategoryConditions)
	assert.Contains(t, categories, CategoryActions)
	assert.Contains(t, categories, CategoryDocumentation)
}

func TestCategorizedErrorString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		err      CategorizedError
		expected string
	}{
		{
			name: "with path",
			err: CategorizedError{
				Category: CategorySyntax,
				Path:     "blueprint.input",
				Message:  "invalid syntax",
			},
			expected: "blueprint.input: invalid syntax",
		},
		{
			name: "without path",
			err: CategorizedError{
				Category: CategorySyntax,
				Path:     "",
				Message:  "file not found",
			},
			expected: "file not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expected, tt.err.String())
		})
	}
}

func TestCategorizedErrorFullString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		err      CategorizedError
		expected string
	}{
		{
			name: "with path",
			err: CategorizedError{
				Category: CategorySchema,
				Path:     "blueprint.domain",
				Message:  "invalid domain",
			},
			expected: "[Schema] blueprint.domain: invalid domain",
		},
		{
			name: "without path",
			err: CategorizedError{
				Category: CategorySyntax,
				Path:     "",
				Message:  "YAML syntax error",
			},
			expected: "[Syntax] YAML syntax error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expected, tt.err.FullString())
		})
	}
}

func TestCategorizedWarningString(t *testing.T) {
	t.Parallel()

	warn := CategorizedWarning{
		Category: CategoryDocumentation,
		Path:     "",
		Message:  "No README.md found",
	}
	assert.Equal(t, "No README.md found", warn.String())

	warnWithPath := CategorizedWarning{
		Category: CategoryInputs,
		Path:     "blueprint.input.foo",
		Message:  "No selector defined",
	}
	assert.Equal(t, "blueprint.input.foo: No selector defined", warnWithPath.String())
}

func TestErrorsByCategory(t *testing.T) {
	t.Parallel()

	errors := []CategorizedError{
		{Category: CategorySyntax, Message: "error1"},
		{Category: CategorySchema, Message: "error2"},
		{Category: CategorySyntax, Message: "error3"},
		{Category: CategoryTemplates, Message: "error4"},
	}

	grouped := ErrorsByCategory(errors)

	assert.Len(t, grouped[CategorySyntax], 2)
	assert.Len(t, grouped[CategorySchema], 1)
	assert.Len(t, grouped[CategoryTemplates], 1)
	assert.Len(t, grouped[CategoryInputs], 0)
}

func TestWarningsByCategory(t *testing.T) {
	t.Parallel()

	warnings := []CategorizedWarning{
		{Category: CategoryDocumentation, Message: "warn1"},
		{Category: CategoryDocumentation, Message: "warn2"},
		{Category: CategoryInputs, Message: "warn3"},
	}

	grouped := WarningsByCategory(warnings)

	assert.Len(t, grouped[CategoryDocumentation], 2)
	assert.Len(t, grouped[CategoryInputs], 1)
	assert.Len(t, grouped[CategorySyntax], 0)
}

func TestFilterErrorsByCategory(t *testing.T) {
	t.Parallel()

	errors := []CategorizedError{
		{Category: CategorySyntax, Message: "error1"},
		{Category: CategorySchema, Message: "error2"},
		{Category: CategoryTemplates, Message: "error3"},
		{Category: CategoryInputs, Message: "error4"},
	}

	t.Run("filter single category", func(t *testing.T) {
		t.Parallel()
		filtered := FilterErrorsByCategory(errors, CategorySyntax)
		assert.Len(t, filtered, 1)
		assert.Equal(t, "error1", filtered[0].Message)
	})

	t.Run("filter multiple categories", func(t *testing.T) {
		t.Parallel()
		filtered := FilterErrorsByCategory(errors, CategorySyntax, CategoryTemplates)
		assert.Len(t, filtered, 2)
	})

	t.Run("empty filter returns all", func(t *testing.T) {
		t.Parallel()
		filtered := FilterErrorsByCategory(errors)
		assert.Len(t, filtered, 4)
	})
}

func TestExcludeErrorsByCategory(t *testing.T) {
	t.Parallel()

	errors := []CategorizedError{
		{Category: CategorySyntax, Message: "error1"},
		{Category: CategorySchema, Message: "error2"},
		{Category: CategoryTemplates, Message: "error3"},
	}

	t.Run("exclude single category", func(t *testing.T) {
		t.Parallel()
		excluded := ExcludeErrorsByCategory(errors, CategorySyntax)
		assert.Len(t, excluded, 2)
		for _, err := range excluded {
			assert.NotEqual(t, CategorySyntax, err.Category)
		}
	})

	t.Run("empty exclusion returns all", func(t *testing.T) {
		t.Parallel()
		excluded := ExcludeErrorsByCategory(errors)
		assert.Len(t, excluded, 3)
	})
}

func TestCountByCategory(t *testing.T) {
	t.Parallel()

	errors := []CategorizedError{
		{Category: CategorySyntax, Message: "error1"},
		{Category: CategorySyntax, Message: "error2"},
		{Category: CategorySchema, Message: "error3"},
	}

	counts := CountByCategory(errors)

	assert.Equal(t, 2, counts[CategorySyntax])
	assert.Equal(t, 1, counts[CategorySchema])
	assert.Equal(t, 0, counts[CategoryTemplates])
}

func TestFormatCategorySummary(t *testing.T) {
	t.Parallel()

	t.Run("with errors and warnings grouped", func(t *testing.T) {
		t.Parallel()
		errors := []CategorizedError{
			{Category: CategorySyntax, Path: "file", Message: "syntax error"},
			{Category: CategorySchema, Path: "", Message: "missing key"},
		}
		warnings := []CategorizedWarning{
			{Category: CategoryDocumentation, Path: "", Message: "no readme"},
		}

		output := FormatCategorySummary(errors, warnings, true)

		assert.Contains(t, output, "Syntax")
		assert.Contains(t, output, "Schema")
		assert.Contains(t, output, "Documentation")
		assert.Contains(t, output, "syntax error")
		assert.Contains(t, output, "missing key")
		assert.Contains(t, output, "no readme")
	})

	t.Run("flat output", func(t *testing.T) {
		t.Parallel()
		errors := []CategorizedError{
			{Category: CategorySyntax, Message: "error1"},
		}

		output := FormatCategorySummary(errors, nil, false)

		assert.Contains(t, output, "ERRORS")
		assert.Contains(t, output, "error1")
		// Should not have category labels in flat mode
		assert.NotContains(t, output, "Syntax ERRORS")
	})

	t.Run("empty lists", func(t *testing.T) {
		t.Parallel()
		output := FormatCategorySummary(nil, nil, true)
		assert.Empty(t, output)
	})
}

func TestSortedCategoryKeys(t *testing.T) {
	t.Parallel()

	grouped := map[ErrorCategory][]CategorizedError{
		CategoryTemplates: {{Message: "t1"}},
		CategorySyntax:    {{Message: "s1"}},
		CategorySchema:    {{Message: "sc1"}},
	}

	sorted := SortedCategoryKeys(grouped)

	// Should be sorted by enum order (CategorySyntax < CategorySchema < CategoryTemplates)
	assert.Len(t, sorted, 3)
	assert.Equal(t, CategorySyntax, sorted[0])
	assert.Equal(t, CategorySchema, sorted[1])
	assert.Equal(t, CategoryTemplates, sorted[2])
}

func TestValidatorCategorizedMethods(t *testing.T) {
	t.Parallel()

	t.Run("AddCategorizedError", func(t *testing.T) {
		t.Parallel()
		v := New("test.yaml")

		v.AddCategorizedError(CategorySyntax, "path.to.error", "test message")

		require.Len(t, v.CategorizedErrors, 1)
		assert.Equal(t, CategorySyntax, v.CategorizedErrors[0].Category)
		assert.Equal(t, "path.to.error", v.CategorizedErrors[0].Path)
		assert.Equal(t, "test message", v.CategorizedErrors[0].Message)

		// Should also add to legacy Errors slice
		require.Len(t, v.Errors, 1)
		assert.Equal(t, "path.to.error: test message", v.Errors[0])
	})

	t.Run("AddCategorizedErrorf", func(t *testing.T) {
		t.Parallel()
		v := New("test.yaml")

		v.AddCategorizedErrorf(CategorySchema, "blueprint", "missing key: %s", "name")

		require.Len(t, v.CategorizedErrors, 1)
		assert.Equal(t, "missing key: name", v.CategorizedErrors[0].Message)
	})

	t.Run("AddCategorizedWarning", func(t *testing.T) {
		t.Parallel()
		v := New("test.yaml")

		v.AddCategorizedWarning(CategoryDocumentation, "", "no readme found")

		require.Len(t, v.CategorizedWarnings, 1)
		assert.Equal(t, CategoryDocumentation, v.CategorizedWarnings[0].Category)

		// Should also add to legacy Warnings slice
		require.Len(t, v.Warnings, 1)
		assert.Equal(t, "no readme found", v.Warnings[0])
	})

	t.Run("AddCategorizedWarningf", func(t *testing.T) {
		t.Parallel()
		v := New("test.yaml")

		v.AddCategorizedWarningf(CategoryInputs, "input.foo", "selector %s not found", "entity")

		require.Len(t, v.CategorizedWarnings, 1)
		assert.Contains(t, v.CategorizedWarnings[0].Message, "selector entity not found")
	})
}

func TestValidatorGetErrorsByCategory(t *testing.T) {
	t.Parallel()

	v := New("test.yaml")
	v.AddCategorizedError(CategorySyntax, "", "syntax error")
	v.AddCategorizedError(CategorySchema, "", "schema error")
	v.AddCategorizedError(CategorySyntax, "", "another syntax error")

	syntaxErrors := v.GetErrorsByCategory(CategorySyntax)
	assert.Len(t, syntaxErrors, 2)

	schemaErrors := v.GetErrorsByCategory(CategorySchema)
	assert.Len(t, schemaErrors, 1)

	multiErrors := v.GetErrorsByCategory(CategorySyntax, CategorySchema)
	assert.Len(t, multiErrors, 3)
}

func TestValidatorErrorCounts(t *testing.T) {
	t.Parallel()

	v := New("test.yaml")
	v.AddCategorizedError(CategorySyntax, "", "error1")
	v.AddCategorizedError(CategorySyntax, "", "error2")
	v.AddCategorizedError(CategorySchema, "", "error3")

	counts := v.ErrorCounts()

	assert.Equal(t, 2, counts[CategorySyntax])
	assert.Equal(t, 1, counts[CategorySchema])
}

func TestValidatorHasCategoryErrors(t *testing.T) {
	t.Parallel()

	v := New("test.yaml")
	v.AddCategorizedError(CategorySyntax, "", "syntax error")

	assert.True(t, v.HasCategoryErrors(CategorySyntax))
	assert.False(t, v.HasCategoryErrors(CategorySchema))
	assert.True(t, v.HasCategoryErrors(CategorySyntax, CategorySchema))
}

func TestValidatorHasCategoryWarnings(t *testing.T) {
	t.Parallel()

	v := New("test.yaml")
	v.AddCategorizedWarning(CategoryDocumentation, "", "no readme")

	assert.True(t, v.HasCategoryWarnings(CategoryDocumentation))
	assert.False(t, v.HasCategoryWarnings(CategoryInputs))
}

func TestValidatorGroupByCategory(t *testing.T) {
	t.Parallel()

	t.Run("default is true", func(t *testing.T) {
		t.Parallel()
		v := New("test.yaml")
		assert.True(t, v.GroupByCategory)
	})
}

func TestFilterWarningsByCategory(t *testing.T) {
	t.Parallel()

	warnings := []CategorizedWarning{
		{Category: CategoryInputs, Message: "warn1"},
		{Category: CategoryDocumentation, Message: "warn2"},
		{Category: CategoryInputs, Message: "warn3"},
	}

	filtered := FilterWarningsByCategory(warnings, CategoryInputs)
	assert.Len(t, filtered, 2)
	for _, w := range filtered {
		assert.Equal(t, CategoryInputs, w.Category)
	}
}

func TestExcludeWarningsByCategory(t *testing.T) {
	t.Parallel()

	warnings := []CategorizedWarning{
		{Category: CategoryInputs, Message: "warn1"},
		{Category: CategoryDocumentation, Message: "warn2"},
	}

	excluded := ExcludeWarningsByCategory(warnings, CategoryInputs)
	assert.Len(t, excluded, 1)
	assert.Equal(t, CategoryDocumentation, excluded[0].Category)
}

func TestCountWarningsByCategory(t *testing.T) {
	t.Parallel()

	warnings := []CategorizedWarning{
		{Category: CategoryInputs, Message: "warn1"},
		{Category: CategoryInputs, Message: "warn2"},
		{Category: CategoryDocumentation, Message: "warn3"},
	}

	counts := CountWarningsByCategory(warnings)

	assert.Equal(t, 2, counts[CategoryInputs])
	assert.Equal(t, 1, counts[CategoryDocumentation])
}

func TestCategorizedWarningFullString(t *testing.T) {
	t.Parallel()

	warn := CategorizedWarning{
		Category: CategoryInputs,
		Path:     "input.test",
		Message:  "test warning",
	}
	assert.Equal(t, "[Inputs] input.test: test warning", warn.FullString())

	warnNoPath := CategorizedWarning{
		Category: CategoryDocumentation,
		Path:     "",
		Message:  "no path warning",
	}
	assert.Equal(t, "[Documentation] no path warning", warnNoPath.FullString())
}

func TestIntegrationCategorizedValidation(t *testing.T) {
	t.Parallel()

	t.Run("validation errors are categorized correctly", func(t *testing.T) {
		t.Parallel()
		v := New("test.yaml")
		v.Data = map[string]interface{}{
			// Missing required keys - should produce CategorySchema errors
		}

		// Simulate validations
		v.ValidateStructure()

		// Should have schema errors for missing required root keys
		schemaErrors := v.GetErrorsByCategory(CategorySchema)
		assert.NotEmpty(t, schemaErrors)
	})
}

func TestFormatCategorySummaryOrder(t *testing.T) {
	t.Parallel()

	// Create errors in reverse category order
	errors := []CategorizedError{
		{Category: CategoryDocumentation, Message: "doc error"},
		{Category: CategoryActions, Message: "action error"},
		{Category: CategorySyntax, Message: "syntax error"},
	}

	output := FormatCategorySummary(errors, nil, true)

	// Verify categories appear in the correct order (Syntax should come before Actions and Documentation)
	syntaxIdx := strings.Index(output, "Syntax")
	actionsIdx := strings.Index(output, "Actions")
	docIdx := strings.Index(output, "Documentation")

	assert.True(t, syntaxIdx < actionsIdx, "Syntax should appear before Actions")
	assert.True(t, actionsIdx < docIdx, "Actions should appear before Documentation")
}
