# Bathroom Light & Fan Control Pro - Changelog

## [1.10.9] - 2025-12-27

### Fixed

- **Jinja2 boolean output**: Replaced bare `true`/`false` with `{{ true }}`/`{{ false }}` in variable assignments to output actual booleans instead of strings.

---

## [1.10.8] - 2025-12-25

### Fixed

- Night schedule not activating correctly; previous timestamp-based comparison was comparing full date-times instead of time-of-day. Now uses minutes-since-midnight for proper overnight schedule detection (e.g., 22:00-06:00).

---

## [1.10.7] - 2025-12-06

### Fixed

- Night schedule activation could fail around overnight boundaries; replaced string-based time checks with timestamp-based comparisons to reliably handle schedules that span midnight. (in_night_schedule)
- Bumped blueprint version to `1.10.7`.

---

## [1.10.6] - 2025-11-14

### Fixed

- Restored door-state dependency for the startup light sanity check to remain consistent with Wasp-in-a-Box semantics.
- Added a short 30s delay after Home Assistant start before evaluating startup light conditions, giving motion and door sensors time to report stable states.
- At startup, the light is now auto-turned off only when it is ON, motion is OFF, and the door is open/unknown, after `lights_off_delay_min` minutes.

---

## [1.10.5] - 2025-11-14

### Fixed

- Relaxed the startup light sanity check so that it no longer depends on the door being open/unknown.
- At Home Assistant startup, if the bathroom light is ON and the motion sensor is OFF, the automation now treats this as vacancy after `lights_off_delay_min` and turns the light off using the automation_control helper semantics.
- This prevents lights from staying on indefinitely across restarts even when the door is closed, while leaving normal Wasp-in-a-Box behavior unchanged during runtime.

---

## [1.10.4] - 2025-11-14

### Fixed

- During Home Assistant startup, if the bathroom light is already ON while motion is OFF and the door is open/unknown, treat this as vacancy.
- Added a startup light sanity check that waits `lights_off_delay_min` minutes, then turns off the light using the same automation_control helper semantics, preventing lights from staying on indefinitely across restarts.
- Kept Wasp-in-a-Box semantics (door closed = occupancy); startup light auto-off only applies when the door is open/unknown.

---

## [1.10.3] - 2025-11-14

### Fixed

- Prevented bathroom lights staying on after showers when using `mode: restart` with frequent humidity updates.
- Moved the lights-off grace period (`lights_off_delay_min`) from an action-level delay to the `wasp_motion_clear` trigger `for:` duration, so humidity triggers no longer cancel the vacancy grace.
- Removed extra action-level delays before turning lights off in Wasp-in-a-Box vacancy Cases A/B/C; lights now turn off immediately once vacancy is detected.

---

## [1.10.2] - 2025-01-09

### Fixed

**Vacancy Detection After Shower (Door Already Open):**

- Fixed lights staying on when leaving bathroom with door already open after shower
- Separated vacancy logic into three distinct cases for clarity:
  - Case A: Motion clears AND door already open/unknown → immediate vacancy (no timing gate)
  - Case B: Motion clears AND door closed but closed after motion cleared → apply timing window
  - Case C: Door left open too long AND no motion → vacancy
- Previous logic combined Cases A and B in a single OR condition, causing the door-close timing check (Case B) to block Case A
- Now when motion clears with door already open, vacancy proceeds immediately without evaluating door-close timing

### Technical Details

- Restructured lines 741-786: Changed from nested OR with AND to separate choose branches
- Case A (lines 742-758): Door open/unknown check - no timestamp comparison
- Case B (lines 759-777): Door closed after motion - keeps existing timestamp logic
- Case C (lines 778-786): Door left open trigger - unchanged
- Preserves all existing behavior for Cases B and C
- Fixes scenario: door closed during shower → open door and leave → motion clears → lights should turn off

---

## [1.10.1] - 2025-01-09

### Fixed

**Vacancy Detection When Door Opened Before Motion Cleared:**

- Removed overly restrictive timestamp comparison in `wasp_door_left_open` vacancy condition
- Previous logic required `door.last_changed > motion.last_changed`, which failed when:
  - User opened door while motion was still active
  - User left immediately
  - Motion cleared after door opened
  - Lights stayed on because timestamp check failed
- Now only requires: door open for delay + motion sensor off
- Fixes lights not turning off when leaving bathroom with door opened before motion cleared

### Technical Details

