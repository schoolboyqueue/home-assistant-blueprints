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
	"github.com/home-assistant-blueprints/ha-ws-client-go/internal/handlers"
	"github.com/home-assistant-blueprints/ha-ws-client-go/internal/output"
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

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		os.Exit(1)
	}
}

// buildCommands creates all CLI commands
func buildCommands() []*cli.Command {
	return []*cli.Command{
		// Basic commands
		{
			Name:      "ping",
			Usage:     "Test connection",
			Action:    wrapHandler(handlers.HandlePing),
			ArgsUsage: "",
		},
		{
			Name:      "state",
			Usage:     "Get single entity state",
			Action:    wrapHandler(handlers.HandleState),
			ArgsUsage: "<entity_id>",
		},
		{
			Name:      "states",
			Usage:     "Get all entity states (summary)",
			Action:    wrapHandler(handlers.HandleStates),
			ArgsUsage: "",
		},
		{
			Name:      "states-json",
			Usage:     "Get all states as JSON array",
			Action:    wrapHandler(handlers.HandleStatesJSON),
			ArgsUsage: "",
		},
		{
			Name:      "states-filter",
			Usage:     "Filter states by entity_id pattern",
			Action:    wrapHandler(handlers.HandleStatesFilter),
			ArgsUsage: "<pattern>",
		},
		{
			Name:      "config",
			Usage:     "Get HA configuration",
			Action:    wrapHandler(handlers.HandleConfig),
			ArgsUsage: "",
		},
		{
			Name:      "services",
			Usage:     "List all services",
			Action:    wrapHandler(handlers.HandleServices),
			ArgsUsage: "",
		},
		{
			Name:      "call",
			Usage:     "Call a service (data as JSON)",
			Action:    wrapHandler(handlers.HandleCall),
			ArgsUsage: "<domain> <service> [data]",
		},
		{
			Name:      "template",
			Usage:     "Render a Jinja template (use - for stdin)",
			Action:    wrapHandler(handlers.HandleTemplate),
			ArgsUsage: "<template>",
		},

		// History commands
		{
			Name:      "logbook",
			Usage:     "Get logbook entries (default 24h)",
			Action:    wrapHandler(handlers.HandleLogbook),
			ArgsUsage: "<entity_id> [hours]",
		},
		{
			Name:      "history",
			Usage:     "Get state history (default 24h)",
			Action:    wrapHandler(handlers.HandleHistory),
			ArgsUsage: "<entity_id> [hours]",
		},
		{
			Name:      "history-full",
			Usage:     "Get history with full attributes",
			Action:    wrapHandler(handlers.HandleHistoryFull),
			ArgsUsage: "<entity_id> [hours]",
		},
		{
			Name:      "attrs",
			Usage:     "Attribute change history (compact)",
			Action:    wrapHandler(handlers.HandleAttrs),
			ArgsUsage: "<entity_id> [hours]",
		},
		{
			Name:      "timeline",
			Usage:     "Multi-entity chronological timeline",
			Action:    wrapHandler(handlers.HandleTimeline),
			ArgsUsage: "<hours> <entity>...",
		},
		{
			Name:      "syslog",
			Usage:     "Get system log errors/warnings",
			Action:    wrapHandler(handlers.HandleSyslog),
			ArgsUsage: "",
		},
		{
			Name:      "stats",
			Usage:     "Get sensor statistics (default 24h)",
			Action:    wrapHandler(handlers.HandleStats),
			ArgsUsage: "<entity_id> [hours]",
		},
		{
			Name:      "stats-multi",
			Usage:     "Get statistics for multiple entities",
			Action:    wrapHandler(handlers.HandleStatsMulti),
			ArgsUsage: "<entity>... [hours]",
		},
		{
			Name:      "context",
			Usage:     "Look up what triggered a state change",
			Action:    wrapHandler(handlers.HandleContext),
			ArgsUsage: "<context_id>",
		},
		{
			Name:      "watch",
			Usage:     "Live subscribe to state changes (default 60s)",
			Action:    wrapHandler(handlers.HandleWatch),
			ArgsUsage: "<entity_id> [seconds]",
		},

		// Registry commands
		{
			Name:      "entities",
			Usage:     "List/search entity registry",
			Action:    wrapHandler(handlers.HandleEntities),
			ArgsUsage: "[pattern]",
		},
		{
			Name:      "devices",
			Usage:     "List/search device registry",
			Action:    wrapHandler(handlers.HandleDevices),
			ArgsUsage: "[pattern]",
		},
		{
			Name:      "areas",
			Usage:     "List all areas",
			Action:    wrapHandler(handlers.HandleAreas),
			ArgsUsage: "",
		},

		// Automation debugging commands
		{
			Name:      "traces",
			Usage:     "List automation traces",
			Action:    wrapHandler(handlers.HandleTraces),
			ArgsUsage: "[automation_id]",
		},
		{
			Name:      "trace",
			Usage:     "Get detailed trace for a run",
			Action:    wrapHandler(handlers.HandleTrace),
			ArgsUsage: "<automation_id> <run_id>",
		},
		{
			Name:      "trace-latest",
			Usage:     "Get the most recent trace",
			Action:    wrapHandler(handlers.HandleTraceLatest),
			ArgsUsage: "<automation_id>",
		},
		{
			Name:      "trace-summary",
			Usage:     "Quick overview of recent runs",
			Action:    wrapHandler(handlers.HandleTraceSummary),
			ArgsUsage: "<automation_id>",
		},
		{
			Name:      "trace-vars",
			Usage:     "Show evaluated variables from trace",
			Action:    wrapHandler(handlers.HandleTraceVars),
			ArgsUsage: "<automation_id> <run_id>",
		},
		{
			Name:      "trace-timeline",
			Usage:     "Step-by-step execution timeline",
			Action:    wrapHandler(handlers.HandleTraceTimeline),
			ArgsUsage: "<automation_id> <run_id>",
		},
		{
			Name:      "trace-trigger",
			Usage:     "Show trigger context details",
			Action:    wrapHandler(handlers.HandleTraceTrigger),
			ArgsUsage: "<automation_id> <run_id>",
		},
		{
			Name:      "trace-actions",
			Usage:     "Show action results",
			Action:    wrapHandler(handlers.HandleTraceActions),
			ArgsUsage: "<automation_id> <run_id>",
		},
		{
			Name:      "trace-debug",
			Usage:     "Comprehensive debug view (all info)",
			Action:    wrapHandler(handlers.HandleTraceDebug),
			ArgsUsage: "<automation_id> <run_id>",
		},
		{
			Name:      "automation-config",
			Usage:     "Get automation configuration",
			Action:    wrapHandler(handlers.HandleAutomationConfig),
			ArgsUsage: "<entity_id>",
		},
		{
			Name:      "blueprint-inputs",
			Usage:     "Validate blueprint inputs vs expected",
			Action:    wrapHandler(handlers.HandleBlueprintInputs),
			ArgsUsage: "<entity_id>",
		},

		// Monitoring commands
		{
			Name:      "monitor",
			Usage:     "Monitor entity state changes",
			Action:    wrapHandler(handlers.HandleMonitor),
			ArgsUsage: "<entity_id> [seconds]",
		},
		{
			Name:      "monitor-multi",
			Usage:     "Monitor multiple entities",
			Action:    wrapHandler(handlers.HandleMonitorMulti),
			ArgsUsage: "<entity>...",
		},
		{
			Name:      "analyze",
			Usage:     "Analyze entity state patterns",
			Action:    wrapHandler(handlers.HandleAnalyze),
			ArgsUsage: "<entity_id>",
		},

		// Diagnostic commands
		{
			Name:      "device-health",
			Usage:     "Check if device is responsive",
			Action:    wrapHandler(handlers.HandleDeviceHealth),
			ArgsUsage: "<entity_id>",
		},
		{
			Name:      "compare",
			Usage:     "Side-by-side entity comparison",
			Action:    wrapHandler(handlers.HandleCompare),
			ArgsUsage: "<entity1> <entity2>",
		},
	}
}

