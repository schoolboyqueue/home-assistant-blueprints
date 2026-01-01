/**
 * Utility functions for the Home Assistant WebSocket API client.
 * @module utils
 */

import { DateParseError, JsonParseError } from './errors.js';
import type { CommandContext, TimeArgs, TimeRange, YamlModule } from './types.js';

// Re-export TimeRange for convenience
export type { TimeRange };

// Lazy-loaded yaml module (only loaded when needed for blueprint-inputs command)
let yamlModule: YamlModule | null = null;

/**
 * Get the js-yaml module, loading it lazily on first use.
 * This avoids loading the YAML parser unless the blueprint-inputs command is used.
 * @returns The js-yaml module
 */
export function getYamlModule(): YamlModule {
  if (yamlModule === null) {
    // Dynamic require for yaml - only loaded when blueprint-inputs command is used
    // eslint-disable-next-line @typescript-eslint/no-require-imports
    yamlModule = require('js-yaml') as YamlModule;
  }
  return yamlModule;
}

/**
 * Parse --from and --to time options from command line arguments.
 * Extracts time filters and returns the remaining arguments.
 *
 * @param argList - Array of command line arguments
 * @returns Object containing fromTime, toTime, and filtered arguments
 *
 * @example
 * ```typescript
 * const { fromTime, toTime, filteredArgs } = parseTimeArgs([
 *   'history', 'sensor.temp', '--from', '2024-01-01 00:00', '--to', '2024-01-02 00:00'
 * ]);
 * // fromTime: Date(2024-01-01T00:00:00)
 * // toTime: Date(2024-01-02T00:00:00)
 * // filteredArgs: ['history', 'sensor.temp']
 * ```
 */
/**
 * Calculate a time range based on optional from/to times and a default hours window.
 * This is a common pattern used across history, automation, and monitoring handlers.
 *
 * @param fromTime - Optional start time (if provided, overrides default calculation)
 * @param toTime - Optional end time (if provided, uses this; otherwise uses current time)
 * @param defaultHours - Number of hours to look back from endTime when fromTime is not provided
 * @returns Object containing startTime and endTime as Date objects
 *
 * @example
 * ```typescript
 * // With no arguments, uses default hours from now
 * const { startTime, endTime } = calculateTimeRange(null, null, 24);
 * // endTime = now, startTime = 24 hours ago
 *
 * // With explicit times from CLI args
 * const { startTime, endTime } = calculateTimeRange(ctx.fromTime, ctx.toTime, 24);
 *
 * // With only toTime specified
 * const { startTime, endTime } = calculateTimeRange(null, new Date('2024-01-15'), 4);
 * // endTime = 2024-01-15, startTime = 4 hours before that
 * ```
 */
export function calculateTimeRange(
  fromTime: Date | null | undefined,
  toTime: Date | null | undefined,
  defaultHours: number
): TimeRange {
  const endTime = toTime ?? new Date();
  const startTime = fromTime ?? new Date(endTime.getTime() - defaultHours * 3_600_000);
  return { startTime, endTime };
}

export function parseTimeArgs(argList: string[]): TimeArgs {
  let fromTime: Date | null = null;
  let toTime: Date | null = null;
  const filteredArgs: string[] = [];

  for (let i = 0; i < argList.length; i++) {
    const arg = argList[i];
    const nextArg = argList[i + 1];
    if (arg === '--from' && nextArg) {
      fromTime = parseFlexibleDate(nextArg);
      i++; // skip next arg
    } else if (arg === '--to' && nextArg) {
      toTime = parseFlexibleDate(nextArg);
      i++; // skip next arg
    } else if (arg) {
      filteredArgs.push(arg);
    }
  }

  return { fromTime, toTime, filteredArgs };
}

/**
 * Parse flexible date formats into a Date object.
 * Supports multiple input formats for user convenience.
 *
 * @param str - Date string in ISO, "YYYY-MM-DD HH:MM", or "MM/DD HH:MM" format
 * @returns Parsed Date object
 * @throws {DateParseError} If the date string cannot be parsed
 *
 * @example
 * ```typescript
 * parseFlexibleDate('2024-01-15T14:30:00Z');     // ISO format
 * parseFlexibleDate('2024-01-15 14:30');         // YYYY-MM-DD HH:MM
 * parseFlexibleDate('01/15 14:30');              // MM/DD HH:MM (current year)
 * ```
 */
export function parseFlexibleDate(str: string): Date {
  // Try ISO format first
  let d = new Date(str);
  if (!Number.isNaN(d.getTime())) return d;

  // Try "YYYY-MM-DD HH:MM" format
  const match = str.match(/^(\d{4})-(\d{2})-(\d{2})\s+(\d{1,2}):(\d{2})(?::(\d{2}))?$/);
  if (match) {
    const year = match[1] ?? '0';
    const month = match[2] ?? '1';
    const day = match[3] ?? '1';
    const hour = match[4] ?? '0';
    const min = match[5] ?? '0';
    const sec = match[6] ?? '0';
    d = new Date(
      parseInt(year, 10),
      parseInt(month, 10) - 1,
      parseInt(day, 10),
      parseInt(hour, 10),
      parseInt(min, 10),
      parseInt(sec, 10)
    );
    if (!Number.isNaN(d.getTime())) return d;
  }

  // Try "MM/DD HH:MM" format (assumes current year)
  const match2 = str.match(/^(\d{1,2})\/(\d{1,2})\s+(\d{1,2}):(\d{2})$/);
  if (match2) {
    const month = match2[1] ?? '1';
    const day = match2[2] ?? '1';
    const hour = match2[3] ?? '0';
    const min = match2[4] ?? '0';
    d = new Date(
      new Date().getFullYear(),
      parseInt(month, 10) - 1,
      parseInt(day, 10),
      parseInt(hour, 10),
      parseInt(min, 10)
    );
    if (!Number.isNaN(d.getTime())) return d;
  }

  throw new DateParseError(str);
}

