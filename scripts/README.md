# Blueprint Validation Scripts

This directory contains tools for validating Home Assistant blueprints before pushing to GitHub or attempting import into Home Assistant.

## Quick Start

```bash
# Validate a single blueprint
python3 scripts/validate-blueprint/validate-blueprint.py <path/to/blueprint.yaml>

# Validate all blueprints in the repository
python3 scripts/validate-blueprint/validate-blueprint.py --all
```

## validate-blueprint.py

A comprehensive Python-based validator that checks blueprint files for common issues **before** you attempt to import them into Home Assistant.

### What It Checks

**YAML & Structure:**

- ✓ Valid YAML syntax
- ✓ Required root-level keys (`blueprint`, `trigger`, `action`)
- ✓ `variables` at root level (not nested under `blueprint`)
- ✓ Blueprint metadata (`name`, `description`, `domain`, `input`)

**Triggers:**

- ✓ Trigger platform/type presence
- ✓ Template triggers cannot reference automation variables (prevents runtime errors)
- ✓ Trigger entity_id fields are static strings (no templates allowed)
- ✓ Entity ID format validation (domain.entity_name)

**Inputs & Selectors:**

- ✓ Input definitions and grouping
- ✓ Valid selector types
- ✓ Proper input nesting

**Actions & Service Calls:**

- ✓ Service call structure
- ✓ Service format validation (domain.service_name)
- ✓ `data:` blocks must not be None/empty
- ✓ Proper nesting of `if`/`then`/`else` blocks
- ✓ `repeat` sequence validation
- ✓ Entity ID format in targets and service calls

**Jinja2 Templates:**

- ✓ No `!input` tags inside `{{ }}` blocks
- ✓ Balanced Jinja2 delimiters (expressions `{{}}`, control blocks `{%%}`, comments `{##}`)
- ✓ Detection of unsupported filters/functions

**Variables:**

- ✓ Proper dictionary structure
- ✓ Presence of `blueprint_version`

**Documentation Sync:**

- ✓ README.md version matches blueprint version
- ✓ README.md exists in blueprint directory

### Usage Examples

**Validate during development:**

```bash
# After editing a blueprint
python3 scripts/validate-blueprint.py adaptive-comfort-control/adaptive_comfort_control_pro_blueprint.yaml
```

**Validate before commit:**

```bash
# Check all blueprints
python3 scripts/validate-blueprint/validate-blueprint.py --all

# Only commit if validation passes
python3 scripts/validate-blueprint/validate-blueprint.py --all && git commit -m "..."
```

**Add to pre-commit hook:**

```bash
# .git/hooks/pre-commit
#!/bin/bash
python3 scripts/validate-blueprint/validate-blueprint.py --all || exit 1
```

### Exit Codes

- `0` - All validations passed
- `1` - Validation failed with errors

### Output Format

**Success:**

```text
Validating: multi-switch-light-control/multi_switch_light_control_pro.yaml

✅ Blueprint is valid!
```

**Errors:**

```text
Validating: example/blueprint.yaml

❌ ERRORS:
  • action[13].then[0].else[0].then[0].data: Cannot be None/empty
  • Missing required root key: 'trigger'

❌ Blueprint validation failed with 2 errors
```

**Warnings:**

```text
Validating: example/blueprint.yaml

⚠️  WARNINGS:
  • No 'blueprint_version' variable defined
  • blueprint.input.my_setting: No selector defined (inputs should have selectors)

✅ Blueprint is valid (with 2 warnings)
```

### Common Errors Detected

**1. Data block with None value**

```yaml
# ❌ WRONG
- service: system_log.write
  data:
level: info  # Not indented under data

# ✅ CORRECT
- service: system_log.write
  data:
    level: info
```

**2. Variables nested under blueprint**

```yaml
# ❌ WRONG
blueprint:
  variables:
    my_var: value

# ✅ CORRECT
blueprint:
  name: My Blueprint
variables:
  my_var: value
```

**3. !input inside templates**

```yaml
# ❌ WRONG
value_template: "{{ !input my_sensor }}"

# ✅ CORRECT
variables:
  my_sensor_var: !input my_sensor
# Then use {{ my_sensor_var }}
```

## Integration with Development Workflow

### Recommended Workflow

1. **Edit blueprint** - Make changes to YAML
2. **Validate locally** - `python3 scripts/validate-blueprint/validate-blueprint.py <file>`
3. **Fix errors** - Address any issues found
4. **Validate all** - `python3 scripts/validate-blueprint/validate-blueprint.py --all`
5. **Commit & push** - Changes are safe to publish

### VS Code Integration (Optional)

Add to `.vscode/tasks.json`:

```json
{
  "version": "2.0.0",
  "tasks": [
    {
      "label": "Validate Blueprint",
      "type": "shell",
      "command": "python3",
      "args": [
        "scripts/validate-blueprint/validate-blueprint.py",
        "${file}"
      ],
      "presentation": {
        "reveal": "always",
        "panel": "new"
      },
      "problemMatcher": []
    },
    {
      "label": "Validate All Blueprints",
      "type": "shell",
      "command": "python3",
      "args": [
        "scripts/validate-blueprint/validate-blueprint.py",
        "--all"
      ],
      "presentation": {
        "reveal": "always",
        "panel": "new"
      },
      "problemMatcher": []
    }
  ]
}
```

Then press `Cmd/Ctrl+Shift+P` → "Tasks: Run Task" → "Validate Blueprint"

## Why This Validator?

**Problem:** Home Assistant's import process only shows errors after you:

1. Push to GitHub
2. Copy the raw URL
3. Go to Home Assistant
4. Try to import
5. See error message
6. Start over

**Solution:** This validator catches 90%+ of structural issues **locally** before you push, saving time and iterations.

**Note:** This validator checks structure and syntax. It cannot validate:

- Entity existence (entities must exist in your HA instance)
- Runtime logic (requires actual execution)
- Device-specific behavior (requires physical devices)

For those, you still need to test the imported blueprint in your Home Assistant instance.

## Future Enhancements

Potential improvements:

- [ ] Check for deprecated service calls
- [ ] Validate entity domain matches (e.g., `light.` entities with `light.turn_on`)
- [ ] Detect circular variable dependencies
- [ ] Check for unused input definitions
- [ ] Validate trigger platform-specific fields
- [ ] Integration with CI/CD (GitHub Actions)
- [ ] JSON Schema validation against HA's blueprint schema

## Troubleshooting

**Import error: `yaml.constructor.ConstructorError: could not determine a constructor for the tag '!input'`**

This is expected when using generic YAML parsers. The validator handles `!input` tags with a custom constructor.

**"No blueprints found in repository"**

Make sure your blueprint files match the naming pattern:

- `*_pro.yaml`
- `*_pro_blueprint.yaml`
- `blueprint.yaml`

**Script fails with Python errors**

Ensure you have Python 3.8+ and PyYAML:

```bash
python3 --version  # Should be 3.8+
pip3 install pyyaml
```
