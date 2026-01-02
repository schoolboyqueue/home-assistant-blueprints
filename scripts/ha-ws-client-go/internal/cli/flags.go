// Package cli provides argument parsing utilities for CLI tools using the standard flag package.
package cli

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/home-assistant-blueprints/ha-ws-client-go/internal/types"
)

// FlagSet creates a new flag.FlagSet configured with standard CLI flags.
// The name parameter is used for error messages and usage output.
func FlagSet(name string) *flag.FlagSet {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(io.Discard) // We handle errors ourselves
	return fs
}

// AddStandardFlags adds help and version flags to the FlagSet.
func AddStandardFlags(fs *flag.FlagSet, f *types.CLIFlags) {
	fs.BoolVar(&f.Help, "help", false, "Show help message")
	fs.BoolVar(&f.Help, "h", false, "Show help message (shorthand)")
	fs.BoolVar(&f.Version, "version", false, "Show version information")
	fs.BoolVar(&f.Version, "v", false, "Show version information (shorthand)")
}

// AddTimeFlags adds --from and --to flags to the FlagSet.
func AddTimeFlags(fs *flag.FlagSet, f *types.CLIFlags) {
	fs.StringVar(&f.From, "from", "", "Start time (YYYY-MM-DD or YYYY-MM-DD HH:MM)")
	fs.StringVar(&f.To, "to", "", "End time (YYYY-MM-DD or YYYY-MM-DD HH:MM)")
}

// AddOutputFlags adds output formatting flags to the FlagSet.
func AddOutputFlags(fs *flag.FlagSet, f *types.CLIFlags) {
	fs.StringVar(&f.Output, "output", "default", "Output format: json, compact, default")
	fs.StringVar(&f.Output, "format", "default", "Output format (alias for --output)")
	fs.BoolVar(&f.Compact, "compact", false, "Use compact output format")
	fs.BoolVar(&f.JSON, "json", false, "Use JSON output format")
	fs.BoolVar(&f.NoHeaders, "no-headers", false, "Hide section headers/titles")
	fs.BoolVar(&f.NoTimestamps, "no-timestamps", false, "Hide timestamps in output")
	fs.BoolVar(&f.ShowAge, "show-age", false, "Show last_updated age")
	fs.IntVar(&f.MaxItems, "max-items", 0, "Limit output to N items")
}

// AddAllFlags adds all standard flags to the FlagSet.
func AddAllFlags(fs *flag.FlagSet, f *types.CLIFlags) {
	AddStandardFlags(fs, f)
	AddTimeFlags(fs, f)
	AddOutputFlags(fs, f)
}

// Parse parses command-line arguments and returns the parsed CLIFlags.
// It handles both flags and positional arguments (commands).
func Parse(args []string) (*types.CLIFlags, error) {
	f := &types.CLIFlags{}
	fs := FlagSet(os.Args[0])
	AddAllFlags(fs, f)

	// Check for bare "help" or "version" commands first
	if len(args) > 0 {
		switch args[0] {
		case "help":
			f.Help = true
			f.Args = args[1:]
			return f, nil
		case "version":
			f.Version = true
			f.Args = args[1:]
			return f, nil
		}
	}

	// Parse the flags
	if err := fs.Parse(args); err != nil {
		return nil, err
	}

	// Get remaining positional arguments
	f.Args = fs.Args()

	// Parse time values if provided
	if f.From != "" {
		t, err := ParseFlexibleDate(f.From)
		if err != nil {
			return nil, fmt.Errorf("invalid --from value: %w", err)
		}
		f.FromTime = &t
	}

	if f.To != "" {
		t, err := ParseFlexibleDate(f.To)
		if err != nil {
			return nil, fmt.Errorf("invalid --to value: %w", err)
		}
		f.ToTime = &t
	}

	return f, nil
}
