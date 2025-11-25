# Adaptive Shades Pro - Changelog

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
