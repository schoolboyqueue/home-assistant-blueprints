# Repository Guidelines

## Project Structure & Module Organization
- Blueprints live one per folder at the repo root: `adaptive-comfort-control/`, `bathroom-light-fan-control/`, `multi-switch-light-control/`.
- Each blueprint folder contains the YAML (`*_blueprint.yaml` or `*_pro.yaml`), a `README.md`, and `CHANGELOG.md`. Treat each folder as a self-contained module with its own version and docs.
- Cross-blueprint guidance sits in `README.md` (overview) and `WARP.md` (agent workflow). No build artifacts or generated files are committed.

## Build, Test, and Development Commands
- There is no build step; edit YAML directly.
- Manual sanity check: `git diff --stat` to verify touched files are limited to the intended blueprint.
- Optional lint if available locally: `yamllint adaptive-comfort-control/adaptive_comfort_control_pro_blueprint.yaml` (or the file you edited).
- Manual runtime test happens in Home Assistant: import the Raw YAML URL, create an automation, and review Traces/Logs.

## Coding Style & Naming Conventions
- YAML with 4-space indentation, double-quoted strings when needed, and folded blocks (`>`) for long descriptions.
- Inputs are grouped under logical sections (`core`, `comfort`, `optional_sensors`, etc.); preserve ordering because later variables depend on earlier ones.
- Blueprint filenames use kebab-case folders with snake_case filenames ending in `_pro` or `_pro_blueprint`.
- When changing behavior, bump the blueprint version string in both `blueprint_version` and `blueprint.name` inside the YAML, and keep semantics aligned with the changelog entry.

## Testing Guidelines
- No automated tests; rely on Home Assistant import and automation runs.
- Prefer adding new optional inputs guarded with Jinja checks (`{% if input is defined %}`) to avoid breaking existing automations.
- When adjusting unit-sensitive logic, validate both °C and °F paths with real device states in Traces.

## Commit & Pull Request Guidelines
- Use Conventional Commits (e.g., `feat(adaptive-comfort): add window-open pause`, `fix(bathroom): clamp humidity delta`). Mark breaking changes with `!`.
- For any blueprint change, also update its `CHANGELOG.md` with a dated entry and note the SemVer bump (MAJOR for breaking, MINOR for features, PATCH for fixes).
- PRs should include: scope of change, blueprint(s) touched, before/after behavior, and trace/log snippets if fixing runtime issues. Keep diffs limited to the affected blueprint folder plus shared docs when necessary.

## Security & Configuration Tips
- Do not commit secrets or environment-specific entity IDs; keep examples generic (`sensor.example_temp`).
- Respect device rate limits and quantization; avoid adding rapid-fire service calls without delays.
