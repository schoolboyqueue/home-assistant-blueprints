# Adaptive Shades Pro - Changelog

## [1.13.12] - 2026-01-03

### Fixed

- **False manual detection on attribute-only updates**: When a cover's position changed but its state remained the same (e.g., "open" at 50% → "open" at 60%), the `trigger.to_state.last_changed` timestamp was stale (from the last state change, not the position change). This caused `seconds_since_cmd` to be negative (cover's `last_changed` older than the helper timestamp), which failed the `>= 0 and < grace_seconds` check and incorrectly flagged the change as manual. Now explicitly handles negative `seconds_since_cmd` as "attribute-only update, not manual" — matching the fix already present in `auto_override_active` but missing from `manual_change_detected`.

## [1.13.11] - 2026-01-01

### Changed

- Standardized logbook message format to use consistent pipe-separated key=value style for improved readability and easier log parsing.

## [1.13.10] - 2025-12-30

### Fixed

- **Helper timestamp updated on every trigger**: The `last_command_helper` timestamp was being updated at the start of every automation run (including sun state changes), even when no shade movement occurred. This caused the grace period logic to be ineffective because the timestamp kept refreshing. Now the helper timestamp is only updated immediately before actually sending a command to the cover.

## [1.13.9] - 2025-12-28

### Fixed

- **Jinja2 min/max syntax error**: Fixed `'list object' has no attribute 'min'` error in `target_angle_deg` calculation. Replaced Python-style list method syntax (`[a, b].min()`) with proper Jinja2 filter syntax (`[a, b] | min`).

## [1.13.8] - 2025-12-28

### Fixed

- **Quiet hours time parsing robustness**: Replaced `strptime()` with string splitting to handle both `HH:MM` and `HH:MM:SS` time formats. Previously, if the time selector output didn't match the expected format exactly, quiet hours detection would fail silently.

## [1.13.7] - 2025-12-27

### Fixed

- **Fixed false manual detection in multi-cover setups**: When the helper timestamp was updated at the start of a run, covers that hadn't moved recently would have a `last_changed` timestamp BEFORE the helper timestamp, causing `seconds_since_cmd` to be negative. This was incorrectly interpreted as a manual change. Now explicitly checks for `changed_before_cmd` and treats it as old state (not manual).

## [1.13.6] - 2025-12-27

### Fixed

- **Jinja2 boolean output**: Replaced bare `true`/`false` with `{{ true }}`/`{{ false }}` across all variable assignments to output actual booleans instead of strings, preventing unexpected truthy behavior in conditionals.

## [1.13.5] - 2025-12-27

### Fixed

- **Fixed inverted shades not closing at night**: The `this_window_open` variable was outputting the literal string `"false"` instead of a boolean `false`. In Jinja2, a non-empty string is truthy, so the condition `{% elif this_window_open %}` was always true, causing inverted shades to always go to position 100 (fully open) instead of the correct night position. Now outputs proper booleans with `{{ false }}` / `{{ true }}` and adds explicit `| bool` filter for safety.

- **Fixed future helper timestamp causing false manual detection**: When a helper timestamp was in the future (due to timezone issues or corrupt data when first created), the manual detection logic would incorrectly evaluate it as valid, causing all automation movements to be blocked. Now explicitly rejects helper timestamps that are in the future (`helper_age < 0`).

## [1.13.4] - 2025-12-26

### Fixed

- **Fixed multi-cover support**: Two issues with multiple covers per automation:
  1. The helper timestamp was updated per-cover inside the loop, causing timing issues when covers report state asynchronously. Now updates helper ONCE before the loop starts.
  2. The `stop: true` for manual override detection was stopping the entire automation instead of just skipping that cover. Now uses a condition to skip remaining actions for that cover while continuing to process other covers.

## [1.13.3] - 2025-12-26

### Fixed

- **Fixed cover trigger false positives**: When the automation moved a cover, the subsequent device state report would trigger `cover_changed` and be falsely detected as manual (device reports have no `parent_id`). The `manual_change_detected` variable now uses the helper timestamp to check if the cover change occurred within 90 seconds of the last automation command, preventing false positives from automation-triggered movements.

## [1.13.2] - 2025-12-26

### Fixed

