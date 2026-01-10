# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Repository Overview

This repository contains production-ready Home Assistant Blueprints for home automation. Blueprints are YAML files with Jinja2 templating that define reusable automation templates for Home Assistant.

See [CONTRIBUTING.md](CONTRIBUTING.md) for contribution guidelines.

## Commands

### Make Targets

| Target | Description |
|--------|-------------|
| `make setup` | Setup development environment (pre-commit, Go tools, docs) |
| `make validate` | Validate all blueprints |
| `make validate-single FILE=<path>` | Validate a single blueprint |
| `make build` | Build all Go tools |
| `make go-init` | Download Go dependencies |
| `make go-tools` | Install Go dev tools (golangci-lint, gofumpt, goimports) |
| `make go-test` | Run Go tests |
| `make go-lint` | Run Go linters (with auto-fix) |
| `make go-format` | Format Go code |
| `make go-vet` | Run go vet |
| `make go-check` | Run all Go checks (format, lint, vet, test) |
| `make go-clean` | Clean Go build artifacts |
| `make docs-check` | Check docs with Biome |
| `make docs-fix` | Fix docs issues with Biome |
| `make check` | Run all checks (Go + blueprints + docs) |
| `make clean` | Clean build artifacts |
| `make help` | Show all available targets |

### Validate Blueprints

```bash
# Validate a single blueprint
./scripts/validate-blueprint-go/build/validate-blueprint <path/to/blueprint.yaml>

# Validate all blueprints in the repository
./scripts/validate-blueprint-go/build/validate-blueprint --all

# Or use make targets
make validate                           # Validate all blueprints
make validate-single FILE=<path>        # Validate a single file
```

The validator checks:

- YAML syntax and blueprint schema
- Input/selector definitions and !input reference validation
- Template syntax (balanced delimiters, no !input inside {{ }})
- Service call structure
- Version sync (blueprint name vs blueprint_version variable)
- Trigger validation (no templates in `for:` duration)
- Condition structure validation
- Mode validation (single, restart, queued, parallel)
- Delay and wait_template/wait_for_trigger validation
- Empty sequence detection
- README.md and CHANGELOG.md existence

## Architecture

### Blueprint Structure

Each blueprint lives in `blueprints/<blueprint-name>/` and contains:

- `*.yaml` - The blueprint file (named `*_pro.yaml` or `*_pro_blueprint.yaml`)
- `README.md` - Documentation
- `CHANGELOG.md` - Version history

### Blueprint YAML Structure

```yaml
blueprint:
  name: "Blueprint Name vX.Y.Z"
  description: >-
    Multi-line description
  domain: automation
  author: "Author Name"
  source_url: https://github.com/...
  input:
    group_name:
      name: Group Label
      icon: mdi:icon-name
      input:
        input_name:
          name: Input Label
          description: Description
          default: value
          selector:
            selector_type:
              options...

variables:
  blueprint_version: "X.Y.Z"
  # Variables defined here, referenced in templates

trigger:
  - platform: state
    entity_id: !input input_name
    # ...

action:
  - if:
      - condition: template
        value_template: "{{ expression }}"
    then:
      - service: domain.service
        target:
          entity_id: !input target_input
```

### Key Patterns

1. **!input tags**: Use `!input input_name` to reference blueprint inputs. Cannot be used inside `{{ }}` templates - bind to a variable first
2. **Variables section**: Must be at root level (not under `blueprint:`). Variables can use `!input` and are available in templates
3. **Selectors**: Every input should have a `selector` (entity, number, boolean, select, etc.)
4. **Grouped inputs**: Inputs are organized into collapsible groups with `name`, `icon`, and nested `input` dict
5. **Debug logging**: Use `logbook.log` service (not `system_log.write`) for debug output - it appears in Home Assistant's Logbook UI which is easier for users to find. Check debug level with direct comparison: `{{ debug_level_v in ['basic', 'verbose'] }}`

## Conventions

### Commits

Uses Conventional Commits:

- `feat(blueprint-name): description` - New features
- `fix(blueprint-name): description` - Bug fixes
- `docs(readme): description` - Documentation changes
- `refactor: description` - Code restructuring

### Versioning

Each blueprint has its own semantic version in:

