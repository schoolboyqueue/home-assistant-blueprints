/**
 * Output configuration and logging module for AI agent context efficiency.
 *
 * This module provides a centralized output system optimized for use by AI agents
 * like Claude Code. It supports multiple output modes:
 * - `default`: Human-readable formatted output
 * - `compact`: Reduced verbosity, single-line entries where possible
 * - `json`: Machine-readable JSON output (most context-efficient)
 *
 * @module output
 */

// =============================================================================
// Types
// =============================================================================

/**
 * Output format modes for controlling verbosity and structure.
 * - `default`: Human-readable with formatting
 * - `compact`: Reduced verbosity, suitable for AI context
 * - `json`: Machine-readable JSON, most context-efficient
 */
export type OutputFormat = 'default' | 'compact' | 'json';

/**
 * Output configuration options.
 */
export interface OutputConfig {
  /** Output format mode */
  format: OutputFormat;
  /** Whether to show timestamps */
  showTimestamps: boolean;
  /** Whether to show headers/separators */
  showHeaders: boolean;
  /** Maximum items to display (0 = unlimited) */
  maxItems: number;
}

/**
 * Structured result data for JSON output.
 */
export interface OutputResult<T = unknown> {
  /** Whether the operation succeeded */
  success: boolean;
  /** The command that was executed */
  command?: string;
  /** Result data */
  data?: T;
  /** Error message if failed */
  error?: string;
  /** Count of items if applicable */
  count?: number;
  /** Summary message */
  summary?: string;
}

// =============================================================================
// Global Configuration
// =============================================================================

/** Global output configuration */
let outputConfig: OutputConfig = {
  format: 'default',
  showTimestamps: true,
  showHeaders: true,
  maxItems: 0,
};

/** Cached adapter instance (forward declaration, set by getOutputAdapter) */
let currentAdapter: OutputAdapter | null = null;

/**
 * Get the current output configuration.
 */
export function getOutputConfig(): Readonly<OutputConfig> {
  return outputConfig;
}

/**
 * Set the output configuration.
 */
export function setOutputConfig(config: Partial<OutputConfig>): void {
  outputConfig = { ...outputConfig, ...config };
  currentAdapter = null; // Reset adapter when config changes
}

/**
 * Parse output format from command line arguments.
 * Extracts --output=<format> or --format=<format> flags.
 *
 * @param args - Command line arguments
 * @returns The remaining arguments after extracting output flags
 */
export function parseOutputArgs(args: string[]): string[] {
  const filtered: string[] = [];

  for (const arg of args) {
    if (arg.startsWith('--output=') || arg.startsWith('--format=')) {
      const format = arg.split('=')[1] as OutputFormat;
      if (format === 'json' || format === 'compact' || format === 'default') {
        setOutputConfig({ format });
      }
    } else if (arg === '--compact') {
      setOutputConfig({ format: 'compact' });
    } else if (arg === '--json') {
      setOutputConfig({ format: 'json' });
    } else if (arg === '--no-headers') {
      setOutputConfig({ showHeaders: false });
    } else if (arg === '--no-timestamps') {
      setOutputConfig({ showTimestamps: false });
    } else if (arg.startsWith('--max-items=')) {
      const max = parseInt(arg.split('=')[1] ?? '0', 10);
      if (!Number.isNaN(max)) {
        setOutputConfig({ maxItems: max });
      }
    } else {
      filtered.push(arg);
    }
  }

  return filtered;
}

// =============================================================================
// Output Functions (delegate to adapters)
// =============================================================================

/**
 * Get or create the output adapter for the current config.
 */
function getAdapter(): OutputAdapter {
  if (!currentAdapter) {
    currentAdapter = getOutputAdapter();
  }
  return currentAdapter;
}

/**
 * Output data in the configured format.
 */
export function output<T>(
  data: T,
  options: { summary?: string; command?: string; count?: number } = {}
): void {
  getAdapter().formatData(data, options);
}

/**
 * Output a simple message or status.
 */
export function outputMessage(message: string): void {
  getAdapter().formatMessage(message);
}

/**
 * Output an error.
 */
export function outputError(error: string | Error, code?: string): void {
  getAdapter().formatError(error, code);
}

/**
 * Output a list of items with optional count.
 */
export function outputList<T>(
  items: T[],
  options: {
    title?: string;
    command?: string;
    itemFormatter?: (item: T, index: number) => string;
    displayKey?: keyof T;
  } = {}
): void {
  getAdapter().formatList(items, options);
}

/**
 * Output entity state data.
 */
