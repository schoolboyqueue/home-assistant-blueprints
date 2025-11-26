# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Repository Overview

This repository contains production-ready **Home Assistant Blueprints** — reusable automation templates written in YAML with Jinja2 templating. Users import these blueprints into their Home Assistant instances and configure them through the UI.

### Key Characteristics

- **No build system or CI/CD** — blueprints are edited directly as YAML files
- **No automated tests** — validation happens by importing into Home Assistant and checking traces/logs
- **Distributed via GitHub URLs** — users import raw YAML URLs directly into Home Assistant
- **Semantic Versioning per blueprint** — each blueprint has independent version tracking
- **Conventional Commits required** — enforced for all changes

### Active Blueprints

1. **Adaptive Comfort Control Pro** (`adaptive-comfort-control/`) — ASHRAE-55 adaptive comfort HVAC automation with psychrometrics, vendor profiles, CO₂ ventilation, and intelligent pause logic
2. **Bathroom Light Fan Control Pro** (`bathroom-light-fan-control/`) — Coordinated bathroom light/fan with "Wasp-in-a-Box" occupancy, humidity delta/rate detection, and night mode
3. **Multi Switch Light Control Pro** (`multi-switch-light-control/`) — Universal switch adapter for Zooz/Inovelli Central Scene and Lutron Pico remotes with auto-detection
4. **Adaptive Shades Pro** (`adaptive-shades/`) — Solar-aware shade positioning using facade geometry, direct/diffuse detection, and temperature-based behavior modes

Each blueprint folder contains: `*_pro_blueprint.yaml` or `*_pro.yaml`, `README.md`, `CHANGELOG.md`

## Development Workflow

### No Build Commands

There is no build, compile, or package step. Development consists of:

1. Edit YAML blueprints directly
2. **Validate locally** (REQUIRED): `python3 scripts/validate-blueprint.py <blueprint.yaml>`
3. Test by importing into Home Assistant and creating an automation
4. Check traces: Settings → Automations & Scenes → [Automation] → Traces
5. Monitor logs: Developer Tools → Logs

**Blueprint Validation (REQUIRED before commit):**
```bash
# Validate single blueprint after editing
python3 scripts/validate-blueprint.py multi-switch-light-control/multi_switch_light_control_pro.yaml

# Validate ALL blueprints before committing
python3 scripts/validate-blueprint.py --all
```

The validator checks for:
- YAML syntax errors
- Missing required keys (`blueprint`, `trigger`, `action`)
- `variables` nested under `blueprint` (common error)
- Empty/None `data:` blocks in service calls
- `!input` tags inside `{{ }}` templates
- Unbalanced template braces
- Invalid selector types
- Proper indentation in `if`/`then`/`else` blocks

See [scripts/README.md](scripts/README.md) for full documentation.

**Manual sanity check before commit:**
```bash
# Verify touched files are limited to intended blueprint
git diff --stat
```

### Version Management

All blueprints follow **Semantic Versioning** (MAJOR.MINOR.PATCH):
- **MAJOR**: Breaking changes (removing/renaming inputs, incompatible behavior changes)
- **MINOR**: New features (backward-compatible additions)
- **PATCH**: Bug fixes only

**When bumping a blueprint version:**

1. Update `blueprint_version` variable in the YAML (e.g., `blueprint_version: "1.11.0"`)
2. Update `blueprint.name` to include new version (e.g., `name: "Blueprint Name v1.11.0"`)
3. Add entry to blueprint's `CHANGELOG.md` with date and categorized changes:
   - `### Added` — New features
   - `### Changed` — Changes in existing functionality
   - `### Deprecated` — Soon-to-be removed features
   - `### Removed` — Removed features (breaking change)
   - `### Fixed` — Bug fixes
   - `### Security` — Security fixes
4. Update blueprint's `README.md` if the version is referenced there
5. Commit with conventional commit format:
   - `feat(component): description` for MINOR
   - `fix(component): description` for PATCH
   - `feat(component)!: description` for MAJOR breaking changes

