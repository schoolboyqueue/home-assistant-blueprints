# Ceiling Fan Climate Control Pro - Changelog

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
