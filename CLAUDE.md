# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Repository Overview

This repository contains production-ready Home Assistant Blueprints for home automation. Blueprints are YAML files with Jinja2 templating that define reusable automation templates for Home Assistant.

## Commands

### Validate Blueprints

```bash
# Validate a single blueprint
python3 scripts/validate-blueprint.py <path/to/blueprint.yaml>

# Validate all blueprints in the repository
python3 scripts/validate-blueprint.py --all
```

The validator checks YAML syntax, blueprint schema, input/selector definitions, template syntax, service call structure, and README version sync.

## Architecture

### Blueprint Structure

Each blueprint lives in `blueprints/<blueprint-name>/` and contains:
- `*.yaml` - The blueprint file (named `*_pro.yaml` or `*_pro_blueprint.yaml`)
- `README.md` - Documentation with version badge
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
3. README.md `**Version:** X.Y.Z` badge

All three must stay in sync (validator checks README vs blueprint).

### Markdown

Uses markdownlint with:
- Line length limit disabled (MD013: false)
- HTML elements allowed: div, h1, p, em, b, a, img, br, details, summary, kbd

### Git Commits

- Never include Claude Code references or co-author lines in commit messages
- Always update the root README.md when adding new blueprints (gallery entry + repository structure)
