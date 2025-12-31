/**
 * Type definitions for the Home Assistant WebSocket API client.
 * @module types
 */

import type WebSocket from 'ws';

// =============================================================================
// WebSocket Message Types
// =============================================================================

/**
 * A pending WebSocket request awaiting a response.
 * Maps message IDs to their resolve/reject handlers.
 */
export interface PendingRequest {
  readonly resolve: (value: unknown) => void;
  readonly reject: (error: Error) => void;
}

/**
 * A message received from the Home Assistant WebSocket API.
 * The structure varies based on the message type.
 */
export interface HAMessage {
  readonly id?: number;
  readonly type: string;
  readonly success?: boolean;
  readonly result?: unknown;
  readonly error?: { readonly message?: string; readonly code?: string };
  readonly message?: string;
  readonly event?: { readonly result?: unknown; readonly variables?: Record<string, unknown> };
}

// =============================================================================
// Entity and State Types
// =============================================================================

/**
 * A Home Assistant entity state.
 * Represents the current state and attributes of an entity.
 */
export interface HAState {
  readonly entity_id: string;
  readonly state: string;
  readonly attributes?: Record<string, unknown>;
  readonly last_changed?: string;
  readonly last_updated?: string;
  readonly context?: {
    readonly id: string;
    readonly parent_id?: string;
    readonly user_id?: string;
  };
}

/**
 * Home Assistant configuration.
 * Contains system-wide settings and information.
 */
export interface HAConfig {
  readonly version: string;
  readonly location_name: string;
  readonly time_zone: string;
  readonly unit_system: Record<string, string>;
  readonly state: string;
  readonly components: readonly string[];
}

// =============================================================================
// History and Logbook Types
// =============================================================================

/**
 * A logbook entry from Home Assistant.
 * Represents a single event in the logbook.
 */
export interface LogbookEntry {
  readonly when: number;
  readonly entity_id?: string;
  readonly state?: string;
  readonly message?: string;
  readonly context_id?: string;
}

/**
 * A historical state entry.
 * May use compact (lu, lc, s, a) or full property names depending on API options.
 */
export interface HistoryState {
  /** Last updated timestamp (Unix epoch, compact format) */
  readonly lu?: number;
  /** Last changed timestamp (Unix epoch, compact format) */
  readonly lc?: number;
  /** Last updated timestamp (ISO string, full format) */
  readonly last_updated?: string;
  /** Last changed timestamp (ISO string, full format) */
  readonly last_changed?: string;
  /** State value (compact format) */
  readonly s?: string;
  /** State value (full format) */
  readonly state?: string;
  /** Attributes (compact format) */
  readonly a?: Record<string, unknown>;
  /** Attributes (full format) */
  readonly attributes?: Record<string, unknown>;
}

// =============================================================================
// Automation and Trace Types
// =============================================================================

/**
 * Summary information about an automation trace.
 * Returned when listing traces.
 */
export interface TraceInfo {
  readonly item_id: string;
  readonly run_id: string;
  readonly state?: string;
  readonly script_execution?: string;
  readonly timestamp: { readonly start: string };
  readonly context?: { readonly id: string };
}

/**
 * Result from an action step in an automation trace.
 * Contains response data, state changes, and any errors.
 */
export interface TraceActionResult {
  readonly error?: unknown;
  readonly response?: unknown;
  readonly params?: Record<string, unknown>;
  readonly running_script?: boolean;
  readonly limit?: number;
  readonly enabled?: boolean;
}

/**
 * A single step within an automation trace.
 * Contains execution details, errors, and evaluated variables.
 */
export interface TraceStep {
  readonly path?: string;
  readonly error?: unknown;
  readonly result?: TraceActionResult;
  readonly variables?: Record<string, unknown>;
  readonly changed_variables?: Record<string, unknown>;
  readonly timestamp?: string;
}

/**
 * Trigger information captured during automation execution.
 * Contains the triggering entity, state changes, and context.
 */
export interface TraceTrigger {
  readonly id?: string;
  readonly idx?: string;
  readonly alias?: string;
  readonly platform?: string;
  readonly entity_id?: string;
  readonly from_state?: {
    readonly entity_id?: string;
    readonly state: string;
    readonly attributes?: Record<string, unknown>;
    readonly last_changed?: string;
    readonly last_updated?: string;
  };
  readonly to_state?: {
    readonly entity_id?: string;
    readonly state: string;
    readonly attributes?: Record<string, unknown>;
    readonly last_changed?: string;
    readonly last_updated?: string;
  };
  readonly for?:
    | string
    | { readonly hours?: number; readonly minutes?: number; readonly seconds?: number };
  readonly description?: string;
}

