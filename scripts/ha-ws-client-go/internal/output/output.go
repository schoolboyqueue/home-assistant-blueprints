// Package output provides output formatting for the CLI.
package output

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
)

// Format represents the output format mode.
type Format string

const (
	FormatDefault Format = "default"
	FormatCompact Format = "compact"
	FormatJSON    Format = "json"
)

// Config holds output configuration.
type Config struct {
	Format         Format
	ShowTimestamps bool
	ShowHeaders    bool
	ShowAge        bool
	MaxItems       int
}

// DefaultConfig returns the default output configuration.
func DefaultConfig() *Config {
	return &Config{
		Format:         FormatDefault,
		ShowTimestamps: true,
		ShowHeaders:    true,
		ShowAge:        false,
		MaxItems:       0,
	}
}

var (
	globalConfig   = DefaultConfig()
	globalConfigMu sync.RWMutex
)

// GetConfig returns the current output configuration.
func GetConfig() *Config {
	globalConfigMu.RLock()
	defer globalConfigMu.RUnlock()
	return globalConfig
}

// SetConfig sets the output configuration.
func SetConfig(cfg *Config) {
	globalConfigMu.Lock()
	defer globalConfigMu.Unlock()
	globalConfig = cfg
}

// ConfigureFromFlags sets the output configuration from parsed CLI flags.
// This integrates with the cli package's flag parsing.
func ConfigureFromFlags(format string, noHeaders, noTimestamps, showAge bool, maxItems int) {
	globalConfigMu.Lock()
	defer globalConfigMu.Unlock()
	switch format {
	case "json":
		globalConfig.Format = FormatJSON
	case "compact":
		globalConfig.Format = FormatCompact
	default:
		globalConfig.Format = FormatDefault
	}
	globalConfig.ShowHeaders = !noHeaders
	globalConfig.ShowTimestamps = !noTimestamps
	globalConfig.ShowAge = showAge
	globalConfig.MaxItems = maxItems
}

// Result represents a structured result for JSON output.
type Result struct {
	Success bool   `json:"success"`
	Command string `json:"command,omitempty"`
	Data    any    `json:"data,omitempty"`
	Error   string `json:"error,omitempty"`
	Count   int    `json:"count,omitempty"`
	Summary string `json:"summary,omitempty"`
	Message string `json:"message,omitempty"`
}

// IsJSON returns true if JSON output mode is enabled.
func IsJSON() bool {
	globalConfigMu.RLock()
	defer globalConfigMu.RUnlock()
	return globalConfig.Format == FormatJSON
}

// IsCompact returns true if compact output mode is enabled.
func IsCompact() bool {
	globalConfigMu.RLock()
	defer globalConfigMu.RUnlock()
	return globalConfig.Format == FormatCompact
}

// ShowAge returns true if --show-age flag is set.
func ShowAge() bool {
	globalConfigMu.RLock()
	defer globalConfigMu.RUnlock()
	return globalConfig.ShowAge
}

// Data outputs data in the configured format.
func Data(data any, opts ...Option) {
	o := &options{}
	for _, opt := range opts {
		opt(o)
	}

	globalConfigMu.RLock()
	cfg := *globalConfig
	globalConfigMu.RUnlock()

	switch cfg.Format {
	case FormatJSON:
		result := Result{Success: true, Data: data}
		if o.command != "" {
			result.Command = o.command
		}
		if o.count > 0 {
			result.Count = o.count
		}
		if o.summary != "" {
			result.Summary = o.summary
		}
		jsonBytes, _ := json.Marshal(result)
		fmt.Println(string(jsonBytes))
	case FormatCompact:
		if o.summary != "" {
			fmt.Println(o.summary)
		}
		printCompact(data)
	default:
		if o.summary != "" && cfg.ShowHeaders {
			fmt.Println(o.summary)
		}
		printDefault(data)
	}
}

// Message outputs a simple message.
func Message(msg string) {
	globalConfigMu.RLock()
	format := globalConfig.Format
	globalConfigMu.RUnlock()

	switch format {
	case FormatJSON:
		result := Result{Success: true, Message: msg}
		jsonBytes, _ := json.Marshal(result)
		fmt.Println(string(jsonBytes))
	default:
		fmt.Println(msg)
	}
}

