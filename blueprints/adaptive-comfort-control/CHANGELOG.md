## [4.22.0] — 2025-12-29

### Added

- **Time-of-day aware learning**: Learning now stores 4 separate offsets (`heat_day`, `heat_night`, `cool_day`, `cool_night`) instead of a single global offset. The system learns your preferences separately for daytime vs nighttime (based on sleep schedule) and heating vs cooling adjustments.
- **Trigger-based template sensor storage**: Replaced `input_number` helper with a trigger-based template sensor for storing learned preferences. This provides persistent JSON storage that survives restarts and can hold multiple values in a single entity. See updated `LEARNING_SETUP.md` for setup instructions.
- **Correct heat/cool learning direction**: Fixed learning to correctly compare manual adjustments against the appropriate setpoint. When you adjust `target_temp_low` (heating), it now compares against `target_low_cli`; when you adjust `target_temp_high` (cooling), it compares against `target_high_cli`. Previously, both were compared against an average, causing incorrect learning direction.

### Fixed

- **Heating setpoint always at minimum comfort**: Fixed `t_adapt_c_guard` clamping that was forcing the adaptive target to stay above `min_c + tol_c`, which caused the heating setpoint (`band_min_c`) to always equal the minimum comfort temperature. The guard now allows the full comfort range, so heating setpoints can properly vary based on the adaptive calculation and learned preferences.

### Changed

- **Input renamed**: `learned_offset_helper` (input_number) replaced with `learned_prefs_sensor` (sensor entity) to support the new multi-value learning system.

## [4.21.2] — 2025-12-28

### Fixed

- **Manual override helper condition failing**: Fixed entity ID variable truthiness checks that were causing the override helper update condition to evaluate as `false` even when all inputs were correctly configured. Replaced unreliable `and entity_var` patterns with explicit `(entity_var | string | length > 0)` checks for consistent boolean evaluation.

## [4.21.1] — 2025-12-28

### Fixed

- **Manual override helper not updating**: Fixed timestamp arithmetic in the override expiry calculation. The variables `_now_ts` and `_duration_sec` were being treated as strings, causing string concatenation instead of numeric addition. Now computes the timestamp correctly within a single Jinja2 block with proper numeric types.
- **Added debug logging for override checks**: When debug level is `verbose`, now logs the override helper state, timestamps, and computed `is_override_active` value on each automation run to help diagnose override persistence issues.

## [4.21.0] — 2025-12-28

### Added

- **Deadband hysteresis for setpoint updates**: The previously reserved `heat_deadband` and `cool_deadband` inputs are now fully implemented. Setpoint commands are skipped when current temperature is within deadband of target, reducing unnecessary thermostat writes and preventing oscillation near setpoints.
- **Dynamic dual-mode support**: Added `preferred_dual_mode` variable that detects whether the thermostat supports `heat_cool` or `auto` mode, and `is_in_dual_mode` for checking current state. Mode-setting logic now dynamically uses the correct mode name instead of hardcoding `heat_cool`.
- **Band inversion prevention comments**: Added detailed inline documentation explaining the nested min/max logic that prevents comfort band inversion when adaptive target approaches band edges.

### Fixed

- **Natural ventilation override persistence**: Added `stop` action after natural ventilation activates to prevent subsequent mode-setting blocks from re-enabling HVAC. Previously, the HVAC could be turned back on in the same run after ventilation was preferred.
- **Psychrometric division safety**: Added `is number` guards to saturation pressure (`psat_in_kpa`, `psat_out_kpa`), vapor pressure (`pv_in_kpa`, `pv_out_kpa`), absolute humidity (`ah_in`, `ah_out`), humidity ratio (`w_in`, `w_out`), and enthalpy (`h_in`, `h_out`) calculations to prevent division/math errors when temperature or pressure inputs are unavailable or non-numeric.
- **Dew point log(0) protection**: Added `x > 0` guards before `log()` calls in dew point calculations to prevent math domain errors when vapor pressure is extremely low.
- **Time parsing robustness**: Improved sleep schedule time parsing to handle both `HH:MM` and `HH:MM:SS` formats with explicit length checks, preventing index errors on malformed time strings.
- **Boolean variable syntax**: Changed `is_imperial` and `climate_is_imperial` assignments to use explicit `{{ true }}`/`{{ false }}` syntax instead of bare literals for consistent Jinja2 boolean output.

