/**
 * Automation debugging command handlers.
 * @module handlers/automation
 */

import * as fs from 'node:fs';
import * as path from 'node:path';
import type WebSocket from 'ws';
import { nextId, pendingRequests, sendMessage } from '../client.js';
import type {
  AutomationConfig,
  BlueprintInput,
  CommandContext,
  HAMessage,
  HAState,
  LogbookEntry,
  TraceDetail,
  TraceInfo,
} from '../types.js';
import { getYamlModule } from '../utils.js';

/** Default number of hours for history queries. */
const DEFAULT_HOURS = 24;

/** Default duration for watch command in seconds. */
const DEFAULT_WATCH_SECONDS = 60;

/**
 * List automation traces.
 * Shows recent execution traces for automations.
 *
 * @param ctx - Command context with WebSocket and arguments
 *
 * @example
 * ```bash
 * npx tsx ha-ws-client.ts traces
 * npx tsx ha-ws-client.ts traces automation.bathroom_lights
 * ```
 */
export async function handleTraces(ctx: CommandContext): Promise<void> {
  let automationId = ctx.args[1];
  const result = await sendMessage<TraceInfo[]>(ctx.ws, 'trace/list', { domain: 'automation' });

  // If automationId looks like an entity_id, resolve it to the numeric item_id
  if (automationId?.includes('_')) {
    const entityId = automationId.startsWith('automation.')
      ? automationId
      : `automation.${automationId}`;
    const states = await sendMessage<HAState[]>(ctx.ws, 'get_states');
    const automationEntity = states.find((s) => s.entity_id === entityId);
    if (automationEntity?.attributes?.id) {
      automationId = String(automationEntity.attributes.id);
    }
  }

  let filtered = result;
  if (automationId) {
    filtered = result.filter(
      (t) => t.item_id === automationId || t.item_id === automationId.replace('automation.', '')
    );
  }

  console.log(
    `Automation traces${automationId ? ` for ${automationId}` : ''}: ${filtered.length}\n`
  );
  for (const t of filtered.slice(0, 15)) {
    const when = new Date(t.timestamp.start).toLocaleString();
    const state = t.state ?? 'unknown';
    const error = t.script_execution === 'error' ? ' [ERROR]' : '';
    console.log(`  [${when}] ${t.item_id}: ${state}${error}`);
    if (t.run_id) console.log(`    run_id: ${t.run_id}`);
  }
  if (filtered.length > 15) {
    console.log(`\n  ... and ${filtered.length - 15} more traces`);
  }
}

/**
 * Get detailed trace for an automation run.
 * Shows execution path, errors, and evaluated values.
 *
 * @param ctx - Command context with WebSocket and arguments
 *
 * @example
 * ```bash
 * npx tsx ha-ws-client.ts trace 01KDQS4E2WHMYJYYXKC7K28XFG
 * npx tsx ha-ws-client.ts trace 01KDQS4E2WHMYJYYXKC7K28XFG bathroom_lights
 * ```
 */
export async function handleTrace(ctx: CommandContext): Promise<void> {
  const runId = ctx.args[1];
  let itemId = ctx.args[2];

  if (!runId) {
    console.error('Usage: trace <run_id> [automation_id]');
    console.error('  run_id: The run ID from traces command');
    console.error('  automation_id: Optional automation ID (will auto-detect if not provided)');
    process.exit(1);
  }

  // If no item_id provided, try to find it from traces list
  if (!itemId) {
    const traces = await sendMessage<TraceInfo[]>(ctx.ws, 'trace/list', { domain: 'automation' });
    const match = traces.find((t) => t.run_id === runId);
    if (match) {
      itemId = match.item_id;
    } else {
      console.error('Could not find automation for run_id. Please provide automation_id.');
      process.exit(1);
    }
  }

  // Normalize item_id (remove automation. prefix if present)
  itemId = itemId.replace('automation.', '');

  const result = await sendMessage<TraceDetail>(ctx.ws, 'trace/get', {
    domain: 'automation',
    item_id: itemId,
    run_id: runId,
  });

  console.log(`Trace for run: ${runId}`);
  console.log(`Automation: ${itemId}`);
  console.log(`State: ${result.script_execution ?? 'unknown'}`);

  if (result.error) {
    console.log('\nERROR:', result.error);
  }

  // Look for errors and variables in trace steps
  if (result.trace) {
    for (const [tracePath, steps] of Object.entries(result.trace)) {
      for (const step of steps) {
        if (step.error) {
          console.log(`\nError at ${tracePath}:`);
          console.log(JSON.stringify(step.error, null, 2));
        }
        if (step.result?.error) {
          console.log(`\nResult error at ${tracePath}:`);
          console.log(JSON.stringify(step.result.error, null, 2));
        }
        const varKeys = step.variables ? Object.keys(step.variables) : [];
        if (varKeys.length > 0 && varKeys.length < 20) {
          console.log(`\nVariables at ${tracePath}:`, varKeys.join(', '));
        }
      }
    }
  }

  if (result.config?.trigger) {
    console.log('\nTrigger config:', JSON.stringify(result.config.trigger, null, 2));
  }

  if (result.context) {
    console.log('\nContext:', JSON.stringify(result.context, null, 2));
  }
}

