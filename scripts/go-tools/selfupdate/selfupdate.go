package selfupdate

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

const (
	// DefaultDownloadTimeout is the default timeout for binary downloads (120 seconds).
	DefaultDownloadTimeout = 120 * time.Second

	// DefaultCheckTimeout is the default timeout for version check operations (30 seconds).
	DefaultCheckTimeout = 30 * time.Second
)

// UpdateResult contains the result of an update check.
type UpdateResult struct {
	CurrentVersion  string // Currently installed version
	LatestVersion   string // Latest available version
	UpdateAvailable bool   // Whether an update is available
	DownloadURL     string // URL to download if update available
	ChecksumURL     string // URL to download checksums.txt
	AssetSize       int64  // Size of the binary asset in bytes
}

// Updater provides self-update functionality for CLI tools.
type Updater struct {
	// ToolName is the name of the tool (e.g., "ha-ws-client", "validate-blueprint").
	ToolName string

	// ToolTag is the GitHub tag prefix (e.g., "ha-ws-client-go", "validate-blueprint-go").
	ToolTag string

	// CurrentVersion is the current version of the tool.
	CurrentVersion string

	// Platform is the detected platform information.
	Platform Platform

	// github is the GitHub API client.
	github *GitHubClient

	// output is where progress and status messages are written.
	output io.Writer

	// downloadTimeout is the timeout for binary downloads.
	downloadTimeout time.Duration

	// quiet disables progress output.
	quiet bool
}

// UpdaterOption is a functional option for configuring Updater.
type UpdaterOption func(*Updater)

// WithOutput sets the output writer for progress and status messages.
func WithOutput(w io.Writer) UpdaterOption {
	return func(u *Updater) {
		u.output = w
	}
}

// WithDownloadTimeout sets the timeout for binary downloads.
func WithDownloadTimeout(timeout time.Duration) UpdaterOption {
	return func(u *Updater) {
		u.downloadTimeout = timeout
	}
}

// WithGitHubClient sets a custom GitHub client (useful for testing).
func WithGitHubClient(client *GitHubClient) UpdaterOption {
	return func(u *Updater) {
		u.github = client
	}
}

// WithQuietMode disables progress output.
func WithQuietMode() UpdaterOption {
	return func(u *Updater) {
		u.quiet = true
	}
}

// NewUpdater creates a new Updater for the given tool.
func NewUpdater(toolName, toolTag, currentVersion string, opts ...UpdaterOption) (*Updater, error) {
	platform, err := DetectPlatform()
	if err != nil {
		return nil, err
	}

	u := &Updater{
		ToolName:        toolName,
		ToolTag:         toolTag,
		CurrentVersion:  currentVersion,
		Platform:        platform,
		github:          NewGitHubClient(),
		output:          os.Stderr,
		downloadTimeout: DefaultDownloadTimeout,
	}

	for _, opt := range opts {
		opt(u)
	}

	return u, nil
}

// Check checks for available updates without downloading.
func (u *Updater) Check() (*UpdateResult, error) {
	release, err := u.github.GetLatestReleaseForToolWithName(u.ToolTag, u.ToolName)
	if err != nil {
		return nil, NewUpdateError("check", u.ToolName, "", err)
	}

	latestVersion, err := u.extractVersionForRelease(release)
	if err != nil {
		return nil, NewUpdateError("check", u.ToolName, "", err)
	}

	result := &UpdateResult{
		CurrentVersion:  u.CurrentVersion,
		LatestVersion:   latestVersion,
		UpdateAvailable: u.needsUpdate(latestVersion),
	}

	if result.UpdateAvailable {
		assetName := u.Platform.AssetName(u.ToolName)
		asset := release.FindAssetByName(assetName)
		if asset != nil {
			result.DownloadURL = asset.BrowserDownloadURL
			result.AssetSize = asset.Size
		}

		checksumAsset := release.FindChecksumsAsset()
		if checksumAsset != nil {
			result.ChecksumURL = checksumAsset.BrowserDownloadURL
		}
	}

	return result, nil
}

// Update updates to the latest version.
func (u *Updater) Update() error {
	return u.UpdateToVersion("")
}

