package selfupdate

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// Checksums holds a map of filename to SHA256 hash.
type Checksums map[string]string

// DownloadChecksums downloads and parses checksums.txt from the given URL.
func DownloadChecksums(url string, timeout time.Duration) (Checksums, error) {
	client := &http.Client{Timeout: timeout}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("User-Agent", "selfupdate-go-client")

	resp, err := client.Do(req)
	if err != nil {
		return nil, &DownloadError{URL: url, Err: err}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, &DownloadError{URL: url, StatusCode: resp.StatusCode}
	}

	return ParseChecksums(resp.Body)
}

// ParseChecksums parses checksums.txt format (SHA256 hash, two spaces, filename).
// Format: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855  ha-ws-client-linux-amd64"
func ParseChecksums(r io.Reader) (Checksums, error) {
	checksums := make(Checksums)
	scanner := bufio.NewScanner(r)

	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		// Standard sha256sum format: hash (64 chars) + two spaces + filename
		// Also handle single space for compatibility
		parts := strings.SplitN(line, "  ", 2)
		if len(parts) != 2 {
			// Try single space
			parts = strings.SplitN(line, " ", 2)
			if len(parts) != 2 {
				return nil, fmt.Errorf("invalid checksum format on line %d: %q", lineNum, line)
			}
		}

		hash := strings.TrimSpace(parts[0])
		filename := strings.TrimSpace(parts[1])

		// Validate hash is 64 hex characters (SHA256)
		if len(hash) != 64 {
			return nil, fmt.Errorf("invalid hash length on line %d: expected 64 chars, got %d", lineNum, len(hash))
		}

		// Validate hash is valid hex
		if _, err := hex.DecodeString(hash); err != nil {
			return nil, fmt.Errorf("invalid hex in hash on line %d: %w", lineNum, err)
		}

		checksums[filename] = strings.ToLower(hash)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("reading checksums: %w", err)
	}

	if len(checksums) == 0 {
		return nil, fmt.Errorf("no checksums found in file")
	}

	return checksums, nil
}

// GetChecksum returns the expected checksum for a filename.
// Returns empty string if not found.
func (c Checksums) GetChecksum(filename string) string {
	return c[filename]
}

// VerifyChecksum verifies a file's SHA256 checksum against the expected hash.
// Returns nil if the checksum matches, or a ChecksumError if it doesn't.
func VerifyChecksum(filePath, expectedHash string) error {
	actualHash, err := ComputeChecksum(filePath)
	if err != nil {
		return fmt.Errorf("computing checksum: %w", err)
	}

	if !strings.EqualFold(actualHash, expectedHash) {
		return &ChecksumError{
			Expected: expectedHash,
			Actual:   actualHash,
			File:     filePath,
		}
	}

	return nil
}

// VerifyReaderChecksum verifies data from an io.Reader against the expected hash.
// Returns nil if the checksum matches, or a ChecksumError if it doesn't.
func VerifyReaderChecksum(r io.Reader, expectedHash, filename string) error {
	actualHash, err := ComputeReaderChecksum(r)
	if err != nil {
		return fmt.Errorf("computing checksum: %w", err)
	}

	if !strings.EqualFold(actualHash, expectedHash) {
		return &ChecksumError{
			Expected: expectedHash,
			Actual:   actualHash,
			File:     filename,
		}
	}

	return nil
}

// ComputeChecksum computes the SHA256 checksum of a file.
func ComputeChecksum(filePath string) (string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("opening file: %w", err)
	}
	defer f.Close()

	return ComputeReaderChecksum(f)
}

// ComputeReaderChecksum computes the SHA256 checksum of data from an io.Reader.
func ComputeReaderChecksum(r io.Reader) (string, error) {
	h := sha256.New()
	if _, err := io.Copy(h, r); err != nil {
		return "", fmt.Errorf("computing hash: %w", err)
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}
