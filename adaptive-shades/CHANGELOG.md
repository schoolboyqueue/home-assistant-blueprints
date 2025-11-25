# Adaptive Shades Pro - Changelog

## [1.5.0] - 2025-11-25

### Added

- Optional climate entity to bias heating/cooling classification and reflect HVAC state in shading decisions.
- Room profile (living/office/bedroom) that adjusts glare sensitivity; effective glare threshold derived from profile.

### Changed

- Comfort mode detection now prefers active HVAC state when provided, falling back to setpoints.
- Glare detection uses the profile-adjusted threshold; documentation updated for clarity.

## [1.5.3] - 2025-11-25

### Fixed

- Replaced unsupported Jinja `exp` filter in clear-sky irradiance calculation with explicit exponent math (uses `e_const ** exponent`).

## [1.5.6] - 2025-11-25

### Fixed

- Adjusted debug level selector to use simple string options for Home Assistant compatibility and aligned version metadata.

## [1.5.4] - 2025-11-25

### Added

- Diagnostics input with `debug_level` (off/basic/verbose) and gated logbook logging, including verbose no-movement traces for troubleshooting.

## [1.5.2] - 2025-11-25

### Fixed

- Replaced unsupported Jinja `radians` filter with explicit degree-to-radian conversion for tilt and sun geometry math to prevent template parse errors.

## [1.5.1] - 2025-11-25

### Fixed

- Corrected tilt capability detection template to avoid template parse errors in Home Assistant (bitwise detection now uses `bitwise_and`).

## [1.4.0] - 2025-11-25

### Added

- Optional `sun_entity` and `weather_entity` to drive sun/irradiance logic and suppress direct-sun detection on overcast conditions when no solarimeter is provided.
- Manual adjustment timeout input that pauses automation after manual shade movement for a configurable duration.

### Changed

- Comfort mode detection now derives from heating/cooling setpoints (with comfort margin) instead of fixed 21/25°C thresholds.
- Shading mode description clarified; irradiance and slat-geometry sections marked as advanced/leave-defaults to reduce setup friction.
## [1.3.1] - 2025-11-25

### Added

- Sunset/sunrise night mode: closes to block position when the sun is below the horizon and resumes adaptive control after sunrise, honoring manual override and quiet hours.

## [1.3.0] - 2025-11-25

### Added

- Zebra shading mode with calibrated admit/dim/block positions and logic driven by sun-on-window, direct-sun detection, temperature band, presence, and glare.
- New `dim_position_pct` input for zebra filtered-light position; shading mode selector now supports zebra.

## [1.2.0] - 2025-11-25

### Added

- Shading mode selector (slat vs zebra placeholder) and tilt-aware control: when the cover supports tilt, send `set_cover_tilt_position`; otherwise use position.
- Wrapped slat logic in a dedicated target to prep for zebra behavior in a later pass.

## [1.1.1] - 2025-11-25

### Changed

- Auto-compute minimum sun elevation from Home Assistant’s `sun.sun` (uses 1° when above horizon, otherwise disables sun-on-window) so users no longer configure this threshold manually.

## [1.1.0] - 2025-11-25

### Added

- Implemented the full venetian-blind control strategy from Energies 13(7):1731: clear-sky irradiance comparison with vertical solar sensor, direct-vs-diffuse branching, temperature bands (winter/intermediate/summer), and occupied vs unoccupied logic.
- Added Eq. 8 glare-limiting slat angle with configurable slat gap ratio, plus max tilt cap for safer mapping to cover position.
- New inputs for ASHRAE clear-sky coefficients, direct sun ratio, diffuse threshold (300 W/m² default), and optional irradiance sensor.

## [1.0.0] - 2025-11-25

### Added

- Initial release of Adaptive Shades Pro based on solar geometry (Energies 13(7):1731) with minimal setup (shade, orientation, comfort bounds).
- Comfort-aware logic that biases open for heating (solar gain) and closed for cooling, using optional indoor/outdoor temperature sensors and hysteresis.
- Glare protection via optional indoor lux threshold, plus presence, manual override, and quiet-hours pauses.
- Import badge and README for quick setup guidance.

---
