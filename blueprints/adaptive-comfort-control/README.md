# Adaptive Comfort Control Pro

**Author:** Jeremy Carter  
**Home Assistant Blueprint for Intelligent HVAC Automation**

[![Import Blueprint](https://my.home-assistant.io/badges/blueprint_import.svg)](https://my.home-assistant.io/redirect/blueprint_import/?blueprint_url=https://github.com/schoolboyqueue/home-assistant-blueprints/blob/main/blueprints/adaptive-comfort-control/adaptive_comfort_control_pro_blueprint.yaml)

---

## Overview

Adaptive Comfort Control Pro is an advanced Home Assistant automation blueprint that implements **ASHRAE-55 adaptive comfort model** with extensive enhancements for real-world residential use. Unlike traditional thermostats with fixed setpoints, this blueprint continuously adjusts comfort targets based on outdoor temperature, season, indoor air quality, sleep patterns, and learned preferences from your manual adjustments.

### Key Features

- 🌡️ **ASHRAE-55 Adaptive Comfort Model** — Scientifically-backed thermal comfort algorithm
- 🧠 **Machine Learning** — Learns from your manual adjustments over time
- 🌍 **Regional Climate Presets** — Optimized defaults for U.S. climate zones
- 🌊 **Built-in Psychrometrics** — Dew point, absolute humidity, enthalpy calculations
- 💨 **Intelligent Natural Ventilation** — Prevents muggy outdoor air from entering
- 🛏️ **Sleep Mode** — Cooler temperatures and tighter bands at night
- 🌱 **CO₂-Driven Ventilation** — Bias comfort when indoor air quality degrades
- ⚠️ **Safety Guards** — Freeze/overheat protection with risk-aware pause acceleration
- 🚪 **Smart HVAC Pause** — Stops heating/cooling when doors/windows open
- 🎛️ **Mixed Units Support** — Handles °C and °F sensors/thermostats seamlessly
- 🔧 **Vendor-Specific Profiles** — Auto-detects Ecobee, Nest, Honeywell, etc.

---

## Quick Start

### Prerequisites

- Home Assistant 2023.4 or newer
- Climate entity (thermostat)
- Indoor temperature sensor
- Outdoor temperature sensor

### Installation

1. **Import the blueprint:**
   - Click the badge above, or
   - Navigate to: Settings → Automations & Scenes → Blueprints → Import Blueprint
   - Paste URL: `https://github.com/schoolboyqueue/home-assistant-blueprints/blob/main/blueprints/adaptive-comfort-control/adaptive_comfort_control_pro_blueprint.yaml`

2. **Create an automation:**
   - Go to: Settings → Automations & Scenes
   - Click "+ Create Automation" → "Use Blueprint"
   - Select "Adaptive Comfort Control Pro"

3. **Minimum required configuration:**
   - **Climate Device:** Your thermostat entity
   - **Indoor Temperature:** Indoor temp sensor
   - **Outdoor Temperature:** Outdoor temp sensor

4. **Save and enable!**

The blueprint runs with sensible defaults. You can refine settings as needed.

---

## Mathematical Foundation

### 1. ASHRAE-55 Adaptive Comfort Model

The core of this blueprint is the **ASHRAE-55-2020 adaptive comfort model**, which establishes optimal indoor temperatures based on recent outdoor climate.

#### Base Comfort Temperature

```text
T_comfort = 18.9 + 0.255 × T_outdoor
```

Where:

- `T_comfort` = Optimal indoor operative temperature (°C)
- `T_outdoor` = Running mean outdoor temperature (°C)

**Valid range:** 10°C ≤ T_outdoor ≤ 40°C (50°F to 104°F)

#### Comfort Band (Tolerance Categories)

ASHRAE-55 defines three acceptability categories based on occupant satisfaction:

| Category | Tolerance       | Satisfaction | Use Case                         |
| -------- | --------------- | ------------ | -------------------------------- |
| **I**    | ±2.0°C (±3.6°F) | ~90%         | Office buildings, high standards |
| **II**   | ±3.0°C (±5.4°F) | ~80%         | Typical residential (default)    |
| **III**  | ±4.0°C (±7.2°F) | ~65%         | Maximum energy savings           |

**Comfort band calculation:**

```text
T_min = T_comfort - tolerance
T_max = T_comfort + tolerance
```

#### Enhanced Comfort Model (with biases)

This blueprint extends the base model with multiple adaptive layers:

```text
T_adapt = T_base + B_seasonal + B_sleep + B_CO₂ + B_learned + B_setback
```

Where:

- `T_base` = ASHRAE-55 baseline (18.9 + 0.255 × T_out)
- `B_seasonal` = Regional seasonal bias (-0.5°C to +0.6°C)
- `B_sleep` = Sleep mode cooling bias (user-configured, typically -1°C to -3°C)
- `B_CO₂` = CO₂-driven ventilation preference (0 to ±2°C)
- `B_learned` = Machine-learned offset from manual adjustments (±10°C range)
- `B_setback` = Energy-saving setback when unoccupied (user-configured)

---

### 2. Psychrometric Calculations

The blueprint computes psychrometric properties to evaluate natural ventilation suitability and prevent introducing humid outdoor air.

#### Magnus Formula (Saturation Vapor Pressure)

```text
P_sat = 0.61078 × exp((17.2694 × T) / (T + 237.3))
```

Where:

- `P_sat` = Saturation vapor pressure (kPa)
- `T` = Dry-bulb temperature (°C)
- `exp` = Exponential function (e^x)

#### Actual Vapor Pressure

```text
P_v = (RH / 100) × P_sat
```

Where:

- `P_v` = Actual vapor pressure (kPa)
- `RH` = Relative humidity (%)

#### Dew Point Temperature

Inverse Magnus formula:

```text
ln(P_v / 0.61078) = (17.2694 × T_dp) / (T_dp + 237.3)

Solving for T_dp:

T_dp = (237.3 × ln(P_v / 0.61078)) / (17.2694 - ln(P_v / 0.61078))
```

Where:

- `T_dp` = Dew point temperature (°C)
- `ln` = Natural logarithm

#### Absolute Humidity

```text
AH = (2167.4 × P_v) / (T + 273.15)
```

Where:

- `AH` = Absolute humidity (g/m³)
- `T` = Dry-bulb temperature (°C)
- Units: grams of water vapor per cubic meter of air

#### Humidity Ratio (Mixing Ratio)

```text
w = 0.62198 × (P_v / (P_total - P_v))
```

Where:

- `w` = Humidity ratio (kg water / kg dry air)
- `P_total` = Total barometric pressure (kPa)

**Safety guard:** Calculation skipped if `(P_total - P_v) < 0.1 kPa` to prevent division by zero.

#### Specific Enthalpy (Moist Air Energy Content)

```text
h = 1.006 × T + w × (2501 + 1.86 × T)
```

Where:

- `h` = Specific enthalpy (kJ/kg dry air)
- `T` = Dry-bulb temperature (°C)
- `w` = Humidity ratio (kg/kg)
- `1.006` = Specific heat of dry air (kJ/kg·K)
- `2501` = Latent heat of vaporization at 0°C (kJ/kg)
- `1.86` = Specific heat of water vapor (kJ/kg·K)

---

### 3. Adaptive Learning Algorithm

The blueprint learns your temperature preferences from manual thermostat adjustments using an **exponential weighted average (EWA)** algorithm.

#### Learning Formula

When you manually adjust the thermostat:

```text
Error = T_manual - T_predicted

B_learned_new = (1 - α) × B_learned_old + α × Error
```

Where:

- `T_manual` = Your manual setpoint adjustment (°C)
- `T_predicted` = What the blueprint predicted (°C)
- `α` = Learning rate (0.05 to 0.5, default: 0.15)
- `B_learned` = Learned offset applied to all future predictions (°C)

#### Convergence Timeline (α = 0.15)

| Adjustments | Adaptation |
| ----------- | ---------- |
| 1           | ~15%       |
| 5           | ~56%       |
| 10          | ~80%       |
| 15          | ~93%       |

**Mathematical basis:** Each new observation is weighted by α, while all previous history retains weight (1 - α). This creates exponential decay of older observations:

```text
B_learned(n) = α × Σ(i=0 to n) [(1-α)^i × Error(n-i)]
```

#### Learning Rate Selection

| α        | Behavior                | Best For                      |
| -------- | ----------------------- | ----------------------------- |
| 0.05     | Very slow, very stable  | Highly consistent preferences |
| 0.10     | Slow, stable            | Most users                    |
| **0.15** | **Moderate, balanced**  | **Recommended default**       |
| 0.20     | Fast, moderately stable | Adapting to new patterns      |
| 0.30     | Very fast, less stable  | Experimentation               |

---

### 4. Natural Ventilation Decision Logic

The blueprint blocks natural ventilation (turning off HVAC and opening windows) if outdoor air is "muggy" or energetically unfavorable.

#### Four-Part Guard Condition

Natural ventilation is **blocked** if any of these conditions are true:

**1. Outdoor dew point exceeds threshold:**

```text
T_dp_outdoor > T_dp_max
```

Typical: 16-19°C (61-66°F) depending on climate zone

**2. Outdoor dew point exceeds indoor (muggy delta):**

```text
T_dp_outdoor > (T_dp_indoor + ΔT_dp)
```

Typical: ΔT_dp = 1.5°C (prevents introducing more humid air)

**3. Outdoor absolute humidity exceeds indoor (economizer delta):**

```text
AH_outdoor > (AH_indoor + ΔAH)
```

Typical: ΔAH = 1.5-3.0 g/m³

**4. Outdoor enthalpy exceeds indoor (energy content):**

```text
h_outdoor > (h_indoor + Δh)
```

Typical: Δh = 2.5-4.0 kJ/kg

**Combined logic:**

```text
Block_ventilation = (T_dp_out > T_dp_max) OR
                    (T_dp_out > T_dp_in + ΔT_dp) OR
                    (AH_out > AH_in + ΔAH) OR
                    (h_out > h_in + Δh)
```

---

### 5. Risk Acceleration Math

When HVAC is paused (due to open doors/windows), the blueprint monitors indoor temperature for freeze/overheat risk and **shortens the resume delay** as risk increases.

#### Risk Calculation

```text
Risk_cold = max(0, T_freeze_guard - T_indoor) / T_warn_band
Risk_hot = max(0, T_indoor - T_overheat_guard) / T_warn_band

Risk_total = max(Risk_cold, Risk_hot)
```

Where:

- `T_freeze_guard` = Freeze protection threshold (e.g., 50°F / 10°C)
- `T_overheat_guard` = Overheat protection threshold (e.g., 95°F / 35°C)
- `T_warn_band` = Warning band width (e.g., 5°F / 3°C)
- `Risk_total` = 0.0 (no risk) to 1.0 (at guard threshold)

#### Accelerated Delay

```text
Delay_resume_effective = Delay_base × (1 - S_accel × Risk_total)

Constrained to: Delay_resume_effective ≥ Delay_min
```

Where:

- `Delay_base` = User-configured pause close delay (e.g., 120 seconds)
- `S_accel` = Acceleration strength (default: 1.0 for resume, 0.5 for pause)
- `Delay_min` = Minimum delay to prevent chattering (e.g., 10 seconds)

**Example:** With Delay_base = 120s, Risk = 0.6, S_accel = 1.0:

```text
Delay_effective = 120 × (1 - 1.0 × 0.6) = 48 seconds
```

The HVAC resumes in 48s instead of 120s, protecting against temperature extremes.

---

### 6. Barometric Pressure Estimation

Psychrometric calculations require barometric pressure. The blueprint supports:

1. **Barometric sensor** (if available)
2. **Elevation-based estimation** (standard atmosphere model)

#### Standard Atmosphere Model

```text
P(h) = 101.325 × (1 - 2.25577 × 10^-5 × h)^5.25588
```

Where:

- `P(h)` = Atmospheric pressure at elevation h (kPa)
- `h` = Elevation above sea level (meters)
- `101.325` = Sea-level standard atmospheric pressure (kPa)

**Derivation:** This is the barometric formula for the U.S. Standard Atmosphere (1976), using a polytropic exponent of 5.25588 (derived from g/RL where g=9.80665 m/s², R=287 J/kg·K, L=-0.0065 K/m).

#### Elevation Lookup Table

If no sensor or manual elevation is provided, the blueprint uses a built-in lookup table with approximate average elevations for all U.S. states:

| State | Elevation (m) | State | Elevation (m) |
| ----- | ------------- | ----- | ------------- |
| FL    | 30            | CO    | 2000          |
| LA    | 30            | WY    | 2040          |
| AZ    | 1250          | UT    | 1890          |
| AK    | 580           | NM    | 1735          |

---

## Configuration Guide

### Core Settings

**Climate Device** (required)  
Your thermostat entity (e.g., `climate.ecobee`, `climate.main_floor_thermostat`)

**Indoor Temperature** (required)  
Primary indoor air temperature sensor. Supports `sensor`, `input_number`, or `weather` entity.

**Outdoor Temperature** (required)  
Outdoor temperature for adaptive model. Supports `sensor`, `input_number`, or `weather` entity.

---

### Comfort & Units

**Temperature Units**

- **Auto-detect** (default): Infers from sensor `unit_of_measurement` attributes
- **Force Celsius (°C)**: Override if sensors report mixed units
- **Force Fahrenheit (°F)**: Override if sensors report mixed units

**ASHRAE-55 Category (Tolerance)**

- **Category I**: ±2°C / ±3.6°F — Office standards (~90% satisfaction)
- **Category II**: ±3°C / ±5.4°F — Typical home (default, ~80% satisfaction)
- **Category III**: ±4°C / ±7.2°F — Maximum savings (~65% satisfaction)

**Comfort Bounds**

Both °C and °F fields are shown; only the active one (based on Units setting) is used.

- **Minimum Comfort (°C)**: Hard lower bound (default: 18°C / 64°F)
- **Maximum Comfort (°C)**: Hard upper bound (default: 28°C / 82°F)
- **Minimum Comfort (°F)**: Hard lower bound (default: 64°F)
- **Maximum Comfort (°F)**: Hard upper bound (default: 82°F)

**Optional Features**

- **Use Operative Temperature**: Averages air temp + mean radiant temp (requires MRT sensor)
- **Typical Air Velocity**: Air movement (m/s) for elevated cooling effect (default: 0.1)
- **Auto Fan Adjustment**: Increases fan speed as room deviates from band
- **Humidity Comfort Correction**: Applies small bias for very high/low RH
- **Precision Comfort Mode**: Includes velocity/humidity in setpoint (more responsive)

---

### Seasonal Bias & Regional Presets

The blueprint adjusts comfort targets by season. You can use:

1. **Regional presets** (auto-configured by U.S. state/climate zone)
2. **Manual bias values**

**Regional Defaults**

Enable regional presets and select your state. The blueprint infers your climate zone:

| Climate Zone    | States                 | Winter Bias | Summer Bias |
| --------------- | ---------------------- | ----------- | ----------- |
| **Very Cold**   | AK                     | +0.6°C      | -0.2°C      |
| **Cold**        | NY, MI, WI, MN, etc.   | +0.5°C      | -0.2°C      |
| **Mixed-Dry**   | CO, WY, ID, MT         | +0.3°C      | -0.2°C      |
| **Mixed-Humid** | NC, TN, PA, OH, etc.   | +0.3°C      | -0.3°C      |
| **Hot-Dry**     | AZ, NV, UT, NM         | 0°C         | -0.4°C      |
| **Hot-Humid**   | FL, LA, MS, GA, HI, TX | +0.1°C      | -0.5°C      |
| **Marine**      | CA, OR, WA             | +0.1°C      | -0.1°C      |

**Manual Seasonal Bias**

If you disable regional presets, configure manually:

- **Winter Bias**: Degrees to add in winter (Dec-Feb)
- **Summer Bias**: Degrees to add in summer (Jun-Aug) — typically negative
- **Shoulder Bias**: Degrees to add in spring/fall (Mar-May, Sep-Nov)

---

### Sleep Mode

**Enable Sleep Mode**

Applies cooler temperatures and tighter comfort bands during sleep.

**Configuration:**

- **Sleep Start Time**: e.g., 22:00 (10 PM)
- **Sleep End Time**: e.g., 07:00 (7 AM)
- **Sleep Mode Entity**: Optional `binary_sensor` or `input_boolean` to override time-based detection
- **Sleep Bias**: Temperature reduction during sleep (default: -1.0°C / -1.8°F)
- **Sleep Band Tightening**: Reduces comfort band half-width for steadier temps (default: 0.5°C / 0.9°F)

**Sleep logic only applies when home is occupied** (requires occupancy sensor).

---

### Manual Override & Learning

**Manual Override Detection**

Automatically pauses automation when you manually adjust the thermostat.

**Settings:**

- **Enable Manual Override Detection**: On (default)
- **Override Duration**: How long to pause (default: 60 minutes)
- **Detection Tolerance**: Minimum change to trigger (default: 1.0°, prevents false triggers)
- **Override Action**: Optional automation/script to run (e.g., send notification)

**Adaptive Learning**

Learn from your manual adjustments over time.

**Settings:**

- **Learn from Manual Adjustments**: On (default)
- **Learning Rate (α)**: 0.05-0.5 (default: 0.15)
  - Lower = slower, more stable
  - Higher = faster, less stable
- **Learned Offset Storage**: Optional `input_number` helper for persistence across HA restarts

**Setup:**

1. Create helper: Settings → Devices & Services → Helpers → Number
2. Configure: Min: -10 (must be negative), Max: 10, Step: 0.01, Unit: °F or °C. Do not leave the minimum at 0 or above, or cooling adjustments cannot be learned.
3. Select in blueprint: "Learned Offset Storage"

See [LEARNING_SETUP.md](LEARNING_SETUP.md) for detailed setup guide.

---

### Natural Ventilation

**Natural Ventilation Enable**

When indoor temp exceeds comfort band AND outdoor conditions are favorable, turns off HVAC and suggests opening windows.

**Threshold:** Degrees above comfort band to trigger (default: 2.0°C / 3.6°F)

**Only Prefer Ventilation When Occupied:** When enabled (default), natural ventilation only runs if occupancy is detected. Disable to allow ventilation even when away.

**HVAC Stay-Off Guarantee:** When natural ventilation is active, final `climate.turn_on` is skipped to avoid re-enabling HVAC during a ventilation run.

**Psychrometric Guards** (see [Math Section](#4-natural-ventilation-decision-logic)):

- **Max Outdoor Dew Point**: Block if outdoor air is too humid (default: 16-19°C depending on region)
- **Muggy Delta**: Block if outdoor dew point exceeds indoor by this amount (default: 1.5°C)
- **Economizer Delta Enthalpy**: Energy content threshold (default: 2.5-4.0 kJ/kg)
- **Economizer Delta AH**: Absolute humidity threshold (default: 1.5-3.0 g/m³)

---

### HVAC Pause

**Pause Sensors**

List of door/window sensors (`binary_sensor`). When any open, HVAC pauses after delay.

**Delays:**

- **Pause after open**: Seconds to wait before pausing (default: 120s, prevents short door opens)
- **Resume after close**: Seconds to wait before resuming (default: 120s)
- **Max Timeout**: Maximum pause duration before forcing resume (default: 60 min)

**Risk Acceleration**

Shortens resume delay as indoor temp approaches freeze/overheat guards (see [Math](#5-risk-acceleration-math)).

**Settings:**

- **Enable Risk Acceleration**: On (default)
- **Warning Band**: Degrees before guard to start acceleration (default: 5°F / 3°C)
- **Acceleration Strength**: 0.0-1.0 (default: 1.0 for resume, 0.5 for pause)
- **Minimum Delays**: Prevents excessive chattering (default: 10s resume, 30s pause)

---

### Safety Features

**Freeze Protection**

- **Enable Freeze Protection**: On (default)
- **Freeze Guard Threshold**: Absolute minimum temp (default: 50°F / 10°C)
- **Block HVAC Pause at Risk**: Prevents pause near freeze threshold

**Overheat Protection**

- **Enable Overheat Protection**: On (default)
- **Overheat Guard Threshold**: Absolute maximum temp (default: 95°F / 35°C)
- **Block HVAC Pause at Risk**: Prevents pause near overheat threshold

---

## Regional Climate Presets

The blueprint includes optimized defaults for U.S. climate zones based on ASHRAE Climate Zone classifications.

### Preset Parameters by Region

| Region          | Seasonal Bias (W/Su/Sh) | Dew Point Max | CO₂ Delta | Economizer ΔH | Economizer ΔAH |
| --------------- | ----------------------- | ------------- | --------- | ------------- | -------------- |
| **Very Cold**   | +0.6 / -0.2 / +0.3°C    | 16.0°C        | 500 ppm   | 3.0 kJ/kg     | 2.0 g/m³       |
| **Cold**        | +0.5 / -0.2 / +0.2°C    | 17.0°C        | 450 ppm   | 3.0 kJ/kg     | 2.0 g/m³       |
| **Mixed-Dry**   | +0.3 / -0.2 / 0°C       | 16.0°C        | 450 ppm   | 2.5 kJ/kg     | 1.5 g/m³       |
| **Mixed-Humid** | +0.3 / -0.3 / 0°C       | 17.0°C        | 450 ppm   | 3.5 kJ/kg     | 2.0 g/m³       |
| **Hot-Dry**     | 0 / -0.4 / -0.1°C       | 19.0°C        | 450 ppm   | 2.0 kJ/kg     | 1.0 g/m³       |
| **Hot-Humid**   | +0.1 / -0.5 / -0.2°C    | 18.0°C        | 450 ppm   | 4.0 kJ/kg     | 3.0 g/m³       |
| **Marine**      | +0.1 / -0.1 / 0°C       | 16.5°C        | 400 ppm   | 3.0 kJ/kg     | 2.0 g/m³       |

**Preset Intensity Scaling:** You can scale all regional presets by 0%-200% (default: 100%).

---

## Thermostat Vendor Profiles

Different thermostats enforce different minimum separations for Auto/Heat-Cool mode. The blueprint auto-detects your thermostat and applies the correct profile.

### Supported Vendors

| Vendor                          | Minimum Separation | Detection Method       |
| ------------------------------- | ------------------ | ---------------------- |
| **Ecobee**                      | 5.0°F (2.8°C)      | Manufacturer attribute |
| **Google Nest**                 | 3.0°F (1.7°C)      | Manufacturer/model     |
| **Honeywell T6 Pro Z-Wave**     | 3.0°F (1.7°C)      | Model string + Z-Wave  |
| **Honeywell/Resideo (general)** | 2.0°F (1.1°C)      | Manufacturer           |
| **Sensi**                       | 3.0°F (1.7°C)      | Entity ID substring    |
| **Generic Zigbee**              | 2.0°F (1.1°C)      | Entity ID substring    |
| **Generic Z-Wave**              | 2.0°F (1.1°C)      | Entity ID substring    |
| **SmartIR**                     | 3.0°F (1.7°C)      | Entity ID substring    |
| **Generic Thermostat**          | 3.0°F (1.7°C)      | Entity ID substring    |
| **Manual Override**             | User-specified     | Blueprint input        |

**Auto-detection order:**

1. Blueprint input (if not "Auto")
2. Device manufacturer/model attributes
3. Entity ID substring matching
4. Device-advertised `min_temp_diff` attribute
5. Fallback: Generic (3.0°F / 1.7°C)

---

## Troubleshooting

### Common Issues

**"Provided temperature X not valid. Accepted range is Y to Z"**

- **Cause:** Blueprint computed a setpoint outside thermostat's range
- **Fix:** Implemented in v4.11+. Blueprint now quantizes to device step and clamps to safe maximum.
- **Verify:** Check `target_temp_step`, `min_temp`, `max_temp` attributes on your climate entity

**Comfort band too wide/narrow**

- **Adjust:** ASHRAE-55 Category (I/II/III) controls band width
- **Sleep Mode:** Reduces band width at night via "Sleep Band Tightening"

**Setpoints not updating**

- **Check triggers:** Blueprint runs on temp changes (30s debounce), time patterns (5 min), HA restart
- **Debug logs:** Enable "Debug Level: basic" and check logbook for trigger IDs
- **Manual override:** If you recently adjusted thermostat, automation pauses for configured duration

**Learning not working**

- **Verify:** "Learn from Manual Adjustments" is enabled
- **Tolerance:** Manual change must exceed tolerance (default: 1.0°)
- **Helper:** If using persistence, verify `input_number` helper exists and is selected
- **Debug:** Enable "Debug Level: verbose" to see learning calculations

**Natural ventilation blocked all the time**

- **Likely cause:** Outdoor air is humid (high dew point or absolute humidity)
- **Check:** Enable "Debug Level: verbose" to see psychrometric values
- **Adjust:** Increase "Max Outdoor Dew Point" or "Muggy Delta" thresholds
- **Regional presets:** May have conservative defaults for your location

**Thermostat profile not detected**

- **Check:** Device attributes via Developer Tools → States → Your climate entity
- **Look for:** `manufacturer`, `model` attributes
- **Override:** Set "Thermostat Profile" input to your vendor manually

---

## Advanced Topics

### Unit Conversion Strategy

The blueprint handles mixed units (°C/°F) using a multi-stage pipeline:

1. **Normalize to °C** — All calculations happen in Celsius internally
2. **Compute adaptive model** — ASHRAE-55, biases, psychrometrics (all in °C)
3. **Convert to "system units"** — User-facing units for display (°C or °F)
4. **Convert to "climate units"** — Thermostat's native units for service calls

This prevents unit drift and ensures compatibility with mixed sensor ecosystems.

### Variable Ordering

The `variables:` section is a **dependency-ordered pipeline**. Later variables can reference earlier ones, but not vice versa. Moving variable definitions can break the blueprint.

**Example dependency chain:**

```text
t_out_c → t_adapt_base_c → t_adapt_c_raw → t_adapt_c_guard → band_min_c → band_min_cli
```

### Psychrometrics Performance

Psychrometric calculations (dew point, AH, enthalpy) are computationally expensive in Jinja2 templates. The blueprint:

1. **Only computes when needed:** Psychrometrics disabled if not using natural ventilation
2. **Debug mode exception:** Verbose debug always computes for visibility
3. **Lazy evaluation:** Debug strings only rendered when debug enabled

### Trigger Debouncing

Temperature sensors can fluctuate rapidly. The blueprint debounces triggers:

- **Indoor temp:** 30-second delay
- **Outdoor temp:** 60-second delay
- **Time patterns:** Every 5 minutes
- **Manual overrides:** Immediate trigger

This reduces automation runs by ~70-80% during temperature fluctuations.

### Separation Enforcement

Auto/Heat-Cool mode requires minimum separation between low/high setpoints. The blueprint:

1. Computes ideal band: `[band_min_cli, band_max_cli]`
2. Quantizes to device step: `target_temp_step`
3. Checks separation: `sep = band_max_cli - band_min_cli`
4. If `sep < vendor_minimum`: Widens band symmetrically around adaptive target
5. Clamps to device range: `[min_temp, max_temp]`

This ensures service calls never fail due to invalid separation.

---

## References

### Scientific Standards

1. **ASHRAE Standard 55-2020:** _Thermal Environmental Conditions for Human Occupancy_  
   American Society of Heating, Refrigerating and Air-Conditioning Engineers  
   <https://www.ashrae.org/technical-resources/bookstore/standard-55-thermal-environmental-conditions-for-human-occupancy>

2. **Magnus Formula (1844):** Saturation vapor pressure approximation  
   Magnus, G. (1844). "Versuche über die Spannkräfte des Wasserdampfes"  
   Annalen der Physik und Chemie, 61, 225-247

3. **ASHRAE Fundamentals Handbook (2021):** Psychrometrics chapter  
   Chapter 1: Psychrometrics  
   <https://www.ashrae.org/technical-resources/ashrae-handbook>

4. **U.S. Standard Atmosphere (1976):** Barometric pressure model  
   NOAA, NASA, USAF  
   <https://ntrs.nasa.gov/citations/19770009539>

### Home Assistant Documentation

- **Blueprints:** <https://www.home-assistant.io/docs/automation/using_blueprints/>
- **Climate Integration:** <https://www.home-assistant.io/integrations/climate/>
- **Templating:** <https://www.home-assistant.io/docs/configuration/templating/>

### Adaptive Comfort Resources

- **CBE Thermal Comfort Tool:** <https://comfort.cbe.berkeley.edu/>  
  Interactive tool for ASHRAE-55 adaptive model calculations

- **ASHRAE Climate Zones:** <https://www.ashrae.org/technical-resources/bookstore/ashrae-climate-design-conditions>

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

**Enjoy adaptive comfort!**
