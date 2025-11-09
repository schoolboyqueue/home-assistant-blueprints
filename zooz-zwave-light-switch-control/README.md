# Zooz Z-Wave Light Switch Control Pro

**Version:** 0.1.0  
**Author:** Jeremy Carter  
**Home Assistant Blueprint for Z-Wave Switch Light Dimming**

[![Import Blueprint](https://my.home-assistant.io/badges/blueprint_import.svg)](https://my.home-assistant.io/redirect/blueprint_import/?blueprint_url=https%3A%2F%2Fgithub.com%2Fschoolboyqueue%2Fhome-assistant-blueprints%2Fblob%2Fmain%2Fzooz-zwave-light-switch-control%2Fzooz_zwave_light_switch_control_pro.yaml)

---

## Overview

Zooz Z-Wave Light Switch Control Pro transforms your Zooz Z-Wave switches into intelligent light dimmers using Central Scene events. Combines single-press on/off with hold-to-dim control inspired by Lutron Pico dimming logic.

Supports ZEN71, ZEN72, ZEN76, ZEN77 switches (both 700 and 800LR series) via Z-Wave JS integration.

### Key Features

#### Basic Control
- 💡 **Single Press Up** — Light turns on (restores last brightness)
- ⚫ **Single Press Down** — Light turns off
- ⬆️ **Hold Up** — Brightness increases smoothly while held
- ⬇️ **Hold Down** — Brightness decreases; turns off at minimum threshold
- 🛑 **Release Detection** — Dimming stops immediately on paddle release

#### Advanced Features
- 🎯 **Optional Double/Triple-Tap Actions** — Custom actions for 2x/3x Up/Down presses
- 🏠 **Area Targeting** — Control entire room instead of single light
- ⚙️ **Configurable Dimming** — Adjustable step size, interval, and thresholds
- 🐛 **Debug Logging** — Off / Basic / Verbose logging for troubleshooting
- 🔌 **Dual Event Support** — Works with both `zwave_js_event` and `zwave_js_value_notification`

---

## Table of Contents

- [Quick Start](#quick-start)
- [How It Works](#how-it-works)
  - [Central Scene Events](#central-scene-events)
  - [Hold-to-Dim Logic](#hold-to-dim-logic)
  - [Brightness Clamping](#brightness-clamping)
- [Configuration Guide](#configuration-guide)
  - [Device and Light](#device-and-light)
  - [Dimming Parameters](#dimming-parameters)
  - [Optional Gesture Actions](#optional-gesture-actions)
  - [Diagnostics](#diagnostics)
- [Use Cases](#use-cases)
- [Troubleshooting](#troubleshooting)
- [Advanced Topics](#advanced-topics)

---

## Quick Start

### Prerequisites

- Home Assistant 2023.4 or newer
- Z-Wave JS integration configured
- Zooz Z-Wave switch (ZEN71/72/76/77; 700 or 800LR series)
- Central Scene / Scene Control enabled on device (parameter 1 on most models)
- Light entity or area to control

### Installation

1. **Import the blueprint:**
   - Click the badge above, or
   - Navigate to: Settings → Automations & Scenes → Blueprints → Import Blueprint
   - Paste URL: `https://github.com/schoolboyqueue/home-assistant-blueprints/blob/main/zooz-zwave-light-switch-control/zooz_zwave_light_switch_control_pro.yaml`

2. **Create an automation:**
   - Go to: Settings → Automations & Scenes
   - Click "+ Create Automation" → "Use Blueprint"
   - Select "Zooz Z-Wave Light Switch Control Pro"

3. **Minimum required configuration:**
   - **Zooz switch device:** Your Z-Wave switch
   - **Light entity:** Light to control

4. **Save and test!**

---

## How It Works

### Central Scene Events

Zooz switches send **Central Scene** events to Home Assistant via Z-Wave JS:

| Event | Scene | Value | Action |
|-------|-------|-------|--------|
| Single press Up | 001 | KeyPressed | Light ON |
| Single press Down | 002 | KeyPressed | Light OFF |
| Hold Up | 001 | KeyHeldDown → KeyReleased | Brighten loop |
| Hold Down | 002 | KeyHeldDown → KeyReleased | Dim loop |
| Double-tap Up | 001 | KeyPressed2x | Custom action |
| Double-tap Down | 002 | KeyPressed2x | Custom action |
| Triple-tap Up | 001 | KeyPressed3x | Custom action |
| Triple-tap Down | 002 | KeyPressed3x | Custom action |

**Dual event type support:**

The blueprint listens to both `zwave_js_event` and `zwave_js_value_notification` for maximum compatibility across Z-Wave JS versions and device firmware.

### Hold-to-Dim Logic

**Hold Up sequence:**

1. KeyHeldDown event received
2. If light is off → turn on to minimum brightness (default: 10%)
3. Start repeat loop:
   - Wait for KeyReleased event (timeout: dim interval, default 200ms)
   - If KeyReleased received → stop dimming
   - If timeout → increase brightness by step (default: 5%)
   - Repeat

**Hold Down sequence:**

1. KeyHeldDown event received
2. If light is off → do nothing (avoid flicker)
3. If light is on → start repeat loop:
   - Wait for KeyReleased event (timeout: dim interval)
   - If KeyReleased received → stop dimming
   - If timeout → decrease brightness by step
   - If brightness ≤ minimum clamp (default: 1%) → turn light off completely
   - Repeat

**Why `wait_for_trigger` with timeout?**

This approach ensures:
- Immediate response to KeyReleased (no polling delay)
- Smooth dimming at configurable interval
- Robust handling of missed or delayed events

### Brightness Clamping

**Minimum clamp (default: 1%):**

When dimming down, if brightness reaches or falls below the minimum clamp threshold, the light turns off completely. This prevents "stuck at 1%" behavior.

**Minimum on brightness (default: 10%):**

When holding Up while light is off, it turns on to at least this brightness before starting to brighten. Prevents starting at imperceptibly dim levels.

**Maximum clamp (default: 100%):**

Logical cap for brightness. Used for information/logging. Home Assistant's `brightness_step_pct` naturally stops at 100%.

---

## Configuration Guide

### Device and Light

**Zooz switch device** (required)  
Your Z-Wave JS device. Blueprint filters for Zooz manufacturer and Z-Wave JS integration.

**Supported models:**
- ZEN71 (On/Off Switch, 700 & 800LR)
- ZEN72 (Dimmer Switch, 700 & 800LR)
- ZEN76 (On/Off Paddle Switch, 700 & 800LR)
- ZEN77 (Dimmer Paddle Switch, 700 & 800LR)

**Light entity** (required)  
The primary light to control. Used for:
- Reading current brightness during hold-to-dim
- Default target for on/off/dim commands (when area not specified)

**Light area** (optional)  
If set, all light commands target this area instead of the single entity. Brightness checks still read from the single light entity.

**Use case:** Bathroom with 3 ceiling lights — select one as primary entity (for brightness checks), select "Bathroom" as area (all 3 lights respond to switch).

---

### Dimming Parameters

**Brightness step (%)** (default: 5%)  
Amount to change brightness per step while holding. Smaller = smoother but slower; larger = faster but chunkier.

**Dim interval (ms)** (default: 200ms)  
Delay between each brightness step. Smaller = faster dimming; larger = slower dimming.

**Combined effect:**
- 5% step @ 200ms = 25%/sec change rate (fast)
- 2% step @ 100ms = 20%/sec change rate (smooth)
- 10% step @ 300ms = 33%/sec change rate (aggressive)

**Minimum brightness on hold-up when off (%)** (default: 10%)  
If light is off and you hold Up, it turns on to at least this brightness before starting to brighten.

**Clamp minimum brightness (%)** (default: 1%)  
If brightness reaches this level or lower during dim-down, light turns off completely.

**Clamp maximum brightness (%)** (default: 100%)  
Logical cap for max brightness. Used for verbose logging. Home Assistant's native brightness controls naturally stop at 100%.

---

### Optional Gesture Actions

**Double-tap Up/Down actions** (optional)  
Custom action sequences to run when Up/Down paddle is double-pressed.

**Triple-tap Up/Down actions** (optional)  
Custom action sequences to run when Up/Down paddle is triple-pressed.

**Examples:**
- Double-tap Up → Set scene "Movie Time"
- Double-tap Down → Turn off all lights in house
- Triple-tap Up → Set brightness to 100%
- Triple-tap Down → Activate "Good Night" routine

---

### Diagnostics

**Debug level** (default: off)  
Controls logging verbosity:

- **off:** No debug output
- **basic:** Key transitions (press/hold/release, light on/off)
- **verbose:** Detailed brightness levels, clamp checks, step calculations

**Debug logs appear in:**
- Settings → System → Logs
- Look for `[Zooz Light Pro]` prefix

---

## Use Cases

### Typical Dimming Sequence

1. **Single press Up**  
   → Light turns on at last brightness (e.g., 70%)

2. **Hold Down**  
   → Brightness decreases: 70% → 65% → 60% → ... → 5% → OFF
   → Release at any point to stop

3. **Single press Up again**  
   → Light turns on at last brightness (5% before it turned off)

4. **Hold Up**  
   → If off: turns on at 10%, then brightens
   → If on: immediately starts brightening: 5% → 10% → 15% → ...
   → Release at 50%

5. **Single press Down**  
   → Light turns off (remembers 50% for next time)

### Area Control

**Setup:**
- 3 recessed lights in kitchen (light.kitchen_1, light.kitchen_2, light.kitchen_3)
- 1 Zooz ZEN77 switch at entrance

**Configuration:**
- Light entity: `light.kitchen_1` (primary, used for brightness checks)
- Light area: `Kitchen`

**Behavior:**
- Single press Up → All 3 kitchen lights turn on
- Hold Down → All 3 lights dim together
- Blueprint reads brightness from `light.kitchen_1` to determine when to stop/turn off

### Double-Tap Scenes

**Configuration:**
- Double-tap Up action: Set scene "Cooking" (bright white, 100%)
- Double-tap Down action: Set scene "Dinner" (warm white, 30%)

**Behavior:**
- Double-tap Up → Instant switch to bright cooking lighting
- Double-tap Down → Instant switch to dim dinner ambiance
- Single press Down → Off

---

## Troubleshooting

### No events received / Switch doesn't trigger automation

**Cause:** Central Scene not enabled or not supported  
**Solutions:**
1. Verify device supports Central Scene (all ZEN71/72/76/77 models do)
2. Check device parameter 1 ("Scene Control" or "Paddle Control"):
   - Navigate to: Settings → Devices & Services → Z-Wave JS → [Your Switch]
   - Look for parameter 1
   - Set to "Scene control enabled" or equivalent
   - Some models require firmware update for Central Scene support
3. Check Z-Wave JS logs: Settings → System → Logs → Filter "zwave"
4. Test with Developer Tools → Events → Listen to `zwave_js_event` and press switch

### Dimming is too fast / too slow

**Too fast:**
- Increase `dim_interval_ms` (e.g., 200 → 300ms)
- Decrease `brightness_step_pct` (e.g., 5% → 3%)

**Too slow:**
- Decrease `dim_interval_ms` (e.g., 200 → 150ms)
- Increase `brightness_step_pct` (e.g., 5% → 7%)

### Dimming doesn't stop when I release paddle

**Likely cause:** KeyReleased events not being sent  
**Solutions:**
1. Check firmware version (older firmware may not send KeyReleased)
2. Update Z-Wave JS to latest version
3. Re-interview device in Z-Wave JS
4. Some non-Zooz devices claiming compatibility may not send KeyReleased properly

### Light ignores brightness_step_pct commands

**Cause:** Some light integrations/drivers don't implement `brightness_step_pct`  
**Solutions:**
1. Try smaller steps with longer intervals (test if stepping works at all)
2. Check light integration documentation for brightness control support
3. If light completely ignores stepping, open GitHub issue for fallback implementation

**Known problematic integrations:**
- Some Zigbee bulbs via ZHA (varies by manufacturer)
- DMX lighting controllers (often lack relative dimming)
- Some IR-controlled LED strips

### Lights turn off unexpectedly during dim-down

**Cause:** Brightness hitting minimum clamp threshold  
**Solution:**  
Lower `clamp_min_pct` (e.g., 1% → 0.5% or 0%). This allows dimming to lower levels before turning off.

### Hold Up when light is off doesn't brighten

**Check:**
1. Verify light actually turns on (check state in Developer Tools)
2. If light turns on but doesn't brighten, increase `min_on_brightness_pct` to ensure visible starting point
3. Enable verbose debug to see brightness step commands being sent

### Area targeting not working

**Check:**
1. Verify area name is correct (case-sensitive)
2. Verify lights are assigned to area: Settings → Areas → [Area] → Check entities
3. Enable basic debug to confirm `area_set` variable is true
4. Test manually: Developer Tools → Services → `light.turn_on` → Target: area

---

## Advanced Topics

### Mode: Restart

The blueprint uses `mode: restart`. When a new trigger fires, the automation restarts from the beginning.

**Why this matters:**

If you start holding Up → release → immediately hold Up again, the second hold event restarts the automation, canceling any cleanup from the first hold. This is desired behavior (immediate response).

**Edge case:** Rapidly pressing Up+Down might cause race conditions. The restart mode handles this gracefully by always processing the latest event.

### Variable Binding Pattern

```yaml
# All !input tags bound to variables first
step_pct_in: !input brightness_step_pct

# Then used in templates
step_pct: "{{ step_pct_in | float(5) }}"
```

**Why?**

Home Assistant's YAML parser has limitations with `!input` inside Jinja2 templates. Binding to intermediate variables first avoids parsing ambiguity.

### Brightness Absolute vs. Relative

The blueprint uses `brightness_step_pct` (relative stepping):

```yaml
service: light.turn_on
data:
  brightness_step_pct: 5  # Increase by 5%
```

**Advantages:**
- Works regardless of current brightness
- Smooth dimming curves
- No need to track state

**Alternative (not implemented):**

Absolute brightness with manual state tracking:

```yaml
current: "{{ state_attr(light, 'brightness') | int }}"
new: "{{ (current + 12) | int }}"
service: light.turn_on
data:
  brightness: "{{ new }}"
```

More complex but allows exact control. Could be added as fallback for lights that don't support `brightness_step_pct`.

### Debug Logging Examples

**Basic logging output:**

```
[Zooz Light Pro] up_single → light.turn_on (restore last)
[Zooz Light Pro] up_hold_start
[Zooz Light Pro] Light is off → turn_on to min_on_brightness_pct=10%
[Zooz Light Pro] up_release → stop brighten
[Zooz Light Pro] down_single → light.turn_off
```

**Verbose logging output:**

```
[Zooz Light Pro] up_hold_start
[Zooz Light Pro] Light is off → turn_on to min_on_brightness_pct=10%
[Zooz Light Pro] brighten step +5% | current=25/255 cap=255
[Zooz Light Pro] brighten step +5% | current=38/255 cap=255
[Zooz Light Pro] brighten step +5% | current=51/255 cap=255
[Zooz Light Pro] up_release → stop brighten
```

### Performance Considerations

**Trigger efficiency:**

Event-based triggers (no polling):
- KeyPressed, KeyHeldDown, KeyReleased events
- Each trigger identified by `id` and `trigger.id` comparison
- No redundant state checks

**Template optimization:**

Variables computed once:
- `min_abs_brightness` = `(clamp_min_pct_v / 100.0 * 255) | round(0) | int`
- Used multiple times in conditions without recalculation

**Mode restart overhead:**

Minimal. Restarting cancels `wait_for_trigger` but doesn't leave orphaned processes.

---

## Contributing

Issues, feature requests, and pull requests welcome at:  
https://github.com/schoolboyqueue/home-assistant-blueprints

---

## Changelog

See [CHANGELOG.md](CHANGELOG.md) for version history.

---

## License

MIT License — See repository for details.

---

## Credits

- Zooz Z-Wave Central Scene event mapping adapted for ZEN71/72/76/77 (700/800 series)
- Hold-to-dim behavior inspired by [SmartQasa's Lutron Pico 5-button blueprint](https://github.com/smartqasa/blueprints/blob/main/lutron_pico_5_light.yaml)

---

**Enjoy intelligent switch control! 💡🎛️**

## Event Mapping

- Up single: KeyPressed on scene 001
- Down single: KeyPressed on scene 002
- Up hold start: KeyHeldDown on scene 001; stop on KeyReleased
- Down hold start: KeyHeldDown on scene 002; stop on KeyReleased
- Up/Down double: KeyPressed2x
- Up/Down triple: KeyPressed3x

Both `zwave_js_event` and `zwave_js_value_notification` event types are supported.

## Notes and Behavior

- When holding Up and the light is off, the blueprint turns it on to at least the configured minimum-on brightness.
- When holding Down, the blueprint will not turn the light on if it is off (avoids flicker). If the light is on and brightness reaches the minimum clamp during hold, it turns off.
- Brightness stepping uses `light.turn_on` with `brightness_step_pct`. Some light integrations may not fully support stepping; if you observe inconsistent dimming, reduce the step size and/or increase the interval. If the light completely ignores `brightness_step_pct`, consider opening an issue to add absolute brightness stepping as a fallback.

## Troubleshooting

- No events received:
  - Confirm the device is paired via Z-Wave JS and supports Central Scene.
  - Ensure "Scene Control" is enabled in device parameters (model-specific).
  - Check Home Assistant logs (Settings → System → Logs) and enable `debug` level here if needed.
- Dimming is too fast/slow:
  - Decrease/increase `brightness_step_pct` or adjust `dim_interval_ms`.
- Lights ignore `brightness_step_pct`:
  - Some integrations/drivers don't implement step properly. Try smaller steps/longer intervals. If still problematic, open an issue to request adding an absolute 0–255 stepping fallback.
- Area vs. Entity:
  - If you set an Area, actions target the area. Brightness checks still use the single selected light entity.

## Credits

- Zooz Z-Wave Central Scene scenes adapted for ZEN71/72/76/77 700/800 series.
- Hold-to-dim behavior inspired by SmartQasa's Lutron Pico 5-button blueprint.