// wrapHandler wraps a handlers.Context function into a cli.ActionFunc
func wrapHandler(handler func(*handlers.Context) error) cli.ActionFunc {
	return func(_ context.Context, cmd *cli.Command) error {
		// Parse time flags from root command
		var fromTime, toTime *time.Time
		if fromStr := cmd.String("from"); fromStr != "" {
			t, err := hacli.ParseFlexibleDate(fromStr)
			if err != nil {
				return fmt.Errorf("invalid --from value: %w", err)
			}
			fromTime = &t
		}
		if toStr := cmd.String("to"); toStr != "" {
			t, err := hacli.ParseFlexibleDate(toStr)
			if err != nil {
				return fmt.Errorf("invalid --to value: %w", err)
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

		// Connect to WebSocket
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

		// Create client
		haClient := client.New(conn)

		// Build args array with command name first (for backward compatibility with handlers)
		args := append([]string{cmd.Name}, cmd.Args().Slice()...)

		// Create handler context
		handlerCtx := &handlers.Context{
			Client:   haClient,
			Args:     args,
			FromTime: fromTime,
			ToTime:   toTime,
		}

		// Execute handler
		if err := handler(handlerCtx); err != nil {
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
		return fmt.Errorf("failed to read auth_required: %w", err)
	}

	if msg.Type != "auth_required" {
		return fmt.Errorf("unexpected message type: %s", msg.Type)
	}

	// Send auth message
	authMsg := map[string]string{
		"type":         "auth",
		"access_token": token,
	}

	if err := conn.WriteJSON(authMsg); err != nil {
		return fmt.Errorf("failed to send auth: %w", err)
	}

	// Read auth result
	var authResult struct {
		Type    string `json:"type"`
		Message string `json:"message,omitempty"`
	}

	if err := conn.ReadJSON(&authResult); err != nil {
		return fmt.Errorf("failed to read auth result: %w", err)
	}

	if authResult.Type != "auth_ok" {
		return fmt.Errorf("authentication failed: %s", authResult.Message)
	}

	return nil
}
