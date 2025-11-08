# Bathroom Light & Fan Control Pro - Changelog

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