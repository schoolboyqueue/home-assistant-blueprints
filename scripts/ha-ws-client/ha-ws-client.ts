#!/usr/bin/env npx
// Usage: npx tsx ha-ws-client.ts <command> [args...]

/**
 * Home Assistant WebSocket API Client
 *
 * A command-line tool for interacting with Home Assistant's WebSocket API.
 * Provides commands for querying entity states, calling services, viewing
 * history, debugging automations, and more.
 *
 * @example
 * ```bash
 * # Get entity state
 * npx tsx ha-ws-client.ts state sun.sun
 *
 * # Call a service
 * npx tsx ha-ws-client.ts call light turn_on '{"entity_id":"light.kitchen"}'
 *
 * # View automation traces
 * npx tsx ha-ws-client.ts traces automation.bathroom_lights
 * ```
 *
 * @module ha-ws-client
 */

// Node.js built-in imports
import WebSocket from 'ws';
import { pendingRequests } from './src/client.js';
import { HAClientError } from './src/errors.js';
import { commandHandlers } from './src/handlers/index.js';
import { outputError, parseOutputArgs } from './src/output.js';
// Local module imports
import type { CommandContext, HAMessage } from './src/types.js';
import { parseTimeArgs } from './src/utils.js';

// =============================================================================
// Constants
// =============================================================================

/** WebSocket URL for connecting to Home Assistant via the Supervisor. */
const WS_URL = 'ws://supervisor/core/api/websocket' as const;

/** Default timeout for WebSocket operations in milliseconds. */
const DEFAULT_TIMEOUT_MS = 30_000 as const;

// =============================================================================
// Environment Validation
// =============================================================================

const TOKEN = process.env.SUPERVISOR_TOKEN;

if (!TOKEN) {
  console.error('Error: SUPERVISOR_TOKEN environment variable not set');
  process.exit(1);
}

// =============================================================================
// Command Line Parsing
// =============================================================================

const rawArgs = process.argv.slice(2);
// Parse output format flags first (--output=json, --compact, --json, etc.)
const argsAfterOutput = parseOutputArgs(rawArgs);
// Then parse time args
const {
  fromTime: globalFromTime,
  toTime: globalToTime,
  filteredArgs,
} = parseTimeArgs(argsAfterOutput);
const command = filteredArgs[0];

// =============================================================================
// Usage Help
// =============================================================================

if (!command) {
  console.log(`Usage: npx tsx ha-ws-client.ts <command> [args...] [--from "TIME"] [--to "TIME"]

Commands:
  state <entity_id>              - Get single entity state
  states                         - Get all entity states (summary)
  states-json                    - Get all states as JSON array
  states-filter <pattern>        - Filter states by entity_id pattern
  config                         - Get HA configuration
  services                       - List all services
  call <domain> <service> [data] - Call a service (data as JSON)
  template <template>            - Render a Jinja template (use - for stdin)
  ping                           - Test connection

Log Commands:
  logbook <entity_id> [hours]    - Get logbook entries (default 24h)
  history <entity_id> [hours]    - Get state history (default 24h)
  history-full <entity_id> [hours] - Get history with full attributes
  attrs <entity_id> [hours]      - Attribute change history (compact)
  timeline <hours> <entity>...   - Multi-entity chronological timeline
  syslog                         - Get system log errors/warnings
  stats <entity_id> [hours]      - Get sensor statistics (default 24h)
  context <context_id>           - Look up what triggered a state change
  watch <entity_id> [seconds]    - Live subscribe to state changes (default 60s)

Registry Commands:
  entities [pattern]             - List/search entity registry
  devices [pattern]              - List/search device registry
  areas                          - List all areas

Automation Debugging:
  traces [automation_id]         - List automation traces
  trace <run_id> [automation_id] - Get detailed trace for a run
  trace-vars <run_id> [auto_id]  - Show evaluated variables from trace
  trace-timeline <run_id> [id]   - Step-by-step execution timeline
  trace-trigger <run_id> [id]    - Show trigger context details
  trace-actions <run_id> [id]    - Show action results
  trace-debug <run_id> [id]      - Comprehensive debug view (all info)
  automation-config <entity_id>  - Get automation configuration
  blueprint-inputs <entity_id>   - Validate blueprint inputs vs expected

Time Filtering Options (for logbook, history, history-full, attrs, timeline):
  --from "YYYY-MM-DD HH:MM"      - Start time (instead of hours ago)
  --to "YYYY-MM-DD HH:MM"        - End time (default: now)

Output Format Options (for AI agent context efficiency):
  --output=json                  - Machine-readable JSON (most context-efficient)
  --output=compact               - Reduced verbosity, single-line entries
  --output=default               - Human-readable formatted output
  --json                         - Shorthand for --output=json
  --compact                      - Shorthand for --output=compact
  --no-headers                   - Hide section headers/titles
  --no-timestamps                - Hide timestamps in output
  --max-items=N                  - Limit output to N items

Examples:
  npx tsx ha-ws-client.ts state sun.sun
  npx tsx ha-ws-client.ts call light turn_on '{"entity_id":"light.kitchen"}'
  npx tsx ha-ws-client.ts attrs light.kitchen 4
  npx tsx ha-ws-client.ts watch binary_sensor.motion 30
  npx tsx ha-ws-client.ts blueprint-inputs automation.bathroom_lights
  npx tsx ha-ws-client.ts trace-vars 01KDQS4E2WHMYJYYXKC7K28XFG
  npx tsx ha-ws-client.ts trace-debug 01KDQS4E2WHMYJYYXKC7K28XFG
  echo "{{ now() }}" | npx tsx ha-ws-client.ts template -
`);
  process.exit(0);
}

