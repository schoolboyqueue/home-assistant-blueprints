/**
 * Type definitions for the Home Assistant WebSocket API client.
 * @module types
 */

import type WebSocket from 'ws';

// =============================================================================
// Utility Types and Type Helpers
// =============================================================================

/**
 * Creates a branded/opaque type for nominal typing.
 * This prevents mixing up values that have the same underlying type.
 *
 * @example
 * ```typescript
 * type UserId = Brand<string, 'UserId'>;
 * type EntityId = Brand<string, 'EntityId'>;
 *
 * const userId: UserId = 'user123' as UserId;
 * const entityId: EntityId = 'light.kitchen' as EntityId;
 *
 * // Type error: cannot assign UserId to EntityId
 * const wrong: EntityId = userId;
 * ```
 */
export type Brand<T, B> = T & { readonly __brand: B };

/**
 * Branded type for entity IDs.
 * Ensures type safety when working with Home Assistant entity identifiers.
 */
export type EntityId = Brand<string, 'EntityId'>;

/**
 * Branded type for context IDs.
 * Used to track causation chains in Home Assistant.
 */
export type ContextId = Brand<string, 'ContextId'>;

/**
 * Branded type for automation/script run IDs.
 */
export type RunId = Brand<string, 'RunId'>;

/**
 * Branded type for message IDs in WebSocket communication.
 */
export type MessageId = Brand<number, 'MessageId'>;

/**
 * Helper to create an EntityId from a string.
 */
export function entityId(id: string): EntityId {
  return id as EntityId;
}

/**
 * Helper to create a ContextId from a string.
 */
export function contextId(id: string): ContextId {
  return id as ContextId;
}

/**
 * Helper to create a RunId from a string.
 */
export function runId(id: string): RunId {
  return id as RunId;
}

// =============================================================================
// Result Type Pattern
// =============================================================================

/**
 * A discriminated union representing either a successful result or an error.
 * This pattern provides explicit error handling without exceptions.
 *
 * @example
 * ```typescript
 * async function getEntity(id: EntityId): Promise<Result<HAState, EntityNotFoundError>> {
 *   const state = states.find(s => s.entity_id === id);
 *   if (!state) {
 *     return err(new EntityNotFoundError(id));
 *   }
 *   return ok(state);
 * }
 *
 * const result = await getEntity(entityId('light.kitchen'));
 * if (result.ok) {
 *   console.log(result.value.state);
 * } else {
 *   console.error(result.error.message);
 * }
 * ```
 */
export type Result<T, E = Error> =
  | { readonly ok: true; readonly value: T }
  | { readonly ok: false; readonly error: E };

/**
 * Creates a successful Result.
 */
export function ok<T>(value: T): Result<T, never> {
  return { ok: true, value };
}

/**
 * Creates a failed Result.
 */
export function err<E>(error: E): Result<never, E> {
  return { ok: false, error };
}

/**
 * Type guard to check if a Result is successful.
 */
export function isOk<T, E>(
  result: Result<T, E>
): result is { readonly ok: true; readonly value: T } {
  return result.ok;
}

/**
 * Type guard to check if a Result is an error.
 */
export function isErr<T, E>(
  result: Result<T, E>
): result is { readonly ok: false; readonly error: E } {
  return !result.ok;
}

/**
 * Unwraps a Result, throwing an error if it's not ok.
 * Use with caution - prefer pattern matching with isOk/isErr.
 */
export function unwrap<T, E extends Error>(result: Result<T, E>): T {
  if (result.ok) {
    return result.value;
  }
  throw result.error;
}

/**
 * Unwraps a Result or returns a default value if it's an error.
 */
export function unwrapOr<T, E>(result: Result<T, E>, defaultValue: T): T {
  return result.ok ? result.value : defaultValue;
}

/**
 * Maps a successful Result value to a new value.
 */
export function mapResult<T, U, E>(result: Result<T, E>, fn: (value: T) => U): Result<U, E> {
  return result.ok ? ok(fn(result.value)) : result;
}

/**
 * Maps an error Result to a new error.
 */
export function mapError<T, E, F>(result: Result<T, E>, fn: (error: E) => F): Result<T, F> {
  return result.ok ? result : err(fn(result.error));
}

/**
 * Chains Results together, short-circuiting on first error.
 */
export function flatMap<T, U, E>(
  result: Result<T, E>,
  fn: (value: T) => Result<U, E>
): Result<U, E> {
  return result.ok ? fn(result.value) : result;
}