/**
 * Show evaluated variables from an automation trace.
 * Displays variables organized by type (boolean, numeric, string, other).
 *
 * @param ctx - Command context with WebSocket and arguments
 *
 * @example
 * ```bash
 * npx tsx ha-ws-client.ts trace-vars 01KDQS4E2WHMYJYYXKC7K28XFG
 * ```
 */
export async function handleTraceVars(ctx: CommandContext): Promise<void> {
  const runId = ctx.args[1];
  let itemId = ctx.args[2];

  if (!runId) {
    console.error('Usage: trace-vars <run_id> [automation_id]');
    process.exit(1);
  }

  // If no item_id provided, try to find it from traces list
  if (!itemId) {
    const traces = await sendMessage<TraceInfo[]>(ctx.ws, 'trace/list', { domain: 'automation' });
    const match = traces.find((t) => t.run_id === runId);
    if (match) {
      itemId = match.item_id;
    } else {
      console.error('Could not find automation for run_id. Please provide automation_id.');
      process.exit(1);
    }
  }

  itemId = itemId.replace('automation.', '');

  const result = await sendMessage<TraceDetail>(ctx.ws, 'trace/get', {
    domain: 'automation',
    item_id: itemId,
    run_id: runId,
  });

  console.log(`Trace variables for: ${itemId}`);
  console.log(`Run ID: ${runId}`);
  console.log(`State: ${result.script_execution ?? 'unknown'}\n`);

  const allVars = new Map<string, unknown>();

  if (result.trace) {
    for (const steps of Object.values(result.trace)) {
      for (const step of steps) {
        if (step.variables && Object.keys(step.variables).length > 0) {
          for (const [k, v] of Object.entries(step.variables)) {
            allVars.set(k, v);
          }
        }
      }
    }
  }

  const skipVars = new Set(['trigger', 'this', 'context']);
  const importantVars = [...allVars.entries()].filter(([k]) => !skipVars.has(k));

  const boolVars = importantVars.filter(([, v]) => typeof v === 'boolean');
  const numVars = importantVars.filter(([, v]) => typeof v === 'number');
  const strVars = importantVars.filter(([, v]) => typeof v === 'string');
  const otherVars = importantVars.filter(
    ([, v]) => typeof v !== 'boolean' && typeof v !== 'number' && typeof v !== 'string'
  );

  if (boolVars.length > 0) {
    console.log('Boolean variables:');
    for (const [k, v] of boolVars) {
      const icon = v ? '[ok]' : '[x]';
      console.log(`  ${icon} ${k}: ${v}`);
    }
    console.log('');
  }

  if (numVars.length > 0) {
    console.log('Numeric variables:');
    for (const [k, v] of numVars) {
      console.log(`  ${k}: ${v}`);
    }
    console.log('');
  }

  if (strVars.length > 0) {
    console.log('String variables:');
    for (const [k, v] of strVars) {
      const str = v as string;
      const display = str.length > 60 ? `${str.substring(0, 60)}...` : str;
      console.log(`  ${k}: ${display}`);
    }
    console.log('');
  }

  if (otherVars.length > 0) {
    console.log('Other variables:');
    for (const [k, v] of otherVars) {
      const display = JSON.stringify(v);
      const truncated = display.length > 80 ? `${display.substring(0, 80)}...` : display;
      console.log(`  ${k}: ${truncated}`);
    }
    console.log('');
  }

  if (allVars.has('trigger')) {
    interface TriggerInfo {
      id?: string;
      entity_id?: string;
      from_state?: { state: string };
      to_state?: { state: string };
    }
    const trigger = allVars.get('trigger') as TriggerInfo;
    console.log('Trigger info:');
    if (trigger.id) console.log(`  id: ${trigger.id}`);
    if (trigger.entity_id) console.log(`  entity: ${trigger.entity_id}`);
    if (trigger.from_state) console.log(`  from: ${trigger.from_state.state}`);
    if (trigger.to_state) console.log(`  to: ${trigger.to_state.state}`);
  }

  if (importantVars.length === 0) {
    console.log('No variables captured in this trace.');
    console.log('(Variables are captured at choose/condition steps)');
  }
}

