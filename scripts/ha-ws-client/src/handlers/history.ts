/**
 * History and logging command handlers.
 * @module handlers/history
 */

import { sendMessage } from '../client.js';
import { getOutputConfig, isJsonOutput, output, outputList, outputMessage } from '../output.js';
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
    ? `${startTime.toISOString()} to ${endTime.toISOString()}`
    : `last ${hours}h`;

  if (isJsonOutput()) {
    output(
      {
        entity_id: entityId,
        time_range: { start: startTime.toISOString(), end: endTime.toISOString() },
        entries: result,
      },
      { command: 'logbook', count: result.length }
    );
    return;
  }

  outputList(result, {
    title: `Logbook entries for ${entityId} (${timeDesc})`,
    command: 'logbook',
    itemFormatter: (entry) => {
      const when = new Date(entry.when * 1000).toLocaleString();
      return `  ${when}: ${entry.state ?? entry.message ?? 'event'}`;
    },
  });
  if (result.length === 0) {
    outputMessage('  No entries found');
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
    ? `${startTime.toISOString()} to ${endTime.toISOString()}`
    : `last ${hours}h`;

  if (isJsonOutput()) {
    output(
      {
        entity_id: entityId,
        time_range: { start: startTime.toISOString(), end: endTime.toISOString() },
        states: states.map((s) => ({
          timestamp: s.lu ? new Date(s.lu * 1000).toISOString() : null,
          state: s.s,
        })),
      },
      { command: 'history', count: states.length }
    );
    return;
  }

  outputList(states, {
    title: `State history for ${entityId} (${timeDesc})`,
    command: 'history',
    itemFormatter: (s) => {
      const when = new Date((s.lu ?? 0) * 1000).toLocaleString();
      return `  ${when}: ${s.s}`;
    },
  });
  if (states.length === 0) {
    outputMessage('  No state changes found');
  } else {
    outputMessage(`\nTotal: ${states.length} state changes`);
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
    ? `${startTime.toISOString()} to ${endTime.toISOString()}`
    : `last ${hours}h`;

  // Helper to get timestamp from state entry
  const getTimestamp = (s: HistoryState): string => {
    if (s.lu) return new Date(s.lu * 1000).toISOString();
    if (s.last_updated) return s.last_updated;
    if (s.lc) return new Date(s.lc * 1000).toISOString();
    if (s.last_changed) return s.last_changed;
    return '';
  };

  if (isJsonOutput()) {
    output(
      {
        entity_id: entityId,
        time_range: { start: startTime.toISOString(), end: endTime.toISOString() },
        states: states.map((s) => ({
          timestamp: getTimestamp(s),
          state: s.state ?? s.s,
          attributes: s.attributes ?? s.a ?? {},
        })),
      },
      { command: 'history-full', count: states.length }
    );
    return;
  }

  outputList(states, {
    title: `Full state history for ${entityId} (${timeDesc})`,
    command: 'history-full',
    itemFormatter: (s) => {
      const when = getTimestamp(s) ? new Date(getTimestamp(s)).toLocaleString() : 'unknown time';
      const state = s.state ?? s.s;
      const attrs = s.attributes ?? s.a ?? {};
      const attrStr = formatEntityAttributes(entityId, attrs);
      return `  ${when}: ${state}${attrStr}`;
    },
  });
  if (states.length === 0) {
    outputMessage('  No state changes found');
  } else {
    outputMessage(`\nTotal: ${states.length} state changes`);
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
    ? `${startTime.toISOString()} to ${endTime.toISOString()}`
    : `last ${hours}h`;

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

  // Helper to get timestamp
  const getTimestamp = (s: HistoryState): string => {
    if (s.lu) return new Date(s.lu * 1000).toISOString();
    if (s.last_updated) return s.last_updated;
    return '';
  };

  // Build attribute change records
  interface AttrChange {
    timestamp: string;
    state: string;
    changes: Record<string, unknown>;
  }

  const attrChanges: AttrChange[] = [];
  let prevAttrs: Record<string, unknown> = {};

  for (const s of states) {
    const state = s.state ?? s.s ?? '';
    const attrs = s.attributes ?? s.a ?? {};
    const changes: Record<string, unknown> = {};

    for (const attr of trackAttrs) {
      const prev = prevAttrs[attr];
      const curr = attrs[attr];
      if (curr !== undefined && JSON.stringify(prev) !== JSON.stringify(curr)) {
        changes[attr] = curr;
      }
    }

    attrChanges.push({
      timestamp: getTimestamp(s),
      state,
      changes,
    });

    prevAttrs = { ...attrs };
  }

  if (isJsonOutput()) {
    output(
      {
        entity_id: entityId,
        time_range: { start: startTime.toISOString(), end: endTime.toISOString() },
        tracked_attributes: trackAttrs,
        changes: attrChanges,
      },
      { command: 'attrs', count: states.length }
    );
    return;
  }

  // Default/compact output
  const { showHeaders } = getOutputConfig();
  if (showHeaders) {
    outputMessage(`Attribute history for ${entityId} (${timeDesc}):\n`);
  }

  if (states.length === 0) {
    outputMessage('  No state changes found');
    return;
  }

  for (const change of attrChanges) {
    const when = change.timestamp ? new Date(change.timestamp).toLocaleString() : 'unknown';
    const changeList = Object.entries(change.changes)
      .map(([attr, curr]) => {
        if (attr === 'brightness' && typeof curr === 'number') {
          return `brightness=${Math.round(curr / 2.55)}%`;
        } else if (attr === 'color_temp_kelvin' && typeof curr === 'number') {
          return `color_temp=${curr}K`;
        } else if (attr === 'color_temp' && typeof curr === 'number') {
          return `color_temp=${curr}mired`;
        } else if (typeof curr === 'object') {
          return `${attr}=${JSON.stringify(curr)}`;
        }
        return `${attr}=${curr}`;
      })
      .join(', ');

    let line = `  ${when}: ${change.state}`;
    if (changeList) {
      line += ` (${changeList})`;
    }
    console.log(line);
  }

  outputMessage(`\nTotal: ${states.length} state changes`);
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
    timestamp: string;
    entity_id: string;
    state: string;
  }

  const allEntries: TimelineEntry[] = [];
  for (const entityId of entityIds) {
    const states = result[entityId] ?? [];
    for (const s of states) {
      const timestamp = s.lu
        ? new Date(s.lu * 1000).toISOString()
        : s.last_updated
          ? s.last_updated
          : '';
      if (timestamp) {
        allEntries.push({
          timestamp,
          entity_id: entityId,
          state: s.s ?? s.state ?? 'unknown',
        });
      }
    }
  }

  // Sort by timestamp
  allEntries.sort((a, b) => new Date(a.timestamp).getTime() - new Date(b.timestamp).getTime());

  const timeDesc = ctx.fromTime
    ? `${startTime.toISOString()} to ${endTime.toISOString()}`
    : `last ${hours}h`;

  if (isJsonOutput()) {
    output(
      {
        entities: entityIds,
        time_range: { start: startTime.toISOString(), end: endTime.toISOString() },
        events: allEntries,
      },
      { command: 'timeline', count: allEntries.length }
    );
    return;
  }

  const { showHeaders } = getOutputConfig();
  if (showHeaders) {
    outputMessage(`Timeline for ${entityIds.length} entities (${timeDesc}):\n`);
  }

  if (allEntries.length === 0) {
    outputMessage('  No events found');
  } else {
    // Calculate max entity_id length for alignment
    const maxLen = Math.max(...entityIds.map((id) => id.length));

    for (const entry of allEntries) {
      const when = new Date(entry.timestamp).toLocaleString();
      const label = entry.entity_id.padEnd(maxLen);
      console.log(`  ${when}  ${label}  ${entry.state}`);
    }

    outputMessage(
      `\nTotal: ${allEntries.length} state changes across ${entityIds.length} entities`
    );
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
  const { format, maxItems } = getOutputConfig();
  const limit = maxItems > 0 ? maxItems : 20;

  if (format === 'json') {
    output(result.slice(0, limit), { command: 'syslog', count: result.length });
    return;
  }

  outputMessage(`System log (${result.length} entries):\n`);
  for (const entry of result.slice(0, limit)) {
    const level = entry.level ?? 'INFO';
    const source = entry.source?.[0] ?? 'unknown';
    const msg = entry.message?.slice(0, 100) ?? '';
    if (format === 'compact') {
      console.log(
        `[${level.toUpperCase()}] ${source}: ${msg}${entry.message && entry.message.length > 100 ? '...' : ''}`
      );
    } else {
      console.log(`[${level.toUpperCase()}] ${source}`);
      console.log(`  ${msg}${entry.message && entry.message.length > 100 ? '...' : ''}`);
      console.log('');
    }
  }
  if (result.length > limit) {
    outputMessage(`... and ${result.length - limit} more entries`);
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

  if (isJsonOutput()) {
    output(
      {
        entity_id: entityId,
        period: hours > DEFAULT_HOURS ? 'day' : 'hour',
        statistics: stats.map((s) => ({
          start: s.start,
          min: s.min,
          max: s.max,
          mean: s.mean,
        })),
      },
      { command: 'stats', count: stats.length }
    );
    return;
  }

  outputMessage(`Statistics for ${entityId} (last ${hours}h):`);
  if (stats.length === 0) {
    outputMessage('  No statistics found (entity may not record statistics)');
  } else {
    for (const s of stats) {
      const when = new Date(s.start).toLocaleString();
      console.log(`  ${when}: min=${s.min}, mean=${s.mean?.toFixed(2)}, max=${s.max}`);
    }
  }
}