// =============================================================================
// Option Type Pattern
// =============================================================================

/**
 * Represents an optional value - either Some with a value or None.
 * Provides safer handling than null/undefined.
 */
export type Option<T> = { readonly some: true; readonly value: T } | { readonly some: false };

/**
 * Creates a Some option with a value.
 */
export function some<T>(value: T): Option<T> {
  return { some: true, value };
}

/**
 * Creates a None option (no value).
 */
export function none<T = never>(): Option<T> {
  return { some: false };
}

/**
 * Creates an Option from a nullable value.
 */
export function fromNullable<T>(value: T | null | undefined): Option<T> {
  return value != null ? some(value) : none();
}

/**
 * Type guard to check if an Option has a value.
 */
export function isSome<T>(option: Option<T>): option is { readonly some: true; readonly value: T } {
  return option.some;
}

/**
 * Type guard to check if an Option is empty.
 */
export function isNone<T>(option: Option<T>): option is { readonly some: false } {
  return !option.some;
}

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
 * Result of subscribing to a trigger.
 * Contains the subscription ID and cleanup function.
 */
export interface TriggerSubscription {
  /** The subscription ID for this trigger */
  readonly subscriptionId: number;
  /** Cleanup function to remove the event listener */
  readonly cleanup: () => void;
}

/**
 * Result of calculateTimeRange function containing start and end times.
 */
export interface TimeRange {
  readonly startTime: Date;
  readonly endTime: Date;
}

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
// Monitoring Types
// =============================================================================

/**
 * A recorded state change entry for historical tracking.
 */
