/**
 * History and logging command handlers.
 * @module handlers/history
 */

import { sendMessage } from '../client.js';
import type {
  CommandContext,
  HistoryState,
  LogbookEntry,
  StatEntry,
  SysLogEntry,
} from '../types.js';
import { calculateTimeRange, formatEntityAttributes, requireArg } from '../utils.js';

/** Default number of hours for history queries. */
const DEFAULT_HOURS = 24;

/**
 * Get logbook entries for an entity.
 * Shows timestamped events from the Home Assistant logbook.
 *
 * @param ctx - Command context with WebSocket and arguments
 *
 * @example
 * ```bash
 * npx tsx ha-ws-client.ts logbook light.kitchen 4
 * npx tsx ha-ws-client.ts logbook light.kitchen --from "2024-01-01 00:00"
 * ```
 */
export async function handleLogbook(ctx: CommandContext): Promise<void> {
  const entityId = requireArg(
    ctx,
    1,
    'Usage: logbook <entity_id> [hours] [--from "TIME"] [--to "TIME"]'
  );
  const hours = parseFloat(ctx.args[2] as string) || DEFAULT_HOURS;

  const { startTime, endTime } = calculateTimeRange(ctx.fromTime, ctx.toTime, hours);

  const result = await sendMessage<LogbookEntry[]>(ctx.ws, 'logbook/get_events', {
    start_time: startTime.toISOString(),
    end_time: endTime.toISOString(),
    entity_ids: [entityId],
  });

  const timeDesc = ctx.fromTime
    ? `${startTime.toLocaleString()} to ${endTime.toLocaleString()}`
    : `last ${hours}h`;
  console.log(`Logbook entries for ${entityId} (${timeDesc}):`);
  if (result.length === 0) {
    console.log('  No entries found');
  } else {
    for (const entry of result) {
      const when = new Date(entry.when * 1000).toLocaleString();
      console.log(`  ${when}: ${entry.state ?? entry.message ?? 'event'}`);
    }
  }
}

/**
 * Get state history for an entity.
 * Shows state changes over time in a compact format.
 *
 * @param ctx - Command context with WebSocket and arguments
 *
 * @example
 * ```bash
 * npx tsx ha-ws-client.ts history sensor.temperature 24
 * npx tsx ha-ws-client.ts history light.kitchen --from "2024-01-01 00:00" --to "2024-01-02 00:00"
 * ```
 */
export async function handleHistory(ctx: CommandContext): Promise<void> {
  const entityId = requireArg(
    ctx,
    1,
    'Usage: history <entity_id> [hours] [--from "TIME"] [--to "TIME"]'
  );
  const hours = parseFloat(ctx.args[2] as string) || DEFAULT_HOURS;

  const { startTime, endTime } = calculateTimeRange(ctx.fromTime, ctx.toTime, hours);

  const result = await sendMessage<Record<string, HistoryState[]>>(
    ctx.ws,
    'history/history_during_period',
    {
      start_time: startTime.toISOString(),
      end_time: endTime.toISOString(),
      entity_ids: [entityId],
      minimal_response: true,
      significant_changes_only: true,
    }
  );

  const states = result[entityId] ?? [];
  const timeDesc = ctx.fromTime
    ? `${startTime.toLocaleString()} to ${endTime.toLocaleString()}`
    : `last ${hours}h`;
  console.log(`State history for ${entityId} (${timeDesc}):`);
  if (states.length === 0) {
    console.log('  No state changes found');
  } else {
    for (const s of states) {
      const when = new Date((s.lu ?? 0) * 1000).toLocaleString();
      console.log(`  ${when}: ${s.s}`);
    }
    console.log(`\nTotal: ${states.length} state changes`);
  }
}

/**
 * Get full state history with attributes.
 * Shows state changes with relevant attribute values.
 *
 * @param ctx - Command context with WebSocket and arguments
 *
 * @example
 * ```bash
 * npx tsx ha-ws-client.ts history-full climate.thermostat 12
 * ```
 */
