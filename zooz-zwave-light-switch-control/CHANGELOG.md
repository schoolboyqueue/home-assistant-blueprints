# Changelog

## 1.0.0 - 2025-01-09
### Added
- Initial stable release of Zooz Z-Wave Light Switch Control Pro
- Central Scene single press on/off (up=on, down=off)
- Hold-to-dim with immediate release detection
- Min-threshold auto-off when dimming reaches bottom
- Area targeting support (target area instead of single entity)
- Debug levels: off, basic, verbose
- Support for both zwave_js_event and zwave_js_value_notification

### Fixed
- YAML validation errors with device selector and debug level options
- Malformed choose/default blocks replaced with if/then/else for binary conditions
- Light off check in up_hold sequence
- All area_set condition blocks converted from choose to if/then/else
- Brightness threshold check converted from choose to if/then/else

### Removed
- Optional gesture action inputs (double/triple tap)
- Triggers still fire for these events - use separate automations if needed
