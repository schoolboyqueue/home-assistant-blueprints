# WARP.md

This file provides guidance to WARP (warp.dev) when working with code in this repository.

## Repository Overview

This repository contains **Home Assistant Blueprints** — reusable automation templates written in YAML with Jinja2 templating. These blueprints are designed to be imported into Home Assistant instances and configured through the UI.

### Blueprints

1. **Adaptive Comfort Control Pro** (`adaptive-comfort-control/`)
   - Advanced HVAC automation implementing ASHRAE-55 adaptive comfort model
   - Features: psychrometrics (dew point, absolute humidity, enthalpy), seasonal bias, regional presets, CO₂-driven ventilation, RMOT support, intelligent HVAC pause with risk acceleration
   - Supports mixed °C/°F units with auto-detection
   - Vendor-specific thermostat profiles (Ecobee, Nest, Honeywell, etc.) with auto-detection
   - Version: 4.11 (as of latest commit)

2. **Bathroom Light Fan Control Pro** (`bathroom-light-fan-control/`)
   - Coordinated bathroom light and fan automation

### File Structure

```
.
├── adaptive-comfort-control/
│   ├── adaptive_comfort_control_pro_blueprint.yaml  # Main blueprint
│   └── CHANGELOG.md
├── bathroom-light-fan-control/
│   ├── bathroom_light_fan_control_pro.yaml          # Main blueprint
│   └── CHANGELOG.md
└── .gitignore
```

## Development Workflow

### No Build System

This repository contains **no build tools, test frameworks, or CI/CD**. Development consists of:
- Editing YAML blueprints directly
- Testing by importing into Home Assistant and creating automations
- Manual validation of Jinja2 template logic

### Version Control

**View commit history:**
```bash
git log --oneline -20
```

**Create new version:**
1. Update `blueprint_version` variable in the YAML (e.g., `blueprint_version: "4.12"`)
2. Document changes in the blueprint's `CHANGELOG.md`
3. Commit with conventional commit format: `feat(component): description` or `fix(component): description`

### Testing Blueprints

**No automated tests exist.** Manual testing workflow:

1. Copy the blueprint URL from GitHub (raw content)
2. Import into Home Assistant: Settings → Automations & Scenes → Blueprints → Import Blueprint
3. Create an automation from the blueprint
4. Configure inputs through the UI
5. Monitor logs for template errors: Developer Tools → Logs
6. Check automation traces: Settings → Automations & Scenes → [Your Automation] → Traces

**Debug logging:**
Both blueprints support debug levels (`off`, `basic`, `verbose`). Set via the blueprint's Debug section.

### Making Changes

**When modifying blueprints:**

1. **Preserve backward compatibility** — existing automations should continue working
2. **Use template guards** — wrap new optional inputs with existence checks:
   ```yaml
   {% if new_optional_input is defined %}...{% endif %}
   ```
3. **Maintain variable ordering** — variables can depend on earlier definitions; order matters
4. **Test unit conversions** — many values support both °C and °F; validate both paths
5. **Document in CHANGELOG.md** — users need to know what changed

## Architecture Deep Dive

### Blueprint Structure

Home Assistant blueprints follow this pattern:

```yaml
blueprint:
  name: ...
  description: ...
  domain: automation
  input:
    # Grouped configuration UI
    group_name:
      input:
        setting_name:
          name: "Display Name"
          description: "Help text"
          selector: { ... }  # UI widget type

variables:
  # Jinja2 template calculations
  computed_value: "{{ some_template_logic }}"

trigger:
  # State changes, time patterns, etc.

action:
  # Automation steps
```

### Key Architectural Patterns

#### 1. **Multi-Stage Variable Pipeline**

The Adaptive Comfort Control blueprint computes values in stages:

1. **Input normalization** — handle `sensor.none`, empty strings, arrays vs strings
2. **Unit detection** — infer °C vs °F from sensor attributes
3. **Unit conversion to °C** — all calculations happen in Celsius internally
4. **Psychrometric calculations** — dew point, absolute humidity, enthalpy
5. **Adaptive model computation** — ASHRAE-55 formula with biases
6. **Band calculation** — comfort ranges with tolerance categories
7. **Safety clamping** — freeze/overheat guards
8. **Climate unit conversion** — convert back to thermostat's native units
9. **Quantization** — round to device step size
10. **Separation enforcement** — ensure minimum gap for Auto/Heat-Cool modes

**Critical insight:** Variable order matters. Variables reference earlier variables. Moving definitions breaks the pipeline.

#### 2. **Regional Presets System**

The blueprint includes climate zone presets:
- U.S. state → climate region mapping (Hot-Humid, Cold, Marine, etc.)
- Regional defaults for seasonal bias, ventilation thresholds, psychrometric limits
- Elevation lookup table for barometric pressure estimation

