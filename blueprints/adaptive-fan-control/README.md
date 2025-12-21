# Ceiling Fan Climate Control Pro

**Author:** Jeremy Carter
**Home Assistant Blueprint for HVAC-Aware Ceiling Fan Automation with Adaptive Comfort**

[![Import Blueprint](https://my.home-assistant.io/badges/blueprint_import.svg)](https://my.home-assistant.io/redirect/blueprint_import/?blueprint_url=https://github.com/schoolboyqueue/home-assistant-blueprints/blob/main/blueprints/adaptive-fan-control/adaptive_fan_control_pro_blueprint.yaml)

**Version:** 2.0.0

---

## Overview

Ceiling Fan Climate Control Pro automates your ceiling fans using the EN 16798 / ASHRAE 55 adaptive comfort model. Instead of fixed temperature thresholds, it dynamically adjusts the comfort band based on outdoor temperature and indoor humidity, matching how humans actually perceive comfort.

### Why Adaptive Comfort?

Fixed thresholds like "turn on at 74°F" don't account for:

- **Humidity**: 77°F at 30% RH feels comfortable; 77°F at 70% RH feels muggy
- **Outdoor conditions**: When it's cold outside, you're adapted to cooler temps; when it's hot, you tolerate warmer indoor temps
- **Seasonal adaptation**: Your body acclimatizes to seasonal temperature ranges

The adaptive comfort model solves this by calculating a dynamic comfort band based on outdoor temperature, then factoring in humidity via heat index.

### Highlights

- 🌡️ **Adaptive comfort model** — EN 16798 / ASHRAE 55 standard where comfort bands shift with outdoor temperature
- 💧 **Humidity-aware** — Optional heat index calculation factors humidity into comfort decisions
- ❄️ **HVAC coordination** — Monitors thermostat hvac_action to work with heating/cooling cycles
- 🔄 **Deviation-based speed** — Fan speed scales with how far above comfort you are, not fixed thresholds
- 👤 **Occupancy-aware** — Only runs when the room is occupied
- 🔃 **Direction control** — Optional reverse mode for winter heating

---

## Quick Start

1. **Import the blueprint** using the badge above or paste the URL into Settings → Automations & Scenes → Blueprints → Import Blueprint.
2. **Create an automation** and select **Ceiling Fan Climate Control Pro**.
3. **Minimum configuration:**
   - **Ceiling fan:** Your fan entity to control
   - **Indoor temperature sensor:** Room temperature sensor
   - **Presence sensor:** Binary sensor for room occupancy
   - **Climate entity:** Your thermostat/HVAC system
4. **Recommended for adaptive comfort:**
   - **Outdoor temperature sensor:** Weather entity or outdoor temp sensor
   - **Indoor humidity sensor:** For heat index calculation
5. **Save and enable.**

---

## Inputs at a Glance

### Fan & Sensors

| Input | Required | Description |
|-------|----------|-------------|
| Ceiling fan | Yes | Fan entity to control |
| Indoor temperature sensor | Yes | Room temperature sensor |
| Indoor humidity sensor | No | Enables heat index calculation |
| Outdoor temperature sensor | No | Enables adaptive comfort mode (supports weather entities) |
| Presence sensor | Yes | Binary sensor for occupancy |
| Climate entity | Yes | Thermostat for HVAC coordination |

### Comfort Settings

| Input | Default | Description |
|-------|---------|-------------|
| Comfort mode | Adaptive | Fixed thresholds or EN 16798 adaptive |
| Comfort category | II (Normal) | I=±2°C strict, II=±3°C normal, III=±4°C relaxed |
| Temperature units | Auto-detect | Fahrenheit, Celsius, or auto-detect from sensor |

### Fan Capabilities

| Input | Default | Description |
|-------|---------|-------------|
| Supports direction | Off | Enable for fans with forward/reverse |
| Reverse when heating | Off | Run reverse during heating to circulate warm air |
| Heating speed | 25% | Fan speed during heating (if reverse enabled) |

### Speed Tiers

| Input | Default | Description |
|-------|---------|-------------|
| Speed mode | Deviation | Fixed thresholds or comfort deviation |
| Medium speed threshold | 2° | Degrees above comfort for medium speed |
| High speed threshold | 4° | Degrees above comfort for high speed |
| Low/Medium/High speed | 33/66/100% | Speed percentages for each tier |

---

## How It Works

### Adaptive Comfort Model (EN 16798)

The comfort temperature is calculated as:

```
T_comfort = 0.33 × T_outdoor + 18.8°C
```

Where `T_outdoor` is clamped to 10-30°C per the standard.

**Example calculations:**

| Outdoor Temp | Comfort Temp | Band (Cat II ±3°C) |
|--------------|--------------|---------------------|
| 50°F (10°C) | 72°F (22.1°C) | 66.7°F - 77.5°F |
| 68°F (20°C) | 78°F (25.4°C) | 72.7°F - 83.9°F |
| 86°F (30°C) | 84°F (28.7°C) | 78.5°F - 90.1°F |

### Heat Index Integration

When a humidity sensor is provided, the blueprint calculates heat index (feels-like temperature) using the Rothfusz regression. This is used instead of raw temperature for comfort decisions.

**Example:** At 77°F with 30% RH, heat index ≈ 76°F (comfortable). At 77°F with 70% RH, heat index ≈ 80°F (fan needed).

### Speed Calculation

In **deviation mode** (default), fan speed is based on how far above the comfort band you are:

| Deviation | Speed |
|-----------|-------|
| 0-2° above | Low (33%) |
| 2-4° above | Medium (66%) |
| 4°+ above | High (100%) |

### HVAC Coordination

| HVAC State | Fan Behavior |
|------------|--------------|
| Heating | Off (or reverse at low speed if enabled) |
| Cooling | On at calculated speed (distributes cool air) |
| Idle | Uses adaptive comfort band |

---

## Your Use Case: Winter Guest Bedroom

With the original v1.0 blueprint, your guest bedroom at 77.5°F would trigger the fan because it exceeded the fixed 74°F threshold.

With v2.0 adaptive comfort:

- **If outdoor temp is 40°F (winter):** Comfort band calculates to ~70-76°F. At 77.5°F you're slightly above comfort, fan runs at low speed.
- **If you add humidity sensor showing 30% RH:** Heat index calculates to ~76°F, which is within the comfort band. Fan stays off.
- **Category III (relaxed):** Widens the band to ±7.2°F, making 77.5°F definitely within comfort.

---

## Notes & Tips

- The fixed thresholds still act as absolute limits even in adaptive mode
- Outdoor temp sensor can be a weather entity (e.g., `weather.home`) or a dedicated sensor
- Debug logging shows the calculated comfort band and heat index values
- The blueprint runs every 5 minutes plus on sensor changes to catch gradual temperature shifts
- Season sensor (`sensor.season`) is used for automatic direction changes in winter

---

## Reference

- **EN 16798-1:2019** — Energy performance of buildings, indoor environmental quality
- **ASHRAE Standard 55** — Thermal Environmental Conditions for Human Occupancy
- **Heat Index** — Rothfusz regression equation (NWS)
