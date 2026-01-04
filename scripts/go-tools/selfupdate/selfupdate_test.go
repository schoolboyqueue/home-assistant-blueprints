package selfupdate

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewUpdater(t *testing.T) {
	u, err := NewUpdater("ha-ws-client", "ha-ws-client-go", "1.5.4")
	if err != nil {
		t.Fatalf("NewUpdater() error = %v", err)
	}

	if u.ToolName != "ha-ws-client" {
		t.Errorf("ToolName = %q, want %q", u.ToolName, "ha-ws-client")
	}
	if u.ToolTag != "ha-ws-client-go" {
		t.Errorf("ToolTag = %q, want %q", u.ToolTag, "ha-ws-client-go")
	}
	if u.CurrentVersion != "1.5.4" {
		t.Errorf("CurrentVersion = %q, want %q", u.CurrentVersion, "1.5.4")
	}
	if u.downloadTimeout != DefaultDownloadTimeout {
		t.Errorf("downloadTimeout = %v, want %v", u.downloadTimeout, DefaultDownloadTimeout)
	}
}

func TestNewUpdater_WithOptions(t *testing.T) {
	var buf bytes.Buffer
	customTimeout := 60 * time.Second

	u, err := NewUpdater("test", "test-go", "1.0.0",
		WithOutput(&buf),
		WithDownloadTimeout(customTimeout),
		WithQuietMode(),
	)
	if err != nil {
		t.Fatalf("NewUpdater() error = %v", err)
	}

	if u.output != &buf {
		t.Error("output was not set correctly")
	}
	if u.downloadTimeout != customTimeout {
		t.Errorf("downloadTimeout = %v, want %v", u.downloadTimeout, customTimeout)
	}
	if !u.quiet {
		t.Error("quiet mode was not set")
	}
}

func TestUpdater_Check(t *testing.T) {
	releases := []Release{
		{
			TagName: "ha-ws-client-go/v1.6.0",
			Assets: []Asset{
				{Name: "ha-ws-client-linux-amd64", BrowserDownloadURL: "https://example.com/ha-ws-client-linux-amd64", Size: 1024},
				{Name: "ha-ws-client-darwin-arm64", BrowserDownloadURL: "https://example.com/ha-ws-client-darwin-arm64", Size: 1024},
				{Name: "checksums.txt", BrowserDownloadURL: "https://example.com/checksums.txt", Size: 256},
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(releases)
	}))
	defer server.Close()

	client := NewGitHubClient(
		WithBaseURL(server.URL),
		WithRepository("test", "test"),
	)

	u, err := NewUpdater("ha-ws-client", "ha-ws-client-go", "1.5.4", WithGitHubClient(client))
	if err != nil {
		t.Fatalf("NewUpdater() error = %v", err)
	}

	result, err := u.Check()
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}

	if result.CurrentVersion != "1.5.4" {
		t.Errorf("CurrentVersion = %q, want %q", result.CurrentVersion, "1.5.4")
	}
	if result.LatestVersion != "1.6.0" {
		t.Errorf("LatestVersion = %q, want %q", result.LatestVersion, "1.6.0")
	}
	if !result.UpdateAvailable {
		t.Error("UpdateAvailable should be true")
	}
	if result.ChecksumURL != "https://example.com/checksums.txt" {
		t.Errorf("ChecksumURL = %q, want %q", result.ChecksumURL, "https://example.com/checksums.txt")
	}
}

func TestUpdater_Check_AlreadyLatest(t *testing.T) {
	releases := []Release{
		{
			TagName: "ha-ws-client-go/v1.5.4",
			Assets: []Asset{
				{Name: "ha-ws-client-linux-amd64", BrowserDownloadURL: "https://example.com/binary"},
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(releases)
	}))
	defer server.Close()

	client := NewGitHubClient(
		WithBaseURL(server.URL),
		WithRepository("test", "test"),
	)

	u, err := NewUpdater("ha-ws-client", "ha-ws-client-go", "1.5.4", WithGitHubClient(client))
	if err != nil {
		t.Fatalf("NewUpdater() error = %v", err)
	}

	result, err := u.Check()
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}

	if result.UpdateAvailable {
		t.Error("UpdateAvailable should be false when already at latest")
	}
}

func TestUpdater_Check_DevVersion(t *testing.T) {
	releases := []Release{
		{
			TagName: "ha-ws-client-go/v1.5.4",
			Assets: []Asset{
				{Name: "ha-ws-client-darwin-arm64", BrowserDownloadURL: "https://example.com/binary"},
				{Name: "checksums.txt", BrowserDownloadURL: "https://example.com/checksums.txt"},
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(releases)
	}))
	defer server.Close()

	client := NewGitHubClient(
		WithBaseURL(server.URL),
		WithRepository("test", "test"),
	)

	// 'dev' version should always show update available
	u, err := NewUpdater("ha-ws-client", "ha-ws-client-go", "dev", WithGitHubClient(client))
	if err != nil {
		t.Fatalf("NewUpdater() error = %v", err)
	}

	result, err := u.Check()
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}

	if !result.UpdateAvailable {
		t.Error("UpdateAvailable should be true for 'dev' version")
	}
}

