/**
 * Entity monitoring command handlers for live state tracking.
 * Provides real-time monitoring with historical tracking, rate-of-change detection,
 * and anomaly highlighting for troubleshooting automation behavior.
 * @module handlers/monitor
 */

import type WebSocket from 'ws';
import { nextId, pendingRequests, sendMessage } from '../client.js';
import type { CommandContext, HAMessage, HAState, HistoryState } from '../types.js';

// =============================================================================
// Types and Interfaces
// =============================================================================

/**
 * A recorded state change entry for historical tracking.
 */
interface StateChangeRecord {
  readonly timestamp: Date;
  readonly state: string;
  readonly previousState: string | null;
  readonly attributes: Record<string, unknown>;
  readonly previousAttributes: Record<string, unknown> | null;
  readonly rateOfChange: number | null;
  readonly isAnomaly: boolean;
  readonly anomalyReason: string | null;
}

/**
 * Configuration for anomaly detection.
 */
interface AnomalyConfig {
  /** Standard deviation threshold for numeric anomalies */
  readonly stdDevThreshold: number;
  /** Minimum rate of change per second to flag as rapid */
  readonly rapidChangeThreshold: number;
  /** Minimum time between changes (seconds) to flag as oscillating */
  readonly oscillationWindow: number;
  /** Number of rapid toggles in window to flag as oscillating */
  readonly oscillationCount: number;
}

/**
 * Statistics for monitored entity values.
 */
interface EntityStats {
  readonly count: number;
  readonly min: number;
  readonly max: number;
  readonly mean: number;
  readonly stdDev: number;
  readonly lastValues: readonly number[];
}

// =============================================================================
// Constants
// =============================================================================

/** Default monitoring duration in seconds */
const DEFAULT_MONITOR_SECONDS = 300;

/** Maximum number of historical records to keep */
const MAX_HISTORY_RECORDS = 1000;

/** Default anomaly detection configuration */
const DEFAULT_ANOMALY_CONFIG: AnomalyConfig = {
  stdDevThreshold: 3,
  rapidChangeThreshold: 10,
  oscillationWindow: 60,
  oscillationCount: 5,
};

/** ANSI color codes for terminal output */
const COLORS = {
  reset: '\x1b[0m',
  red: '\x1b[31m',
  green: '\x1b[32m',
  yellow: '\x1b[33m',
  blue: '\x1b[34m',
  magenta: '\x1b[35m',
  cyan: '\x1b[36m',
  dim: '\x1b[2m',
  bold: '\x1b[1m',
} as const;

// =============================================================================
// Utility Functions
// =============================================================================

/**
 * Parse a value as a number.
 */
function parseNumeric(value: unknown): number | null {
  if (typeof value === 'number') return value;
  if (typeof value === 'string') {
    const num = parseFloat(value);
    if (!Number.isNaN(num) && Number.isFinite(num)) return num;
  }
  return null;
}

/**
 * Calculate statistics for an array of numbers.
 */
function calculateStats(values: readonly number[]): EntityStats {
  if (values.length === 0) {
    return { count: 0, min: 0, max: 0, mean: 0, stdDev: 0, lastValues: [] };
  }

  const count = values.length;
  const min = Math.min(...values);
  const max = Math.max(...values);
  const mean = values.reduce((sum, v) => sum + v, 0) / count;

  const squaredDiffs = values.map((v) => (v - mean) ** 2);
  const avgSquaredDiff = squaredDiffs.reduce((sum, v) => sum + v, 0) / count;
  const stdDev = Math.sqrt(avgSquaredDiff);

  return {
    count,
    min,
    max,
    mean,
    stdDev,
    lastValues: values.slice(-50),
  };
}

/**
 * Format a duration in milliseconds to a human-readable string.
 */
function formatDuration(ms: number): string {
  if (ms < 1000) return `${ms}ms`;
  if (ms < 60000) return `${(ms / 1000).toFixed(1)}s`;
  if (ms < 3600000) return `${(ms / 60000).toFixed(1)}m`;
  return `${(ms / 3600000).toFixed(1)}h`;
}

/**
 * Format a rate of change value.
 */
function formatRateOfChange(rate: number | null, unit: string = ''): string {
  if (rate === null) return 'N/A';
  const absRate = Math.abs(rate);
  const sign = rate >= 0 ? '+' : '-';
  if (absRate < 0.01) return `${sign}0${unit}/s`;
  if (absRate < 1) return `${sign}${absRate.toFixed(3)}${unit}/s`;
  if (absRate < 100) return `${sign}${absRate.toFixed(1)}${unit}/s`;
  return `${sign}${absRate.toFixed(0)}${unit}/s`;
}

