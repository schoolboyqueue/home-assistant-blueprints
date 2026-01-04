# Changelog

All notable changes to ha-ws-client-go will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.5.4] - 2026-01-03

### Fixed

- Add thread-safety to global output configuration with mutex locks to prevent race conditions in concurrent access

### Added

- Add unit tests for handlers with mock WebSocket client

## [1.5.3] - 2026-01-03

### Changed

- Extract test fixtures to shared `scripts/testfixtures` package for code reuse across Go tools

## [1.5.2] - 2026-01-03

### Fixed

- Trace commands now accept entity_id format (e.g., `automation.guest_bedroom_adaptive_shade`) in addition to internal numeric IDs
  - Added `resolveAutomationInternalID` helper to automatically look up the automation's internal ID from its state attributes
  - Fixed `traces`, `trace`, `trace-latest`, `trace-summary`, `trace-debug`, `automation-config`, and `getTraceDetail` to resolve entity names

## [1.5.1] - 2026-01-03

### Added

- Centralized error handling system with new `internal/errors` package
  - `Error` type with typed error categories (Network, Validation, Parsing, Timeout, NotFound, Auth, API, Internal, Canceled, Subscription)
  - Error registry pattern with predefined error codes and factory functions
  - Helper functions for error type checking (`IsNotFound`, `IsValidation`, etc.)
  - Common error factories (`ErrEntityNotFound`, `ErrMissingArgument`, `ErrInvalidJSON`, etc.)

### Changed

- Refactor all handlers to use centralized error types instead of ad-hoc error creation
- Update error handling in client, middleware, and message dispatcher to use typed errors

## [1.5.0] - 2026-01-03

### Added

- Message dispatcher abstractions for unified WebSocket request/response handling
  - `MessageRequest[T]` generic type for type-safe message requests with `Execute()` and `ExecuteAndOutput()` methods
  - `MessageDispatcher[T]` fluent API with `Transform()` and `Output()` pipeline stages
  - `ListRequest[T]` for simplified list-based command handlers
  - `TimelineRequest[T]` for timeline display handlers
  - `MapRequest[T]` for map-based response extraction
  - `SimpleHandler` and `TransformHandler` factory functions for common patterns
- Unit tests for all new dispatcher abstractions

### Changed

- Refactor `handleTrace` and `handleTraceDebug` to use `MessageRequest` pattern
- Refactor `HandleStatesJSON` and `HandleConfig` to use `MessageRequest` pattern
- Refactor `HandleSyslog` to use `ListRequest` pattern
- Refactor `HandleAreas` to use `ListRequest` pattern
- Rename `errors` variable to `errs` in shutdown tests (lint fix)

## [1.4.0] - 2026-01-03

### Added

- Graceful shutdown coordination with new `internal/shutdown` package
  - `Coordinator` type for managing shutdown lifecycle with configurable grace periods
  - Signal handling for SIGINT/SIGTERM with proper cleanup sequencing
  - `PartialResult` tracker for reporting progress on interruption
  - `WrapContext` helper for combining parent context with shutdown coordination
- Context propagation throughout the application for cancellation support
  - `NewWithContext` constructor for creating context-aware WebSocket clients
  - `Ctx` field in handler context for graceful operation cancellation
- Exit code 130 for SIGINT interruption (standard Unix convention)

### Changed

- Refactor main.go to use shutdown coordinator for signal handling
- Update client creation to pass context for cancellation awareness
- Improve handler wrapper to propagate context through command execution

## [1.3.1] - 2026-01-02

### Fixed

- Fix `state --compact` output displaying raw Go struct instead of formatted `entity_id=state` format
  - The issue occurred because typed structs (like HAState) weren't being converted to maps before compact formatting
  - Added `structToMap()` helper that uses JSON marshaling to convert any struct to `map[string]any`
- Fix `traces --json` output producing multiple JSON lines instead of a single JSON object when no traces exist
  - When an automation has `last_triggered` but no stored traces, the info is now returned as a single JSON object with `entity_id`, `traces`, `last_triggered`, and `message` fields

### Added

- Add unit tests for struct-to-map conversion and compact output formatting
  - `TestPrintCompact_Struct` - verifies typed structs output correctly in compact mode
  - `TestStructToMap` - tests JSON-based struct conversion edge cases
  - `TestPrintCompactMap` - tests entity state, trace info, and generic map formats
  - `TestData_CompactFormat_WithStruct` - integration test for Data() with struct input

## [1.3.0] - 2026-01-02

### Added

- Add `context` command support for entity_id input (previously only accepted context_id)
  - Now accepts either `context <entity_id>` or `context <context_id>`
  - Shows related state changes that share the same context chain
  - Displays helpful message when no matches found with suggestions
- Add `traces` command discrepancy detection when automation has `last_triggered` but no stored traces
  - Shows the last_triggered timestamp and suggests checking trace storage settings
