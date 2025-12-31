# Shared Regional Profile Database

This directory contains the centralized regional profile database for Home Assistant blueprints. It provides climate presets for U.S. states and international regions with seasonal bias adjustments, adaptive comfort baselines, and vendor-specific thermostat profiles.

## Version

**v1.0.0** (2025-12-31)

## Files

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
5. Run `python3 scripts/validate-blueprint/validate-blueprint.py --all` to validate any blueprints using this data
