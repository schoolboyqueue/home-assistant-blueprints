# ha-ws-client-go

A high-performance Home Assistant WebSocket API client written in Go, optimized for running on resource-constrained devices like Raspberry Pi.

## Why Go?

Go offers several advantages for embedded devices like Raspberry Pi:

| Aspect | Interpreted (Node.js) | Go |
|--------|----------------------|-----|
| Binary Size | ~50MB+ (runtime) | ~5-8MB (single binary) |
| Memory Usage | ~50MB+ baseline | ~5-10MB |
| Startup Time | ~500ms+ (JIT) | ~10ms |
| Dependencies | Runtime + packages | None (static binary) |
| Cross-compilation | Complex | Built-in support |

## Features

- **Comprehensive commands** - states, history, automation debugging, monitoring
- **Zero runtime dependencies** - single static binary
- **Cross-compilation** for all Raspberry Pi variants
- **Efficient memory usage** - perfect for Pi Zero/1
- **Multiple output formats** - JSON, compact, default
- **AI-agent optimized** - context-efficient output

## Installation

### Pre-built Binaries

Download the appropriate binary for your platform from the releases page:

- `ha-ws-client-linux-arm64` - Raspberry Pi 4 (64-bit)
- `ha-ws-client-linux-armv7` - Raspberry Pi 2/3 (32-bit)
- `ha-ws-client-linux-armv6` - Raspberry Pi Zero/1
- `ha-ws-client-linux-amd64` - Linux x86_64
- `ha-ws-client-darwin-arm64` - macOS Apple Silicon
- `ha-ws-client-darwin-amd64` - macOS Intel

### Build from Source

Requires Go 1.22 or later:

```bash
# Clone the repository
cd scripts/ha-ws-client-go

# Build for current platform
make build

# Build for all Raspberry Pi variants
make build-pi

# Build for all platforms
make build-all
```

## Usage

```bash
# Set environment variable (required)
export SUPERVISOR_TOKEN="your-token-here"

# Basic commands
ha-ws-client state sun.sun
ha-ws-client states --json
ha-ws-client call light turn_on '{"entity_id":"light.kitchen"}'

# History commands
ha-ws-client history sensor.temperature 24
ha-ws-client logbook light.kitchen --from "2024-01-15 10:00"

# Automation debugging
ha-ws-client traces automation.morning_routine
ha-ws-client trace-debug 01KDQS4E2WHMYJYYXKC7K28XFG

# Real-time monitoring
ha-ws-client watch binary_sensor.motion 60
ha-ws-client monitor-multi light.kitchen light.living_room
```

## Commands

### Basic Commands
| Command | Description |
|---------|-------------|
| `state <entity_id>` | Get single entity state |
| `states` | Get all entity states (summary) |
| `states-json` | Get all states as JSON array |
| `states-filter <pattern> [--show-age]` | Filter states by entity_id pattern (--show-age shows staleness) |
| `config` | Get HA configuration |
| `services` | List all services |
| `call <domain> <service> [data]` | Call a service (data as JSON) |
| `template <template>` | Render a Jinja template |
| `ping` | Test connection |

### Diagnostic Commands
| Command | Description |
|---------|-------------|
| `device-health <entity_id>` | Check device responsiveness and stale detection |
| `compare <entity1> <entity2>` | Side-by-side entity comparison with attribute diffs |

### Log Commands
| Command | Description |
|---------|-------------|
| `logbook <entity_id> [hours]` | Get logbook entries (default 24h) |
| `history <entity_id> [hours]` | Get state history (default 24h) |
| `history-full <entity_id> [hours]` | Get history with full attributes |
| `attrs <entity_id> [hours]` | Attribute change history |
| `timeline <hours> <entity>...` | Multi-entity chronological timeline |
| `syslog` | Get system log errors/warnings |
| `stats <entity_id> [hours]` | Get sensor statistics |
| `context <context_id>` | Look up what triggered a state change |
| `watch <entity_id> [seconds]` | Live subscribe to state changes |

### Registry Commands
| Command | Description |
|---------|-------------|
| `entities [pattern]` | List/search entity registry |
| `devices [pattern]` | List/search device registry |
| `areas` | List all areas |