### Agent Wrap-up Expectations (Auto-commit & Push)

When modifying any blueprint YAML (or its README/CHANGELOG), by default:

1. Determine the correct SemVer bump (MAJOR/MINOR/PATCH) based on the change
2. Update `blueprint_version` and `blueprint.name` to the new version
3. Update the corresponding `CHANGELOG.md` with a dated entry summarizing the change
4. Update the blueprint `README.md` version header to match
5. Run a basic YAML syntax check (optional but preferred)
6. Stage the relevant files, create a conventional commit, and push to the default branch

This should happen automatically as part of the workflow without needing an explicit user request.

### Common Commands

**Git/Repository:**
```bash
# View recent commits
git log --oneline --graph -20

# Check for uncommitted changes
git status
git diff

# Find all blueprint versions
git grep -n "blueprint_version:"

# Find all references to a variable in a blueprint
git grep -n "variable_name" adaptive-comfort-control/
```

**Blueprint Development:**
```bash
# Extract input section from blueprint
awk '/^blueprint:/{p=1} /^variables:/{p=0} p' adaptive-comfort-control/adaptive_comfort_control_pro_blueprint.yaml

# Extract variables section
awk '/^variables:/{p=1} /^trigger:/{p=0} p' adaptive-comfort-control/adaptive_comfort_control_pro_blueprint.yaml

# Bump version (macOS sed example)
NEW_VERSION="4.12"
sed -E -i '' "s/^([[:space:]]*blueprint_version:[[:space:]]*).*/\1\"${NEW_VERSION}\"/" adaptive-comfort-control/adaptive_comfort_control_pro_blueprint.yaml

# Find psychrometric calculations
git grep -nEi "dew|enthalpy|humidity|psychrom" adaptive-comfort-control/

# Find regional preset logic
git grep -nEi "region|preset|climate.*zone" adaptive-comfort-control/
```

## Home Assistant YAML & Jinja2 Critical Rules

### YAML Structure Requirements

- **Single document only** — no multiple `---` separators in blueprint files
- **`variables:` must be at root level** — same indentation as `blueprint`, `trigger`, `action`
  - ❌ WRONG: nested under `blueprint:` (causes "extra keys not allowed @ data['blueprint']['variables']" error)
  - ✅ CORRECT: top-level key alongside `blueprint:`
- **Variable ordering matters** — later variables can depend on earlier ones; order is critical
- **Selector types matter:**
  - Use `selector: entity` for triggers (returns entity_id string)
  - Use `selector: target` for service call targets (returns dictionary with entity_id/area_id/device_id)
  - `selector: target` cannot be used directly in `entity_id:` of triggers
- **Selector options formatting:**
  - Use quoted strings (`- "off"`) **OR** full `label`/`value` pairs (both quoted)
  - Don't mix styles or leave values unquoted (causes import errors)
  - Example: `options: ["off", "basic", "verbose"]` or `options: [{label: "Off", value: "off"}]`

### Jinja2 Template Constraints

Home Assistant uses a **restricted Jinja2 environment**, not full Python or generic Jinja2:

**Do NOT use:**
- Python imports or calls: `import math`, `math.sin()`, `pow()`, etc.
- Jinja macros/extensions: `macro`, `import`, `include`, `call`
- Unsupported filters: `hypot`, `cosh`, `sinh`, `clip`, etc.
- Bitwise operators directly in templates: use `bitwise_and` filter instead of `&`
- `!input` tags inside `{{ ... }}` blocks — bind to variables first

**Safe baseline filters:**
- Type conversion: `int`, `float`, `string`, `round`, `abs`, `default`
- String: `lower`, `upper`, `title`, `capitalize`, `replace`, `regex_replace`, `length`
- List: `min`, `max`, `sort`, `unique` (if already present in file)

**Math & trigonometry:**
- Prefer explicit arithmetic: `+`, `-`, `*`, `/`
- Convert degrees/radians manually:
  ```jinja2
  {# Degrees → radians #}
  {{ (angle_deg | float * pi / 180) }}

  {# Radians → degrees #}
  {{ (angle_rad | float * 180 / pi) }}
  ```
