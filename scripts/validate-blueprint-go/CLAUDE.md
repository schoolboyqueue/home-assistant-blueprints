# CLAUDE.md

## Overview

Go CLI for validating Home Assistant Blueprint YAML files. Checks YAML syntax, blueprint schema, input/selector validation, Jinja2 templates, service calls, triggers, conditions, and more.

## Architecture

```
validate-blueprint-go/
├── main.go                      # CLI entry point and orchestration
├── internal/
│   └── validator/               # Core validation package
│       ├── validator.go         # BlueprintValidator struct & Validate() orchestration
│       ├── constants.go         # Configuration constants (ValidModes, ValidConditionTypes, etc.)
│       ├── yaml.go              # YAML loading with !input tag support
│       ├── schema.go            # Structure & blueprint section validation
│       ├── inputs.go            # Input/selector validation
│       ├── triggers.go          # Trigger validation
│       ├── conditions.go        # Condition validation
│       ├── actions.go           # Action/service validation
│       ├── templates.go         # Jinja2 template validation
│       ├── hysteresis.go        # Variable & hysteresis boundary validation
│       ├── helpers.go           # Utility functions
│       └── reporter.go          # Result formatting & display
├── go.mod                       # Go module definition
├── go.sum                       # Dependency checksums
├── Makefile                     # Build, format, lint, test targets
├── .golangci.yml                # Linter configuration
├── .editorconfig                # Editor settings
├── README.md                    # User documentation
└── build/                       # Build output directory (created by make build)
    └── validate-blueprint       # Built binary
```

## Package Structure

### `internal/validator`

The core validation logic is organized into focused files:

| File | Purpose |
|------|---------|
| `validator.go` | Main `BlueprintValidator` struct and `Validate()` orchestration |
| `constants.go` | All configuration constants (ValidModes, ValidConditionTypes, ValidSelectorTypes, HysteresisPatterns, Jinja2Builtins) |
| `yaml.go` | YAML parsing with custom `!input` tag handling |
| `schema.go` | Root structure validation, blueprint section, mode, version sync |
| `inputs.go` | Input definitions and selector validation |
| `triggers.go` | Trigger validation and entity_id checks |
| `conditions.go` | Condition validation including nested conditions |
| `actions.go` | Action/service validation including choose/if/repeat blocks |
| `templates.go` | Jinja2 template syntax validation |
| `hysteresis.go` | Variable validation and hysteresis boundary detection |
| `helpers.go` | Utility functions (ToFloat, Abs, ContainsVariableRef, etc.) |
| `reporter.go` | Result formatting with colored output |

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

This project uses **golangci-lint** with 23+ linters and **gofumpt** for stricter formatting.

Key conventions:
- **Tabs** for indentation (Go standard)
- Error wrapping with `fmt.Errorf("...: %w", err)`
- Named return values where it improves clarity
- Preallocate slices when size is known

### Go Guidelines

The project uses standard Go idioms:

```go
// Always handle errors
data, err := os.ReadFile(path)
if err != nil {
    return fmt.Errorf("failed to read file: %w", err)
}

// Use colored output for validation results
color.Green("✓ Valid blueprint: %s", filename)
color.Red("✗ Error: %s", message)
color.Yellow("⚠ Warning: %s", message)
```

### Validation Flow

1. Parse YAML with custom `!input` tag handler (`yaml.go`)
2. Check required root-level keys (`schema.go`)
3. Validate blueprint metadata (`schema.go`)
4. Validate mode settings (`schema.go`)
5. Validate all input definitions and selector types (`inputs.go`)
6. Validate hysteresis boundaries (`hysteresis.go`)
7. Validate variables section (`hysteresis.go`)
8. Validate version sync (`schema.go`)
9. Check trigger definitions and entity_id usage (`triggers.go`)
10. Validate conditions (`conditions.go`)
11. Validate service calls structure (`actions.go`)
12. Check Jinja2 template syntax (`templates.go`)
13. Validate input references (`inputs.go`)
14. Check for documentation files (`reporter.go`)

### Adding New Validations

1. Identify the appropriate file based on validation type
2. Add validation method to `BlueprintValidator` receiver
3. Call the new method from `Validate()` in `validator.go`
4. Update `README.md` with new check documentation

Validation method pattern:
```go
// In the appropriate file (e.g., schema.go, inputs.go, etc.)
func (v *BlueprintValidator) ValidateSomething() {
    // Perform validation
    if invalid {
        v.AddErrorf("path: description of error")
    }
    if suspicious {
        v.AddWarningf("path: description of warning")
    }
}
```

### Error Handling

Use the `AddError`/`AddWarning` methods on `BlueprintValidator`:

```go
// Add error
v.AddError("'variables' must be a dictionary")
v.AddErrorf("Missing required key: '%s'", key)

// Add warning
v.AddWarning("No variables section defined")
v.AddWarningf("Unknown selector type '%s'", selectorType)
```

## Usage

```bash
# Validate a single blueprint
./build/validate-blueprint path/to/blueprint.yaml

# Validate all blueprints in repository
./build/validate-blueprint --all
```

## Exit Codes

| Code | Description |
|------|-------------|
| `0` | All validations passed |
| `1` | Validation failed with errors |

## Dependencies

- `gopkg.in/yaml.v3` - YAML parsing with node access
- `github.com/fatih/color` - Colored terminal output

Dev tools (installed via `make tools`):
- `golangci-lint` - Comprehensive linting
- `gofumpt` - Stricter code formatting
- `goimports` - Import organization

## Versioning

Version information is embedded at build time via ldflags:

```bash
# Check version
./build/validate-blueprint --version

# Build with specific version
make build VERSION=1.2.3
```

**Version synchronization requirements:**
1. Update `VERSION` in Makefile before release (or use git tags)
2. Add entry to `CHANGELOG.md` following Keep a Changelog format
3. Keep version synchronized with ha-ws-client-go when making coordinated releases
4. GitHub Actions will automatically build and attach binaries to releases

**Pre-commit hook enforces:**
- CHANGELOG.md must exist
- Makefile VERSION must match latest CHANGELOG.md version
- Warning (not blocking) if versions differ between validate-blueprint-go and ha-ws-client-go

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

Binary sizes are ~5-8MB (vs ~50MB+ for Node.js equivalent).
