# Adaptive Learning Setup Guide

## Overview

The learning system tracks your manual temperature adjustments and adapts to your preferences over time. It stores **16 separate offsets** organized by:

- **Mode**: Heating vs Cooling
- **Time of Day**: Day vs Night (based on sun position)
- **Season**: Winter, Spring, Summer, or Autumn

| Key                 | Description                             |
| ------------------- | --------------------------------------- |
| `heat_day_winter`   | Heating preference during winter days   |
| `heat_night_winter` | Heating preference during winter nights |
| `cool_day_winter`   | Cooling preference during winter days   |
| `cool_night_winter` | Cooling preference during winter nights |
| `heat_day_spring`   | Heating preference during spring days   |
| `heat_night_spring` | Heating preference during spring nights |
| `cool_day_spring`   | Cooling preference during spring days   |
| `cool_night_spring` | Cooling preference during spring nights |
| `heat_day_summer`   | Heating preference during summer days   |
| `heat_night_summer` | Heating preference during summer nights |
| `cool_day_summer`   | Cooling preference during summer days   |
| `cool_night_summer` | Cooling preference during summer nights |
| `heat_day_autumn`   | Heating preference during autumn days   |
| `heat_night_autumn` | Heating preference during autumn nights |
| `cool_day_autumn`   | Cooling preference during autumn days   |
| `cool_night_autumn` | Cooling preference during autumn nights |

This allows the system to learn complex seasonal patterns. For example, you might prefer it cooler during summer nights but warmer during winter mornings, and these preferences will be applied automatically as conditions change.

## Setup Instructions

### Step 1: Create the Variables Sensor

Add this to your `configuration.yaml` (or a separate file included via `template:`):

```yaml
template:
  - trigger:
      - platform: event
        event_type: set_variable
      - platform: event
        event_type: remove_variable
      - platform: event
        event_type: clear_variables
    condition:
      - condition: template
        value_template: >
          {{
              (
                trigger.event.event_type == 'set_variable'
                and trigger.event.data is defined
                and trigger.event.data.key is defined
                and trigger.event.data.value is defined
              )
              or
              (
                trigger.event.event_type == 'remove_variable'
                and trigger.event.data is defined
                and trigger.event.data.key is defined
              )
              or 
              trigger.event.event_type == 'clear_variables'
          }}
    action:
      - if: "{{ trigger.event.data.get('log', state_attr('sensor.comfort_learning', 'log_events')) }}"
        then:
          - service: logbook.log
            data:
              name: "Comfort Learning"
              message: >
                {{ trigger.event.data.key | default('All variables removed') }}
                {%- if trigger.event.event_type == 'set_variable' -%}
                  : {{ trigger.event.data.value }}
                {%- endif -%}
    sensor:
      - unique_id: comfort_learning_prefs
        name: Comfort Learning
        state: "{{ this.attributes.variables | default({}) | length }} preferences"
        attributes:
          log_events: true
          variables: >
            {% set others = dict(this.attributes.get('variables', {}).items() | rejectattr('0', 'eq', trigger.event.data.key)) %}
            {% if trigger.event.event_type == 'set_variable' %}
              {% set new = {trigger.event.data.key: trigger.event.data.value} %}
              {{ dict(others, **new) }}
            {% elif trigger.event.event_type == 'remove_variable' %}
              {{ others }}
            {% elif trigger.event.event_type == 'clear_variables' %}
              {}
            {% endif %}
```

After adding this, restart Home Assistant or reload template entities.

### Step 2: Configure the Blueprint

In your automation:

1. **Manual Override Detection** section:
   - ✅ **Enable Manual Override Detection**: On
   - **Override Duration**: 60 minutes (adjust as needed)
   - ✅ **Learn from Manual Adjustments**: On
   - **Learning Rate**: 0.15 (start conservative)
   - **Learned Preferences Sensor**: Select `sensor.comfort_learning`

### Step 3: Verify Setup

1. Go to **Developer Tools → States**
2. Search for `sensor.comfort_learning`
3. You should see it with `variables: {}` in attributes (empty until you make adjustments)

## How It Works

### Learning Process

When you manually adjust the thermostat:

1. **Blueprint detects the change** (heating setpoint vs cooling setpoint)
2. **Determines time of day** (day vs night based on your sleep schedule)
3. **Calculates the error** (what you set - what we predicted)
4. **Updates the appropriate offset** using exponential moving average

Example:

```text
You adjust heating setpoint: 66°F → 70°F (during daytime)
Blueprint predicted: 66°F
Error: +4°F

Key updated: heat_day
Calculation: new = 0.85 * old + 0.15 * (+4°F)
           = 0.85 * 0 + 0.6
           = +0.6°F

Future daytime heating: now +0.6°F warmer
```

### Time-of-Day Awareness (Sun-Based)

The system uses the **Sun Entity** (configured in Optional Sensors) to determine day vs night:

- **Night**: When the sun is below the horizon
- **Day**: When the sun is above the horizon

This automatically adapts to seasonal daylight changes. During summer, "day" is longer; during winter, "night" is longer. Your preferences are learned and applied according to actual daylight conditions, not fixed times.

If no sun entity is configured, the system falls back to your Sleep Schedule times.

### Seasonal Awareness

The system detects the current season using either:

1. **Season Entity** (if configured in Seasonal Bias & Regional Presets)
2. **Month-based fallback**:
   - Dec, Jan, Feb = Winter
   - Mar, Apr, May = Spring
   - Jun, Jul, Aug = Summer
   - Sep, Oct, Nov = Autumn