// UpdateToVersion updates to a specific version (or latest if version is empty).
func (u *Updater) UpdateToVersion(version string) error {
	var release *Release
	var err error

	if version == "" {
		release, err = u.github.GetLatestReleaseForToolWithName(u.ToolTag, u.ToolName)
	} else {
		release, err = u.github.GetReleaseForToolVersionWithName(u.ToolTag, u.ToolName, version)
	}
	if err != nil {
		return NewUpdateError("check", u.ToolName, version, err)
	}

	targetVersion, err := u.extractVersionForRelease(release)
	if err != nil {
		return NewUpdateError("check", u.ToolName, version, err)
	}

	// Check if we're already at this version
	if !u.needsUpdate(targetVersion) && version == "" {
		return ErrAlreadyLatest
	}

	// Find the binary asset for our platform
	assetName := u.Platform.AssetName(u.ToolName)
	asset := release.FindAssetByName(assetName)
	if asset == nil {
		return NewUpdateError("download", u.ToolName, targetVersion, &ArchitectureError{
			OS:                     u.Platform.OS,
			Arch:                   u.Platform.ArchString(),
			SupportedArchitectures: SupportedArchitectures,
		})
	}

	// Find checksums.txt
	checksumAsset := release.FindChecksumsAsset()
	if checksumAsset == nil {
		return NewUpdateError("verify", u.ToolName, targetVersion, ErrMissingChecksum)
	}

	// Download checksums
	checksums, err := DownloadChecksums(checksumAsset.BrowserDownloadURL, DefaultCheckTimeout)
	if err != nil {
		return NewUpdateError("verify", u.ToolName, targetVersion, err)
	}

	expectedChecksum := checksums.GetChecksum(assetName)
	if expectedChecksum == "" {
		return NewUpdateError("verify", u.ToolName, targetVersion, fmt.Errorf("checksum not found for %s", assetName))
	}

	// Get the current binary path (resolving symlinks)
	binaryPath, err := u.getBinaryPath()
	if err != nil {
		return NewUpdateError("replace", u.ToolName, targetVersion, err)
	}

	// Download to temp file in same directory as binary (for atomic rename)
	targetDir := filepath.Dir(binaryPath)
	tempPath, err := u.download(asset.BrowserDownloadURL, asset.Size, targetDir)
	if err != nil {
		return NewUpdateError("download", u.ToolName, targetVersion, err)
	}
	defer os.Remove(tempPath) // Clean up on error

	// Verify checksum
	if err := VerifyChecksum(tempPath, expectedChecksum); err != nil {
		return NewUpdateError("verify", u.ToolName, targetVersion, err)
	}

	// Replace the binary atomically
	if err := u.replaceBinary(tempPath, binaryPath); err != nil {
		return NewUpdateError("replace", u.ToolName, targetVersion, err)
	}

	return nil
}

// needsUpdate returns true if the current version is different from the target version.
// For 'dev' builds, always returns true to allow testing the update flow.
func (u *Updater) needsUpdate(targetVersion string) bool {
	if u.CurrentVersion == "dev" {
		return true
	}
	return u.CurrentVersion != targetVersion
}

// getBinaryPath returns the path to the current binary, resolving symlinks.
func (u *Updater) getBinaryPath() (string, error) {
	execPath, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("getting executable path: %w", err)
	}

	// Resolve symlinks to get the actual binary location
	realPath, err := filepath.EvalSymlinks(execPath)
	if err != nil {
		return "", fmt.Errorf("resolving symlinks: %w", err)
	}

	// Check if we can write to the directory
	dir := filepath.Dir(realPath)
	if err := u.checkWritable(dir); err != nil {
		return "", &PermissionError{Path: realPath, Op: "write"}
	}

	return realPath, nil
}

// checkWritable checks if a directory is writable.
func (u *Updater) checkWritable(dir string) error {
	testFile := filepath.Join(dir, ".selfupdate-test")
	f, err := os.Create(testFile)
	if err != nil {
		return err
	}
	f.Close()
	return os.Remove(testFile)
}

