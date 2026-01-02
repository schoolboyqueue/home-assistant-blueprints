package cli

import (
	"fmt"
	"time"
)

// ParseFlexibleDate parses a date/time string in various common formats.
// Supported formats:
//   - RFC3339 (e.g., "2024-01-02T15:04:05Z07:00")
//   - "2006-01-02T15:04:05"
//   - "2006-01-02 15:04:05"
//   - "2006-01-02 15:04"
//   - "2006-01-02" (date only)
func ParseFlexibleDate(s string) (time.Time, error) {
	formats := []string{
		time.RFC3339,
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
		"2006-01-02 15:04",
		"2006-01-02",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, s); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse date: %s", s)
}
