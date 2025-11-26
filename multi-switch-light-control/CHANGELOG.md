# Changelog

## 1.6.0 - 2025-11-26
### Changed
- Single tap up now always sets brightness to 100% instead of restoring previous brightness

## 1.5.9 - 2025-11-27
### Added
- Optional state-sync toggle that mirrors the selected switch/light entity with the target lights (any light >0% → switch on, all lights off → switch off). Includes auto-detection of a suitable entity with an override input.
### Changed
- Automation now runs in `parallel` mode so light state sync triggers don't interrupt button-driven hold loops.

## 1.5.8 - 2025-11-27
### Fixed
- Match Lutron Pico button presses by `button_type` (on/raise/stop/lower/off) instead of vendor-specific button numbers so every remote reports the correct action mappings regardless of serial numbering.
- Update raise/lower release listeners to the same `button_type` filters, restoring hold-to-dim behavior for Pico remotes.

## 1.5.7 - 2025-11-27
### Fixed
- Clean up Zigbee action sensor detection so empty placeholders produce true empty strings (not quoted `""`), keeping Pico remotes in the Lutron branch.

## 1.5.6 - 2025-11-27
### Changed
- Revert Pico triggers to filtered `lutron_caseta_button_event` listeners (device_id + button number) so the correct branch executes without requiring Lutron config entries.

## 1.5.5 - 2025-11-27
### Fixed
- Guard Lutron Pico event filters with defaults so missing `serial_number` attributes no longer raise template warnings when upgrading from older blueprint versions.

## 1.5.4 - 2025-11-27
### Fixed
- Add verbose trigger/event logging to help diagnose Lutron Pico inputs and confirm trigger IDs/platforms at runtime.

## 1.5.3 - 2025-11-27
### Fixed
- Use Lutron Caseta button events with explicit button numbers (2-6) instead of device triggers so Pico automations validate and log actions correctly across integrations.

## 1.5.2 - 2025-11-27
### Fixed
- Switch Lutron Pico triggers back to event-based listeners so non-Lutron devices no longer error with "Device has no config entry from domain 'lutron_caseta'" during blueprint validation/import.

## 1.5.1 - 2025-11-27
### Fixed
- Restore device-based Lutron Pico press/release triggers so Pico remotes fire button actions again.

## 1.5.0 - 2025-11-27
### Added
- Add Zigbee2MQTT device action triggers so button events run without relying on the action sensor entity, including hold/release detection inside the dimming loops.

### Fixed
- Auto-select the detected Zigbee action sensor when available and keep `entity_id` references static, logging a hint if we discover a candidate but the input remains at the placeholder.

## 1.4.4 - 2025-11-27
### Fixed
- Require the Zigbee action sensor input to reference a concrete entity and remove templated `entity_id` fields, preventing "Entity is neither a valid entity ID" errors when the action sensor is unconfigured. Added a placeholder default plus debug hints when an auto-detected sensor exists but is not selected.

## 1.4.3 - 2025-11-27
### Fixed
- Improve Zigbee action sensor auto-detection to only select `sensor.*_action` entities and prefer matches containing the switch name, prompting users to pick an input when multiple candidates remain.

## 1.4.2 - 2025-11-27
### Fixed
- Sanitize Zigbee action sensor triggers/wait loops so empty inputs no longer create invalid entity IDs, preventing "Entity is neither a valid entity ID nor a valid UUID" import errors when no Zigbee action sensor is configured.

## 1.4.1 - 2025-11-26
### Fixed
- Execute light service calls per entity when an area target is not provided, preventing "Entity is neither a valid entity ID nor a valid UUID" errors when multiple lights are selected.

## 1.4.0 - 2025-11-26
### Added
- Zigbee2MQTT and ZHA support for Inovelli Zigbee switches
- Optional `zigbee_action_sensor` input for Zigbee switches (auto-detected if left blank, manual selection available)
- State triggers for all Zigbee button actions: `up_single`, `down_single`, `up_held`, `down_held`, `up_release`, `down_release`, `up_double`, `down_double`, `up_triple`, `down_triple`, `up_quadruple`, `down_quadruple`, `up_quintuple`, `down_quintuple`
- Auto-detect action sensor entities (e.g., `sensor.*_action`) from device registry with fallback to manual input
- Enhanced protocol detection in `switch_type` variable to identify Zigbee vs Z-Wave vs Lutron based on presence of action sensor entity
- Zigbee release detection in hold-to-dim loops using `up_release` and `down_release` state changes
- Updated all trigger ID checks to support both Z-Wave and Zigbee2MQTT trigger IDs (e.g., `up_single` and `up_single_z2m`)

### Changed
- Updated blueprint description to include Inovelli Zigbee switches (Zigbee2MQTT/ZHA) alongside Z-Wave and Lutron devices
- README updated with Zigbee support documentation, protocol detection details, and troubleshooting for Zigbee2MQTT

### Fixed
- Replace template triggers with state triggers to avoid "undefined variable" warnings in Home Assistant logs
- State triggers use `!input zigbee_action_sensor` directly since template triggers cannot access automation variables

## 1.3.3 - 2025-11-26
### Fixed
- Replace Lutron Caseta device triggers with event-based triggers to fix "Device has no config entry from domain 'lutron_caseta'" error when using Z-Wave or Inovelli switches
- Convert all Lutron triggers to use `lutron_caseta_button_event` platform events instead of device-specific triggers, allowing the blueprint to work with any device type without requiring Lutron integration

## 1.3.2 - 2025-11-26
### Fixed
- Fix malformed YAML indentation in down_triple debug logging that caused blueprint import error ("template value is None for dictionary value @ data['actions'][13]['then'][0]['else'][0]['then'][0]['data']")

## 1.3.1 - 2025-11-26
### Fixed
- Replace `numeric_state` conditions with `template` conditions in Lutron raise/lower repeat loops to fix blueprint import error ("Entity {{ light_entity }} is neither a valid entity ID nor a valid UUID")

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