/**
 * Detect relevant attribute changes between two attribute objects.
 */
function detectAttributeChanges(
  prev: Record<string, unknown>,
  curr: Record<string, unknown>,
  entityId: string
): string[] {
  const changes: string[] = [];
  const trackAttrs = getTrackedAttributes(entityId);

  for (const attr of trackAttrs) {
    const prevVal = prev[attr];
    const currVal = curr[attr];

    if (currVal !== undefined && JSON.stringify(prevVal) !== JSON.stringify(currVal)) {
      const prevStr = formatAttributeValue(attr, prevVal);
      const currStr = formatAttributeValue(attr, currVal);
      changes.push(`${attr}: ${prevStr} → ${currStr}`);
    }
  }

  return changes;
}

/**
 * Get tracked attributes based on entity type.
 */
function getTrackedAttributes(entityId: string): string[] {
  if (entityId.startsWith('light.')) {
    return ['brightness', 'color_temp', 'color_temp_kelvin', 'rgb_color'];
  }
  if (entityId.startsWith('climate.')) {
    return [
      'hvac_action',
      'temperature',
      'current_temperature',
      'humidity',
      'fan_mode',
      'preset_mode',
    ];
  }
  if (entityId.startsWith('cover.')) {
    return ['current_position', 'current_tilt_position'];
  }
  if (entityId.startsWith('fan.')) {
    return ['percentage', 'preset_mode', 'direction'];
  }
  if (entityId.startsWith('media_player.')) {
    return ['volume_level', 'source', 'media_title'];
  }
  if (entityId.startsWith('sensor.') || entityId.startsWith('binary_sensor.')) {
    return [];
  }
  return [];
}

/**
 * Format an attribute value for display.
 */
function formatAttributeValue(attr: string, value: unknown): string {
  if (value === undefined || value === null) return 'null';
  if (attr === 'brightness' && typeof value === 'number') {
    return `${Math.round(value / 2.55)}%`;
  }
  if (attr === 'color_temp_kelvin' && typeof value === 'number') {
    return `${value}K`;
  }
  if (attr === 'volume_level' && typeof value === 'number') {
    return `${Math.round(value * 100)}%`;
  }
  if (typeof value === 'object') {
    return JSON.stringify(value);
  }
  return String(value);
}

// =============================================================================
// Anomaly Detection
// =============================================================================

/**
 * Check for anomalies in a state change.
 */
function detectAnomaly(
  record: Omit<StateChangeRecord, 'isAnomaly' | 'anomalyReason'>,
  history: readonly StateChangeRecord[],
  stats: EntityStats,
  config: AnomalyConfig
): { isAnomaly: boolean; reason: string | null } {
  const reasons: string[] = [];

  // Check for numeric anomalies using standard deviation
  const numericValue = parseNumeric(record.state);
  if (numericValue !== null && stats.count >= 10 && stats.stdDev > 0) {
    const zScore = Math.abs(numericValue - stats.mean) / stats.stdDev;
    if (zScore > config.stdDevThreshold) {
      reasons.push(`Value ${numericValue.toFixed(2)} is ${zScore.toFixed(1)} std devs from mean`);
    }
  }

  // Check for rapid rate of change
  if (record.rateOfChange !== null && Math.abs(record.rateOfChange) > config.rapidChangeThreshold) {
    reasons.push(`Rapid change: ${formatRateOfChange(record.rateOfChange)}`);
  }

  // Check for oscillation (rapid toggling between states)
  const now = record.timestamp.getTime();
  const windowStart = now - config.oscillationWindow * 1000;
  const recentChanges = history.filter((r) => r.timestamp.getTime() >= windowStart);

  if (recentChanges.length >= config.oscillationCount) {
    // Check if it's toggling between two states
    const states = new Set(recentChanges.map((r) => r.state));
    if (states.size === 2) {
      reasons.push(
        `Oscillating between states (${recentChanges.length} changes in ${config.oscillationWindow}s)`
      );
    }
  }

  // Check for unusual binary sensor patterns
  if (
    record.state === 'unavailable' ||
    record.state === 'unknown' ||
    record.previousState === 'unavailable' ||
    record.previousState === 'unknown'
  ) {
    if (record.state === 'unavailable') {
      reasons.push('Entity became unavailable');
    } else if (record.previousState === 'unavailable') {
      reasons.push('Entity recovered from unavailable');
    }
  }

  return {
    isAnomaly: reasons.length > 0,
    reason: reasons.length > 0 ? reasons.join('; ') : null,
  };
}

