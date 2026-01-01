/**
 * Automation debugging command handlers.
 * @module handlers/automation
 */

import * as fs from 'node:fs';
import * as path from 'node:path';
import { sendMessage, subscribeToTrigger } from '../client.js';
import { getOutputConfig, isJsonOutput, output, outputList, outputMessage } from '../output.js';
import type {
  AutomationConfig,
  AutomationConfigResult,
  BlueprintInput,
  CommandContext,
  HAState,
  LogbookEntry,
  StateTriggerVariables,
  TraceDetail,
  TraceInfo,
  TraceStep,
  TraceTrigger,
  TriggerInfo,
} from '../types.js';
import { calculateTimeRange, getYamlModule, requireArg } from '../utils.js';

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

  const { format, maxItems } = getOutputConfig();
  const limit = maxItems > 0 ? maxItems : 15;

  if (format === 'json') {
    output(
      filtered.slice(0, limit).map((t) => ({
        item_id: t.item_id,
        run_id: t.run_id,
        state: t.state,
        script_execution: t.script_execution,
        timestamp: t.timestamp.start,
        context_id: t.context?.id,
      })),
      { command: 'traces', count: filtered.length }
    );
    return;
  }

  outputList(filtered.slice(0, limit), {
    title: `Automation traces${automationId ? ` for ${automationId}` : ''}`,
    command: 'traces',
    itemFormatter: (t) => {
      const when = new Date(t.timestamp.start).toLocaleString();
      const state = t.state ?? 'unknown';
      const error = t.script_execution === 'error' ? ' [ERROR]' : '';
      if (format === 'compact') {
        return `${t.item_id} ${t.run_id} ${state}${error}`;
      }
      let line = `  [${when}] ${t.item_id}: ${state}${error}`;
      if (t.run_id) line += `\n    run_id: ${t.run_id}`;
      return line;
    },
  });
  if (filtered.length > limit) {
    outputMessage(`\n  ... and ${filtered.length - limit} more traces`);
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
  const runId = requireArg(
    ctx,
    1,
    'Usage: trace <run_id> [automation_id]\n' +
      '  run_id: The run ID from traces command\n' +
      '  automation_id: Optional automation ID (will auto-detect if not provided)'
  );
  let itemId = ctx.args[2];

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

  if (isJsonOutput()) {
    output(
      {
        run_id: runId,
        automation_id: itemId,
        state: result.script_execution,
        error: result.error,
        trace: result.trace,
        trigger: result.trigger,
        context: result.context,
        timestamp: result.timestamp,
      },
      { command: 'trace' }
    );
    return;
  }

  outputMessage(`Trace for run: ${runId}`);
  outputMessage(`Automation: ${itemId}`);
  outputMessage(`State: ${result.script_execution ?? 'unknown'}`);

  if (result.error) {
    outputMessage(`\nERROR: ${result.error}`);
  }

  // Look for errors and variables in trace steps
  if (result.trace) {
    for (const [tracePath, steps] of Object.entries(result.trace)) {
      for (const step of steps) {
        if (step.error) {
          outputMessage(`\nError at ${tracePath}:`);
          console.log(JSON.stringify(step.error, null, 2));
        }
        if (step.result?.error) {
          outputMessage(`\nResult error at ${tracePath}:`);
          console.log(JSON.stringify(step.result.error, null, 2));
        }
        const varKeys = step.variables ? Object.keys(step.variables) : [];
        if (varKeys.length > 0 && varKeys.length < 20) {
          outputMessage(`\nVariables at ${tracePath}: ${varKeys.join(', ')}`);
        }
      }
    }
  }

  if (result.config?.trigger) {
    outputMessage('\nTrigger config:');
    console.log(JSON.stringify(result.config.trigger, null, 2));
  }

  if (result.context) {
    outputMessage('\nContext:');
    console.log(JSON.stringify(result.context, null, 2));
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
  const runId = requireArg(ctx, 1, 'Usage: trace-vars <run_id> [automation_id]');
  let itemId = ctx.args[2];

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
  const entityId = requireArg(ctx, 1, 'Usage: automation-config <entity_id>');

  const result = await sendMessage<AutomationConfigResult>(ctx.ws, 'automation/config', {
    entity_id: entityId.startsWith('automation.') ? entityId : `automation.${entityId}`,
  });

  if (isJsonOutput()) {
    output({ entity_id: entityId, config: result.config }, { command: 'automation-config' });
    return;
  }

  outputMessage(`Configuration for ${entityId}:\n`);
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
  const contextId = requireArg(
    ctx,
    1,
    'Usage: context <context_id>\n' +
      'Example: context 01KDQS4E2WHMYJYYXKC7K28XFG\n' +
      '\nContext IDs can be found in the "context" field of entity states'
  );

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

  const { startTime, endTime } = calculateTimeRange(null, null, DEFAULT_HOURS);

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
  const entityId = requireArg(ctx, 1, 'Usage: blueprint-inputs <automation_entity_id>');

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
  const entityId = requireArg(
    ctx,
    1,
    'Usage: watch <entity_id> [seconds]\n' + 'Example: watch light.kitchen 30'
  );
  const seconds = parseInt(ctx.args[2] as string, 10) || DEFAULT_WATCH_SECONDS;

  console.log(`Watching ${entityId} for ${seconds} seconds...`);
  console.log('Press Ctrl+C to stop early.\n');

  // Get and display initial state
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

  // Subscribe to state changes using the helper function
  const { cleanup } = await subscribeToTrigger(
    ctx.ws,
    { platform: 'state', entity_id: entityId },
    (variables) => {
      eventCount++;
      const vars = variables as StateTriggerVariables;
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
  );

  // Wait for the watch duration
  await new Promise((resolve) => setTimeout(resolve, seconds * 1000));

  // Clean up the subscription
  cleanup();
  console.log(`\nWatched for ${seconds}s, captured ${eventCount} state change(s).`);
}

/**
 * Helper function to get trace detail with automatic item_id resolution.
 *
 * @param ctx - Command context with WebSocket and arguments
 * @param runId - The run ID of the trace
 * @param providedItemId - Optional item ID (will be auto-detected if not provided)
 * @returns The trace detail
 */
async function getTraceDetail(
  ctx: CommandContext,
  runId: string,
  providedItemId?: string
): Promise<{ trace: TraceDetail; itemId: string }> {
  let itemId = providedItemId;

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

  return { trace: result, itemId };
}

/**
 * Format a trace step path into a human-readable description.
 *
 * @param tracePath - The trace path (e.g., "action/0", "condition/1/if/0")
 * @returns Human-readable description
 */
function formatTracePath(tracePath: string): string {
  const parts = tracePath.split('/');
  const descriptions: string[] = [];

  for (let i = 0; i < parts.length; i++) {
    const part = parts[i];
    const nextPart = parts[i + 1];

    if (part === 'trigger' && nextPart !== undefined) {
      descriptions.push(`Trigger ${nextPart}`);
      i++;
    } else if (part === 'condition' && nextPart !== undefined) {
      descriptions.push(`Condition ${nextPart}`);
      i++;
    } else if (part === 'action' && nextPart !== undefined) {
      descriptions.push(`Action ${nextPart}`);
      i++;
    } else if (part === 'sequence' && nextPart !== undefined) {
      descriptions.push(`Step ${nextPart}`);
      i++;
    } else if (part === 'if' && nextPart !== undefined) {
      descriptions.push(`If branch ${nextPart}`);
      i++;
    } else if (part === 'else' && nextPart !== undefined) {
      descriptions.push(`Else branch ${nextPart}`);
      i++;
    } else if (part === 'then' && nextPart !== undefined) {
      descriptions.push(`Then ${nextPart}`);
      i++;
    } else if (part === 'choose' && nextPart !== undefined) {
      descriptions.push(`Choose option ${nextPart}`);
      i++;
    } else if (part === 'default' && nextPart !== undefined) {
      descriptions.push(`Default action ${nextPart}`);
      i++;
    } else if (part === 'repeat' && nextPart !== undefined) {
      descriptions.push(`Repeat ${nextPart}`);
      i++;
    } else if (part === 'parallel' && nextPart !== undefined) {
      descriptions.push(`Parallel ${nextPart}`);
      i++;
    }
  }

  return descriptions.length > 0 ? descriptions.join(' > ') : tracePath;
}

/**
 * Show step-by-step execution timeline for an automation trace.
 * Displays each step with timestamps, duration, and execution results.
 *
 * @param ctx - Command context with WebSocket and arguments
 *
 * @example
 * ```bash
 * npx tsx ha-ws-client.ts trace-timeline 01KDQS4E2WHMYJYYXKC7K28XFG
 * ```
 */
export async function handleTraceTimeline(ctx: CommandContext): Promise<void> {
  const runId = requireArg(
    ctx,
    1,
    'Usage: trace-timeline <run_id> [automation_id]\n' +
      '  run_id: The run ID from traces command\n' +
      '  automation_id: Optional automation ID (will auto-detect if not provided)'
  );
  const providedItemId = ctx.args[2];

  const { trace: result, itemId } = await getTraceDetail(ctx, runId, providedItemId);

  console.log(`Execution Timeline for: automation.${itemId}`);
  console.log(`Run ID: ${runId}`);
  console.log(`Status: ${result.script_execution ?? 'unknown'}`);

  if (result.timestamp) {
    const start = new Date(result.timestamp.start);
    console.log(`Started: ${start.toLocaleString()}`);
    if (result.timestamp.finish) {
      const finish = new Date(result.timestamp.finish);
      const durationMs = finish.getTime() - start.getTime();
      console.log(`Duration: ${durationMs}ms`);
    }
  }

  if (result.error) {
    console.log(`\n[ERROR] ${result.error}`);
  }

  console.log('\nSteps:');

  if (!result.trace || Object.keys(result.trace).length === 0) {
    console.log('No trace steps recorded.');
    return;
  }

  // Collect all steps with their paths and timestamps for sorting
  const allSteps: Array<{ path: string; step: TraceStep; index: number }> = [];

  for (const [tracePath, steps] of Object.entries(result.trace)) {
    for (let i = 0; i < steps.length; i++) {
      const step = steps[i];
      if (step) {
        allSteps.push({ path: tracePath, step, index: i });
      }
    }
  }

  // Sort by timestamp if available
  allSteps.sort((a, b) => {
    const timeA = a.step.timestamp ? new Date(a.step.timestamp).getTime() : 0;
    const timeB = b.step.timestamp ? new Date(b.step.timestamp).getTime() : 0;
    return timeA - timeB;
  });

  let prevTime: Date | null = null;

  for (const { path, step } of allSteps) {
    const pathDesc = formatTracePath(path);
    const timestamp = step.timestamp ? new Date(step.timestamp) : null;
    const timeStr = timestamp
      ? timestamp.toLocaleTimeString('en-US', { hour12: false })
      : '??:??:??';

    // Calculate delta from previous step
    let delta = '';
    if (timestamp && prevTime) {
      const deltaMs = timestamp.getTime() - prevTime.getTime();
      if (deltaMs >= 1000) {
        delta = ` (+${(deltaMs / 1000).toFixed(2)}s)`;
      } else {
        delta = ` (+${deltaMs}ms)`;
      }
    }
    prevTime = timestamp;

    // Determine status icon
    let statusIcon = '[ok]';
    if (step.error) {
      statusIcon = '[ERR]';
    } else if (step.result?.error) {
      statusIcon = '[ERR]';
    } else if (step.result?.enabled === false) {
      statusIcon = '[SKIP]';
    }

    console.log(`${statusIcon} [${timeStr}]${delta} ${pathDesc}`);

    // Show errors inline
    if (step.error) {
      const errorStr =
        typeof step.error === 'string' ? step.error : JSON.stringify(step.error, null, 2);
      console.log(`     Error: ${errorStr}`);
    }
    if (step.result?.error) {
      const errorStr =
        typeof step.result.error === 'string'
          ? step.result.error
          : JSON.stringify(step.result.error, null, 2);
      console.log(`     Result error: ${errorStr}`);
    }

    // Show changed variables (not full state, just what changed)
    if (step.changed_variables && Object.keys(step.changed_variables).length > 0) {
      const changedVars = Object.entries(step.changed_variables)
        .filter(([k]) => !['trigger', 'this', 'context'].includes(k))
        .slice(0, 5);
      if (changedVars.length > 0) {
        const varStr = changedVars.map(([k, v]) => `${k}=${JSON.stringify(v)}`).join(', ');
        console.log(`     Changed: ${varStr}`);
      }
    }
  }
}

/**
 * Display detailed trigger context for an automation trace.
 * Shows the trigger type, triggering entity, state changes, and any trigger variables.
 *
 * @param ctx - Command context with WebSocket and arguments
 *
 * @example
 * ```bash
 * npx tsx ha-ws-client.ts trace-trigger 01KDQS4E2WHMYJYYXKC7K28XFG
 * ```
 */
export async function handleTraceTrigger(ctx: CommandContext): Promise<void> {
  const runId = requireArg(
    ctx,
    1,
    'Usage: trace-trigger <run_id> [automation_id]\n' +
      '  run_id: The run ID from traces command\n' +
      '  automation_id: Optional automation ID (will auto-detect if not provided)'
  );
  const providedItemId = ctx.args[2];

  const { trace: result, itemId } = await getTraceDetail(ctx, runId, providedItemId);

  console.log(`Trigger Context for: automation.${itemId}`);
  console.log(`Run ID: ${runId}`);
  console.log(`Status: ${result.script_execution ?? 'unknown'}\n`);

  // Look for trigger info in trace variables
  let triggerInfo: TraceTrigger | null = null;

  // First check the trace for trigger variables (check both variables and changed_variables)
  if (result.trace) {
    for (const steps of Object.values(result.trace)) {
      for (const step of steps) {
        // Check changed_variables first (where trigger info usually appears)
        if (step.changed_variables?.trigger) {
          triggerInfo = step.changed_variables.trigger as TraceTrigger;
          break;
        }
        // Also check variables as fallback
        if (step.variables?.trigger) {
          triggerInfo = step.variables.trigger as TraceTrigger;
          break;
        }
      }
      if (triggerInfo) break;
    }
  }

  // Also use top-level trigger info if available (merge with found info)
  if (result.trigger && typeof result.trigger === 'object') {
    triggerInfo = { ...triggerInfo, ...result.trigger };
  }

  if (!triggerInfo) {
    console.log('No trigger information found in this trace.');
    console.log('(Trigger data may not be captured for all automation types)');
    return;
  }

  console.log('Details:');

  if (triggerInfo.platform) {
    console.log(`Platform: ${triggerInfo.platform}`);
  }

  if (triggerInfo.id !== undefined) {
    console.log(`Trigger ID: ${triggerInfo.id}`);
  }

  if (triggerInfo.idx !== undefined) {
    console.log(`Trigger Index: ${triggerInfo.idx}`);
  }

  if (triggerInfo.alias) {
    console.log(`Alias: ${triggerInfo.alias}`);
  }

  if (triggerInfo.entity_id) {
    console.log(`Entity: ${triggerInfo.entity_id}`);
  }

  if (triggerInfo.description) {
    console.log(`Description: ${triggerInfo.description}`);
  }

  // Show state transition
  if (triggerInfo.from_state || triggerInfo.to_state) {
    console.log('\nState change:');

    if (triggerInfo.from_state) {
      const from = triggerInfo.from_state;
      console.log(`From State:`);
      console.log(`  Value: ${from.state}`);
      if (from.last_changed) {
        console.log(`  Changed: ${new Date(from.last_changed).toLocaleString()}`);
      }
      if (from.attributes && Object.keys(from.attributes).length > 0) {
        const importantAttrs = [
          'brightness',
          'color_temp',
          'temperature',
          'hvac_action',
          'position',
          'percentage',
          'friendly_name',
        ];
        const attrs = Object.entries(from.attributes)
          .filter(([k]) => importantAttrs.includes(k))
          .slice(0, 5);
        if (attrs.length > 0) {
          console.log(`  Attributes: ${attrs.map(([k, v]) => `${k}=${v}`).join(', ')}`);
        }
      }
    }

    if (triggerInfo.to_state) {
      const to = triggerInfo.to_state;
      console.log(`\nTo State:`);
      console.log(`  Value: ${to.state}`);
      if (to.last_changed) {
        console.log(`  Changed: ${new Date(to.last_changed).toLocaleString()}`);
      }
      if (to.attributes && Object.keys(to.attributes).length > 0) {
        const importantAttrs = [
          'brightness',
          'color_temp',
          'temperature',
          'hvac_action',
          'position',
          'percentage',
          'friendly_name',
        ];
        const attrs = Object.entries(to.attributes)
          .filter(([k]) => importantAttrs.includes(k))
          .slice(0, 5);
        if (attrs.length > 0) {
          console.log(`  Attributes: ${attrs.map(([k, v]) => `${k}=${v}`).join(', ')}`);
        }
      }
    }
  }

  // Show "for" duration if present
  if (triggerInfo.for) {
    console.log('\nDuration:');
    if (typeof triggerInfo.for === 'string') {
      console.log(`For: ${triggerInfo.for}`);
    } else {
      const parts: string[] = [];
      if (triggerInfo.for.hours) parts.push(`${triggerInfo.for.hours}h`);
      if (triggerInfo.for.minutes) parts.push(`${triggerInfo.for.minutes}m`);
      if (triggerInfo.for.seconds) parts.push(`${triggerInfo.for.seconds}s`);
      console.log(`For: ${parts.join(' ')}`);
    }
  }

  // Show trigger configuration from automation config
  if (result.config?.trigger && Array.isArray(result.config.trigger)) {
    const idx = triggerInfo.idx !== undefined ? parseInt(triggerInfo.idx, 10) : 0;
    const triggerConfig = result.config.trigger[idx];
    if (triggerConfig) {
      console.log('\nConfig:');
      console.log(JSON.stringify(triggerConfig, null, 2));
    }
  }
}

/**
 * Display action results from an automation trace.
 * Shows each action's execution result, service calls, and any data returned.
 *
 * @param ctx - Command context with WebSocket and arguments
 *
 * @example
 * ```bash
 * npx tsx ha-ws-client.ts trace-actions 01KDQS4E2WHMYJYYXKC7K28XFG
 * ```
 */
export async function handleTraceActions(ctx: CommandContext): Promise<void> {
  const runId = requireArg(
    ctx,
    1,
    'Usage: trace-actions <run_id> [automation_id]\n' +
      '  run_id: The run ID from traces command\n' +
      '  automation_id: Optional automation ID (will auto-detect if not provided)'
  );
  const providedItemId = ctx.args[2];

  const { trace: result, itemId } = await getTraceDetail(ctx, runId, providedItemId);

  console.log(`Action Results for: automation.${itemId}`);
  console.log(`Run ID: ${runId}`);
  console.log(`Status: ${result.script_execution ?? 'unknown'}\n`);

  if (!result.trace || Object.keys(result.trace).length === 0) {
    console.log('No trace steps recorded.');
    return;
  }

  // Filter for action-related trace steps
  const actionPaths = Object.entries(result.trace)
    .filter(([path]) => path.includes('action') || path.includes('sequence'))
    .sort(([a], [b]) => a.localeCompare(b));

  if (actionPaths.length === 0) {
    console.log('No action steps found in this trace.');
    return;
  }

  console.log('Actions:');

  let actionNum = 1;
  for (const [tracePath, steps] of actionPaths) {
    const pathDesc = formatTracePath(tracePath);

    for (const step of steps) {
      // Determine status
      let status = '[ok]';
      if (step.error) {
        status = '[FAIL]';
      } else if (step.result?.error) {
        status = '[FAIL]';
      } else if (step.result?.enabled === false) {
        status = '[SKIP]';
      }

      console.log(`${status} Action ${actionNum}: ${pathDesc}`);

      // Show timestamp
      if (step.timestamp) {
        console.log(`    Time: ${new Date(step.timestamp).toLocaleString()}`);
      }

      // Show result details
      if (step.result) {
        // Show parameters if present
        if (step.result.params && Object.keys(step.result.params).length > 0) {
          console.log(`    Params:`);
          for (const [key, value] of Object.entries(step.result.params)) {
            const valueStr = typeof value === 'object' ? JSON.stringify(value) : String(value);
            const truncated = valueStr.length > 60 ? `${valueStr.substring(0, 60)}...` : valueStr;
            console.log(`      ${key}: ${truncated}`);
          }
        }

        // Show response if present
        if (step.result.response !== undefined) {
          console.log(`    Response:`);
          const responseStr = JSON.stringify(step.result.response, null, 2);
          const lines = responseStr.split('\n');
          for (const line of lines.slice(0, 10)) {
            console.log(`      ${line}`);
          }
          if (lines.length > 10) {
            console.log(`      ... (${lines.length - 10} more lines)`);
          }
        }

        // Show running_script status
        if (step.result.running_script !== undefined) {
          console.log(`    Running script: ${step.result.running_script}`);
        }

        // Show limit
        if (step.result.limit !== undefined) {
          console.log(`    Limit: ${step.result.limit}`);
        }
      }

      // Show errors
      if (step.error) {
        console.log(`    Error:`);
        const errorStr =
          typeof step.error === 'string' ? step.error : JSON.stringify(step.error, null, 2);
        const lines = errorStr.split('\n');
        for (const line of lines) {
          console.log(`      ${line}`);
        }
      }

      if (step.result?.error) {
        console.log(`    Result Error:`);
        const errorStr =
          typeof step.result.error === 'string'
            ? step.result.error
            : JSON.stringify(step.result.error, null, 2);
        const lines = errorStr.split('\n');
        for (const line of lines) {
          console.log(`      ${line}`);
        }
      }

      console.log('');
      actionNum++;
    }
  }

  // Show summary
  const successCount = actionPaths
    .flatMap(([, steps]) => steps)
    .filter((s) => !s.error && !s.result?.error && s.result?.enabled !== false).length;
  const failCount = actionPaths
    .flatMap(([, steps]) => steps)
    .filter((s) => s.error || s.result?.error).length;
  const skipCount = actionPaths
    .flatMap(([, steps]) => steps)
    .filter((s) => s.result?.enabled === false).length;

  console.log('Summary:');
  console.log(`Total actions: ${actionNum - 1}`);
  console.log(`Successful: ${successCount}`);
  if (failCount > 0) console.log(`Failed: ${failCount}`);
  if (skipCount > 0) console.log(`Skipped: ${skipCount}`);
}

/**
 * Display a comprehensive debug view of an automation trace.
 * Combines trigger, variables, and action information in one output.
 *
 * @param ctx - Command context with WebSocket and arguments
 *
 * @example
 * ```bash
 * npx tsx ha-ws-client.ts trace-debug 01KDQS4E2WHMYJYYXKC7K28XFG
 * ```
 */
export async function handleTraceDebug(ctx: CommandContext): Promise<void> {
  const runId = requireArg(
    ctx,
    1,
    'Usage: trace-debug <run_id> [automation_id]\n' +
      '  run_id: The run ID from traces command\n' +
      '  automation_id: Optional automation ID (will auto-detect if not provided)\n' +
      '\nThis command provides a comprehensive debug view combining:\n' +
      '  - Trigger context\n' +
      '  - Variable values at each step\n' +
      '  - Action results\n' +
      '  - Any errors encountered'
  );
  const providedItemId = ctx.args[2];

  const { trace: result, itemId } = await getTraceDetail(ctx, runId, providedItemId);

  console.log(`Trace: automation.${itemId}`);
  console.log(`Run ID: ${runId}`);
  console.log(`Status: ${result.script_execution ?? 'unknown'}`);

  if (result.timestamp) {
    const start = new Date(result.timestamp.start);
    console.log(`Started: ${start.toLocaleString()}`);
    if (result.timestamp.finish) {
      const finish = new Date(result.timestamp.finish);
      const durationMs = finish.getTime() - start.getTime();
      console.log(`Duration: ${durationMs}ms`);
    }
  }

  if (result.error) {
    console.log(`\n!!! ERROR: ${result.error}`);
  }

  // Section 1: Trigger Context
  console.log(`\nTrigger:`);

  let triggerInfo: TraceTrigger | null = null;
  if (result.trace) {
    for (const steps of Object.values(result.trace)) {
      for (const step of steps) {
        // Check changed_variables first (where trigger info usually appears)
        if (step.changed_variables?.trigger) {
          triggerInfo = step.changed_variables.trigger as TraceTrigger;
          break;
        }
        // Also check variables as fallback
        if (step.variables?.trigger) {
          triggerInfo = step.variables.trigger as TraceTrigger;
          break;
        }
      }
      if (triggerInfo) break;
    }
  }

  if (result.trigger && typeof result.trigger === 'object') {
    triggerInfo = { ...triggerInfo, ...result.trigger };
  }

  if (triggerInfo) {
    if (triggerInfo.platform) console.log(`Platform: ${triggerInfo.platform}`);
    if (triggerInfo.entity_id) console.log(`Entity: ${triggerInfo.entity_id}`);
    if (triggerInfo.from_state) {
      console.log(`From: ${triggerInfo.from_state.state}`);
    }
    if (triggerInfo.to_state) {
      console.log(`To: ${triggerInfo.to_state.state}`);
    }
    if (triggerInfo.description) console.log(`Description: ${triggerInfo.description}`);
  } else {
    console.log('(No trigger information captured)');
  }

  // Section 2: Variables
  console.log(`\nVariables:`);

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

  if (importantVars.length > 0) {
    for (const [key, value] of importantVars) {
      const valueStr = typeof value === 'object' ? JSON.stringify(value) : String(value);
      const truncated = valueStr.length > 60 ? `${valueStr.substring(0, 60)}...` : valueStr;
      const typeStr = typeof value;
      console.log(`[${typeStr}] ${key}: ${truncated}`);
    }
  } else {
    console.log('(No variables captured - variables are captured at condition/choose steps)');
  }

  // Section 3: Execution Timeline
  console.log(`\nTimeline:`);

  if (result.trace && Object.keys(result.trace).length > 0) {
    // Collect and sort all steps
    const allSteps: Array<{ path: string; step: TraceStep }> = [];
    for (const [tracePath, steps] of Object.entries(result.trace)) {
      for (const step of steps) {
        allSteps.push({ path: tracePath, step });
      }
    }

    allSteps.sort((a, b) => {
      const timeA = a.step.timestamp ? new Date(a.step.timestamp).getTime() : 0;
      const timeB = b.step.timestamp ? new Date(b.step.timestamp).getTime() : 0;
      return timeA - timeB;
    });

    for (const { path, step } of allSteps) {
      const pathDesc = formatTracePath(path);
      const timeStr = step.timestamp
        ? new Date(step.timestamp).toLocaleTimeString('en-US', { hour12: false })
        : '??:??:??';

      let statusIcon = '[ok]';
      if (step.error || step.result?.error) {
        statusIcon = '[ERR]';
      } else if (step.result?.enabled === false) {
        statusIcon = '[SKIP]';
      }

      console.log(`\n${statusIcon} [${timeStr}] ${pathDesc}`);

      // Show result params
      if (step.result?.params) {
        const params = Object.entries(step.result.params).slice(0, 3);
        for (const [k, v] of params) {
          const vStr = typeof v === 'object' ? JSON.stringify(v) : String(v);
          const truncated = vStr.length > 50 ? `${vStr.substring(0, 50)}...` : vStr;
          console.log(`    ${k}: ${truncated}`);
        }
      }

      // Show errors
      if (step.error) {
        console.log(
          `    !!! Error: ${typeof step.error === 'string' ? step.error : JSON.stringify(step.error)}`
        );
      }
      if (step.result?.error) {
        console.log(
          `    !!! Result error: ${typeof step.result.error === 'string' ? step.result.error : JSON.stringify(step.result.error)}`
        );
      }
    }
  } else {
    console.log('(No trace steps recorded)');
  }

  // Section 4: Context
  console.log(`\nContext:`);

  if (result.context) {
    if (result.context.id) console.log(`Context ID: ${result.context.id}`);
    if (result.context.parent_id) console.log(`Parent ID: ${result.context.parent_id}`);
    if (result.context.user_id) console.log(`User ID: ${result.context.user_id}`);
  } else {
    console.log('(No context information)');
  }
}
