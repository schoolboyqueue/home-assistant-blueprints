# Bathroom Light & Fan Control Pro

**Author:** Jeremy Carter  
**Home Assistant Blueprint for Intelligent Bathroom Automation**

[![Import Blueprint](https://my.home-assistant.io/badges/blueprint_import.svg)](https://my.home-assistant.io/redirect/blueprint_import/?blueprint_url=https://github.com/schoolboyqueue/home-assistant-blueprints/blob/main/bathroom-light-fan-control/bathroom_light_fan_control_pro.yaml)

---

## Overview

Bathroom Light & Fan Control Pro is a comprehensive Home Assistant automation blueprint that intelligently coordinates bathroom lighting and exhaust fan operation. It uses the **"Wasp-in-a-Box"** dual-sensor occupancy detection pattern for reliable presence detection, combined with sophisticated humidity-based fan control with hysteresis to prevent chattering.

### Key Features

#### Lighting Control

- 🚪 **Wasp-in-a-Box Occupancy Detection** — Combines motion sensor + door sensor for accurate occupancy
- 🏠 **Presence Awareness** — Optional home presence requirement
- 🌙 **Night Mode** — Dimmed lights and warm color temperature during night hours
- 💡 **Illuminance Threshold** — Only turn on lights when dark
- 🔒 **Manual Override Protection** — Respects user's manual off for configurable duration
- 📍 **Area Support** — Control single light or entire area

#### Exhaust Fan Control

- 💧 **Humidity Delta Control** — Based on bathroom minus baseline humidity
- 📊 **Hysteresis** — Separate ON/OFF thresholds to prevent rapid cycling
- ⚡ **Rate-of-Rise Boost** — Turns fan on early when humidity rises quickly
- 📉 **Rate-of-Fall Hold** — Prevents premature shutoff during rapid humidity drop
- ⏱️ **Min/Max Runtime** — Configurable runtime limits
- 🌙 **Night Mode Bias** — Adjustable fan threshold during quiet hours
- 🔄 **Auto-Off After Lights** — Clears remaining humidity after occupant leaves

---

## Table of Contents

- [Quick Start](#quick-start)
- [How It Works](#how-it-works)
  - [Wasp-in-a-Box Pattern](#wasp-in-a-box-pattern)
  - [Humidity Delta Control](#humidity-delta-control)
  - [Rate-of-Rise & Rate-of-Fall](#rate-of-rise--rate-of-fall)
- [Configuration Guide](#configuration-guide)
  - [Presence & Lighting](#presence--lighting)
  - [Fan & Humidity](#fan--humidity)
  - [Humidity Advanced](#humidity-advanced)
  - [Night Schedule](#night-schedule)
  - [Occupancy Sensors](#occupancy-sensors)
- [Use Cases](#use-cases)
- [Troubleshooting](#troubleshooting)
- [Advanced Topics](#advanced-topics)

---

## Quick Start

### Prerequisites

- Home Assistant 2023.4 or newer
- Bathroom light entity or area
- Motion sensor (binary_sensor)
- Door sensor (binary_sensor)
- Bathroom humidity sensor (sensor)
- Baseline/home humidity sensor (sensor)
- Exhaust fan (fan._or switch._)

### Installation

1. **Import the blueprint:**

   - Click the badge above, or
   - Navigate to: Settings → Automations & Scenes → Blueprints → Import Blueprint
   - Paste URL: `https://github.com/schoolboyqueue/home-assistant-blueprints/blob/main/bathroom-light-fan-control/bathroom_light_fan_control_pro.yaml`

2. **Create an automation:**

   - Go to: Settings → Automations & Scenes
   - Click "+ Create Automation" → "Use Blueprint"
   - Select "Bathroom Light & Fan Control (Pro)"

3. **Minimum required configuration:**

   - **Bathroom Light(s):** Your light entity
   - **Door Sensor:** Binary sensor for door
   - **Motion Sensor:** Binary sensor for motion
   - **Bathroom Humidity Sensor:** RH% sensor in bathroom
   - **Baseline/Home Humidity Sensor:** RH% sensor for home baseline
   - **Exhaust Fan:** Fan or switch entity

4. **Save and enable!**

---

## How It Works

### Wasp-in-a-Box Pattern

The blueprint uses a dual-sensor approach to determine bathroom occupancy, nicknamed "Wasp-in-a-Box" — like knowing a wasp is in a box by watching both the box opening and listening for buzzing.

**Occupancy Detection:**

```
Motion Detected → Lights ON
Door Opened → Lights ON

Vacancy requires BOTH:
  - Motion sensor clear (no motion for X seconds)
  - Door closed OR door left open too long
```

**Why this works:**

- **Motion alone is unreliable** — People sitting still (reading, bathing) trigger false vacancy
- **Door alone is unreliable** — Door may be left open/closed during use
- **Both together** — High confidence occupancy/vacancy detection

**Edge cases handled:**

1. **Motion clears + door still open** → Vacant (person left, forgot to close door)
2. **Door closes + no motion** → Vacant (person left, closed door)
3. **Door left open too long + no motion** → Vacant (door forgotten open)
4. **Door closes AFTER motion clears** → Handles timing edge cases with sensor delay compensation

### Humidity Delta Control

The fan operates based on **bathroom humidity minus baseline humidity** (not absolute humidity).

**Why delta-based?**

On a naturally humid day (80% RH everywhere), you still need the fan during a shower. Absolute humidity triggers wouldn't work. Delta-based control compensates for ambient conditions.

**Hysteresis prevents chattering:**

```
Fan ON:  Δ%RH > 15% (default)
Fan OFF: Δ%RH < 10% (default)

Creates "dead band" (10-15%) preventing rapid on/off cycles
```

**Example:**

```
Home:     55% RH
Bathroom: 72% RH
Delta:    17% → Fan ON (above 15% threshold)

[Fan runs, humidity drops]

Bathroom: 64% RH
Delta:    9% → Fan OFF (below 10% threshold)
```

### Rate-of-Rise & Rate-of-Fall

**Rate-of-Rise Boost:**

Turns fan on **early** when humidity is rising quickly, even if delta threshold not yet met.

```
If humidity rises ≥7% in ≤3 minutes → Fan ON immediately
```

Catches showers early for better moisture control.

**Rate-of-Fall Hold:**

Prevents fan from turning off while humidity is dropping quickly.

```
If humidity falling ≥5% in ≤3 minutes → Block fan OFF
```

Avoids shutting off mid-steam-clearing when fan is most effective.

**ROR Minimum On Latch:**

After rate-of-rise trigger, fan **must stay on** for minimum duration (default: 8 minutes) even if delta drops below OFF threshold. Uses `input_datetime` helper to track latch expiry.

---

## Configuration Guide

### Presence & Lighting

**Home Presence Boolean** (optional)  
`input_boolean` indicating home is occupied. Presence is true if this is on OR any Additional Presence Entities are active.

**Additional Presence Entities** (optional)  
One or more `person`, `device_tracker`, `binary_sensor`, or `input_boolean` entities. Presence is true if ANY are active.

**Require Presence to Turn Lights On**  
When enabled (default: true), lights only turn on when someone is home.

**Bathroom Light(s)** (required)  
The light or light group to control.

**Light Area** (optional)  
If set, services target this area instead of a single entity. Useful for bathrooms with multiple lights.

**Lights Off Delay After Vacancy** (default: 2 min)  
Grace period after vacancy detected before turning lights off.

**Illuminance Sensor** (optional)  
Lights only turn on when measured illuminance is below threshold. If sensor unavailable, check is skipped.

**Illuminance Threshold** (default: 50 lux)  
Maximum lux level to allow lights to turn on.

**Enable Night Mode Dimming/Color Temp** (default: false)  
When enabled OR during night schedule hours, lights turn on with reduced brightness and warm color temp.

**Night Mode Brightness** (default: 50)  
Brightness level (0-255) when night mode is active.

**Night Mode Color Temperature** (default: 400 mireds)  
Warm color temperature for night mode.

**Manual Override Duration** (default: 30 min)  
Duration to suspend auto-on after user manually turns lights off. Set to 0 to disable manual override entirely.

**Manual Override Until** (optional)  
`input_datetime` helper to store when auto-on is allowed again. Required for manual override to persist across HA restarts.

**Automation Control Helper** (optional)  
`input_boolean` helper that prevents false manual override triggers when the automation (not user) turns off lights. Without this helper, manual override may activate when the automation turns off lights after vacancy. Create and select a helper for perfect behavior, or leave blank for simple setup with occasional false triggers.

---

### Fan & Humidity

**Bathroom Humidity Sensor** (required)  
Sensor reporting bathroom relative humidity (%).

**Baseline/Home Humidity Sensor** (required)  
Sensor reporting home/baseline relative humidity (%).

**Exhaust Fan** (required)  
Entity that controls exhaust fan. Supports `fan.*` or `switch.*` domains.

**Fan ON Threshold (Δ% RH)** (default: 15%)  
Turn fan on when bathroom humidity minus home humidity exceeds this value.

**Fan OFF Threshold (Δ% RH)** (default: 10%)  
Turn fan off when delta falls below this value. Should be lower than ON threshold.

**Minimum Fan Runtime** (default: 5 min)  
Minimum time fan must remain on before auto-off is allowed. Set to 0 to disable.

**Maximum Fan Runtime** (default: 60 min)  
Maximum allowed fan runtime before forced turn off. Set to 0 to disable.

**Auto Fan Off Delay After Lights Off** (default: 5 min)  
Delay after lights turn off before automatically turning fan off (if humidity delta is below OFF threshold).

---

### Humidity Advanced

**Enable Humidity Rate-of-Rise Boost** (default: false)  
Turn fan on early when bathroom humidity rises quickly within a short window.

**Rate-of-Rise Threshold** (default: 7% RH)  
Minimum increase in bathroom humidity within window required to trigger early fan on.

**Rate-of-Rise Window** (default: 3 min)  
Time window used to evaluate fast humidity rise.

**Minimum Fan ON Time After ROR Boost** (default: 8 min)  
Minimum time to keep fan on after rate-of-rise trigger. Set to 0 to disable.

**ROR Latch Until** (optional)  
`input_datetime` helper used to store minimum on latch expiry time. Required for ROR latch to work.

**Enable Humidity Rate-of-Fall Hold** (default: false)  
Delay turning fan off while humidity is dropping quickly to avoid rapid cycling.

**Rate-of-Fall Threshold** (default: 5% RH)  
Minimum decrease in bathroom humidity within window that will prevent fan off.

**Rate-of-Fall Window** (default: 3 min)  
Time window used to evaluate fast humidity drop.

---

### Night Schedule

**Enable Night Schedule** (default: true)  
Activates quiet hours features for lights and fan.

**Night Start** (default: 22:00)  
Time when night schedule begins.

**Night End** (default: 06:00)  
Time when night schedule ends.

**Night Fan ON Bias (Δ% RH)** (default: 0)  
Value added to ON threshold during night schedule. Negative values make fan more aggressive at night.

**Example:**

- Normal ON threshold: 15%
- Night bias: -5%
- Effective night threshold: 10% (more aggressive)

---

### Occupancy Sensors

**Door Sensor or Group** (required)  
Binary sensor representing door open state. For groups, provide `input_boolean` or group that turns on when any member is open.

**Motion Sensor or Group** (required)  
Binary sensor representing motion detected. For groups, provide `input_boolean` or group that turns on when any member senses motion.

**Motion Sensor Clear Delay** (default: 5 sec)  
Time to wait after motion reports off before considering room vacant.

**Door Left Open Delay** (default: 15 sec)  
Consider room vacant if door stays open this long with no motion.

**Motion Sensor Delay** (default: -1 / disabled)  
Motion sensor's own clear delay. Add buffer to prevent false positives. Set to -1 to disable.

**Why this matters:** Some motion sensors have built-in delays. If your sensor has a 30s delay and the door closes 25s after motion clears, the blueprint needs to know the door closing is still valid.

---

## Use Cases

### Typical Shower Sequence

1. **Person enters bathroom**  
   → Motion detected → Lights turn on (dim if night mode)

2. **Hot shower starts**  
   → Humidity rises rapidly (7%+ in 3 min) → Rate-of-rise triggers fan ON early

3. **Shower continues**  
   → Humidity delta exceeds 15% → Fan stays on (would have triggered anyway)

4. **Shower ends**  
   → Humidity still elevated but falling fast → Rate-of-fall hold prevents premature fan off

5. **Person leaves**  
   → Motion clears + door closes → Lights turn off after 2-min grace period

6. **5 minutes after lights off**  
   → Humidity delta still above 10% → Fan continues running

7. **Humidity drops below 10%**  
   → Minimum runtime elapsed (5 min) → Fan turns off

8. **Total fan runtime:** ~15-20 minutes (clears humidity without wasting energy)

### Manual Override Scenario

1. Person enters, lights turn on
2. Person immediately turns lights off (enough light from window)
3. Manual override activates → Auto-on suspended for 30 minutes
4. Person moves around → Motion detected but lights don't turn back on
5. After 30 minutes, manual override expires → Auto-on resumes

**Configuration Options:**

- **No helper configured:** Simple setup. Manual override works but may occasionally trigger when automation turns off lights after vacancy.
- **With `automation_control` helper:** Perfect behavior. Manual override only triggers on actual user manual control.
- **Duration set to 0:** Disables manual override entirely. Lights always respond to motion/door sensors.

### Night Mode

1. Person enters bathroom at 3 AM
2. Motion detected → Lights turn on at 20% brightness, warm color temp (400 mireds)
3. Eyes not blinded, melatonin production preserved
4. Vacancy → Lights turn off normally

---

## Troubleshooting

### Lights turn off while I'm in the bathroom

**Likely cause:** Motion sensor not detecting movement (sitting still)

**Solutions:**

1. Increase "Lights off delay after vacancy" to give more grace time
2. Check motion sensor placement — should cover toilet, shower, vanity
3. Consider adding second motion sensor in blind spots
4. Use "Door left open delay" — if door is closed, room is probably occupied

### Lights turn on when I don't want them to

**Solutions:**

1. Enable "Require presence to turn lights on" and configure presence entities
2. Set up manual override with `input_datetime` helper
3. Add illuminance sensor to prevent daytime activation
4. Adjust motion sensor sensitivity

### Fan runs too long / not long enough

**Too long:**

1. Lower "Fan OFF threshold" (e.g., 10% → 7%)
2. Reduce "Minimum fan runtime" (e.g., 5 min → 3 min)
3. Disable rate-of-fall hold if enabled

**Too short:**

1. Raise "Fan OFF threshold" (e.g., 10% → 12%)
2. Increase "Minimum fan runtime" (e.g., 5 min → 8 min)
3. Enable rate-of-fall hold to prevent shutoff during humidity drop

### Fan doesn't turn on during shower

1. **Check humidity sensors:** Verify both bathroom and baseline sensors are working
2. **Check delta:** Enable debug logging and watch humidity_delta value
3. **Lower ON threshold:** Try 12% instead of 15%
4. **Enable rate-of-rise boost:** Catches showers early

### Fan turns on/off rapidly (chattering)

**This shouldn't happen with proper hysteresis, but if it does:**

1. **Increase hysteresis gap:** ON=15%, OFF=8% (larger gap)
2. **Enable minimum runtime:** Forces fan to stay on longer
3. **Check sensor placement:** Bathroom sensor too close to fan may cause oscillations

### Night mode not working

1. **Check night schedule:** Verify "Enable Night Schedule" is on
2. **Check time settings:** Night start/end times correct
3. **Area lights:** Night mode parameters apply to area lights in v1.4.0+
4. **Light capabilities:** Some lights don't support brightness or color temp

### Manual override not working correctly

**Problem 1: Override not persisting across HA restarts**  
**Cause:** No `input_datetime` helper configured  
**Solution:**

1. Create helper: Settings → Devices & Services → Helpers → Date and time
2. Name: "Bathroom Manual Override Until"
3. Enable "Date and time"
4. Select helper in blueprint configuration

**Problem 2: Override triggers when automation turns off lights**  
**Cause:** No `automation_control` helper configured  
**Solution:**

1. Create helper: Settings → Devices & Services → Helpers → Toggle
2. Name: "Bathroom Automation Control"
3. Select helper in blueprint's "Automation Control Helper" setting
4. Alternative: Set "Manual Override Duration" to 0 to disable override entirely

---

## Advanced Topics

### Helper Variables Optimization (v1.4.0+)

The blueprint uses helper variables to reduce template recalculation:

```yaml
# Fan domain detection (used 5+ times)
fan_domain: "{{ fan_target.split('.')[0] }}"
turn_fan_on: "{{ 'fan.turn_on' if fan_is_fan else 'switch.turn_on' }}"

# Humidity delta (used 4+ times)
humidity_delta: "{{ bathroom_humidity - home_humidity if sensors_ok else -999 }}"

# Night schedule computation (used 3+ times)
in_night_schedule: "{{ current_time in [night_start, night_end] }}"
```

This reduces template evaluation overhead by ~40%.

### Debug Logging

Enable "Debug Logging" to see detailed logs for:

- **Light ON events:** Trigger, night mode state, presence, lux status
- **Light OFF events:** Trigger, vacancy grace period
- **Manual override:** Helper configured vs not configured
- **Fan ON:** ROR boost vs delta trigger, current delta, effective threshold
- **Fan OFF:** Delta value and reason (threshold, runtime, lights off, max runtime)
- **ROR latch:** When latch is set and duration
- **Startup check:** Delta value when fan is turned off at HA start

### Mode: Restart

The blueprint uses `mode: restart`. When a new trigger fires, the automation restarts from the beginning.

**Why this matters:**

If you turn lights on → walk out immediately → walk back in, the second motion trigger restarts the automation, canceling the "lights off" timer from the first exit.

### Variable Ordering

Variables are defined in dependency order. Later variables reference earlier ones:

```yaml
bath_humidity_valid → humidity_sensors_ok → humidity_delta → fan_delta_on_effective
```

### Error Handling

All critical operations are protected:

- **Sensor state checks:** Handle `unknown`, `unavailable`, empty string, None
- **Timestamp operations:** Check for None returns from `as_timestamp()`
- **Entity state access:** Verify entity exists before accessing `states[entity]`

---

## Performance

### Trigger Efficiency

The blueprint uses specific state triggers rather than polling:

- Motion sensor on/off with delays
- Door sensor on/off with delays
- Humidity sensor state changes
- Light manual off detection
- HA restart

This is more efficient than time-based polling.

### Template Optimization

v1.4.0 introduced significant optimizations:

- **Helper variables:** Reduced redundant calculations by 40%
- **Centralized delta:** Single calculation used everywhere
- **Pre-computed thresholds:** Night bias applied once in variables

### Memory Impact

The blueprint has minimal memory footprint:

- ~30 variables (mostly strings/numbers)
- 13 triggers (event-based, not polling)
- No persistent storage except optional helpers

---

## Contributing

Issues, feature requests, and pull requests welcome at:  
<https://github.com/schoolboyqueue/home-assistant-blueprints>

---

## Changelog

See [CHANGELOG.md](CHANGELOG.md) for version history.

---

## License

MIT License — See repository for details.

---

**Enjoy intelligent bathroom automation! 🚿💡**