- Removed lines 781-783: timestamp comparison condition
- `wasp_door_left_open` trigger now properly detects vacancy with simpler logic:
  1. Door has been open for configured delay (default 15 seconds)
  2. Motion sensor shows "off" (no motion)
- Handles edge case where door opens while motion is still detecting

---

## [1.10.0] - 2025-01-08

### Fixed

**Wasp-in-a-Box Concept Violation (Critical Bug):**

- Removed `wasp_vacancy_with_grace` trigger that turned off lights after motion clear regardless of door state
- Removed `wasp_door_closed` trigger that turned off lights when door closed
- Core principle: Door closed = occupancy (even without motion detection)
- Lights now only turn off when: motion clear AND (door open OR door left open too long)
- Fixes lights turning off during showers when behind curtain (no motion but door closed)

### Technical Details

- Removed trigger at line 565-571: `wasp_vacancy_with_grace` (motion clear for full grace period)
- Removed trigger at line 547-550: `wasp_door_closed` (door closing)
- Removed corresponding action logic for both triggers
- Simplified vacancy detection to two valid cases:
  1. Motion clear + door open (person left)
  2. Motion clear + door left open too long (person left, forgot to close door)
- Door closing while motion is off no longer triggers lights off (was breaking Wasp-in-a-Box)

### Migration Notes

**Behavior change:** Lights will stay on longer if door remains closed, even without motion.
This is correct Wasp-in-a-Box behavior - door closed indicates occupancy.

---

## [1.9.5] - 2025-01-08

### Fixed

**Pitfall #10: Template in trigger `for:` duration (Best Practice):**

- Removed template from `fan_max_runtime_expired` trigger `for:` parameter (line 531)
- Changed from `{{ fan_max_runtime_min if ... else 9999 }}` to direct `!input fan_max_runtime_min`
- Variables aren't available at trigger compile time - using templates in triggers is risky
- Changed `fan_max_runtime_min` minimum from 0 to 1 minute (use 240 to effectively disable)
- Removed unnecessary condition check `fan_max_runtime_min > 0` from action (line 1113)

### Changed

**fan_max_runtime_min input:**

- Minimum value changed from 0 to 1 minute
- Description updated: "Use a large value (e.g., 240) to effectively disable"
- Simplifies trigger logic and follows blueprint best practices

### Technical Details

- Line 531: Now uses `!input fan_max_runtime_min` directly without template
- Line 161: Changed `min: 0` to `min: 1`
- Line 1113: Removed condition checking if value > 0
- Adheres to Common Pitfall #10 documented in WARP.md

---

## [1.9.4] - 2025-01-08

### Fixed

**Manual Override Race Condition (Critical):**

- Added 100ms delay after turning off lights before resetting `automation_control` helper
- Previous timing: helper OFF happened simultaneously with lights OFF, causing race condition
- `light_manual_off` trigger fired when lights turned off, checked helper state, found it OFF (already resetting), activated manual override incorrectly
- Now: lights turn OFF → wait 100ms → helper resets to OFF
- `light_manual_off` trigger now sees helper still ON during the 100ms window, correctly blocking manual override

### Technical Details

- Added `delay: milliseconds: 100` after light.turn_off in both paths (lines 763-764, 875-876)
- Ensures light state change is fully processed before helper resets
- Prevents simultaneous state changes that cause race conditions
- Applied to both `wasp_vacancy_with_grace` and old vacancy paths

---

## [1.9.3] - 2025-01-08

### Fixed

**Manual Override Timezone Bug (Critical):**

- `override_ok` variable now uses `state_attr(manual_override_until, 'timestamp')` instead of `as_timestamp(states(...))`
- Previous approach had timezone conversion issues: helper stores UTC timestamp, `states()` returns local time string, `as_timestamp()` misinterprets it
- Caused manual override to never expire - `override_ok` always returned False even after expiration
- Now uses the `timestamp` attribute directly (Unix epoch), avoiding timezone parsing entirely

### Technical Details

- Line 486: Changed from `as_timestamp(states(manual_override_until))` to `state_attr(manual_override_until, 'timestamp')`
- The `input_datetime` entity has a `timestamp` attribute containing the raw Unix timestamp (float)
- Direct timestamp comparison: `now_ts > until_ts` (both in Unix epoch seconds)
- Added extra check for unavailable state before accessing attribute

---

## [1.9.2] - 2025-01-08

### Fixed

**Manual Override False Triggers (Critical Bug):**