// =============================================================================
// Output Formatting
// =============================================================================

/**
 * Format a state change record for terminal output.
 */
function formatStateChange(
  record: StateChangeRecord,
  entityId: string,
  showDetails: boolean
): string {
  const lines: string[] = [];
  const timeStr = record.timestamp.toLocaleTimeString('en-US', { hour12: false });

  // Main state change line
  let mainLine = `[${timeStr}] `;

  if (record.isAnomaly) {
    mainLine += `${COLORS.red}${COLORS.bold}[!]${COLORS.reset} `;
  }

  mainLine += `${record.previousState ?? '(initial)'} → ${COLORS.cyan}${record.state}${COLORS.reset}`;

  if (record.rateOfChange !== null) {
    const rateColor =
      Math.abs(record.rateOfChange) > DEFAULT_ANOMALY_CONFIG.rapidChangeThreshold
        ? COLORS.yellow
        : COLORS.dim;
    mainLine += ` ${rateColor}(${formatRateOfChange(record.rateOfChange)})${COLORS.reset}`;
  }

  lines.push(mainLine);

  // Anomaly reason
  if (record.isAnomaly && record.anomalyReason) {
    lines.push(`         ${COLORS.red}⚠ ${record.anomalyReason}${COLORS.reset}`);
  }

  // Attribute changes
  if (showDetails && record.previousAttributes) {
    const attrChanges = detectAttributeChanges(
      record.previousAttributes,
      record.attributes,
      entityId
    );
    if (attrChanges.length > 0) {
      lines.push(`         ${COLORS.dim}Attributes: ${attrChanges.join(', ')}${COLORS.reset}`);
    }
  }

  return lines.join('\n');
}

/**
 * Print monitoring statistics summary.
 */
function printStats(
  entityId: string,
  history: readonly StateChangeRecord[],
  stats: EntityStats,
  duration: number
): void {
  console.log(`\n${COLORS.bold}═══ Monitoring Summary ═══${COLORS.reset}`);
  console.log(`Entity: ${entityId}`);
  console.log(`Duration: ${formatDuration(duration)}`);
  console.log(`State changes: ${history.length}`);

  if (stats.count > 0) {
    console.log(`\n${COLORS.bold}Numeric Statistics:${COLORS.reset}`);
    console.log(`  Min: ${stats.min.toFixed(2)}`);
    console.log(`  Max: ${stats.max.toFixed(2)}`);
    console.log(`  Mean: ${stats.mean.toFixed(2)}`);
    console.log(`  Std Dev: ${stats.stdDev.toFixed(2)}`);
  }

  const anomalies = history.filter((r) => r.isAnomaly);
  if (anomalies.length > 0) {
    console.log(
      `\n${COLORS.red}${COLORS.bold}Anomalies Detected: ${anomalies.length}${COLORS.reset}`
    );
    for (const a of anomalies.slice(-5)) {
      console.log(`  ${a.timestamp.toLocaleTimeString()}: ${a.state} - ${a.anomalyReason}`);
    }
    if (anomalies.length > 5) {
      console.log(`  ... and ${anomalies.length - 5} more`);
    }
  }

  // State distribution
  const stateCounts = new Map<string, number>();
  for (const r of history) {
    stateCounts.set(r.state, (stateCounts.get(r.state) ?? 0) + 1);
  }

  if (stateCounts.size > 1) {
    console.log(`\n${COLORS.bold}State Distribution:${COLORS.reset}`);
    const sorted = [...stateCounts.entries()].sort((a, b) => b[1] - a[1]);
    for (const [state, count] of sorted.slice(0, 10)) {
      const pct = ((count / history.length) * 100).toFixed(1);
      console.log(`  ${state}: ${count} (${pct}%)`);
    }
  }
}

// =============================================================================
// Command Handlers
// =============================================================================

/**
 * Monitor a single entity's state changes in real-time.
 * Provides live monitoring with historical tracking, rate-of-change detection,
 * and anomaly highlighting.
 *
 * @param ctx - Command context with WebSocket and arguments
 *
 * @example
 * ```bash
 * npx tsx ha-ws-client.ts monitor sensor.temperature 300
 * npx tsx ha-ws-client.ts monitor light.kitchen 60
 * npx tsx ha-ws-client.ts monitor binary_sensor.motion 120
 * ```
 */
