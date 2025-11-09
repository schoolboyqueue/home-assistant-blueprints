# WARP.md

This file provides guidance to WARP (warp.dev) when working with code in this repository.

**IMPORTANT:** This file should be kept up-to-date whenever:
- New blueprints are added to the repository
- New architectural patterns are introduced
- Common pitfalls are discovered
- Remote testing workflows change

**When adding a new blueprint, update:**
1. The "Blueprints" section below with blueprint name and features (no version numbers)
2. The "File Structure" tree
3. The "Remote testing workflow" section with copy command
4. Any architecture sections if the blueprint introduces new patterns

**META-INSTRUCTION FOR WARP:**
Whenever you discover new patterns, pitfalls, solutions, or architectural insights while working on this repository, **proactively update this WARP.md file** with that knowledge. Do not wait for the user to ask. This includes:
- New common errors and their solutions
- New architectural patterns that work well
- New pitfalls to avoid
- Updates to existing sections when you learn better approaches
- Clarifications when existing documentation is ambiguous or incomplete

This ensures the knowledge base grows continuously and future sessions benefit from past learnings.

## Repository Overview

This repository contains **Home Assistant Blueprints** — reusable automation templates written in YAML with Jinja2 templating. These blueprints are designed to be imported into Home Assistant instances and configured through the UI.

### Blueprints

1. **Adaptive Comfort Control Pro** (`adaptive-comfort-control/`)
   - Advanced HVAC automation implementing ASHRAE-55 adaptive comfort model
   - Features: psychrometrics (dew point, absolute humidity, enthalpy), seasonal bias, regional presets, CO₂-driven ventilation, RMOT support, intelligent HVAC pause with risk acceleration, manual override persistence
   - Supports mixed °C/°F units with auto-detection
   - Vendor-specific thermostat profiles (Ecobee, Nest, Honeywell, etc.) with auto-detection
   - Optimized to skip unnecessary thermostat commands (50-80% reduction)

2. **Bathroom Light Fan Control Pro** (`bathroom-light-fan-control/`)
   - Coordinated bathroom light and fan automation using "Wasp-in-a-Box" occupancy detection
   - Features: humidity delta control, rate-of-rise/fall detection, night mode, manual override, presence-based activation
   - Supports mixed light entity and area control, fan or switch domains

3. **Zooz Z-Wave Light Switch Control Pro** (`zooz-zwave-light-switch-control/`)
   - Z-Wave switch light dimming using Central Scene events (ZEN71/72/76/77)
   - Features: single press on/off, hold-to-dim with release detection, area targeting
   - Supports both zwave_js_event and zwave_js_value_notification event types
   - Configurable dimming parameters (step size, interval, brightness thresholds)
   - Note: Double/triple tap triggers fire but don't execute actions (use separate automations if needed)

### File Structure

```
.
├── adaptive-comfort-control/
│   ├── adaptive_comfort_control_pro_blueprint.yaml  # Main blueprint
│   ├── CHANGELOG.md
│   └── README.md
├── bathroom-light-fan-control/
│   ├── bathroom_light_fan_control_pro.yaml          # Main blueprint
│   ├── CHANGELOG.md
│   └── README.md
├── zooz-zwave-light-switch-control/
│   ├── zooz_zwave_light_switch_control_pro.yaml     # Main blueprint
│   ├── CHANGELOG.md
│   └── README.md
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

**Semantic Versioning:**

All blueprints follow [Semantic Versioning](https://semver.org/) (MAJOR.MINOR.PATCH):
- **MAJOR** (X.0.0): Breaking changes (removing inputs, changing behavior incompatibly)
- **MINOR** (x.X.0): New features, backward-compatible additions
- **PATCH** (x.x.X): Bug fixes, no new features or breaking changes

**Create new version:**
1. Determine version increment based on change type:
   - Breaking change? Increment MAJOR (e.g., 1.10.1 → 2.0.0)
   - New feature? Increment MINOR (e.g., 1.10.1 → 1.11.0)
   - Bug fix? Increment PATCH (e.g., 1.10.1 → 1.10.2)
2. Update `blueprint_version` variable in the YAML (e.g., `blueprint_version: "1.11.0"`)
3. Update the `name:` field in the blueprint to include the new version (e.g., `name: "Blueprint Name v1.11.0"`)
4. Document changes in the blueprint's `CHANGELOG.md` with categorized sections:
   - `### Added` - New features
   - `### Changed` - Changes in existing functionality
   - `### Deprecated` - Soon-to-be removed features
   - `### Removed` - Removed features (breaking change)
   - `### Fixed` - Bug fixes
   - `### Security` - Security fixes
5. Commit with conventional commit format:
   - Breaking: `feat(component)!: description` or `fix(component)!: description`
   - Feature: `feat(component): description`
   - Fix: `fix(component): description`
   - Chore: `chore: description`