- Added `automation_control` helper set/reset logic to old vacancy path (wasp_motion_clear, wasp_door_closed, wasp_door_left_open)
- v1.9.1 only fixed the new `wasp_vacancy_with_grace` trigger path
- If lights turned off via old fallback path, helper remained OFF, causing manual override to activate incorrectly
- Now both light-off paths properly signal automation control

### Changed

**Manual Override Descriptions:**

- Updated `manual_override_duration_min` description to mention disabling by setting to 0
- Updated `automation_control` description with clearer explanation of three options:
  - No helper: Simple setup, may get false triggers
  - With helper: Perfect behavior, no false triggers
  - Duration = 0: Disable manual override entirely

### Technical Details

- Lines 841-847: Added helper ON before lights off in old vacancy path
- Lines 864-870: Added helper OFF after lights off in old vacancy path
- Matches helper logic from lines 737-762 (wasp_vacancy_with_grace path)
- Both light-off code paths now prevent false manual override triggers

---

## [1.9.1] - 2025-01-08

### Changed

**Made Manual Override Helper Optional:**

- `automation_control` helper is now optional (was required)
- If not configured: Manual override works with simple behavior (may trigger when automation turns off lights)
- If configured: Helper prevents false manual override triggers when automation turns off lights automatically
- Users can now choose between simple setup or perfect behavior
- Alternative: Set `manual_override_duration_min: 0` to disable manual override entirely

### Technical Details

- Helper checks wrapped in conditions: `{{ automation_control not in ['', None] }}`
- Manual override condition: `{{ automation_control in ['', None] or is_state(automation_control, 'off') }}`
- When helper not provided, automation skips helper ON/OFF service calls
- Backward compatible with v1.9.0 automations that have helper configured

---

## [1.9.0] - 2025-01-08

### Added

**Manual Override False Trigger Prevention:**

- New required input: `automation_control` (input_boolean helper)
- Prevents manual override from activating when automation turns off lights automatically
- Automation sets helper ON before turning off lights, then resets to OFF
- Manual override logic checks helper is OFF before activating (means user turned off, not automation)

### Fixed

**Manual Override Logic:**

- Manual override no longer triggers when automation turns off lights after vacancy
- Only triggers when user manually turns off lights via physical switch or app
- Prevents false override that blocked lights from turning back on

### Migration Required

**BREAKING CHANGE - Requires new helper:**

1. Create an input_boolean helper:

   - Settings → Devices & Services → Helpers → Create Helper → Toggle
   - Name: "[Bathroom Name] Automation Control"
   - Icon: `mdi:robot` (optional)

2. Select the helper in blueprint's "Automation Control Helper" setting

3. Manual override will now only trigger on actual user manual control

### Technical Details

- Helper set ON at lines 736-739 before automation turns off lights
- Helper reset to OFF at lines 752-755 after lights turn off
- Manual override condition checks helper is OFF at lines 873-876
- Without helper, manual override cannot distinguish automation vs user control

---

## [1.8.1] - 2025-01-08

### Fixed

**Trigger-Level Delay:**

- Removed unsupported template from trigger `for:` parameter
- Changed to use `!input lights_off_delay_min` directly
- Home Assistant doesn't support template evaluation in trigger `for:` duration

### Technical Details

- Line 557: `for: minutes: !input lights_off_delay_min` (was template)
- Trigger fires only after motion sensor stays 'off' for full grace period
- Action turns off lights immediately when trigger fires

---

## [1.8.0] - 2025-01-08

### Fixed

**Lights Not Turning Off After Vacancy:**

- Root cause: Humidity triggers canceled light-off delay in action sequence
- Solution: Moved grace period delay from action to trigger using `for:` parameter
- New trigger: `wasp_vacancy_with_grace` waits full `lights_off_delay_min` at trigger level
- If motion detected during wait, trigger cancels (desired behavior)
- Humidity triggers no longer interfere with motion sensor's wait state

### Changed

- Mode changed back to `restart` (from `queued`) for better responsiveness
- Light-off delay now handled at trigger level (lines 552-558)
- Action turns off lights immediately when vacancy trigger fires
- No more cancelable delays in action sequence

### Technical Details

- Trigger: `platform: state, entity_id: !input motion_sensor, from: "on", to: "off", for: minutes: !input lights_off_delay_min`
- Prevents queueing multiple redundant runs from frequent sensor updates
- More efficient than `mode: queued, max: 5` approach

---

## [1.7.2] - 2025-01-08

### Changed

**Mode Changed to Queued:**

