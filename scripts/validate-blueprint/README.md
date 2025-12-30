# validate-blueprint

A comprehensive Python-based validator that checks blueprint files for common issues **before** you attempt to import them into Home Assistant.

## Quick Start

```bash
# From repository root
python3 scripts/validate-blueprint/validate-blueprint.py <path/to/blueprint.yaml>

# Validate all blueprints in the repository
python3 scripts/validate-blueprint/validate-blueprint.py --all
```

## What It Checks

### YAML & Structure

- Valid YAML syntax
- Required root-level keys (`blueprint`, `trigger`, `action`)
- `variables` at root level (not nested under `blueprint`)
- Blueprint metadata (`name`, `description`, `domain`, `input`)

### Triggers

- Trigger platform/type presence
- Template triggers cannot reference automation variables
- Trigger entity_id fields are static strings (no templates allowed)
- Entity ID format validation (domain.entity_name)

### Inputs & Selectors

- Input definitions and grouping
- Valid selector types
- Proper input nesting

### Actions & Service Calls

- Service call structure and format validation
- `data:` blocks must not be None/empty
- Proper nesting of `if`/`then`/`else` blocks
- `repeat` sequence validation

### Jinja2 Templates

- No `!input` tags inside `{{ }}` blocks
- Balanced Jinja2 delimiters
- Detection of unsupported filters/functions

### Documentation Sync

- README.md version matches blueprint version
- README.md exists in blueprint directory

## Exit Codes

- `0` - All validations passed
- `1` - Validation failed with errors

## VS Code Integration

Add to `.vscode/tasks.json`:

```json
{
  "version": "2.0.0",
  "tasks": [
    {
      "label": "Validate Blueprint",
      "type": "shell",
      "command": "python3",
      "args": ["scripts/validate-blueprint/validate-blueprint.py", "${file}"],
      "problemMatcher": []
    },
    {
      "label": "Validate All Blueprints",
      "type": "shell",
      "command": "python3",
      "args": ["scripts/validate-blueprint/validate-blueprint.py", "--all"],
      "problemMatcher": []
    }
  ]
}
```

## Pre-commit Hook

```bash
# .git/hooks/pre-commit
#!/bin/bash
python3 scripts/validate-blueprint/validate-blueprint.py --all || exit 1
```
