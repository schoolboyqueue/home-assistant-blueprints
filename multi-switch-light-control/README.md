# Multi Switch Light Control Pro

**Version:** 1.1.1  
**Author:** Jeremy Carter  
**Home Assistant blueprint that supports Zooz/Inovelli Z-Wave switches plus Lutron Pico remotes.**

[![Import Blueprint](https://my.home-assistant.io/badges/blueprint_import.svg)](https://my.home-assistant.io/redirect/blueprint_import/?blueprint_url=https%3A%2F%2Fgithub.com%2Fschoolboyqueue%2Fhome-assistant-blueprints%2Fblob%2Fmain%2Fmulti-switch-light-control%2Fmulti_switch_light_control_pro.yaml)

---

## Overview

This blueprint inspects the selected device's registry entry, prints the detected name/model/type when debugging, and automatically switches between Z-Wave Central Scene logic (Zooz/Inovelli) and Lutron Pico button handling. That means the same automation can control your Zooz paddle switch, an Inovelli scene-capable switch, or a Lutron Pico remote without copy/pasting different blueprints.

Supported behaviors:

- **Central scene devices (Zooz/Inovelli):** single press on/off, hold-to-dim/up, double/triple tap logging for custom actions, optional area targeting, and configurable step/interval/clamp values.
- **Lutron Pico:** On/Raise/Lower/Off/Stop (favorite) button sequences with hold-to-dim pacing, favorite button brightness/color defaults, and optional actions for the middle button.
- **Device info logging:** When debug level is `basic` or `verbose`, the automation writes the detected vendor/model/type to the system log so you can verify the auto-detection.

## Quick Start

1. Import the blueprint via the badge or by using this URL:
   `https://github.com/schoolboyqueue/home-assistant-blueprints/blob/main/multi-switch-light-control/multi_switch_light_control_pro.yaml`
2. Create an automation and select a supported switch device. The blueprint adapts its trigger set automatically.
3. Configure the light entity (and optionally an area) and adjust dimming/Lutron tuning inputs as needed.
4. Enable `basic` debug to confirm the detected device in Settings → System → Logs (`[Multi Switch Light Pro] Detected switch ...`).

## Inputs at a Glance

- **Switch device:** Pick any Zooz/Inovelli Z-Wave JS switch, or a Lutron Pico remote. No need to duplicate the automation for each vendor.
- **Light entity/area:** The blueprint always reads brightness from the chosen light entity, even when commands target an area.
- **Dimming parameters:** Adjust brightness step, interval, and min/max clamps. These values drive both Z-Wave and Lutron hold loops.
- **Lutron tuning:** Configure the favorite button defaults, transition speeds, and hold step delay specific to Pico remotes.
- **Diagnostics:** `basic`/`verbose` logging prints both central scene transitions and brightness calculations as the automation runs.

## Debug & Troubleshooting

- Enable **basic** or **verbose** debug to see `[Multi Switch Light Pro] Detected switch ...` and follow the action logs for each button.
- If the automation fires but nothing happens, confirm the selected device actually sends the expected events (Central Scene for Z-Wave, Lutrons for Pico).
- For hold loops, the blueprint respects the `dim_interval_ms`/`lutron_step_delay_ms` until a release event is received. If releases are dropped, the loop stops once the light hits the configured clamp.
- Favorite button custom actions are optional. If none are provided the blueprint sets the configured brightness/kelvin/transition.

## Changelog

See [CHANGELOG.md](CHANGELOG.md) for the full version history.