export function outputEntity(entity: {
  entity_id: string;
  state: string;
  attributes?: Record<string, unknown>;
}): void {
  getAdapter().formatEntity(entity);
}

/**
 * Output history/timeline data.
 */
export function outputTimeline<T extends { timestamp?: Date | string | number; when?: number }>(
  entries: T[],
  options: { title?: string; command?: string; entryFormatter?: (entry: T) => string } = {}
): void {
  getAdapter().formatTimeline(entries, options);
}

// =============================================================================
// Formatting Helpers
// =============================================================================

/**
 * Format an item for compact output.
 */
function formatCompact(item: unknown): string {
  if (item === null || item === undefined) return 'null';
  if (typeof item !== 'object') return String(item);

  const obj = item as Record<string, unknown>;

  // Entity state
  if ('entity_id' in obj && 'state' in obj) {
    return `${obj.entity_id}=${obj.state}`;
  }

  // Trace info
  if ('run_id' in obj && 'item_id' in obj) {
    const timestamp = obj.timestamp as { start?: string } | undefined;
    return `${obj.item_id} ${obj.run_id} ${timestamp?.start ?? ''}`.trim();
  }

  // Timeline entry
  if ('when' in obj || 'timestamp' in obj) {
    const time = getEntryTime(obj);
    const state = obj.state ?? obj.message ?? obj.value ?? '';
    return time ? `${formatTime(time)} ${state}` : String(state);
  }

  // Generic object - output key=value pairs
  const pairs = Object.entries(obj)
    .filter(([, v]) => v !== undefined && v !== null)
    .slice(0, 5)
    .map(([k, v]) => `${k}=${typeof v === 'object' ? JSON.stringify(v) : v}`);

  return pairs.join(' ');
}

/**
 * Format an item for default output.
 */
function formatDefault(item: unknown): string {
  if (item === null || item === undefined) return 'null';
  if (typeof item !== 'object') return String(item);
  return JSON.stringify(item, null, 2);
}

/**
 * Get timestamp from an entry.
 */
function getEntryTime(entry: Record<string, unknown>): Date | null {
  if (entry.timestamp instanceof Date) return entry.timestamp;
  if (typeof entry.timestamp === 'string') return new Date(entry.timestamp);
  if (typeof entry.timestamp === 'number') return new Date(entry.timestamp);
  if (typeof entry.when === 'number') return new Date(entry.when * 1000);
  if (typeof entry.lu === 'number') return new Date(entry.lu * 1000);
  if (typeof entry.last_updated === 'string') return new Date(entry.last_updated);
  return null;
}

/**
 * Format a date for display based on output config.
 */
function formatTime(date: Date): string {
  const { format } = outputConfig;
  if (format === 'compact') {
    return date.toISOString();
  }
  return date.toLocaleString();
}

/**
 * Check if we're in JSON output mode.
 */
export function isJsonOutput(): boolean {
  return outputConfig.format === 'json';
}

/**
 * Check if we're in compact output mode.
 */
export function isCompactOutput(): boolean {
  return outputConfig.format === 'compact';
}

/**
 * Check if we're in default output mode.
 */
export function isDefaultOutput(): boolean {
  return outputConfig.format === 'default';
}

// =============================================================================
// Strategy Pattern - Output Adapters
// =============================================================================

/**
 * Output adapter interface for the Strategy pattern.
 * Allows different output formatting strategies to be plugged in.
 */
export interface OutputAdapter {
  /** Format and output data */
  formatData<T>(data: T, options?: { summary?: string; command?: string; count?: number }): void;
  /** Format and output a list */
  formatList<T>(
    items: T[],
    options?: {
      title?: string;
      command?: string;
      itemFormatter?: (item: T, index: number) => string;
      displayKey?: keyof T;
    }
  ): void;
  /** Format and output an error */
  formatError(error: string | Error, code?: string): void;
  /** Format and output a message */
  formatMessage(message: string): void;
  /** Format and output an entity */
  formatEntity(entity: {
    entity_id: string;
    state: string;
    attributes?: Record<string, unknown>;
  }): void;
  /** Format and output a timeline */
  formatTimeline<T extends { timestamp?: Date | string | number; when?: number }>(
    entries: T[],
    options?: { title?: string; command?: string; entryFormatter?: (entry: T) => string }
  ): void;
}

/**
 * JSON output adapter - outputs data as JSON.
 */