- **Fixed fresh helper causing false positives**: A newly created `input_datetime` helper with no automation-set value would cause all cover changes to be flagged as manual. Now validates that the helper timestamp is recent (within 2x the manual timeout period). If the helper hasn't been set by this automation yet, falls back to UI-only detection.

## [1.13.1] - 2025-12-26

### Fixed

- **Fixed helper timestamp parsing**: The `input_datetime` helper stores state as `YYYY-MM-DD HH:MM:SS` which `as_timestamp()` cannot parse. Now uses `state_attr(helper, 'timestamp')` to get the Unix timestamp directly, which is the correct way to read `input_datetime` values for time comparisons.

## [1.13.0] - 2025-12-26

### Added

- **Last command timestamp helper** (`last_command_helper`): New optional `input_datetime` helper input that enables reliable detection of manual adjustments from physical remotes or buttons. The automation records a timestamp before each cover command; any cover state change that occurs more than 90 seconds after the last command is considered manual. Create one helper per automation instance with **"Date and time" mode enabled** (not just date or just time) and select it in the blueprint configuration.

### Changed

- **Improved manual detection logic**: With the helper configured, the automation can now reliably distinguish between:
  - Changes from this automation (within 90s of recorded command timestamp)
  - Changes from Home Assistant UI (detected via `user_id` in context)
  - Changes from other automations (detected via `parent_id` in context)
  - Changes from physical remotes/buttons (detected as manual when helper is configured)
  
  Without the helper, behavior is unchanged from v1.12.4 (only UI changes detected as manual).

## [1.12.4] - 2025-12-26

### Fixed

- **Fixed false positive manual detection**: Removed unreliable position-comparison fallback from manual adjustment detection. The previous logic compared current position to calculated target when no `user_id` or `parent_id` was present in state context, but this caused false positives because device state reports after automation commands have no `parent_id` and may differ from target due to rounding or drift. Manual detection now only triggers on clear UI-initiated changes (`user_id` present). Physical remote/button presses cannot be reliably distinguished from automation results and are no longer detected—users should use the manual override input_boolean for explicit manual control periods.

## [1.12.3] - 2025-12-26

### Fixed

- **Fixed attribute access error**: Resolved `ReadOnlyDict has no attribute 'current_position'` error by using `state_attr()` helper instead of direct attribute access on state objects.

## [1.12.2] - 2025-12-26

### Fixed

- **Fixed circular variable dependency error**: Resolved `UndefinedError: 'desired_position_room' is undefined` that occurred during variable rendering. The `manual_change_detected` and `auto_override_active` variables were referencing `desired_position_room` before it was defined in the variable chain. Manual detection now relies on context-based detection (UI vs automation) at the top level, with position-based comparison deferred to the per-cover loop where all variables are available.

## [1.12.1] - 2025-12-26

### Fixed

- **Pause conditions now properly stop execution**: Added missing `stop:` actions for manual override, quiet hours, and presence requirement conditions. Previously, these conditions would log a message but continue to move shades anyway.
- **Prevent commanding shades while moving**: The automation now skips sending position commands when a cover is in `opening` or `closing` state, preventing interruption of in-progress movements (whether from automation or manual control).
- **Improved manual adjustment detection**: The v1.11.0 `parent_id` approach caused false positives when devices reported state updates after automation commands, blocking all automatic movements. New approach detects manual changes by:
  1. Ignoring state changes while cover is actively moving (`opening`/`closing` states)
  2. UI interactions (`user_id` present) - always treated as manual
  3. Automation/script changes (`parent_id` present) - never treated as manual
  4. Device state reports (no `user_id` or `parent_id`) - treated as manual only if the final position differs significantly from the automation's target position (beyond the hysteresis threshold)
  
  This correctly detects physical remote control, voice assistant commands, and native app changes without false positives from intermediate position reports during shade movement or final position reports after automation-triggered movements.

## [1.12.0] - 2025-12-25

### Added

- **Window open delay** (`window_open_delay`): New configurable delay (0–300 seconds) before a shade opens when its mapped window/door sensor triggers. Set to 0 for immediate response, or use a delay (e.g., 30–60 seconds) to avoid shade movements when briefly opening a window.

## [1.11.0] - 2025-12-24

### Added

- **Window contact trigger**: Added state trigger for `window_contacts` with id `window_changed` so the automation reacts immediately when a window contact sensor changes state, rather than waiting for the next 5-minute time pattern trigger.

