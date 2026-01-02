# CLAUDE.md

## Overview

Go CLI for Home Assistant WebSocket API. Runs in HA add-on environment or on Raspberry Pi. Commands for entity states, services, history, and automation debugging.

## Architecture

```
ha-ws-client-go/
├── cmd/ha-ws-client/
│   └── main.go           # Entry point - CLI parsing, WebSocket, routing
├── internal/
│   ├── client/
│   │   └── client.go     # WebSocket request/response, message IDs
│   ├── handlers/
│   │   ├── basic.go      # ping, state, states, config, services, call, template
│   │   ├── history.go    # logbook, history, attrs, timeline, stats, syslog
│   │   ├── registry.go   # entities, devices, areas
│   │   ├── automation.go # traces, trace, trace-vars, trace-timeline, trace-debug, etc.
│   │   └── monitor.go    # watch, monitor, monitor-multi, analyze
│   ├── output/
│   │   └── output.go     # Output formatting (json/compact/default)
│   └── types/
│       └── types.go      # All Go type definitions
├── go.mod
├── go.sum
├── Makefile              # Build, format, lint, test targets
├── .golangci.yml         # Linter configuration
└── .editorconfig         # Editor settings
```

## Development

### Setup

```bash
make init      # Download dependencies
make tools     # Install dev tools (golangci-lint, gofumpt, goimports)
```

### Commands

```bash
make build        # Build for current platform
make build-pi     # Build for all Raspberry Pi variants (ARM64, ARMv7, ARMv6)
make build-all    # Build for all platforms

make format       # Format code (gofumpt + goimports)
make lint         # Run golangci-lint
make lint-fix     # Run golangci-lint with auto-fix
make vet          # Run go vet

make test         # Run all tests
make test-cover   # Run tests with coverage report

make check        # Run all checks (format-check, lint, vet, test)
make pre-commit   # Pre-commit workflow (format, lint-fix, test)
make ci           # CI workflow (strict, no fixes)

make help         # Show all available commands
```

### Code Style

This project uses **golangci-lint** with 25+ linters and **gofumpt** for stricter formatting.

Key conventions:
- **Tabs** for indentation (Go standard)
- Error wrapping with `fmt.Errorf("...: %w", err)`
- Named return values where it improves clarity
- Preallocate slices when size is known

### Go Guidelines

The project uses standard Go idioms:

```go
// Always handle errors
result, err := client.SendMessage(...)
if err != nil {
    return fmt.Errorf("failed to get states: %w", err)
}

// Use typed generics for API responses
states, err := client.SendMessageTyped[[]types.HAState](c, "get_states", nil)

// Named returns for complex functions
func parseTimeArgs(args []string) (filtered []string, fromTime, toTime *time.Time) {
    // ...
}
```

### Adding New Commands

1. Create handler function in appropriate file under `internal/handlers/`
2. Export from handler file (public function name)
3. Register in `commandRegistry` map in `cmd/ha-ws-client/main.go`
4. Update help text in `showHelp()` function
5. Update `README.md` CLI documentation

Handler signature:
```go
func HandleMyCommand(ctx *Context) error {
    // ctx.Client - WebSocket client
    // ctx.Args - Command arguments
    // ctx.FromTime, ctx.ToTime - Time range filters

    // Implementation
    return nil
}
```

### Before Committing

**Always run the formatter and linter before committing changes:**

```bash
# In the HA add-on environment (no make available):
PATH=/config/.gopath/bin:$PATH gofumpt -l -w .
PATH=/config/.gopath/bin:$PATH goimports -local github.com/home-assistant-blueprints/ha-ws-client-go -w .
PATH=/config/.gopath/bin:$PATH GOPATH=/config/.gopath GOCACHE=/config/.gopath/cache golangci-lint run --fix

# With make available:
make pre-commit   # Runs format, lint-fix, and tests
```

**Common linter issues to watch for:**
- Repeated string literals → extract to constants
- Unchecked type assertions → use `val, ok := x.(Type)` pattern
- Assignment operators → use `x /= 2` instead of `x = x / 2`
- Error return values not checked → always check or explicitly ignore with `_ =`

