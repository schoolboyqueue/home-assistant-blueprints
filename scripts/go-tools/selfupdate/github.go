package selfupdate

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	// DefaultGitHubAPIBase is the base URL for GitHub API.
	DefaultGitHubAPIBase = "https://api.github.com"

	// DefaultRepoOwner is the default GitHub repository owner.
	DefaultRepoOwner = "schoolboyqueue"

	// DefaultRepoName is the default GitHub repository name.
	DefaultRepoName = "home-assistant-blueprints"
)

// Release represents a GitHub release.
type Release struct {
	TagName     string    `json:"tag_name"`
	Name        string    `json:"name"`
	PublishedAt time.Time `json:"published_at"`
	Assets      []Asset   `json:"assets"`
}

// Asset represents a downloadable file from a GitHub release.
type Asset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
	Size               int64  `json:"size"`
}

// GitHubClient is a client for interacting with the GitHub API.
type GitHubClient struct {
	httpClient *http.Client
	baseURL    string
	repoOwner  string
	repoName   string
}

// GitHubClientOption is a functional option for configuring GitHubClient.
type GitHubClientOption func(*GitHubClient)

// WithHTTPClient sets a custom HTTP client.
func WithHTTPClient(client *http.Client) GitHubClientOption {
	return func(c *GitHubClient) {
		c.httpClient = client
	}
}

// WithBaseURL sets a custom base URL (useful for testing).
func WithBaseURL(baseURL string) GitHubClientOption {
	return func(c *GitHubClient) {
		c.baseURL = baseURL
	}
}

// WithRepository sets a custom repository owner and name.
func WithRepository(owner, name string) GitHubClientOption {
	return func(c *GitHubClient) {
		c.repoOwner = owner
		c.repoName = name
	}
}

