/**
 * Registry command handlers for entities, devices, and areas.
 * @module handlers/registry
 */

import { sendMessage } from '../client.js';
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

  console.log(
    `Entity registry${pattern ? ` matching "${pattern}"` : ''}: ${filtered.length} entities\n`
  );
  for (const e of filtered.slice(0, 30)) {
    const name = e.name ?? e.original_name ?? '';
    const platform = e.platform ?? '';
    const disabled = e.disabled_by ? ' [DISABLED]' : '';
    console.log(`  ${e.entity_id}${disabled}`);
    if (name) console.log(`    name: ${name}`);
    if (platform) console.log(`    platform: ${platform}`);
  }
  if (filtered.length > 30) {
    console.log(`\n  ... and ${filtered.length - 30} more`);
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

  console.log(
    `Device registry${pattern ? ` matching "${pattern}"` : ''}: ${filtered.length} devices\n`
  );
  for (const d of filtered.slice(0, 20)) {
    console.log(`  ${d.name ?? d.name_by_user ?? 'Unnamed'}`);
    console.log(`    id: ${d.id}`);
    if (d.manufacturer) console.log(`    manufacturer: ${d.manufacturer}`);
    if (d.model) console.log(`    model: ${d.model}`);
    if (d.area_id) console.log(`    area: ${d.area_id}`);
    console.log('');
  }
  if (filtered.length > 20) {
    console.log(`  ... and ${filtered.length - 20} more`);
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
  console.log(`Areas: ${result.length}\n`);
  for (const a of result) {
    console.log(`  ${a.name} (${a.area_id})`);
    if (a.aliases && a.aliases.length > 0) {
      console.log(`    aliases: ${a.aliases.join(', ')}`);
    }
  }
}
