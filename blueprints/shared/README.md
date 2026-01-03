# Shared Resources for Home Assistant Blueprints

This directory contains shared resources for Home Assistant blueprints, including:

1. **Regional Profile Database** - Climate presets for U.S. states and international regions
2. **Template Macros** - Reusable Jinja2 calculation patterns (psychrometrics, solar geometry, adaptive comfort)

## Version

**v1.1.0** (2026-01-03)

## Files

### `templates.yaml` (NEW)

Reusable Jinja2 template macros for common calculation patterns. This file serves as the **single source of truth** for mathematical formulas used across blueprints.

**Included Templates:**

| Template | Description | Used By |
|----------|-------------|---------|
| Unit Conversion | C/F temperature conversion | All temperature blueprints |
| Heat Index | Rothfusz regression (feels-like temp) | adaptive-fan-control |
| Adaptive Comfort | EN 16798 / ASHRAE 55 model | adaptive-fan, adaptive-comfort |
| Psychrometrics | Dew point, enthalpy, humidity ratio | adaptive-comfort-control |
| Atmospheric Pressure | Barometric formula from elevation | adaptive-comfort-control |
| Solar Geometry | Sun position relative to window | adaptive-shades |
| Clear-Sky Irradiance | ASHRAE clear-sky model | adaptive-shades |
| Slat Angle | Venetian blind optimal angle | adaptive-shades |
| Time Schedule | Overnight schedule parsing | bathroom-light-fan, adaptive-shades |
| Presence Detection | Multi-entity presence check | All occupancy-based blueprints |
| Sensor Validation | Safe sensor value extraction | All blueprints |
| Climate Helpers | Thermostat attribute reading | adaptive-comfort, adaptive-shades |

**Why This File Exists:**
Home Assistant blueprints don't support external imports. This file documents the canonical formulas so:
- Developers can copy-paste verified implementations
- Formulas are documented with mathematical references
- Changes can be audited in one place before propagating to blueprints

### `regional_profiles.json`

Complete JSON database containing:

- **Climate Zones**: 8 climate zones with characteristics (Hot-Humid, Hot-Dry, Marine, Mixed-Humid, Mixed-Dry, Cold, Very Cold, Subarctic)
- **US States**: All 50 states + DC mapped to climate regions with elevation data
- **International Regions**: 50+ international regions (Canada, UK, Australia, New Zealand, Europe, Asia, Middle East, Africa, South America)
- **Regional Presets**: Climate-specific settings for:
  - Psychrometric thresholds (dew point, enthalpy, humidity)
  - Seasonal bias (winter, summer, shoulder seasons)
  - Adaptive comfort baselines (ASHRAE-55 based)
  - Ventilation thresholds
- **Thermostat Profiles**: 22 vendor profiles with minimum separation values

### `regional_profile_helpers.yaml`

Jinja2 template variables for use in Home Assistant blueprints. Contains:

- State/region to climate zone mappings
- Elevation lookup tables
- Psychrometric threshold lookups
- Seasonal bias presets
- Adaptive comfort baselines
- Ventilation thresholds
- Thermostat vendor profiles

## Usage in Blueprints

Since Home Assistant blueprints don't support external imports, copy the relevant variables from `regional_profile_helpers.yaml` into your blueprint's `variables:` section.

### Example: Get Climate Region from State Code

```yaml
variables:
  # Copy from regional_profile_helpers.yaml
  _us_state_to_region: >
      {{ {
        'AL': 'Hot-Humid',
        'AK': 'Very Cold',
        # ... rest of mapping
      } }}

  # Parse state code from "XX - State Name" format
  _state_code: "{{ (user_state_code.split(' ')[0] if user_state_code else '') }}"

  # Get the climate region
  climate_region: "{{ _us_state_to_region.get(_state_code, 'Mixed-Humid') }}"
```

### Example: Apply Regional Psychrometric Thresholds

```yaml
variables:
  _psychro_dp_max_c: >
      {{ {
        'Hot-Humid': 18.0,
        'Hot-Dry': 19.0,
        # ... rest of mapping
      } }}

  # Get dew point threshold for the region
  max_dew_point_c: "{{ _psychro_dp_max_c.get(climate_region, 17.0) }}"

  # Convert to Fahrenheit if needed
  max_dew_point_f: "{{ max_dew_point_c * 9/5 + 32 }}"
```

### Example: Seasonal Bias with Intensity Factor

```yaml
variables:
  _seasonal_winter_bias_c: >
      {{ {
        'Hot-Humid': 0.1,
        'Cold': 0.5,
        'Very Cold': 0.6,
        # ... rest of mapping
      } }}

  # Apply intensity factor (1.0 = recommended)
  bias_intensity: !input bias_preset_intensity

  # Calculate effective bias
  winter_bias_c: "{{ (_seasonal_winter_bias_c.get(region, 0.3) * (bias_intensity | float(1.0))) | round(2) }}"
```