/**
 * Get automation configuration.
 * Returns the raw configuration for an automation.
 *
 * @param ctx - Command context with WebSocket and arguments
 *
 * @example
 * ```bash
 * npx tsx ha-ws-client.ts automation-config automation.bathroom_lights
 * ```
 */
export async function handleAutomationConfig(ctx: CommandContext): Promise<void> {
  const entityId = ctx.args[1];
  if (!entityId) {
    console.error('Usage: automation-config <entity_id>');
    process.exit(1);
  }

  interface AutomationConfigResult {
    readonly config: Record<string, unknown>;
  }
  const result = await sendMessage<AutomationConfigResult>(ctx.ws, 'automation/config', {
    entity_id: entityId.startsWith('automation.') ? entityId : `automation.${entityId}`,
  });

  console.log(`Configuration for ${entityId}:\n`);
  console.log(JSON.stringify(result.config, null, 2));
}

/**
 * Look up what caused a state change by context ID.
 * Searches entities, logbook, and traces for the context.
 *
 * @param ctx - Command context with WebSocket and arguments
 *
 * @example
 * ```bash
 * npx tsx ha-ws-client.ts context 01KDQS4E2WHMYJYYXKC7K28XFG
 * ```
 */
export async function handleContext(ctx: CommandContext): Promise<void> {
  const contextId = ctx.args[1];
  if (!contextId) {
    console.error('Usage: context <context_id>');
    console.error('Example: context 01KDQS4E2WHMYJYYXKC7K28XFG');
    console.error('\nContext IDs can be found in the "context" field of entity states');
    process.exit(1);
  }

  const states = await sendMessage<HAState[]>(ctx.ws, 'get_states');

  const matchingEntity = states.find((s) => s.context && s.context.id === contextId);

  console.log(`Context lookup for: ${contextId}\n`);

  if (matchingEntity) {
    console.log(`Current entity with this context:`);
    console.log(`  Entity: ${matchingEntity.entity_id}`);
    console.log(`  State: ${matchingEntity.state}`);
    console.log(`  Last changed: ${new Date(matchingEntity.last_changed ?? '').toLocaleString()}`);
    if (matchingEntity.context?.parent_id) {
      console.log(`  Parent context: ${matchingEntity.context.parent_id}`);
    }
    if (matchingEntity.context?.user_id) {
      console.log(`  User ID: ${matchingEntity.context.user_id}`);
    }
  }

  const endTime = new Date();
  const startTime = new Date(endTime.getTime() - DEFAULT_HOURS * 3_600_000);

  const logbookResult = await sendMessage<LogbookEntry[]>(ctx.ws, 'logbook/get_events', {
    start_time: startTime.toISOString(),
    end_time: endTime.toISOString(),
    context_id: contextId,
  });

  if (logbookResult.length > 0) {
    console.log(`\nLogbook entries with this context:`);
    for (const entry of logbookResult) {
      const when = new Date(entry.when * 1000).toLocaleString();
      const entity = entry.entity_id ?? 'unknown';
      const state = entry.state ?? entry.message ?? 'event';
      console.log(`  ${when}: ${entity} -> ${state}`);
    }
  }

  try {
    const traces = await sendMessage<TraceInfo[]>(ctx.ws, 'trace/list', { domain: 'automation' });
    const matchingTraces = traces.filter(
      (t) => t.run_id === contextId || t.context?.id === contextId
    );

    if (matchingTraces.length > 0) {
      console.log(`\nMatching automation traces:`);
      for (const t of matchingTraces) {
        const when = new Date(t.timestamp.start).toLocaleString();
        console.log(`  ${when}: automation.${t.item_id}`);
        console.log(`    Run ID: ${t.run_id}`);
        console.log(`    State: ${t.state ?? t.script_execution ?? 'unknown'}`);
      }
    }
  } catch {
    // Trace lookup may fail, that's ok
  }

  if (!matchingEntity && logbookResult.length === 0) {
    console.log(`No current data found for context ID: ${contextId}`);
    console.log('(Context may have expired or entity state has changed since then)');
  }
}

/**
 * Validate blueprint inputs against expected inputs.
 * Compares provided inputs with the blueprint definition.
 *
 * @param ctx - Command context with WebSocket and arguments
 *
 * @example
 * ```bash
 * npx tsx ha-ws-client.ts blueprint-inputs automation.bathroom_lights
 * ```
 */
