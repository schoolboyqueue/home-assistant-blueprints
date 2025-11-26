# WARP.md

This file provides guidance to WARP (warp.dev) when working with code in this repository.

**IMPORTANT:** This file should be kept up-to-date whenever:
- New blueprints are added to the repository
- New architectural patterns are introduced
- Common pitfalls are discovered
- Documentation and distribution guidance changes

**When adding a new blueprint, update:**
1. The "Blueprints" section below with blueprint name and features (no version numbers)
2. The "File Structure" tree
3. The "Blueprint Distribution" section if distribution approach changes
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

3. **Multi Switch Light Control Pro** (`multi-switch-light-control/`)
   - Universal light control for Zooz/Inovelli Z-Wave switches and Lutron Pico remotes
   - Features: auto-detection of switch type, single press on/off, hold-to-dim, Central Scene actions (1x-5x taps), Lutron favorite button
   - Supports both zwave_js_event and zwave_js_value_notification event types
   - Multi-entity targeting with first entity used for brightness reads
   - Configurable dimming parameters (step size, interval, brightness thresholds, transitions)

4. **Adaptive Shades Pro** (`adaptive-shades/`)
   - Automated shade positioning (slat and zebra/roller) using solar geometry and comfort signals
   - Features: facade-based direct/diffuse detection, clear-sky irradiance comparison, multi-mode behavior (winter/intermediate/summer), presence-aware behavior, quiet hours, manual override, optional weather and climate integration

### File Structure

```text
.
├── adaptive-comfort-control/
│   ├── adaptive_comfort_control_pro_blueprint.yaml  # Main blueprint
│   ├── CHANGELOG.md
│   └── README.md
├── bathroom-light-fan-control/
│   ├── bathroom_light_fan_control_pro.yaml          # Main blueprint
│   ├── CHANGELOG.md
│   └── README.md
├── multi-switch-light-control/
│   ├── multi_switch_light_control_pro.yaml          # Main blueprint
│   ├── CHANGELOG.md
│   └── README.md
├── adaptive-shades/
│   ├── adaptive_shades_pro.yaml                     # Main blueprint
│   ├── CHANGELOG.md
│   └── README.md
├── scripts/
│   ├── validate-blueprint.py                        # Blueprint validator
│   └── README.md                                    # Validation documentation
├── CLAUDE.md                                        # Claude Code guidance
├── WARP.md                                          # WARP agent guidance
├── AGENTS.md                                        # Repository structure
└── .gitignore
```

## Development Workflow

### No Build System

This repository contains **no build tools, test frameworks, or CI/CD**. Development consists of:
- Editing YAML blueprints directly
- **Validating locally using the blueprint validator** (REQUIRED before commit)
- Testing by importing into Home Assistant and creating automations
- Manual validation of Jinja2 template logic

**Blueprint Validation (REQUIRED):**

Before committing any blueprint changes, always run:
```bash
# Validate single blueprint after editing
python3 scripts/validate-blueprint.py <path/to/blueprint.yaml>

# Validate ALL blueprints before committing
python3 scripts/validate-blueprint.py --all
```

The validator catches structural errors locally:
- YAML syntax errors
- Missing required keys (`blueprint`, `trigger`, `action`)
- `variables` nested under `blueprint` (common error that breaks imports)
- Empty/None `data:` blocks in service calls
- `!input` tags inside `{{ }}` templates
- Unbalanced template braces
- Invalid selector types
- Improper indentation in `if`/`then`/`else` blocks

See [scripts/README.md](scripts/README.md) for complete validation documentation.

### Version Control

**View commit history:**
```bash
git log --oneline -20
```

**Agent wrap-up expectations (auto-commit & push):**
- When an agent modifies any blueprint YAML (or its README/CHANGELOG), it should by default:
  1. Determine the correct SemVer bump (MAJOR/MINOR/PATCH) based on the change.
  2. Update `blueprint_version` and `blueprint.name` to the new version.
  3. Update the corresponding `CHANGELOG.md` with a dated entry summarizing the change.
  4. Update the blueprint `README.md` version header to match.
  5. **Run the blueprint validator** (REQUIRED): `python3 scripts/validate-blueprint.py <blueprint.yaml>`
  6. Fix any validation errors before proceeding.
  7. Stage the relevant files, create a conventional commit, and push to the default branch.
- This should happen automatically as part of the agent's workflow without needing an explicit user request.

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
- `verbose`: Detailed state information, sensor values, condition breakdowns, brightness/shade calculations

### Making Changes

**When modifying blueprints:**

1. **Preserve backward compatibility** — existing automations should continue working.
2. **Use template guards** — wrap new optional inputs with existence checks or null-safe fallbacks.
3. **Maintain variable ordering** — variables can depend on earlier definitions; order matters.
4. **Test unit conversions** — many values support both °C and °F; validate both paths.
5. **Document in CHANGELOG.md** — users need to know what changed.
6. **Beware Jinja2 variable scoping** — variables set inside loops don't persist; use `namespace()` for mutable state.

---

## Home Assistant Jinja & Blueprint Editing Rules

WARP must treat Home Assistant Jinja as a **restricted runtime**, not full Python or full Jinja2.

### Core Constraints

When editing templates:

- Do **not** introduce Python code (no `math.sin`, `pow`, `import`, etc.).
- Do **not** add custom Jinja extensions or macros (`macro`, `import`, `include`, `call`).
- Use only filters known to be supported by Home Assistant, or that are already present in the file.

Safe baseline filters:
- `int`, `float`, `string`
- `round`, `abs`
- `default`
- `lower`, `upper`, `title`, `capitalize`
- `replace`, `regex_replace`
- `length`

### Math & Angle Handling