1. Blueprint `name` field: `"Blueprint Name vX.Y.Z"`
2. `blueprint_version` variable
3. `CHANGELOG.md` - Add entry for new version

The blueprint name and variable must stay in sync.

### Go Tool Versioning

The Go tools (ha-ws-client-go and validate-blueprint-go) each have their own semantic version:

1. **Makefile VERSION**: Set `VERSION=X.Y.Z` or pass via `make build VERSION=X.Y.Z`
2. **CHANGELOG.md**: Add entry for each release following Keep a Changelog format
3. **Version flag**: Run `--version` to check current version

When updating Go tools:

1. Update the VERSION in Makefile (or rely on git tag for releases)
2. Add entry to CHANGELOG.md with date and changes
3. Keep both tools' versions synchronized when making coordinated changes
4. GitHub Actions will automatically build and release binaries on version tags

**Pre-commit hooks:**

- The project uses [pre-commit](https://pre-commit.com) instead of Husky
- Hooks validate blueprints, Go code, commit messages, and more
- Configuration: `.pre-commit-config.yaml`
- Hook scripts: `.pre-commit/hooks/`
- Setup: `pip install pre-commit && pre-commit install`

**Pre-commit checks for Go tools:**

- CHANGELOG.md must exist for both tools
- Makefile VERSION must match latest CHANGELOG.md version entry
- Warning if tool versions are not synchronized (not blocking)

### Markdown

Uses markdownlint with:

- Line length limit disabled (MD013: false)
- HTML elements allowed: div, h1, p, em, b, a, img, br, details, summary, kbd

### Git Commits

- Never include Claude Code references or co-author lines in commit messages
- Always update the root README.md when adding new blueprints (gallery entry + repository structure)

### GitHub Pages Website

The project website is served from `docs/` and must be kept in sync with blueprints:

**Files:**

- `docs/index.html` - Main website with blueprint gallery
- `docs/styles.css` - Styling
- `docs/script.js` - Interactive functionality
- `_config.yml` - Jekyll configuration

**When adding a new blueprint:**

1. Update the blueprint count in the hero stats section (`<span class="stat-value">`)
2. Add a new `<article class="blueprint-card">` in the blueprints gallery section
3. Include: icon SVG, title, description, tags, import URL, and docs link
4. Import URL format: `https://my.home-assistant.io/redirect/blueprint_import/?blueprint_url=https://raw.githubusercontent.com/schoolboyqueue/home-assistant-blueprints/main/blueprints/<name>/<file>.yaml`

**When updating a blueprint:**

- Update the description if features changed significantly
- Update tags if new capabilities were added

**When removing a blueprint:**

1. Remove the blueprint card from the gallery
2. Update the blueprint count in the hero stats

## Documentation Maintenance

**Keep documentation in sync with code changes.** When making changes, update all affected documentation:

### Files to Update

| Change Type | Files to Update |
|-------------|-----------------|
| New blueprint | Root README.md (gallery + structure), docs/index.html, CLAUDE.md if patterns change |
| New Go tool feature | Tool's README.md, CLAUDE.md (architecture if new files), CHANGELOG.md |
| New Go tool internal package | Tool's README.md + CLAUDE.md (architecture sections), root README.md (structure) |
| New workflow file | Root README.md (structure section) |
| Changed directory structure | All README.md and CLAUDE.md files with architecture/structure sections |
| New npm script | package.json, CLAUDE.md (npm Scripts table), CONTRIBUTING.md (Available npm Scripts) |
| Contribution process change | CONTRIBUTING.md, CLAUDE.md if it affects documented workflows |

### Architecture Sections

These files contain directory structure diagrams that must stay current:

- `README.md` (root) - Repository structure
- `CONTRIBUTING.md` - Development setup and npm scripts table
- `scripts/ha-ws-client-go/README.md` - Architecture section
- `scripts/ha-ws-client-go/CLAUDE.md` - Architecture section
- `scripts/validate-blueprint-go/README.md` - Architecture section
- `scripts/validate-blueprint-go/CLAUDE.md` - Architecture + Package Structure sections

### Before Committing

1. If you added/removed/renamed files, update relevant architecture sections
2. If you added new commands or features, update README.md command tables
3. If you changed tool behavior, update CLAUDE.md usage examples
4. Run `git diff --stat` to see changed files and verify docs are updated