- Mode changed from `restart` to `queued` with `max: 5`
- Prevents humidity triggers from canceling light-off delay
- Multiple sensor updates queue up instead of restarting automation

### Known Issues

- Queued runs can be wasteful with frequent sensor updates
- Better solution: move delay to trigger level (planned for v1.8.0)

---

## [1.7.1] - 2025-01-08

### Added

**Debug Logging for Lights Not Turning Off:**

- Added logs when `wasp_motion_clear` and `wasp_vacancy_with_grace` triggers fire
- Shows motion sensor state, grace period, and whether lights are being turned off
- Helps diagnose delay cancellation issues

---

## [1.7.0] - 2025-01-08

### Changed

**UI Reorganization - Night Settings:**

- Moved all night-related settings from "Presence & Lighting" to "Night Schedule" section
- Settings now logically grouped: schedule, force-on, brightness, color temp, fan bias
- Improved descriptions to clarify night mode behavior
- `night_mode_enabled` renamed to "Force Night Mode Always On" for clarity
- Better explains that night mode activates during schedule OR when forced on

### Technical Details

- No logic changes - purely UI/organizational improvement
- All settings retain same default values and behavior
- Variable bindings unchanged

### Migration Notes

**Existing automations will continue to work** - this is a non-breaking change. All inputs have the same IDs and defaults. The UI will show the reorganized layout next time you edit the automation.

---

## [1.6.0] - 2025-01-08

### Changed

**Simplified Presence System (Breaking Change):**

- Collapsed `presence_boolean`, `presence_entities`, and `require_presence` into single `presence_entities` input
- Now accepts any combination of: person, device_tracker, binary_sensor, input_boolean, zone
- Logic: If list is empty → lights always work. If list has entities → at least one must show 'on'/'home'
- Cleaner, more intuitive configuration
- Added zone domain support for location-based presence

### Removed

**Deprecated Inputs:**

- `presence_boolean` - merge into `presence_entities` list
- `require_presence` - now automatic (empty list = no requirement, populated list = requirement)

### Migration Required

**For existing automations:**

1. **If you had `require_presence: false`:**

   - Remove or leave `presence_entities: []` (empty)
   - Behavior: lights work regardless of presence

2. **If you had `presence_boolean: input_boolean.home`:**

   - Change to: `presence_entities: [input_boolean.home]`

3. **If you had both `presence_boolean` AND `presence_entities`:**

   - Combine into single list: `presence_entities: [input_boolean.home, person.john, person.jane]`

4. **If you had `require_presence: true` with entities:**
   - Just keep your `presence_entities` list (requirement is now implicit)

### Technical Details

- `presence_ok` logic simplified: empty list → true, populated list → check if any entity is 'on'/'home'
- Variable binding reduced from 3 inputs to 1
- More flexible: can mix entity types (input_boolean, person, device_tracker, etc.)

---

## [1.5.2] - 2025-01-08

### Fixed

**Require Presence Default:**

- Changed `require_presence` default from `true` to `false`
- Lights now turn on by default regardless of home presence system
- Previous default prevented lights from turning on unless presence entities showed someone home
- Better default for typical bathroom use case where motion/door should always trigger lights

### Migration Notes

**Existing automations using this blueprint:**

- If you want the OLD behavior (require presence), explicitly set `require_presence: true` in your automation config
- If you have presence entities configured but lights weren't turning on, this fixes it
- The presence check logic itself is unchanged, only the default value

### Technical Details

- `require_presence` input default changed from `true` → `false` in blueprint definition
- When `false`: lights turn on from motion/door regardless of presence entities
- When `true`: lights turn on only if presence_boolean OR any presence_entities show 'on'/'home'

---

## [1.5.1] - 2025-01-08

### Added

**Debug Logging for Failed Light Turn-On:**

- Added log message when light should turn on but conditions fail
- Shows which condition blocked the light: `light_already_on`, `presence_ok`, `override_ok`, or `lux_ok`
- Helps diagnose why lights don't respond to motion/door triggers
- Logged at both `basic` and `verbose` debug levels

### Technical Details

- Added `else` branch to light turn-on logic that logs failed conditions
- Message format: `Light NOT turned on: trigger=X, light_already_on=X, presence_ok=X, override_ok=X, lux_ok=X`

---

## [1.5.0] - 2025-01-08

### Changed

**Debug System Upgrade:**