// Error outputs an error message.
func Error(err error, code string) {
	msg := err.Error()

	globalConfigMu.RLock()
	format := globalConfig.Format
	globalConfigMu.RUnlock()

	switch format {
	case FormatJSON:
		result := Result{Success: false, Error: msg}
		jsonBytes, _ := json.Marshal(result)
		fmt.Println(string(jsonBytes))
	case FormatCompact:
		if code != "" {
			fmt.Fprintf(os.Stderr, "[%s] %s\n", code, msg)
		} else {
			fmt.Fprintln(os.Stderr, msg)
		}
	default:
		if code != "" {
			fmt.Fprintf(os.Stderr, "Error [%s]: %s\n", code, msg)
		} else {
			fmt.Fprintf(os.Stderr, "Error: %s\n", msg)
		}
	}
}

// List outputs a list of items.
func List[T any](items []T, opts ...ListOption[T]) {
	o := &listOptions[T]{}
	for _, opt := range opts {
		opt(o)
	}

	globalConfigMu.RLock()
	cfg := *globalConfig
	globalConfigMu.RUnlock()

	count := len(items)
	displayItems := items
	if cfg.MaxItems > 0 && count > cfg.MaxItems {
		displayItems = items[:cfg.MaxItems]
	}

	switch cfg.Format {
	case FormatJSON:
		result := Result{Success: true, Data: items, Count: count}
		if o.command != "" {
			result.Command = o.command
		}
		jsonBytes, _ := json.Marshal(result)
		fmt.Println(string(jsonBytes))
	case FormatCompact:
		if o.title != "" && cfg.ShowHeaders {
			fmt.Printf("%s: %d\n", o.title, count)
		}
		for i, item := range displayItems {
			if o.formatter != nil {
				fmt.Println(o.formatter(item, i))
			} else {
				printCompactItem(item)
			}
		}
		if cfg.MaxItems > 0 && count > cfg.MaxItems {
			fmt.Printf("+%d more\n", count-cfg.MaxItems)
		}
	default:
		if o.title != "" && cfg.ShowHeaders {
			fmt.Printf("%s: %d\n\n", o.title, count)
		}
		for i, item := range displayItems {
			if o.formatter != nil {
				fmt.Println(o.formatter(item, i))
			} else {
				printDefaultItem(item)
			}
		}
		if cfg.MaxItems > 0 && count > cfg.MaxItems {
			fmt.Printf("\n... and %d more\n", count-cfg.MaxItems)
		}
	}
}

// Entity outputs entity state data.
func Entity(entityID, state string, attributes map[string]any) {
	globalConfigMu.RLock()
	format := globalConfig.Format
	globalConfigMu.RUnlock()

	switch format {
	case FormatJSON:
		data := map[string]any{
			"entity_id":  entityID,
			"state":      state,
			"attributes": attributes,
		}
		result := Result{Success: true, Data: data}
		jsonBytes, _ := json.Marshal(result)
		fmt.Println(string(jsonBytes))
	case FormatCompact:
		fmt.Printf("%s=%s\n", entityID, state)
	default:
		data := map[string]any{
			"entity_id":  entityID,
			"state":      state,
			"attributes": attributes,
		}
		jsonBytes, _ := json.MarshalIndent(data, "", "  ")
		fmt.Println(string(jsonBytes))
	}
}

// Timeline outputs timeline/history data.
func Timeline[T any](entries []T, opts ...TimelineOption[T]) {
	o := &timelineOptions[T]{}
	for _, opt := range opts {
		opt(o)
	}

	globalConfigMu.RLock()
	cfg := *globalConfig
	globalConfigMu.RUnlock()

	count := len(entries)
	displayEntries := entries
	if cfg.MaxItems > 0 && count > cfg.MaxItems {
		displayEntries = entries[:cfg.MaxItems]
	}

	switch cfg.Format {
	case FormatJSON:
		result := Result{Success: true, Data: entries, Count: count}
		if o.command != "" {
			result.Command = o.command
		}
		jsonBytes, _ := json.Marshal(result)
		fmt.Println(string(jsonBytes))
	case FormatCompact:
		if o.title != "" && cfg.ShowHeaders {
			fmt.Printf("%s: %d\n", o.title, count)
		}
		for _, entry := range displayEntries {
			if o.formatter != nil {
				fmt.Println(o.formatter(entry))
			} else {
				printCompactItem(entry)
			}
		}
		if cfg.MaxItems > 0 && count > cfg.MaxItems {
			fmt.Printf("+%d more\n", count-cfg.MaxItems)
		}
	default:
		if o.title != "" && cfg.ShowHeaders {
			fmt.Printf("%s:\n\n", o.title)
		}
		for _, entry := range displayEntries {
			if o.formatter != nil {
				fmt.Println(o.formatter(entry))
			} else {
				printDefaultItem(entry)
			}
		}
		if cfg.MaxItems > 0 && count > cfg.MaxItems {
			fmt.Printf("\n... and %d more\n", count-cfg.MaxItems)
		}
		if cfg.ShowHeaders {
			fmt.Printf("\nTotal: %d entries\n", count)
		}
	}
}

