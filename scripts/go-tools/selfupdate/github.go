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

// ListReleasesForTool returns releases filtered by tool-specific tags.
// Tool-specific tags follow the pattern: {toolTag}/vX.Y.Z
// For example: "ha-ws-client-go/v1.5.4" or "validate-blueprint-go/v1.6.0"
func (c *GitHubClient) ListReleasesForTool(toolTag string) ([]Release, error) {
	releases, err := c.ListReleases()
	if err != nil {
		return nil, err
	}

	prefix := toolTag + "/v"
	var filtered []Release
	for _, r := range releases {
		if strings.HasPrefix(r.TagName, prefix) {
			filtered = append(filtered, r)
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
	releases, err := c.ListReleasesForTool(toolTag)
	if err != nil {
		return nil, err
	}

	// Releases are returned in chronological order (newest first)
	// Return the first one as the latest
	return &releases[0], nil
}

// GetReleaseForToolVersion returns a specific version release for a tool.
// toolTag is like "ha-ws-client-go", version is like "1.5.4" (without 'v' prefix)
func (c *GitHubClient) GetReleaseForToolVersion(toolTag, version string) (*Release, error) {
	tag := fmt.Sprintf("%s/v%s", toolTag, version)
	release, err := c.GetReleaseByTag(tag)
	if err != nil {
		// Check if it's a 404
		var downloadErr *DownloadError
		if errors.As(err, &downloadErr) && downloadErr.StatusCode == 404 {
			return nil, ErrVersionNotFound
		}
		return nil, err
	}
	return release, nil
}

// ExtractVersion extracts the version from a tool-specific tag.
// For "ha-ws-client-go/v1.5.4", returns "1.5.4"
func ExtractVersion(tagName, toolTag string) string {
	prefix := toolTag + "/v"
	if after, ok :=strings.CutPrefix(tagName, prefix); ok  {
		return after
	}
	return ""
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

// doRequest performs an HTTP GET request and handles common error cases.
func (c *GitHubClient) doRequest(url string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
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
			remaining, _ := strconv.Atoi(resp.Header.Get("X-RateLimit-Remaining"))
			return &RateLimitError{ResetTime: resetTime, Remaining: remaining}
		}
		return &DownloadError{URL: resp.Request.URL.String(), StatusCode: resp.StatusCode}

	case http.StatusNotFound:
		return &DownloadError{URL: resp.Request.URL.String(), StatusCode: resp.StatusCode}

	default:
		return &DownloadError{URL: resp.Request.URL.String(), StatusCode: resp.StatusCode}
	}
}
