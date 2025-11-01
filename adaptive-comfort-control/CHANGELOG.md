## [4.8.1] - 2025-10-31
### Fixed
- Map adaptive fan control to actual supported modes. When a thermostat only exposes `on`/`auto`, the blueprint now chooses between those and never attempts `high`/`medium`/`low`.
- Prevent the inside-band **low** setpoint from collapsing to the device minimum by clamping it to the adaptive band floor (addresses night/sleep cases).
- Define occupancy/setback (`occ_now` / `setback_apply_c`) before sleep logic to eliminate transient undefined/boolean-cast errors.

### Notes
- No behavior intent change; these fixes improve robustness across diverse climate entities and avoid invalid-service errors.

## [4.8.0] - 2025-10-31
### Added
- **Sleep / Circadian Preferences:** optional nighttime cooling and comfort bias.
  - Configurable enable flag, start/end time, or external sleep mode sensor.
  - Negative (cooler) bias applied to the adaptive comfort target during sleep.
  - Optional “band tightening” for steadier overnight temperatures.
  - Full unit awareness (°C/°F) and integrated with setback, occupancy, and seasonal logic.
- Debug summary now reports `Sleep=<bool>` with effective bias and tightening delta.

### Improved
- Adaptive model incorporates circadian context for more realistic comfort behavior.
- Seamless interaction with occupancy and energy-saving logic—sleep bias automatically disables if unoccupied.
- Fully compatible with CO₂ and RMOT bias layers, including cumulative effects.

### Notes
- The sleep mode can be activated via time window or an entity (e.g., `input_boolean.bedtime`).
- No manual configuration changes are required for existing users; defaults preserve prior behavior.

## [4.7.1] - 2025-10-31
### Fixed
- Variable order dependency for occupancy-aware setback (`occ_now` / `setback_apply_c`) that could cause undefined or stale values at runtime.
- Hardened setpoint output path to use climate-unit conversion and thermostat min/max clamping to prevent out-of-range errors.

### Notes
- No behavior change to the adaptive model itself; fix ensures reliable evaluation order and safer setpoint writes across thermostats.

## [4.7.0] - 2025-10-31
### Added
- Optional barometric pressure sensor input with automatic unit conversion (kPa, hPa/mbar, Pa, inHg, mmHg).
- Optional site elevation input and automatic fallback to state-based LUT when regional presets are enabled.
- Dynamic pressure calculation now used for humidity ratio and psychrometric equations.
- Debug summary now includes resolved pressure (kPa) for verification.

### Improved
- Psychrometric accuracy at altitude (especially in CO/UT/WY) for dewpoint, AH, and enthalpy-based ventilation guards.
- Blueprint now self-corrects pressure assumptions for high-elevation sites.

### Fixed
- Removed previous hard-coded sea-level pressure (101.325 kPa) for improved realism and consistency.