export interface StateChangeRecord {
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
export interface AnomalyConfig {
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
export interface EntityStats {
  readonly count: number;
  readonly min: number;
  readonly max: number;
  readonly mean: number;
  readonly stdDev: number;
  readonly lastValues: readonly number[];
}

// =============================================================================
// Automation Handler Types
// =============================================================================

/**
 * Result from the automation/config API call.
 */
export interface AutomationConfigResult {
  readonly config: Record<string, unknown>;
}

/**
 * Trigger information extracted from trace variables for trace-vars command.
 */
export interface TriggerInfo {
  readonly id?: string;
  readonly entity_id?: string;
  readonly from_state?: { readonly state: string };
  readonly to_state?: { readonly state: string };
}

/**
 * Trigger variables received from watch/monitor state subscriptions.
 */
export interface StateTriggerVariables {
  readonly trigger?: {
    readonly entity_id?: string;
    readonly from_state?: { readonly state: string; readonly attributes?: Record<string, unknown> };
    readonly to_state?: { readonly state: string; readonly attributes?: Record<string, unknown> };
  };
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

// =============================================================================
// Runtime Validation Types and Helpers
// =============================================================================

/**
 * Validation error containing path and message.
 */
export interface ValidationError {
  readonly path: string;
  readonly message: string;
  readonly expected?: string;
  readonly received?: string;
}

/**
 * Result of schema validation.
 */
export type ValidationResult<T> = Result<T, readonly ValidationError[]>;

/**
 * Schema definition for runtime type validation.
 * Provides a lightweight validation system without external dependencies.
 */
export interface Schema<T> {
  /** Validates unknown input and returns a Result */
  readonly validate: (value: unknown, path?: string) => ValidationResult<T>;
  /** Optional description for error messages */
  readonly description?: string;
}

/**
 * Schema builder functions for common types.
 */
export const Schema = {
  /**
   * Creates a string schema.
   */
  string(options?: { minLength?: number; maxLength?: number; pattern?: RegExp }): Schema<string> {
    return {
      description: 'string',
      validate(value: unknown, path = ''): ValidationResult<string> {
        if (typeof value !== 'string') {
          return err([
            { path, message: 'Expected string', expected: 'string', received: typeof value },
          ]);
        }
        if (options?.minLength !== undefined && value.length < options.minLength) {
          return err([
            { path, message: `String must be at least ${options.minLength} characters` },
          ]);
        }
        if (options?.maxLength !== undefined && value.length > options.maxLength) {
          return err([{ path, message: `String must be at most ${options.maxLength} characters` }]);
        }
        if (options?.pattern !== undefined && !options.pattern.test(value)) {
          return err([{ path, message: `String must match pattern ${options.pattern}` }]);
        }
        return ok(value);
      },
    };
  },

  /**
   * Creates a number schema.
   */
  number(options?: { min?: number; max?: number; integer?: boolean }): Schema<number> {
    return {
      description: 'number',
      validate(value: unknown, path = ''): ValidationResult<number> {
        if (typeof value !== 'number' || Number.isNaN(value)) {
          return err([
            { path, message: 'Expected number', expected: 'number', received: typeof value },
          ]);
        }
        if (options?.integer && !Number.isInteger(value)) {
          return err([{ path, message: 'Expected integer' }]);
        }
        if (options?.min !== undefined && value < options.min) {
          return err([{ path, message: `Number must be at least ${options.min}` }]);
        }
        if (options?.max !== undefined && value > options.max) {
          return err([{ path, message: `Number must be at most ${options.max}` }]);
        }
        return ok(value);
      },
    };
  },

  /**
   * Creates a boolean schema.
   */
  boolean(): Schema<boolean> {
    return {
      description: 'boolean',
      validate(value: unknown, path = ''): ValidationResult<boolean> {
        if (typeof value !== 'boolean') {
          return err([
            { path, message: 'Expected boolean', expected: 'boolean', received: typeof value },
          ]);
        }
        return ok(value);
      },
    };
  },

  /**
   * Creates an array schema.
   */
  array<T>(itemSchema: Schema<T>): Schema<readonly T[]> {
    return {
      description: `array of ${itemSchema.description ?? 'items'}`,
      validate(value: unknown, path = ''): ValidationResult<readonly T[]> {
        if (!Array.isArray(value)) {
          return err([
            { path, message: 'Expected array', expected: 'array', received: typeof value },
          ]);
        }
        const results: T[] = [];
        const errors: ValidationError[] = [];
        for (let i = 0; i < value.length; i++) {
          const result = itemSchema.validate(value[i], `${path}[${i}]`);
          if (result.ok) {
            results.push(result.value);
          } else {
            errors.push(...result.error);
          }
        }
        return errors.length > 0 ? err(errors) : ok(results);
      },
    };
  },

  /**
   * Creates an object schema.
   */
  object<T extends Record<string, unknown>>(shape: { [K in keyof T]: Schema<T[K]> }): Schema<T> {
    return {
      description: 'object',
      validate(value: unknown, path = ''): ValidationResult<T> {
        if (typeof value !== 'object' || value === null || Array.isArray(value)) {
          return err([
            { path, message: 'Expected object', expected: 'object', received: typeof value },
          ]);
        }
        const obj = value as Record<string, unknown>;
        const result: Record<string, unknown> = {};
        const errors: ValidationError[] = [];

        for (const [key, schema] of Object.entries(shape) as [string, Schema<unknown>][]) {
          const fieldPath = path ? `${path}.${key}` : key;
          const fieldResult = schema.validate(obj[key], fieldPath);
          if (fieldResult.ok) {
            result[key] = fieldResult.value;
          } else {
            errors.push(...fieldResult.error);
          }
        }

        return errors.length > 0 ? err(errors) : ok(result as T);
      },
    };
  },

  /**
   * Creates an optional schema.
   */
  optional<T>(schema: Schema<T>): Schema<T | undefined> {
    return {
      description: `optional ${schema.description ?? 'value'}`,
      validate(value: unknown, path = ''): ValidationResult<T | undefined> {
        if (value === undefined) {
          return ok(undefined);
        }
        return schema.validate(value, path);
      },
    };
  },

  /**
   * Creates a nullable schema.
   */
  nullable<T>(schema: Schema<T>): Schema<T | null> {
    return {
      description: `nullable ${schema.description ?? 'value'}`,
      validate(value: unknown, path = ''): ValidationResult<T | null> {
        if (value === null) {
          return ok(null);
        }
        return schema.validate(value, path);
      },
    };
  },

  /**
   * Creates a literal schema for exact value matching.
   */
  literal<T extends string | number | boolean>(literalValue: T): Schema<T> {
    return {
      description: `literal ${JSON.stringify(literalValue)}`,
      validate(value: unknown, path = ''): ValidationResult<T> {
        if (value !== literalValue) {
          return err([
            {
              path,
              message: `Expected ${JSON.stringify(literalValue)}`,
              expected: String(literalValue),
              received: String(value),
            },
          ]);
        }
        return ok(value as T);
      },
    };
  },

  /**
   * Creates a union schema for multiple possible types.
   */
  union<T extends readonly Schema<unknown>[]>(
    schemas: T
  ): Schema<T[number] extends Schema<infer U> ? U : never> {
    return {
      description: schemas.map((s) => s.description ?? 'unknown').join(' | '),
      validate(
        value: unknown,
        path = ''
      ): ValidationResult<T[number] extends Schema<infer U> ? U : never> {
        for (const schema of schemas) {
          const result = schema.validate(value, path);
          if (result.ok) {
            return result as ValidationResult<T[number] extends Schema<infer U> ? U : never>;
          }
        }
        return err([
          { path, message: `Expected one of: ${schemas.map((s) => s.description).join(', ')}` },
        ]);
      },
    };
  },

  /**
   * Creates a record schema (object with string keys).
   */
  record<T>(valueSchema: Schema<T>): Schema<Record<string, T>> {
    return {
      description: `record of ${valueSchema.description ?? 'values'}`,
      validate(value: unknown, path = ''): ValidationResult<Record<string, T>> {
        if (typeof value !== 'object' || value === null || Array.isArray(value)) {
          return err([
            { path, message: 'Expected object', expected: 'object', received: typeof value },
          ]);
        }
        const obj = value as Record<string, unknown>;
        const result: Record<string, T> = {};
        const errors: ValidationError[] = [];

        for (const [key, val] of Object.entries(obj)) {
          const fieldPath = path ? `${path}.${key}` : key;
          const fieldResult = valueSchema.validate(val, fieldPath);
          if (fieldResult.ok) {
            result[key] = fieldResult.value;
          } else {
            errors.push(...fieldResult.error);
          }
        }

        return errors.length > 0 ? err(errors) : ok(result);
      },
    };
  },

  /**
   * Creates an unknown schema that accepts any value.
   */
  unknown(): Schema<unknown> {
    return {
      description: 'unknown',
      validate(value: unknown): ValidationResult<unknown> {
        return ok(value);
      },
    };
  },
};

// =============================================================================
// Common Schemas for Home Assistant Types
// =============================================================================

/**
 * Schema for HAState objects received from the API.
 */
export const HAStateSchema: Schema<HAState> = Schema.object({
  entity_id: Schema.string(),
  state: Schema.string(),
  attributes: Schema.optional(Schema.record(Schema.unknown())),
  last_changed: Schema.optional(Schema.string()),
  last_updated: Schema.optional(Schema.string()),
  context: Schema.optional(
    Schema.object({
      id: Schema.string(),
      parent_id: Schema.optional(Schema.string()),
      user_id: Schema.optional(Schema.string()),
    })
  ),
}) as Schema<HAState>;

/**
 * Schema for LogbookEntry objects.
 */
export const LogbookEntrySchema: Schema<LogbookEntry> = Schema.object({
  when: Schema.number(),
  entity_id: Schema.optional(Schema.string()),
  state: Schema.optional(Schema.string()),
  message: Schema.optional(Schema.string()),
  context_id: Schema.optional(Schema.string()),
}) as Schema<LogbookEntry>;

// =============================================================================
// Pipe and Compose Utilities
// =============================================================================

/**
 * Pipes a value through a series of functions.
 * Each function receives the output of the previous one.
 *
 * @example
 * ```typescript
 * const result = pipe(
 *   value,
 *   addOne,
 *   double,
 *   toString
 * );
 * ```
 */
export function pipe<A>(a: A): A;
export function pipe<A, B>(a: A, ab: (a: A) => B): B;
export function pipe<A, B, C>(a: A, ab: (a: A) => B, bc: (b: B) => C): C;
export function pipe<A, B, C, D>(a: A, ab: (a: A) => B, bc: (b: B) => C, cd: (c: C) => D): D;
export function pipe<A, B, C, D, E>(
  a: A,
  ab: (a: A) => B,
  bc: (b: B) => C,
  cd: (c: C) => D,
  de: (d: D) => E
): E;
export function pipe(value: unknown, ...fns: Array<(arg: unknown) => unknown>): unknown {
  return fns.reduce((acc, fn) => fn(acc), value);
}

/**
 * Composes functions from right to left.
 *
 * @example
 * ```typescript
 * const addOneAndDouble = compose(double, addOne);
 * addOneAndDouble(5); // (5 + 1) * 2 = 12
 * ```
 */
export function compose<A, B>(ab: (a: A) => B): (a: A) => B;
export function compose<A, B, C>(bc: (b: B) => C, ab: (a: A) => B): (a: A) => C;
export function compose<A, B, C, D>(cd: (c: C) => D, bc: (b: B) => C, ab: (a: A) => B): (a: A) => D;
export function compose(...fns: Array<(arg: unknown) => unknown>): (arg: unknown) => unknown {
  return (value: unknown) => fns.reduceRight((acc, fn) => fn(acc), value);
}
