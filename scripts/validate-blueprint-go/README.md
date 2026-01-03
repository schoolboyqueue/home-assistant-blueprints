# validate-blueprint-go

A comprehensive Home Assistant Blueprint validator written in Go, optimized for fast validation and easy distribution as a single binary.

## Why Go?

Go offers several advantages for validation tools:

| Aspect | Interpreted (Node.js) | Go |
|--------|----------------------|-----|
| Binary Size | ~50MB+ (runtime) | ~5-8MB (single binary) |
| Memory Usage | ~50MB+ baseline | ~5-10MB |
| Startup Time | ~500ms+ (JIT) | ~10ms |
| Dependencies | Runtime + packages | None (static binary) |
| Cross-compilation | Complex | Built-in support |

## Features

- **YAML Syntax Validation** - Validates YAML syntax with support for Home Assistant's `!input` tag
- **Blueprint Schema Validation** - Checks required keys and structure
- **Input/Selector Validation** - Validates input definitions and selector types
- **Template Syntax Checking** - Validates Jinja2 template syntax
- **Service Call Validation** - Validates service call structure and format
- **Trigger Validation** - Validates trigger definitions and entity_id usage
- **Condition Validation** - Validates condition structure and types
- **Mode Validation** - Validates automation modes (single, restart, queued, parallel)
- **Input Reference Validation** - Ensures all `!input` references point to defined inputs
- **Hysteresis Validation** - Detects inverted or missing hysteresis boundaries
- **Math Safety Checks** - Warns about potential division by zero, sqrt of negative, etc.
- **Boolean Literal Detection** - Warns about bare `true`/`false` that become strings

## Installation

### Pre-built Binaries

Download the appropriate binary for your platform from the releases page:

- `validate-blueprint-linux-arm64` - Raspberry Pi 4 (64-bit)
- `validate-blueprint-linux-armv7` - Raspberry Pi 2/3 (32-bit)
- `validate-blueprint-linux-armv6` - Raspberry Pi Zero/1
- `validate-blueprint-linux-amd64` - Linux x86_64
- `validate-blueprint-darwin-arm64` - macOS Apple Silicon
- `validate-blueprint-darwin-amd64` - macOS Intel

### Build from Source

Requires Go 1.22 or later:

```bash
# Clone the repository
cd scripts/validate-blueprint-go

# Build for current platform
make build

# Build for all Raspberry Pi variants
make build-pi

# Build for all platforms
make build-all
```

## Usage

```bash
# Validate a single blueprint
./build/validate-blueprint path/to/blueprint.yaml

# Validate all blueprints in the repository
./build/validate-blueprint --all
```

This finds all blueprint files matching patterns like `*_pro.yaml`, `*_pro_blueprint.yaml`, or `blueprint.yaml` in the repository.

## Exit Codes

| Code | Description |
|------|-------------|
| `0` | All validations passed |
| `1` | Validation failed with errors |

## What It Checks

### YAML & Structure
| Check | Description |
|-------|-------------|
| YAML syntax | Valid YAML syntax |
| Required keys | `blueprint`, `trigger`, `action` at root level |
| Variables placement | `variables` at root level (not nested under `blueprint`) |
| Blueprint metadata | `name`, `description`, `domain`, `input` |

### Triggers
| Check | Description |
|-------|-------------|
| Trigger platform | Presence of trigger platform/type |
| Template triggers | Cannot reference automation variables |
| Entity ID fields | Must be static strings (no templates allowed) |

### Inputs & Selectors
| Check | Description |
|-------|-------------|
| Input definitions | Valid input definitions and grouping |
| Selector types | Valid selector type names |
| Input nesting | Proper input structure |
| Select options | Valid select option values |

### Actions & Service Calls
| Check | Description |
|-------|-------------|
| Service format | Valid service call structure |
| Data blocks | Must not be None/empty |
| Control flow | Proper `if`/`then`/`else` nesting |
| Repeat sequences | Valid `repeat` block structure |
| Choose blocks | Valid `choose` block structure |