export async function handleHistoryFull(ctx: CommandContext): Promise<void> {
  const entityId = requireArg(
    ctx,
    1,
    'Usage: history-full <entity_id> [hours] [--from "TIME"] [--to "TIME"]'
  );
  const hours = parseFloat(ctx.args[2] as string) || DEFAULT_HOURS;

  const { startTime, endTime } = calculateTimeRange(ctx.fromTime, ctx.toTime, hours);

  const result = await sendMessage<Record<string, HistoryState[]>>(
    ctx.ws,
    'history/history_during_period',
    {
      start_time: startTime.toISOString(),
      end_time: endTime.toISOString(),
      entity_ids: [entityId],
      minimal_response: false,
      significant_changes_only: true,
      no_attributes: false,
    }
  );

  const states = result[entityId] ?? [];
  const timeDesc = ctx.fromTime
    ? `${startTime.toLocaleString()} to ${endTime.toLocaleString()}`
    : `last ${hours}h`;
  console.log(`Full state history for ${entityId} (${timeDesc}):`);
  if (states.length === 0) {
    console.log('  No state changes found');
  } else {
    for (const s of states) {
      let when: string;
      if (s.lu) {
        when = new Date(s.lu * 1000).toLocaleString();
      } else if (s.last_updated) {
        when = new Date(s.last_updated).toLocaleString();
      } else if (s.lc) {
        when = new Date(s.lc * 1000).toLocaleString();
      } else if (s.last_changed) {
        when = new Date(s.last_changed).toLocaleString();
      } else {
        when = 'unknown time';
      }
      const state = s.state ?? s.s;
      const attrs = s.attributes ?? s.a ?? {};
      const attrStr = formatEntityAttributes(entityId, attrs);
      console.log(`  ${when}: ${state}${attrStr}`);
    }
    console.log(`\nTotal: ${states.length} state changes`);
  }
}

/**
 * Get attribute change history for an entity.
 * Shows changes in specific attributes relevant to the entity type.
 *
 * @param ctx - Command context with WebSocket and arguments
 *
 * @example
 * ```bash
 * npx tsx ha-ws-client.ts attrs light.kitchen 4
 * npx tsx ha-ws-client.ts attrs climate.thermostat 24
 * ```
 */