export async function handleBlueprintInputs(ctx: CommandContext): Promise<void> {
  const entityId = ctx.args[1];
  if (!entityId) {
    console.error('Usage: blueprint-inputs <automation_entity_id>');
    process.exit(1);
  }

  const fullEntityId = entityId.startsWith('automation.') ? entityId : `automation.${entityId}`;

  const states = await sendMessage<HAState[]>(ctx.ws, 'get_states');
  const automationEntity = states.find((s) => s.entity_id === fullEntityId);

  if (!automationEntity) {
    console.error(`Automation not found: ${fullEntityId}`);
    process.exit(1);
  }

  const automationId = automationEntity.attributes?.id as string | undefined;
  if (!automationId) {
    console.error(`Could not find automation ID for ${fullEntityId}`);
    process.exit(1);
  }

  const yaml = getYamlModule();

  let automationsContent: string;
  try {
    automationsContent = fs.readFileSync('/config/automations.yaml', 'utf8');
  } catch {
    console.error('Could not read automations.yaml');
    process.exit(1);
  }

  const automations = yaml.load(automationsContent) as AutomationConfig[];
  const automation = automations.find((a) => a.id === automationId);

  if (!automation) {
    console.error(`Automation ID ${automationId} not found in automations.yaml`);
    process.exit(1);
  }

  if (!automation.use_blueprint) {
    console.log(`${fullEntityId} is not a blueprint-based automation`);
    process.exit(0);
  }

  const blueprintPath = automation.use_blueprint.path;
  const providedInputs = automation.use_blueprint.input || {};

  console.log(`Blueprint: ${blueprintPath}`);
  console.log(`\nProvided inputs:`);

  const blueprintFile = path.join('/config/blueprints/automation', blueprintPath);

  const expectedInputs: Record<string, { name: string; default?: unknown; description?: string }> =
    {};
  let blueprintFound = false;

  try {
    if (fs.existsSync(blueprintFile)) {
      const blueprintContent = fs.readFileSync(blueprintFile, 'utf8');
      const InputType = new yaml.Type('!input', {
        kind: 'scalar',
        construct: (data: string) => ({ __input: data }),
      });
      const IncludeType = new yaml.Type('!include', {
        kind: 'scalar',
        construct: (data: string) => ({ __include: data }),
      });
      const HA_SCHEMA = yaml.DEFAULT_SCHEMA.extend([InputType, IncludeType]);
      const blueprint = yaml.load(blueprintContent, { schema: HA_SCHEMA }) as {
        blueprint?: { input?: Record<string, BlueprintInput> };
      };
      blueprintFound = true;

      const extractInputs = (obj: Record<string, BlueprintInput>): void => {
        if (!obj) return;
        for (const [key, value] of Object.entries(obj)) {
          if (value && typeof value === 'object') {
            if (value.input && typeof value.input === 'object') {
              extractInputs(value.input);
            } else if (value.selector || value.default !== undefined || value.name) {
              const inputInfo: { name: string; default?: unknown; description?: string } = {
                name: value.name || key,
              };
              if (value.default !== undefined) {
                inputInfo.default = value.default;
              }
              if (value.description !== undefined) {
                inputInfo.description = value.description;
              }
              expectedInputs[key] = inputInfo;
            }
          }
        }
      };

      if (blueprint.blueprint?.input) {
        extractInputs(blueprint.blueprint.input);
      }
    }
  } catch {
    // YAML parsing failed, continue without blueprint comparison
  }

  const providedKeys = Object.keys(providedInputs);
  const expectedKeys = Object.keys(expectedInputs);

  const expectedKeySet = new Set(expectedKeys);
  const unknownInputs = providedKeys.filter(
    (k) => expectedKeys.length > 0 && !expectedKeySet.has(k)
  );

  for (const key of providedKeys) {
    const value = providedInputs[key];
    const displayValue = typeof value === 'object' ? JSON.stringify(value) : String(value);
    const isUnknown = unknownInputs.includes(key);
    const marker = isUnknown ? ' [UNKNOWN INPUT]' : '';
    console.log(`  ${key}: ${displayValue}${marker}`);
  }

  if (blueprintFound && expectedKeys.length > 0) {
    console.log(`\nExpected inputs from blueprint:`);
    const providedKeySet = new Set(providedKeys);
    for (const key of expectedKeys) {
      const info = expectedInputs[key];
      if (!info) continue;
      const provided = providedKeySet.has(key);
      const marker = provided ? '[ok]' : info.default !== undefined ? '(default)' : 'MISSING';
      const defaultStr =
        info.default !== undefined ? ` [default: ${JSON.stringify(info.default)}]` : '';
      console.log(`  ${marker} ${key}${defaultStr}`);
    }

    if (unknownInputs.length > 0) {
      console.log(`\nUnknown inputs (not in blueprint):`);
      for (const key of unknownInputs) {
        console.log(`  - ${key}`);
        const keyPrefix = key.toLowerCase().split('_')[0] ?? '';
        const similar = expectedKeys.filter((ek) => {
          const ekPrefix = ek.toLowerCase().split('_')[0] ?? '';
          return ek.toLowerCase().includes(keyPrefix) || key.toLowerCase().includes(ekPrefix);
        });
        if (similar.length > 0) {
          console.log(`    Did you mean: ${similar.join(', ')}?`);
        }
      }
    }
  } else if (!blueprintFound) {
    console.log(`\n(Blueprint file not found at ${blueprintFile} for validation)`);
  }
}