6. Version numbers appear in blueprint names and variables only (not in WARP.md to save tokens)

### Testing Blueprints

**No automated tests exist.** Manual testing workflow:

1. Copy the blueprint URL from GitHub (raw content)
2. Import into Home Assistant: Settings → Automations & Scenes → Blueprints → Import Blueprint
3. Create an automation from the blueprint
4. Configure inputs through the UI
5. Monitor logs for template errors: Developer Tools → Logs
6. Check automation traces: Settings → Automations & Scenes → [Your Automation] → Traces

**Debug logging:**
All blueprints support debug levels (`off`, `basic`, `verbose`). Set via the blueprint's Debug/Diagnostics section.
- `off`: No debug output
- `basic`: Key events (light ON/OFF, fan ON/OFF, manual override, press/hold/release, etc.)
- `verbose`: Detailed state information, sensor values, condition breakdowns, brightness calculations

### Home Assistant Server Access

**SSH access is configured** for direct server interaction:

```bash
# Connect to Home Assistant server
ssh homeassistant
```

**Remote testing workflow:**

1. **Copy blueprint to server** (SCP doesn't work, use pipe through SSH):
   ```bash
   cat adaptive-comfort-control/adaptive_comfort_control_pro_blueprint.yaml | \
       ssh homeassistant "cat > /tmp/blueprint.yaml && sudo mv /tmp/blueprint.yaml /config/blueprints/automation/schoolboyqueue/adaptive_comfort_control_pro_blueprint.yaml"
   
   cat bathroom-light-fan-control/bathroom_light_fan_control_pro.yaml | \
       ssh homeassistant "cat > /tmp/blueprint.yaml && sudo mv /tmp/blueprint.yaml /config/blueprints/automation/schoolboyqueue/bathroom_light_fan_control_pro.yaml"
   
   cat zooz-zwave-light-switch-control/zooz_zwave_light_switch_control_pro.yaml | \
       ssh homeassistant "cat > /tmp/blueprint.yaml && sudo mv /tmp/blueprint.yaml /config/blueprints/automation/schoolboyqueue/zooz_zwave_light_switch_control_pro.yaml"
   ```

2. **Reload automations** (via Home Assistant UI or CLI if available):
   - Settings → System → Restart → Quick reload (automations only)
   - Or: Developer Tools → YAML → Automations

3. **Check Home Assistant logs:**
   ```bash
   ssh homeassistant "tail -f /config/home-assistant.log"
   ```

4. **View blueprint files on server:**
   ```bash
   ssh homeassistant "ls -la /config/blueprints/automation/"
   ```

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
6. **Beware Jinja2 variable scoping** — variables set inside loops don't persist; use `namespace()` for mutable state

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

6. **Jinja2 variable scoping in loops** — Setting a variable inside a loop doesn't persist outside the loop. Use `namespace()` for mutable state:
   ```yaml
   # WRONG - ok will always be false outside loop
   {% set ok = false %}
   {% for e in entities %}
     {% if states(e) == 'on' %}
       {% set ok = true %}  # This doesn't persist!
     {% endif %}
   {% endfor %}
   {{ ok }}  # Always false
   
   # CORRECT - use namespace for mutable state
   {% set ns = namespace(found=false) %}
   {% for e in entities %}
     {% if states(e) == 'on' %}
       {% set ns.found = true %}  # This persists!
     {% endif %}
   {% endfor %}
   {{ ns.found }}  # Correctly returns true if any entity was 'on'
   ```

7. **input_datetime timezone issues** — When comparing `input_datetime` helpers with current time, **always use the `timestamp` attribute**, not `states()`:
   ```yaml
   # WRONG - timezone conversion causes comparison failures
   {% set until_ts = as_timestamp(states(input_datetime.helper)) %}
   {% set now_ts = as_timestamp(now()) %}
   {{ now_ts > until_ts }}  # May fail due to timezone mismatch
   
   # CORRECT - use timestamp attribute directly
   {% set until_ts = state_attr(input_datetime.helper, 'timestamp') %}
   {% set now_ts = as_timestamp(now()) %}
   {{ now_ts > until_ts }}  # Both in Unix epoch seconds
   ```
   
   **Why this fails:** `states(input_datetime.helper)` returns a **string** in local timezone (e.g., "2025-11-08 11:34:47"). When `as_timestamp()` parses this string, it may interpret it as UTC, causing comparison failures. The `timestamp` attribute is a **float** (Unix epoch) with no timezone ambiguity.

8. **State change trigger race conditions** — When an automation changes entity state AND that state change triggers another branch, add a small delay to avoid race conditions:
   ```yaml
   # WRONG - simultaneous state changes cause race conditions
   - service: input_boolean.turn_on
     target:
       entity_id: input_boolean.flag
   - service: light.turn_off  # This triggers light_manual_off
     target:
       entity_id: light.bathroom
   - service: input_boolean.turn_off  # Might happen before trigger fires
     target:
       entity_id: input_boolean.flag
   
   # CORRECT - add delay to ensure trigger sees correct state
   - service: input_boolean.turn_on
     target:
       entity_id: input_boolean.flag
   - service: light.turn_off  # This triggers light_manual_off
     target:
       entity_id: light.bathroom
   - delay:
       milliseconds: 100  # Let state change settle
   - service: input_boolean.turn_off
     target:
       entity_id: input_boolean.flag
   ```
   
   **Why this matters:** When an entity state change triggers another automation or trigger within the same blueprint, the trigger fires immediately. If you change another entity's state in the same action sequence, both state changes might be processed simultaneously, causing the trigger to see the wrong state. A small delay (100ms) ensures the first state change is fully processed before the second occurs.

9. **Using `target:` selector types in trigger `entity_id` fields** — A `selector:` of type `target:` may allow selection of devices, areas, or entities, returning a dictionary (not the plain entity ID list that state triggers expect).
   ```yaml
   # WRONG - target selector returns dictionary
   input:
     my_entity:
       selector:
         target: {}  # Returns {area_id: ..., device_id: ..., entity_id: [...]}
   trigger:
     - platform: state
       entity_id: !input my_entity  # Fails - expects entity ID string/list
   
   # CORRECT - use entity selector for triggers
   input:
     my_entity:
       selector:
         entity: {}  # Returns entity ID string directly
   trigger:
     - platform: state
       entity_id: !input my_entity  # Works
   ```
   
   **Why this fails:** State triggers expect entity IDs, not target dictionaries. Blueprint fails to import or triggers silently never fire.

10. **Using variables or `!input` in trigger templates or `for:` durations** — Variables defined in `variables:` section aren't available when triggers are evaluated. Templates in triggers have limited context.
    ```yaml
    # WRONG - variable not available in trigger
    variables:
      my_duration: !input delay_minutes
    trigger:
      - platform: state
        entity_id: sensor.test
        for:
          minutes: "{{ my_duration }}"  # Undefined variable error
    
    # CORRECT - use !input directly or hardcode
    trigger:
      - platform: state
        entity_id: sensor.test
        for:
          minutes: !input delay_minutes  # Works
    ```
    
    **Why this fails:** Triggers are evaluated at blueprint compile time, before variables are available. Move logic to `action:` section using `wait_for_trigger` if needed.

11. **Incorrect variable ordering in `variables:` section** — Variables referencing other variables must come after those variables in the list. Home Assistant evaluates variables sequentially.
    ```yaml
    # WRONG - order matters
    variables:
      result: "{{ base_value * 2 }}"  # Undefined variable error
      base_value: 10
    
    # CORRECT - dependencies first
    variables:
      base_value: 10
      result: "{{ base_value * 2 }}"  # Works
    ```
    
    **Why this fails:** Variables are evaluated in order. Later variables can reference earlier ones, but not vice versa. See Pitfall #3.

12. **Using `!input` inside Jinja2 template blocks** — Embedding `!input` directly inside `{{ }}` confuses YAML/Jinja parsing.
    ```yaml
    # WRONG - !input inside template
    variables:
      result: "{{ !input my_value * 2 }}"  # YAML parsing error
    
    # CORRECT - bind input to variable first
    variables:
      my_value: !input my_setting
      result: "{{ my_value * 2 }}"  # Works
    ```
    
    **Why this fails:** `!input` is a YAML tag processed before Jinja2 evaluation. Map inputs to variables first, then use variables in templates.

13. **Blueprint limitations with complex trigger types** — Some trigger types (`template` triggers, webhook triggers) have inconsistent support in blueprints or may not work as expected.
    ```yaml
    # RISKY - template triggers in blueprints
    trigger:
      - platform: template
        value_template: "{{ states('sensor.temp') | float > 25 }}"  # May not work reliably
    
    # SAFER - use well-supported triggers
    trigger:
      - platform: numeric_state
        entity_id: sensor.temp
        above: 25  # Better blueprint support
    ```
    
    **Why this matters:** Stick to `state`, `numeric_state`, `time_pattern` triggers in blueprints. Test thoroughly if using advanced trigger types.

14. **Not documenting required helpers/entities** — Users may not know they need to create helpers (input_boolean, input_datetime, etc.) before using the blueprint.
    ```yaml
    # Good practice - document requirements clearly
    blueprint:
      description: >
        REQUIRED: Create these helpers before using:
        1. input_datetime.my_override_until (date and time enabled)
        2. input_boolean.my_automation_control
      input:
        helper:
          description: "REQUIRED: Select the input_datetime helper you created"
          selector:
            entity:
              domain: input_datetime
    ```
    
    **Why this matters:** Clear documentation and selector filters prevent misconfiguration. Users should know what to create before importing the blueprint.

15. **Time pattern triggers causing load spikes** — Multiple installations using identical `time_pattern` triggers fire simultaneously, potentially overwhelming external services.
    ```yaml
    # RISKY - all instances fire at same time
    trigger:
      - platform: time_pattern
        minutes: "/5"  # Everyone fires at :00, :05, :10, etc.
    action:
      - service: notify.pushover  # May rate-limit if many instances
    
    # BETTER - add jitter/stagger
    trigger:
      - platform: time_pattern
        minutes: "/5"
    action:
      - delay:
          seconds: "{{ range(0, 60) | random }}"  # Random 0-60s delay
      - service: notify.pushover
    ```
    
    **Why this matters:** Synchronized triggers can cause rate limiting, service failures, or external API throttling. Add randomization for external service calls.

16. **Using entity attributes in places expecting state** — Some automations try to use attributes directly in triggers/conditions that expect state values.
    ```yaml
    # WRONG - attribute in state trigger
    trigger:
      - platform: state
        entity_id: sensor.temp
        attribute: unit_of_measurement  # Triggers on attribute change
        to: "°F"  # This works for attributes
    
    # BETTER - use attribute in template/condition
    trigger:
      - platform: state
        entity_id: sensor.temp  # Trigger on state change
    condition:
      - condition: template
        value_template: "{{ state_attr('sensor.temp', 'unit_of_measurement') == '°F' }}"  # Check attribute
    ```
    
    **Why this matters:** Attributes and state are different. State triggers with `attribute:` key work but may not behave as expected. Use templates for attribute logic when possible.

17. **Using `choose` when you mean `if/then/else`** — When you have a simple binary condition (one check, two outcomes), use `if/then/else`, not `choose`. Using `choose` with a `default:` section that contains more action items at the same indentation level creates malformed YAML.
    ```yaml
    # WRONG - choose for binary condition
    - choose:
        - conditions: "{{ area_set }}"
          sequence:
            - service: light.turn_on
              target:
                area_id: "{{ light_area }}"
      default:
        - service: light.turn_on
          target:
            entity_id: "{{ light_entity }}"
    
    # CORRECT - if/then/else for binary condition
    - if:
        - condition: template
          value_template: "{{ area_set }}"
      then:
        - service: light.turn_on
          target:
            area_id: "{{ light_area }}"
      else:
        - service: light.turn_on
          target:
            entity_id: "{{ light_entity }}"
    
    # WRONG - wrapping optional !input actions in choose with conditions
    - choose:
        - conditions: []
          sequence: !input double_tap_action
    
    # WRONG - calling !input actions directly as list item
    - !input double_tap_action  # Error: expects dictionary, got list
    
    # CORRECT - use choose with empty condition list and default
    - choose: []
      default: !input double_tap_action
    ```
    
    **Why this fails:** `choose` is for multiple conditions (like switch/case), while `if/then/else` is for binary decisions. The Home Assistant schema validator rejects `choose` blocks with `default:` followed by additional sequence items, producing errors like "extra keys not allowed @ data['actions'][...]['default']". This error occurs when `choose`/`default` is used instead of `if/then/else`, especially in nested action sequences.
    
    **Optional input actions:** When an input has `default: []` (empty action list), you cannot call it directly as a sequence item because `!input` returns a **list** of actions, not a single action. The correct pattern is `choose: []` with `default: !input action_name`. This means "no conditions match, so always execute the default actions". The empty `choose: []` is required for proper YAML structure.

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
3. **Update blueprint input descriptions** if changing user-facing behavior or adding inputs
4. **Update README.md** to reflect changes (version number, configuration sections, troubleshooting)
5. Commit: `feat(adaptive-comfort): description` or `fix(adaptive-comfort): description`
6. Users will see the new version when they update the blueprint in Home Assistant

### Update blueprint input descriptions and README

When making user-facing changes to inputs or behavior:

1. **Update input descriptions** in blueprint YAML (`description:` field)
   - Explain options clearly (e.g., "Set to 0 to disable", "Optional: ...", "Required: ...")
   - Mention defaults and edge cases
   - Keep descriptions concise but complete

2. **Update README.md**:
   - Version number at top of file
   - Configuration Guide sections (explain new inputs, update existing descriptions)
   - Use Cases section if behavior changed
   - Troubleshooting section for new edge cases or common issues

3. **Why this matters**: Users rely on descriptions in the UI and README to configure blueprints correctly. Outdated docs lead to misconfiguration and support questions.

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