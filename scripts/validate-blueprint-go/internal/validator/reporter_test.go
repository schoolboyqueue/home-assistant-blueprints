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
