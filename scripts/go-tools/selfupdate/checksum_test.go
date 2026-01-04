package selfupdate

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestParseChecksums(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    Checksums
		wantErr bool
	}{
		{
			name: "standard format with two spaces",
			input: `e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855  ha-ws-client-linux-amd64
a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2  ha-ws-client-darwin-arm64`,
			want: Checksums{
				"ha-ws-client-linux-amd64":  "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
				"ha-ws-client-darwin-arm64": "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2",
			},
			wantErr: false,
		},
		{
			name:  "single space format",
			input: `e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855 ha-ws-client-linux-amd64`,
			want: Checksums{
				"ha-ws-client-linux-amd64": "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
			},
			wantErr: false,
		},
		{
			name: "with empty lines",
			input: `e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855  file1

a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2  file2`,
			want: Checksums{
				"file1": "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
				"file2": "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2",
			},
			wantErr: false,
		},
		{
			name:  "uppercase hash",
			input: `E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855  file1`,
			want: Checksums{
				"file1": "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
			},
			wantErr: false,
		},
		{
			name:    "empty file",
			input:   "",
			wantErr: true,
		},
		{
			name:    "invalid format - no space",
			input:   "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855filename",
			wantErr: true,
		},
		{
			name:    "invalid hash length",
			input:   "e3b0c44298fc1c14  filename",
			wantErr: true,
		},
		{
			name:    "invalid hex characters",
			input:   "z3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855  filename",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseChecksums(strings.NewReader(tt.input))
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseChecksums() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}

			if len(got) != len(tt.want) {
				t.Errorf("ParseChecksums() returned %d entries, want %d", len(got), len(tt.want))
			}

			for k, v := range tt.want {
				if got[k] != v {
					t.Errorf("ParseChecksums()[%q] = %q, want %q", k, got[k], v)
				}
			}
		})
	}
}

func TestChecksums_GetChecksum(t *testing.T) {
	checksums := Checksums{
		"file1": "hash1",
		"file2": "hash2",
	}

	tests := []struct {
		filename string
		want     string
	}{
		{"file1", "hash1"},
		{"file2", "hash2"},
		{"file3", ""},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			got := checksums.GetChecksum(tt.filename)
			if got != tt.want {
				t.Errorf("GetChecksum(%q) = %q, want %q", tt.filename, got, tt.want)
			}
		})
	}
}

func TestVerifyChecksum(t *testing.T) {
	// Create a temp file with known content
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "testfile")
	content := []byte("hello world")
	if err := os.WriteFile(testFile, content, 0o644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Compute expected hash
	h := sha256.Sum256(content)
	expectedHash := hex.EncodeToString(h[:])

	t.Run("matching checksum", func(t *testing.T) {
		err := VerifyChecksum(testFile, expectedHash)
		if err != nil {
			t.Errorf("VerifyChecksum() error = %v, want nil", err)
		}
	})

	t.Run("case insensitive matching", func(t *testing.T) {
		err := VerifyChecksum(testFile, strings.ToUpper(expectedHash))
		if err != nil {
			t.Errorf("VerifyChecksum() error = %v, want nil", err)
		}
	})

	t.Run("mismatched checksum", func(t *testing.T) {
		wrongHash := "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
		err := VerifyChecksum(testFile, wrongHash)
		if err == nil {
			t.Error("VerifyChecksum() expected error for mismatched checksum")
		}

		if !errors.Is(err, ErrChecksumMismatch) {
			t.Errorf("error should be ErrChecksumMismatch, got %v", err)
		}

		var checksumErr *ChecksumError
		if errors.As(err, &checksumErr) {
			if checksumErr.Expected != wrongHash {
				t.Errorf("ChecksumError.Expected = %q, want %q", checksumErr.Expected, wrongHash)
			}
			if checksumErr.Actual != expectedHash {
				t.Errorf("ChecksumError.Actual = %q, want %q", checksumErr.Actual, expectedHash)
			}
		} else {
			t.Error("error should be *ChecksumError")
		}
	})

	t.Run("nonexistent file", func(t *testing.T) {
		err := VerifyChecksum("/nonexistent/file", expectedHash)
		if err == nil {
			t.Error("VerifyChecksum() expected error for nonexistent file")
		}
	})
}

func TestComputeChecksum(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "testfile")
	content := []byte("test content")
	if err := os.WriteFile(testFile, content, 0o644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	h := sha256.Sum256(content)
	expected := hex.EncodeToString(h[:])

	got, err := ComputeChecksum(testFile)
	if err != nil {
		t.Fatalf("ComputeChecksum() error = %v", err)
	}

	if got != expected {
		t.Errorf("ComputeChecksum() = %q, want %q", got, expected)
	}
}

func TestComputeReaderChecksum(t *testing.T) {
	content := "test content"
	h := sha256.Sum256([]byte(content))
	expected := hex.EncodeToString(h[:])

	got, err := ComputeReaderChecksum(strings.NewReader(content))
	if err != nil {
		t.Fatalf("ComputeReaderChecksum() error = %v", err)
	}

	if got != expected {
		t.Errorf("ComputeReaderChecksum() = %q, want %q", got, expected)
	}
}

func TestVerifyReaderChecksum(t *testing.T) {
	content := "test content"
	h := sha256.Sum256([]byte(content))
	correctHash := hex.EncodeToString(h[:])

	t.Run("matching checksum", func(t *testing.T) {
		err := VerifyReaderChecksum(strings.NewReader(content), correctHash, "testfile")
		if err != nil {
			t.Errorf("VerifyReaderChecksum() error = %v", err)
		}
	})

	t.Run("mismatched checksum", func(t *testing.T) {
		wrongHash := "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
		err := VerifyReaderChecksum(strings.NewReader(content), wrongHash, "testfile")
		if err == nil {
			t.Error("VerifyReaderChecksum() expected error")
		}
		if !errors.Is(err, ErrChecksumMismatch) {
			t.Errorf("error should be ErrChecksumMismatch, got %v", err)
		}
	})
}

func TestDownloadChecksums(t *testing.T) {
	checksumContent := `e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855  ha-ws-client-linux-amd64
a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2  ha-ws-client-darwin-arm64`

	t.Run("successful download", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte(checksumContent))
		}))
		defer server.Close()

		checksums, err := DownloadChecksums(server.URL, 30*time.Second)
		if err != nil {
			t.Fatalf("DownloadChecksums() error = %v", err)
		}

		if len(checksums) != 2 {
			t.Errorf("DownloadChecksums() returned %d entries, want 2", len(checksums))
		}

		if checksums["ha-ws-client-linux-amd64"] != "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855" {
			t.Error("DownloadChecksums() missing or incorrect checksum for ha-ws-client-linux-amd64")
		}
	})

	t.Run("404 error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer server.Close()

		_, err := DownloadChecksums(server.URL, 30*time.Second)
		if err == nil {
			t.Error("DownloadChecksums() expected error for 404")
		}

		var downloadErr *DownloadError
		if !errors.As(err, &downloadErr) {
			t.Errorf("error should be *DownloadError, got %T", err)
		}
	})

	t.Run("invalid checksum content", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte("invalid content"))
		}))
		defer server.Close()

		_, err := DownloadChecksums(server.URL, 30*time.Second)
		if err == nil {
			t.Error("DownloadChecksums() expected error for invalid content")
		}
	})
}
