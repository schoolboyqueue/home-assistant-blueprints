# Changelog

All notable changes to ha-ws-client-go will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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
