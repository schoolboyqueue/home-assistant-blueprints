<div align="center">
  <h1>Home Assistant Blueprints</h1>
  <p><em>Pro‑grade automation templates for real homes</em></p>
  <p>
    <a href="LICENSE">
      <img alt="MIT License" src="https://img.shields.io/badge/License-MIT-green.svg">
    </a>
    <a href="https://www.conventionalcommits.org/en/v1.0.0/">
      <img alt="Conventional Commits" src="https://img.shields.io/badge/Conventional%20Commits-1.0.0-yellow.svg?logo=conventionalcommits&logoColor=white">
    </a>
    <a href="https://semver.org/">
      <img alt="Semantic Versioning" src="https://img.shields.io/badge/SemVer-2.0.0-blue.svg?logo=semver&logoColor=white">
    </a>
    <a href="https://github.com/schoolboyqueue/home-assistant-blueprints/stargazers">
      <img alt="GitHub Stars" src="https://img.shields.io/github/stars/schoolboyqueue/home-assistant-blueprints?style=flat">
    </a>
    <a href="https://github.com/schoolboyqueue/home-assistant-blueprints/pulls">
      <img alt="PRs Welcome" src="https://img.shields.io/badge/PRs-welcome-brightgreen.svg">
    </a>
    <a href="https://buymeacoffee.com/schoolboyqueue">
      <img alt="Buy Me A Coffee" src="https://img.shields.io/badge/Buy%20Me%20A%20Coffee-%23FFDD00.svg?logo=buy-me-a-coffee&logoColor=black&label=Support">
    </a>
  </p>
  <p>
    <a href="https://buymeacoffee.com/schoolboyqueue"><b>☕ Like these? Buy me a coffee</b></a>
  </p>
</div>

---

## Overview

This repository hosts polished, production‑ready Home Assistant Blueprints that solve real household needs. Each blueprint is:
- **Robust:** battle‑tested logic with clear input validation and sane defaults
- **Efficient:** avoids unnecessary service calls, respects device limits, and reduces chatter
- **Configurable:** sensible presets with advanced controls when you want them
- **Documented:** README per blueprint, changelog, and conventional commits
- **MIT‑licensed:** permissive for personal and commercial use

Blueprints use YAML with Jinja2 templating and follow Semantic Versioning.

## Blueprints Gallery

Quickly explore and import blueprints. "Import" buttons link to the YAML on GitHub (blob). For direct import in Home Assistant, use the **Raw** link in Quick Links.

### 1) Adaptive Comfort Control Pro
Advanced HVAC automation implementing ASHRAE‑55 adaptive comfort with built‑in psychrometrics (dew point, absolute humidity, enthalpy), vendor auto‑profiles, seasonal bias, CO₂‑aware ventilation, and smart pause acceleration.

<a href="https://github.com/schoolboyqueue/home-assistant-blueprints/blob/main/adaptive-comfort-control/adaptive_comfort_control_pro_blueprint.yaml"><img alt="Import Adaptive Comfort" src="https://img.shields.io/badge/Import-Home%20Assistant-2D9BF0?logo=homeassistant&logoColor=white"></a>

**Quick Links:**
<a href="https://raw.githubusercontent.com/schoolboyqueue/home-assistant-blueprints/main/adaptive-comfort-control/adaptive_comfort_control_pro_blueprint.yaml">Raw</a> •
<a href="https://github.com/schoolboyqueue/home-assistant-blueprints/blob/main/adaptive-comfort-control/README.md">Docs</a> •
<a href="https://github.com/schoolboyqueue/home-assistant-blueprints/blob/main/adaptive-comfort-control/CHANGELOG.md">Changelog</a>

---

### 2) Bathroom Light Fan Control Pro
Coordinated bathroom light and fan automation with occupancy ("Wasp-in-a-Box"), humidity delta and rate-of-change detection, night mode, and manual override.

<a href="https://github.com/schoolboyqueue/home-assistant-blueprints/blob/main/bathroom-light-fan-control/bathroom_light_fan_control_pro.yaml"><img alt="Import Bathroom Light & Fan" src="https://img.shields.io/badge/Import-Home%20Assistant-2D9BF0?logo=homeassistant&logoColor=white"></a>

**Quick Links:**
<a href="https://raw.githubusercontent.com/schoolboyqueue/home-assistant-blueprints/main/bathroom-light-fan-control/bathroom_light_fan_control_pro.yaml">Raw</a> •
<a href="https://github.com/schoolboyqueue/home-assistant-blueprints/blob/main/bathroom-light-fan-control/README.md">Docs</a> •
<a href="https://github.com/schoolboyqueue/home-assistant-blueprints/blob/main/bathroom-light-fan-control/CHANGELOG.md">Changelog</a>

---

### 3) Multi Switch Light Control Pro
A single automation that adapts to Zooz/Inovelli Central Scene switches or Lutron Pico remotes, auto-detects the selected hardware, and unifies dimming/transition tuning across every trigger.