### Jinja2 Templates
| Check | Description |
|-------|-------------|
| Input tags | No `!input` tags inside `{{ }}` blocks |
| Delimiters | Balanced Jinja2 delimiters |

### Hysteresis Boundaries
| Check | Description |
|-------|-------------|
| Thresholds | ON/OFF threshold relationships |
| Gaps | Warns about inverted or missing gaps |

### Math Operations
| Check | Description |
|-------|-------------|
| Logarithms | Warns about potential `log()` with non-positive values |
| Square roots | Warns about potential `sqrt()` with negative values |
| List methods | Detects Python-style methods (should use Jinja2 filters) |

### Documentation
| Check | Description |
|-------|-------------|
| README.md | Exists in blueprint directory |
| CHANGELOG.md | Exists in blueprint directory |

## Raspberry Pi Deployment

### For Pi 4 (64-bit Raspberry Pi OS)

```bash
# Build
make build-linux-arm64

# Copy to Pi
scp build/validate-blueprint-linux-arm64 pi@raspberrypi:/usr/local/bin/validate-blueprint
```

### For Pi 2/3 (32-bit Raspberry Pi OS)

```bash
# Build
make build-linux-armv7

# Copy to Pi
scp build/validate-blueprint-linux-armv7 pi@raspberrypi:/usr/local/bin/validate-blueprint
```

### For Pi Zero/1

```bash
# Build
make build-linux-armv6

# Copy to Pi
scp build/validate-blueprint-linux-armv6 pi@raspberrypi:/usr/local/bin/validate-blueprint
```

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
make run ARGS="path/to/blueprint.yaml"

# Validate all blueprints
make run-all
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
validate-blueprint-go/
├── main.go                  # CLI entry point and orchestration
├── internal/
│   ├── common/              # Shared utilities
│   │   └── validators.go    # Common validation helpers
│   ├── shutdown/            # Graceful shutdown coordination
│   │   └── shutdown.go
│   ├── testfixtures/        # Test fixture utilities
│   │   └── fixtures.go
│   └── validator/           # Core validation package
│       ├── actions.go       # Action/service validation
│       ├── categories.go    # Category classification
│       ├── conditions.go    # Condition validation
│       ├── constants.go     # Configuration constants
│       ├── context.go       # Validation context
│       ├── helpers.go       # Utility functions
│       ├── hysteresis.go    # Hysteresis boundary validation
│       ├── inputs.go        # Input/selector validation
│       ├── reporter.go      # Result formatting & display
│       ├── schema.go        # Structure & blueprint validation
│       ├── templates.go     # Jinja2 template validation
│       ├── triggers.go      # Trigger validation
│       ├── types.go         # Type definitions
│       ├── validator.go     # Main validator orchestration
│       └── yaml.go          # YAML loading with !input support
├── go.mod                   # Go module definition
├── go.sum                   # Dependency checksums
├── Makefile                 # Development task automation
├── .golangci.yml            # Linter configuration
├── .editorconfig            # Editor settings
├── .gitignore               # Git ignore rules
├── README.md                # This file
└── build/                   # Build output directory (created by make build)
    └── validate-blueprint   # Built binary
```

## Integration

### VS Code Task

Add to `.vscode/tasks.json`:

```json
{
  "version": "2.0.0",
  "tasks": [
    {
      "label": "Validate Blueprint (Go)",
      "type": "shell",
      "command": "${workspaceFolder}/scripts/validate-blueprint-go/build/validate-blueprint",
      "args": ["${file}"],
      "problemMatcher": []
    }
  ]
}
```

### Pre-commit Hook

```bash
#!/bin/bash
./scripts/validate-blueprint-go/build/validate-blueprint --all || exit 1
```

### npm Script

The Go validator is configured as the default in package.json:

```json
{
  "scripts": {
    "validate": "./scripts/validate-blueprint-go/build/validate-blueprint --all",
    "validate:single": "./scripts/validate-blueprint-go/build/validate-blueprint"
  }
}
```

## License

Same license as the parent project.
