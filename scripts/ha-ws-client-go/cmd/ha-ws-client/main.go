// Package main is the entry point for the ha-ws-client CLI.
package main

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gorilla/websocket"

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

var commandRegistry = map[string]func(*handlers.Context) error{
	// Basic commands
	"ping":          handlers.HandlePing,
	"state":         handlers.HandleState,
	"states":        handlers.HandleStates,
	"states-json":   handlers.HandleStatesJSON,
	"states-filter": handlers.HandleStatesFilter,
	"config":        handlers.HandleConfig,
	"services":      handlers.HandleServices,
	"call":          handlers.HandleCall,
	"template":      handlers.HandleTemplate,

	// History commands
	"logbook":      handlers.HandleLogbook,
	"history":      handlers.HandleHistory,
	"history-full": handlers.HandleHistoryFull,
	"attrs":        handlers.HandleAttrs,
	"timeline":     handlers.HandleTimeline,
	"syslog":       handlers.HandleSyslog,
	"stats":        handlers.HandleStats,
	"context":      handlers.HandleContext,
	"watch":        handlers.HandleWatch,

	// Registry commands
	"entities": handlers.HandleEntities,
	"devices":  handlers.HandleDevices,
	"areas":    handlers.HandleAreas,

	// Automation commands
	"traces":            handlers.HandleTraces,
	"trace":             handlers.HandleTrace,
	"trace-latest":      handlers.HandleTraceLatest,
	"trace-summary":     handlers.HandleTraceSummary,
	"trace-vars":        handlers.HandleTraceVars,
	"trace-timeline":    handlers.HandleTraceTimeline,
	"trace-trigger":     handlers.HandleTraceTrigger,
	"trace-actions":     handlers.HandleTraceActions,
	"trace-debug":       handlers.HandleTraceDebug,
	"automation-config": handlers.HandleAutomationConfig,
	"blueprint-inputs":  handlers.HandleBlueprintInputs,

	// Monitor commands
	"monitor":       handlers.HandleMonitor,
	"monitor-multi": handlers.HandleMonitorMulti,
	"analyze":       handlers.HandleAnalyze,

	// Diagnostic commands
	"device-health": handlers.HandleDeviceHealth,
	"compare":       handlers.HandleCompare,
}

func showHelp() {
	fmt.Print(`Usage: ha-ws-client <command> [args...] [--from "TIME"] [--to "TIME"]

Commands:
  state <entity_id>              - Get single entity state
  states                         - Get all entity states (summary)
  states-json                    - Get all states as JSON array
  states-filter <pattern>        - Filter states by entity_id pattern (--show-age)
  config                         - Get HA configuration
  services                       - List all services
  call <domain> <service> [data] - Call a service (data as JSON)
  template <template>            - Render a Jinja template (use - for stdin)
  ping                           - Test connection

Log Commands:
  logbook <entity_id> [hours]    - Get logbook entries (default 24h)
  history <entity_id> [hours]    - Get state history (default 24h)
  history-full <entity_id> [hours] - Get history with full attributes
  attrs <entity_id> [hours]      - Attribute change history (compact)
  timeline <hours> <entity>...   - Multi-entity chronological timeline
  syslog                         - Get system log errors/warnings
  stats <entity_id> [hours]      - Get sensor statistics (default 24h)
  context <context_id>           - Look up what triggered a state change
  watch <entity_id> [seconds]    - Live subscribe to state changes (default 60s)

Registry Commands:
  entities [pattern]             - List/search entity registry
  devices [pattern]              - List/search device registry
  areas                          - List all areas

Automation Debugging:
  traces [automation_id]         - List automation traces (supports --from)
  trace <automation_id> <run_id> - Get detailed trace for a run
  trace-latest <automation_id>   - Get the most recent trace
  trace-summary <automation_id>  - Quick overview of recent runs
  trace-vars <auto_id> <run_id>  - Show evaluated variables from trace
  trace-timeline <id> <run_id>   - Step-by-step execution timeline
  trace-trigger <id> <run_id>    - Show trigger context details
  trace-actions <id> <run_id>    - Show action results
  trace-debug <id> <run_id>      - Comprehensive debug view (all info)
  automation-config <entity_id>  - Get automation configuration
  blueprint-inputs <entity_id>   - Validate blueprint inputs vs expected

Monitoring Commands:
  monitor <entity_id> [seconds]  - Monitor entity state changes
  monitor-multi <entity>...      - Monitor multiple entities
  analyze <entity_id>            - Analyze entity state patterns

Diagnostic Commands:
  device-health <entity_id>      - Check if device is responsive (stale detection)
  compare <entity1> <entity2>    - Side-by-side entity comparison

Time Filtering Options (for logbook, history, history-full, attrs, timeline, traces):
  --from "YYYY-MM-DD HH:MM"      - Start time (instead of hours ago)
  --to "YYYY-MM-DD HH:MM"        - End time (default: now)

Output Format Options (for AI agent context efficiency):
  --output=json                  - Machine-readable JSON (most context-efficient)
  --output=compact               - Reduced verbosity, single-line entries
  --output=default               - Human-readable formatted output
  --json                         - Shorthand for --output=json
  --compact                      - Shorthand for --output=compact
  --no-headers                   - Hide section headers/titles
  --no-timestamps                - Hide timestamps in output
  --max-items=N                  - Limit output to N items
  --show-age                     - Show last_updated age (states-filter)

Global Options:
  --help, -h, help               - Show this help message
  --version, -v, version         - Show version information

Examples:
  ha-ws-client state sun.sun
  ha-ws-client call light turn_on '{"entity_id":"light.kitchen"}'
  ha-ws-client attrs light.kitchen 4
  ha-ws-client watch binary_sensor.motion 30
  ha-ws-client states-filter "cover.*" --show-age --compact
  ha-ws-client trace-latest automation.bathroom_lights
  ha-ws-client trace-summary automation.adaptive_shades
  ha-ws-client traces automation.kitchen --from "2024-01-01"
  ha-ws-client device-health cover.guest_bedroom_shade
  ha-ws-client compare cover.living_room_shade cover.guest_bedroom_shade
  echo "{{ now() }}" | ha-ws-client template -
`)
}

