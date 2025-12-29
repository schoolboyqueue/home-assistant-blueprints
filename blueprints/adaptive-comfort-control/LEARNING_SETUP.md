# Adaptive Learning Setup Guide

## Overview

The learning system tracks your manual temperature adjustments and adapts to your preferences over time. It stores **4 separate offsets**:

| Key | Description |
|-----|-------------|
| `heat_day` | Your heating preference during daytime |
| `heat_night` | Your heating preference during nighttime (sleep hours) |
| `cool_day` | Your cooling preference during daytime |
| `cool_night` | Your cooling preference during nighttime |

This means the system learns that you might want it warmer in the morning (heat_day) but cooler at night (cool_night), or any other combination of preferences.

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

### Time-of-Day Awareness

The system uses your **Sleep Schedule** settings to determine day vs night:

- **Night**: Between Sleep Start and Sleep End times
- **Day**: All other times

This means if you configure:

- Sleep Start: 22:00
- Sleep End: 07:00

Then adjustments made between 10 PM and 7 AM affect the `_night` offsets, and adjustments made between 7 AM and 10 PM affect the `_day` offsets.

### Heating vs Cooling Detection

| Trigger | Offset Updated |
|---------|----------------|
| You adjust `target_temp_low` (heat setpoint) | `heat_day` or `heat_night` |
| You adjust `target_temp_high` (cool setpoint) | `cool_day` or `cool_night` |
| Single setpoint + HVAC mode is "heat" | `heat_day` or `heat_night` |
| Single setpoint + HVAC mode is "cool" | `cool_day` or `cool_night` |

## Monitoring

### Check Current Learned Preferences

Go to **Developer Tools → States**, find `sensor.comfort_learning`, and look at the `variables` attribute:

```yaml
variables:
  heat_day: 1.5
  heat_night: 0.8
  cool_day: -0.5
  cool_night: -1.2
```

This example shows:

- Daytime heating: +1.5°F warmer than baseline
- Nighttime heating: +0.8°F warmer than baseline
- Daytime cooling: -0.5°F cooler than baseline (higher setpoint)
- Nighttime cooling: -1.2°F cooler than baseline

### Debug Logs

Enable **Debug Level: basic** or **verbose** to see learning in action:

```text
Manual override detected (climate_manual_change_low).
Manual=21.7°C, Predicted=20.0°C, Error=+1.7°C.
Learning [heat_day]: 0.50 → 0.76°F.
Pausing for 60 min (helper set).
```

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

| Rate | Speed | Stability | Best For |
|------|-------|-----------|----------|
| 0.05 | Very Slow | Very Stable | Highly consistent preferences |
| 0.10 | Slow | Stable | Most users |
| **0.15** | **Moderate** | **Balanced** | **Recommended default** |
| 0.20 | Fast | Moderate | Quickly adapting to new patterns |
| 0.30 | Very Fast | Less Stable | Experimental/testing |

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

**Q: What if I make random adjustments?**
The exponential average smooths out noise. Consistent patterns emerge over time.

**Q: Can I have different learning for different rooms?**
Yes! Create multiple sensors (e.g., `sensor.comfort_learning_bedroom`, `sensor.comfort_learning_living`) and configure each automation to use its own sensor.

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
3. Check that sleep schedule is correctly configured for day/night detection
