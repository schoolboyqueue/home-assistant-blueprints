// Package main is the entry point for the ha-ws-client CLI.
package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/websocket"
	"github.com/urfave/cli/v3"

	hacli "github.com/home-assistant-blueprints/ha-ws-client-go/internal/cli"
	"github.com/home-assistant-blueprints/ha-ws-client-go/internal/client"
	errs "github.com/home-assistant-blueprints/ha-ws-client-go/internal/errors"
	"github.com/home-assistant-blueprints/ha-ws-client-go/internal/handlers"
	"github.com/home-assistant-blueprints/ha-ws-client-go/internal/output"
	"github.com/home-assistant-blueprints/ha-ws-client-go/internal/shutdown"
	"github.com/home-assistant-blueprints/selfupdate"
)

// Version information - set by ldflags during build
var (
	Version   = "dev"
	BuildTime = "unknown"
	GitCommit = "unknown"
)

const (
	wsURL          = "ws://supervisor/core/api/websocket"
	defaultTimeout = 30 * time.Second
)

func main() {
	// Create shutdown coordinator with signal handling
	coord, ctx := shutdown.New(
		shutdown.WithGracePeriod(5*time.Second),
		shutdown.WithOnShutdown(func(reason string) {
			output.Message(fmt.Sprintf("\nShutting down: %s", reason))
		}),
		shutdown.WithOnCleanupTimeout(func() {
			output.Message("Warning: cleanup timed out, some operations may not have completed")
		}),
	)
	coord.HandleSignals()

	cmd := &cli.Command{
		Name:    "ha-ws-client",
		Usage:   "Home Assistant WebSocket API client",
		Version: Version,
		Flags: []cli.Flag{
			// Time filtering flags
			&cli.StringFlag{
				Name:  "from",
				Usage: "Start time for filtering (YYYY-MM-DD or YYYY-MM-DD HH:MM)",
			},
			&cli.StringFlag{
				Name:  "to",
				Usage: "End time for filtering (YYYY-MM-DD or YYYY-MM-DD HH:MM)",
			},
			// Output format flags
			&cli.StringFlag{
				Name:    "output",
				Aliases: []string{"format"},
				Value:   "default",
				Usage:   "Output format: json, compact, or default",
			},
			&cli.BoolFlag{
				Name:  "compact",
				Usage: "Use compact output format (single-line entries)",
			},
			&cli.BoolFlag{
				Name:  "json",
				Usage: "Use JSON output format (machine-readable)",
			},
			&cli.BoolFlag{
				Name:  "no-headers",
				Usage: "Hide section headers and titles",
			},
			&cli.BoolFlag{
				Name:  "no-timestamps",
				Usage: "Hide timestamps in output",
			},
			&cli.BoolFlag{
				Name:  "show-age",
				Usage: "Show last_updated age for states-filter command",
			},
			&cli.IntFlag{
				Name:  "max-items",
				Usage: "Limit output to N items (0 = unlimited)",
			},
		},
		Commands: buildCommands(),
	}

	if err := cmd.Run(ctx, os.Args); err != nil {
		// Check if it was an interrupt
		if coord.IsShuttingDown() {
			os.Exit(130) // Standard exit code for SIGINT
		}
		os.Exit(1)
	}
}

// buildCommands creates all CLI commands from the handler registry.
// Commands are auto-registered via init() functions in each handler file.
func buildCommands() []*cli.Command {
	// Get all registered commands from the handlers package
	registeredCmds := handlers.GetAllCommands()
	commands := make([]*cli.Command, 0, len(registeredCmds)+1) // +1 for update command

	for _, cmd := range registeredCmds {
		commands = append(commands, &cli.Command{
			Name:      cmd.Name,
			Usage:     cmd.Usage,
			ArgsUsage: cmd.ArgsUsage,
			Category:  cmd.Category,
			Action:    wrapHandler(cmd.Handler),
		})
	}

	// Add update command (bypasses WebSocket handler wrapper)
	commands = append(commands, buildUpdateCommand())

	return commands
}

// buildUpdateCommand creates the update command that doesn't require WebSocket connection.
func buildUpdateCommand() *cli.Command {
	return &cli.Command{
		Name:     "update",
		Usage:    "Check for and install updates",
		Category: "Utility",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "check",
				Aliases: []string{"c"},
				Usage:   "Only check for updates, don't install",
			},
			&cli.StringFlag{
				Name:    "version",
				Aliases: []string{"v"},
				Usage:   "Install a specific version (e.g., 1.6.0)",
			},
		},
		Action: runUpdate,
	}
}