func TestUpdater_needsUpdate(t *testing.T) {
	u := &Updater{CurrentVersion: "1.5.4"}

	tests := []struct {
		currentVersion string
		targetVersion  string
		want           bool
	}{
		{"1.5.4", "1.6.0", true},
		{"1.5.4", "1.5.4", false},
		{"dev", "1.5.4", true},
		{"dev", "1.0.0", true},
	}

	for _, tt := range tests {
		t.Run(tt.currentVersion+"_"+tt.targetVersion, func(t *testing.T) {
			u.CurrentVersion = tt.currentVersion
			got := u.needsUpdate(tt.targetVersion)
			if got != tt.want {
				t.Errorf("needsUpdate(%q) = %v, want %v", tt.targetVersion, got, tt.want)
			}
		})
	}
}

func TestUpdater_ListAvailableVersions(t *testing.T) {
	releases := []Release{
		{TagName: "ha-ws-client-go/v1.6.0"},
		{TagName: "ha-ws-client-go/v1.5.4"},
		{TagName: "ha-ws-client-go/v1.5.3"},
		{TagName: "validate-blueprint-go/v1.0.0"}, // Different tool
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(releases)
	}))
	defer server.Close()

	client := NewGitHubClient(
		WithBaseURL(server.URL),
		WithRepository("test", "test"),
	)

	u, err := NewUpdater("ha-ws-client", "ha-ws-client-go", "1.5.4", WithGitHubClient(client))
	if err != nil {
		t.Fatalf("NewUpdater() error = %v", err)
	}

	versions, err := u.ListAvailableVersions()
	if err != nil {
		t.Fatalf("ListAvailableVersions() error = %v", err)
	}

	expected := []string{"1.6.0", "1.5.4", "1.5.3"}
	if len(versions) != len(expected) {
		t.Errorf("ListAvailableVersions() returned %d versions, want %d", len(versions), len(expected))
	}

	for i, v := range expected {
		if versions[i] != v {
			t.Errorf("versions[%d] = %q, want %q", i, versions[i], v)
		}
	}
}

func TestUpdater_Update_AlreadyLatest(t *testing.T) {
	releases := []Release{
		{
			TagName: "ha-ws-client-go/v1.5.4",
			Assets: []Asset{
				{Name: "ha-ws-client-darwin-arm64", BrowserDownloadURL: "https://example.com/binary"},
				{Name: "checksums.txt", BrowserDownloadURL: "https://example.com/checksums.txt"},
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(releases)
	}))
	defer server.Close()

	client := NewGitHubClient(
		WithBaseURL(server.URL),
		WithRepository("test", "test"),
	)

	u, err := NewUpdater("ha-ws-client", "ha-ws-client-go", "1.5.4", WithGitHubClient(client))
	if err != nil {
		t.Fatalf("NewUpdater() error = %v", err)
	}

	err = u.Update()
	if !errors.Is(err, ErrAlreadyLatest) {
		t.Errorf("Update() error = %v, want ErrAlreadyLatest", err)
	}
}

func TestUpdater_Update_MissingChecksum(t *testing.T) {
	releases := []Release{
		{
			TagName: "ha-ws-client-go/v1.6.0",
			Assets: []Asset{
				{Name: "ha-ws-client-darwin-arm64", BrowserDownloadURL: "https://example.com/binary"},
				// No checksums.txt
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(releases)
	}))
	defer server.Close()

	client := NewGitHubClient(
		WithBaseURL(server.URL),
		WithRepository("test", "test"),
	)

	u, err := NewUpdater("ha-ws-client", "ha-ws-client-go", "1.5.4", WithGitHubClient(client))
	if err != nil {
		t.Fatalf("NewUpdater() error = %v", err)
	}

	err = u.Update()
	if !errors.Is(err, ErrMissingChecksum) {
		t.Errorf("Update() error = %v, want ErrMissingChecksum", err)
	}
}