export async function handleAttrs(ctx: CommandContext): Promise<void> {
  const entityId = requireArg(ctx, 1, 'Usage: attrs <entity_id> [hours]');
  const hours = parseFloat(ctx.args[2] as string) || DEFAULT_HOURS;

  const { startTime, endTime } = calculateTimeRange(ctx.fromTime, ctx.toTime, hours);

  const result = await sendMessage<Record<string, HistoryState[]>>(
    ctx.ws,
    'history/history_during_period',
    {
      start_time: startTime.toISOString(),
      end_time: endTime.toISOString(),
      entity_ids: [entityId],
      minimal_response: false,
      significant_changes_only: false,
      no_attributes: false,
    }
  );

  const states = result[entityId] ?? [];
  const timeDesc = ctx.fromTime
    ? `${startTime.toLocaleString()} to ${endTime.toLocaleString()}`
    : `last ${hours}h`;

  console.log(`Attribute history for ${entityId} (${timeDesc}):\n`);

  if (states.length === 0) {
    console.log('  No state changes found');
    return;
  }

  // Determine which attributes to track based on entity type
  let trackAttrs: string[] = [];
  if (entityId.startsWith('light.')) {
    trackAttrs = ['brightness', 'color_temp', 'color_temp_kelvin', 'rgb_color', 'hs_color'];
  } else if (entityId.startsWith('climate.')) {
    trackAttrs = [
      'hvac_action',
      'temperature',
      'target_temp_high',
      'target_temp_low',
      'current_temperature',
      'humidity',
      'fan_mode',
      'preset_mode',
    ];
  } else if (entityId.startsWith('cover.')) {
    trackAttrs = ['current_position', 'current_tilt_position'];
  } else if (entityId.startsWith('fan.')) {
    trackAttrs = ['percentage', 'preset_mode', 'direction'];
  } else if (entityId.startsWith('media_player.')) {
    trackAttrs = ['volume_level', 'source', 'media_title', 'media_artist'];
  } else {
    // For unknown entity types, detect attributes dynamically
    const allAttrs = new Set<string>();
    for (const s of states) {
      const attrs = s.attributes ?? s.a ?? {};
      for (const k of Object.keys(attrs)) {
        allAttrs.add(k);
      }
    }
    const noise = new Set([
      'friendly_name',
      'icon',
      'supported_features',
      'entity_id',
      'supported_color_modes',
      'min_mireds',
      'max_mireds',
      'min_color_temp_kelvin',
      'max_color_temp_kelvin',
      'device_class',
      'state_class',
      'unit_of_measurement',
    ]);
    trackAttrs = [...allAttrs].filter((a) => !noise.has(a));
  }

  let prevAttrs: Record<string, unknown> = {};
  for (const s of states) {
    let when: string;
    if (s.lu) {
      when = new Date(s.lu * 1000).toLocaleString();
    } else if (s.last_updated) {
      when = new Date(s.last_updated).toLocaleString();
    } else {
      when = 'unknown';
    }

    const state = s.state ?? s.s;
    const attrs = s.attributes ?? s.a ?? {};

    const changes: string[] = [];
    for (const attr of trackAttrs) {
      const prev = prevAttrs[attr];
      const curr = attrs[attr];
      if (curr !== undefined && JSON.stringify(prev) !== JSON.stringify(curr)) {
        if (attr === 'brightness' && typeof curr === 'number') {
          changes.push(`brightness=${Math.round(curr / 2.55)}%`);
        } else if (attr === 'color_temp_kelvin' && typeof curr === 'number') {
          changes.push(`color_temp=${curr}K`);
        } else if (attr === 'color_temp' && typeof curr === 'number') {
          changes.push(`color_temp=${curr}mired`);
        } else if (typeof curr === 'object') {
          changes.push(`${attr}=${JSON.stringify(curr)}`);
        } else {
          changes.push(`${attr}=${curr}`);
        }
      }
    }

    let line = `  ${when}: ${state}`;
    if (changes.length > 0) {
      line += ` (${changes.join(', ')})`;
    }
    console.log(line);

    prevAttrs = { ...attrs };
  }

  console.log(`\nTotal: ${states.length} state changes`);
}

/**
 * Show a chronological timeline of multiple entities.
 * Displays state changes from multiple entities sorted by time.
 * Uses the history API to capture all state changes including sensors.
 *
 * @param ctx - Command context with WebSocket and arguments
 *
 * @example
 * ```bash
 * npx tsx ha-ws-client.ts timeline 2 climate.thermostat binary_sensor.door sensor.temperature
 * ```
 */
