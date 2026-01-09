# Pre-commit Hooks

This project uses [pre-commit](https://pre-commit.com) to run validation checks before commits.

## Installation

```bash
# Install pre-commit
pip install pre-commit

# Install the git hooks
pre-commit install

# Install commit-msg hook (for Conventional Commits validation)
pre-commit install --hook-type commit-msg
```

## Usage

Hooks run automatically on `git commit`. You can also run them manually:

```bash
# Run all hooks on all files
pre-commit run --all-files

# Run all hooks on staged files only
pre-commit run

# Run a specific hook
pre-commit run blueprint-validation --all-files
```

## Available Hooks

### Standard Hooks

- **trailing-whitespace**: Remove trailing whitespace
- **end-of-file-fixer**: Ensure files end with a newline
- **check-yaml**: Validate YAML syntax
- **check-json**: Validate JSON syntax
- **check-merge-conflict**: Check for merge conflict markers
- **mixed-line-ending**: Ensure consistent line endings (LF)
- **markdownlint-cli2**: Lint Markdown files

### Go Hooks

- **go-fmt**: Format Go code
- **go-vet**: Run go vet
- **go-mod-tidy**: Tidy go.mod files

### Go Linting

- **golangci-lint**: Official golangci-lint hook for Go code quality

### Conventional Commits

- **conventional-pre-commit**: Validates commit messages follow Conventional Commits format

### Custom Blueprint Hooks

- **blueprint-validation**: Validate blueprint YAML against schema (includes !input in template checks)
- **blueprint-version-check**: Check version consistency
- **blueprint-documentation**: Ensure README and CHANGELOG exist

### Custom Go Tool Hooks

- **go-tools-version-check**: Validate versions and changelogs

## Hook Scripts

Custom hook scripts in `.pre-commit/hooks/`:

- `validate-blueprints.sh`: Blueprint validation
- `check-blueprint-versions.sh`: Version consistency
- `check-documentation.sh`: Documentation checks
- `check-go-versions.sh`: Go tool version/changelog validation

## Configuration

The main configuration is in `.pre-commit-config.yaml` at the repository root.

To update hook versions:

```bash
pre-commit autoupdate
```

## Skipping Hooks

To skip hooks (use sparingly):

```bash
# Skip all hooks for a commit
git commit --no-verify

# Skip specific hooks
SKIP=blueprint-validation,golangci-lint git commit -m "..."
```

## Troubleshooting

### Error: "Cowardly refusing to install hooks with `core.hooksPath` set"

If you previously used Husky, you need to unset the git config:

```bash
git config --unset-all core.hooksPath
pre-commit install
pre-commit install --hook-type commit-msg
```

### Hooks not running

```bash
# Reinstall hooks
pre-commit uninstall
pre-commit install
pre-commit install --hook-type commit-msg
```

### Blueprint validator not found

```bash
# Build Go tools
npm run go:build
```

### golangci-lint not found

```bash
# Install Go dev tools
npm run go:tools
```

### Clean and reinstall

```bash
pre-commit clean
pre-commit install-hooks
```
