/**
 * Home Assistant WebSocket Client
 *
 * A modern TypeScript client for the Home Assistant WebSocket API.
 *
 * @module ha-ws-client
 */

// =============================================================================
// Core Types and Utilities
// =============================================================================

export * from './errors.js';
export * from './types.js';

// =============================================================================
// WebSocket Client
// =============================================================================

export { nextId, pendingRequests, sendMessage, subscribeToTrigger } from './client.js';

// =============================================================================
// Output System
// =============================================================================

export type {
  OutputAdapter,
  OutputConfig,
  OutputFormat,
  OutputResult,
  TableColumn,
} from './output.js';
export {
  CompactOutputAdapter,
  DefaultOutputAdapter,
  formatTable,
  getOutputAdapter,
  getOutputConfig,
  isCompactOutput,
  isDefaultOutput,
  isJsonOutput,
  JsonOutputAdapter,
  output,
  outputEntity,
  outputError,
  outputList,
  outputMessage,
  outputTimeline,
  ProgressIndicator,
  parseOutputArgs,
  Spinner,
  setOutputConfig,
} from './output.js';

// =============================================================================
// Utilities
// =============================================================================

export {
  calculateTimeRange,
  formatEntityAttributes,
  getYamlModule,
  parseFlexibleDate,
  parseJsonArg,
  parseTimeArgs,
  requireArg,
} from './utils.js';

// =============================================================================
// Command Handlers
// =============================================================================

export { commandHandlers } from './handlers/index.js';
