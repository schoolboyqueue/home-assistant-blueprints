# Scripts

This directory contains tools for developing and debugging Home Assistant blueprints.

## Tools

| Tool                                              | Language | Description                                                    |
| ------------------------------------------------- | -------- | -------------------------------------------------------------- |
| [validate-blueprint-go](./validate-blueprint-go/) | Go       | Validate blueprint YAML structure before importing to HA       |
| [ha-ws-client-go](./ha-ws-client-go/)             | Go       | WebSocket API client for querying HA and debugging automations |

## Version Information

Both tools support `--version` to display version information:

```bash
./validate-blueprint --version
./ha-ws-client --version
```

## Recommended Workflow

1. **Edit blueprint** - Make changes to YAML
2. **Validate locally** - `./scripts/validate-blueprint-go/build/validate-blueprint <file>`
3. **Debug with ha-ws-client-go** - Query states, check history, view traces
4. **Validate all** - `./scripts/validate-blueprint-go/build/validate-blueprint --all`
5. **Commit & push** - Git hooks automatically validate before commit

## Building Tools

### validate-blueprint-go

```bash
cd scripts/validate-blueprint-go
make init      # Download dependencies
make tools     # Install dev tools (golangci-lint, gofumpt, goimports)
make build     # Build the binary
make build-pi  # Build for all Raspberry Pi variants
make build-all # Build for all platforms
make help      # Show all available commands
```

### ha-ws-client-go

```bash
cd scripts/ha-ws-client-go
make init      # Download dependencies
make tools     # Install dev tools (golangci-lint, gofumpt, goimports)
make build     # Build the binary
make build-pi  # Build for all Raspberry Pi variants
make build-all # Build for all platforms
make help      # Show all available commands
```

## Pre-built Binaries

Pre-built binaries for all platforms are available in the [GitHub Releases](https://github.com/jeremiah-k/home-assistant-blueprints/releases) section.