Prefer explicit math over fragile filters:

- Degrees → radians:

  ```jinja2
  {{ (angle_deg | float * pi / 180) }}
  ```

- Radians → degrees:

  ```jinja2
  {{ (angle_rad | float * 180 / pi) }}
  ```

Use `sin`, `cos`, `tan`, `sqrt`, `exp`, `radians`, `degrees` **only** if they:
1. Already appear in the blueprint, and
2. Have been observed to work in Home Assistant.

If in doubt, avoid them.

### YAML & Schema Rules

- Each blueprint file must be a **single YAML document** — no extra `---` separators.
- `variables:` must be at the **top level** (same indentation as `blueprint`, `trigger`, `action`).
- `!input` tags must not be embedded inside `{{ ... }}`. Bind inputs to variables first:

  ```yaml
  variables:
    cooling_setpoint: !input cooling_setpoint
    cooling_needed: >
      {{ indoor_temp is not none and indoor_temp >= cooling_setpoint }}
  ```

- `selector: target` returns dictionaries and must not be used directly in `entity_id:` of triggers; use `selector: entity` for entities expected by triggers.

### Safe Edit Strategy

Before editing:

- Skim the blueprint to see what filters and patterns are already in use.
- Mimic existing style and math patterns.
- Keep changes minimal and local.

During editing:

- Do not refactor working math into “cleaner” but unsupported forms.
- Do not add loops or deeply nested Jinja unless absolutely required.
- Respect variable ordering; if `A` depends on `B`, `B` must be defined first.

After editing:

- Ensure indentation, quoting, and YAML structure are valid.
- Run a quick YAML parse locally if possible.
- Import into HA and check Traces/Logs for template errors.

---

## Architecture Deep Dive

*(This section remains as in the original WARP.md, with details on Adaptive Comfort Control’s pipeline, regional presets, thermostat vendor detection, risk-aware pause, psychrometrics, etc. Keep this content up to date as blueprints evolve.)*

[Retain your existing detailed subsections here, including “Blueprint Structure”, “Key Architectural Patterns”, and “Common Pitfalls”, as they already provide valuable reference. If new patterns emerge in `adaptive-shades/` or others, add them under new subsections.]

---

## Common Tasks

### Add a new blueprint input

1. Define in `blueprint.input.<group>.input.<name>` with selector.
2. Bind to variable: `my_input: !input my_setting`.
3. Use in calculations: `{{ my_input | float(default) }}`.
4. Update CHANGELOG.md.
5. Increment `blueprint_version`.

### Add a new thermostat vendor

1. Add detection logic in `thermostat_profile_auto` (check manufacturer/model/name).
2. Add minimum separation in `vendor_sep_cli_profile`.
3. Test with both °C and °F climate entities.

### Add a new shade/comfort feature (Adaptive Shades)

1. Decide whether behavior should apply to **slat**, **zebra**, or both.
2. Add input(s) under an appropriate group (e.g., `core`, `sensors_and_targets`, `scheduling_and_overrides`).
3. Bind to new variables in `variables:` and integrate with existing geometry/comfort logic.
4. Respect façade-based geometry (sun azimuth/elevation + window orientation).
5. Test in Home Assistant with both simple and advanced configurations (with/without sensors).

---

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
awk '/^blueprint:/{p=1} /^variables:/{p=0} p' adaptive-comfort-control/adaptive_comfort_control_pro_blueprint.yaml
```

**Extract variables section:**
```bash
awk '/^variables:/{p=1} /^trigger:/{p=0} p' adaptive-comfort-control/adaptive_comfort_control_pro_blueprint.yaml
```

**Validate blueprint (REQUIRED):**
```bash
# Validate specific blueprint
python3 scripts/validate-blueprint.py adaptive-comfort-control/adaptive_comfort_control_pro_blueprint.yaml

# Validate all blueprints
python3 scripts/validate-blueprint.py --all
```

**Bump version (macOS sed):**
```bash
NEW_VERSION="4.12"
sed -E -i '' "s/^([[:space:]]*blueprint_version:[[:space:]]*).*/\1"${NEW_VERSION}"/" adaptive-comfort-control/adaptive_comfort_control_pro_blueprint.yaml
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

---

## Blueprint Distribution

**Blueprints are distributed via GitHub URLs:**

1. Push to GitHub.
2. Users import via raw GitHub URL in Home Assistant UI.
3. Home Assistant caches the blueprint; users manually click "Update" to get new versions.

**Testing locally:**
- Copy blueprint YAML to Home Assistant config: `config/blueprints/automation/<category>/`.
- Restart Home Assistant or reload automations.
- Blueprint appears in UI.

---

## Jinja2 Template Reference (Quick)

Commonly used patterns in these blueprints:

```jinja2
{# Comments #}

{# State access #}
{{ states('sensor.temperature') }}
{{ state_attr('sensor.temp', 'unit_of_measurement') }}
{{ is_state('binary_sensor.door', 'on') }}

{# Type conversion #}
{{ value | float(0) }}
{{ value | int(0) }}
{{ value | round(2) }}

{# String operations #}
{{ string | lower }}
{{ 'substring' in string }}

{# Lists and filters #}
{{ expand(entity_list) | selectattr('state','eq','on') | list | count }}

{# Conditionals #}
{% if condition %} ... {% elif ... %} ... {% else %} ... {% endif %}
{{ value if condition else other }}

{# None/null checks #}
{{ value is none }}
{{ value is not none }}
```

---

## Further Reading

- Home Assistant Blueprint docs: https://www.home-assistant.io/docs/automation/using_blueprints/
- Home Assistant templating docs: https://www.home-assistant.io/docs/configuration/templating/
- Semantic Versioning: https://semver.org/
