# Changelog

All notable changes to ha-ws-client-go will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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
