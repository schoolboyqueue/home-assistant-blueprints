# Scripts

This directory contains tools for developing and debugging Home Assistant blueprints.

## Tools

| Tool | Description |
|------|-------------|
| [validate-blueprint](./validate-blueprint/) | Validate blueprint YAML structure before importing to HA |
| [ha-ws-client](./ha-ws-client/) | WebSocket API client for querying HA and debugging automations |

## Recommended Workflow

1. **Edit blueprint** - Make changes to YAML
2. **Validate locally** - `python3 scripts/validate-blueprint/validate-blueprint.py <file>`
3. **Debug with ha-ws-client** - Query states, check history, view traces
4. **Validate all** - `python3 scripts/validate-blueprint/validate-blueprint.py --all`
5. **Commit & push** - Git hooks automatically validate before commit