export async function handleTimeline(ctx: CommandContext): Promise<void> {
  const hoursArg = requireArg(
    ctx,
    1,
    'Usage: timeline <hours> <entity1> <entity2> ... [--from "TIME"] [--to "TIME"]\n' +
      'Example: timeline 2 climate.thermostat binary_sensor.door sensor.temperature'
  );
  const hours = parseFloat(hoursArg);
  const entityIds = ctx.args.slice(2) as string[];

  if (entityIds.length === 0) {
    console.error('Usage: timeline <hours> <entity1> <entity2> ... [--from "TIME"] [--to "TIME"]');
    console.error('Example: timeline 2 climate.thermostat binary_sensor.door sensor.temperature');
    process.exit(1);
  }

  const { startTime, endTime } = calculateTimeRange(ctx.fromTime, ctx.toTime, hours);

  // Use history API instead of logbook to capture all state changes including sensors
  const result = await sendMessage<Record<string, HistoryState[]>>(
    ctx.ws,
    'history/history_during_period',
    {
      start_time: startTime.toISOString(),
      end_time: endTime.toISOString(),
      entity_ids: entityIds,
      minimal_response: true,
      significant_changes_only: true,
    }
  );

  // Collect all entries from all entities
  interface TimelineEntry {
    when: number;
    entity_id: string;
    state: string;
  }

  const allEntries: TimelineEntry[] = [];
  for (const entityId of entityIds) {
    const states = result[entityId] ?? [];
    for (const s of states) {
      const timestamp = s.lu
        ? s.lu * 1000
        : s.last_updated
          ? new Date(s.last_updated).getTime()
          : 0;
      if (timestamp > 0) {
        allEntries.push({
          when: timestamp,
          entity_id: entityId,
          state: s.s ?? s.state ?? 'unknown',
        });
      }
    }
  }

  // Sort by timestamp
  allEntries.sort((a, b) => a.when - b.when);

  const timeDesc = ctx.fromTime
    ? `${startTime.toLocaleString()} to ${endTime.toLocaleString()}`
    : `last ${hours}h`;
  console.log(`Timeline for ${entityIds.length} entities (${timeDesc}):\n`);

  if (allEntries.length === 0) {
    console.log('  No events found');
  } else {
    // Calculate max entity_id length for alignment
    const maxLen = Math.max(...entityIds.map((id) => id.length));

    for (const entry of allEntries) {
      const when = new Date(entry.when).toLocaleString();
      const label = entry.entity_id.padEnd(maxLen);
      console.log(`  ${when}  ${label}  ${entry.state}`);
    }

    console.log(`\nTotal: ${allEntries.length} state changes across ${entityIds.length} entities`);
  }
}

/**
 * Get system log entries.
 * Shows recent errors and warnings from Home Assistant's system log.
 *
 * @param ctx - Command context with WebSocket and arguments
 *
 * @example
 * ```bash
 * npx tsx ha-ws-client.ts syslog
 * ```
 */
export async function handleSyslog(ctx: CommandContext): Promise<void> {
  const result = await sendMessage<SysLogEntry[]>(ctx.ws, 'system_log/list');
  console.log(`System log (${result.length} entries):\n`);
  for (const entry of result.slice(0, 20)) {
    const level = entry.level ?? 'INFO';
    const source = entry.source?.[0] ?? 'unknown';
    const msg = entry.message?.slice(0, 100) ?? '';
    console.log(`[${level.toUpperCase()}] ${source}`);
    console.log(`  ${msg}${entry.message && entry.message.length > 100 ? '...' : ''}`);
    console.log('');
  }
  if (result.length > 20) {
    console.log(`... and ${result.length - 20} more entries`);
  }
}

/**
 * Get sensor statistics.
 * Shows min/max/mean values for sensors that record statistics.
 *
 * @param ctx - Command context with WebSocket and arguments
 *
 * @example
 * ```bash
 * npx tsx ha-ws-client.ts stats sensor.temperature 24
 * ```
 */
export async function handleStats(ctx: CommandContext): Promise<void> {
  const entityId = requireArg(ctx, 1, 'Usage: stats <entity_id> [hours]');
  const hours = parseFloat(ctx.args[2] as string) || DEFAULT_HOURS;
  const { startTime } = calculateTimeRange(null, null, hours);

  const result = await sendMessage<Record<string, StatEntry[]>>(
    ctx.ws,
    'recorder/statistics_during_period',
    {
      start_time: startTime.toISOString(),
      statistic_ids: [entityId],
      period: hours > DEFAULT_HOURS ? 'day' : 'hour',
    }
  );

  const stats = result[entityId] ?? [];
  console.log(`Statistics for ${entityId} (last ${hours}h):`);
  if (stats.length === 0) {
    console.log('  No statistics found (entity may not record statistics)');
  } else {
    for (const s of stats) {
      const when = new Date(s.start).toLocaleString();
      console.log(`  ${when}: min=${s.min}, mean=${s.mean?.toFixed(2)}, max=${s.max}`);
    }
  }
}
