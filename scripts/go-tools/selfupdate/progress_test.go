package selfupdate

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"
)

func TestNewProgressWriter(t *testing.T) {
	var buf bytes.Buffer
	p := NewProgressWriter(&buf, 1000)

	if p.total != 1000 {
		t.Errorf("total = %d, want 1000", p.total)
	}
	if p.Current() != 0 {
		t.Errorf("Current() = %d, want 0", p.Current())
	}
}

func TestProgressWriter_Write(t *testing.T) {
	var buf bytes.Buffer
	p := NewProgressWriter(&buf, 100)

	// Write 50 bytes
	n, err := p.Write(make([]byte, 50))
	if err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	if n != 50 {
		t.Errorf("Write() returned %d, want 50", n)
	}
	if p.Current() != 50 {
		t.Errorf("Current() = %d, want 50", p.Current())
	}

	// Check that progress was rendered
	output := buf.String()
	if !strings.Contains(output, "50%") {
		t.Errorf("output should contain '50%%', got %q", output)
	}
}

func TestProgressWriter_Quiet(t *testing.T) {
	var buf bytes.Buffer
	p := NewProgressWriter(&buf, 100, WithQuiet())

	_, _ = p.Write(make([]byte, 50))
	p.Finish()

	if buf.Len() != 0 {
		t.Errorf("quiet mode should not produce output, got %q", buf.String())
	}
}

func TestProgressWriter_UnknownTotal(t *testing.T) {
	var buf bytes.Buffer
	p := NewProgressWriter(&buf, 0) // Unknown total

	_, _ = p.Write(make([]byte, 50))
	p.Finish()

	// With unknown total, no percentage is shown
	if buf.Len() != 0 {
		t.Errorf("unknown total should not produce output, got %q", buf.String())
	}
}

func TestProgressWriter_Finish(t *testing.T) {
	var buf bytes.Buffer
	p := NewProgressWriter(&buf, 100)

	_, _ = p.Write(make([]byte, 100))
	p.Finish()

	output := buf.String()
	if !strings.Contains(output, "100%") {
		t.Errorf("finished output should contain '100%%', got %q", output)
	}
	if !strings.HasSuffix(output, "\n") {
		t.Error("finished output should end with newline")
	}
}

func TestProgressWriter_ProgressBar(t *testing.T) {
	var buf bytes.Buffer
	p := NewProgressWriter(&buf, 100)

	// Write 50 bytes
	_, _ = p.Write(make([]byte, 50))

	output := buf.String()

	// Check that the progress bar contains filled and empty characters
	if !strings.Contains(output, progressFilledChar) {
		t.Error("progress bar should contain filled characters")
	}
	if !strings.Contains(output, progressEmptyChar) {
		t.Error("progress bar should contain empty characters at 50%")
	}
}

func TestFormatSize(t *testing.T) {
	tests := []struct {
		bytes int64
		want  string
	}{
		{0, "0 B"},
		{100, "100 B"},
		{1023, "1023 B"},
		{1024, "1.00 KB"},
		{1536, "1.50 KB"},
		{1048576, "1.00 MB"},
		{1572864, "1.50 MB"},
		{1073741824, "1.00 GB"},
		{1610612736, "1.50 GB"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := formatSize(tt.bytes)
			if got != tt.want {
				t.Errorf("formatSize(%d) = %q, want %q", tt.bytes, got, tt.want)
			}
		})
	}
}

func TestProgressReader(t *testing.T) {
	content := strings.Repeat("x", 100)
	reader := strings.NewReader(content)

	var buf bytes.Buffer
	progress := NewProgressWriter(&buf, 100)
	progressReader := NewProgressReader(reader, progress)

	// Read all data
	data, err := io.ReadAll(progressReader)
	if err != nil {
		t.Fatalf("ReadAll() error = %v", err)
	}
	if len(data) != 100 {
		t.Errorf("read %d bytes, want 100", len(data))
	}
	if progress.Current() != 100 {
		t.Errorf("Current() = %d, want 100", progress.Current())
	}
}

func TestProgressReader_PartialReads(t *testing.T) {
	content := strings.Repeat("x", 100)
	reader := strings.NewReader(content)

	var buf bytes.Buffer
	progress := NewProgressWriter(&buf, 100)
	progressReader := NewProgressReader(reader, progress)

	// Read in chunks
	buffer := make([]byte, 25)
	for i := range 4 {
		n, err := progressReader.Read(buffer)
		if err != nil && !errors.Is(err, io.EOF) {
			t.Fatalf("Read() error = %v", err)
		}
		if n != 25 {
			t.Errorf("read %d bytes, want 25", n)
		}
		expected := int64((i + 1) * 25)
		if progress.Current() != expected {
			t.Errorf("after read %d: Current() = %d, want %d", i+1, progress.Current(), expected)
		}
	}
}

func TestProgressWriter_CarriageReturn(t *testing.T) {
	var buf bytes.Buffer
	p := NewProgressWriter(&buf, 100)

	p.Write(make([]byte, 50))

	output := buf.String()
	if !strings.HasPrefix(output, "\r") {
		t.Error("progress output should start with carriage return")
	}
}

func TestProgressWriter_Total(t *testing.T) {
	var buf bytes.Buffer
	p := NewProgressWriter(&buf, 500)

	if p.Total() != 500 {
		t.Errorf("Total() = %d, want 500", p.Total())
	}
}