- Only use `sin`, `cos`, `tan`, `sqrt`, `exp`, `radians`, `degrees` if they already exist in the blueprint

**Safe Jinja patterns:**
```jinja2
{# Comments #}
{# This is a comment #}

{# State access #}
{{ states('sensor.temperature') }}
{{ state_attr('sensor.temp', 'unit_of_measurement') }}
{{ is_state('binary_sensor.door', 'on') }}

{# None checks (critical for optional sensors) #}
{{ value is none }}
{{ value is not none }}
{{ value not in ['unavailable', 'unknown', 'none'] }}

{# Type conversion with defaults #}
{{ value | float(0) }}
{{ value | int(0) }}
{{ value | round(2) }}

{# Time deltas (convert to timestamps first) #}
{{ (as_timestamp(now()) - as_timestamp(state.last_changed)) / 60 }}

{# String operations #}
{{ string | lower }}
{{ string | upper }}
{{ 'substring' in string }}

{# Lists and filters #}
{{ expand(entity_list) | selectattr('state','eq','on') | list | count }}

{# Conditionals #}
{% if condition %} ... {% elif ... %} ... {% else %} ... {% endif %}
{{ value if condition else other }}

{# Bitwise operations (checking feature flags) #}
{{ (supported_features | int(0) | bitwise_and(16)) > 0 }}
```

**Handling Existing Templates:**

When editing existing Jinja templates:
1. Preserve working expressions and filters unless explicitly asked to refactor
2. If a blueprint already uses filters like `radians`, `exp`, `sin`, `cos`, keep them exactly as shown
3. Do not "optimize" a working expression by replacing supported math with unsupported filters
4. Mimic existing style and math patterns in the file

### Common Pitfalls

1. **Extra keys error on import** — `variables:` nested under `blueprint:` instead of at root level (error: "extra keys not allowed @ data['blueprint']['variables']")
2. **Bitwise operator errors** — raw `&` in Jinja triggers parse errors; use `bitwise_and` filter instead
3. **Multi-entity inputs** — when input allows multiple entities, derive a single representative (e.g., first entity) before calling `state_attr()`, `is_state()`, or accessing `last_changed`
4. **Time delta type errors** — convert datetimes to timestamps with `as_timestamp()` before subtraction: `(as_timestamp(now()) - as_timestamp(state.last_changed)) / 60`
5. **Selector options malformed** — use quoted strings (`- "off"`) or full `label`/`value` pairs (both quoted); don't mix styles
6. **Template guards for optional sensors** — always check with `is not none` and `not in ['unavailable', 'unknown', 'none']`
7. **Variable scoping in loops** — variables set inside loops don't persist; use `namespace()` for mutable state
8. **!input in templates** — do not embed `!input` tags inside `{{ ... }}` blocks; bind them to variables first, then use the variables

## Making Changes to Blueprints

### Coding Style & Naming Conventions

- **Indentation:** 4-space indentation for YAML (use spaces, not tabs)
- **Strings:** Double-quoted when needed, folded blocks (`>`) for long descriptions
- **Input grouping:** Inputs are grouped under logical sections (`core`, `comfort`, `optional_sensors`, etc.)
- **Blueprint filenames:** Use kebab-case folders with snake_case filenames ending in `_pro` or `_pro_blueprint`
- **Variable naming:** snake_case for variable names
- **Preserve ordering:** Inputs and variables have dependencies; maintain order when editing

### Backward Compatibility First

- Existing automations must continue working after updates
- Use template guards for new optional inputs with fallbacks
- Preserve variable ordering — later variables depend on earlier definitions
- Test both unit paths (°C and °F) when touching temperature logic
- Avoid refactoring working math into "cleaner" but unsupported forms
- Keep changes minimal and local

### Unit Handling

Many blueprints support mixed °C/°F environments:
- Auto-detect units from sensor `unit_of_measurement` attributes
- Validate conversions and quantization against device step sizes
- Test with both metric and imperial sensor configurations

