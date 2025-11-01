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