/**
 * Live subscribe to entity state changes.
 * Watches an entity and displays state changes in real-time.
 *
 * @param ctx - Command context with WebSocket and arguments
 *
 * @example
 * ```bash
 * npx tsx ha-ws-client.ts watch light.kitchen 30
 * npx tsx ha-ws-client.ts watch binary_sensor.motion 60
 * ```
 */
export async function handleWatch(ctx: CommandContext): Promise<void> {
  const entityId = ctx.args[1];
  const seconds = parseInt(ctx.args[2] as string, 10) || DEFAULT_WATCH_SECONDS;

  if (!entityId) {
    console.error('Usage: watch <entity_id> [seconds]');
    console.error('Example: watch light.kitchen 30');
    process.exit(1);
  }

  console.log(`Watching ${entityId} for ${seconds} seconds...`);
  console.log('Press Ctrl+C to stop early.\n');

  const subId = nextId();
  const subPromise = new Promise<void>((resolve, reject) => {
    pendingRequests.set(subId, { resolve: resolve as (value: unknown) => void, reject });
  });

  ctx.ws.send(
    JSON.stringify({
      id: subId,
      type: 'subscribe_trigger',
      trigger: {
        platform: 'state',
        entity_id: entityId,
      },
    })
  );

  await subPromise;

  const states = await sendMessage<HAState[]>(ctx.ws, 'get_states');
  const initialState = states.find((s) => s.entity_id === entityId);
  if (initialState) {
    console.log(`Initial: ${initialState.state}`);
    const attrs = initialState.attributes ?? {};
    const importantAttrs = [
      'brightness',
      'color_temp',
      'temperature',
      'hvac_action',
      'position',
      'percentage',
    ] as const;
    const attrStr = importantAttrs
      .filter((a) => attrs[a] !== undefined)
      .map((a) => `${a}=${attrs[a]}`)
      .join(', ');
    if (attrStr) console.log(`  (${attrStr})`);
    console.log('');
  }

  let eventCount = 0;
  const eventHandler = (data: WebSocket.Data): void => {
    let msg: HAMessage;
    try {
      msg = JSON.parse(data.toString()) as HAMessage;
    } catch {
      return;
    }
    if (msg.type === 'event' && msg.event?.variables) {
      eventCount++;
      interface WatchTrigger {
        trigger?: {
          from_state?: { state: string; attributes?: Record<string, unknown> };
          to_state?: { state: string; attributes?: Record<string, unknown> };
        };
      }
      const vars = msg.event.variables as WatchTrigger;
      const when = new Date().toLocaleTimeString();

      let output = `[${when}] `;
      if (vars.trigger?.from_state && vars.trigger?.to_state) {
        const from = vars.trigger.from_state;
        const to = vars.trigger.to_state;
        output += `${from.state} -> ${to.state}`;

        const fromAttrs = from.attributes ?? {};
        const toAttrs = to.attributes ?? {};
        const changedAttrs: string[] = [];
        const watchAttrs = [
          'brightness',
          'color_temp',
          'temperature',
          'hvac_action',
          'position',
          'percentage',
          'preset_mode',
        ];
        for (const attr of watchAttrs) {
          if (fromAttrs[attr] !== toAttrs[attr] && toAttrs[attr] !== undefined) {
            changedAttrs.push(`${attr}: ${fromAttrs[attr]} -> ${toAttrs[attr]}`);
          }
        }
        if (changedAttrs.length > 0) {
          output += `\n         ${changedAttrs.join(', ')}`;
        }
      }
      console.log(output);
    }
  };

  ctx.ws.on('message', eventHandler);

  await new Promise((resolve) => setTimeout(resolve, seconds * 1000));

  ctx.ws.removeListener('message', eventHandler);
  console.log(`\nWatched for ${seconds}s, captured ${eventCount} state change(s).`);
}