export async function handleMonitor(ctx: CommandContext): Promise<void> {
  const entityId = ctx.args[1];
  const seconds = parseInt(ctx.args[2] as string, 10) || DEFAULT_MONITOR_SECONDS;
  const showDetails = !ctx.args.includes('--no-details');

  if (!entityId) {
    console.error('Usage: monitor <entity_id> [seconds] [--no-details]');
    console.error('  entity_id: Entity to monitor (e.g., sensor.temperature)');
    console.error('  seconds: Monitoring duration (default: 300)');
    console.error('  --no-details: Hide attribute changes');
    console.error('\nFeatures:');
    console.error('  - Live state change tracking');
    console.error('  - Historical tracking of all changes');
    console.error('  - Rate-of-change detection for numeric sensors');
    console.error('  - Anomaly detection and highlighting');
    console.error('  - Summary statistics at end');
    process.exit(1);
  }

  console.log(`${COLORS.bold}Monitoring ${entityId} for ${seconds} seconds...${COLORS.reset}`);
  console.log('Press Ctrl+C to stop early.\n');

  // Subscribe to state changes
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

  // Get initial state
  const states = await sendMessage<HAState[]>(ctx.ws, 'get_states');
  const initialState = states.find((s) => s.entity_id === entityId);

  if (!initialState) {
    console.error(`Entity not found: ${entityId}`);
    process.exit(1);
  }

  // Initialize tracking variables
  const history: StateChangeRecord[] = [];
  const numericValues: number[] = [];
  let lastState = initialState.state;
  let lastAttributes = initialState.attributes ?? {};
  let lastTimestamp = new Date();
  const startTime = Date.now();

  // Display initial state
  console.log(`${COLORS.bold}Initial state:${COLORS.reset} ${initialState.state}`);
  const trackedAttrs = getTrackedAttributes(entityId);
  if (trackedAttrs.length > 0) {
    const attrValues = trackedAttrs
      .filter((a) => lastAttributes[a] !== undefined)
      .map((a) => `${a}=${formatAttributeValue(a, lastAttributes[a])}`);
    if (attrValues.length > 0) {
      console.log(`${COLORS.dim}  Attributes: ${attrValues.join(', ')}${COLORS.reset}`);
    }
  }

  // Track initial numeric value
  const initialNumeric = parseNumeric(initialState.state);
  if (initialNumeric !== null) {
    numericValues.push(initialNumeric);
  }

  console.log(`\n${COLORS.bold}─── Live State Changes ───${COLORS.reset}\n`);

  // Event handler for state changes
  const eventHandler = (data: WebSocket.Data): void => {
    let msg: HAMessage;
    try {
      msg = JSON.parse(data.toString()) as HAMessage;
    } catch {
      return;
    }

    if (msg.type === 'event' && msg.event?.variables) {
      interface StateTrigger {
        trigger?: {
          from_state?: { state: string; attributes?: Record<string, unknown> };
          to_state?: { state: string; attributes?: Record<string, unknown> };
        };
      }

      const vars = msg.event.variables as StateTrigger;
      const toState = vars.trigger?.to_state;

      if (!toState) return;

      const now = new Date();
      const newState = toState.state;
      const newAttributes = toState.attributes ?? {};

      // Calculate rate of change for numeric values
      let rateOfChange: number | null = null;
      const numericValue = parseNumeric(newState);
      const prevNumeric = parseNumeric(lastState);

      if (numericValue !== null) {
        numericValues.push(numericValue);
        if (numericValues.length > MAX_HISTORY_RECORDS) {
          numericValues.shift();
        }

        if (prevNumeric !== null) {
          const timeDiff = (now.getTime() - lastTimestamp.getTime()) / 1000;
          if (timeDiff > 0) {
            rateOfChange = (numericValue - prevNumeric) / timeDiff;
          }
        }
      }

      // Calculate current statistics
      const stats = calculateStats(numericValues);

      // Create the record without anomaly info first
      const partialRecord: Omit<StateChangeRecord, 'isAnomaly' | 'anomalyReason'> = {
        timestamp: now,
        state: newState,
        previousState: lastState,
        attributes: newAttributes,
        previousAttributes: lastAttributes,
        rateOfChange,
      };

      // Detect anomalies
      const { isAnomaly, reason } = detectAnomaly(
        partialRecord,
        history,
        stats,
        DEFAULT_ANOMALY_CONFIG
      );

      // Create the full record
      const record: StateChangeRecord = {
        ...partialRecord,
        isAnomaly,
        anomalyReason: reason,
      };

      // Add to history
      history.push(record);
      if (history.length > MAX_HISTORY_RECORDS) {
        history.shift();
      }

      // Display the state change
      console.log(formatStateChange(record, entityId, showDetails));

      // Update tracking variables
      lastState = newState;
      lastAttributes = newAttributes;
      lastTimestamp = now;
    }
  };

  ctx.ws.on('message', eventHandler);

  // Wait for the monitoring duration
  await new Promise((resolve) => setTimeout(resolve, seconds * 1000));

  // Clean up
  ctx.ws.removeListener('message', eventHandler);

  // Print summary statistics
  const duration = Date.now() - startTime;
  const finalStats = calculateStats(numericValues);
  printStats(entityId, history, finalStats, duration);

  console.log(`\n${COLORS.green}Monitoring complete.${COLORS.reset}`);
}