func TestUpdater_UpdateWithVerification(t *testing.T) {
	// Create test binary content
	binaryContent := []byte("test binary content for update test")
	h := sha256.Sum256(binaryContent)
	binaryChecksum := hex.EncodeToString(h[:])

	// Create temp directory for test binary
	tmpDir := t.TempDir()
	testBinaryPath := filepath.Join(tmpDir, "test-binary")
	if err := os.WriteFile(testBinaryPath, []byte("old binary"), 0755); err != nil {
		t.Fatalf("Failed to create test binary: %v", err)
	}

	assetName := "ha-ws-client-darwin-arm64"
	checksumContent := binaryChecksum + "  " + assetName

	releases := []Release{
		{
			TagName: "ha-ws-client-go/v1.6.0",
			Assets: []Asset{
				{Name: assetName, BrowserDownloadURL: "/binary", Size: int64(len(binaryContent))},
				{Name: "checksums.txt", BrowserDownloadURL: "/checksums.txt", Size: int64(len(checksumContent))},
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/repos/test/test/releases":
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(releases)
		case "/binary":
			w.Write(binaryContent)
		case "/checksums.txt":
			w.Write([]byte(checksumContent))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	// Update the release URLs to use the test server
	releases[0].Assets[0].BrowserDownloadURL = server.URL + "/binary"
	releases[0].Assets[1].BrowserDownloadURL = server.URL + "/checksums.txt"

	client := NewGitHubClient(
		WithBaseURL(server.URL),
		WithRepository("test", "test"),
	)

	u, err := NewUpdater("ha-ws-client", "ha-ws-client-go", "1.5.4",
		WithGitHubClient(client),
		WithQuietMode(),
	)
	if err != nil {
		t.Fatalf("NewUpdater() error = %v", err)
	}

	// Override getBinaryPath to use our test binary
	// Since we can't easily mock os.Executable(), we'll test the components individually

	// Test download function (use system temp dir for test)
	tempPath, err := u.download(server.URL+"/binary", int64(len(binaryContent)), os.TempDir())
	if err != nil {
		t.Fatalf("download() error = %v", err)
	}
	defer os.Remove(tempPath)

	// Verify the downloaded content
	downloadedContent, err := os.ReadFile(tempPath)
	if err != nil {
		t.Fatalf("reading downloaded file: %v", err)
	}
	if string(downloadedContent) != string(binaryContent) {
		t.Error("downloaded content doesn't match expected")
	}

	// Test checksum verification
	err = VerifyChecksum(tempPath, binaryChecksum)
	if err != nil {
		t.Errorf("VerifyChecksum() error = %v", err)
	}

	// Test checksum verification with wrong hash
	err = VerifyChecksum(tempPath, "0000000000000000000000000000000000000000000000000000000000000000")
	if !errors.Is(err, ErrChecksumMismatch) {
		t.Errorf("VerifyChecksum() should fail with ErrChecksumMismatch, got %v", err)
	}
}

func TestUpdater_checkWritable(t *testing.T) {
	u := &Updater{}

	// Writable directory
	tmpDir := t.TempDir()
	err := u.checkWritable(tmpDir)
	if err != nil {
		t.Errorf("checkWritable() error = %v for writable dir", err)
	}

	// Non-existent directory
	err = u.checkWritable("/nonexistent/directory/path")
	if err == nil {
		t.Error("checkWritable() should fail for non-existent directory")
	}
}

func TestUpdater_replaceBinary(t *testing.T) {
	u := &Updater{}

	tmpDir := t.TempDir()

	// Create original binary
	originalPath := filepath.Join(tmpDir, "original")
	if err := os.WriteFile(originalPath, []byte("original"), 0755); err != nil {
		t.Fatalf("Failed to create original: %v", err)
	}

	// Create new binary
	newPath := filepath.Join(tmpDir, "new")
	if err := os.WriteFile(newPath, []byte("new content"), 0644); err != nil {
		t.Fatalf("Failed to create new: %v", err)
	}

	// Replace
	err := u.replaceBinary(newPath, originalPath)
	if err != nil {
		t.Fatalf("replaceBinary() error = %v", err)
	}

	// Verify content was replaced
	content, err := os.ReadFile(originalPath)
	if err != nil {
		t.Fatalf("reading replaced binary: %v", err)
	}
	if string(content) != "new content" {
		t.Errorf("content = %q, want %q", string(content), "new content")
	}

	// Verify permissions were preserved
	info, err := os.Stat(originalPath)
	if err != nil {
		t.Fatalf("stat replaced binary: %v", err)
	}
	if info.Mode() != 0755 {
		t.Errorf("mode = %o, want 0755", info.Mode())
	}
}

func TestUpdateError(t *testing.T) {
	err := NewUpdateError("download", "ha-ws-client", "1.6.0", ErrNetworkTimeout)

	if err.Op != "download" {
		t.Errorf("Op = %q, want %q", err.Op, "download")
	}
	if err.Tool != "ha-ws-client" {
		t.Errorf("Tool = %q, want %q", err.Tool, "ha-ws-client")
	}
	if err.Version != "1.6.0" {
		t.Errorf("Version = %q, want %q", err.Version, "1.6.0")
	}

	expected := "download ha-ws-client v1.6.0: network timeout"
	if err.Error() != expected {
		t.Errorf("Error() = %q, want %q", err.Error(), expected)
	}

	// Test Unwrap
	if !errors.Is(err, ErrNetworkTimeout) {
		t.Error("errors.Is should match ErrNetworkTimeout")
	}
}

func TestUpdateError_NoVersion(t *testing.T) {
	err := NewUpdateError("check", "ha-ws-client", "", ErrNoRelease)

	expected := "check ha-ws-client: no release found"
	if err.Error() != expected {
		t.Errorf("Error() = %q, want %q", err.Error(), expected)
	}
}