### Automation Debugging
| Command | Description |
|---------|-------------|
| `traces [automation_id] [--from]` | List automation traces (supports time filtering) |
| `trace <automation_id> <run_id>` | Get detailed trace |
| `trace-latest <automation_id>` | Get most recent trace without listing first |
| `trace-summary <automation_id>` | Quick overview of recent automation runs |
| `trace-vars <run_id>` | Show evaluated variables |
| `trace-timeline <run_id>` | Step-by-step execution timeline |
| `trace-trigger <run_id>` | Show trigger context details |
| `trace-actions <run_id>` | Show action results |
| `trace-debug <run_id>` | Comprehensive debug view |
| `automation-config <entity_id>` | Get automation configuration |
| `blueprint-inputs <entity_id>` | Validate blueprint inputs |

### Monitoring Commands
| Command | Description |
|---------|-------------|
| `monitor <entity_id>` | Monitor entity state changes |
| `monitor-multi <entity>...` | Monitor multiple entities |
| `analyze <entity_id>` | Analyze entity state patterns |

## Output Formats

The client supports three output formats optimized for different use cases:

```bash
# Default - Human readable
ha-ws-client state sun.sun

# Compact - Reduced verbosity
ha-ws-client state sun.sun --compact

# JSON - Machine readable (best for AI agents)
ha-ws-client state sun.sun --json
```

### Output Options

| Option | Description |
|--------|-------------|
| `--output=json` | Machine-readable JSON |
| `--output=compact` | Reduced verbosity |
| `--output=default` | Human-readable formatted |
| `--json` | Shorthand for `--output=json` |
| `--compact` | Shorthand for `--output=compact` |
| `--no-headers` | Hide section headers/titles |
| `--no-timestamps` | Hide timestamps in output |
| `--max-items=N` | Limit output to N items |

### Time Filtering

```bash
# Using hours
ha-ws-client history sensor.temp 48

# Using absolute times
ha-ws-client history sensor.temp --from "2024-01-15 10:00" --to "2024-01-15 18:00"
```

## Raspberry Pi Deployment

### For Pi 4 (64-bit Raspberry Pi OS)

```bash
# Build
make build-linux-arm64

# Copy to Pi
scp build/ha-ws-client-linux-arm64 pi@raspberrypi:/usr/local/bin/ha-ws-client
```

### For Pi 2/3 (32-bit Raspberry Pi OS)

```bash
# Build
make build-linux-armv7

# Copy to Pi
scp build/ha-ws-client-linux-armv7 pi@raspberrypi:/usr/local/bin/ha-ws-client
```

### For Pi Zero/1

```bash
# Build
make build-linux-armv6

# Copy to Pi
scp build/ha-ws-client-linux-armv6 pi@raspberrypi:/usr/local/bin/ha-ws-client
```

## Performance

Tested on Raspberry Pi 4:

| Metric | Value |
|--------|-------|
| Startup | ~12ms |
| Memory (idle) | ~6MB |
| Binary size | ~6MB |
| `states` command | ~45ms |

## Development

### Initial Setup

```bash
# Download dependencies
make init

# Install development tools (golangci-lint, gofumpt, goimports)
make tools
```

### Common Commands

```bash
# Format code (recommended before committing)
make format

# Run linter
make lint

# Run linter with auto-fix
make lint-fix

# Run all checks (format check + lint + vet + test)
make check

# Run tests
make test

# Run tests with coverage
make test-cover

# Build and run
make run ARGS="state sun.sun"
```

### Pre-commit Workflow

```bash
# Format, lint with fixes, and test
make pre-commit
```

### CI/CD

```bash
# Strict checks for CI (no auto-fixes)
make ci
```

### Available Make Targets

Run `make help` to see all available targets organized by category.

## Architecture

```
ha-ws-client-go/
├── cmd/ha-ws-client/    # Main entry point
│   └── main.go
├── internal/
│   ├── client/          # WebSocket client
│   │   └── client.go
│   ├── handlers/        # Command handlers
│   │   ├── basic.go
│   │   ├── history.go
│   │   ├── automation.go
│   │   ├── registry.go
│   │   └── monitor.go
│   ├── output/          # Output formatting
│   │   └── output.go
│   └── types/           # Type definitions
│       └── types.go
├── go.mod
├── go.sum
├── Makefile
├── .golangci.yml        # Linter configuration
├── .editorconfig        # Editor settings
├── .gitignore
└── README.md
```

## License

Same license as the parent project.