export class JsonOutputAdapter implements OutputAdapter {
  formatData<T>(data: T, options?: { summary?: string; command?: string; count?: number }): void {
    const result: OutputResult<T> = { success: true, data };
    if (options?.command) result.command = options.command;
    if (options?.count !== undefined) result.count = options.count;
    if (options?.summary) result.summary = options.summary;
    console.log(JSON.stringify(result));
  }

  formatList<T>(items: T[], options?: { title?: string; command?: string }): void {
    const result: OutputResult<T[]> = {
      success: true,
      data: items,
      count: items.length,
    };
    if (options?.command) result.command = options.command;
    console.log(JSON.stringify(result));
  }

  formatError(error: string | Error, code?: string): void {
    const message = error instanceof Error ? error.message : error;
    console.log(JSON.stringify({ success: false, error: message, code }));
  }

  formatMessage(message: string): void {
    console.log(JSON.stringify({ success: true, message }));
  }

  formatEntity(entity: {
    entity_id: string;
    state: string;
    attributes?: Record<string, unknown>;
  }): void {
    console.log(JSON.stringify({ success: true, data: entity }));
  }

  formatTimeline<T>(entries: T[], options?: { title?: string; command?: string }): void {
    console.log(
      JSON.stringify({
        success: true,
        command: options?.command,
        data: entries,
        count: entries.length,
      })
    );
  }
}

/**
 * Compact output adapter - minimal, space-efficient output.
 */
export class CompactOutputAdapter implements OutputAdapter {
  formatData<T>(data: T, options?: { summary?: string }): void {
    if (options?.summary) {
      console.log(options.summary);
    }
    if (Array.isArray(data)) {
      const maxItems = outputConfig.maxItems;
      const items = maxItems > 0 ? data.slice(0, maxItems) : data;
      for (const item of items) {
        console.log(formatCompact(item));
      }
      if (maxItems > 0 && data.length > maxItems) {
        console.log(`... and ${data.length - maxItems} more`);
      }
    } else if (typeof data === 'object' && data !== null) {
      console.log(formatCompact(data));
    } else {
      console.log(String(data));
    }
  }

  formatList<T>(
    items: T[],
    options?: {
      title?: string;
      itemFormatter?: (item: T, index: number) => string;
      displayKey?: keyof T;
    }
  ): void {
    const { maxItems, showHeaders } = outputConfig;
    const count = items.length;
    const displayItems = maxItems > 0 ? items.slice(0, maxItems) : items;

    if (options?.title && showHeaders) {
      console.log(`${options.title}: ${count}`);
    }

    for (let i = 0; i < displayItems.length; i++) {
      const item = displayItems[i];
      if (item === undefined) continue;
      if (options?.itemFormatter) {
        console.log(options.itemFormatter(item, i));
      } else if (options?.displayKey && typeof item === 'object' && item !== null) {
        console.log(String(item[options.displayKey]));
      } else {
        console.log(formatCompact(item));
      }
    }

    if (maxItems > 0 && count > maxItems) {
      console.log(`+${count - maxItems} more`);
    }
  }

  formatError(error: string | Error, code?: string): void {
    const message = error instanceof Error ? error.message : error;
    console.error(code ? `[${code}] ${message}` : message);
  }

  formatMessage(message: string): void {
    console.log(message);
  }

  formatEntity(entity: { entity_id: string; state: string }): void {
    console.log(`${entity.entity_id}=${entity.state}`);
  }

  formatTimeline<T extends { timestamp?: Date | string | number; when?: number }>(
    entries: T[],
    options?: { title?: string; entryFormatter?: (entry: T) => string }
  ): void {
    const { maxItems, showHeaders, showTimestamps } = outputConfig;
    const count = entries.length;
    const displayEntries = maxItems > 0 ? entries.slice(0, maxItems) : entries;

    if (options?.title && showHeaders) {
      console.log(`${options.title}: ${count}`);
    }

    for (const entry of displayEntries) {
      if (options?.entryFormatter) {
        console.log(options.entryFormatter(entry));
      } else {
        const time = getEntryTime(entry as Record<string, unknown>);
        const timeStr = showTimestamps && time ? `${time.toISOString()} ` : '';
        console.log(`${timeStr}${formatCompact(entry)}`);
      }
    }

    if (maxItems > 0 && count > maxItems) {
      console.log(`+${count - maxItems} more`);
    }
  }
}

/**
 * Default output adapter - human-readable formatted output.
 */
export class DefaultOutputAdapter implements OutputAdapter {
  formatData<T>(data: T, options?: { summary?: string }): void {
    if (options?.summary && outputConfig.showHeaders) {
      console.log(options.summary);
    }
    if (Array.isArray(data)) {
      for (const item of data) {
        console.log(formatDefault(item));
      }
    } else if (typeof data === 'object' && data !== null) {
      console.log(JSON.stringify(data, null, 2));
    } else {
      console.log(String(data));
    }
  }