/**
 * Format entity attributes for display based on entity type.
 * Returns a formatted string showing relevant attributes for the entity domain.
 *
 * @param entityId - The entity_id to format attributes for
 * @param attrs - The attributes object from the entity state
 * @returns Formatted string like " (attr1=val1, attr2=val2)" or empty string
 *
 * @example
 * ```typescript
 * formatEntityAttributes('climate.thermostat', {
 *   hvac_action: 'heating',
 *   temperature: 72,
 *   current_temperature: 68
 * });
 * // Returns: " (action=heating, target=72, current=68)"
 * ```
 */
export function formatEntityAttributes(entityId: string, attrs: Record<string, unknown>): string {
  if (entityId.startsWith('climate.')) {
    const parts: string[] = [];
    if (attrs.hvac_action) parts.push(`action=${attrs.hvac_action}`);
    if (attrs.temperature) parts.push(`target=${attrs.temperature}`);
    if (attrs.target_temp_high) parts.push(`high=${attrs.target_temp_high}`);
    if (attrs.target_temp_low) parts.push(`low=${attrs.target_temp_low}`);
    if (attrs.current_temperature) parts.push(`current=${attrs.current_temperature}`);
    if (attrs.fan_mode) parts.push(`fan=${attrs.fan_mode}`);
    return parts.length > 0 ? ` (${parts.join(', ')})` : '';
  }

  if (entityId.startsWith('light.')) {
    const parts: string[] = [];
    if (attrs.brightness) {
      parts.push(`brightness=${Math.round((attrs.brightness as number) / 2.55)}%`);
    }
    if (attrs.color_temp) parts.push(`temp=${attrs.color_temp}`);
    return parts.length > 0 ? ` (${parts.join(', ')})` : '';
  }

  if (entityId.startsWith('cover.')) {
    const parts: string[] = [];
    if (attrs.current_position !== undefined) {
      parts.push(`position=${attrs.current_position}%`);
    }
    return parts.length > 0 ? ` (${parts.join(', ')})` : '';
  }

  if (entityId.startsWith('fan.')) {
    const parts: string[] = [];
    if (attrs.percentage) parts.push(`speed=${attrs.percentage}%`);
    if (attrs.preset_mode) parts.push(`preset=${attrs.preset_mode}`);
    if (attrs.direction) parts.push(`direction=${attrs.direction}`);
    return parts.length > 0 ? ` (${parts.join(', ')})` : '';
  }

  // For other entities, show a few key attributes
  const skipKeys = new Set(['friendly_name', 'unit_of_measurement']);
  const otherAttrs = Object.entries(attrs)
    .filter(([k]) => !skipKeys.has(k))
    .slice(0, 3)
    .map(([k, v]) => `${k}=${typeof v === 'object' ? JSON.stringify(v) : v}`);
  return otherAttrs.length > 0 ? ` (${otherAttrs.join(', ')})` : '';
}

/**
 * Validate that a required argument exists at the specified index.
 * If the argument is missing, prints the usage message and exits.
 *
 * @param ctx - Command context with arguments
 * @param index - The argument index to check (1-based, where 0 is the command name)
 * @param usage - The usage message to display if validation fails
 * @returns The argument value as a string (non-null guaranteed)
 *
 * @example
 * ```typescript
 * // Single required argument
 * const entityId = requireArg(ctx, 1, 'Usage: state <entity_id>');
 *
 * // Multiple lines of usage info
 * const runId = requireArg(ctx, 1,
 *   'Usage: trace <run_id> [automation_id]\n' +
 *   '  run_id: The run ID from traces command\n' +
 *   '  automation_id: Optional automation ID'
 * );
 * ```
 */
export function requireArg(ctx: CommandContext, index: number, usage: string): string {
  const arg = ctx.args[index];
  if (!arg) {
    console.error(usage);
    process.exit(1);
  }
  return arg;
}

/**
 * Parse a JSON string argument with structured error handling.
 * Provides helpful error messages with context when parsing fails.
 *
 * @param arg - The JSON string to parse
 * @param fieldName - A descriptive name for the field (used in error messages)
 * @returns The parsed JSON value
 * @throws {JsonParseError} If the JSON string cannot be parsed
 *
 * @example
 * ```typescript
 * // Parse service call data
 * const data = parseJsonArg('{"entity_id":"light.kitchen"}', 'service data');
 *
 * // Parse with error handling
 * try {
 *   const payload = parseJsonArg(userInput, 'webhook payload');
 * } catch (err) {
 *   if (err instanceof JsonParseError) {
 *     console.error(err.message);
 *     // Error message includes field name, position, and helpful tip
 *   }
 * }
 * ```
 */
export function parseJsonArg<T = unknown>(arg: string, fieldName: string): T {
  try {
    return JSON.parse(arg) as T;
  } catch (err) {
    throw new JsonParseError(fieldName, err as Error, arg);
  }
}