## Climate Zones

| Zone | Description | Example Regions |
|------|-------------|-----------------|
| Hot-Humid | High temps & humidity year-round | FL, LA, TX, HI, Singapore, Queensland |
| Hot-Dry | High temps, low humidity | AZ, NV, NM, Dubai, Egypt |
| Marine | Mild coastal climate | CA, OR, WA, UK, NZ, Portugal |
| Mixed-Humid | Warm humid summers, mild winters | NC, TN, VA, Japan, Italy |
| Mixed-Dry | Warm dry summers, cold winters | CO, WY, ID, Spain, Greece |
| Cold | Warm summers, cold winters | NY, MA, MI, Germany, Ontario |
| Very Cold | Short cool summers, long cold winters | AK, MN, Finland, Norway |
| Subarctic | Very short summers, extreme cold | Yukon, NWT, Nunavut |

## Data Sources

- ASHRAE Standard 55-2020 (Thermal Environmental Conditions for Human Occupancy)
- IECC Climate Zone Maps (International Energy Conservation Code)
- Köppen-Geiger Climate Classification

## Maintenance

When updating this database:

1. Update `regional_profiles.json` with new data
2. Update `regional_profile_helpers.yaml` with corresponding Jinja2 lookups
3. Increment the version in both files
4. Update the `lastUpdated` field in the JSON metadata
5. Run `validate-blueprint --all` to validate any blueprints using this data (see `scripts/validate-blueprint-go/`)

## Using Template Macros

### Example: Adding Heat Index to Your Blueprint

```yaml
variables:
  # Required constants
  e_const: 2.718281828459045

  # Your input sensors
  t_in_c: "{{ states(temp_sensor) | float(21) }}"
  rh_in: "{{ states(humidity_sensor) | float(50) }}"

  # Heat index calculation (copy from templates.yaml)
  heat_index_c: >
    {% set T = t_in_c | float %}
    {% set R = rh_in | float %}
    {% if T < 20 or R < 40 %}
      {{ T }}
    {% else %}
      {% set c1 = -8.78469475556 %}
      {% set c2 = 1.61139411 %}
      {% set c3 = 2.33854883889 %}
      {% set c4 = -0.14611605 %}
      {% set c5 = -0.012308094 %}
      {% set c6 = -0.0164248277778 %}
      {% set c7 = 0.002211732 %}
      {% set c8 = 0.00072546 %}
      {% set c9 = -0.000003582 %}
      {% set HI = c1 + c2*T + c3*R + c4*T*R + c5*T*T + c6*R*R + c7*T*T*R + c8*T*R*R + c9*T*T*R*R %}
      {{ HI }}
    {% endif %}
```

### Example: Adding Adaptive Comfort Model

```yaml
variables:
  # Outdoor temperature (clamped to valid range)
  t_rm_c: "{{ [[t_out_c | float, 10] | max, 30] | min }}"

  # EN 16798 adaptive comfort formula
  comfort_temp_c: "{{ 0.33 * t_rm_c + 18.8 }}"

  # Tolerance based on category (I=strict, II=normal, III=relaxed)
  tol_c: >
    {% if comfort_category == 'I' %}2.0
    {% elif comfort_category == 'II' %}3.0
    {% else %}4.0{% endif %}

  # Comfort band
  comfort_upper_c: "{{ comfort_temp_c + tol_c | float }}"
  comfort_lower_c: "{{ comfort_temp_c - tol_c | float }}"
```

### Example: Time Schedule Parsing (Overnight Support)

```yaml
variables:
  schedule_start: !input quiet_hours_start
  schedule_end: !input quiet_hours_end

  in_quiet_hours: >
    {% set start_parts = schedule_start.split(':') if ':' in (schedule_start | string) else [] %}
    {% set end_parts = schedule_end.split(':') if ':' in (schedule_end | string) else [] %}
    {% if start_parts | length < 2 or end_parts | length < 2 %}
      {{ false }}
    {% else %}
      {% set start_m = (start_parts[0] | int(0)) * 60 + (start_parts[1] | int(0)) %}
      {% set end_m = (end_parts[0] | int(0)) * 60 + (end_parts[1] | int(0)) %}
      {% set now_m = now().hour * 60 + now().minute %}
      {% if start_m <= end_m %}
        {{ start_m <= now_m < end_m }}
      {% else %}
        {{ now_m >= start_m or now_m < end_m }}
      {% endif %}
    {% endif %}
```

## Template Development Guidelines

When adding new shared templates:

1. **Document the formula** with mathematical notation and references
2. **Include required inputs** clearly in comments
3. **Handle edge cases** (unavailable sensors, division by zero, etc.)
4. **Provide unit conversion** variants if applicable
5. **Test in isolation** before copying to blueprints
6. **Update this README** with the new template in the table above