### Debug Logging

All blueprints support debug levels (`off`, `basic`, `verbose`):
- `off`: No debug output
- `basic`: Key events (state changes, mode switches, manual overrides)
- `verbose`: Detailed calculations, sensor values, condition breakdowns

Enable via blueprint's Debug/Diagnostics section in the UI.

### Troubleshooting & Debugging Tips

- **Enable debug level** (basic or verbose) to see key decisions and sensor values in traces
- **Units matter** — internal calculations often in °C; thermostats may be °F. Ensure conversions and quantization align with device step sizes
- **Optional sensors** — guard against `unavailable`/`unknown`/`none` states with template checks
- **State/trigger race conditions** — add small delays (e.g., 100ms) if a state change should settle before the next action
- **Use entity selectors (not target selectors) for triggers** — target selectors return dictionaries unsuitable for trigger `entity_id:`
- **Use `!input` directly in trigger `for:` durations** — bind to variable for calculations, but can use directly in trigger config
- **Check automation traces** — Settings → Automations & Scenes → [Automation] → Traces shows execution path and variable values
- **Monitor logs** — Developer Tools → Logs shows template errors and warnings

### Security & Device Limits

- **No secrets in code** — do not commit environment-specific entity IDs or API keys; keep examples generic (`sensor.example_temp`)
- **Respect device rate limits** — avoid rapid-fire service calls without delays
- **Quantization matters** — align temperature steps with device capabilities (e.g., 0.5°C for some thermostats, 1°F for others)
- **Vendor limits** — respect manufacturer-specific constraints (e.g., minimum separation between heat/cool setpoints)
- **Unit correctness** — ensure conversions are accurate and tested with real device states

## Common Tasks

### Add a New Blueprint Input

1. Define in `blueprint.input.<group>.input.<name>` with appropriate selector
2. Bind to variable in `variables:` section: `my_input: !input my_setting`
3. Use in calculations with type conversion: `{{ my_input | float(default) }}`
4. Add template guards for optional inputs: `{{ my_input is not none }}`
5. Update blueprint `CHANGELOG.md` with the addition
6. Increment `blueprint_version` (MINOR bump for new feature)
7. Update `README.md` if the input affects user-facing documentation

### Add a New Thermostat Vendor (Adaptive Comfort)

1. Add detection logic in `thermostat_profile_auto` variable
2. Check manufacturer/model/name attributes from climate entity
3. Add corresponding minimum separation in `vendor_sep_cli_profile` variable
4. Test with both °C and °F climate entities
5. Document in `CHANGELOG.md` under "Added" section

### Add a New Shade Feature (Adaptive Shades)

1. Decide whether behavior applies to **slat**, **zebra**, or both modes
2. Add input(s) under appropriate group (`core`, `sensors_and_targets`, `scheduling_and_overrides`)
3. Bind to new variables in `variables:` section
4. Integrate with existing geometry/comfort logic (respect façade-based sun azimuth/elevation)
5. Test with both simple and advanced configurations (with/without optional sensors)
6. Add entry to `CHANGELOG.md` and bump version

## Blueprint Distribution

**Blueprints are distributed via GitHub URLs:**

1. Push changes to GitHub (main branch)
2. Users import via raw GitHub URL in Home Assistant UI: Settings → Automations & Scenes → Blueprints → Import Blueprint
3. Home Assistant caches the blueprint; users must manually click "Update" to get new versions
4. Raw URL format: `https://raw.githubusercontent.com/schoolboyqueue/home-assistant-blueprints/main/<folder>/<file>.yaml`

**Testing locally (before pushing):**

- Copy blueprint YAML to Home Assistant config directory: `config/blueprints/automation/<category>/`
- Restart Home Assistant or reload automations
- Blueprint appears in UI for testing

## Git Commit Process

### Conventional Commits Required

Format: `<type>(<scope>): <description>`