// runUpdate handles the update command (bypasses WebSocket connection).
func runUpdate(_ context.Context, cmd *cli.Command) error {
	checkOnly := cmd.Bool("check")
	targetVersion := cmd.String("version")

	updater, err := selfupdate.NewUpdater(
		"ha-ws-client",
		"ha-ws-client-go",
		Version,
		selfupdate.WithOutput(os.Stderr),
	)
	if err != nil {
		return cli.Exit(fmt.Sprintf("Error: %v", err), 1)
	}

	if checkOnly {
		return runUpdateCheck(updater)
	}

	if targetVersion != "" {
		return runUpdateToVersion(updater, targetVersion)
	}

	return runUpdateInstall(updater)
}

// runUpdateCheck checks for updates and displays version information.
func runUpdateCheck(updater *selfupdate.Updater) error {
	result, err := updater.Check()
	if err != nil {
		// Handle specific error types with user-friendly messages
		var rateLimitErr *selfupdate.RateLimitError
		if errors.As(err, &rateLimitErr) {
			return cli.Exit(fmt.Sprintf("Error: %v\nPlease wait and try again later.", rateLimitErr), 1)
		}
		return cli.Exit(fmt.Sprintf("Error checking for updates: %v", err), 1)
	}

	fmt.Printf("Current version: %s\n", result.CurrentVersion)
	fmt.Printf("Latest version:  %s\n", result.LatestVersion)

	if result.UpdateAvailable {
		fmt.Printf("\nA new version is available! Run 'ha-ws-client update' to install.\n")
	} else {
		fmt.Printf("\nYou are running the latest version.\n")
	}

	return nil
}

// runUpdateToVersion downloads and installs a specific version.
func runUpdateToVersion(updater *selfupdate.Updater, version string) error {
	// Check if version exists and show warning for downgrades
	result, err := updater.Check()
	if err != nil {
		var rateLimitErr *selfupdate.RateLimitError
		if errors.As(err, &rateLimitErr) {
			return cli.Exit(fmt.Sprintf("Error: %v\nPlease wait and try again later.", rateLimitErr), 1)
		}
		return cli.Exit(fmt.Sprintf("Error checking for updates: %v", err), 1)
	}

	// Warn if downgrading
	if result.CurrentVersion != "dev" && version < result.CurrentVersion {
		fmt.Printf("Warning: Installing older version %s (current: %s)\n", version, result.CurrentVersion)
	}

	fmt.Printf("Installing version %s...\n", version)

	if err := updater.UpdateToVersion(version); err != nil {
		if errors.Is(err, selfupdate.ErrVersionNotFound) {
			// List available versions to help user
			versions, listErr := updater.ListAvailableVersions()
			if listErr == nil && len(versions) > 0 {
				fmt.Printf("\nAvailable versions:\n")
				for _, v := range versions {
					fmt.Printf("  - %s\n", v)
				}
			}
			return cli.Exit(fmt.Sprintf("Error: Version %s not found.", version), 1)
		}
		if errors.Is(err, selfupdate.ErrPermissionDenied) {
			return cli.Exit("Error: Permission denied. Try running with elevated privileges.", 1)
		}
		if errors.Is(err, selfupdate.ErrChecksumMismatch) {
			return cli.Exit("Error: Downloaded file verification failed. Please try again.", 1)
		}
		return cli.Exit(fmt.Sprintf("Error updating: %v", err), 1)
	}

	fmt.Printf("\nSuccessfully installed version %s!\n", version)
	return nil
}

// runUpdateInstall downloads and installs the latest version.
func runUpdateInstall(updater *selfupdate.Updater) error {
	// First check if update is available
	result, err := updater.Check()
	if err != nil {
		var rateLimitErr *selfupdate.RateLimitError
		if errors.As(err, &rateLimitErr) {
			return cli.Exit(fmt.Sprintf("Error: %v\nPlease wait and try again later.", rateLimitErr), 1)
		}
		return cli.Exit(fmt.Sprintf("Error checking for updates: %v", err), 1)
	}

	if !result.UpdateAvailable {
		fmt.Printf("Already at latest version (%s).\n", result.CurrentVersion)
		return nil
	}

	fmt.Printf("Updating from %s to %s...\n", result.CurrentVersion, result.LatestVersion)

	if err := updater.Update(); err != nil {
		// Handle specific error types
		if errors.Is(err, selfupdate.ErrAlreadyLatest) {
			fmt.Printf("Already at latest version (%s).\n", result.CurrentVersion)
			return nil
		}
		if errors.Is(err, selfupdate.ErrPermissionDenied) {
			return cli.Exit("Error: Permission denied. Try running with elevated privileges.", 1)
		}
		if errors.Is(err, selfupdate.ErrChecksumMismatch) {
			return cli.Exit("Error: Downloaded file verification failed. Please try again.", 1)
		}
		return cli.Exit(fmt.Sprintf("Error updating: %v", err), 1)
	}

	fmt.Printf("\nSuccessfully updated to version %s!\n", result.LatestVersion)
	return nil
}