### Changed

- **Simplified optional sensor parsing**: Standardized 9 optional sensor ID extraction blocks into a cleaner 3-line pattern, removing redundant `| default('', true)` calls and improving readability.
- **Removed dead code**: Removed the unused `_nat_vent_active` runtime variable assignment (natural ventilation state is now handled via `stop` action and pre-computed `nat_vent_would_activate`).

## [4.20.5] — 2025-12-28

### Fixed

- **Manual override detection reliability**: Completely rewrote the manual override detection logic. Instead of comparing the thermostat setpoint against the predicted adaptive target (which can drift due to changing conditions), now compares against the exact setpoints the automation WOULD set (`target_low_cli`/`target_high_cli` for dual mode, `t_adapt_cli_q` for single mode). This prevents the automation from detecting its own temperature changes as manual overrides, which was causing a feedback loop where learning would trigger on every automation update.

## [4.20.4] — 2025-12-28

### Changed

- **Season inference alignment**: Month-based fallback now outputs `spring`/`autumn` to match Home Assistant's `season` sensor output instead of the internal `shoulder` value. The shoulder bias is still applied to spring and autumn seasons.
- **Code cleanup**: Removed unused `heat_db_sys` and `cool_db_sys` variable bindings (inputs are preserved for future use). Removed unused `prev_cli` variables from manual override detection logic.

## [4.20.3] — 2025-12-28

### Fixed

- **HVAC mode detection**: Fixed `hvac_mode_now` to read from entity state (`states()`) instead of incorrectly reading from attribute (`state_attr()`). This was causing incorrect mode detection and potentially unnecessary mode changes.
- **Absolute humidity operator precedence**: Fixed `ah_in` and `ah_out` calculations where the `| round(2)` filter was only applying to `1000.0` instead of the entire expression due to operator precedence.
- **Enthalpy null safety**: Added guard for `w_in`/`w_out` being `none` in enthalpy (`h_in`/`h_out`) calculations to prevent math errors when humidity ratio cannot be computed.
- **Dew point log(0) protection**: Added guard against computing `log(0)` in dew point calculations when vapor pressure is extremely low (near 0% RH).
- **Natural ventilation variable scope**: Fixed `_nat_vent_active` flag not propagating across `choose` block boundaries. Now pre-computes natural ventilation eligibility in the variables section as `nat_vent_would_activate`.
- **Missing pause/resume actions**: Added execution of `pause_action` and `resume_action` inputs which were defined but never called in the action sequence.

## [4.20.2] — 2025-12-28

### Fixed

- **Manual override tolerance conversion**: Fixed unit conversion for manual override tolerance that incorrectly treated the delta as an absolute temperature (subtracting 32 before multiplying by 5/9). For imperial users, this caused the tolerance to become a large negative number, making every thermostat change appear as a manual override.
- **Self-triggering detection**: Added logic to distinguish automation-initiated setpoint changes from true manual overrides by checking if the new setpoint is near the automation's predicted target. Previously, the automation's own temperature updates would trigger false manual override detections.

## [4.20.1] — 2025-12-27

### Fixed

- **Jinja2 boolean output**: Replaced bare `true`/`false` with `{{ true }}`/`{{ false }}` in variable assignments to output actual booleans instead of strings, preventing unexpected truthy behavior in conditionals.

## [4.20.0] — 2025-11-25

### Added

- **Precision/velocity offsets:** New humidity and air-velocity offsets in precision comfort mode to nudge targets when RH is high/low and air movement is above 0.3 m/s in warm conditions.
- **Adaptive fan control:** Computes temperature error vs. adaptive target and sets fan mode (auto/low/medium/high) when available.
- **HVAC mode cooldown:** Respects `mode_cooldown_min` before switching to `heat_cool` (safety re-enables unaffected).

### Changed

- **Humidity offset rework:** Unified humidity offset handling for precision and humidity comfort paths using `humid_offset_c`.
- **Occupancy/ventilation resilience:** Natural ventilation keeps HVAC off via `_nat_vent_active`, and occupancy fallbacks stay false on unknown states.