/**
 * Monitor multiple entities simultaneously.
 * Shows a unified timeline of state changes across all monitored entities.
 *
 * @param ctx - Command context with WebSocket and arguments
 *
 * @example
 * ```bash
 * npx tsx ha-ws-client.ts monitor-multi 60 sensor.temp1 sensor.temp2 light.kitchen
 * ```
 */
export async function handleMonitorMulti(ctx: CommandContext): Promise<void> {
  const seconds = parseInt(ctx.args[1] as string, 10) || DEFAULT_MONITOR_SECONDS;
  const entityIds = ctx.args.slice(2).filter((a) => !a.startsWith('--')) as string[];

  if (entityIds.length === 0) {
    console.error('Usage: monitor-multi <seconds> <entity1> <entity2> ... [--no-details]');
    console.error('  seconds: Monitoring duration');
    console.error('  entity1, entity2, ...: Entities to monitor');
    console.error('\nExample:');
    console.error('  monitor-multi 60 sensor.temp1 sensor.temp2 binary_sensor.motion');
    process.exit(1);
  }

  const showDetails = !ctx.args.includes('--no-details');

  console.log(
    `${COLORS.bold}Monitoring ${entityIds.length} entities for ${seconds} seconds...${COLORS.reset}`
  );
  console.log('Entities:');
  for (const id of entityIds) {
    console.log(`  - ${id}`);
  }
  console.log('Press Ctrl+C to stop early.\n');

  // Create labels for compact display
  const labels = new Map<string, string>();
  const maxLabelLength = 15;
  for (const id of entityIds) {
    const parts = id.split('.');
    const domain = (parts[0] ?? '').substring(0, 3);
    const name = (parts[1] ?? '').substring(0, maxLabelLength - 4);
    labels.set(id, `${domain}.${name}`);
  }

  // Subscribe to all entities
  const subscriptions: number[] = [];
  for (const entityId of entityIds) {
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
    subscriptions.push(subId);
  }

  // Get initial states
  const states = await sendMessage<HAState[]>(ctx.ws, 'get_states');
  const entityStates = new Map<string, HAState>();
  for (const entityId of entityIds) {
    const state = states.find((s) => s.entity_id === entityId);
    if (state) {
      entityStates.set(entityId, state);
    }
  }

  // Display initial states
  console.log(`${COLORS.bold}Initial states:${COLORS.reset}`);
  for (const [entityId, state] of entityStates) {
    console.log(`  ${labels.get(entityId)}: ${state.state}`);
  }

  console.log(`\n${COLORS.bold}─── Live State Changes ───${COLORS.reset}\n`);

  // Track changes
  let totalEvents = 0;
  const startTime = Date.now();

  // Event handler
  const eventHandler = (data: WebSocket.Data): void => {
    let msg: HAMessage;
    try {
      msg = JSON.parse(data.toString()) as HAMessage;
    } catch {
      return;
    }

    if (msg.type === 'event' && msg.event?.variables) {
      interface StateTrigger {
        trigger?: {
          entity_id?: string;
          from_state?: { state: string; attributes?: Record<string, unknown> };
          to_state?: { state: string; attributes?: Record<string, unknown> };
        };
      }

      const vars = msg.event.variables as StateTrigger;
      const triggerEntityId = vars.trigger?.entity_id;

      if (!triggerEntityId || !entityIds.includes(triggerEntityId)) return;

      totalEvents++;
      const now = new Date();
      const timeStr = now.toLocaleTimeString('en-US', { hour12: false });
      const label = (labels.get(triggerEntityId) ?? triggerEntityId).padEnd(maxLabelLength);
      const fromState = vars.trigger?.from_state?.state ?? '?';
      const toState = vars.trigger?.to_state?.state ?? '?';

      let line = `[${timeStr}] ${COLORS.cyan}${label}${COLORS.reset} `;
      line += `${fromState} → ${COLORS.bold}${toState}${COLORS.reset}`;

      console.log(line);

      // Track attribute changes if detailed
      if (
        showDetails &&
        vars.trigger?.from_state?.attributes &&
        vars.trigger?.to_state?.attributes
      ) {
        const attrChanges = detectAttributeChanges(
          vars.trigger.from_state.attributes,
          vars.trigger.to_state.attributes,
          triggerEntityId
        );
        if (attrChanges.length > 0) {
          console.log(
            `${' '.repeat(timeStr.length + 3)}${COLORS.dim}${attrChanges.join(', ')}${COLORS.reset}`
          );
        }
      }
    }
  };

  ctx.ws.on('message', eventHandler);

  // Wait for monitoring duration
  await new Promise((resolve) => setTimeout(resolve, seconds * 1000));

  // Clean up
  ctx.ws.removeListener('message', eventHandler);

  // Print summary
  const duration = Date.now() - startTime;
  console.log(`\n${COLORS.bold}═══ Monitoring Summary ═══${COLORS.reset}`);
  console.log(`Duration: ${formatDuration(duration)}`);
  console.log(`Total state changes: ${totalEvents}`);
  console.log(`Entities monitored: ${entityIds.length}`);

  console.log(`\n${COLORS.green}Monitoring complete.${COLORS.reset}`);
}