  formatList<T>(
    items: T[],
    options?: {
      title?: string;
      itemFormatter?: (item: T, index: number) => string;
      displayKey?: keyof T;
    }
  ): void {
    const { maxItems, showHeaders } = outputConfig;
    const count = items.length;
    const displayItems = maxItems > 0 ? items.slice(0, maxItems) : items;

    if (options?.title && showHeaders) {
      console.log(`${options.title}: ${count}\n`);
    }

    for (let i = 0; i < displayItems.length; i++) {
      const item = displayItems[i];
      if (item === undefined) continue;
      if (options?.itemFormatter) {
        console.log(options.itemFormatter(item, i));
      } else if (options?.displayKey && typeof item === 'object' && item !== null) {
        console.log(String(item[options.displayKey]));
      } else {
        console.log(formatDefault(item));
      }
    }

    if (maxItems > 0 && count > maxItems) {
      console.log(`\n... and ${count - maxItems} more`);
    }
  }

  formatError(error: string | Error, code?: string): void {
    const message = error instanceof Error ? error.message : error;
    if (code) {
      console.error(`Error [${code}]: ${message}`);
    } else {
      console.error(`Error: ${message}`);
    }
  }

  formatMessage(message: string): void {
    console.log(message);
  }

  formatEntity(entity: {
    entity_id: string;
    state: string;
    attributes?: Record<string, unknown>;
  }): void {
    console.log(JSON.stringify(entity, null, 2));
  }

  formatTimeline<T extends { timestamp?: Date | string | number; when?: number }>(
    entries: T[],
    options?: { title?: string; entryFormatter?: (entry: T) => string }
  ): void {
    const { maxItems, showHeaders, showTimestamps } = outputConfig;
    const count = entries.length;
    const displayEntries = maxItems > 0 ? entries.slice(0, maxItems) : entries;

    if (options?.title && showHeaders) {
      console.log(`${options.title}:\n`);
    }

    for (const entry of displayEntries) {
      if (options?.entryFormatter) {
        console.log(options.entryFormatter(entry));
      } else {
        const time = getEntryTime(entry as Record<string, unknown>);
        const timeStr = showTimestamps && time ? `${time.toLocaleString()} ` : '';
        console.log(`${timeStr}${formatCompact(entry)}`);
      }
    }

    if (maxItems > 0 && count > maxItems) {
      console.log(`\n... and ${count - maxItems} more`);
    }

    if (showHeaders) {
      console.log(`\nTotal: ${count} entries`);
    }
  }
}

/**
 * Get the appropriate output adapter based on current configuration.
 */
export function getOutputAdapter(): OutputAdapter {
  switch (outputConfig.format) {
    case 'json':
      return new JsonOutputAdapter();
    case 'compact':
      return new CompactOutputAdapter();
    default:
      return new DefaultOutputAdapter();
  }
}

// =============================================================================
// Table Formatting Utility
// =============================================================================

/**
 * Column definition for table formatting.
 */
export interface TableColumn<T> {
  /** Column header text */
  readonly header: string;
  /** Function to extract cell value from row data */
  readonly value: (row: T) => string;
  /** Minimum width for the column */
  readonly minWidth?: number;
  /** Maximum width for the column */
  readonly maxWidth?: number;
  /** Text alignment */
  readonly align?: 'left' | 'right' | 'center';
}

/**
 * Format data as a table.
 * Only used in default output mode.
 *
 * @param rows - Array of row data
 * @param columns - Column definitions
 * @param options - Table options
 */
