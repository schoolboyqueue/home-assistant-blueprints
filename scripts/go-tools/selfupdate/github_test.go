package selfupdate

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewGitHubClient_Defaults(t *testing.T) {
	client := NewGitHubClient()

	if client.baseURL != DefaultGitHubAPIBase {
		t.Errorf("baseURL = %q, want %q", client.baseURL, DefaultGitHubAPIBase)
	}
	if client.repoOwner != DefaultRepoOwner {
		t.Errorf("repoOwner = %q, want %q", client.repoOwner, DefaultRepoOwner)
	}
	if client.repoName != DefaultRepoName {
		t.Errorf("repoName = %q, want %q", client.repoName, DefaultRepoName)
	}
}

func TestNewGitHubClient_WithOptions(t *testing.T) {
	customClient := &http.Client{Timeout: 60 * time.Second}

	client := NewGitHubClient(
		WithHTTPClient(customClient),
		WithBaseURL("https://custom.api.example.com"),
		WithRepository("custom-owner", "custom-repo"),
	)

	if client.httpClient != customClient {
		t.Error("httpClient was not set correctly")
	}
	if client.baseURL != "https://custom.api.example.com" {
		t.Errorf("baseURL = %q, want %q", client.baseURL, "https://custom.api.example.com")
	}
	if client.repoOwner != "custom-owner" {
		t.Errorf("repoOwner = %q, want %q", client.repoOwner, "custom-owner")
	}
	if client.repoName != "custom-repo" {
		t.Errorf("repoName = %q, want %q", client.repoName, "custom-repo")
	}
}

func TestGitHubClient_ListReleases(t *testing.T) {
	releases := []Release{
		{
			TagName:     "ha-ws-client-go/v1.6.0",
			Name:        "ha-ws-client-go v1.6.0",
			PublishedAt: time.Now(),
			Assets: []Asset{
				{Name: "ha-ws-client-linux-amd64", BrowserDownloadURL: "https://example.com/ha-ws-client-linux-amd64", Size: 1024},
				{Name: "checksums.txt", BrowserDownloadURL: "https://example.com/checksums.txt", Size: 256},
			},
		},
		{
			TagName:     "validate-blueprint-go/v1.5.0",
			Name:        "validate-blueprint-go v1.5.0",
			PublishedAt: time.Now().Add(-24 * time.Hour),
			Assets: []Asset{
				{Name: "validate-blueprint-linux-amd64", BrowserDownloadURL: "https://example.com/validate-blueprint-linux-amd64", Size: 2048},
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/repos/test-owner/test-repo/releases" {
			t.Errorf("unexpected path: %s", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(releases)
	}))
	defer server.Close()

	client := NewGitHubClient(
		WithBaseURL(server.URL),
		WithRepository("test-owner", "test-repo"),
	)

	result, err := client.ListReleases()
	if err != nil {
		t.Fatalf("ListReleases() error = %v", err)
	}

	if len(result) != 2 {
		t.Errorf("ListReleases() returned %d releases, want 2", len(result))
	}

	if result[0].TagName != "ha-ws-client-go/v1.6.0" {
		t.Errorf("first release TagName = %q, want %q", result[0].TagName, "ha-ws-client-go/v1.6.0")
	}
}

func TestGitHubClient_GetReleaseByTag(t *testing.T) {
	release := Release{
		TagName:     "ha-ws-client-go/v1.5.4",
		Name:        "ha-ws-client-go v1.5.4",
		PublishedAt: time.Now(),
		Assets: []Asset{
			{Name: "ha-ws-client-linux-amd64", BrowserDownloadURL: "https://example.com/ha-ws-client-linux-amd64", Size: 1024},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expectedPath := "/repos/test-owner/test-repo/releases/tags/ha-ws-client-go/v1.5.4"
		if r.URL.Path != expectedPath {
			t.Errorf("unexpected path: %s, want %s", r.URL.Path, expectedPath)
			w.WriteHeader(http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(release)
	}))
	defer server.Close()

	client := NewGitHubClient(
		WithBaseURL(server.URL),
		WithRepository("test-owner", "test-repo"),
	)

	result, err := client.GetReleaseByTag("ha-ws-client-go/v1.5.4")
	if err != nil {
		t.Fatalf("GetReleaseByTag() error = %v", err)
	}

	if result.TagName != "ha-ws-client-go/v1.5.4" {
		t.Errorf("TagName = %q, want %q", result.TagName, "ha-ws-client-go/v1.5.4")
	}
}

func TestGitHubClient_ListReleasesForTool(t *testing.T) {
	releases := []Release{
		{TagName: "ha-ws-client-go/v1.6.0"},
		{TagName: "validate-blueprint-go/v1.5.0"},
		{TagName: "ha-ws-client-go/v1.5.4"},
		{TagName: "ha-ws-client-go/v1.5.3"},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(releases)
	}))
	defer server.Close()

	client := NewGitHubClient(
		WithBaseURL(server.URL),
		WithRepository("test-owner", "test-repo"),
	)

	t.Run("filter ha-ws-client-go releases", func(t *testing.T) {
		result, err := client.ListReleasesForTool("ha-ws-client-go")
		if err != nil {
			t.Fatalf("ListReleasesForTool() error = %v", err)
		}

		if len(result) != 3 {
			t.Errorf("ListReleasesForTool() returned %d releases, want 3", len(result))
		}

		for _, r := range result {
			if r.TagName[:len("ha-ws-client-go")] != "ha-ws-client-go" {
				t.Errorf("unexpected release in result: %s", r.TagName)
			}
		}
	})

	t.Run("filter validate-blueprint-go releases", func(t *testing.T) {
		result, err := client.ListReleasesForTool("validate-blueprint-go")
		if err != nil {
			t.Fatalf("ListReleasesForTool() error = %v", err)
		}

		if len(result) != 1 {
			t.Errorf("ListReleasesForTool() returned %d releases, want 1", len(result))
		}
	})
}