- Add unit tests for context helper functions (`findEntityByID`, `findStatesByContext`, `addParentContextMatches`, `formatContextInfo`)
- Add integration tests for `context`, `traces`, and `automation-config` handlers
- Add test automation with `stored_traces: 0` for testing traces discrepancy feature

### Changed

- Improve `automation-config` command to retrieve full config from traces when available
  - For blueprint automations, tries to get resolved config from most recent trace
  - Falls back to API when no traces available with helpful message
- Refactor `handleContext` to reduce cyclomatic complexity using helper functions

## [1.2.0] - 2026-01-03

### Added

- Add concurrent batch processing framework (`batch.go`) for parallel operations
- Add registry handler tests (`registry_test.go`)
- Add time parsing tests (`time_test.go`)

### Changed

- Refactor CLI flag parsing by removing deprecated flags module
- Enhance history and monitor handlers with improved output formatting
- Update urfave/cli dependency

### Fixed

- Fix unused parameter lint warning in wrapHandler
- Fix unnecessary int conversion lint warning

## [1.1.2] - 2026-01-02

### Added

- Add integration tests for all handler commands with real Home Assistant connections
- Add remote testing support via `HA_HOST`, `HA_WS_URL`, and `HA_TOKEN` environment variables
- Add comprehensive Testing section to README with setup instructions

### Changed

- Refactor integration test client to support both local add-on and remote connections

## [1.1.1] - 2026-01-02

### Fixed

- Fix `template` command timing out for non-string results (integers, floats)
  - HA returns numeric template results (e.g., `{{ 1 + 1 }}` → `2`) as their native types
  - The subscription callback was only accepting string types, causing silent failures
  - Now properly converts any result type to string using `fmt.Sprintf("%v", result)`

## [1.1.0] - 2026-01-02

### Changed

- Refactor handlers to use composable middleware pattern for argument validation
- Extract CLI utilities (flag parsing, time parsing) into internal/cli package
- Add HandlerConfig struct to Context for passing validated data between middleware

## [1.0.5] - 2026-01-02

### Fixed

- Fix `history-full --compact` output formatting (was printing raw Go struct format)

## [1.0.4] - 2026-01-02

### Added

- Add `trace-latest <automation_id>` command to get most recent trace without listing first
- Add `trace-summary <automation_id>` command for quick overview of recent automation runs
- Add `device-health <entity_id>` command to check device responsiveness and stale detection
- Add `compare <entity1> <entity2>` command for side-by-side entity comparison with attribute diffs
- Add `--show-age` flag to `states-filter` to display last_updated age with stale indicators
- Add `--from` time filtering support to `traces` command

## [1.0.3] - 2026-01-02

### Changed

- Change trace command argument order from `<run_id> [automation_id]` to `<automation_id> <run_id>` for better UX
- Update `blueprint-inputs` message to clarify HA API limitation for blueprint automations

## [1.0.2] - 2026-01-01

### Fixed

- Fix `HistoryState.LU/LC` type from `int64` to `float64` to handle decimal Unix timestamps
- Fix `SysLogEntry.Source` type from `[]string` to `any` with `GetSource()` helper method
- Fix `SubscribeToTemplate` race condition by registering event handler before sending message
- Fix `getHistory` in monitor.go to preserve float64 timestamp precision
- Fix event handling for `render_template` which uses `Event.Result` instead of `Event.Variables`
- Fix `StatEntry.GetStartTime()` to handle millisecond timestamps from statistics API
- Fix `TraceDetail.Trigger` type from `*TraceTrigger` to `any` with helper methods for variable response types

## [1.0.1] - 2026-01-01

### Changed

- Remove duplicate `TimeRange` and `HistoryState` types from monitor.go (now use shared types from types.go)
- Replace custom `join()` function with `strings.Join` from stdlib
- Remove unused `CommandContext` type

### Fixed

- Add auto-cleanup for template subscriptions when timeout is specified
- Improve timeline command error message formatting
- Align SysLogEntry struct field tags for consistency

## [1.0.0] - 2025-01-01

### Added

- Initial release of ha-ws-client-go
- WebSocket client for Home Assistant API
- Basic commands: ping, state, states, states-json, states-filter, config, services, call, template
- Log commands: logbook, history, history-full, attrs, timeline, syslog, stats, context, watch
- Registry commands: entities, devices, areas
- Automation debugging: traces, trace, trace-vars, trace-timeline, trace-trigger, trace-actions, trace-debug, automation-config, blueprint-inputs
- Monitoring commands: monitor, monitor-multi, analyze
- Time filtering with --from and --to options
- Output format options: --json, --compact, --no-headers, --no-timestamps, --max-items
- Cross-platform builds for Linux (amd64, arm64, armv7, armv6), macOS (amd64, arm64), Windows (amd64)
- Version information with --version flag
