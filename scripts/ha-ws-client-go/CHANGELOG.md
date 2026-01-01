# Changelog

All notable changes to ha-ws-client-go will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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
