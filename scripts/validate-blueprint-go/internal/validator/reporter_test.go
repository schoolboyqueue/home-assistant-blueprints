package validator

import (
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Note: TestReportResults cannot be parallel because it modifies os.Stdout
func TestReportResults(t *testing.T) {
	// These tests are sequential because they modify os.Stdout
	t.Run("no errors or warnings", func(t *testing.T) {
		v := New("test.yaml")

		// Capture stdout
		oldStdout := os.Stdout
		r, w, err := os.Pipe()
		require.NoError(t, err)
		os.Stdout = w

		result := v.ReportResults()

		w.Close()
		os.Stdout = oldStdout
		out, err := io.ReadAll(r)
		require.NoError(t, err)

		assert.True(t, result)
		assert.Contains(t, string(out), "valid")
	})

	t.Run("only warnings", func(t *testing.T) {
		v := New("test.yaml")
		v.AddWarning("test warning 1")
		v.AddWarning("test warning 2")

		// Capture stdout
		oldStdout := os.Stdout
		r, w, err := os.Pipe()
		require.NoError(t, err)
		os.Stdout = w

		result := v.ReportResults()

		w.Close()
		os.Stdout = oldStdout
		out, err := io.ReadAll(r)
		require.NoError(t, err)

		assert.True(t, result)
		assert.Contains(t, string(out), "WARNINGS")
		assert.Contains(t, string(out), "2 warnings")
	})

	t.Run("only errors", func(t *testing.T) {
		v := New("test.yaml")
		v.AddError("test error 1")
		v.AddError("test error 2")

		// Capture stdout
		oldStdout := os.Stdout
		r, w, err := os.Pipe()
		require.NoError(t, err)
		os.Stdout = w

		result := v.ReportResults()

		w.Close()
		os.Stdout = oldStdout
		out, err := io.ReadAll(r)
		require.NoError(t, err)

		assert.False(t, result)
		assert.Contains(t, string(out), "ERRORS")
		assert.Contains(t, string(out), "FAIL")
	})

	t.Run("errors and warnings", func(t *testing.T) {
		v := New("test.yaml")
		v.AddError("test error")
		v.AddWarning("test warning")

		// Capture stdout
		oldStdout := os.Stdout
		r, w, err := os.Pipe()
		require.NoError(t, err)
		os.Stdout = w

		result := v.ReportResults()

		w.Close()
		os.Stdout = oldStdout
		out, err := io.ReadAll(r)
		require.NoError(t, err)

		assert.False(t, result)
		assert.Contains(t, string(out), "ERRORS")
		assert.Contains(t, string(out), "WARNINGS")
	})
}

func TestCheckReadmeExists(t *testing.T) {
	t.Parallel()

	t.Run("README.md exists", func(t *testing.T) {
		t.Parallel()
		tmpDir := t.TempDir()
		blueprintPath := filepath.Join(tmpDir, "blueprint.yaml")
		readmePath := filepath.Join(tmpDir, "README.md")

		// Create files
		err := os.WriteFile(blueprintPath, []byte("test"), 0o644)
		require.NoError(t, err)
		err = os.WriteFile(readmePath, []byte("# Test"), 0o644)
		require.NoError(t, err)

		v := New(blueprintPath)
		v.CheckReadmeExists()

		assert.Empty(t, v.Warnings)
	})

	t.Run("README.md does not exist", func(t *testing.T) {
		t.Parallel()
		tmpDir := t.TempDir()
		blueprintPath := filepath.Join(tmpDir, "blueprint.yaml")

		// Create only blueprint file
		err := os.WriteFile(blueprintPath, []byte("test"), 0o644)
		require.NoError(t, err)

		v := New(blueprintPath)
		v.CheckReadmeExists()

		require.Len(t, v.Warnings, 1)
		assert.Contains(t, v.Warnings[0], "README.md")
	})
}

func TestCheckChangelogExists(t *testing.T) {
	t.Parallel()

	t.Run("CHANGELOG.md exists", func(t *testing.T) {
		t.Parallel()
		tmpDir := t.TempDir()
		blueprintPath := filepath.Join(tmpDir, "blueprint.yaml")
		changelogPath := filepath.Join(tmpDir, "CHANGELOG.md")

		// Create files
		err := os.WriteFile(blueprintPath, []byte("test"), 0o644)
		require.NoError(t, err)
		err = os.WriteFile(changelogPath, []byte("# Changelog"), 0o644)
		require.NoError(t, err)

		v := New(blueprintPath)
		v.CheckChangelogExists()

		assert.Empty(t, v.Warnings)
	})

	t.Run("CHANGELOG.md does not exist", func(t *testing.T) {
		t.Parallel()
		tmpDir := t.TempDir()
		blueprintPath := filepath.Join(tmpDir, "blueprint.yaml")

		// Create only blueprint file
		err := os.WriteFile(blueprintPath, []byte("test"), 0o644)
		require.NoError(t, err)

		v := New(blueprintPath)
		v.CheckChangelogExists()

		require.Len(t, v.Warnings, 1)
		assert.Contains(t, v.Warnings[0], "CHANGELOG.md")
	})
}

func TestReportResultsCategorized(t *testing.T) {
	// Note: These tests are sequential because they modify os.Stdout
	t.Run("categorized errors output", func(t *testing.T) {
		v := New("test.yaml")
		v.AddCategorizedError(CategorySyntax, "file", "syntax error")
		v.AddCategorizedError(CategorySchema, "blueprint", "schema error")

		// Capture stdout
		oldStdout := os.Stdout
		r, w, err := os.Pipe()
		require.NoError(t, err)
		os.Stdout = w

		result := v.ReportResults()

		w.Close()
		os.Stdout = oldStdout
		out, err := io.ReadAll(r)
		require.NoError(t, err)

		assert.False(t, result)
		assert.Contains(t, string(out), "Syntax")
		assert.Contains(t, string(out), "Schema")
		assert.Contains(t, string(out), "syntax error")
		assert.Contains(t, string(out), "schema error")
	})

	t.Run("categorized warnings output", func(t *testing.T) {
		v := New("test.yaml")
		v.AddCategorizedWarning(CategoryDocumentation, "", "no readme")
		v.AddCategorizedWarning(CategoryInputs, "input.foo", "no selector")

		// Capture stdout
		oldStdout := os.Stdout
		r, w, err := os.Pipe()
		require.NoError(t, err)
		os.Stdout = w

		result := v.ReportResults()

		w.Close()
		os.Stdout = oldStdout
		out, err := io.ReadAll(r)
		require.NoError(t, err)

		assert.True(t, result) // Warnings don't fail
		assert.Contains(t, string(out), "Documentation")
		assert.Contains(t, string(out), "Inputs")
	})

	t.Run("flat output when GroupByCategory is false", func(t *testing.T) {
		v := New("test.yaml")
		v.GroupByCategory = false
		v.AddCategorizedError(CategorySyntax, "", "test error")

		// Capture stdout
		oldStdout := os.Stdout
		r, w, err := os.Pipe()
		require.NoError(t, err)
		os.Stdout = w

		result := v.ReportResults()

		w.Close()
		os.Stdout = oldStdout
		out, err := io.ReadAll(r)
		require.NoError(t, err)

		assert.False(t, result)
		// Should show flat output without category labels
		assert.Contains(t, string(out), "ERRORS")
	})
}

func TestReportResultsFlat(t *testing.T) {
	t.Run("flat output method", func(t *testing.T) {
		v := New("test.yaml")
		v.AddCategorizedError(CategorySyntax, "", "error1")
		v.AddCategorizedWarning(CategoryDocumentation, "", "warning1")

		// Capture stdout
		oldStdout := os.Stdout
		r, w, err := os.Pipe()
		require.NoError(t, err)
		os.Stdout = w

		result := v.ReportResultsFlat()

		w.Close()
		os.Stdout = oldStdout
		out, err := io.ReadAll(r)
		require.NoError(t, err)

		assert.False(t, result)
		assert.Contains(t, string(out), "ERRORS")
		assert.Contains(t, string(out), "WARNINGS")
	})
}

func TestReportFilteredResults(t *testing.T) {
	t.Run("filter by single category", func(t *testing.T) {
		v := New("test.yaml")
		v.AddCategorizedError(CategorySyntax, "", "syntax error")
		v.AddCategorizedError(CategorySchema, "", "schema error")

		// Capture stdout
		oldStdout := os.Stdout
		r, w, err := os.Pipe()
		require.NoError(t, err)
		os.Stdout = w

		result := v.ReportFilteredResults(CategorySyntax)

		w.Close()
		os.Stdout = oldStdout
		out, err := io.ReadAll(r)
		require.NoError(t, err)

		assert.False(t, result)
		assert.Contains(t, string(out), "syntax error")
		// Schema error should not be in output since we filtered by Syntax only
	})

	t.Run("no errors in selected category", func(t *testing.T) {
		v := New("test.yaml")
		v.AddCategorizedError(CategorySyntax, "", "syntax error")

		// Capture stdout
		oldStdout := os.Stdout
		r, w, err := os.Pipe()
		require.NoError(t, err)
		os.Stdout = w

		result := v.ReportFilteredResults(CategorySchema)

		w.Close()
		os.Stdout = oldStdout
		out, err := io.ReadAll(r)
		require.NoError(t, err)

		assert.True(t, result) // No errors in Schema category
		assert.Contains(t, string(out), "No issues found")
	})
}

func TestDocumentationWarningsAreCategorized(t *testing.T) {
	t.Parallel()

	t.Run("README warning has correct category", func(t *testing.T) {
		t.Parallel()
		tmpDir := t.TempDir()
		blueprintPath := filepath.Join(tmpDir, "blueprint.yaml")

		err := os.WriteFile(blueprintPath, []byte("test"), 0o644)
		require.NoError(t, err)

		v := New(blueprintPath)
		v.CheckReadmeExists()

		require.Len(t, v.CategorizedWarnings, 1)
		assert.Equal(t, CategoryDocumentation, v.CategorizedWarnings[0].Category)
	})

	t.Run("CHANGELOG warning has correct category", func(t *testing.T) {
		t.Parallel()
		tmpDir := t.TempDir()
		blueprintPath := filepath.Join(tmpDir, "blueprint.yaml")

		err := os.WriteFile(blueprintPath, []byte("test"), 0o644)
		require.NoError(t, err)

		v := New(blueprintPath)
		v.CheckChangelogExists()

		require.Len(t, v.CategorizedWarnings, 1)
		assert.Equal(t, CategoryDocumentation, v.CategorizedWarnings[0].Category)
	})
}
