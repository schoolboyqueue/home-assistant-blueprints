/**
 * Basic command handlers for common Home Assistant operations.
 * @module handlers/basic
 */

import * as fs from 'node:fs';
import type WebSocket from 'ws';
import { HAClientError, nextId, sendMessage } from '../client.js';
import { EntityNotFoundError } from '../errors.js';
import { getOutputConfig, isJsonOutput, output, outputList, outputMessage } from '../output.js';
import type { CommandContext, HAConfig, HAMessage, HAState } from '../types.js';
import { parseJsonArg, requireArg } from '../utils.js';

/**
 * Test the WebSocket connection with a ping/pong.
 * Useful for verifying connectivity and measuring latency.
 *
 * @param ctx - Command context with WebSocket and arguments
 *
 * @example
 * ```bash
 * npx tsx ha-ws-client.ts ping
 * # Output: Pong! (42ms)
 * ```
 */
export async function handlePing(ctx: CommandContext): Promise<void> {
  const start = Date.now();
  await sendMessage(ctx.ws, 'ping');
  const latency = Date.now() - start;

  if (isJsonOutput()) {
    output({ latency_ms: latency }, { command: 'ping' });
  } else {
    outputMessage(`Pong! (${latency}ms)`);
  }
}

/**
 * Get the state of a single entity.
 * Returns the full state object including attributes.
 *
 * @param ctx - Command context with WebSocket and arguments
 * @throws {EntityNotFoundError} If the entity does not exist
 *
 * @example
 * ```bash
 * npx tsx ha-ws-client.ts state sun.sun
 * ```
 */
export async function handleState(ctx: CommandContext): Promise<void> {
  const entityId = requireArg(ctx, 1, 'Usage: state <entity_id>');
  const states = await sendMessage<HAState[]>(ctx.ws, 'get_states');
  const entity = states.find((s) => s.entity_id === entityId);
  if (!entity) {
    throw new EntityNotFoundError(entityId);
  }
  output(entity, { command: 'state' });
}

/**
 * Get a summary of all entity states.
 * Shows total count and a sample of 10 entities.
 *
 * @param ctx - Command context with WebSocket and arguments
 *
 * @example
 * ```bash
 * npx tsx ha-ws-client.ts states
 * # Output: Total entities: 150
 * #         Sample entities:
 * #           sun.sun: above_horizon
 * #           ...
 * ```
 */
export async function handleStates(ctx: CommandContext): Promise<void> {
  const states = await sendMessage<HAState[]>(ctx.ws, 'get_states');
  const { format, maxItems } = getOutputConfig();
  const limit = maxItems > 0 ? maxItems : 10;

  if (format === 'json') {
    output(
      { total: states.length, sample: states.slice(0, limit) },
      { command: 'states', count: states.length }
    );
  } else {
    outputList(states.slice(0, limit), {
      title: `Total entities: ${states.length}\nSample`,
      command: 'states',
      itemFormatter: (s) => `  ${s.entity_id}: ${s.state}`,
    });
  }
}

/**
 * Get all entity states as JSON.
 * Outputs the complete state array for programmatic use.
 *
 * @param ctx - Command context with WebSocket and arguments
 *
 * @example
 * ```bash
 * npx tsx ha-ws-client.ts states-json > all-states.json
 * ```
 */
export async function handleStatesJson(ctx: CommandContext): Promise<void> {
  const states = await sendMessage<HAState[]>(ctx.ws, 'get_states');
  output(states, { command: 'states-json', count: states.length });
}

/**
 * Filter states by entity_id pattern.
 * Supports glob-like patterns with * wildcards.
 *
 * @param ctx - Command context with WebSocket and arguments
 *
 * @example
 * ```bash
 * npx tsx ha-ws-client.ts states-filter "light.*"
 * npx tsx ha-ws-client.ts states-filter "sensor.*temperature*"
 * ```
 */
export async function handleStatesFilter(ctx: CommandContext): Promise<void> {
  const pattern = requireArg(ctx, 1, 'Usage: states-filter <pattern>');
  const states = await sendMessage<HAState[]>(ctx.ws, 'get_states');
  const regex = new RegExp(pattern.replace(/\*/g, '.*'));
  const filtered = states.filter((s) => regex.test(s.entity_id));

  outputList(filtered, {
    title: `Found ${filtered.length} matching entities`,
    command: 'states-filter',
    itemFormatter: (s) => `${s.entity_id}: ${s.state}`,
  });
}

/**
 * Get Home Assistant configuration.
 * Returns version, location, timezone, and component count.
 *
 * @param ctx - Command context with WebSocket and arguments
 *
 * @example
 * ```bash
 * npx tsx ha-ws-client.ts config
 * ```
 */
export async function handleConfig(ctx: CommandContext): Promise<void> {
  const config = await sendMessage<HAConfig>(ctx.ws, 'get_config');
  const summary = {
    version: config.version,
    location_name: config.location_name,
    time_zone: config.time_zone,
    unit_system: config.unit_system,
    state: config.state,
    components_count: config.components.length,
  };
  output(summary, { command: 'config' });
}

