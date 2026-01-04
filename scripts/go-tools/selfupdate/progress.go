package selfupdate

import (
	"fmt"
	"io"
	"strings"
	"sync/atomic"
)

const (
	// progressBarWidth is the width of the progress bar in characters.
	progressBarWidth = 40

	// progressFilledChar is the character used for completed progress.
	progressFilledChar = "█"

	// progressEmptyChar is the character used for remaining progress.
	progressEmptyChar = "░"
)

// ProgressWriter wraps an io.Writer and tracks download progress.
type ProgressWriter struct {
	writer      io.Writer    // Underlying writer for output
	total       int64        // Total expected bytes (0 if unknown)
	current     atomic.Int64 // Current bytes written
	lastPercent int          // Last displayed percentage
	quiet       bool         // If true, don't output progress
}

// ProgressOption is a functional option for configuring ProgressWriter.
type ProgressOption func(*ProgressWriter)

// WithQuiet disables progress output.
func WithQuiet() ProgressOption {
	return func(p *ProgressWriter) {
		p.quiet = true
	}
}

// NewProgressWriter creates a new ProgressWriter that tracks progress.
// total is the expected total bytes (set to 0 if unknown).
// output is where progress updates are written (typically os.Stderr).
func NewProgressWriter(output io.Writer, total int64, opts ...ProgressOption) *ProgressWriter {
	p := &ProgressWriter{
		writer:      output,
		total:       total,
		lastPercent: -1,
	}

	for _, opt := range opts {
		opt(p)
	}

	return p
}

// Write implements io.Writer and updates progress.
func (p *ProgressWriter) Write(data []byte) (int, error) {
	n := len(data)
	newCurrent := p.current.Add(int64(n))

	if !p.quiet && p.total > 0 {
		percent := int(float64(newCurrent) / float64(p.total) * 100)
		if percent != p.lastPercent {
			p.lastPercent = percent
			p.render(newCurrent, percent)
		}
	}

	return n, nil
}

// render displays the progress bar using ANSI escape sequences.
func (p *ProgressWriter) render(current int64, percent int) {
	// Calculate filled/empty portions
	filled := (percent * progressBarWidth) / 100
	if filled > progressBarWidth {
		filled = progressBarWidth
	}
	empty := progressBarWidth - filled

	// Build progress bar
	bar := strings.Repeat(progressFilledChar, filled) + strings.Repeat(progressEmptyChar, empty)

	// Format size display
	currentSize := formatSize(current)
	totalSize := formatSize(p.total)

	// Use carriage return to overwrite line
	fmt.Fprintf(p.writer, "\r[%s] %3d%% (%s / %s)", bar, percent, currentSize, totalSize)
}

// Finish completes the progress bar and moves to a new line.
func (p *ProgressWriter) Finish() {
	if !p.quiet && p.total > 0 {
		current := p.current.Load()
		percent := 100
		if current < p.total {
			percent = int(float64(current) / float64(p.total) * 100)
		}
		p.render(current, percent)
		fmt.Fprintln(p.writer) // New line after completion
	}
}

// Current returns the current number of bytes written.
func (p *ProgressWriter) Current() int64 {
	return p.current.Load()
}

// Total returns the expected total bytes.
func (p *ProgressWriter) Total() int64 {
	return p.total
}

// formatSize formats a byte count into human-readable format (KB, MB, GB).
func formatSize(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/GB)
	case bytes >= MB:
		return fmt.Sprintf("%.2f MB", float64(bytes)/MB)
	case bytes >= KB:
		return fmt.Sprintf("%.2f KB", float64(bytes)/KB)
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}

// ProgressReader wraps an io.Reader and tracks read progress.
// Use this for downloads where you want to track progress as data is read.
type ProgressReader struct {
	reader   io.Reader
	progress *ProgressWriter
}

// NewProgressReader creates a new ProgressReader that tracks read progress.
func NewProgressReader(reader io.Reader, progress *ProgressWriter) *ProgressReader {
	return &ProgressReader{
		reader:   reader,
		progress: progress,
	}
}

// Read implements io.Reader and updates progress.
func (r *ProgressReader) Read(p []byte) (int, error) {
	n, err := r.reader.Read(p)
	if n > 0 {
		//nolint:errcheck // Progress tracking is non-critical
		r.progress.Write(p[:n])
	}
	return n, err
}