## [4.19.0] — 2025-11-25

### Added

- **Precision/Humidity offset:** Small humidity-driven offset applied when Precision Comfort or Humidity Comfort is enabled, cooling slightly when RH > 60% and warming when RH < 30%.
- **Occupancy-gated ventilation:** New `Only Prefer Ventilation When Occupied` toggle (default: on) blocks natural ventilation when no occupancy is detected.
- **Sensor validity guard:** Aborts runs when indoor/outdoor temps are invalid/out of range and logs when debug is enabled.

### Changed

- **Natural ventilation safety:** Tracks an internal `_nat_vent_active` flag so the final `climate.turn_on` is skipped after a ventilation decision, avoiding HVAC re-enables while windows are preferred.
- **Manual override detection:** Uses prediction-error vs. adaptive target (unit-aware tolerance) instead of raw delta to detect manual overrides.
- **Occupancy fallback:** Unknown/unavailable occupancy states now resolve to `false` rather than truthy, preventing unintended occupied behavior.

## [4.18.6] — 2025-11-15

### Changed

- **Directional safety re-enable:** When freeze-protect is enabled and indoor temperature approaches the freeze guard, the blueprint will re-enable HVAC from a manual off state in heat-only mode and hold the safety minimum. When overheat-protect is enabled and indoor temperature approaches the overheat guard, it will re-enable HVAC in cool-only mode and hold the safety maximum. Manual off is still respected when guards are disabled or risk is low.
- **Safety defaults updated:** Adjusted `safety_min_sys` and `freeze_guard_sys` defaults to 40°F (≈4.4°C) to better match typical freeze-protection recommendations.

## [4.18.5] — 2025-11-15

### Changed

- **Preemptive safety re-enable:** When freeze or overheat protection is enabled, the blueprint will re-enable HVAC from a manual off state once `risk_near` exceeds 0.7 (approaching configured guards), rather than waiting until the indoor temperature actually crosses the guard thresholds.

## [4.18.4] — 2025-11-15

### Changed

- **Safety-first auto-on when guards enabled:** Final `climate.turn_on` is now allowed even from a manual `hvac_mode: off` state *only* when freeze or overheat protection is enabled and currently blocking pause (`pause_blocked_now`), preserving manual off in normal conditions but restoring HVAC when approaching configured safety guards.

## [4.18.3] — 2025-11-15

### Changed

- **Learning decoupled from override duration:** Manual learning now triggers whenever a manual setpoint change exceeds the tolerance, even if manual override duration is set to 0 or override is disabled; override pause behavior (helper timestamp, stop, and custom override actions) only runs when override is enabled and duration > 0.
- **Respect manual HVAC off:** Final `climate.turn_on` call is now skipped when the climate entity's `hvac_mode` is currently `'off'`, so the blueprint will not automatically re-enable the system if you've manually turned it off.

## [4.18.2] — 2025-11-15

### Fixed

- **Learning helper update path instrumentation:** Added explicit debug log when writing to the learned offset helper and tightened the conditional to require a non-empty helper id string, making it easier to verify that the helper-write branch executes.

## [4.18.1] — 2025-11-14

### Fixed

- **Manual override persistence timestamp:** `is_override_active` now uses the `input_datetime` helper's `timestamp` attribute instead of parsing the string state, eliminating timezone parsing edge cases and making override checks robust across restarts.
- **Learning helper unit alignment:** When learning from manual adjustments, the blueprint now writes the learned offset back to the helper in system units (°C/°F) consistently with how it is read, preserving correct scaling in imperial setups.

### Documentation

- Updated blueprint input description for **Learned Offset Storage** to explicitly require a negative minimum (e.g., -10) and positive maximum so both warmer and cooler manual changes can be learned.
- Updated README version and learning helper setup instructions to call out the need for negative minimum values and clarify recommended helper configuration.

## [4.18] — 2025-01-08

### Optimized

