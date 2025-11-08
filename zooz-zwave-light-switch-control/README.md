# Zooz Z-Wave Light Switch Control Pro

Version: 0.1.0  
Author: Jeremy Carter

A Home Assistant blueprint that combines Zooz Z-Wave Central Scene presses with hold-to-dim control inspired by the Lutron Pico dimming logic. Works with ZEN71/72/76/77 (700/800 series) via Z-Wave JS.

## Features

- Single press Up → light turns on (restores last brightness)
- Single press Down → light turns off
- Hold Up → brightness increases repeatedly while held; stops on release
- Hold Down → brightness decreases repeatedly while held; stops on release; turns off when minimum threshold is reached
- Optional custom actions for double-tap and triple-tap (Up/Down)
- Optional area targeting (actions apply to area instead of single entity)
- Debug logging (off / basic / verbose)

## Requirements

- Home Assistant with the Z-Wave JS integration
- Zooz Z-Wave switch supporting Central Scene (ZEN71/72/76/77; includes 800LR variants)
- Central Scene/Scene Control must be enabled on the device (check device parameters in Z-Wave JS). On many Zooz models this is "Scene Control" parameter (often parameter 1). Ensure your firmware supports Central Scene.

## Configuration

1. Import this blueprint into Home Assistant.
2. Create an automation from the blueprint.
3. Select the Zooz switch device (integration: Z-Wave JS).
4. Select the primary light entity to control (used to read brightness and default target).
5. Optional: Choose an Area to target (actions affect the entire area, but brightness checks are still read from the single light entity).
6. Adjust dimming parameters as needed (defaults are sane):
   - Brightness step (%): default 5%
   - Dim interval (ms): default 200 ms
   - Minimum brightness on hold-up when off (%): default 10%
   - Clamp minimum (%): default 1% (below/at this threshold during dim-down → turn off)
   - Clamp maximum (%): default 100% (used for information/logging)
7. Optional: Add custom actions for double-tap and triple-tap (Up/Down).
8. Select a debug level (`off`, `basic`, or `verbose`).

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