// download downloads the binary from the given URL to a temp file in the target directory.
// The target directory should be the same filesystem as the final destination to enable atomic rename.
func (u *Updater) download(url string, size int64, targetDir string) (string, error) {
	client := &http.Client{Timeout: u.downloadTimeout}

	req, err := http.NewRequest(http.MethodGet, url, http.NoBody)
	if err != nil {
		return "", fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("User-Agent", "selfupdate-go-client")

	resp, err := client.Do(req)
	if err != nil {
		return "", &DownloadError{URL: url, Err: err}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", &DownloadError{URL: url, StatusCode: resp.StatusCode}
	}

	// Create temp file in the target directory (same filesystem for atomic rename)
	tempFile, err := os.CreateTemp(targetDir, ".selfupdate-*")
	if err != nil {
		return "", fmt.Errorf("creating temp file: %w", err)
	}
	tempPath := tempFile.Name()

	// Set up progress tracking
	var progressOpts []ProgressOption
	if u.quiet {
		progressOpts = append(progressOpts, WithQuiet())
	}
	progress := NewProgressWriter(u.output, size, progressOpts...)
	progressReader := NewProgressReader(resp.Body, progress)

	// Copy with progress tracking
	_, err = io.Copy(tempFile, progressReader)
	if closeErr := tempFile.Close(); closeErr != nil && err == nil {
		err = closeErr
	}

	progress.Finish()

	if err != nil {
		os.Remove(tempPath)
		return "", fmt.Errorf("downloading binary: %w", err)
	}

	return tempPath, nil
}

// replaceBinary atomically replaces the current binary with the new one.
// On Unix, this uses atomic rename. On Windows, we rename the old binary first
// since a running exe can be renamed but not overwritten.
func (u *Updater) replaceBinary(tempPath, binaryPath string) error {
	// Get the current binary's permissions
	info, err := os.Stat(binaryPath)
	if err != nil {
		return fmt.Errorf("getting binary info: %w", err)
	}

	// Set the same permissions on the temp file
	if err := os.Chmod(tempPath, info.Mode()); err != nil {
		return fmt.Errorf("setting permissions: %w", err)
	}

	// Strategy: rename old binary out of the way, then rename new binary into place.
	// This works on both Unix and Windows:
	// - Unix: allows atomic rename over existing file, but this approach also works
	// - Windows: running exe can be renamed but not overwritten, so we must rename it first
	oldPath := binaryPath + ".old"

	// Remove any leftover .old file from previous updates
	os.Remove(oldPath)

	// Rename current binary to .old
	if err := os.Rename(binaryPath, oldPath); err != nil {
		// On Unix, try direct rename (atomic overwrite)
		if err := os.Rename(tempPath, binaryPath); err != nil {
			return fmt.Errorf("replacing binary: %w", err)
		}
		return nil
	}

	// Move new binary into place
	if err := os.Rename(tempPath, binaryPath); err != nil {
		// Try to restore the old binary
		if restoreErr := os.Rename(oldPath, binaryPath); restoreErr != nil {
			return fmt.Errorf("replacing binary failed and could not restore: %w (restore error: %w)", err, restoreErr)
		}
		return fmt.Errorf("replacing binary: %w", err)
	}

	// Clean up old binary (may fail on Windows if still running - that's ok)
	_ = os.Remove(oldPath)

	return nil
}

// ListAvailableVersions returns a list of available versions for the tool.
// Each release must have a versions.json asset with the tool's version.
func (u *Updater) ListAvailableVersions() ([]string, error) {
	releases, err := u.github.ListReleasesForToolWithName(u.ToolTag, u.ToolName)
	if err != nil {
		return nil, NewUpdateError("list", u.ToolName, "", err)
	}

	versions := make([]string, 0, len(releases))
	for i := range releases {
		version, err := u.extractVersionForRelease(&releases[i])
		if err != nil {
			// Skip releases without valid versions.json
			continue
		}
		versions = append(versions, version)
	}

	if len(versions) == 0 {
		return nil, NewUpdateError("list", u.ToolName, "", fmt.Errorf("no releases with versions.json found"))
	}

	return versions, nil
}

// PrintAvailableVersions lists all available versions to the configured output.
// It marks the current version with "(current)" suffix.
func (u *Updater) PrintAvailableVersions() error {
	versions, err := u.ListAvailableVersions()
	if err != nil {
		return err
	}

	if len(versions) == 0 {
		fmt.Fprintln(u.output, "No versions available.")
		return nil
	}

	fmt.Fprintf(u.output, "Available versions (%d):\n", len(versions))
	for _, v := range versions {
		if v == u.CurrentVersion {
			fmt.Fprintf(u.output, "  - %s (current)\n", v)
		} else {
			fmt.Fprintf(u.output, "  - %s\n", v)
		}
	}

	return nil
}

// extractVersionForRelease extracts the version for this tool from a release.
// Requires versions.json asset in the release.
func (u *Updater) extractVersionForRelease(release *Release) (string, error) {
	versionsAsset := release.FindVersionsAsset()
	if versionsAsset == nil {
		return "", fmt.Errorf("versions.json not found in release %s", release.TagName)
	}

	versions, err := DownloadVersions(versionsAsset.BrowserDownloadURL, DefaultCheckTimeout)
	if err != nil {
		return "", fmt.Errorf("downloading versions.json: %w", err)
	}

	if versions == nil {
		return "", fmt.Errorf("versions.json is empty in release %s", release.TagName)
	}

	version := versions.GetVersion(u.ToolTag)
	if version == "" {
		return "", fmt.Errorf("tool %s not found in versions.json for release %s", u.ToolTag, release.TagName)
	}

	return version, nil
}
