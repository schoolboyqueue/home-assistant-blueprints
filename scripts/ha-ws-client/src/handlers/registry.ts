/**
 * Registry command handlers for entities, devices, and areas.
 * @module handlers/registry
 */

import { sendMessage } from '../client.js';
import { getOutputConfig, output, outputList, outputMessage } from '../output.js';
import type { AreaEntry, CommandContext, DeviceEntry, EntityEntry } from '../types.js';

/**
 * List or search the entity registry.
 * Shows entity metadata including platform and disabled status.
 *
 * @param ctx - Command context with WebSocket and arguments
 *
 * @example
 * ```bash
 * npx tsx ha-ws-client.ts entities
 * npx tsx ha-ws-client.ts entities "kitchen"
 * npx tsx ha-ws-client.ts entities "light.*"
 * ```
 */
export async function handleEntities(ctx: CommandContext): Promise<void> {
  const pattern = ctx.args[1];
  const result = await sendMessage<EntityEntry[]>(ctx.ws, 'config/entity_registry/list');

  let filtered = result;
  if (pattern) {
    const regex = new RegExp(pattern, 'i');
    filtered = result.filter(
      (e) =>
        regex.test(e.entity_id) || regex.test(e.name ?? '') || regex.test(e.original_name ?? '')
    );
  }

  const { format, maxItems } = getOutputConfig();
  const limit = maxItems > 0 ? maxItems : 30;

  if (format === 'json') {
    output(filtered.slice(0, limit), { command: 'entities', count: filtered.length });
    return;
  }

  outputList(filtered.slice(0, limit), {
    title: `Entity registry${pattern ? ` matching "${pattern}"` : ''}: ${filtered.length} entities`,
    command: 'entities',
    itemFormatter: (e) => {
      const name = e.name ?? e.original_name ?? '';
      const platform = e.platform ?? '';
      const disabled = e.disabled_by ? ' [DISABLED]' : '';
      if (format === 'compact') {
        return `${e.entity_id}${disabled}${name ? ` (${name})` : ''}`;
      }
      let output = `  ${e.entity_id}${disabled}`;
      if (name) output += `\n    name: ${name}`;
      if (platform) output += `\n    platform: ${platform}`;
      return output;
    },
  });
  if (filtered.length > limit) {
    outputMessage(`\n  ... and ${filtered.length - limit} more`);
  }
}

/**
 * List or search the device registry.
 * Shows device metadata including manufacturer, model, and area.
 *
 * @param ctx - Command context with WebSocket and arguments
 *
 * @example
 * ```bash
 * npx tsx ha-ws-client.ts devices
 * npx tsx ha-ws-client.ts devices "philips"
 * npx tsx ha-ws-client.ts devices "living_room"
 * ```
 */
export async function handleDevices(ctx: CommandContext): Promise<void> {
  const pattern = ctx.args[1];
  const result = await sendMessage<DeviceEntry[]>(ctx.ws, 'config/device_registry/list');

  let filtered = result;
  if (pattern) {
    const regex = new RegExp(pattern, 'i');
    filtered = result.filter(
      (d) =>
        regex.test(d.name ?? '') ||
        regex.test(d.name_by_user ?? '') ||
        regex.test(d.manufacturer ?? '') ||
        regex.test(d.model ?? '') ||
        regex.test(d.area_id ?? '')
    );
  }

  const { format, maxItems } = getOutputConfig();
  const limit = maxItems > 0 ? maxItems : 20;

  if (format === 'json') {
    output(filtered.slice(0, limit), { command: 'devices', count: filtered.length });
    return;
  }

  outputList(filtered.slice(0, limit), {
    title: `Device registry${pattern ? ` matching "${pattern}"` : ''}: ${filtered.length} devices`,
    command: 'devices',
    itemFormatter: (d) => {
      const name = d.name ?? d.name_by_user ?? 'Unnamed';
      if (format === 'compact') {
        const parts = [name, d.manufacturer, d.model, d.area_id].filter(Boolean);
        return parts.join(' | ');
      }
      let output = `  ${name}\n    id: ${d.id}`;
      if (d.manufacturer) output += `\n    manufacturer: ${d.manufacturer}`;
      if (d.model) output += `\n    model: ${d.model}`;
      if (d.area_id) output += `\n    area: ${d.area_id}`;
      return output;
    },
  });
  if (filtered.length > limit) {
    outputMessage(`\n  ... and ${filtered.length - limit} more`);
  }
}

/**
 * List all areas in the area registry.
 * Shows area names, IDs, and any aliases.
 *
 * @param ctx - Command context with WebSocket and arguments
 *
 * @example
 * ```bash
 * npx tsx ha-ws-client.ts areas
 * ```
 */
export async function handleAreas(ctx: CommandContext): Promise<void> {
  const result = await sendMessage<AreaEntry[]>(ctx.ws, 'config/area_registry/list');
  const { format } = getOutputConfig();

  if (format === 'json') {
    output(result, { command: 'areas', count: result.length });
    return;
  }

  outputList(result, {
    title: 'Areas',
    command: 'areas',
    itemFormatter: (a) => {
      if (format === 'compact') {
        const aliases = a.aliases && a.aliases.length > 0 ? ` [${a.aliases.join(', ')}]` : '';
        return `${a.area_id}: ${a.name}${aliases}`;
      }
      let output = `  ${a.name} (${a.area_id})`;
      if (a.aliases && a.aliases.length > 0) {
        output += `\n    aliases: ${a.aliases.join(', ')}`;
      }
      return output;
    },
  });
}
