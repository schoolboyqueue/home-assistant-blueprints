# Contributing to Home Assistant Blueprints

Thank you for your interest in contributing! This guide covers contributions to both the Home Assistant blueprints and the Go tools in this repository.

## Table of Contents

- [Getting Started](#getting-started)
- [Contributing Blueprints](#contributing-blueprints)
- [Contributing to Go Tools](#contributing-to-go-tools)
- [Development Setup](#development-setup)
- [Coding Standards](#coding-standards)
- [Commit Conventions](#commit-conventions)
- [Pull Request Process](#pull-request-process)

## Getting Started

1. Fork the repository
2. Clone your fork:
   ```bash
   git clone https://github.com/YOUR_USERNAME/home-assistant-blueprints.git
   cd home-assistant-blueprints
   ```
3. Set up the development environment:
   ```bash
   npm install                    # Install husky for pre-commit hooks
   npm run go:tools               # Install Go development tools
   npm run go:build               # Build Go tools
   ```

## Contributing Blueprints

### Blueprint Structure

Each blueprint lives in its own directory under `blueprints/`:

```
blueprints/
└── my-blueprint/
    ├── my_blueprint_pro.yaml    # The blueprint file
    ├── README.md                # Documentation
    └── CHANGELOG.md             # Version history
```

### Blueprint Requirements

1. **YAML Structure**: Follow the Home Assistant blueprint schema:
   ```yaml
   blueprint:
     name: "My Blueprint vX.Y.Z"
     description: >-
       Description of what the blueprint does
     domain: automation
     author: "Your Name"
     source_url: https://github.com/...
     input:
       # Input definitions...

   variables:
     blueprint_version: "X.Y.Z"
     # Other variables...

   trigger:
     # Trigger definitions...

   action:
     # Action definitions...
   ```

2. **Version Sync**: The version in `name:` must match `blueprint_version`

3. **Documentation**: Include README.md and CHANGELOG.md

4. **Validation**: Must pass the blueprint validator:
   ```bash
   npm run validate:single blueprints/my-blueprint/my_blueprint_pro.yaml
   ```

### Blueprint Best Practices

- Use grouped inputs with icons for better organization
- Bind `!input` values to variables before using in templates (never use `!input` inside `{{ }}`)
- Include appropriate selectors for all inputs
- Use `logbook.log` for debug output (not `system_log.write`)
- Follow semantic versioning for blueprint versions

### Updating Existing Blueprints

1. Make your changes to the blueprint YAML
2. Update the version in both `name:` and `blueprint_version`
3. Add an entry to CHANGELOG.md
4. Update README.md if features changed
5. Run validation before committing

## Contributing to Go Tools

This repository includes two Go tools:

- **validate-blueprint-go**: CLI for validating blueprint YAML files
- **ha-ws-client-go**: CLI for Home Assistant WebSocket API

### Go Development Workflow

```bash
# Navigate to the tool directory
cd scripts/validate-blueprint-go   # or scripts/ha-ws-client-go

# Install dependencies
make init

# Install development tools (golangci-lint, gofumpt, goimports)
make tools

# Build
make build

# Run tests
make test

# Format code
make format

# Run linter
make lint

# Run all checks (format, lint, vet, test)
make check

# Pre-commit workflow (format, lint-fix, test)
make pre-commit
```

### Adding New Features

#### validate-blueprint-go

1. Identify the appropriate file based on validation type:
   - `schema.go` - Root structure and blueprint section validation
   - `inputs.go` - Input and selector validation
   - `triggers.go` - Trigger validation
   - `conditions.go` - Condition validation
   - `actions.go` - Action/service validation
   - `templates.go` - Jinja2 template validation

2. Add validation method to `BlueprintValidator`:
   ```go
   func (v *BlueprintValidator) ValidateSomething() {
       if invalid {
           v.AddErrorf("path: description of error")
       }
   }
   ```

3. Call from `Validate()` in `validator.go`

4. Update README.md with new check documentation

#### ha-ws-client-go

1. Add handler function in `internal/handlers/`:
   ```go
   func HandleMyCommand(ctx *Context) error {
       // Implementation
       return nil
   }
   ```

2. Register in `commandRegistry` in `cmd/ha-ws-client/main.go`

3. Update help text in `showHelp()`

4. Update README.md command tables

### Go Code Style

- Use **tabs** for indentation (Go standard)
- Wrap errors with `fmt.Errorf("...: %w", err)`
- Use named return values where it improves clarity
- Preallocate slices when size is known
- Follow [golangci-lint](https://golangci-lint.run/) recommendations

### Go Tool Versioning

When making functional code changes (not docs/tests only):

1. Update `VERSION` in the tool's Makefile
2. Add entry to CHANGELOG.md following [Keep a Changelog](https://keepachangelog.com/) format
3. Use semantic versioning: patch for fixes, minor for features, major for breaking changes

## Development Setup

### Prerequisites

- **Node.js** (for Husky pre-commit hooks)
- **Go 1.21+** (for Go tools)
- **npm** (for package management)

### Initial Setup

```bash
# Install npm dependencies (sets up Husky)
npm install

# Install Go development tools
npm run go:tools

# Build Go tools
npm run go:build

# Verify setup
npm run validate   # Should validate all blueprints
```

### Available npm Scripts

| Script | Description |
|--------|-------------|
| `npm run validate` | Validate all blueprints |
| `npm run validate:single <path>` | Validate a single blueprint |
| `npm run go:init` | Download Go dependencies |
| `npm run go:tools` | Install Go development tools |
| `npm run go:build` | Build Go tools |
| `npm run go:test` | Run Go tests |
| `npm run go:lint` | Run Go linters |
| `npm run go:format` | Format Go code |
| `npm run go:vet` | Run go vet |
| `npm run go:check` | Run all Go checks (format, lint, vet, test) |
| `npm run go:clean` | Clean Go build artifacts |
| `npm run docs:check` | Check docs with Biome |
| `npm run docs:fix` | Fix docs issues with Biome |

## Coding Standards

### Blueprints (YAML)

- Use 2-space indentation
- No tabs
- No trailing whitespace
- Include selectors for all inputs
- Group related inputs together

### Go Code

- Format with `gofumpt` (stricter than `gofmt`)
- Organize imports with `goimports`
- Pass `golangci-lint` checks
- Include tests for new functionality
- Document exported functions

### Markdown

- Follow markdownlint rules (configured in `.markdownlint.json`)
- No trailing whitespace
- Single blank line at end of file

## Commit Conventions

This project uses [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>[scope]: <description>

[optional body]

[optional footer(s)]
```

### Types

| Type | Description |
|------|-------------|
| `feat` | New feature |
| `fix` | Bug fix |
| `docs` | Documentation only |
| `style` | Formatting, no code change |
| `refactor` | Code restructuring |
| `perf` | Performance improvement |
| `test` | Adding/fixing tests |
| `build` | Build system changes |
| `ci` | CI configuration |
| `chore` | Maintenance |
| `revert` | Revert previous commit |

### Examples

```bash
# Blueprint changes
feat(adaptive-fan): add humidity-based control
fix(bathroom-light): correct timeout calculation

# Go tool changes
feat(validate): add hysteresis boundary detection
fix(ha-ws-client): handle connection timeout

# Documentation
docs(readme): update installation instructions
```

### Scope Suggestions

- Blueprint name: `adaptive-fan`, `bathroom-light`, etc.
- Go tools: `validate`, `ha-ws-client`
- General: `readme`, `ci`, `deps`

## Pull Request Process

### Before Submitting

1. **Run pre-commit checks locally**:
   ```bash
   # For blueprints - validator runs automatically via Husky
   git add .
   git commit -m "..."   # Pre-commit hook will validate

   # For Go code
   cd scripts/<tool>
   make pre-commit
   ```

2. **Ensure all checks pass**:
   - Blueprint validation
   - Go linting and tests (if applicable)
   - Version/changelog sync (for Go tools)

3. **Update documentation**:
   - Blueprint README.md and CHANGELOG.md
   - Go tool README.md for new commands/features
   - Root README.md if adding new blueprints

### PR Description

Include:
- Summary of changes
- Related issue (if applicable)
- Testing performed
- Screenshots (for UI-related changes)

### What the Pre-commit Hook Checks

The repository includes a comprehensive pre-commit hook that validates:

**For Blueprints:**
- YAML syntax and schema
- Version consistency between `name:` and `blueprint_version`
- No `!input` inside `{{ }}` templates
- No tabs (spaces preferred)
- No trailing whitespace
- README.md and CHANGELOG.md existence

**For Go Code:**
- golangci-lint passes
- Makefile VERSION matches CHANGELOG.md version
- CHANGELOG.md exists

### After Submitting

- Respond to review feedback
- Keep the PR updated with the base branch
- Squash commits if requested

## Questions?

If you have questions about contributing:

1. Check existing issues for similar questions
2. Open a new issue with the `question` label
3. Be specific about what you're trying to accomplish

Thank you for contributing!