export function formatTable<T>(
  rows: readonly T[],
  columns: readonly TableColumn<T>[],
  options: { showHeaders?: boolean; maxRows?: number } = {}
): string {
  const { showHeaders = true, maxRows } = options;
  const displayRows = maxRows && maxRows > 0 ? rows.slice(0, maxRows) : rows;

  // Calculate column widths
  const widths = columns.map((col) => {
    const headerWidth = col.header.length;
    const maxDataWidth = displayRows.reduce((max, row) => {
      const value = col.value(row);
      return Math.max(max, value.length);
    }, 0);
    const width = Math.max(headerWidth, maxDataWidth, col.minWidth ?? 0);
    return col.maxWidth ? Math.min(width, col.maxWidth) : width;
  });

  const lines: string[] = [];

  // Header row
  if (showHeaders) {
    const headerRow = columns
      .map((col, i) => padString(col.header, widths[i] ?? 0, col.align ?? 'left'))
      .join(' | ');
    lines.push(headerRow);

    // Separator
    const separator = widths.map((w) => '-'.repeat(w)).join('-+-');
    lines.push(separator);
  }

  // Data rows
  for (const row of displayRows) {
    const dataRow = columns
      .map((col, i) => {
        const value = col.value(row);
        const width = widths[i] ?? 0;
        const truncated = value.length > width ? `${value.slice(0, width - 1)}…` : value;
        return padString(truncated, width, col.align ?? 'left');
      })
      .join(' | ');
    lines.push(dataRow);
  }

  // Overflow indicator
  if (maxRows && rows.length > maxRows) {
    lines.push(`... and ${rows.length - maxRows} more rows`);
  }

  return lines.join('\n');
}

/**
 * Pad a string to a specific width with alignment.
 */
function padString(str: string, width: number, align: 'left' | 'right' | 'center'): string {
  if (str.length >= width) return str;

  const padding = width - str.length;

  switch (align) {
    case 'right':
      return ' '.repeat(padding) + str;
    case 'center': {
      const leftPad = Math.floor(padding / 2);
      const rightPad = padding - leftPad;
      return ' '.repeat(leftPad) + str + ' '.repeat(rightPad);
    }
    default:
      return str + ' '.repeat(padding);
  }
}

// =============================================================================
// Progress and Spinner Utilities
// =============================================================================

/**
 * Progress indicator for long-running operations.
 * Only shows in non-JSON modes.
 */
export class ProgressIndicator {
  private readonly message: string;
  private readonly total: number;
  private current = 0;
  private intervalId: ReturnType<typeof setInterval> | null = null;

  constructor(message: string, total: number) {
    this.message = message;
    this.total = total;
  }

  /**
   * Start the progress indicator.
   */
  start(): void {
    if (outputConfig.format === 'json') return;
    this.render();
  }

  /**
   * Update progress.
   */
  update(current: number): void {
    this.current = current;
    if (outputConfig.format !== 'json') {
      this.render();
    }
  }

  /**
   * Increment progress by 1.
   */
  increment(): void {
    this.update(this.current + 1);
  }

  /**
   * Complete the progress.
   */
  complete(message?: string): void {
    if (this.intervalId) {
      clearInterval(this.intervalId);
      this.intervalId = null;
    }

    if (outputConfig.format !== 'json') {
      process.stdout.write(`\r${' '.repeat(80)}\r`);
      if (message) {
        console.log(message);
      }
    }
  }

  private render(): void {
    const percentage = this.total > 0 ? Math.round((this.current / this.total) * 100) : 0;
    const bar = this.renderBar(percentage);
    process.stdout.write(`\r${this.message} ${bar} ${percentage}% (${this.current}/${this.total})`);
  }

  private renderBar(percentage: number): string {
    const width = 20;
    const filled = Math.round(width * (percentage / 100));
    const empty = width - filled;
    return `[${'█'.repeat(filled)}${'░'.repeat(empty)}]`;
  }
}

/**
 * Spinner for indeterminate progress.
 * Only shows in non-JSON modes.
 */
export class Spinner {
  private readonly message: string;
  private readonly frames = ['⠋', '⠙', '⠹', '⠸', '⠼', '⠴', '⠦', '⠧', '⠇', '⠏'];
  private frameIndex = 0;
  private intervalId: ReturnType<typeof setInterval> | null = null;

  constructor(message: string) {
    this.message = message;
  }

  /**
   * Start the spinner.
   */
  start(): void {
    if (outputConfig.format === 'json') return;

    this.intervalId = setInterval(() => {
      this.frameIndex = (this.frameIndex + 1) % this.frames.length;
      process.stdout.write(`\r${this.frames[this.frameIndex]} ${this.message}`);
    }, 80);
  }

  /**
   * Stop the spinner.
   */
  stop(message?: string): void {
    if (this.intervalId) {
      clearInterval(this.intervalId);
      this.intervalId = null;
    }

    if (outputConfig.format !== 'json') {
      process.stdout.write(`\r${' '.repeat(this.message.length + 4)}\r`);
      if (message) {
        console.log(message);
      }
    }
  }

  /**
   * Stop with success indicator.
   */
  succeed(message?: string): void {
    this.stop(`✓ ${message ?? this.message}`);
  }

  /**
   * Stop with failure indicator.
   */
  fail(message?: string): void {
    this.stop(`✗ ${message ?? this.message}`);
  }
}