// wrapHandler wraps a handlers.Context function into a cli.ActionFunc
func wrapHandler(handler func(*handlers.Context) error) cli.ActionFunc {
	return func(ctx context.Context, cmd *cli.Command) error {
		// Parse time flags from root command
		var fromTime, toTime *time.Time
		if fromStr := cmd.String("from"); fromStr != "" {
			t, err := hacli.ParseFlexibleDate(fromStr)
			if err != nil {
				return errs.Wrap(errs.ErrorTypeValidation, err, "invalid --from value")
			}
			fromTime = &t
		}
		if toStr := cmd.String("to"); toStr != "" {
			t, err := hacli.ParseFlexibleDate(toStr)
			if err != nil {
				return errs.Wrap(errs.ErrorTypeValidation, err, "invalid --to value")
			}
			toTime = &t
		}

		// Determine output format
		outputFormat := cmd.String("output")
		if cmd.Bool("json") {
			outputFormat = "json"
		} else if cmd.Bool("compact") {
			outputFormat = "compact"
		}

		// Configure output settings
		output.ConfigureFromFlags(
			outputFormat,
			cmd.Bool("no-headers"),
			cmd.Bool("no-timestamps"),
			cmd.Bool("show-age"),
			cmd.Int("max-items"),
		)

		// Get supervisor token
		token := os.Getenv("SUPERVISOR_TOKEN")
		if token == "" {
			return cli.Exit("Error: SUPERVISOR_TOKEN environment variable not set", 1)
		}

		// Connect to WebSocket with context for cancellation
		header := http.Header{}
		header.Set("Authorization", "Bearer "+token)

		dialer := websocket.Dialer{
			HandshakeTimeout: defaultTimeout,
		}

		conn, _, err := dialer.Dial(wsURL, header)
		if err != nil {
			return cli.Exit(fmt.Sprintf("Failed to connect to Home Assistant: %v", err), 1)
		}
		defer conn.Close()

		// Authenticate
		if err := authenticate(conn, token); err != nil {
			return cli.Exit(fmt.Sprintf("Authentication failed: %v", err), 1)
		}

		// Create client with context for graceful shutdown
		haClient := client.NewWithContext(ctx, conn)

		// Build args array with command name first (for backward compatibility with handlers)
		args := append([]string{cmd.Name}, cmd.Args().Slice()...)

		// Create handler context with context.Context for cancellation
		handlerCtx := &handlers.Context{
			Ctx:      ctx,
			Client:   haClient,
			Args:     args,
			FromTime: fromTime,
			ToTime:   toTime,
		}

		// Execute handler
		if err := handler(handlerCtx); err != nil {
			// Check if it was a context cancellation (graceful shutdown)
			if ctx.Err() != nil {
				return ctx.Err()
			}
			var clientErr *client.HAClientError
			if errors.As(err, &clientErr) {
				output.Error(err, clientErr.Code)
			} else {
				output.Error(err, "")
			}
			return cli.Exit("", 1)
		}

		return nil
	}
}

func authenticate(conn *websocket.Conn, token string) error {
	// Read auth_required message
	var msg struct {
		Type string `json:"type"`
	}

	if err := conn.ReadJSON(&msg); err != nil {
		return errs.Wrap(errs.ErrorTypeNetwork, err, "failed to read auth_required")
	}

	if msg.Type != "auth_required" {
		return errs.Newf(errs.ErrorTypeAuth, "unexpected message type: %s", msg.Type)
	}

	// Send auth message
	authMsg := map[string]string{
		"type":         "auth",
		"access_token": token,
	}

	if err := conn.WriteJSON(authMsg); err != nil {
		return errs.Wrap(errs.ErrorTypeNetwork, err, "failed to send auth")
	}

	// Read auth result
	var authResult struct {
		Type    string `json:"type"`
		Message string `json:"message,omitempty"`
	}

	if err := conn.ReadJSON(&authResult); err != nil {
		return errs.Wrap(errs.ErrorTypeNetwork, err, "failed to read auth result")
	}

	if authResult.Type != "auth_ok" {
		return errs.ErrAuthFailed(authResult.Message)
	}

	return nil
}
