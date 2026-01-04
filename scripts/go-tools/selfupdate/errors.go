// Package selfupdate provides self-update functionality for CLI tools.
package selfupdate

import (
	"errors"
	"fmt"
)

// Sentinel errors for update operations.
var (
	// ErrNoRelease indicates no release was found for the tool.
	ErrNoRelease = errors.New("no release found")

	// ErrChecksumMismatch indicates the downloaded file checksum doesn't match.
	ErrChecksumMismatch = errors.New("checksum mismatch")

	// ErrNetworkTimeout indicates a network operation timed out.
	ErrNetworkTimeout = errors.New("network timeout")

	// ErrPermissionDenied indicates insufficient permissions for the operation.
	ErrPermissionDenied = errors.New("permission denied")

	// ErrRateLimited indicates GitHub API rate limit was exceeded.
	ErrRateLimited = errors.New("rate limit exceeded")

	// ErrUnsupportedArchitecture indicates the current architecture is not supported.
	ErrUnsupportedArchitecture = errors.New("unsupported architecture")

	// ErrMissingChecksum indicates the checksums.txt file is missing from the release.
	ErrMissingChecksum = errors.New("checksums.txt missing from release")

	// ErrAssetNotFound indicates the binary for the current architecture is missing.
	ErrAssetNotFound = errors.New("asset not found for architecture")

	// ErrAlreadyLatest indicates the tool is already at the latest version.
	ErrAlreadyLatest = errors.New("already at latest version")

	// ErrVersionNotFound indicates the specified version was not found.
	ErrVersionNotFound = errors.New("version not found")
)

// UpdateError wraps an error with additional context about the update operation.
type UpdateError struct {
	Op      string // Operation that failed (e.g., "check", "download", "verify", "replace")
	Tool    string // Tool name (e.g., "ha-ws-client", "validate-blueprint")
	Version string // Version involved in the operation (if applicable)
	Err     error  // Underlying error
}

// Error implements the error interface.
func (e *UpdateError) Error() string {
	if e.Version != "" {
		return fmt.Sprintf("%s %s v%s: %v", e.Op, e.Tool, e.Version, e.Err)
	}
	return fmt.Sprintf("%s %s: %v", e.Op, e.Tool, e.Err)
}

// Unwrap returns the underlying error for errors.Is and errors.As.
func (e *UpdateError) Unwrap() error {
	return e.Err
}

// NewUpdateError creates a new UpdateError with the given context.
func NewUpdateError(op, tool, version string, err error) *UpdateError {
	return &UpdateError{
		Op:      op,
		Tool:    tool,
		Version: version,
		Err:     err,
	}
}

// RateLimitError provides details about rate limiting.
type RateLimitError struct {
	ResetTime string // When the rate limit resets (human-readable)
	Remaining int    // Remaining requests (usually 0 when this error occurs)
}

// Error implements the error interface.
func (e *RateLimitError) Error() string {
	if e.ResetTime != "" {
		return fmt.Sprintf("GitHub API rate limit exceeded, resets at %s", e.ResetTime)
	}
	return "GitHub API rate limit exceeded"
}

// Is allows errors.Is to match against ErrRateLimited.
func (e *RateLimitError) Is(target error) bool {
	return target == ErrRateLimited
}

// ArchitectureError provides details about unsupported architectures.
type ArchitectureError struct {
	OS                     string   // Current OS
	Arch                   string   // Current architecture
	SupportedArchitectures []string // List of supported architectures
}

// Error implements the error interface.
func (e *ArchitectureError) Error() string {
	return fmt.Sprintf("architecture %s-%s is not supported; supported: %v", e.OS, e.Arch, e.SupportedArchitectures)
}

// Is allows errors.Is to match against ErrUnsupportedArchitecture.
func (e *ArchitectureError) Is(target error) bool {
	return target == ErrUnsupportedArchitecture
}

// ChecksumError provides details about checksum verification failures.
type ChecksumError struct {
	Expected string // Expected checksum from checksums.txt
	Actual   string // Actual checksum of downloaded file
	File     string // File that was verified
}

// Error implements the error interface.
func (e *ChecksumError) Error() string {
	return fmt.Sprintf("checksum mismatch for %s: expected %s, got %s", e.File, e.Expected, e.Actual)
}

// Is allows errors.Is to match against ErrChecksumMismatch.
func (e *ChecksumError) Is(target error) bool {
	return target == ErrChecksumMismatch
}

// DownloadError provides details about download failures.
type DownloadError struct {
	URL        string // URL that failed to download
	StatusCode int    // HTTP status code (if applicable)
	Err        error  // Underlying error
}

// Error implements the error interface.
func (e *DownloadError) Error() string {
	if e.StatusCode > 0 {
		return fmt.Sprintf("download failed for %s: HTTP %d", e.URL, e.StatusCode)
	}
	return fmt.Sprintf("download failed for %s: %v", e.URL, e.Err)
}

// Unwrap returns the underlying error.
func (e *DownloadError) Unwrap() error {
	return e.Err
}

// PermissionError provides details about permission failures.
type PermissionError struct {
	Path string // Path that couldn't be accessed
	Op   string // Operation that failed (e.g., "write", "replace")
}

// Error implements the error interface.
func (e *PermissionError) Error() string {
	return fmt.Sprintf("permission denied: cannot %s %s (try running with elevated privileges)", e.Op, e.Path)
}

// Is allows errors.Is to match against ErrPermissionDenied.
func (e *PermissionError) Is(target error) bool {
	return target == ErrPermissionDenied
}
