/**
 * Basic command handlers for common Home Assistant operations.
 * @module handlers/basic
 */

import * as fs from 'node:fs';
import type WebSocket from 'ws';
import { HAClientError, nextId, sendMessage } from '../client.js';
import { EntityNotFoundError } from '../errors.js';
import type { CommandContext, HAConfig, HAMessage, HAState } from '../types.js';

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
  console.log(`Pong! (${Date.now() - start}ms)`);
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
  const entityId = ctx.args[1];
  if (!entityId) {
    console.error('Usage: state <entity_id>');
    process.exit(1);
  }
  const states = await sendMessage<HAState[]>(ctx.ws, 'get_states');
  const entity = states.find((s) => s.entity_id === entityId);
  if (!entity) {
    throw new EntityNotFoundError(entityId);
  }
  console.log(JSON.stringify(entity, null, 2));
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
  console.log(`Total entities: ${states.length}`);
  console.log('\nSample entities:');
  for (const s of states.slice(0, 10)) {
    console.log(`  ${s.entity_id}: ${s.state}`);
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
  console.log(JSON.stringify(states, null, 2));
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
  const pattern = ctx.args[1];
  if (!pattern) {
    console.error('Usage: states-filter <pattern>');
    process.exit(1);
  }
  const states = await sendMessage<HAState[]>(ctx.ws, 'get_states');
  const regex = new RegExp(pattern.replace(/\*/g, '.*'));
  const filtered = states.filter((s) => regex.test(s.entity_id));
  for (const s of filtered) {
    console.log(`${s.entity_id}: ${s.state}`);
  }
  console.log(`\nFound ${filtered.length} matching entities`);
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
  console.log(
    JSON.stringify(
      {
        version: config.version,
        location_name: config.location_name,
        time_zone: config.time_zone,
        unit_system: config.unit_system,
        state: config.state,
        components_count: config.components.length,
      },
      null,
      2
    )
  );
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
  console.log(`Domains: ${domains.length}`);
  for (const domain of domains) {
    const domainServices = services[domain];
    const svcList = domainServices ? Object.keys(domainServices).join(', ') : '';
    console.log(`  ${domain}: ${svcList}`);
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
  const domain = ctx.args[1];
  const service = ctx.args[2];
  const serviceData: Record<string, unknown> = ctx.args[3]
    ? (JSON.parse(ctx.args[3]) as Record<string, unknown>)
    : {};

  if (!domain || !service) {
    console.error('Usage: call <domain> <service> [data]');
    process.exit(1);
  }

  const result = await sendMessage<Record<string, unknown> | null>(ctx.ws, 'call_service', {
    domain,
    service,
    service_data: serviceData,
  });
  console.log('Service called successfully');
  if (result && Object.keys(result).length > 0) {
    console.log('Response:', JSON.stringify(result, null, 2));
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
  console.log(result);
}