/**
 * Detailed trace information for an automation run.
 * Contains the full execution path, trigger context, and evaluated values.
 */
export interface TraceDetail {
  readonly script_execution?: string;
  readonly error?: string;
  readonly trace?: Record<string, readonly TraceStep[]>;
  readonly config?: {
    readonly id?: string;
    readonly alias?: string;
    readonly trigger?: readonly unknown[];
    readonly condition?: readonly unknown[];
    readonly action?: readonly unknown[];
  };
  readonly context?: {
    readonly id?: string;
    readonly parent_id?: string;
    readonly user_id?: string;
  };
  readonly trigger?: TraceTrigger;
  readonly run_id?: string;
  readonly domain?: string;
  readonly item_id?: string;
  readonly timestamp?: {
    readonly start: string;
    readonly finish?: string;
  };
}

/**
 * Automation configuration from automations.yaml.
 * Includes blueprint usage information if applicable.
 */
export interface AutomationConfig {
  readonly id: string;
  readonly alias?: string;
  readonly use_blueprint?: {
    readonly path: string;
    readonly input?: Record<string, unknown>;
  };
}

/**
 * Blueprint input definition.
 * Describes an input parameter for a blueprint.
 */
export interface BlueprintInput {
  readonly name?: string;
  readonly default?: unknown;
  readonly description?: string;
  readonly selector?: unknown;
  /** Nested inputs for grouped input sections */
  readonly input?: Record<string, BlueprintInput>;
}

// =============================================================================
// Utility Types
// =============================================================================

/**
 * Parsed time arguments from command line.
 * Contains optional from/to times and remaining arguments.
 */
export interface TimeArgs {
  readonly fromTime: Date | null;
  readonly toTime: Date | null;
  readonly filteredArgs: readonly string[];
}

/**
 * Context passed to all command handlers.
 * Contains the WebSocket connection, arguments, and time filters.
 */
export interface CommandContext {
  /** Active WebSocket connection to Home Assistant */
  readonly ws: WebSocket;
  /** Command arguments (command name at index 0) */
  readonly args: readonly string[];
  /** Optional start time filter */
  readonly fromTime: Date | null;
  /** Optional end time filter */
  readonly toTime: Date | null;
}

/**
 * Command handler function signature.
 * All command handlers must conform to this type.
 */
export type CommandHandler = (ctx: CommandContext) => Promise<void>;

// =============================================================================
// Registry Types
// =============================================================================

/**
 * Entity registry entry.
 * Contains metadata about a registered entity.
 */
export interface EntityEntry {
  readonly entity_id: string;
  readonly name?: string;
  readonly original_name?: string;
  readonly platform?: string;
  readonly disabled_by?: string;
}

/**
 * Device registry entry.
 * Contains metadata about a registered device.
 */
export interface DeviceEntry {
  readonly id: string;
  readonly name?: string;
  readonly name_by_user?: string;
  readonly manufacturer?: string;
  readonly model?: string;
  readonly area_id?: string;
}

/**
 * Area registry entry.
 * Contains metadata about a registered area.
 */
export interface AreaEntry {
  readonly area_id: string;
  readonly name: string;
  readonly aliases?: readonly string[];
}

/**
 * System log entry.
 * Contains a log message from Home Assistant.
 */
export interface SysLogEntry {
  readonly level?: string;
  readonly source?: readonly string[];
  readonly message?: string;
}

/**
 * Statistics entry for a sensor.
 * Contains min/max/mean values for a time period.
 */
export interface StatEntry {
  readonly start: string;
  readonly min: number;
  readonly max: number;
  readonly mean?: number;
}

// =============================================================================
// YAML Module Type
// =============================================================================

/**
 * Type definition for the js-yaml module.
 * Used for lazy-loading YAML parsing functionality.
 */
export type YamlModule = {
  load: (content: string, options?: { schema?: unknown }) => unknown;
  Type: new (
    tag: string,
    options: { kind: string; construct: (data: string) => unknown }
  ) => unknown;
  DEFAULT_SCHEMA: { extend: (types: unknown[]) => unknown };
};