// Options

type options struct {
	command string
	count   int
	summary string
}

// Option configures output options.
type Option func(*options)

// WithCommand sets the command name for the output.
func WithCommand(cmd string) Option {
	return func(o *options) { o.command = cmd }
}

// WithCount sets the count for the output.
func WithCount(n int) Option {
	return func(o *options) { o.count = n }
}

// WithSummary sets the summary for the output.
func WithSummary(s string) Option {
	return func(o *options) { o.summary = s }
}

type listOptions[T any] struct {
	title     string
	command   string
	formatter func(T, int) string
}

// ListOption configures list output options.
type ListOption[T any] func(*listOptions[T])

// ListTitle sets the title for list output.
func ListTitle[T any](title string) ListOption[T] {
	return func(o *listOptions[T]) { o.title = title }
}

// ListCommand sets the command name for list output.
func ListCommand[T any](cmd string) ListOption[T] {
	return func(o *listOptions[T]) { o.command = cmd }
}

// ListFormatter sets the item formatter for list output.
func ListFormatter[T any](f func(T, int) string) ListOption[T] {
	return func(o *listOptions[T]) { o.formatter = f }
}

type timelineOptions[T any] struct {
	title     string
	command   string
	formatter func(T) string
}

// TimelineOption configures timeline output options.
type TimelineOption[T any] func(*timelineOptions[T])

// TimelineTitle sets the title for timeline output.
func TimelineTitle[T any](title string) TimelineOption[T] {
	return func(o *timelineOptions[T]) { o.title = title }
}

// TimelineCommand sets the command name for timeline output.
func TimelineCommand[T any](cmd string) TimelineOption[T] {
	return func(o *timelineOptions[T]) { o.command = cmd }
}

// TimelineFormatter sets the entry formatter for timeline output.
func TimelineFormatter[T any](f func(T) string) TimelineOption[T] {
	return func(o *timelineOptions[T]) { o.formatter = f }
}

// Helper functions

func printCompact(data any) {
	switch v := data.(type) {
	case []any:
		for _, item := range v {
			printCompactItem(item)
		}
	default:
		printCompactItem(data)
	}
}

func printCompactItem(item any) {
	// First try to convert struct to map using JSON marshaling
	// This handles typed structs like HAState gracefully
	switch v := item.(type) {
	case map[string]any:
		printCompactMap(v)
	default:
		// Try to convert struct to map via JSON
		if m := structToMap(item); m != nil {
			printCompactMap(m)
			return
		}
		fmt.Printf("%v\n", item)
	}
}

// structToMap converts a struct to map[string]any via JSON marshaling.
// Returns nil if the conversion fails or the result is not a map.
func structToMap(item any) map[string]any {
	data, err := json.Marshal(item)
	if err != nil {
		return nil
	}
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		return nil
	}
	return m
}

// printCompactMap handles compact output for map types.
func printCompactMap(v map[string]any) {
	// Entity state
	if entityID, ok := v["entity_id"].(string); ok {
		if state, ok := v["state"].(string); ok {
			fmt.Printf("%s=%s\n", entityID, state)
			return
		}
	}
	// Trace info
	if runID, ok := v["run_id"].(string); ok {
		if itemID, ok := v["item_id"].(string); ok {
			timestamp := ""
			if ts, ok := v["timestamp"].(map[string]any); ok {
				if start, ok := ts["start"].(string); ok {
					timestamp = start
				}
			}
			fmt.Printf("%s %s %s\n", itemID, runID, timestamp)
			return
		}
	}
	// Generic: print key=value pairs
	var pairs []string
	for k, val := range v {
		if val != nil {
			pairs = append(pairs, fmt.Sprintf("%s=%v", k, val))
		}
		if len(pairs) >= 5 {
			break
		}
	}
	fmt.Println(strings.Join(pairs, " "))
}

func printDefault(data any) {
	jsonBytes, _ := json.MarshalIndent(data, "", "  ")
	fmt.Println(string(jsonBytes))
}

func printDefaultItem(item any) {
	jsonBytes, _ := json.MarshalIndent(item, "", "  ")
	fmt.Println(string(jsonBytes))
}

// FormatTime formats a time for display.
func FormatTime(t time.Time) string {
	globalConfigMu.RLock()
	format := globalConfig.Format
	globalConfigMu.RUnlock()

	if format == FormatCompact {
		return t.Format(time.RFC3339)
	}
	return t.Local().Format("2006-01-02 15:04:05")
}