**Types:**
- `feat`: New feature (MINOR version bump)
- `fix`: Bug fix (PATCH version bump)
- `docs`: Documentation only
- `chore`: Maintenance, no version bump
- Breaking changes: add `!` (e.g., `feat(bathroom)!: remove deprecated input`)

**Scopes:** Use blueprint folder name (`adaptive-comfort`, `bathroom`, `multi-switch`, `adaptive-shades`)

**Examples:**
```
feat(adaptive-comfort): add window-open pause detection
fix(bathroom): clamp humidity delta to prevent overflow
docs(readme): clarify import instructions
chore(adaptive-shades): refactor geometry calculations
```

### When Making a Change

1. Determine version increment (MAJOR/MINOR/PATCH)
2. Update `blueprint_version` and `blueprint.name` in YAML
3. Add dated entry to `CHANGELOG.md` with categorized sections (Added/Changed/Fixed/Removed)
4. Update `README.md` if version is referenced
5. Commit with conventional format
6. Push to main branch

## Architecture Notes

### Adaptive Comfort Control Pipeline

1. **Unit detection** — auto-detect from sensor `unit_of_measurement` or use override
2. **Regional presets** — apply seasonal bias, ventilation, psychrometric defaults by U.S. state/region
3. **Vendor profile detection** — auto-detect thermostat manufacturer from entity attributes
4. **Psychrometrics** — calculate dew point, absolute humidity, enthalpy from temperature and RH
5. **Comfort band calculation** — ASHRAE-55 adaptive comfort with category tolerance (I/II/III)
6. **Risk-aware pause** — intelligent HVAC pause with acceleration based on delta from comfort band
7. **Command optimization** — skip unnecessary service calls (50-80% reduction) by comparing current vs target state

### Adaptive Shades Geometry

- **Facade-based detection** — sun azimuth/elevation vs window orientation determines direct vs diffuse exposure
- **Clear-sky irradiance** — compare measured vs theoretical to detect clouds
- **Temperature bands** — winter (admit heat), intermediate (balanced), summer (block heat)
- **Presence-aware** — different behavior when occupied vs unoccupied
- **Manual override** — contact sensor or automation-triggered input_boolean to suspend automation

### Multi-Switch Auto-Detection

- **Hardware detection** — identify Zooz/Inovelli vs Lutron from event data structure
- **Unified dimming** — same step size and transition settings across all switch types
- **Central Scene events** — handle both `zwave_js_event` and `zwave_js_value_notification`

### Bathroom Light Fan Control

- **Wasp-in-a-Box occupancy** — binary sensor-based occupancy detection
- **Humidity delta** — compare bathroom humidity to baseline (e.g., another room or outdoor)
- **Rate-of-change detection** — trigger fan based on rising or falling humidity rate
- **Night mode** — reduced brightness during quiet hours
- **Manual override** — respect user-initiated light/fan changes for a timeout period

## Keeping Documentation Up-to-Date

**IMPORTANT:** When working on this repository, proactively update documentation when you discover new patterns, pitfalls, or solutions:

- **Update CLAUDE.md** when you find new common errors, architectural patterns, or clarify ambiguous instructions
- **Update WARP.md** when adding new blueprints (update blueprints list, file structure, and relevant sections)
- **Update AGENTS.md** when discovering new Home Assistant schema pitfalls or Jinja2 restrictions
- **Update blueprint README.md** when changing inputs, behavior, or adding features
- **Update CHANGELOG.md** for every blueprint change (required for version tracking)

This ensures the knowledge base grows continuously and future sessions benefit from past learnings.

**When adding a new blueprint, update:**
1. This CLAUDE.md file (Active Blueprints section)
2. WARP.md (Blueprints and File Structure sections)
3. Repository README.md (Blueprints Gallery section with import badge and links)

## Reference Documentation

For deeper architectural details, see:
- **WARP.md** — Complete workflow guidance for all agents (includes detailed subsections on each blueprint)
- **AGENTS.md** — Repository structure and Home Assistant Jinja2 compliance rules
- Individual blueprint READMEs for feature documentation