// =============================================================================
// Main Entry Point
// =============================================================================

/**
 * Main entry point - establishes WebSocket connection and handles messages.
 * Sets up event handlers for authentication, message routing, and errors.
 */
async function main(): Promise<void> {
  const ws = new WebSocket(WS_URL);

  const timeoutId = setTimeout(() => {
    console.error('Timeout');
    ws.close();
    process.exit(1);
  }, DEFAULT_TIMEOUT_MS);

  ws.on('error', (err: Error) => {
    clearTimeout(timeoutId);
    console.error('WebSocket error:', err.message);
    process.exit(1);
  });

  ws.on('close', () => {
    clearTimeout(timeoutId);
  });

  ws.on('message', async (data: WebSocket.Data) => {
    let msg: HAMessage;
    try {
      msg = JSON.parse(data.toString()) as HAMessage;
    } catch {
      console.error('Invalid JSON received from WebSocket');
      return;
    }

    switch (msg.type) {
      case 'auth_required':
        ws.send(JSON.stringify({ type: 'auth', access_token: TOKEN }));
        break;

      case 'auth_ok':
        try {
          await executeCommand(ws);
        } catch (err) {
          if (err instanceof HAClientError) {
            outputError(err.message, err.code);
          } else {
            outputError(err instanceof Error ? err.message : String(err));
          }
        }
        clearTimeout(timeoutId);
        ws.close();
        process.exit(0);
        break;

      case 'auth_invalid':
        clearTimeout(timeoutId);
        console.error('Authentication failed:', msg.message);
        process.exit(1);
        break;

      case 'result': {
        const msgId = msg.id;
        if (msgId === undefined) break;

        const pending = pendingRequests.get(msgId);
        if (pending) {
          pendingRequests.delete(msgId);
          if (msg.success) {
            pending.resolve(msg.result);
          } else {
            pending.reject(
              new HAClientError(msg.error?.message ?? 'Unknown error', msg.error?.code)
            );
          }
        }
        break;
      }

      case 'pong': {
        // Handle pong response for ping command
        const msgId = msg.id;
        if (msgId === undefined) break;

        const pending = pendingRequests.get(msgId);
        if (pending) {
          pendingRequests.delete(msgId);
          pending.resolve('pong');
        }
        break;
      }
    }
  });
}

/**
 * Execute the requested command using the registered handler.
 * Looks up the command in the handler registry and invokes it with context.
 *
 * @param ws - Active WebSocket connection to Home Assistant
 * @throws {Error} If the command is not recognized
 */
async function executeCommand(ws: WebSocket): Promise<void> {
  if (!command) {
    console.error('No command specified');
    process.exit(1);
  }
  const handler = commandHandlers[command];
  if (!handler) {
    console.error(`Unknown command: ${command}`);
    process.exit(1);
  }

  const ctx: CommandContext = {
    ws,
    args: filteredArgs,
    fromTime: globalFromTime,
    toTime: globalToTime,
  };

  await handler(ctx);
}

// Start the client
main();