**Pattern:**
```yaml
_region_from_state: >  # Map state code to region
  {% set s = state_code %}
  {% if s in ['FL','LA','MS'...] %} Hot-Humid
  ...

_preset_winter_c: >  # Get preset value for region
  {% set r = region_pick %}
  {% if r == 'Hot-Humid' %} 0.1
  ...

winter_bias_sys: >  # Use preset or manual override
  {{ _preset_winter_c if regional_enabled else manual_value }}
```

#### 3. **Thermostat Vendor Auto-Detection**

Different thermostats enforce different minimum temperature separations in Auto/Heat-Cool mode. The blueprint:

1. Reads device manufacturer/model attributes
2. Matches against known vendors (Ecobee=5°F, Nest=3°F, etc.)
3. Falls back to generic profiles (Zigbee, Z-Wave)
4. Allows manual override

**Pattern:**
```yaml
thermostat_profile_auto: >
  {% set m = device_attr(..., 'manufacturer') | lower %}
  {% if 'ecobee' in m %} Ecobee
  {% elif 'nest' in m %} Google Nest
  ...

vendor_sep_cli_profile: >
  {% set p = thermostat_profile_eff %}
  {% if p == 'Ecobee' %} 5.0  # °F minimum
  ...
```

#### 4. **Risk-Aware Pause Acceleration**

When HVAC is paused (due to open windows/doors), the blueprint monitors indoor temperature:
- Near freeze risk → shorten resume delay
- Near overheat risk → shorten resume delay
- Prevents damage from prolonged HVAC shutdown

**Pattern:**
```yaml
risk_near: >  # 0.0 to 1.0 scale
  {% set risk_cold = freeze_guard - t_indoor %}
  {% set risk_hot = t_indoor - overheat_guard %}
  {{ [risk_cold, risk_hot] | max / warn_band }}

pause_close_delay_eff: >
  {% if risk_near > 0 %}
    {{ base_delay * (1 - accel_strength * risk_near) }}
  {% else %}
    {{ base_delay }}
  {% endif %}
```

#### 5. **Psychrometric Calculations**

Built-in psychrometrics prevent "muggy" natural ventilation:
- Computes dew point (Magnus formula)
- Computes absolute humidity (g/m³)
- Computes enthalpy (kJ/kg)
- Blocks ventilation if outdoor air is more humid than indoor

**All done in Jinja2 templates** — no external libraries.

### Common Pitfalls

1. **Unit confusion** — Some values are in "system units" (user preference), some in °C (calculations), some in "climate units" (thermostat native). Always check variable name suffix: `_sys`, `_c`, `_cli`.

2. **Float precision errors** — Home Assistant thermostats reject `33.333...°C` when max is `33.3°C`. Always quantize:
   ```yaml
   quantized: "{{ (value / step) | round(0,'floor') * step }}"
   ```

3. **Variable ordering** — If you add a new variable that depends on `foo`, it must come AFTER `foo` in the `variables:` section.

4. **Optional sensor handling** — Always check for `none`, `'unknown'`, `'unavailable'`, `''`, and `sensor.none`:
   ```yaml
   sensor_value: >
     {% if sensor_id and states(sensor_id) not in ['unknown','unavailable',''] %}
       {{ states(sensor_id) | float }}
     {% else %} {{ default_value }} {% endif %}
   ```

5. **Time window wrap-around** — Sleep schedules like `22:00 → 07:00` cross midnight. Handle explicitly:
   ```yaml
   {% if start <= end %}
     {{ (now >= start) and (now < end) }}
   {% else %}
     {{ (now >= start) or (now < end) }}
   {% endif %}
   ```

## Common Tasks

### Add a new blueprint input

1. Define in `blueprint.input.<group>.input.<name>` with selector
2. Bind to variable: `my_input: !input my_setting`
3. Use in calculations: `{{ my_input | float(default) }}`
4. Update CHANGELOG.md
5. Increment `blueprint_version`

### Add a new regional preset

1. Update state mapping in `_region_from_state`
2. Add region-specific values in preset variables (`_preset_w_c`, `_dp_max_c`, etc.)
3. Test both manual and preset modes

### Add a new thermostat vendor

1. Add detection logic in `thermostat_profile_auto` (check manufacturer/model/name)
2. Add minimum separation in `vendor_sep_cli_profile`
3. Test with both °C and °F climate entities

### Debug template errors

1. Enable `debug_level: verbose` in the blueprint
2. Check Developer Tools → Template to test snippets:
   ```jinja2
   {% set t_out_c = 20 %}
   {{ (18.9 + 0.255 * t_out_c) | round(2) }}
   ```
3. View full automation trace: Settings → Automations → [Name] → Traces → Latest run

### Update blueprint version