/**
 * Analyze historical state changes for anomalies.
 * Fetches history and applies anomaly detection retroactively.
 *
 * @param ctx - Command context with WebSocket and arguments
 *
 * @example
 * ```bash
 * npx tsx ha-ws-client.ts analyze sensor.temperature 24
 * npx tsx ha-ws-client.ts analyze binary_sensor.motion 4
 * ```
 */
export async function handleAnalyze(ctx: CommandContext): Promise<void> {
  const entityId = ctx.args[1];
  const hours = parseFloat(ctx.args[2] as string) || 24;

  if (!entityId) {
    console.error('Usage: analyze <entity_id> [hours]');
    console.error('  entity_id: Entity to analyze');
    console.error('  hours: Historical period to analyze (default: 24)');
    console.error('\nAnalyzes historical state changes for:');
    console.error('  - Rate of change patterns');
    console.error('  - Anomalous values');
    console.error('  - Oscillation/flapping behavior');
    console.error('  - Unavailability periods');
    process.exit(1);
  }

  console.log(
    `${COLORS.bold}Analyzing ${entityId} for the last ${hours} hours...${COLORS.reset}\n`
  );

  const endTime = new Date();
  const startTime = new Date(endTime.getTime() - hours * 3600000);

  // Fetch history
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

  const rawStates = result[entityId] ?? [];

  if (rawStates.length === 0) {
    console.log('No historical data found for this entity.');
    return;
  }

  console.log(`Found ${rawStates.length} state changes to analyze.\n`);

  // Convert to our record format
  const history: StateChangeRecord[] = [];
  const numericValues: number[] = [];
  let prevState: string | null = null;
  let prevTimestamp: Date | null = null;
  let prevAttributes: Record<string, unknown> | null = null;

  for (const s of rawStates) {
    const timestamp =
      s.lu !== undefined
        ? new Date(s.lu * 1000)
        : s.last_updated
          ? new Date(s.last_updated)
          : new Date();

    const state = s.state ?? s.s ?? '';
    const attributes = s.attributes ?? s.a ?? {};

    // Calculate rate of change
    let rateOfChange: number | null = null;
    const numericValue = parseNumeric(state);
    const prevNumeric = prevState !== null ? parseNumeric(prevState) : null;

    if (numericValue !== null) {
      numericValues.push(numericValue);

      if (prevNumeric !== null && prevTimestamp !== null) {
        const timeDiff = (timestamp.getTime() - prevTimestamp.getTime()) / 1000;
        if (timeDiff > 0) {
          rateOfChange = (numericValue - prevNumeric) / timeDiff;
        }
      }
    }

    const stats = calculateStats(numericValues);

    const partialRecord: Omit<StateChangeRecord, 'isAnomaly' | 'anomalyReason'> = {
      timestamp,
      state,
      previousState: prevState,
      attributes,
      previousAttributes: prevAttributes,
      rateOfChange,
    };

    const { isAnomaly, reason } = detectAnomaly(
      partialRecord,
      history,
      stats,
      DEFAULT_ANOMALY_CONFIG
    );

    history.push({
      ...partialRecord,
      isAnomaly,
      anomalyReason: reason,
    });

    prevState = state;
    prevTimestamp = timestamp;
    prevAttributes = attributes;
  }

  // Calculate final statistics
  const stats = calculateStats(numericValues);

  // Report findings
  console.log(`${COLORS.bold}═══ Analysis Results ═══${COLORS.reset}`);
  console.log(`\nPeriod: ${startTime.toLocaleString()} to ${endTime.toLocaleString()}`);
  console.log(`Total state changes: ${history.length}`);

  // Numeric statistics
  if (stats.count > 0) {
    console.log(`\n${COLORS.bold}Numeric Statistics:${COLORS.reset}`);
    console.log(`  Samples: ${stats.count}`);
    console.log(`  Min: ${stats.min.toFixed(2)}`);
    console.log(`  Max: ${stats.max.toFixed(2)}`);
    console.log(`  Mean: ${stats.mean.toFixed(2)}`);
    console.log(`  Std Dev: ${stats.stdDev.toFixed(2)}`);
    console.log(`  Range: ${(stats.max - stats.min).toFixed(2)}`);
  }

  // Anomalies
  const anomalies = history.filter((r) => r.isAnomaly);
  if (anomalies.length > 0) {
    console.log(
      `\n${COLORS.red}${COLORS.bold}Anomalies Detected: ${anomalies.length}${COLORS.reset}`
    );
    console.log('');

    // Group by type
    const byReason = new Map<string, StateChangeRecord[]>();
    for (const a of anomalies) {
      const key = a.anomalyReason?.split(';')[0]?.trim() ?? 'Unknown';
      const list = byReason.get(key) ?? [];
      list.push(a);
      byReason.set(key, list);
    }

    for (const [reason, records] of byReason) {
      console.log(`${COLORS.yellow}${reason}:${COLORS.reset} ${records.length} occurrences`);
      for (const r of records.slice(0, 3)) {
        console.log(`  - ${r.timestamp.toLocaleString()}: ${r.previousState ?? '?'} → ${r.state}`);
      }
      if (records.length > 3) {
        console.log(`  ... and ${records.length - 3} more`);
      }
    }
  } else {
    console.log(`\n${COLORS.green}No anomalies detected.${COLORS.reset}`);
  }

  // State distribution
  const stateCounts = new Map<string, number>();
  for (const r of history) {
    stateCounts.set(r.state, (stateCounts.get(r.state) ?? 0) + 1);
  }

  if (stateCounts.size > 1) {
    console.log(`\n${COLORS.bold}State Distribution:${COLORS.reset}`);
    const sorted = [...stateCounts.entries()].sort((a, b) => b[1] - a[1]);
    for (const [state, count] of sorted.slice(0, 10)) {
      const pct = ((count / history.length) * 100).toFixed(1);
      const bar = '█'.repeat(Math.ceil(parseFloat(pct) / 5));
      console.log(
        `  ${state.padEnd(20)} ${count.toString().padStart(5)} (${pct.padStart(5)}%) ${COLORS.dim}${bar}${COLORS.reset}`
      );
    }
  }

  // Unavailability analysis
  const unavailable = history.filter((r) => r.state === 'unavailable' || r.state === 'unknown');
  if (unavailable.length > 0) {
    console.log(
      `\n${COLORS.yellow}${COLORS.bold}Unavailability Events: ${unavailable.length}${COLORS.reset}`
    );
    for (const u of unavailable.slice(0, 5)) {
      console.log(`  - ${u.timestamp.toLocaleString()}: became ${u.state}`);
    }
    if (unavailable.length > 5) {
      console.log(`  ... and ${unavailable.length - 5} more`);
    }
  }

  console.log(`\n${COLORS.green}Analysis complete.${COLORS.reset}`);
}
