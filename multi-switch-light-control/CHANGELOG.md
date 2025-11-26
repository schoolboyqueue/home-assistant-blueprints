# Changelog

## 1.3.0 - 2025-11-26
### Added
- Allow the light target input to select multiple entities, applying commands to every selected light while using the first one for brightness/state reads.

## 1.2.0 - 2025-11-24
### Added
- Custom Central Scene actions for up/down presses (1x through 5x) so each tap count can run a user-defined automation, including full logging and quadruple/quintuple event support.
## 1.1.1 - 2025-11-24
### Added
- Display the selected device's vendor/model/name below the switch selector so you can confirm the detected hardware before saving.
### Changed
- Generalized the hold-step delay and on/off transition inputs so the dimming controls apply to both Central Scene switches and Lutron Pico remotes, keeping the fade timings consistent.

## 1.1.0 - 2025-11-24
### Added
- Rename the blueprint to **Multi Switch Light Control Pro** and allow a single automation to target Zooz/Z-Wave, Inovelli, or Lutron Pico hardware.
- Auto-detect the selected device's manufacturer/model/type and log it when diagnostics are enabled so you can verify the detection.
- Add dedicated Lutron Pico tuning inputs (favorite button defaults, transition speeds, hold step delay) and a favorite button override action.
- Include Lutron press/release triggers alongside the existing Z-Wave Central Scene events so each device type uses the proper button grammar.
### Changed
- Reworked the description, README, and folder/file names to describe the new multi-switch experience.
- Keep the existing Zooz/Inovelli hold-to-dim logic while offering the same configurable dimming inputs to Lutrons.

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
