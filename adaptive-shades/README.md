# Adaptive Shades Pro

**Version:** 1.10.0  
**Author:** Jeremy Carter  
**Home Assistant Blueprint for Solar-Adaptive Shade Control**

[![Import Blueprint](https://my.home-assistant.io/badges/blueprint_import.svg)](https://my.home-assistant.io/redirect/blueprint_import/?blueprint_url=https://github.com/schoolboyqueue/home-assistant-blueprints/blob/main/adaptive-shades/adaptive_shades_pro.yaml)

---

## Overview

Adaptive Shades Pro automates vertical blackout or zebra shades using the solar-geometry strategy outlined in *Energies 13(7):1731*. It keeps setup minimal—pick your shade, set the window orientation, and choose comfort bounds—while balancing glare control, passive solar gain, and cooling load.

### Highlights

- ☀️ **Solar-aware positioning** — Uses sun azimuth/elevation and your window orientation to move only when the sun is on that facade.
- 🌡️ **Comfort-biased** — Biases toward opening for heating (solar gain) or closing for cooling when temps cross your bounds.
- 👀 **Glare protection** — Optional indoor lux sensor triggers the block position when glare spikes.
- 💤 **Quiet hours & overrides** — Pause motion overnight or with a manual override helper.
- 🏠 **Presence-aware (optional)** — Only adjusts when someone is home if you provide a presence helper.

---

## Quick Start

1) **Import the blueprint** using the badge above or paste the URL into Settings → Automations & Scenes → Blueprints → Import Blueprint.  
2) **Create an automation** and select **Adaptive Shades Pro**.  
3) **Minimum configuration:**  
   - **Shade cover:** Your vertical shade (cover entity).  
   - **Shading mode:** `slat` for venetian/tilt-style; `zebra` for banded shades using calibrated admit/dim/block positions.  
   - **Window orientation:** Degrees the window faces (0° = North, 90° = East, 180° = South, 270° = West).  
   - (Optional) Indoor temp, outdoor temp, indoor lux, weather entity, climate entity, room profile, irradiance sensor, presence, quiet hours, manual override helper.
4) **Save and enable.** The blueprint will adjust every 5 minutes, at HA startup, and whenever the sun entity updates.

---

## Inputs at a Glance

- **Shade & Geometry**  
  - Shade cover, window orientation, azimuth tolerance, minimum sun elevation.  
  - Preferred open (admit) and block positions (0–100%).
- **Comfort & Glare (optional)**  
  - Indoor temp, outdoor temp, cooling upper bound, heating lower bound, comfort margin.  
  - Indoor lux sensor with glare threshold.
- **Scheduling & Overrides (optional)**  
  - Presence entity (only act when home/on).  
  - Manual override helper (input_boolean).  
  - Quiet hours start/end.

---

## Control Logic (Summary)

- Classify direct vs diffuse sun: compares a vertical irradiance sensor (if provided) to a clear-sky model (ASHRAE A/B), else falls back to sun azimuth/elevation and window orientation.
- Temperature bands driven by your heating/cooling setpoints (with comfort margin) and climate mode if provided (heat/cool).
  - **Occupied:** winter uses glare-limiting slat angle (Eq. 8); intermediate uses Eq. 8 when direct, 80° when diffuse; summer tilts to 45° for minimum daylight.  
  - **Unoccupied:** winter aligns to sun (β+90°) or diffuse empirical law (120 − 0.66·α) and uses G < 300 W/m² threshold; intermediate 80°; summer closes.
- Maps computed slat angle (0° open, 180° closed) to cover position and caps at `max_tilt_angle`; if the cover supports tilt, uses `cover.set_cover_tilt_position`, otherwise falls back to position. Zebra mode maps to admit/dim/block positions based on sun/glare/temp/occupancy.
- Night behavior: closes to the block position after sunset and resumes adaptive control after sunrise (respects manual override and quiet hours).
- Optional weather bias suppresses direct-sun classification on cloudy/rainy states when no irradiance sensor is present; optional manual timeout pauses automation after manual shade moves; optional climate entity biases heating/cooling and room profiles adjust glare sensitivity.

---

## Notes & Tips

- Positions use `cover.set_cover_position` (0 = fully open, 100 = fully closed). Zebra shades typically need partial positions; adjust the admit/block defaults to your fabric.  
- Quiet hours keep the last position until the end time.  
- Presence is optional; leave empty to run 24/7.  
- The blueprint uses the `sun.sun` entity—no extra dependencies.

---

## Reference

- B. K. (2020). *The Control of Venetian Blinds: A Solution for Reduction of Energy Consumption Preserving Visual Comfort,* **Energies 13(7):1731**. The blueprint implements the direct/diffuse classification, temperature bands, and slat-angle equations (including Eq. 8) adapted for Home Assistant covers.