func parseTimeArgs(args []string) (filtered []string, fromTime, toTime *time.Time) {
	i := 0
	for i < len(args) {
		arg := args[i]
		if arg == "--from" && i+1 < len(args) {
			t, err := parseFlexibleDate(args[i+1])
			if err == nil {
				fromTime = &t
			}
			i += 2
			continue
		}
		if arg == "--to" && i+1 < len(args) {
			t, err := parseFlexibleDate(args[i+1])
			if err == nil {
				toTime = &t
			}
			i += 2
			continue
		}
		if after, ok := strings.CutPrefix(arg, "--from="); ok {
			t, err := parseFlexibleDate(after)
			if err == nil {
				fromTime = &t
			}
			i++
			continue
		}
		if after, ok := strings.CutPrefix(arg, "--to="); ok {
			t, err := parseFlexibleDate(after)
			if err == nil {
				toTime = &t
			}
			i++
			continue
		}
		filtered = append(filtered, arg)
		i++
	}

	return filtered, fromTime, toTime
}

func parseFlexibleDate(s string) (time.Time, error) {
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

func isHelpRequested(args []string) bool {
	for _, arg := range args {
		if arg == "--help" || arg == "-h" || arg == "help" {
			return true
		}
	}
	return false
}

func isVersionRequested(args []string) bool {
	for _, arg := range args {
		if arg == "--version" || arg == "-v" || arg == "version" {
			return true
		}
	}
	return false
}

func showVersion() {
	fmt.Printf("ha-ws-client %s\n", Version)
	fmt.Printf("  Build time: %s\n", BuildTime)
	fmt.Printf("  Git commit: %s\n", GitCommit)
}

func main() {
	os.Exit(run())
}

func run() int {
	rawArgs := os.Args[1:]

	// Check for version first
	if isVersionRequested(rawArgs) {
		showVersion()
		return 0
	}

	// Check for help
	if isHelpRequested(rawArgs) || len(rawArgs) == 0 {
		showHelp()
		return 0
	}

	// Parse output args first
	argsAfterOutput := output.ParseArgs(rawArgs)

	// Parse time args
	filteredArgs, fromTime, toTime := parseTimeArgs(argsAfterOutput)

	if len(filteredArgs) == 0 {
		showHelp()
		return 0
	}

	command := filteredArgs[0]

	// Get supervisor token
	token := os.Getenv("SUPERVISOR_TOKEN")
	if token == "" {
		fmt.Fprintln(os.Stderr, "Error: SUPERVISOR_TOKEN environment variable not set")
		return 1
	}

	// Connect to WebSocket
	header := http.Header{}
	header.Set("Authorization", "Bearer "+token)

	dialer := websocket.Dialer{
		HandshakeTimeout: defaultTimeout,
	}

	conn, _, err := dialer.Dial(wsURL, header)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to connect to Home Assistant: %v\n", err)
		return 1
	}
	defer conn.Close()

	// Authenticate
	if err := authenticate(conn, token); err != nil {
		fmt.Fprintf(os.Stderr, "Authentication failed: %v\n", err)
		return 1
	}

	// Create client
	c := client.New(conn)

	// Create handler context
	ctx := &handlers.Context{
		Client:   c,
		Args:     filteredArgs,
		FromTime: fromTime,
		ToTime:   toTime,
	}

	// Look up and execute command
	handler, ok := commandRegistry[command]
	if !ok {
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", command)
		return 1
	}

	if err := handler(ctx); err != nil {
		var clientErr *client.HAClientError
		if errors.As(err, &clientErr) {
			output.Error(err, clientErr.Code)
		} else {
			output.Error(err, "")
		}
		return 1
	}

	return 0
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