<a href="https://github.com/schoolboyqueue/home-assistant-blueprints/blob/main/multi-switch-light-control/multi_switch_light_control_pro.yaml"><img alt="Import Multi Switch Light Control" src="https://img.shields.io/badge/Import-Home%20Assistant-2D9BF0?logo=homeassistant&logoColor=white"></a>

**Quick Links:**
<a href="https://raw.githubusercontent.com/schoolboyqueue/home-assistant-blueprints/main/multi-switch-light-control/multi_switch_light_control_pro.yaml">Raw</a> •
<a href="https://github.com/schoolboyqueue/home-assistant-blueprints/blob/main/multi-switch-light-control/README.md">Docs</a> •
<a href="https://github.com/schoolboyqueue/home-assistant-blueprints/blob/main/multi-switch-light-control/CHANGELOG.md">Changelog</a>

---

### 4) Zooz Z-Wave Light Switch Control Pro
Z-Wave dimming via Central Scene events (ZEN71/72/76/77). Single press on/off, hold-to-dim with release detection, and configurable parameters. Supports `zwave_js_event` and `zwave_js_value_notification`.

<a href="https://github.com/schoolboyqueue/home-assistant-blueprints/blob/main/zooz-zwave-light-switch-control/zooz_zwave_light_switch_control_pro.yaml"><img alt="Import Zooz Z-Wave" src="https://img.shields.io/badge/Import-Home%20Assistant-2D9BF0?logo=homeassistant&logoColor=white"></a>

**Quick Links:**
<a href="https://raw.githubusercontent.com/schoolboyqueue/home-assistant-blueprints/main/zooz-zwave-light-switch-control/zooz_zwave_light_switch_control_pro.yaml">Raw</a> •
<a href="https://github.com/schoolboyqueue/home-assistant-blueprints/blob/main/zooz-zwave-light-switch-control/README.md">Docs</a> •
<a href="https://github.com/schoolboyqueue/home-assistant-blueprints/blob/main/zooz-zwave-light-switch-control/CHANGELOG.md">Changelog</a>

---

## Quick Start

1. Click an **Import** badge above to view the blueprint YAML on GitHub.
2. For Home Assistant's Import dialog, copy the **Raw** link for your chosen blueprint.
3. In Home Assistant: **Settings → Automations & Scenes → Blueprints → Import Blueprint** → paste the Raw URL.
4. Create an automation from the blueprint and configure inputs via the UI.
5. Check **Traces** and **Logs** if anything looks off.

## Design Principles

- **Safety first:** device vendor limits and unit correctness
- **Minimalism:** only the inputs you need, with clearly described defaults
- **Predictability:** deterministic logic and quantization to device step sizes
- **Observability:** debug levels (off/basic/verbose) and traceable decision paths
- **Performance:** avoid unnecessary service calls and rate-limit where needed

## Repository Structure

```
.
├── adaptive-comfort-control/
│   ├── adaptive_comfort_control_pro_blueprint.yaml
│   ├── CHANGELOG.md
│   └── README.md
├── bathroom-light-fan-control/
│   ├── bathroom_light_fan_control_pro.yaml
│   ├── CHANGELOG.md
│   └── README.md
├── multi-switch-light-control/
│   ├── multi_switch_light_control_pro_blueprint.yaml
│   ├── CHANGELOG.md
│   └── README.md
├── zooz-zwave-light-switch-control/
│   ├── zooz_zwave_light_switch_control_pro.yaml
│   ├── CHANGELOG.md
│   └── README.md
├── WARP.md
└── README.md (this file)
```

## Troubleshooting & Debugging

- Enable the blueprint's **Debug level** (basic or verbose) to see key decisions and sensor values.
- **Units matter:** internal calcs often in °C; thermostats may be °F. Ensure conversions and quantization align with device step sizes.
- **Optional sensors:** guard against `unavailable`/`unknown`/`none` states.
- **State/trigger race conditions:** add small delays (e.g., 100ms) if a change should settle before the next action.
- Use **entity selectors** (not target selectors) for triggers; use `!input` directly in trigger `for:` durations.

For deeper guidance, see [WARP.md](WARP.md).

## Contributing

Contributions are welcome!

**Conventional Commits required** (examples):
- `docs(readme): clarify quick start`
- `feat(bathroom): add night mode brightness floor`
- `fix(adaptive-comfort): correct °F step quantization`

**Semantic Versioning per blueprint:**
- **MAJOR:** breaking changes
- **MINOR:** new features
- **PATCH:** bug fixes

Open a PR with a clear description and tests/trace screenshots when relevant.

## License

MIT — see the [LICENSE](LICENSE) file. You can use these blueprints in personal or commercial projects.

## Support

If these blueprints help you, consider supporting:
- ☕ **Buy Me A Coffee:** https://buymeacoffee.com/schoolboyqueue

---

## Star History

[![Star History Chart](https://api.star-history.com/svg?repos=schoolboyqueue/home-assistant-blueprints&type=Date)](https://star-history.com/#schoolboyqueue/home-assistant-blueprints&Date)
