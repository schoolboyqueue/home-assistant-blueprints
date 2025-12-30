/**
 * Command handler exports and registry.
 * @module handlers
 */

import type { CommandHandler } from '../types.js';

export {
  handleAutomationConfig,
  handleBlueprintInputs,
  handleContext,
  handleTrace,
  handleTraces,
  handleTraceVars,
  handleWatch,
} from './automation.js';
// Re-export all handlers for direct use
export {
  handleCall,
  handleConfig,
  handlePing,
  handleServices,
  handleState,
  handleStates,
  handleStatesFilter,
  handleStatesJson,
  handleTemplate,
} from './basic.js';
export {
  handleAttrs,
  handleHistory,
  handleHistoryFull,
  handleLogbook,
  handleStats,
  handleSyslog,
  handleTimeline,
} from './history.js';
export {
  handleAreas,
  handleDevices,
  handleEntities,
} from './registry.js';

import {
  handleAutomationConfig,
  handleBlueprintInputs,
  handleContext,
  handleTrace,
  handleTraces,
  handleTraceVars,
  handleWatch,
} from './automation.js';
// Import handlers for registry
import {
  handleCall,
  handleConfig,
  handlePing,
  handleServices,
  handleState,
  handleStates,
  handleStatesFilter,
  handleStatesJson,
  handleTemplate,
} from './basic.js';
import {
  handleAttrs,
  handleHistory,
  handleHistoryFull,
  handleLogbook,
  handleStats,
  handleSyslog,
  handleTimeline,
} from './history.js';
import { handleAreas, handleDevices, handleEntities } from './registry.js';

/**
 * Registry mapping command names to their handler functions.
 * Used by the main entry point to dispatch commands.
 */
export const commandHandlers: Record<string, CommandHandler> = {
  // Basic commands
  ping: handlePing,
  state: handleState,
  states: handleStates,
  'states-json': handleStatesJson,
  'states-filter': handleStatesFilter,
  config: handleConfig,
  services: handleServices,
  call: handleCall,
  template: handleTemplate,

  // History commands
  logbook: handleLogbook,
  history: handleHistory,
  'history-full': handleHistoryFull,
  attrs: handleAttrs,
  timeline: handleTimeline,
  syslog: handleSyslog,
  stats: handleStats,

  // Registry commands
  entities: handleEntities,
  devices: handleDevices,
  areas: handleAreas,

  // Automation commands
  traces: handleTraces,
  trace: handleTrace,
  'trace-vars': handleTraceVars,
  'automation-config': handleAutomationConfig,
  context: handleContext,
  'blueprint-inputs': handleBlueprintInputs,
  watch: handleWatch,
};