- **Skip unnecessary thermostat commands:** Automation now checks if calculated setpoints match current thermostat state before sending commands.
  - Compares target temperatures to current temperatures with tolerance (0.6 × step size) to account for quantization.
  - Separately validates dual setpoint modes (auto/heat_cool) and single setpoint modes (heat/cool).
  - Only sends `climate.set_temperature` when setpoints differ from current values.
  - Only sends `climate.set_hvac_mode` when mode needs to change.
  - Reduces unnecessary writes that could trigger manual override detection.
  - Reduces network traffic for cloud-connected thermostats.
  - Reduces logbook noise from redundant setpoint commands.

### Added

- **Verbose logging for skipped commands:** When debug mode is `verbose`, automation logs when setpoints are unchanged and commands are skipped.
- **State comparison variables:** New helper variables `_dual_mode_matches`, `_single_mode_matches`, and `needs_temp_update` track whether thermostat commands are necessary.

### Technical Details

- Tolerance set to 60% of device step size (e.g., 0.3°F for 0.5°F step thermostats) to handle floating point rounding.
- Comparison happens after all unit conversions and quantization, ensuring accurate state matching.
- `climate.turn_on` always called (safe to call when already on).

### Performance Impact

- Estimated 50-80% reduction in thermostat service calls during stable conditions.
- Prevents false manual override detection on no-change updates.

## [4.17] — 2025-01-08

### Fixed

- **CRITICAL: Manual override persistence bug:** Fixed issue where manual override would not persist across automation restarts.
  - **Problem:** With `mode: restart`, any trigger (temp change, time pattern, etc.) would restart the automation, canceling the 60-minute delay and immediately resuming normal operation.
  - **Solution:** Replaced delay-based override with persistent `input_datetime` helper to track override expiry timestamp.
  - **Breaking change:** Manual override now **requires** `manual_override_until` input_datetime helper to function correctly.
  - Without helper configured, override is detected and logged, but will NOT persist across automation restarts (warning logged).

### Added

- **Manual Override Until helper input:** New required `input_datetime` helper stores when manual override expires.
- **Override persistence check:** Automation now checks `is_override_active` variable at start of each run to skip normal operation if override is still active.
- **Enhanced logging:** Override detection logs now indicate whether helper is configured ("helper set" vs "WARNING: no helper - will NOT persist!").

### Changed

- Manual override sequence no longer uses `delay` action, relying entirely on helper timestamp comparison.
- Override check happens before all other automation logic, ensuring override is respected even when automation is triggered frequently.

### Migration Required

1. Create `input_datetime` helper: Settings → Devices & Services → Helpers → Date and time
2. Name: "Adaptive HVAC Manual Override Until"
3. Enable "Date and time"
4. Select helper in blueprint: "Manual Override Until (input_datetime helper)"

### Technical Details

- Automation has 13+ triggers including time patterns (every 10 minutes), temperature changes (30s/60s debounce), and state changes.
- Previous implementation assumed no triggers would fire during 60-minute delay, which was unrealistic.
- New implementation sets helper to `now() + 60 minutes` and checks `now() < helper` on every run.
- Override state persists across HA restarts, automation reloads, and trigger-based restarts.

## [4.16] — 2025-11-07

### Improved

- **UI clarity for temperature unit fields:** Added visual indicators and clearer descriptions for Celsius/Fahrenheit input fields.
  - Temperature fields now labeled "❄️ Minimum Comfort (°C) — Metric Only" and "🔥 Maximum Comfort (°F) — Imperial Only"
  - Descriptions now explicitly state when each field is used vs ignored
  - Units override selector options improved: "Auto-detect from sensors", "Force Celsius (°C)", "Force Fahrenheit (°F)"
  - Added warning emoji and guidance in units_override description to clarify which fields to use

### Notes

- Home Assistant blueprints do not support conditional input visibility, so both °C and °F fields remain visible.
- This update makes it clearer which fields are active based on the "Temperature Units" setting.

## [4.15] — 2025-11-07

### Added