/**
 * List all available services.
 * Groups services by domain with their available service calls.
 *
 * @param ctx - Command context with WebSocket and arguments
 *
 * @example
 * ```bash
 * npx tsx ha-ws-client.ts services
 * # Output: Domains: 45
 * #           light: turn_on, turn_off, toggle
 * #           ...
 * ```
 */
export async function handleServices(ctx: CommandContext): Promise<void> {
  const services = await sendMessage<Record<string, Record<string, unknown>>>(
    ctx.ws,
    'get_services'
  );
  const domains = Object.keys(services).sort();
  const { format } = getOutputConfig();

  if (format === 'json') {
    const data = domains.map((domain) => ({
      domain,
      services: Object.keys(services[domain] ?? {}),
    }));
    output(data, { command: 'services', count: domains.length });
  } else {
    outputList(domains, {
      title: `Domains`,
      command: 'services',
      itemFormatter: (domain) => {
        const domainServices = services[domain];
        const svcList = domainServices ? Object.keys(domainServices).join(', ') : '';
        return `  ${domain}: ${svcList}`;
      },
    });
  }
}

/**
 * Call a Home Assistant service.
 * Executes a service call with optional JSON data.
 *
 * @param ctx - Command context with WebSocket and arguments
 *
 * @example
 * ```bash
 * npx tsx ha-ws-client.ts call light turn_on '{"entity_id":"light.kitchen"}'
 * npx tsx ha-ws-client.ts call script turn_on '{"entity_id":"script.morning_routine"}'
 * ```
 */
export async function handleCall(ctx: CommandContext): Promise<void> {
  const domain = requireArg(ctx, 1, 'Usage: call <domain> <service> [data]');
  const service = requireArg(ctx, 2, 'Usage: call <domain> <service> [data]');
  const serviceData: Record<string, unknown> = ctx.args[3]
    ? parseJsonArg<Record<string, unknown>>(ctx.args[3], 'service data')
    : {};

  const result = await sendMessage<Record<string, unknown> | null>(ctx.ws, 'call_service', {
    domain,
    service,
    service_data: serviceData,
  });

  const { format } = getOutputConfig();
  if (format === 'json') {
    output(
      {
        domain,
        service,
        service_data: serviceData,
        response: result,
      },
      { command: 'call', summary: 'Service called successfully' }
    );
  } else {
    outputMessage('Service called successfully');
    if (result && Object.keys(result).length > 0) {
      output(result, { summary: 'Response:' });
    }
  }
}

/**
 * Render a Jinja2 template.
 * Evaluates a template string using Home Assistant's template engine.
 * Supports reading template from stdin using '-' argument.
 *
 * @param ctx - Command context with WebSocket and arguments
 *
 * @example
 * ```bash
 * npx tsx ha-ws-client.ts template "{{ states('sun.sun') }}"
 * npx tsx ha-ws-client.ts template "{{ now() }}"
 * echo "{{ states.light | list | count }}" | npx tsx ha-ws-client.ts template -
 * ```
 */
export async function handleTemplate(ctx: CommandContext): Promise<void> {
  let template = ctx.args[1];

  // If no argument provided, read from stdin
  if (!template) {
    try {
      template = fs.readFileSync(0, 'utf-8').trim();
    } catch {
      console.error('Usage: template <template>');
      console.error(
        '  Or pipe template via stdin: echo "{{ now() }}" | npx tsx ha-ws-client.ts template -'
      );
      process.exit(1);
    }
  }

  // Handle "-" as explicit stdin marker
  if (template === '-') {
    template = fs.readFileSync(0, 'utf-8').trim();
  }

  if (!template) {
    console.error('Usage: template <template>');
    process.exit(1);
  }

  // render_template is a subscription: first sends result (confirmation), then event (rendered value)
  const id = nextId();
  const resultPromise = new Promise<string>((resolve, reject) => {
    let subscribed = false;
    const handler = (data: WebSocket.Data): void => {
      let msg: HAMessage;
      try {
        msg = JSON.parse(data.toString()) as HAMessage;
      } catch {
        return;
      }
      if (msg.id === id) {
        if (msg.type === 'result') {
          if (!msg.success) {
            ctx.ws.removeListener('message', handler);
            reject(
              new HAClientError(msg.error?.message ?? 'Template render failed', 'TEMPLATE_ERROR')
            );
          } else {
            subscribed = true;
          }
        } else if (msg.type === 'event' && subscribed) {
          ctx.ws.removeListener('message', handler);
          resolve(String(msg.event?.result ?? ''));
        }
      }
    };
    ctx.ws.on('message', handler);
  });

  ctx.ws.send(
    JSON.stringify({
      id,
      type: 'render_template',
      template,
    })
  );

  const result = await resultPromise;

  if (isJsonOutput()) {
    output({ template, result }, { command: 'template' });
  } else {
    outputMessage(result);
  }
}
