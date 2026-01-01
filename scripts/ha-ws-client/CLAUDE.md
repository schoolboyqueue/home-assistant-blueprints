# CLAUDE.md

## Overview

TypeScript CLI for Home Assistant WebSocket API. Runs in HA add-on environment. Commands for entity states, services, history, and automation debugging.

## Architecture

```
ha-ws-client/
├── ha-ws-client.ts      # Entry point - CLI parsing, WebSocket, routing
├── src/
│   ├── client.ts        # WebSocket request/response, message IDs
│   ├── errors.ts        # Custom errors (HAClientError, EntityNotFoundError)
│   ├── types.ts         # All TypeScript interfaces + Result/Option types + Schema validation
│   ├── utils.ts         # Date parsing, attribute formatting, YAML loading
│   ├── output.ts        # Output formatting with strategy pattern (json/compact/default adapters)
│   ├── index.ts         # Barrel exports
│   └── handlers/
│       ├── index.ts     # Command handler registry
│       ├── basic.ts     # ping, state, states, config, services, call, template
│       ├── history.ts   # logbook, history, attrs, timeline, stats, syslog
│       ├── registry.ts  # entities, devices, areas
│       ├── automation.ts # traces, trace, trace-vars, trace-timeline, trace-debug, etc.
│       └── monitor.ts   # watch, monitor, monitor-multi, analyze
```

## Development

### Commands

```bash
npm run ha -- <command>     # Run the CLI
npm run typecheck           # Type check without emitting
npm run lint                # Run Biome linter
npm run format              # Format with Biome
npm run fix                 # Auto-fix lint/format issues
npm run check               # Run all Biome checks
```

### Code Style

This project uses **Biome** for linting and formatting. Key rules:

- **Single quotes** for strings
- **Semicolons** required
- **2-space indentation**
- **100 character line width**
- **ES5 trailing commas**
- **Node.js import protocol** required (e.g., `import * as fs from 'node:fs'`)

### TypeScript Guidelines

The project uses strict TypeScript settings:

- `strict: true` - All strict checks enabled
- `noUncheckedIndexedAccess: true` - Array/object access returns `T | undefined`
- `noImplicitReturns: true` - All code paths must return
- `noUnusedLocals/Parameters: true` - No unused variables
- `exactOptionalPropertyTypes: true` - Distinguishes `undefined` from optional

**Important patterns:**

```typescript
// Always handle potentially undefined array access
const arg = ctx.args[1];  // string | undefined
if (!arg) { /* handle missing arg */ }

// Use readonly for type interfaces
interface HAState {
  readonly entity_id: string;
  readonly state: string;
}

// Export types with 'type' keyword
import type { CommandContext } from './types.js';
```

### Adding New Commands

1. Create handler function in appropriate file under `src/handlers/`
2. Export from `src/handlers/index.ts`
3. Register in `commandHandlers` record in `src/handlers/index.ts`
4. Update help text in `ha-ws-client.ts`
5. Update `README.md` CLI documentation

Handler signature:
```typescript
export async function handleMyCommand(ctx: CommandContext): Promise<void> {
  const { ws, args, fromTime, toTime } = ctx;
  // Implementation
}
```

### WebSocket API Patterns

```typescript
// Simple request/response
const states = await sendMessage<HAState[]>(ws, 'get_states');

// With parameters
const result = await sendMessage<HistoryState[]>(ws, 'history/history_during_period', {
  start_time: startTime.toISOString(),
  entity_ids: [entityId],
});

// For subscription-based commands (template, watch), handle message events directly
const id = nextId();
ws.on('message', handler);
ws.send(JSON.stringify({ id, type: 'subscribe_events' }));
```

### Error Handling

Use custom error classes from `src/errors.ts`:

```typescript
import { EntityNotFoundError, HAClientError } from '../errors.js';

// Throw for missing entities
throw new EntityNotFoundError(entityId);

// Throw for HA API errors (handled automatically by sendMessage)
throw new HAClientError('Something failed', 'ERROR_CODE');
```

## Environment

- **Runtime**: Node.js 20+ with tsx for direct TypeScript execution
- **API Access**: Via `SUPERVISOR_TOKEN` environment variable (set by HA add-on)
- **WebSocket**: `ws://supervisor/core/api/websocket`

## Testing

No automated tests yet. Test manually:

```bash
# Basic connectivity
npm run ha -- ping

# Entity operations
npm run ha -- state sun.sun
npm run ha -- states-filter "light.*"

# History queries
npm run ha -- history sensor.temperature 4
npm run ha -- attrs climate.thermostat 12

# Automation debugging
npm run ha -- traces
npm run ha -- blueprint-inputs automation.some_automation

# Advanced trace debugging
npm run ha -- trace-timeline <run_id>   # Execution timeline
npm run ha -- trace-trigger <run_id>    # Trigger context
npm run ha -- trace-actions <run_id>    # Action results
npm run ha -- trace-debug <run_id>      # Comprehensive view
```

## Dependencies

- `ws` - WebSocket client
- `js-yaml` - YAML parsing for blueprint-inputs command (lazy-loaded)
- `tsx` - TypeScript execution
- `@biomejs/biome` - Linting and formatting
- `typescript` - Type checking

## Output Formats

All commands support output format flags for context-efficient AI consumption:

| Flag | Format | Use Case |
|------|--------|----------|
| `--json` | JSON | Machine-readable, most context-efficient |
| `--compact` | Compact | Reduced verbosity, single-line entries |
| (default) | Default | Human-readable with formatting |

**Additional flags:**
- `--no-headers` - Suppress headers/separators
- `--no-timestamps` - Suppress timestamps
- `--max-items=N` - Limit output items

**Examples:**
```bash
npm run ha -- states --json              # JSON array of all states
npm run ha -- history sensor.temp 4 --compact  # One-line-per-entry
npm run ha -- traces --json --max-items=5     # Last 5 traces as JSON
```

**JSON output structure:**
```json
{"success": true, "data": [...], "count": 42, "command": "states"}
{"success": true, "message": "pong"}
{"success": false, "error": "Entity not found: light.kitchen", "code": "ENTITY_NOT_FOUND"}
```

**Error codes:** `ENTITY_NOT_FOUND`, `AUTH_FAILED`, `INVALID_DATE`, `INVALID_JSON`

## Key Type Patterns

**Result type** - Explicit error handling (no exceptions):
```typescript
import { Result, ok, err, isOk } from './src/types.js';
// Return Result<T, E> instead of throwing
```

**Branded types** - Nominal typing for entity_id, context_id, etc.:
```typescript
import { EntityId, entityId } from './src/types.js';
const id: EntityId = entityId('light.kitchen');  // Type-safe
```

**Schema validation** - Runtime validation without dependencies:
```typescript
import { Schema, HAStateSchema } from './src/types.js';
const result = HAStateSchema.validate(unknownData);
```
