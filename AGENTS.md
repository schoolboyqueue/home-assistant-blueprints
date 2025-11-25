# Repository Guidelines

## Project Structure & Module Organization
- Blueprints live one per folder at the repo root: `adaptive-comfort-control/`, `bathroom-light-fan-control/`, `multi-switch-light-control/`, `adaptive-shades/`.
- Each blueprint folder contains the YAML (`*_blueprint.yaml` or `*_pro.yaml`), a `README.md`, and `CHANGELOG.md`. Treat each folder as a self-contained module with its own version and docs.
- Cross-blueprint guidance sits in `README.md` (overview) and `WARP.md` (agent workflow). No build artifacts or generated files are committed.

## Build, Test, and Development Commands
- There is no build step; edit YAML directly.
- Manual sanity check: `git diff --stat` to verify touched files are limited to the intended blueprint.
- Optional lint if available locally: `yamllint adaptive-comfort-control/adaptive_comfort_control_pro_blueprint.yaml` (or the file you edited).
- Manual runtime test happens in Home Assistant: import the Raw YAML URL, create an automation, and review Traces/Logs.

## Home Assistant Blueprint Schema Pitfalls
- `variables:` must be at the root of the YAML (same level as `blueprint`, `trigger`, `action`), not nested under `blueprint:`—otherwise import fails with “extra keys not allowed @ data['blueprint']['variables']”.
- Avoid raw bitwise operators in Jinja (`&`), which can trigger template parse errors in Home Assistant. Use filters like `bitwise_and` instead (e.g., `{{ (supported_features | int(0) | bitwise_and(16)) > 0 }}`) when checking feature flags.
- YAML must contain a **single document**. Do not include more than one `---` document separator in blueprint files.
- Home Assistant’s Jinja environment is restricted. Do not assume arbitrary Jinja extensions or Python are available.
 - HA selectors: when using `selector.select`, `options` should be simple strings (quote them, e.g., `- "off"`) or full `label`/`value` pairs. Avoid mixed/partial structures; malformed options will cause import errors.

## Home Assistant Jinja2 Compliance

You are editing **Home Assistant** automations and blueprints written in **YAML** and using **Home Assistant’s Jinja2 templating**. All templates must be valid for Home Assistant’s implementation, not generic Jinja or Python.

### General Rules
- Templates are evaluated by Home Assistant, not a generic Jinja engine.
- Use `{{ ... }}` for expressions and `{% ... %}` for control flow.
- Do **not** use:
  - Python imports (`import math`, `from ... import ...`)
  - Python calls (`math.sin(...)`, `pow(...)`, etc.)
  - Jinja macros/imports (`macro`, `import`, `include`, `call`)

### Allowed Filters (Baseline)
You may safely use these filters unless the user’s blueprint suggests otherwise:

- Type/utility filters:
  - `int`, `float`, `string`
  - `round`, `abs`
  - `default`
  - `lower`, `upper`, `title`, `capitalize`
  - `replace`
  - `regex_replace`
  - `length`

- List helpers already used in this repo (if present in the file you’re editing):
  - `min`, `max`, `sort`, `unique`, etc., only when they are already part of the existing blueprint.

If you are not sure a filter exists, **do not invent it**.

### Math & Trig Filters

Do **not** assume advanced filters exist. The safest approach:

- Prefer plain arithmetic using operators: `+`, `-`, `*`, `/`.
- Prefer **degree/radian conversion by hand**, instead of relying on `radians` or `degrees`.

Use these patterns:

```jinja2
{# Degrees → radians #}
{{ (angle_deg | float * pi / 180) }}

{# Radians → degrees #}
{{ (angle_rad | float * 180 / pi) }}
```

Only use math filters like `sin`, `cos`, `tan`, `sqrt`, `exp`, `radians`, `degrees` if:
1. They already appear in the current blueprint, and
2. The user has confirmed they work in their HA environment.

Otherwise, avoid them.

### Explicit “Do Not Use” List

Do **not** generate templates that use:

- Python-style calls: `math.sin(...)`, `pow(a, b)`, etc.
- Custom or unknown filters: `hypot`, `cosh`, `sinh`, `clip`, `pow10`, `exp2`, etc.
- Jinja extensions: `macro`, `call`, `import`, `include` (beyond HA’s standard behavior).
- `!input` tags directly inside `{{ ... }}` blocks (bind them to variables first, then use the variables).

### Handling Existing Templates

When editing existing Jinja templates:

1. Preserve working expressions and filters unless explicitly asked to refactor them.
2. If a blueprint already uses filters like `radians`, `exp`, `sin`, `cos`, you may keep them exactly as shown.
3. Do not “optimize” a working expression by replacing supported math with unsupported filters.

### Example: Safe vs Unsafe

Safe in Home Assistant:

```jinja2
cos_az_offset: "{{ ((sun_azimuth - window_orientation_deg) | float * pi / 180) | cos }}"
```

Unsafe (do not generate):

```jinja2
cos_az_offset: "{{ (sun_azimuth - window_orientation_deg) | radians | cos }}"
cos_az_offset: "{{ math.cos(math.radians(sun_azimuth - window_orientation_deg)) }}"
```

If unsure, prefer the explicit arithmetic pattern.

## Coding Style & Naming Conventions
- YAML with 4-space indentation, double-quoted strings when needed, and folded blocks (`>`) for long descriptions.
- Inputs are grouped under logical sections (`core`, `comfort`, `optional_sensors`, etc.); preserve ordering because later variables depend on earlier ones.
- Blueprint filenames use kebab-case folders with snake_case filenames ending in `_pro` or `_pro_blueprint`.
- When changing behavior, bump the blueprint version string in both `blueprint_version` and `blueprint.name` inside the YAML, and keep semantics aligned with the changelog entry.

## Testing Guidelines
- No automated tests; rely on Home Assistant import and automation runs.
- Prefer adding new optional inputs guarded with Jinja checks and fallbacks, to avoid breaking existing automations.
- When adjusting unit-sensitive logic, validate both °C and °F paths with real device states in Traces.

## Commit & Pull Request Guidelines
- Use Conventional Commits (e.g., `feat(adaptive-comfort): add window-open pause`, `fix(bathroom): clamp humidity delta`). Mark breaking changes with `!`.
- For any blueprint change, also update its `CHANGELOG.md` with a dated entry and note the SemVer bump (MAJOR for breaking, MINOR for features, PATCH for fixes).
- PRs should include: scope of change, blueprint(s) touched, before/after behavior, and trace/log snippets if fixing runtime issues. Keep diffs limited to the affected blueprint folder plus shared docs when necessary.

## Security & Configuration Tips
- Do not commit secrets or environment-specific entity IDs; keep examples generic (`sensor.example_temp`).
- Respect device rate limits and quantization; avoid adding rapid-fire service calls without delays.
