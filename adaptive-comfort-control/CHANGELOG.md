## [4.11] — 2025-11-02
### Added
- **Thermostat vendor profiles + auto-detect:** Infers profile from device manufacturer/model/name; enforces vendor minimum Auto/Heat-Cool separation (e.g., Ecobee = **5°F**). Optional user override preserved.
- **Safety & Guards:** Absolute floor/ceiling in system units (freeze/overheat) with optional block on HVAC pause when at risk.
- **Pause acceleration near risk:** Shortens resume (and optionally open) delays when indoor temp approaches guard thresholds; bounded by user-set minimums to avoid chatter.
- **Regional presets:** State/region based seasonal/psychrometric defaults with intensity scaling.
- **Verbose diagnostics:** Rich single-line logbook entries (no panel notifications) including band, setpoint, IN/OUT, device step, bounds, and separation enforcement status.

### Changed
- **Quantization strategy:** Floor to device `target_temp_step` before separation checks; clamp all targets/bands to device allowed range.
- **Band & target guards:** Always compute in °C internally; convert late to climate and UI units to avoid unit drift; keep ordering sane without a second pass.
- **Natural ventilation gate:** Consolidated dewpoint/AH/enthalpy checks behind one condition; respects regional and psychrometric presets.

### Fixed
- **“Provided temperature … not valid”**: Prevented overflow to `33.333…` by clamping and quantizing before service calls.
- **Trigger/action tail restoration:** Reinstated full `trigger:` set (state sensors, time patterns, HA start) and the main action graph after earlier truncation.
- **Undefined template vars & duplicate keys:** Resolved lingering lint errors (e.g., effective pause delays, unique map keys).
- **Ecobee auto bounds too tight:** Enforce minimum separation explicitly even when device reports ambiguous diffs; avoids 74–75°F “auto” gaps.

## [4.9.0] - 2025-11-01
### Fixed
- Corrected seasonal bias preset conversion from °C→°F (removed +32 offset on delta values).
- Eliminated extreme target values (e.g., 91°F) caused by unguarded adaptive target exceeding comfort bounds.
- Added two-sided clamping to `band_min_c` / `band_max_c` to prevent inversion and ensure stable adaptive band formation.
- Guarded adaptive target (`t_adapt_c_guard`) introduced to maintain band integrity within user comfort range.

### Changed
- Overhauled HVAC pause/resume logic:
  - Added trigger-based debounce (no pause on short door opens).
  - Removed redundant action delays to align with trigger `for:` timing.
  - Always resumes HVAC to a valid mode when doors are closed.
- Aligned all temperature outputs to use the guarded adaptive target for consistent reporting.
- Expanded diagnostic logs to include all bias terms, tolerance, and adaptive intermediates.

### Notes
- Normal daytime operation now stabilizes around 70–72°F with outdoor ≈ 29°F.
- Comfort band: ~65–76°F (configurable via inputs).
- Backward compatible; strongly recommended for all v4.8.x users to eliminate 91°F runaway conditions.

## [4.8.6] - 2025-11-01
### Fixed
- Eliminated floating-point spillover that produced invalid setpoints like `33.333333…` against devices capped at `33.3°C`.
- Quantized all computed setpoints using **floor** rounding (never rounding up past device max) and clamped to a **safe maximum** (`max_allowed_cli − 0.001`).
- Applied safe-max clamp to all `climate.set_temperature` calls (single and low/high) and standardized rounding to **one decimal place**.

### Changed
- Added `cli_eps` and `cli_max_safe` helper variables for precision control.
- Reworked the “step rounding to device resolution + final clamp” block to use floor quantization and safe-limit clamping.
- Unified rounding behavior across temperature writes for consistent °C/°F device support.

### Notes
- Backward compatible with previous versions.
- Behavior near the upper temperature limit may shift downward by ≤ one device step, which is intentional to prevent HA range errors.
- Resolves recurring “Provided temperature … not valid. Accepted range is 7.2 to 33.3” log errors.

## [4.8.5] - 2025-10-31
### Fixed
- Replaced `wait_for_trigger` in HVAC Pause and Resume branches with simple **delay and recheck** logic to prevent deadlocks when already in the target state (e.g., door already closed when resume runs).
- Removed strict numeric sanity check from global conditions that blocked pause/resume execution; replaced with a non-blocking **sanity probe log** at the start of `action:`.
- Pause/Resume delays now reliably respect user-configured **Pause after open** and **Resume after close** timings across all integrations (`on/off`, `open/closed`, etc.).

### Improved
- Enhanced log visibility: preflight trigger ID, open-sensor lists in pause logs, and a `t_in_c` / `t_out_c` / `band_min_c` / `band_max_c` sanity snapshot when debug is basic or verbose.
- Increased blueprint resilience to mixed door/window entity reporting conventions.

### Notes
- No change to adaptive comfort, sleep bias, or ventilation behavior.
- Recommended update for anyone using door/window-based HVAC pause.

## [4.8.4] - 2025-10-31
### Fixed
- Eliminated boundary errors (e.g., `33.333333333333336 > 33.3`) by adding
  step-rounded **final** setpoints and **service-level clamps** on every
  `climate.set_temperature` write.
- Resolved `supports_heat_cool` undefined warnings by moving HVAC capability
  variables earlier in the variables block (before any references).
- Ensured inside-band Heat/Cool writes use `inside_low/high` (final, clamped)
  for stable sleep-bias behavior and predictable Auto mode operation.
- Final debug log now surfaces device step and bounds for quick validation.

### Notes
- No behavioral change if your device already enforced limits; this update
  simply guarantees payloads are always within the device’s accepted range.
- Blueprint field `blueprint_version` should be bumped to `4.8.4`.

## [4.8.2] - 2025-10-31
### Fixed
- Resolved issue where the low Auto-mode setpoint could not drop below ~69°F when no minimum separation is required.
- Corrected separation enforcement to be **optional and device-aware** rather than hard-coded.
- Adjusted inside-band and off-band control logic to write the **inside_low/high** targets instead of band edges for more accurate sleep bias handling.

### Added
- New input: **Auto Mode Min Separation (system units)**, default 0.0.
  - Allows users to explicitly set or disable separation in Heat/Cool (Auto) mode.
  - Automatically detects device-advertised minimums (`min_temp_diff`, `temperature_difference`) when present.
- Extended debug output to include `sep_enforce` and `sep_min` for verification.

### Notes
- When `Auto Mode Min Separation` = 0.0 and the device does not report a required gap, the comfort band is free to fully contract—allowing lower nighttime bias.
- Existing behavior for devices that require a gap remains unchanged.

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