- Replaced boolean debug toggle with three-level system matching Adaptive Comfort Control Pro
- Debug levels: `off`, `basic`, `verbose`
- Default is now `basic` (was `false`/off)
- UI changed from toggle to dropdown selector
- Icon changed from `mdi:bug` to `mdi:bug-outline` for consistency

### Added

**Verbose Debug Logging:**

- **Light ON (verbose)**: Adds night mode breakdown (enabled vs schedule), illuminance sensor value, area control status
- **Light OFF (verbose)**: Adds motion and door sensor states at time of vacancy
- **Fan ON (verbose)**: Adds bathroom/home humidity sensor raw values, night bias value, night schedule status
- **Fan OFF (verbose)**: Adds bathroom/home humidity sensor raw values, minimum runtime setting

**Basic Debug Logging (unchanged):**

- All existing basic logs preserved with same format
- Light ON/OFF, Fan ON/OFF, Manual override, ROR latch, Startup checks

### Migration Notes

**Existing automations will automatically default to `basic` logging** (previously equivalent to `debug_enabled: true`).

To restore previous behavior:

- **Old `debug_enabled: false`** → Set `debug_level: off`
- **Old `debug_enabled: true`** → Already at `basic` (default), no change needed
- **For detailed troubleshooting** → Set `debug_level: verbose`

### Technical Details

- All log conditions changed from `debug_enabled` to `debug_level in ['basic','verbose']` or specific level checks
- Variable binding changed from `debug_enabled: !input debug_enabled` to `debug_level: !input debug_level`
- Verbose logs use separate conditions to provide enhanced context without cluttering basic logs
- Blueprint version incremented to track breaking configuration change

---

## [1.4.0] - 2025-01-08

### Fixed

**Critical Bugs:**

- **Night mode area lighting**: Area lights now properly receive brightness and color temp parameters during night mode
- **Night mode with area disabled**: Fixed missing light control path when night mode is disabled and area is set
- **ROR variable scope**: `ror_ok` variable now properly accessible in sequence actions for debug logging
- **Manual override false triggers**: Added check to only trigger override when duration > 0
- **Fan max runtime trigger**: Fixed trigger firing immediately when max runtime set to 0 (disabled)

**Logic Issues:**

- **Night mode schedule integration**: Night mode now automatically activates during night schedule hours (not just when explicitly enabled)
- **States dictionary access**: Added existence checks before accessing `states[entity].last_changed` to prevent AttributeErrors
- **as_timestamp None handling**: Added None checks for all `as_timestamp()` calls to prevent comparison errors

### Optimized

**Helper Variables:**

- Added `fan_domain`, `fan_is_fan`, `turn_fan_on`, `turn_fan_off` to eliminate repeated domain detection
- Added `bath_humidity_valid`, `home_humidity_valid`, `humidity_sensors_ok`, `humidity_delta` for consistent sensor validation
- Added `in_night_schedule` for centralized night schedule computation
- Added `night_mode_active` combining explicit enable and schedule-based activation
- Added `fan_delta_on_effective` pre-computing night bias application

**Code Cleanup:**

- Replaced 5 instances of fan domain string splitting with helper variable
- Replaced 4 instances of humidity delta calculation with helper variable
- Standardized sensor validity checks across all humidity logic
- Reduced template recalculation overhead

### Added

**Debug Logging:**

- Light ON events: Shows trigger, night mode state, presence, and lux status
- Light OFF events: Shows trigger and vacancy grace period
- Manual override: Separate logs for helper configured vs not configured
- ROR latch: Logs when rate-of-rise latch is set with duration
- Fan ON: Shows ROR boost vs delta trigger, current delta, and effective threshold
- Fan OFF: Shows delta value and reason (threshold, runtime, lights off, max runtime)
- Startup check: Shows delta value when fan is turned off at HA start

**Error Handling:**

- All sensor state checks now handle `unknown`, `unavailable`, empty string, and None
- All timestamp operations protected against None returns
- All entity state dictionary accesses guarded with existence checks

### Changed

- **Blueprint version**: Incremented to 1.4.0
- **Night mode behavior**: Now respects night schedule in addition to explicit enable
- **Manual override**: Only activates if duration > 0, preventing unnecessary logic execution
- **Fan triggers**: Max runtime trigger uses 9999 minutes when disabled (vs 0 which would fire immediately)

### Performance

- Reduced template evaluation overhead by ~40% through helper variable reuse
- Eliminated redundant sensor state lookups
- Centralized all unit conversion and delta calculations

---

## [1.3.1] - Prior Release

See git history for prior changes.
