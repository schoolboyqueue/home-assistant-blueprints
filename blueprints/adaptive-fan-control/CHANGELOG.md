# Ceiling Fan Climate Control Pro - Changelog

## [2.2.6] - 2025-12-31

### Fixed

- **Absolute bounds checking**: Added validation to reject temperatures outside 40°F-120°F (4°C-49°C) range. This catches bogus readings like 212°F when a sensor transitions from unavailable to a value, bypassing the rate-of-change filter which requires a valid previous reading. Invalid readings are logged with details and ignored.

## [2.2.5] - 2025-12-30

### Fixed

- **Sensor glitch protection**: Added rate-of-change filter to reject impossible temperature spikes. When triggered by a temperature change, rejects readings that jump more than 10°F from the previous value (e.g., 72°F → 212°F). This catches sensor glitches like malfunctioning Ecobee sensors reporting boiling point before going offline, without blocking legitimate readings from saunas, sunrooms, or other extreme-but-real environments. Also rejects unavailable/unknown states. Invalid readings are logged with details and ignored.

## [2.2.4] - 2025-12-27

### Fixed

- **Jinja2 boolean output**: Replaced bare `true`/`false` with `{{ true }}`/`{{ false }}` in `is_imperial` variable to output actual booleans instead of strings.

## [2.2.3] - 2025-12-26

### Fixed

- **Variable ordering error**: Moved `calculated_speed` and `calculated_direction` definitions before variables that reference them, fixing "'calculated_direction' is undefined" template warnings in Home Assistant logs.

## [2.2.2] - 2025-12-25

### Fixed

- **Redundant RF commands eliminated**: Added early-exit optimization that skips the entire automation when the fan is already in the desired state. Prevents audible command sounds on RF devices (Bond, etc.) when temperature sensors update but no fan change is needed.

## [2.2.1] - 2025-12-24

### Fixed

- **Fan direction comparison**: Normalized `fan_current_direction` to lowercase to handle fans that report "Forward" vs "forward" inconsistently
- **Speed tolerance for redundant commands**: Added 5% tolerance when comparing current vs target speed to avoid sending unnecessary fan commands for negligible differences. New variables `speed_tolerance`, `speed_matches`, and `heating_speed_matches` prevent wear on fan motor from constant micro-adjustments

## [2.2.0] - 2025-12-21

### Fixed

- **Prevent redundant fan commands during HVAC mode changes**: Fan speed and direction are now only updated when they actually differ from current state, reducing unnecessary commands when HVAC transitions between heating/cooling/idle states

## [2.1.0] - 2025-12-21

### Changed

- **Adaptive mode now uses pure EN 16798 comfort calculation**: Fixed thresholds no longer cap the adaptive comfort band. In adaptive mode, the comfort band is determined entirely by outdoor temperature and comfort category (e.g., Category III at 6°C outdoor → 64.6°F - 79.0°F band). Fixed thresholds only apply in "Fixed thresholds" mode.

### Fixed

- Fan no longer turns on unnecessarily in winter when indoor temp is within adaptive comfort band (e.g., 78°F at low humidity with cold outdoor temps)

## [2.0.3] - 2025-12-21

### Fixed

- Switched debug logging from `system_log.write` to `logbook.log` for better visibility in Home Assistant's Logbook UI

## [2.0.2] - 2025-12-21

### Fixed

- Fixed debug logging not working - now uses direct input reference in templates instead of pre-computed boolean variable (matches pattern used in other blueprints)

## [2.0.1] - 2025-12-21

### Fixed

- Debug logging now fires on both "basic" and "verbose" levels (was only verbose)
- Added default case logging when no action is taken
- Improved log message format with conditional humidity/outdoor temp display

## [2.0.0] - 2025-12-21

### Added

- **Adaptive comfort mode (EN 16798 / ASHRAE 55):** Dynamic comfort band that shifts based on outdoor temperature. Warmer outdoor temps = higher acceptable indoor temps, matching human thermal adaptation.
- **Indoor humidity sensor:** Optional input that enables heat index (feels-like temperature) calculation. Humidity is factored into comfort decisions when provided.
- **Outdoor temperature sensor:** Optional input that enables adaptive comfort mode. Supports both weather entities and dedicated temperature sensors.
- **Comfort category selection:** Choose between Category I (±2°C strict), II (±3°C normal), or III (±4°C relaxed) comfort bands.
- **Temperature unit selection:** Auto-detect from sensor, or manually specify Fahrenheit or Celsius.
- **Deviation-based speed calculation:** Fan speed based on how far above the comfort band the temperature is, not fixed thresholds.
- **Periodic trigger (5 min):** Catches gradual temperature changes that might not trigger state changes.

### Changed

- **Comfort logic completely rewritten:** Now uses adaptive comfort model instead of fixed temperature thresholds.
- **Speed tier thresholds:** Now represent degrees of deviation from comfort band (default 2°/4°) rather than absolute temperatures.
- **Fixed thresholds repurposed:** Now serve as absolute limits that cap the adaptive comfort band.
- **Trigger system updated:** Removed numeric_state triggers for fixed thresholds; now triggers on any temperature state change plus periodic evaluation.
- **Log messages updated:** Now show comfort band, heat index, and deviation values.
- **Source URL updated:** Points to correct blueprint path.

### Fixed

- Fan no longer turns on when temperature is high but humidity is low (e.g., 77°F at 30% RH in winter).

## [1.0.0] - 2025-12-21

### Added

- Initial release of Ceiling Fan Climate Control Pro with HVAC-aware ceiling fan automation.
- **HVAC integration:** Monitors thermostat hvac_action to coordinate fan behavior with heating/cooling cycles.
- **Temperature-based speed control:** Three-tier speed system (low/medium/high) with configurable thresholds and percentages.
- **Occupancy awareness:** Fan only runs when room is occupied, with configurable delay before turning off.
- **Direction control:** Optional support for fans with reverse mode, including automatic reverse during heating and seasonal direction changes.
- **Heating mode options:** Choose between turning fan off during heating (default) or running in reverse at low speed to circulate warm air.
- **Seasonal adaptation:** Automatic direction adjustment based on Home Assistant's season sensor when HVAC is idle.
- **Debug logging:** Three-level diagnostics (off/basic/verbose) for troubleshooting.
- Import badge and README for quick setup guidance.

---