- **Adaptive Learning from Manual Overrides:** Blueprint now learns your temperature preferences over time.
  - Implements exponential weighted average algorithm to track manual adjustments.
  - Calculates error between manual temperature and predicted comfort temperature.
  - Updates learned offset using configurable learning rate (default: 0.15).
  - Formula: `new_offset = (1 - α) * old_offset + α * error`
  - Stores learned offset in optional `input_number` helper for persistence across restarts.
  - Applied on top of ASHRAE-55 base model + seasonal bias + sleep bias + CO₂ bias.
  - Works with both single setpoint (heat/cool) and dual setpoint (auto/heat_cool) modes.

### Configuration

- **Learn from Manual Adjustments:** Enable/disable learning (default: enabled)
- **Learning Rate:** 0.05-0.5, controls adaptation speed (default: 0.15)
- **Learned Offset Storage:** Optional `input_number` helper for persistence

### How It Works

1. User manually adjusts thermostat (e.g., 70°F → 73°F)
2. Blueprint calculates error: `manual_temp - predicted_temp = +3°F`
3. Updates learned offset: `new = 0.85 * old + 0.15 * 3°F`
4. Applies learned offset to all future comfort calculations
5. Gradually adapts to your preferences while respecting seasonal patterns

### Debug Logging

- Enhanced override detection logs show:
  - Manual temperature vs predicted temperature
  - Calculated error in °C
  - Old learned offset → new learned offset
  - Pause duration

### Notes

- **Seasonal adaptation preserved**: Learned offset adds to regional/seasonal biases (important for Colorado's Mixed-Dry climate).
- **Conservative by default**: 0.15 learning rate means full adaptation takes ~10-15 adjustments.
- **Optional persistence**: Without helper, offset resets on HA restart (still learns within session).
- **Compatible with all existing features**: Works alongside RMOT, CO₂, sleep mode, etc.

## [4.14] — 2025-11-07

### Optimized

- **Trigger debouncing:** Added 30-second delay for indoor temp, 60-second for outdoor temp to reduce rapid re-evaluations by ~70-80%.
- **Lazy debug computation:** Debug strings now computed only when debug mode is enabled, saving template rendering on every run.
- **Inline band variables:** Eliminated intermediate `_band_min_ordered` and `_band_max_ordered` variables by inlining calculations.
- **Unit conversion helpers:** Added `_to_sys_mult` and `_to_sys_add` variables to reduce repetition in temperature conversions.
- **Vendor separation lookup:** Replaced 40+ line if-elif chain with dict lookup table, reducing code by ~85% and improving maintainability.
- **Removed unused variable:** Deleted `vendor_sep_sys_default` which was defined but never referenced.

### Performance Impact

- Estimated 70-80% reduction in automation triggers during temperature fluctuations.
- Faster template evaluation when debug is disabled.
- More maintainable vendor profile management.

### Notes

- All changes are backward-compatible with no behavioral changes.
- Optimization focused on reducing unnecessary computation and improving code clarity.

## [4.13] — 2025-11-07

### Added

- **Manual Override Detection:** Automatically pauses automation when users manually adjust thermostat setpoints.
  - Configurable pause duration (default: 60 minutes, up to 8 hours).
  - Adjustable detection tolerance to prevent false triggers from minor drift.
  - Optional notification action when override is detected.
  - Monitors single setpoint (heat/cool) and dual setpoint (auto/heat_cool) modes.
  - Respects user control while automatically resuming adaptive comfort after the pause period.

### Notes

- Feature is enabled by default but can be disabled via blueprint input.
- Tolerance default (1.0 climate unit) prevents triggering on quantization noise.
- When override detected, automation stops and waits for configured duration before resuming.

## [4.12] — 2025-11-07

### Fixed

- **Debug logging unit conversion:** Corrected `sep_cli_min` output that was incorrectly converting already-converted climate units to system units.

### Improved

- **Division by zero protection:** Added safety guard for humidity ratio calculation to prevent edge cases where vapor pressure approaches total pressure.
- **Code documentation:** Added explanatory comments for complex logic sections:
  - Natural ventilation psychrometric guards (4-part condition explanation)
  - Minimum separation enforcement for Auto/Heat-Cool modes
  - Comfort band ordering and inversion prevention
  - Risk acceleration asymmetry (50% strength for open delay vs full strength for resume)

### Notes

- No behavioral changes to automation logic.
- Improved maintainability and debugging clarity.

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