func TestGitHubClient_ListReleasesForTool_NoReleases(t *testing.T) {
	releases := []Release{
		{TagName: "ha-ws-client-go/v1.6.0"},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(releases)
	}))
	defer server.Close()

	client := NewGitHubClient(
		WithBaseURL(server.URL),
		WithRepository("test-owner", "test-repo"),
	)

	_, err := client.ListReleasesForTool("nonexistent-tool")
	if !errors.Is(err, ErrNoRelease) {
		t.Errorf("ListReleasesForTool() error = %v, want ErrNoRelease", err)
	}
}

func TestGitHubClient_GetLatestReleaseForTool(t *testing.T) {
	// Releases are in chronological order (newest first)
	releases := []Release{
		{TagName: "ha-ws-client-go/v1.6.0"},
		{TagName: "ha-ws-client-go/v1.5.4"},
		{TagName: "ha-ws-client-go/v1.5.3"},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(releases)
	}))
	defer server.Close()

	client := NewGitHubClient(
		WithBaseURL(server.URL),
		WithRepository("test-owner", "test-repo"),
	)

	result, err := client.GetLatestReleaseForTool("ha-ws-client-go")
	if err != nil {
		t.Fatalf("GetLatestReleaseForTool() error = %v", err)
	}

	if result.TagName != "ha-ws-client-go/v1.6.0" {
		t.Errorf("TagName = %q, want %q", result.TagName, "ha-ws-client-go/v1.6.0")
	}
}

func TestGitHubClient_RateLimitError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("X-RateLimit-Remaining", "0")
		w.Header().Set("X-RateLimit-Reset", "1704412800") // Fixed timestamp for testing
		w.WriteHeader(http.StatusForbidden)
	}))
	defer server.Close()

	client := NewGitHubClient(
		WithBaseURL(server.URL),
		WithRepository("test-owner", "test-repo"),
	)

	_, err := client.ListReleases()
	if err == nil {
		t.Fatal("ListReleases() should return error on rate limit")
	}

	if !errors.Is(err, ErrRateLimited) {
		t.Errorf("error should be ErrRateLimited, got %v", err)
	}

	var rateLimitErr *RateLimitError
	if errors.As(err, &rateLimitErr) {
		if rateLimitErr.ResetTime == "" {
			t.Error("RateLimitError.ResetTime should be set")
		}
	} else {
		t.Error("error should be *RateLimitError")
	}
}

func TestGitHubClient_NotFoundError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := NewGitHubClient(
		WithBaseURL(server.URL),
		WithRepository("test-owner", "test-repo"),
	)

	_, err := client.GetReleaseByTag("nonexistent-tag")
	if err == nil {
		t.Fatal("GetReleaseByTag() should return error on 404")
	}

	var downloadErr *DownloadError
	if !errors.As(err, &downloadErr) {
		t.Errorf("error should be *DownloadError, got %T", err)
	}
	if downloadErr != nil && downloadErr.StatusCode != 404 {
		t.Errorf("StatusCode = %d, want 404", downloadErr.StatusCode)
	}
}

