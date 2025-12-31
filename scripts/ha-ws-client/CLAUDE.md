# CLAUDE.md

## Overview

This is a TypeScript CLI tool for interacting with Home Assistant's WebSocket API. It runs within a Home Assistant add-on environment and provides commands for querying entity states, calling services, viewing history, debugging automations, and more.

## Architecture

```
ha-ws-client/
├── ha-ws-client.ts      # Entry point - CLI parsing, WebSocket connection, message routing
├── src/
│   ├── client.ts        # WebSocket request/response handling, message ID management
│   ├── errors.ts        # Custom error classes (HAClientError, EntityNotFoundError, etc.)
│   ├── types.ts         # TypeScript interfaces for all data structures
│   ├── utils.ts         # Date parsing, attribute formatting, YAML module loading
│   └── handlers/
│       ├── index.ts     # Command handler registry and exports
│       ├── basic.ts     # Core commands: ping, state, states, config, services, call, template
│       ├── history.ts   # History commands: logbook, history, attrs, timeline, stats, syslog
│       ├── registry.ts  # Registry commands: entities, devices, areas
│       └── automation.ts # Debug commands: traces, trace, trace-vars, trace-timeline, trace-trigger, trace-actions, trace-debug, context, watch, blueprint-inputs
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
