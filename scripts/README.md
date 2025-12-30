# Scripts

This directory contains tools for developing and debugging Home Assistant blueprints.

## Tools

| Tool | Description |
|------|-------------|
| [validate-blueprint](./validate-blueprint/) | Validate blueprint YAML structure before importing to HA |
| [ha-ws-client](./ha-ws-client/) | WebSocket API client for querying HA and debugging automations |

## validate-blueprint

A comprehensive Python-based validator that checks blueprint files for common issues **before** you attempt to import them into Home Assistant.

### Quick Start

```bash
# Validate a single blueprint
python3 scripts/validate-blueprint/validate-blueprint.py <path/to/blueprint.yaml>

# Validate all blueprints in the repository
python3 scripts/validate-blueprint/validate-blueprint.py --all
```

### What It Checks

**YAML & Structure:**

- Valid YAML syntax
- Required root-level keys (`blueprint`, `trigger`, `action`)
- `variables` at root level (not nested under `blueprint`)
- Blueprint metadata (`name`, `description`, `domain`, `input`)

**Triggers:**

- Trigger platform/type presence
- Template triggers cannot reference automation variables
- Trigger entity_id fields are static strings (no templates allowed)
- Entity ID format validation (domain.entity_name)

**Inputs & Selectors:**

- Input definitions and grouping
- Valid selector types
- Proper input nesting

**Actions & Service Calls:**

- Service call structure and format validation
- `data:` blocks must not be None/empty
- Proper nesting of `if`/`then`/`else` blocks
- `repeat` sequence validation

**Jinja2 Templates:**

- No `!input` tags inside `{{ }}` blocks
- Balanced Jinja2 delimiters
- Detection of unsupported filters/functions

**Documentation Sync:**

- README.md version matches blueprint version
- README.md exists in blueprint directory

### Exit Codes

- `0` - All validations passed
- `1` - Validation failed with errors

## ha-ws-client

A TypeScript CLI tool for interacting with Home Assistant's WebSocket API. Useful for querying entity states, viewing history, and debugging automations.

### Quick Start

```bash
cd scripts/ha-ws-client
npm install

# Run commands
npx tsx ha-ws-client.ts <command> [args...]
```

### Commands

**Basic Commands:**

```bash
state <entity_id>              # Get single entity state
states                         # Get all entity states (summary)
states-json                    # Get all states as JSON
states-filter <pattern>        # Filter states by entity_id pattern
config                         # Get HA configuration
services                       # List all services
call <domain> <service> [data] # Call a service (data as JSON)
template <template>            # Render a Jinja template
ping                           # Test connection
```

**History Commands:**

```bash
logbook <entity_id> [hours]      # Get logbook entries (default 24h)
history <entity_id> [hours]      # Get state history (default 24h)
history-full <entity_id> [hours] # Get history with full attributes
attrs <entity_id> [hours]        # Attribute change history (compact)
timeline <hours> <entity>...     # Multi-entity chronological timeline
syslog                           # Get system log errors/warnings
stats <entity_id> [hours]        # Get sensor statistics
watch <entity_id> [seconds]      # Live subscribe to state changes
```

**Registry Commands:**

```bash
entities [pattern]    # List/search entity registry
devices [pattern]     # List/search device registry
areas                 # List all areas
```

**Automation Debugging:**

```bash
traces [automation_id]           # List automation traces
trace <run_id> [automation_id]   # Get detailed trace for a run
trace-vars <run_id> [auto_id]    # Show evaluated variables from trace
automation-config <entity_id>    # Get automation configuration
blueprint-inputs <entity_id>     # Validate blueprint inputs vs expected
context <context_id>             # Look up what triggered a state change
```

**Time Filtering (for history commands):**

```bash
--from "YYYY-MM-DD HH:MM"    # Start time
--to "YYYY-MM-DD HH:MM"      # End time
```

### Examples

```bash
# Get entity state
npx tsx ha-ws-client.ts state sun.sun

# Call a service
npx tsx ha-ws-client.ts call light turn_on '{"entity_id":"light.kitchen"}'

# View attribute changes for a light
npx tsx ha-ws-client.ts attrs light.kitchen 4

# Watch for state changes in real-time
npx tsx ha-ws-client.ts watch binary_sensor.motion 30

# Debug automation - list recent traces
npx tsx ha-ws-client.ts traces automation.bathroom_lights

# View variables from a specific trace
npx tsx ha-ws-client.ts trace-vars 01KDQS4E2WHMYJYYXKC7K28XFG

# Validate blueprint inputs match expected
npx tsx ha-ws-client.ts blueprint-inputs automation.bathroom_lights

# Render a template
echo "{{ states.light | list | count }}" | npx tsx ha-ws-client.ts template -
```

### Development

```bash
npm run check      # Run Biome linter + formatter check
npm run fix        # Auto-fix lint/format issues
npm run typecheck  # TypeScript type checking
```

### Requirements

- Node.js 18+
- Must run inside Home Assistant (requires `SUPERVISOR_TOKEN` environment variable)

## Integration with Development Workflow

### Recommended Workflow

1. **Edit blueprint** - Make changes to YAML
2. **Validate locally** - `python3 scripts/validate-blueprint/validate-blueprint.py <file>`
3. **Debug with ha-ws-client** - Query states, check history, view traces
4. **Validate all** - `python3 scripts/validate-blueprint/validate-blueprint.py --all`
5. **Commit & push** - Changes are safe to publish

### Pre-commit Hook

```bash
# .git/hooks/pre-commit
#!/bin/bash
python3 scripts/validate-blueprint/validate-blueprint.py --all || exit 1
```

### VS Code Integration

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