## [1.10.0] - 2025-11-29

### Added

- **Night/quiet hours position** (`night_position_pct`): New input to set a dedicated shade position for nighttime and quiet hours, independent of the glare block position. Previously, `block_position_pct` served both purposes.
- **Minimum position change** (`position_hysteresis`): New configurable input (default 3%) to control the minimum position delta required before shades move. Set higher to reduce frequent small adjustments, or lower for more precise control.

### Changed

- Renamed "Preferred block position" to "Glare block position" for clarity—this position is now used exclusively for blocking direct sun/glare during daytime.
- Night and quiet hours behavior now uses the dedicated `night_position_pct` instead of `block_position_pct`.

## [1.9.3] - 2025-11-25

### Fixed

- Ensures zebra shades close to the configured block position whenever the sun is below the horizon, even on manual runs or if other triggers were missed.

## [1.9.2] - 2025-11-25

### Fixed

- Manual adjustment timeout now only activates when the last cover change includes a user context (manual interaction). Automation-initiated movements no longer trigger the timeout.

## [1.9.1] - 2025-11-25

### Fixed

- Manual adjustment timeout is now evaluated per cover inside the repeat loop, so manual moves on any shade pause automation for that shade.

## [1.9.0] - 2025-11-25

### Added

- Optional window contact mapping (one per shade, index-matched) that forces a shade fully open when its window is open.
- Per-cover loop for movement, hysteresis, and logging; honors inverted shades per cover.
- Presence/sun/comfort logic remains shared at room level.

## [1.8.4] - 2025-11-25

### Changed

- Expanded manual comfort setpoint ranges (0–120) so users can input Fahrenheit or Celsius; climate-derived setpoints still take priority when a climate entity is provided.

## [1.8.3] - 2025-11-25

### Fixed

- Multi-cover support hardened: state/feature lookups now use the first cover (base_cover) to avoid unhashable list errors in HA templates.

## [1.8.0] - 2025-11-25

### Added

- Presence-required toggle and refined presence/glare handling for zebra mode: presence now gates movement (when required) and drives glare-sensitive behavior while unoccupied paths still consider heating/cooling needs.

## [1.7.0] - 2025-11-25

### Added

- Inverted shades option to flip open/block percentages for devices that report 0% as fully closed and 100% as fully open.

## [1.6.0] - 2025-11-25

### Added

- Shade cover selector now supports multiple covers, enabling grouped shade control.

### Fixed

- Manual override timeout now uses the first selected cover safely for last_changed and supported_features checks.

## [1.5.7] - 2025-11-25

### Fixed

- Clear-sky irradiance calculation now uses Jinja `| max` on a list (replacing Python `.max()`), preventing template errors in Home Assistant.

## [1.5.6] - 2025-11-25

### Fixed

- Adjusted debug level selector to use simple string options for Home Assistant compatibility and aligned version metadata.

## [1.5.4] - 2025-11-25

### Added

- Diagnostics input with `debug_level` (off/basic/verbose) and gated logbook logging, including verbose no-movement traces for troubleshooting.

## [1.5.3] - 2025-11-25

### Fixed

- Replaced unsupported Jinja `exp` filter in clear-sky irradiance calculation with explicit exponent math (uses `e_const ** exponent`).

## [1.5.2] - 2025-11-25

### Fixed

- Replaced unsupported Jinja `radians` filter with explicit degree-to-radian conversion for tilt and sun geometry math to prevent template parse errors.

## [1.5.1] - 2025-11-25

### Fixed

- Corrected tilt capability detection template to avoid template parse errors in Home Assistant (bitwise detection now uses `bitwise_and`).

## [1.5.0] - 2025-11-25

### Added

- Optional climate entity to bias heating/cooling classification and reflect HVAC state in shading decisions.
- Room profile (living/office/bedroom) that adjusts glare sensitivity; effective glare threshold derived from profile.

### Changed

- Comfort mode detection now prefers active HVAC state when provided, falling back to setpoints.
- Comfort setpoints now derive from the climate entity when provided; manual setpoints are used only when no climate is configured.
- Glare detection uses the profile-adjusted threshold; documentation updated for clarity.

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

- Auto-compute minimum sun elevation from Home Assistant's `sun.sun` (uses 1° when above horizon, otherwise disables sun-on-window) so users no longer configure this threshold manually.

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
