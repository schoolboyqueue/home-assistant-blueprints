# Multi Switch Light Control Pro

**Author:** Jeremy Carter
**Home Assistant blueprint that supports Inovelli Zigbee switches (Zigbee2MQTT/ZHA), Zooz/Inovelli Z-Wave switches, and Lutron Pico remotes.**

[![Import Blueprint](https://my.home-assistant.io/badges/blueprint_import.svg)](https://my.home-assistant.io/redirect/blueprint_import/?blueprint_url=https%3A%2F%2Fgithub.com%2Fschoolboyqueue%2Fhome-assistant-blueprints%2Fblob%2Fmain%2Fblueprints%2Fmulti-switch-light-control%2Fmulti_switch_light_control_pro.yaml)

---

## Overview

This blueprint features **intelligent hardware auto-detection** that identifies your specific device model, auto-configures trigger mappings, and selects optimal control strategies based on device capabilities. It inspects the selected device's registry entry, detects the protocol (Zigbee2MQTT, ZHA, Z-Wave, or Lutron), and automatically adapts its triggers and logic.

### Supported Device Profiles

The blueprint recognizes these specific device models:

| Manufacturer | Models | Protocol | Features |
|-------------|--------|----------|----------|
| **Zooz** | ZEN71, ZEN72, ZEN73, ZEN74, ZEN76, ZEN77 | Z-Wave | LED, 5x multi-tap, Central Scene |
| **Inovelli Blue** | VZM31 (2-in-1), VZM35 (Fan) | Zigbee | LED, Config button, 5x multi-tap, Release detection |
| **Inovelli Red** | LZW30, LZW31, LZW36 | Z-Wave | LED, Config button, 5x multi-tap |
| **Lutron Pico** | 2-button, 3-button, 4-button, 5-button | Lutron | Release detection, Favorite button |

Supported behaviors:

- **Inovelli Zigbee (Zigbee2MQTT/ZHA):** Auto-detects action sensor entity, supports single tap on/off, hold-to-dim/brighten with release detection, double/triple/quad/quint tap for custom actions.
- **Z-Wave Central Scene (Zooz/Inovelli):** Single press on/off, hold-to-dim/up, double/triple/quad/quint tap logging for custom actions, optional area targeting, and configurable step/interval/clamp values.
- **Lutron Pico:** On/Raise/Lower/Off/Stop (favorite) button sequences with hold-to-dim pacing, favorite button brightness/color defaults, and optional actions for the middle button.
- **Switch state sync (optional):** Keep the physical switch/light entity mirrored with the target lights so other automations stay reflected in your switch state.
- **Protocol detection:** When debug level is `basic` or `verbose`, the automation writes the detected vendor/model/protocol to the system log so you can verify the auto-detection.

## Quick Start

1. Import the blueprint via the badge or by using this URL:
   `https://github.com/schoolboyqueue/home-assistant-blueprints/blob/main/blueprints/multi-switch-light-control/multi_switch_light_control_pro.yaml`
2. Create an automation and select a supported switch device. The blueprint adapts its trigger set automatically.
3. Configure the light entity (and optionally an area) and adjust dimming/Lutron tuning inputs as needed.
4. Enable `basic` debug to confirm the detected device in Settings → System → Logs (`[Multi Switch Light Pro] Detected switch ...`).

## Inputs at a Glance

- **Switch device:** Pick any Inovelli Zigbee switch (Zigbee2MQTT/ZHA), Zooz/Inovelli Z-Wave switch, or Lutron Pico remote. The blueprint auto-detects protocol - no need to duplicate the automation for each vendor.
- **Light entity/area:** The blueprint always reads brightness from the chosen light entity, even when commands target an area.
- **Dimming parameters:** Adjust brightness step, interval, and min/max clamps. These values drive Zigbee, Z-Wave, and Lutron hold loops.
- **Lutron tuning:** Configure the favorite button defaults, transition speeds, and hold step delay specific to Pico remotes.
- **Multi-tap actions:** Optional custom sequences for up/down presses (1x-5x) let you wire multi-tap gestures to other automations while still running the default turn on/off/dim behavior when left blank. Works with both Zigbee and Z-Wave switches.
- **Diagnostics:** `basic`/`verbose` logging prints protocol detection, button events, and brightness calculations as the automation runs.

## Debug & Troubleshooting

- Enable **basic** debug to see `[Multi Switch Light Pro] Detected switch ... | type=... | profile=...` showing the detected protocol and specific device profile (e.g., `zooz_zen77`, `inovelli_blue_2in1`, `lutron_pico_5button`).
- Enable **verbose** debug to also see full device capabilities including control strategy, protocol, max multi-tap level, button count, and feature flags (LED, release detection, etc.).
- **Zigbee switches:** The blueprint auto-detects the action sensor entity (e.g., `sensor.kitchen_switch_action`). If auto-detection fails or events aren't firing, manually select your action sensor in the "Zigbee action sensor" input. Verify Zigbee2MQTT is publishing action values correctly.
- **Z-Wave switches:** Confirm the device sends Central Scene events. Check Developer Tools → Events for `zwave_js_event` or `zwave_js_value_notification`.
- **Lutron Pico:** Verify `lutron_caseta_button_event` events appear when pressing buttons.
- For hold loops, the blueprint waits for release events (`up_release`, `down_release` for Zigbee; `KeyReleased` for Z-Wave; `release_raise`/`release_lower` for Lutron). If releases are dropped, loops stop when the light hits min/max clamps.
- Favorite button custom actions (Lutron) and multi-tap actions (Zigbee/Z-Wave) are optional. If none are provided, the blueprint runs default on/off/dim behavior.

## Changelog

See [CHANGELOG.md](CHANGELOG.md) for the full version history.
