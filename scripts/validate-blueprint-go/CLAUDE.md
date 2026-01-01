# CLAUDE.md

## Overview

Go CLI for validating Home Assistant Blueprint YAML files. Checks YAML syntax, blueprint schema, input/selector validation, Jinja2 templates, service calls, triggers, conditions, and more.

## Architecture

```
validate-blueprint-go/
├── main.go              # Main validator implementation
│                        # - YAML parsing with !input tag support
│                        # - Blueprint schema validation
│                        # - Input/selector validation
│                        # - Jinja2 template checking
│                        # - Service call validation
│                        # - Trigger/condition validation
├── go.mod               # Go module definition
├── go.sum               # Dependency checksums
├── Makefile             # Build, format, lint, test targets
├── .golangci.yml        # Linter configuration
├── .editorconfig        # Editor settings
├── README.md            # User documentation
└── build/               # Build output directory (created by make build)
    └── validate-blueprint  # Built binary
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

1. Parse YAML with custom `!input` tag handler
2. Check required root-level keys (`blueprint`, `trigger`, `action`)
3. Validate blueprint metadata (`name`, `description`, `domain`, `input`)
4. Validate all input definitions and selector types
5. Check trigger definitions and entity_id usage
6. Validate service calls structure
7. Check Jinja2 template syntax
8. Validate hysteresis boundaries
9. Check for documentation files

### Adding New Validations

1. Create validation function following existing patterns
2. Add to appropriate validation section in `main.go`
3. Update `README.md` with new check documentation

Validation function pattern:
```go
func validateSomething(node *yaml.Node, path string, results *ValidationResults) {
    // Perform validation
    if invalid {
        results.addError(path, "description of error")
    }
    if suspicious {
        results.addWarning(path, "description of warning")
    }
}
```

### Error Handling

Use the ValidationResults pattern:

```go
type ValidationResults struct {
    Errors   []ValidationIssue
    Warnings []ValidationIssue
}

func (r *ValidationResults) addError(path, message string) {
    r.Errors = append(r.Errors, ValidationIssue{Path: path, Message: message})
}

func (r *ValidationResults) addWarning(path, message string) {
    r.Warnings = append(r.Warnings, ValidationIssue{Path: path, Message: message})
}
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