Adjustments you make are stored with the current season, so preferences learned in summer are applied in summer, winter preferences in winter, etc.

### Heating vs Cooling Detection

| Trigger                                       | Offset Updated               |
| --------------------------------------------- | ---------------------------- |
| You adjust `target_temp_low` (heat setpoint)  | `heat_{day\|night}_{season}` |
| You adjust `target_temp_high` (cool setpoint) | `cool_{day\|night}_{season}` |
| Single setpoint + HVAC mode is "heat"         | `heat_{day\|night}_{season}` |
| Single setpoint + HVAC mode is "cool"         | `cool_{day\|night}_{season}` |

The `{day|night}` component is determined by sun position, and `{season}` is one of `summer`, `winter`, or `shoulder`.

## Monitoring

### Check Current Learned Preferences

Go to **Developer Tools → States**, find `sensor.comfort_learning`, and look at the `variables` attribute:

```yaml
variables:
  heat_day_winter: 1.8
  heat_night_winter: 1.2
  cool_day_winter: 0.0
  cool_night_winter: -0.3
  heat_day_spring: 1.0
  heat_night_spring: 0.6
  cool_day_spring: -0.5
  cool_night_spring: -0.8
  heat_day_summer: 0.5
  heat_night_summer: 0.3
  cool_day_summer: -1.2
  cool_night_summer: -1.5
  heat_day_autumn: 0.8
  heat_night_autumn: 0.5
  cool_day_autumn: -0.6
  cool_night_autumn: -0.9
```

This example shows seasonal preferences:

- **Winter**: Stronger heating offsets (+1.2 to +1.8°F) for warmth
- **Spring/Autumn**: Moderate adjustments as weather transitions
- **Summer**: Stronger cooling offsets (-1.2 to -1.5°F) for comfort in heat

### Debug Logs

Enable **Debug Level: basic** or **verbose** to see learning in action:

```text
Manual override detected (climate_manual_change_low).
Manual=21.7°C, Predicted=20.0°C, Error=+1.7°C.
Learning [heat_day_winter]: 0.50 → 0.76°F.
Pausing for 60 min (helper set).
```

The key now includes the season (e.g., `heat_day_winter`) for visibility into which seasonal slot is being updated.

### Dashboard Card (Optional)

```yaml
type: entities
title: Comfort Learning
entities:
  - entity: sensor.comfort_learning
    name: Preferences Count
  - type: attribute
    entity: sensor.comfort_learning
    attribute: variables
    name: Current Offsets
```

## Reset Learning

To reset all learned preferences:

1. **Developer Tools → Events**
2. Fire event: `clear_variables`
3. No event data needed

Or reset a single offset:

```yaml
event: remove_variable
event_data:
  key: heat_day
```

## Tuning Tips

### Learning Rate Guide

| Rate     | Speed        | Stability    | Best For                         |
| -------- | ------------ | ------------ | -------------------------------- |
| 0.05     | Very Slow    | Very Stable  | Highly consistent preferences    |
| 0.10     | Slow         | Stable       | Most users                       |
| **0.15** | **Moderate** | **Balanced** | **Recommended default**          |
| 0.20     | Fast         | Moderate     | Quickly adapting to new patterns |
| 0.30     | Very Fast    | Less Stable  | Experimental/testing             |

### Convergence Timeline

With **α = 0.15** (default):

- After 1 adjustment: ~15% adapted
- After 5 adjustments: ~56% adapted
- After 10 adjustments: ~80% adapted
- After 15 adjustments: ~93% adapted

## FAQ

**Q: What if I don't set up the sensor?**
Learning still works during the current session but resets when Home Assistant restarts.

**Q: Does this replace seasonal adaptation?**
No! Learned offsets add to your regional/seasonal biases. Both work together:

```text
Final Comfort = ASHRAE-55 + Regional Bias + Seasonal Bias + Learned Offset + Sleep Bias + CO₂ Bias
```

The learned offset is now season-aware, so it applies the right adjustment for the current season automatically.

**Q: What if I make random adjustments?**
The exponential average smooths out noise. Consistent patterns emerge over time.

**Q: Can I have different learning for different rooms?**
Yes! Create multiple sensors (e.g., `sensor.comfort_learning_bedroom`, `sensor.comfort_learning_living`) and configure each automation to use its own sensor.

**Q: How does sun-based day/night differ from the sleep schedule?**
The sun entity tracks actual sunrise/sunset times which change throughout the year. In summer, "day" might be 5 AM to 9 PM, while in winter it might be 7 AM to 5 PM. This means your daytime preferences are applied when there's actually daylight, not at fixed times.

**Q: I have old 4-key preferences from a previous version. What happens?**
Old 4-key preferences (`heat_day`, `heat_night`, etc.) are no longer used. You'll start fresh with the 16-key system. Your new preferences will be learned as you make manual adjustments throughout the year.

## Troubleshooting

**Learning not working:**

1. Check "Learn from Manual Adjustments" is enabled
2. Verify sensor entity exists (`sensor.comfort_learning`)
3. Ensure manual changes exceed tolerance (default: 1.0°)
4. Check debug logs for learning messages

**Sensor not updating:**

1. Verify the YAML configuration is correct
2. Check Home Assistant logs for template errors
3. Try firing a test event manually in Developer Tools

**Offset values seem wrong:**

1. Fire `clear_variables` event to reset
2. Reduce learning rate for more conservative updates
3. Check that sun entity is configured (defaults to `sun.sun`)
4. Verify season detection is correct (check `sensor.season` or month-based fallback)