1. Edit `blueprint_version: "X.Y"` in YAML
2. Document in CHANGELOG.md (see existing format)
3. Commit: `feat(adaptive-comfort): description` or `fix(adaptive-comfort): description`
4. Users will see the new version when they update the blueprint in Home Assistant

## Common Commands

### Repository Management

**View recent commits:**
```bash
git log --oneline --graph -20
```

**Find blueprint versions:**
```bash
git grep -n "blueprint_version:"
```

**Check for uncommitted changes:**
```bash
git status
git diff
```

### Blueprint Development

**Extract input section:**
```bash
awk '/^input:/{p=1} p; /^variables:/{p=0}' adaptive-comfort-control/adaptive_comfort_control_pro_blueprint.yaml
```

**Extract variables section:**
```bash
awk '/^variables:/{p=1} p; /^trigger:/{p=0}' adaptive-comfort-control/adaptive_comfort_control_pro_blueprint.yaml
```

**Validate YAML syntax (optional):**
```bash
python3 -c 'import sys, yaml; yaml.safe_load(open(sys.argv[1]))' adaptive-comfort-control/adaptive_comfort_control_pro_blueprint.yaml
```

**Bump version (macOS sed):**
```bash
NEW_VERSION="4.12"
sed -E -i '' "s/^([[:space:]]*blueprint_version:[[:space:]]*).*/\\1\"${NEW_VERSION}\"/" adaptive-comfort-control/adaptive_comfort_control_pro_blueprint.yaml
```

### Searching Code

**Find all references to a variable:**
```bash
git grep -n "variable_name" adaptive-comfort-control/
```

**Find regional preset logic:**
```bash
git grep -nEi "region|preset|climate.*zone" adaptive-comfort-control/
```

**Find psychrometric calculations:**
```bash
git grep -nEi "dew|enthalpy|humidity|psychrom" adaptive-comfort-control/
```

## Blueprint Distribution

**Blueprints are distributed via GitHub URLs:**

1. Push to GitHub
2. Users import via raw GitHub URL in Home Assistant UI
3. Home Assistant caches the blueprint; users manually click "Update" to get new versions

**Testing locally:**
- Copy blueprint YAML to Home Assistant config: `config/blueprints/automation/<category>/`
- Restart Home Assistant or reload automations
- Blueprint appears in UI

## Jinja2 Template Reference

**Commonly used patterns in these blueprints:**

```jinja2
{# Comments #}

{# State access #}
{{ states('sensor.temperature') }}          # Current state
{{ state_attr('sensor.temp', 'unit') }}     # Attribute
{{ is_state('binary_sensor.door', 'on') }}  # Boolean check

{# Device attributes (vendor detection) #}
{{ device_attr('climate.thermostat', 'manufacturer') }}

{# Type conversion #}
{{ value | float(default) }}   # Float with fallback
{{ value | int(default) }}     # Integer
{{ value | round(2) }}         # Round to 2 decimals
{{ value | round(0, 'floor') }}  # Floor

{# String operations #}
{{ string | lower }}
{{ string | string }}          # Force to string
{{ 'substring' in string }}

{# Math #}
{{ [a, b, c] | min }}
{{ [a, b, c] | max }}
{{ log(x) }}                   # Natural log
{{ value ** exponent }}        # Power

{# Lists and filters #}
{{ expand(entity_list) | selectattr('state','eq','on') | list | count }}

{# Conditionals #}
{% if condition %} ... {% elif ... %} ... {% else %} ... {% endif %}
{{ value if condition else other }}

{# None/null checks #}
{{ value is none }}
{{ value is not none }}
{{ value is number }}
{{ value is string }}
```

## Home Assistant Blueprint Quirks

1. **No loops** — Jinja2 loops (`{% for %}`) work but are limited. Complex iterations can cause performance issues.

2. **No custom filters** — Stuck with Home Assistant's built-in Jinja2 filters.

3. **No imports** — Can't split blueprints into modules. Everything in one file.

4. **Variables are immutable** — Can't reassign. Use different variable names for each stage.

5. **Trigger context** — Use `trigger.id` to identify which trigger fired:
   ```yaml
   trigger:
     - platform: state
       id: temp_changed
       entity_id: sensor.temp
   action:
     - choose:
       - conditions:
           - "{{ trigger.id == 'temp_changed' }}"
   ```

6. **Service calls need data** — When calling Home Assistant services, provide exact data:
   ```yaml
   - service: climate.set_temperature
     target:
       entity_id: "{{ climate_entity_id }}"
     data:
       temperature: "{{ target_temp }}"
   ```

## Further Reading

- **Home Assistant Blueprint docs**: https://www.home-assistant.io/docs/automation/using_blueprints/
- **Jinja2 template reference**: https://www.home-assistant.io/docs/configuration/templating/
- **ASHRAE-55 adaptive comfort**: https://comfort.cbe.berkeley.edu/