// NewGitHubClient creates a new GitHub API client.
func NewGitHubClient(opts ...GitHubClientOption) *GitHubClient {
	c := &GitHubClient{
		httpClient: &http.Client{Timeout: 30 * time.Second},
		baseURL:    DefaultGitHubAPIBase,
		repoOwner:  DefaultRepoOwner,
		repoName:   DefaultRepoName,
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// ListReleases fetches all releases from the repository.
func (c *GitHubClient) ListReleases() ([]Release, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/releases?per_page=100", c.baseURL, c.repoOwner, c.repoName)

	resp, err := c.doRequest(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}

	var releases []Release
	if err := json.Unmarshal(body, &releases); err != nil {
		return nil, fmt.Errorf("parsing releases JSON: %w", err)
	}

	return releases, nil
}

// GetReleaseByTag fetches a specific release by tag name.
func (c *GitHubClient) GetReleaseByTag(tag string) (*Release, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/releases/tags/%s", c.baseURL, c.repoOwner, c.repoName, tag)

	resp, err := c.doRequest(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}

	var release Release
	if err := json.Unmarshal(body, &release); err != nil {
		return nil, fmt.Errorf("parsing release JSON: %w", err)
	}

	return &release, nil
}

// ListReleasesForTool returns releases filtered by tool-specific tags or combined releases.
// Matches two patterns:
//   - Tool-specific tags: {toolTag}/vX.Y.Z (e.g., "ha-ws-client-go/v1.5.4")
//   - Combined releases: vX.Y.Z (e.g., "v1.6.0") that contain the tool's binary
//
// The toolName parameter is the binary name (e.g., "ha-ws-client") used to check
// if a combined release contains the tool's assets.
func (c *GitHubClient) ListReleasesForTool(toolTag string) ([]Release, error) {
	return c.ListReleasesForToolWithName(toolTag, "")
}

// ListReleasesForToolWithName returns releases filtered by tool-specific tags or combined releases.
// If toolName is provided, combined releases (v*) are also checked for matching assets.
func (c *GitHubClient) ListReleasesForToolWithName(toolTag, toolName string) ([]Release, error) {
	releases, err := c.ListReleases()
	if err != nil {
		return nil, err
	}

	toolPrefix := toolTag + "/v"
	var filtered []Release
	for _, r := range releases {
		// Match tool-specific tags: ha-ws-client-go/v1.5.4
		if strings.HasPrefix(r.TagName, toolPrefix) {
			filtered = append(filtered, r)
			continue
		}

		// Match combined releases: v1.6.0 (must have tool's binary in assets)
		if toolName != "" && strings.HasPrefix(r.TagName, "v") && !strings.Contains(r.TagName, "/") {
			// Check if release has this tool's binary
			for _, asset := range r.Assets {
				if strings.HasPrefix(asset.Name, toolName+"-") {
					filtered = append(filtered, r)
					break
				}
			}
		}
	}

	if len(filtered) == 0 {
		return nil, ErrNoRelease
	}

	return filtered, nil
}

// GetLatestReleaseForTool returns the latest release for a specific tool.
// Tool-specific tags follow the pattern: {toolTag}/vX.Y.Z
func (c *GitHubClient) GetLatestReleaseForTool(toolTag string) (*Release, error) {
	return c.GetLatestReleaseForToolWithName(toolTag, "")
}

// GetLatestReleaseForToolWithName returns the latest release for a specific tool.
// If toolName is provided, combined releases (v*) are also considered.
func (c *GitHubClient) GetLatestReleaseForToolWithName(toolTag, toolName string) (*Release, error) {
	releases, err := c.ListReleasesForToolWithName(toolTag, toolName)
	if err != nil {
		return nil, err
	}

	// Releases are returned in chronological order (newest first)
	// Return the first one as the latest
	return &releases[0], nil
}

// GetReleaseForToolVersion returns a specific version release for a tool.
// toolTag is like "ha-ws-client-go", version is like "1.5.4" (without 'v' prefix)
// toolName is the binary name (e.g., "ha-ws-client") for searching combined releases.
func (c *GitHubClient) GetReleaseForToolVersion(toolTag, version string) (*Release, error) {
	return c.GetReleaseForToolVersionWithName(toolTag, "", version)
}

// GetReleaseForToolVersionWithName returns a specific version release for a tool.
// It first tries the tool-specific tag (ha-ws-client-go/v1.5.4), then searches
// combined releases (v*) by checking versions.json.
func (c *GitHubClient) GetReleaseForToolVersionWithName(toolTag, toolName, version string) (*Release, error) {
	// First, try tool-specific tag (fast path)
	tag := fmt.Sprintf("%s/v%s", toolTag, version)
	release, err := c.GetReleaseByTag(tag)
	if err == nil {
		return release, nil
	}

	// If not a 404, return the error
	var downloadErr *DownloadError
	if !errors.As(err, &downloadErr) || downloadErr.StatusCode != 404 {
		return nil, err
	}

	// Tool-specific tag not found, search combined releases
	if toolName == "" {
		return nil, ErrVersionNotFound
	}

	// Get all combined releases that have this tool
	releases, err := c.ListReleasesForToolWithName(toolTag, toolName)
	if err != nil {
		if errors.Is(err, ErrNoRelease) {
			return nil, ErrVersionNotFound
		}
		return nil, err
	}

	// Search for a combined release with matching version in versions.json
	for i := range releases {
		r := &releases[i]
		// Skip tool-specific releases (already tried above)
		if strings.HasPrefix(r.TagName, toolTag+"/") {
			continue
		}

		// Check versions.json for this release
		versionsAsset := r.FindVersionsAsset()
		if versionsAsset == nil {
			continue
		}

		versions, err := DownloadVersions(versionsAsset.BrowserDownloadURL, 30*time.Second)
		if err != nil || versions == nil {
			continue
		}

		if versions.GetVersion(toolTag) == version {
			return r, nil
		}
	}

	return nil, ErrVersionNotFound
}

// FindAssetByName finds an asset by its name in a release.
func (r *Release) FindAssetByName(name string) *Asset {
	for i := range r.Assets {
		if r.Assets[i].Name == name {
			return &r.Assets[i]
		}
	}
	return nil
}

// FindChecksumsAsset finds the checksums.txt asset in a release.
func (r *Release) FindChecksumsAsset() *Asset {
	return r.FindAssetByName("checksums.txt")
}

// FindVersionsAsset finds the versions.json asset in a release.
func (r *Release) FindVersionsAsset() *Asset {
	return r.FindAssetByName("versions.json")
}

// ToolVersions maps tool tags to their versions from versions.json.
type ToolVersions map[string]string

// GetVersion returns the version for a specific tool tag.
// Returns empty string if the tool is not in the versions map.
func (v ToolVersions) GetVersion(toolTag string) string {
	return v[toolTag]
}

// DownloadVersions downloads and parses versions.json from the given URL.
// Returns nil (not an error) if the file doesn't exist (404), allowing
// fallback to legacy version extraction for older releases.
func DownloadVersions(url string, timeout time.Duration) (ToolVersions, error) {
	client := &http.Client{Timeout: timeout}

	req, err := http.NewRequest(http.MethodGet, url, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("User-Agent", "selfupdate-go-client")

	resp, err := client.Do(req)
	if err != nil {
		return nil, &DownloadError{URL: url, Err: err}
	}
	defer resp.Body.Close()

	// Return nil for 404 - versions.json may not exist in older releases
	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	if resp.StatusCode != http.StatusOK {
		return nil, &DownloadError{URL: url, StatusCode: resp.StatusCode}
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}

	var versions ToolVersions
	if err := json.Unmarshal(body, &versions); err != nil {
		return nil, fmt.Errorf("parsing versions JSON: %w", err)
	}

	return versions, nil
}

// doRequest performs an HTTP GET request and handles common error cases.
func (c *GitHubClient) doRequest(url string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, url, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "selfupdate-go-client")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, &DownloadError{URL: url, Err: err}
	}

	if err := c.handleHTTPError(resp); err != nil {
		resp.Body.Close()
		return nil, err
	}

	return resp, nil
}

// handleHTTPError checks for and returns appropriate errors based on HTTP status code.
func (c *GitHubClient) handleHTTPError(resp *http.Response) error {
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}

	switch resp.StatusCode {
	case http.StatusForbidden:
		// Could be rate limiting
		if resp.Header.Get("X-RateLimit-Remaining") == "0" {
			resetTime := ""
			if resetUnix := resp.Header.Get("X-RateLimit-Reset"); resetUnix != "" {
				if ts, err := strconv.ParseInt(resetUnix, 10, 64); err == nil {
					resetTime = time.Unix(ts, 0).Format(time.RFC1123)
				}
			}
			return &RateLimitError{ResetTime: resetTime, Remaining: 0}
		}
		return &DownloadError{URL: resp.Request.URL.String(), StatusCode: resp.StatusCode}

	case http.StatusNotFound:
		return &DownloadError{URL: resp.Request.URL.String(), StatusCode: resp.StatusCode}

	default:
		return &DownloadError{URL: resp.Request.URL.String(), StatusCode: resp.StatusCode}
	}
}