**Install dev tools if not present:**
```bash
GOPATH=/config/.gopath go install mvdan.cc/gofumpt@v0.7.0
GOPATH=/config/.gopath go install golang.org/x/tools/cmd/goimports@latest
GOPATH=/config/.gopath go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.62.2
```

### WebSocket API Patterns

```go
// Simple request/response
states, err := client.SendMessageTyped[[]types.HAState](c, "get_states", nil)

// With parameters
result, err := client.SendMessageTyped[[][]types.HistoryState](c,
    "history/period/"+startTime.Format(time.RFC3339),
    map[string]any{
        "filter_entity_id": entityID,
        "end_time":         endTime.Format(time.RFC3339),
    })

// For subscriptions (watch, monitor)
id, cleanup, err := c.SubscribeToTrigger(trigger, func(vars map[string]any) {
    // Handle event
}, timeout)
defer cleanup()
```

### Error Handling

Use the custom error type from `internal/client/`:

```go
import "github.com/home-assistant-blueprints/ha-ws-client-go/internal/client"

// HAClientError is returned for API errors
var clientErr *client.HAClientError
if errors.As(err, &clientErr) {
    output.Error(err, clientErr.Code)
}

// For missing entities
return fmt.Errorf("%w: %s", handlers.ErrEntityNotFound, entityID)
```

## Environment

- **Runtime**: Single static binary (no runtime dependencies)
- **API Access**: Via `SUPERVISOR_TOKEN` environment variable (set by HA add-on)
- **WebSocket**: `ws://supervisor/core/api/websocket`

## Testing

```bash
# Build and test help
make build && ./build/ha-ws-client --help

# Test in HA environment
export SUPERVISOR_TOKEN="your-token"
./ha-ws-client ping
./ha-ws-client state sun.sun
./ha-ws-client states-filter "light.*"
./ha-ws-client history sensor.temperature 4
./ha-ws-client traces
```

## Dependencies

- `github.com/gorilla/websocket` - WebSocket client
- `gopkg.in/yaml.v3` - YAML parsing

Dev tools (installed via `make tools`):
- `golangci-lint` - Comprehensive linting
- `gofumpt` - Stricter code formatting
- `goimports` - Import organization

## Output Formats

All commands support output format flags for context-efficient AI consumption:

| Flag | Format | Use Case |
|------|--------|----------|
| `--json` | JSON | Machine-readable, most context-efficient |
| `--compact` | Compact | Reduced verbosity, single-line entries |
| (default) | Default | Human-readable with formatting |

**Additional flags:**
- `--no-headers` - Suppress headers/separators
- `--no-timestamps` - Suppress timestamps
- `--max-items=N` - Limit output items

**Examples:**
```bash
./ha-ws-client states --json              # JSON array of all states
./ha-ws-client history sensor.temp 4 --compact  # One-line-per-entry
./ha-ws-client traces --json --max-items=5     # Last 5 traces as JSON
```

**JSON output structure:**
```json
{"success": true, "data": [...], "count": 42, "command": "states"}
{"success": true, "message": "pong"}
{"success": false, "error": "Entity not found: light.kitchen"}
```

## Versioning

Version information is embedded at build time via ldflags:

```bash
# Check version
./build/ha-ws-client --version

# Build with specific version
make build VERSION=1.2.3
```

**Version synchronization requirements:**
1. Update `VERSION` in Makefile before release (or use git tags)
2. Add entry to `CHANGELOG.md` following Keep a Changelog format
3. Keep version synchronized with validate-blueprint-go when making coordinated releases
4. GitHub Actions will automatically build and attach binaries to releases

**Pre-commit hook enforces:**
- CHANGELOG.md must exist
- Makefile VERSION must match latest CHANGELOG.md version
- Warning (not blocking) if versions differ between ha-ws-client-go and validate-blueprint-go

## Cross-Compilation

Build for Raspberry Pi and other platforms:

```bash
make build-pi          # All Pi variants (ARM64, ARMv7, ARMv6)
make build-linux-arm64 # Pi 4 (64-bit)
make build-linux-armv7 # Pi 2/3 (32-bit)
make build-linux-armv6 # Pi Zero/1
make build-all         # All platforms
make sizes             # Show binary sizes
```

Binary sizes are ~5.5-5.8MB (vs ~50MB+ for Node.js equivalent).