func TestRelease_FindAssetByName(t *testing.T) {
	release := Release{
		Assets: []Asset{
			{Name: "ha-ws-client-linux-amd64", BrowserDownloadURL: "https://example.com/1"},
			{Name: "ha-ws-client-darwin-arm64", BrowserDownloadURL: "https://example.com/2"},
			{Name: "checksums.txt", BrowserDownloadURL: "https://example.com/checksums"},
		},
	}

	t.Run("find existing asset", func(t *testing.T) {
		asset := release.FindAssetByName("ha-ws-client-linux-amd64")
		if asset == nil {
			t.Fatal("FindAssetByName() returned nil for existing asset")
		}
		if asset.BrowserDownloadURL != "https://example.com/1" {
			t.Errorf("BrowserDownloadURL = %q, want %q", asset.BrowserDownloadURL, "https://example.com/1")
		}
	})

	t.Run("find checksums", func(t *testing.T) {
		asset := release.FindChecksumsAsset()
		if asset == nil {
			t.Fatal("FindChecksumsAsset() returned nil")
		}
		if asset.Name != "checksums.txt" {
			t.Errorf("Name = %q, want %q", asset.Name, "checksums.txt")
		}
	})

	t.Run("asset not found", func(t *testing.T) {
		asset := release.FindAssetByName("nonexistent")
		if asset != nil {
			t.Error("FindAssetByName() should return nil for nonexistent asset")
		}
	})
}

func TestGitHubClient_GetReleaseForToolVersion(t *testing.T) {
	release := Release{
		TagName: "ha-ws-client-go/v1.5.4",
		Name:    "ha-ws-client-go v1.5.4",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/repos/test-owner/test-repo/releases/tags/ha-ws-client-go/v1.5.4" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(release)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := NewGitHubClient(
		WithBaseURL(server.URL),
		WithRepository("test-owner", "test-repo"),
	)

	t.Run("existing version", func(t *testing.T) {
		result, err := client.GetReleaseForToolVersion("ha-ws-client-go", "1.5.4")
		if err != nil {
			t.Fatalf("GetReleaseForToolVersion() error = %v", err)
		}
		if result.TagName != "ha-ws-client-go/v1.5.4" {
			t.Errorf("TagName = %q, want %q", result.TagName, "ha-ws-client-go/v1.5.4")
		}
	})

	t.Run("nonexistent version", func(t *testing.T) {
		_, err := client.GetReleaseForToolVersion("ha-ws-client-go", "9.9.9")
		if err == nil {
			t.Fatal("GetReleaseForToolVersion() should return error for nonexistent version")
		}
		// The error should be ErrVersionNotFound (which wraps the 404)
		if !errors.Is(err, ErrVersionNotFound) {
			t.Errorf("error should be ErrVersionNotFound, got %T: %v", err, err)
		}
	})
}

func TestGitHubClient_GetReleaseForToolVersionWithName_CombinedRelease(t *testing.T) {
	versionsJSON := `{"ha-ws-client-go": "1.6.0", "validate-blueprint-go": "1.7.0"}`

	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/repos/test-owner/test-repo/releases/tags/ha-ws-client-go/v1.6.0":
			// Tool-specific tag doesn't exist
			w.WriteHeader(http.StatusNotFound)
		case "/repos/test-owner/test-repo/releases":
			// Return combined release
			releases := []Release{
				{
					TagName: "v1.7.0",
					Name:    "ha-ws-client-go v1.6.0, validate-blueprint-go v1.7.0",
					Assets: []Asset{
						{Name: "ha-ws-client-linux-amd64", BrowserDownloadURL: "https://example.com/binary"},
						{Name: "versions.json", BrowserDownloadURL: server.URL + "/versions.json"},
					},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(releases)
		case "/versions.json":
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(versionsJSON))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client := NewGitHubClient(
		WithBaseURL(server.URL),
		WithRepository("test-owner", "test-repo"),
	)

	t.Run("find version in combined release", func(t *testing.T) {
		result, err := client.GetReleaseForToolVersionWithName("ha-ws-client-go", "ha-ws-client", "1.6.0")
		if err != nil {
			t.Fatalf("GetReleaseForToolVersionWithName() error = %v", err)
		}
		if result.TagName != "v1.7.0" {
			t.Errorf("TagName = %q, want %q", result.TagName, "v1.7.0")
		}
	})

	t.Run("version not in any release", func(t *testing.T) {
		_, err := client.GetReleaseForToolVersionWithName("ha-ws-client-go", "ha-ws-client", "9.9.9")
		if !errors.Is(err, ErrVersionNotFound) {
			t.Errorf("error should be ErrVersionNotFound, got %T: %v", err, err)
		}
	})
}
