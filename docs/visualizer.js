/**
 * Blueprint Visualizer - Interactive Documentation System
 * Parses Home Assistant blueprint YAML files and generates visual learning materials
 *
 * Key design principles for human-friendly diagrams:
 * 1. Use natural language descriptions (e.g., "When motion detected" not "trigger.id == 'motion'")
 * 2. Group related functionality together (e.g., "Light Control", "Fan Control")
 * 3. Show the "why" not just the "what" (e.g., "Turn on fan to remove moisture")
 * 4. Use icons and colors consistently to indicate action types
 */

(function () {
  // Blueprint URLs mapping
  const BLUEPRINT_URLS = {
    'adaptive-comfort-control':
      'https://raw.githubusercontent.com/schoolboyqueue/home-assistant-blueprints/main/blueprints/adaptive-comfort-control/adaptive_comfort_control_pro_blueprint.yaml',
    'bathroom-light-fan-control':
      'https://raw.githubusercontent.com/schoolboyqueue/home-assistant-blueprints/main/blueprints/bathroom-light-fan-control/bathroom_light_fan_control_pro.yaml',
    'multi-switch-light-control':
      'https://raw.githubusercontent.com/schoolboyqueue/home-assistant-blueprints/main/blueprints/multi-switch-light-control/multi_switch_light_control_pro.yaml',
    'adaptive-shades':
      'https://raw.githubusercontent.com/schoolboyqueue/home-assistant-blueprints/main/blueprints/adaptive-shades/adaptive_shades_pro.yaml',
    'adaptive-fan-control':
      'https://raw.githubusercontent.com/schoolboyqueue/home-assistant-blueprints/main/blueprints/adaptive-fan-control/adaptive_fan_control_pro_blueprint.yaml',
  };

  // Human-friendly blueprint summaries - these provide context that's hard to extract from YAML
  const BLUEPRINT_SUMMARIES = {
    'adaptive-comfort-control': {
      purpose: 'Automatically controls HVAC to maintain optimal comfort based on temperature, humidity, and occupancy',
      mainFeatures: [
        'Smart temperature control',
        'Humidity management',
        'Occupancy-aware',
        'Energy efficient scheduling',
      ],
      userBenefit: 'Keep your home comfortable while saving energy',
    },
    'bathroom-light-fan-control': {
      purpose: 'Intelligently controls bathroom lights and exhaust fan based on motion, door, and humidity sensors',
      mainFeatures: [
        'Motion-activated lights',
        'Humidity-triggered fan',
        'Night mode dimming',
        'Wasp-in-a-Box occupancy',
      ],
      userBenefit: 'Hands-free bathroom automation that handles lights and moisture',
    },
    'multi-switch-light-control': {
      purpose: 'Advanced light control supporting multiple switches, scenes, and dimming options',
      mainFeatures: ['Multi-switch support', 'Scene activation', 'Dimming control', 'Time-based behavior'],
      userBenefit: 'Flexible lighting control with multiple switches',
    },
    'adaptive-shades': {
      purpose: 'Automatically adjusts window shades based on sun position, temperature, and preferences',
      mainFeatures: ['Sun tracking', 'Temperature-based adjustment', 'Schedule support', 'Manual override'],
      userBenefit: 'Optimize natural light and temperature automatically',
    },
    'adaptive-fan-control': {
      purpose: 'Smart ceiling fan control based on temperature and occupancy',
      mainFeatures: ['Temperature-based speed', 'Occupancy awareness', 'HVAC coordination', 'Seasonal modes'],
      userBenefit: 'Comfortable airflow that adapts to conditions',
    },
  };

  // Detailed trigger-to-action mappings for each blueprint
  // This provides explicit human-readable descriptions for each trigger's logic path
  const BLUEPRINT_TRIGGER_LOGIC = {
    'bathroom-light-fan-control': {
      ha_start: {
        conditions: ['Fan is currently on', 'Humidity is below threshold'],
        actions: ['Turn off fan if humidity is low', 'Turn off lights if room is vacant'],
        description: 'Checks if fan/lights should be off after Home Assistant restart',
      },
      humidity_bath_change: {
        conditions: ['Humidity sensors are valid', 'Humidity delta exceeds threshold'],
        actions: ['Turn on fan if humidity is high', 'Turn off fan if humidity normalized'],
        description: 'Responds to bathroom humidity changes to control exhaust fan',
      },
      humidity_home_change: {
        conditions: ['Humidity sensors are valid', 'Recalculate humidity delta'],
        actions: ['Adjust fan based on new baseline humidity'],
        description: 'Updates baseline humidity for delta calculation',
      },
      light_manual_off: {
        conditions: ['Light was turned off manually'],
        actions: ['Set manual override timer', 'Prevent auto-on for override duration'],
        description: 'Respects manual light control by setting override period',
      },
      fan_max_runtime_expired: {
        conditions: ['Fan has run for maximum allowed time'],
        actions: ['Turn off fan to prevent running indefinitely'],
        description: 'Safety shutoff after fan runs too long',
      },
      lights_off_for_fan_auto_off: {
        conditions: ['Lights have been off for configured time', 'Humidity is below threshold'],
        actions: ['Turn off fan since bathroom is likely unoccupied'],
        description: 'Auto-shutoff fan after lights off indicates room is empty',
      },
      wasp_motion: {
        conditions: ['Motion detected', 'Light is off', 'Someone is home', 'No manual override', 'Room is dark enough'],
        actions: ['Turn on light (dimmed if night mode)', 'Apply night mode settings if active'],
        description: 'Motion-activated lighting with night mode support',
      },
      wasp_door_opened: {
        conditions: ['Door opened', 'Light is off', 'Someone is home', 'No manual override', 'Room is dark enough'],
        actions: ['Turn on light (dimmed if night mode)'],
        description: 'Turn on light when door opens (entering bathroom)',
      },
      wasp_motion_clear: {
        conditions: ['No motion for configured delay', 'Door is open (room vacant)'],
        actions: ['Turn off light after vacancy detected'],
        description: 'Turn off light when room becomes vacant',
      },
      wasp_door_left_open: {
        conditions: ['Door has been open for configured time', 'No motion detected'],
        actions: ['Turn off light (assume person left with door open)'],
        description: 'Handle door left open indicating room is empty',
      },
    },
    'adaptive-comfort-control': {
      indoor_temp_state: {
        conditions: ['Indoor temperature changed', 'Within operating hours'],
        actions: ['Recalculate comfort setpoints', 'Adjust HVAC if needed'],
        description: 'Responds to indoor temperature changes',
      },
      outdoor_temp_state: {
        conditions: ['Outdoor temperature changed'],
        actions: ['Update adaptive setpoints based on outdoor conditions'],
        description: 'Adapts comfort targets based on outdoor temperature',
      },
      occupancy_state: {
        conditions: ['Occupancy state changed'],
        actions: ['Enable comfort mode when occupied', 'Enable eco mode when away'],
        description: 'Adjusts HVAC based on home occupancy',
      },
      season_state: {
        conditions: ['Season or mode changed'],
        actions: ['Switch between heating/cooling modes', 'Update seasonal setpoints'],
        description: 'Adapts to seasonal changes',
      },
      pause_any: {
        conditions: ['Manual pause requested'],
        actions: ['Pause automation temporarily', 'Resume after timeout'],
        description: 'Temporarily pauses automatic control',
      },
      climate_manual_change: {
        conditions: ['User manually adjusted thermostat'],
        actions: ['Respect manual change', 'Set override timer'],
        description: 'Respects manual thermostat adjustments',
      },
      rh_in_state: {
        conditions: ['Indoor humidity changed'],
        actions: ['Consider humidity in comfort calculation'],
        description: 'Factors humidity into comfort decisions',
      },
      tick_10m: {
        conditions: ['10-minute interval elapsed'],
        actions: ['Evaluate current conditions', 'Make adjustments if needed'],
        description: 'Periodic comfort evaluation',
      },
      tick_hour: {
        conditions: ['Hourly interval elapsed'],
        actions: ['Log status', 'Check for schedule changes'],
        description: 'Hourly status check and logging',
      },
      ha_start: {
        conditions: ['Home Assistant started'],
        actions: ['Initialize state', 'Apply current settings'],
        description: 'Initializes after Home Assistant restart',
      },
    },
    'adaptive-fan-control': {
      temp_changed: {
        conditions: ['Temperature changed', 'Room is occupied'],
        actions: ['Calculate optimal fan speed', 'Adjust fan accordingly'],
        description: 'Adjusts fan speed based on temperature',
      },
      presence_cleared: {
        conditions: ['Room became unoccupied'],
        actions: ['Turn off fan after delay', 'Save energy when away'],
        description: 'Turns off fan when room is empty',
      },
      presence_detected: {
        conditions: ['Motion or presence detected', 'Temperature warrants fan'],
        actions: ['Turn on fan at appropriate speed'],
        description: 'Activates fan when room becomes occupied',
      },
      hvac_changed: {
        conditions: ['HVAC state changed'],
        actions: ['Coordinate with HVAC', 'Avoid conflicting operation'],
        description: 'Coordinates fan with HVAC system',
      },
      season_changed: {
        conditions: ['Season mode changed'],
        actions: ['Adjust fan direction', 'Update speed thresholds'],
        description: 'Adapts to seasonal mode changes',
      },
      periodic: {
        conditions: ['Periodic check interval'],
        actions: ['Evaluate current conditions', 'Adjust if needed'],
        description: 'Regular condition check',
      },
    },
    'multi-switch-light-control': {
      // Standard switch presses
      up_single: {
        conditions: ['Up button pressed'],
        actions: ['Turn on lights', 'Run custom action if configured'],
        description: 'Turn lights on with single press',
      },
      down_single: {
        conditions: ['Down button pressed'],
        actions: ['Turn off lights', 'Run custom action if configured'],
        description: 'Turn lights off with single press',
      },
      up_double: {
        conditions: ['Up button double-pressed'],
        actions: ['Set lights to full brightness', 'Run custom action'],
        description: 'Full brightness with double press',
      },
      down_double: {
        conditions: ['Down button double-pressed'],
        actions: ['Run custom action if configured'],
        description: 'Custom action with double press',
      },
      up_triple: {
        conditions: ['Up button triple-pressed'],
        actions: ['Run custom action or scene'],
        description: 'Scene or custom action with triple press',
      },
      down_triple: {
        conditions: ['Down button triple-pressed'],
        actions: ['Run custom action or scene'],
        description: 'Scene or custom action with triple press',
      },
      up_quad: {
        conditions: ['Up button pressed 4 times'],
        actions: ['Run custom action or scene'],
        description: 'Extended action with 4 presses',
      },
      down_quad: {
        conditions: ['Down button pressed 4 times'],
        actions: ['Run custom action or scene'],
        description: 'Extended action with 4 presses',
      },
      up_quint: {
        conditions: ['Up button pressed 5 times'],
        actions: ['Run custom action or scene'],
        description: 'Extended action with 5 presses',
      },
      down_quint: {
        conditions: ['Down button pressed 5 times'],
        actions: ['Run custom action or scene'],
        description: 'Extended action with 5 presses',
      },
      up_hold_start: {
        conditions: ['Up button being held'],
        actions: ['Increase brightness gradually'],
        description: 'Dim up while holding button',
      },
      down_hold_start: {
        conditions: ['Down button being held'],
        actions: ['Decrease brightness gradually'],
        description: 'Dim down while holding button',
      },
      up_release: {
        conditions: ['Up button released'],
        actions: ['Stop brightness adjustment'],
        description: 'Stop dimming when released',
      },
      down_release: {
        conditions: ['Down button released'],
        actions: ['Stop brightness adjustment'],
        description: 'Stop dimming when released',
      },
      // Zigbee2MQTT variants (same behavior, different event format)
      up_single_z2m: {
        conditions: ['Up button pressed (Zigbee2MQTT)'],
        actions: ['Turn on lights', 'Run custom action if configured'],
        description: 'Turn lights on with single press',
      },
      down_single_z2m: {
        conditions: ['Down button pressed (Zigbee2MQTT)'],
        actions: ['Turn off lights', 'Run custom action if configured'],
        description: 'Turn lights off with single press',
      },
      up_double_z2m: {
        conditions: ['Up button double-pressed (Zigbee2MQTT)'],
        actions: ['Set lights to full brightness', 'Run custom action'],
        description: 'Full brightness with double press',
      },
      down_double_z2m: {
        conditions: ['Down button double-pressed (Zigbee2MQTT)'],
        actions: ['Run custom action if configured'],
        description: 'Custom action with double press',
      },
      up_triple_z2m: {
        conditions: ['Up button triple-pressed (Zigbee2MQTT)'],
        actions: ['Run custom action or scene'],
        description: 'Scene or custom action with triple press',
      },
      down_triple_z2m: {
        conditions: ['Down button triple-pressed (Zigbee2MQTT)'],
        actions: ['Run custom action or scene'],
        description: 'Scene or custom action with triple press',
      },
      up_quad_z2m: {
        conditions: ['Up button pressed 4 times (Zigbee2MQTT)'],
        actions: ['Run custom action or scene'],
        description: 'Extended action with 4 presses',
      },
      down_quad_z2m: {
        conditions: ['Down button pressed 4 times (Zigbee2MQTT)'],
        actions: ['Run custom action or scene'],
        description: 'Extended action with 4 presses',
      },
      up_quint_z2m: {
        conditions: ['Up button pressed 5 times (Zigbee2MQTT)'],
        actions: ['Run custom action or scene'],
        description: 'Extended action with 5 presses',
      },
      down_quint_z2m: {
        conditions: ['Down button pressed 5 times (Zigbee2MQTT)'],
        actions: ['Run custom action or scene'],
        description: 'Extended action with 5 presses',
      },
      up_held_z2m: {
        conditions: ['Up button being held (Zigbee2MQTT)'],
        actions: ['Increase brightness gradually'],
        description: 'Dim up while holding button',
      },
      down_held_z2m: {
        conditions: ['Down button being held (Zigbee2MQTT)'],
        actions: ['Decrease brightness gradually'],
        description: 'Dim down while holding button',
      },
      up_release_z2m: {
        conditions: ['Up button released (Zigbee2MQTT)'],
        actions: ['Stop brightness adjustment'],
        description: 'Stop dimming when released',
      },
      down_release_z2m: {
        conditions: ['Down button released (Zigbee2MQTT)'],
        actions: ['Stop brightness adjustment'],
        description: 'Stop dimming when released',
      },
      // Zigbee generic
      zigbee_action: {
        conditions: ['Zigbee switch action received'],
        actions: ['Map action to light control', 'Execute appropriate command'],
        description: 'Handles Zigbee switch events',
      },
      // Lutron Pico remotes
      lutron_on: {
        conditions: ['Pico on button pressed'],
        actions: ['Turn on lights with transition'],
        description: 'Turn on lights with Pico remote',
      },
      lutron_off: {
        conditions: ['Pico off button pressed'],
        actions: ['Turn off lights with transition'],
        description: 'Turn off lights with Pico remote',
      },
      lutron_raise: {
        conditions: ['Pico raise button held'],
        actions: ['Increase brightness gradually'],
        description: 'Brighten lights with Pico remote',
      },
      lutron_lower: {
        conditions: ['Pico lower button held'],
        actions: ['Decrease brightness gradually'],
        description: 'Dim lights with Pico remote',
      },
      lutron_raise_release: {
        conditions: ['Pico raise button released'],
        actions: ['Stop brightness increase'],
        description: 'Stop brightening when released',
      },
      lutron_lower_release: {
        conditions: ['Pico lower button released'],
        actions: ['Stop brightness decrease'],
        description: 'Stop dimming when released',
      },
      lutron_stop: {
        conditions: ['Pico dimming stopped'],
        actions: ['Stop any brightness adjustment'],
        description: 'Immediately stop dimming',
      },
      lutron_favorite: {
        conditions: ['Pico favorite button pressed'],
        actions: ['Set favorite brightness/color', 'Run custom action'],
        description: 'Set favorite scene with Pico center button',
      },
      // State sync
      state_sync: {
        conditions: ['Light state changed externally'],
        actions: ['Sync switch indicator with light state'],
        description: 'Keep switch LED in sync with light',
      },
      state_sync_light_on: {
        conditions: ['Light turned on externally'],
        actions: ['Update switch indicator to on'],
        description: 'Sync when light turns on elsewhere',
      },
      state_sync_light_off: {
        conditions: ['Light turned off externally'],
        actions: ['Update switch indicator to off'],
        description: 'Sync when light turns off elsewhere',
      },
    },
    'adaptive-shades': {
      time_pattern: {
        conditions: ['Every 5 minutes'],
        actions: ['Calculate optimal shade position', 'Adjust if sun angle changed'],
        description: 'Periodic sun position check',
      },
      sun_state: {
        conditions: ['Sun position changed significantly'],
        actions: ['Recalculate shade positions', 'Apply new positions'],
        description: 'Responds to sun movement',
      },
      cover_changed: {
        conditions: ['Cover position changed manually'],
        actions: ['Detect manual override', 'Pause automation temporarily'],
        description: 'Respects manual shade adjustments',
      },
      window_opened: {
        conditions: ['Window contact opened'],
        actions: ['Open shades for ventilation', 'Wait configured delay'],
        description: 'Opens shades when window opens',
      },
      window_closed: {
        conditions: ['Window contact closed'],
        actions: ['Resume normal shade control'],
        description: 'Resumes automation when window closes',
      },
      ha_start: {
        conditions: ['Home Assistant started'],
        actions: ['Initialize shade positions', 'Apply current sun position'],
        description: 'Initializes after restart',
      },
    },
  };

  // Human-friendly trigger name mappings based on common trigger IDs
  const TRIGGER_FRIENDLY_NAMES = {
    // Bathroom Light Fan Control
    ha_start: {
      name: 'Home Assistant Starts',
      icon: '🏠',
      description: 'When Home Assistant restarts, check current state',
    },
    humidity_bath_change: {
      name: 'Bathroom Humidity Changes',
      icon: '💧',
      description: 'When bathroom humidity sensor updates',
    },
    humidity_home_change: { name: 'Home Humidity Changes', icon: '💧', description: 'When baseline humidity changes' },
    light_manual_off: {
      name: 'Light Turned Off Manually',
      icon: '💡',
      description: 'When someone manually turns off the light',
    },
    fan_max_runtime_expired: {
      name: 'Fan Max Runtime Reached',
      icon: '⏱️',
      description: 'Fan has run for maximum allowed time',
    },
    lights_off_for_fan_auto_off: {
      name: 'Lights Off Delay Passed',
      icon: '💨',
      description: 'Lights have been off long enough to consider turning off fan',
    },
    wasp_motion: { name: 'Motion Detected', icon: '🚶', description: 'Someone moved in the room' },
    wasp_door_opened: { name: 'Door Opened', icon: '🚪', description: 'The door was opened' },
    wasp_motion_clear: { name: 'Motion Cleared', icon: '🚶', description: 'No motion for the configured delay' },
    wasp_door_left_open: { name: 'Door Left Open', icon: '🚪', description: 'Door has been open for too long' },

    // Adaptive Comfort Control
    indoor_temp_state: { name: 'Indoor Temp Changed', icon: '🌡️', description: 'Indoor temperature sensor updated' },
    outdoor_temp_state: { name: 'Outdoor Temp Changed', icon: '🌤️', description: 'Outdoor temperature changed' },
    occupancy_state: { name: 'Occupancy Changed', icon: '👥', description: 'Home occupancy state changed' },
    season_state: { name: 'Season Changed', icon: '🍂', description: 'Season or mode changed' },
    pause_any: { name: 'Pause Requested', icon: '⏸️', description: 'Manual pause requested' },
    climate_manual_change: { name: 'Thermostat Adjusted', icon: '🎛️', description: 'User manually adjusted thermostat' },
    climate_manual_change_low: { name: 'Setpoint Lowered', icon: '🎛️', description: 'User lowered thermostat setpoint' },
    climate_manual_change_high: { name: 'Setpoint Raised', icon: '🎛️', description: 'User raised thermostat setpoint' },
    rh_in_state: { name: 'Indoor Humidity Changed', icon: '💧', description: 'Indoor humidity sensor updated' },
    rh_out_state: { name: 'Outdoor Humidity Changed', icon: '💧', description: 'Outdoor humidity changed' },
    rmot_state: { name: 'Radiant Mean Temp Changed', icon: '🌡️', description: 'Radiant temperature updated' },
    baro_state: { name: 'Barometric Pressure Changed', icon: '📊', description: 'Barometric pressure updated' },
    tick_10m: { name: 'Every 10 Minutes', icon: '⏰', description: 'Periodic 10-minute check' },
    tick_hour: { name: 'Every Hour', icon: '⏰', description: 'Hourly status check' },

    // Adaptive Fan Control
    temp_changed: { name: 'Temperature Changed', icon: '🌡️', description: 'Temperature sensor updated' },
    presence_cleared: { name: 'Room Vacated', icon: '🚶', description: 'Room became unoccupied' },
    presence_detected: { name: 'Presence Detected', icon: '🚶', description: 'Motion or presence detected' },
    hvac_changed: { name: 'HVAC State Changed', icon: '❄️', description: 'HVAC system state changed' },
    season_changed: { name: 'Season Mode Changed', icon: '🍂', description: 'Season operating mode changed' },
    periodic: { name: 'Periodic Check', icon: '⏰', description: 'Regular interval check' },

    // Multi-Switch Light Control - Standard switches
    up_single: { name: 'Up Button Pressed Once', icon: '🔼', description: 'Single press on the up/on button' },
    down_single: { name: 'Down Button Pressed Once', icon: '🔽', description: 'Single press on the down/off button' },
    up_double: { name: 'Up Button Double-Pressed', icon: '⏫', description: 'Quick double-press on up button' },
    down_double: { name: 'Down Button Double-Pressed', icon: '⏬', description: 'Quick double-press on down button' },
    up_triple: { name: 'Up Button Triple-Pressed', icon: '🔼', description: 'Quick triple-press on up button' },
    down_triple: { name: 'Down Button Triple-Pressed', icon: '🔽', description: 'Quick triple-press on down button' },
    up_quad: { name: 'Up Button 4x Pressed', icon: '🔼', description: 'Four quick presses on up button' },
    down_quad: { name: 'Down Button 4x Pressed', icon: '🔽', description: 'Four quick presses on down button' },
    up_quint: { name: 'Up Button 5x Pressed', icon: '🔼', description: 'Five quick presses on up button' },
    down_quint: { name: 'Down Button 5x Pressed', icon: '🔽', description: 'Five quick presses on down button' },
    up_hold_start: { name: 'Holding Up Button', icon: '⬆️', description: 'Started holding up button' },
    down_hold_start: { name: 'Holding Down Button', icon: '⬇️', description: 'Started holding down button' },
    up_release: { name: 'Up Button Released', icon: '🔼', description: 'Released the up button after holding' },
    down_release: { name: 'Down Button Released', icon: '🔽', description: 'Released the down button after holding' },

    // Multi-Switch - Zigbee2MQTT variants (z2m)
    up_single_z2m: { name: 'Up Button Pressed Once', icon: '🔼', description: 'Single press on up (Zigbee2MQTT)' },
    down_single_z2m: {
      name: 'Down Button Pressed Once',
      icon: '🔽',
      description: 'Single press on down (Zigbee2MQTT)',
    },
    up_double_z2m: { name: 'Up Button Double-Pressed', icon: '⏫', description: 'Double-press on up (Zigbee2MQTT)' },
    down_double_z2m: {
      name: 'Down Button Double-Pressed',
      icon: '⏬',
      description: 'Double-press on down (Zigbee2MQTT)',
    },
    up_triple_z2m: { name: 'Up Button Triple-Pressed', icon: '🔼', description: 'Triple-press on up (Zigbee2MQTT)' },
    down_triple_z2m: {
      name: 'Down Button Triple-Pressed',
      icon: '🔽',
      description: 'Triple-press on down (Zigbee2MQTT)',
    },
    up_quad_z2m: { name: 'Up Button 4x Pressed', icon: '🔼', description: 'Four presses on up (Zigbee2MQTT)' },
    down_quad_z2m: { name: 'Down Button 4x Pressed', icon: '🔽', description: 'Four presses on down (Zigbee2MQTT)' },
    up_quint_z2m: { name: 'Up Button 5x Pressed', icon: '🔼', description: 'Five presses on up (Zigbee2MQTT)' },
    down_quint_z2m: { name: 'Down Button 5x Pressed', icon: '🔽', description: 'Five presses on down (Zigbee2MQTT)' },
    up_held_z2m: { name: 'Holding Up Button', icon: '⬆️', description: 'Holding up button (Zigbee2MQTT)' },
    down_held_z2m: { name: 'Holding Down Button', icon: '⬇️', description: 'Holding down button (Zigbee2MQTT)' },
    up_release_z2m: { name: 'Up Button Released', icon: '🔼', description: 'Released up button (Zigbee2MQTT)' },
    down_release_z2m: { name: 'Down Button Released', icon: '🔽', description: 'Released down button (Zigbee2MQTT)' },

    // Multi-Switch - Lutron Pico remotes
    lutron_on: { name: 'Pico On Button Pressed', icon: '💡', description: 'Pressed the on button on Pico remote' },
    lutron_off: { name: 'Pico Off Button Pressed', icon: '🔌', description: 'Pressed the off button on Pico remote' },
    lutron_raise: { name: 'Pico Raise Button Held', icon: '⬆️', description: 'Holding the raise/brighten button' },
    lutron_lower: { name: 'Pico Lower Button Held', icon: '⬇️', description: 'Holding the lower/dim button' },
    lutron_raise_release: { name: 'Pico Raise Button Released', icon: '⬆️', description: 'Released the raise button' },
    lutron_lower_release: { name: 'Pico Lower Button Released', icon: '⬇️', description: 'Released the lower button' },
    lutron_stop: { name: 'Pico Stop', icon: '⏹️', description: 'Stop dimming action' },
    lutron_favorite: {
      name: 'Pico Favorite Button Pressed',
      icon: '⭐',
      description: 'Pressed the favorite/center button',
    },

    // Multi-Switch - State sync
    state_sync: {
      name: 'Light State Changed Externally',
      icon: '🔄',
      description: 'Light was controlled by another source',
    },
    state_sync_light_on: {
      name: 'Light Turned On Externally',
      icon: '💡',
      description: 'Light was turned on by another automation or app',
    },
    state_sync_light_off: {
      name: 'Light Turned Off Externally',
      icon: '🔌',
      description: 'Light was turned off by another automation or app',
    },

    // Zigbee generic
    zigbee_action: { name: 'Zigbee Switch Action', icon: '📡', description: 'Received action from Zigbee switch' },

    // Adaptive Shades
    time_pattern: { name: 'Every 5 Minutes', icon: '⏰', description: 'Periodic check for sun position updates' },
    sun_state: { name: 'Sun Position Changed', icon: '☀️', description: 'Sun moved significantly in the sky' },
    cover_changed: { name: 'Shade Moved Manually', icon: '🪟', description: 'Someone manually adjusted the shade' },
    window_opened: { name: 'Window Opened', icon: '🪟', description: 'Window contact sensor detected open' },
    window_closed: { name: 'Window Closed', icon: '🪟', description: 'Window contact sensor detected closed' },

    // Common generic triggers
    motion_detected: { name: 'Motion Detected', icon: '🚶', description: 'Motion sensor triggered' },
    no_motion: { name: 'No Motion', icon: '🚶', description: 'Motion cleared after delay' },
    temperature_change: { name: 'Temperature Changed', icon: '🌡️', description: 'Temperature sensor updated' },
    time_trigger: { name: 'Scheduled Time', icon: '⏰', description: 'Time-based schedule triggered' },
    sun_event: { name: 'Sunrise/Sunset', icon: '🌅', description: 'Sun position changed' },
    state_change: { name: 'State Changed', icon: '🔄', description: 'Entity state updated' },
  };

  // Human-friendly action descriptions
  const ACTION_FRIENDLY_NAMES = {
    'light.turn_on': { name: 'Turn On Light', icon: '💡' },
    'light.turn_off': { name: 'Turn Off Light', icon: '🔌' },
    'fan.turn_on': { name: 'Turn On Fan', icon: '💨' },
    'fan.turn_off': { name: 'Turn Off Fan', icon: '🔇' },
    'switch.turn_on': { name: 'Turn On Switch', icon: '🔛' },
    'switch.turn_off': { name: 'Turn Off Switch', icon: '🔴' },
    'climate.set_temperature': { name: 'Set Temperature', icon: '🌡️' },
    'climate.turn_on': { name: 'Turn On HVAC', icon: '❄️' },
    'climate.turn_off': { name: 'Turn Off HVAC', icon: '🔌' },
    'cover.open_cover': { name: 'Open Shades', icon: '☀️' },
    'cover.close_cover': { name: 'Close Shades', icon: '🌙' },
    'cover.set_cover_position': { name: 'Set Shade Position', icon: '↕️' },
    'logbook.log': { name: 'Log Event', icon: '📝' },
    'input_boolean.turn_on': { name: 'Enable Flag', icon: '✅' },
    'input_boolean.turn_off': { name: 'Disable Flag', icon: '⬜' },
    'input_datetime.set_datetime': { name: 'Set Timer', icon: '⏱️' },
  };

  // State
  let currentBlueprint = null;
  let parsedData = null;

  // DOM Elements
  const blueprintSelect = document.getElementById('blueprint-select');
  const diagramPlaceholder = document.getElementById('diagram-placeholder');
  const diagramLoading = document.getElementById('diagram-loading');
  const matrixContent = document.getElementById('matrix-content');
  const vizInfo = document.getElementById('viz-info');

  // Initialize Mermaid with dark theme
  if (typeof mermaid !== 'undefined') {
    mermaid.initialize({
      startOnLoad: false,
      theme: 'dark',
      themeVariables: {
        primaryColor: '#a78bfa',
        primaryTextColor: '#e2e8f0',
        primaryBorderColor: '#a78bfa',
        lineColor: '#64748b',
        secondaryColor: '#22d3ee',
        tertiaryColor: '#0a0a1a',
        background: '#030014',
        mainBkg: '#0a0a1a',
        nodeBorder: '#a78bfa',
        clusterBkg: 'rgba(167, 139, 250, 0.1)',
        clusterBorder: '#a78bfa',
        titleColor: '#e2e8f0',
        edgeLabelBackground: '#0a0a1a',
        nodeTextColor: '#e2e8f0',
      },
      flowchart: {
        htmlLabels: true,
        curve: 'basis',
        padding: 15,
        nodeSpacing: 50,
        rankSpacing: 50,
        useMaxWidth: true,
      },
      securityLevel: 'loose',
    });
  }

  // Event Listeners
  if (blueprintSelect) {
    blueprintSelect.addEventListener('change', handleBlueprintSelect);
  }

  /**
   * Handle blueprint selection
   */
  async function handleBlueprintSelect(e) {
    const blueprintId = e.target.value;
    if (!blueprintId) {
      showPlaceholder();
      return;
    }

    currentBlueprint = blueprintId;
    showLoading();

    try {
      const yaml = await fetchBlueprint(blueprintId);
      parsedData = parseBlueprint(yaml);
      updateInfoPanel(parsedData);
      renderVisualization();
    } catch (error) {
      console.error('Error loading blueprint:', error);
      showError(error.message);
    }
  }

  /**
   * Fetch blueprint YAML from GitHub
   */
  async function fetchBlueprint(blueprintId) {
    const url = BLUEPRINT_URLS[blueprintId];
    if (!url) throw new Error('Blueprint not found');

    const response = await fetch(url);
    if (!response.ok) throw new Error(`Failed to fetch blueprint: ${response.status}`);

    return await response.text();
  }

  /**
   * Create custom YAML schema to handle Home Assistant tags
   */
  function createHASchema() {
    // Define custom types for Home Assistant YAML tags
    const inputType = new jsyaml.Type('!input', {
      kind: 'scalar',
      construct: function (data) {
        return { __ha_type: 'input', value: data };
      },
    });

    const includeType = new jsyaml.Type('!include', {
      kind: 'scalar',
      construct: function (data) {
        return { __ha_type: 'include', value: data };
      },
    });

    const secretType = new jsyaml.Type('!secret', {
      kind: 'scalar',
      construct: function (data) {
        return { __ha_type: 'secret', value: data };
      },
    });

    const envVarType = new jsyaml.Type('!env_var', {
      kind: 'scalar',
      construct: function (data) {
        return { __ha_type: 'env_var', value: data };
      },
    });

    const includeListType = new jsyaml.Type('!include_dir_list', {
      kind: 'scalar',
      construct: function (data) {
        return { __ha_type: 'include_dir_list', value: data };
      },
    });

    const includeMergeType = new jsyaml.Type('!include_dir_merge_list', {
      kind: 'scalar',
      construct: function (data) {
        return { __ha_type: 'include_dir_merge_list', value: data };
      },
    });

    const includeNamedType = new jsyaml.Type('!include_dir_named', {
      kind: 'scalar',
      construct: function (data) {
        return { __ha_type: 'include_dir_named', value: data };
      },
    });

    const includeMergeNamedType = new jsyaml.Type('!include_dir_merge_named', {
      kind: 'scalar',
      construct: function (data) {
        return { __ha_type: 'include_dir_merge_named', value: data };
      },
    });

    // Create schema extending DEFAULT_SCHEMA
    return jsyaml.DEFAULT_SCHEMA.extend([
      inputType,
      includeType,
      secretType,
      envVarType,
      includeListType,
      includeMergeType,
      includeNamedType,
      includeMergeNamedType,
    ]);
  }

  /**
   * Parse blueprint YAML into structured data
   */
  function parseBlueprint(yamlText) {
    if (typeof jsyaml === 'undefined') {
      throw new Error('js-yaml library not loaded');
    }

    // Use custom schema to handle HA tags
    const haSchema = createHASchema();
    const doc = jsyaml.load(yamlText, { schema: haSchema });

    const data = {
      name: doc.blueprint?.name || 'Unknown Blueprint',
      description: doc.blueprint?.description || '',
      domain: doc.blueprint?.domain || 'automation',
      version: extractVersion(doc.blueprint?.name || ''),
      inputs: parseInputs(doc.blueprint?.input || {}),
      triggers: parseTriggers(doc.trigger || []),
      conditions: parseConditions(doc.condition || []),
      actions: parseActions(doc.action || []),
      variables: doc.variables || {},
      mode: doc.mode || 'single',
    };

    // Calculate statistics
    data.stats = {
      triggers: data.triggers.length,
      conditions: countConditions(data.conditions) + countActionConditions(data.actions),
      actions: countActions(data.actions),
      branches: countBranches(data.actions),
    };

    return data;
  }

  /**
   * Extract version from blueprint name
   */
  function extractVersion(name) {
    const match = name.match(/v?(\d+\.\d+\.?\d*)/i);
    return match ? match[1] : '1.0';
  }

  /**
   * Parse blueprint inputs
   */
  function parseInputs(input) {
    const inputs = [];

    function extractInputs(obj, section = null) {
      for (const [key, value] of Object.entries(obj)) {
        if (value && typeof value === 'object') {
          if (value.input) {
            // This is a section with nested inputs
            extractInputs(value.input, value.name || key);
          } else if (value.name || value.selector) {
            // This is an input definition
            inputs.push({
              id: key,
              name: value.name || key,
              description: value.description || '',
              default: value.default,
              section: section,
              selector: value.selector,
            });
          }
        }
      }
    }

    extractInputs(input);
    return inputs;
  }

  /**
   * Parse triggers
   */
  function parseTriggers(triggers) {
    if (!Array.isArray(triggers)) {
      triggers = [triggers];
    }

    return triggers.map((trigger, index) => ({
      id: trigger.id || `trigger_${index}`,
      platform: trigger.platform || 'unknown',
      entity_id: trigger.entity_id || null,
      event: trigger.event || null,
      to: trigger.to || null,
      from: trigger.from || null,
      for: trigger.for || null,
      description: generateTriggerDescription(trigger),
    }));
  }

  /**
   * Generate human-readable trigger description
   * Priority: 1. Known trigger ID mapping, 2. Smart description from platform/entity, 3. Generic fallback
   */
  function generateTriggerDescription(trigger) {
    const triggerId = trigger.id || '';
    const platform = trigger.platform || 'unknown';

    // First check if we have a human-friendly name for this trigger ID
    if (triggerId && TRIGGER_FRIENDLY_NAMES[triggerId]) {
      return TRIGGER_FRIENDLY_NAMES[triggerId].name;
    }

    // Generate smart description based on platform and context
    switch (platform) {
      case 'state':
        return generateStateDescription(trigger);
      case 'homeassistant':
        if (trigger.event === 'start') {
          return 'Home Assistant Starts';
        }
        return `System Event: ${trigger.event || 'event'}`;
      case 'time':
        return `At ${formatTimeValue(trigger.at)}`;
      case 'sun':
        if (trigger.event === 'sunrise') return 'At Sunrise';
        if (trigger.event === 'sunset') return 'At Sunset';
        return `Sun: ${trigger.event || 'event'}`;
      case 'numeric_state':
        return generateNumericStateDescription(trigger);
      case 'time_pattern':
        return generateTimePatternDescription(trigger);
      default:
        // Try to create a readable name from the trigger ID
        if (triggerId) {
          return formatTriggerId(triggerId);
        }
        return `${capitalizeFirst(platform)} Trigger`;
    }
  }

  /**
   * Generate description for state triggers
   */
  function generateStateDescription(trigger) {
    const entityName = formatEntityId(trigger.entity_id);
    const entityDomain = getEntityDomain(trigger.entity_id);

    // Special handling based on domain
    if (entityDomain === 'binary_sensor') {
      if (trigger.from === 'off' && trigger.to === 'on') {
        // Try to determine sensor type from entity name
        if (entityName.toLowerCase().includes('motion')) {
          return 'Motion Detected';
        }
        if (entityName.toLowerCase().includes('door')) {
          return 'Door Opened';
        }
        if (entityName.toLowerCase().includes('window')) {
          return 'Window Opened';
        }
        return `${entityName} Activated`;
      }
      if (trigger.from === 'on' && trigger.to === 'off') {
        if (entityName.toLowerCase().includes('motion')) {
          return trigger.for ? `No Motion for ${formatDuration(trigger.for)}` : 'Motion Cleared';
        }
        if (entityName.toLowerCase().includes('door')) {
          return 'Door Closed';
        }
        return `${entityName} Deactivated`;
      }
    }

    if (entityDomain === 'light') {
      if (trigger.to === 'off') {
        return trigger.from === 'on' ? 'Light Turned Off' : 'Light Off';
      }
      if (trigger.to === 'on') {
        return 'Light Turned On';
      }
    }

    if (entityDomain === 'fan' || entityDomain === 'switch') {
      const deviceName = entityDomain === 'fan' ? 'Fan' : entityName;
      if (trigger.to === 'on') {
        if (trigger.for) {
          return `${deviceName} On for ${formatDuration(trigger.for)}`;
        }
        return `${deviceName} Turned On`;
      }
      if (trigger.to === 'off') {
        return `${deviceName} Turned Off`;
      }
    }

    if (entityDomain === 'sensor') {
      return `${entityName} Changed`;
    }

    // Generic state change description
    let desc = `${entityName}`;
    if (trigger.from && trigger.to) {
      desc += ` changes ${trigger.from} → ${trigger.to}`;
    } else if (trigger.to) {
      desc += ` becomes ${trigger.to}`;
    }
    if (trigger.for) {
      desc += ` for ${formatDuration(trigger.for)}`;
    }
    return desc;
  }

  /**
   * Generate description for numeric_state triggers
   */
  function generateNumericStateDescription(trigger) {
    const entityName = formatEntityId(trigger.entity_id);

    if (trigger.above !== undefined && trigger.below !== undefined) {
      return `${entityName} between ${trigger.above} and ${trigger.below}`;
    }
    if (trigger.above !== undefined) {
      return `${entityName} rises above ${trigger.above}`;
    }
    if (trigger.below !== undefined) {
      return `${entityName} drops below ${trigger.below}`;
    }
    return `${entityName} Value Changed`;
  }

  /**
   * Generate description for time_pattern triggers
   */
  function generateTimePatternDescription(trigger) {
    if (trigger.minutes && trigger.minutes === '/5') {
      return 'Every 5 Minutes';
    }
    if (trigger.minutes && trigger.minutes === '/1') {
      return 'Every Minute';
    }
    if (trigger.hours) {
      return `Every ${trigger.hours} Hour(s)`;
    }
    return 'On Schedule';
  }

  /**
   * Format a trigger ID into readable text
   */
  function formatTriggerId(triggerId) {
    return triggerId
      .replace(/_/g, ' ')
      .replace(/\b\w/g, (c) => c.toUpperCase())
      .replace(/^Wasp /, '') // Remove "Wasp" prefix for readability
      .trim();
  }

  /**
   * Format time value for display
   */
  function formatTimeValue(timeValue) {
    if (!timeValue) return 'scheduled time';
    if (typeof timeValue === 'object' && timeValue.__ha_type) {
      return 'configured time';
    }
    // Try to format HH:MM:SS to HH:MM AM/PM
    if (typeof timeValue === 'string' && timeValue.includes(':')) {
      const parts = timeValue.split(':');
      if (parts.length >= 2) {
        let hours = parseInt(parts[0], 10);
        const mins = parts[1];
        const ampm = hours >= 12 ? 'PM' : 'AM';
        hours = hours % 12 || 12;
        return `${hours}:${mins} ${ampm}`;
      }
    }
    return timeValue;
  }

  /**
   * Get entity domain from entity_id
   */
  function getEntityDomain(entityId) {
    if (!entityId) return '';
    if (typeof entityId === 'object' && entityId.__ha_type) {
      return 'input'; // Template input
    }
    if (typeof entityId === 'string') {
      return entityId.split('.')[0] || '';
    }
    return '';
  }

  /**
   * Capitalize first letter
   */
  function capitalizeFirst(str) {
    if (!str) return '';
    return str.charAt(0).toUpperCase() + str.slice(1);
  }

  /**
   * Parse conditions
   */
  function parseConditions(conditions) {
    if (!Array.isArray(conditions)) {
      return conditions ? [conditions] : [];
    }
    return conditions.map((cond) => parseCondition(cond));
  }

  /**
   * Parse single condition
   */
  function parseCondition(condition) {
    if (!condition) return null;

    const type = condition.condition || 'template';

    return {
      type: type,
      entity_id: condition.entity_id,
      state: condition.state,
      value_template: condition.value_template,
      conditions: condition.conditions ? condition.conditions.map((c) => parseCondition(c)) : null,
      description: generateConditionDescription(condition),
    };
  }

  /**
   * Generate human-readable condition description
   */
  function generateConditionDescription(condition) {
    if (!condition) return 'Unknown';

    const type = condition.condition || 'template';

    switch (type) {
      case 'state':
        return generateStateConditionDescription(condition);
      case 'template':
        return generateTemplateConditionDescription(condition.value_template);
      case 'and':
        return 'All conditions must be true';
      case 'or':
        return 'Any condition can be true';
      case 'not':
        return 'Condition must NOT be true';
      case 'numeric_state':
        return generateNumericConditionDescription(condition);
      case 'time':
        return generateTimeConditionDescription(condition);
      case 'sun':
        return generateSunConditionDescription(condition);
      case 'zone':
        return 'In specified zone';
      default:
        return capitalizeFirst(type) + ' condition';
    }
  }

  /**
   * Generate description for state conditions
   */
  function generateStateConditionDescription(condition) {
    const entityName = formatEntityId(condition.entity_id);
    const entityDomain = getEntityDomain(condition.entity_id);
    const state = condition.state;

    // Create human-friendly descriptions
    if (entityDomain === 'light') {
      return state === 'on' ? 'Light is on' : 'Light is off';
    }
    if (entityDomain === 'fan') {
      return state === 'on' ? 'Fan is running' : 'Fan is off';
    }
    if (entityDomain === 'binary_sensor') {
      if (entityName.toLowerCase().includes('motion')) {
        return state === 'on' ? 'Motion detected' : 'No motion';
      }
      if (entityName.toLowerCase().includes('door')) {
        return state === 'on' ? 'Door is open' : 'Door is closed';
      }
      if (entityName.toLowerCase().includes('occupancy') || entityName.toLowerCase().includes('presence')) {
        return state === 'on' ? 'Room occupied' : 'Room vacant';
      }
    }
    if (entityDomain === 'input_boolean') {
      return state === 'on' ? `${entityName} enabled` : `${entityName} disabled`;
    }
    if (entityDomain === 'person') {
      return state === 'home' ? 'Someone is home' : 'Away from home';
    }

    return `${entityName} is ${state}`;
  }

  /**
   * Generate description for template conditions
   * Comprehensively parses common Home Assistant template patterns
   */
  function generateTemplateConditionDescription(template) {
    if (!template) return 'Condition met';
    if (typeof template !== 'string') return 'Condition evaluated';

    // Clean up the template for analysis
    const clean = template.replace(/\{\{/g, '').replace(/\}\}/g, '').replace(/\s+/g, ' ').trim();

    // Look for common patterns and translate to human-readable

    // Trigger ID checks - exact match
    const triggerIdMatch = clean.match(/trigger\.id\s*[=!]=?\s*['"]([^'"]+)['"]/);
    if (triggerIdMatch) {
      const triggerId = triggerIdMatch[1];
      const isEquals = !clean.includes('!=');
      const friendlyName = TRIGGER_FRIENDLY_NAMES[triggerId]?.name || formatTriggerId(triggerId);
      return isEquals ? `Triggered by: ${friendlyName}` : `Not triggered by: ${friendlyName}`;
    }

    // Trigger ID in list
    const triggerInMatch = clean.match(/trigger\.id\s+in\s+\[([^\]]+)\]/);
    if (triggerInMatch) {
      const ids = triggerInMatch[1].match(/['"]([^'"]+)['"]/g);
      if (ids && ids.length <= 3) {
        const names = ids.map((id) => {
          const cleanId = id.replace(/['"]/g, '');
          return TRIGGER_FRIENDLY_NAMES[cleanId]?.name || formatTriggerId(cleanId);
        });
        return `Triggered by: ${names.join(' or ')}`;
      }
      return `Triggered by one of ${ids?.length || 'several'} events`;
    }

    // State checks - is_state function
    const isStateMatch = clean.match(/is_state\s*\(\s*['"]?([^'"]+)['"]?\s*,\s*['"]([^'"]+)['"]\s*\)/);
    if (isStateMatch) {
      const entity = isStateMatch[1];
      const state = isStateMatch[2];
      const entityName = formatEntityId(entity);
      if (state === 'on') return `${entityName} is on`;
      if (state === 'off') return `${entityName} is off`;
      if (state === 'home') return `${entityName} is home`;
      return `${entityName} is "${state}"`;
    }

    // States function comparisons
    const statesMatch = clean.match(/states\s*\(\s*['"]?([^'"]+)['"]?\s*\)/);
    if (statesMatch) {
      const entity = statesMatch[1];
      const entityName = formatEntityId(entity);
      if (clean.includes('not in') && clean.includes('unknown')) {
        return `${entityName} is available`;
      }
      if (clean.includes('>') || clean.includes('>=')) {
        return `${entityName} is above threshold`;
      }
      if (clean.includes('<') || clean.includes('<=')) {
        return `${entityName} is below threshold`;
      }
      return `${entityName} state checked`;
    }

    // Common blueprint variable checks
    if (clean.includes('presence_ok')) {
      if (clean.includes('not') || clean.includes('false')) {
        return 'No one is home';
      }
      return 'Someone is home';
    }
    if (clean.includes('override_ok')) {
      if (clean.includes('not') || clean.includes('false')) {
        return 'Manual override is active';
      }
      return 'No manual override';
    }
    if (clean.includes('lux_ok')) {
      return 'Room is dark enough';
    }
    if (clean.includes('humidity_sensors_ok')) {
      return 'Humidity sensors are valid';
    }
    if (clean.includes('night_mode_active')) {
      return 'Night mode is active';
    }
    if (clean.includes('in_night_schedule')) {
      return 'Within night hours';
    }
    if (clean.includes('should_turn_on_light')) {
      return 'Conditions met to turn on light';
    }
    if (clean.includes('area_set')) {
      return 'Using area-based control';
    }
    if (clean.includes('fan_is_fan')) {
      return 'Target is a fan entity';
    }
    if (clean.includes('automation_control')) {
      return 'Automation control helper set';
    }

    // Humidity delta checks
    if (clean.includes('humidity_delta')) {
      if (clean.includes('humidity_delta_on') || (clean.includes('>') && !clean.includes('<'))) {
        return 'Humidity exceeds ON threshold';
      }
      if (clean.includes('humidity_delta_off') || (clean.includes('<') && !clean.includes('>'))) {
        return 'Humidity below OFF threshold';
      }
      return 'Humidity level checked';
    }

    // Debug level checks - don't show as condition
    if (clean.includes('debug_level')) {
      return 'Debug logging enabled';
    }

    // Time-based checks
    if (clean.includes('now()') || clean.includes('as_timestamp')) {
      if (clean.includes('after') || clean.includes('>')) {
        return 'Current time is after threshold';
      }
      if (clean.includes('before') || clean.includes('<')) {
        return 'Current time is before threshold';
      }
      return 'Time-based check';
    }

    // Rate of rise/fall checks
    if (clean.includes('rate_of_rise')) {
      return 'Humidity rising quickly';
    }
    if (clean.includes('rate_of_fall')) {
      return 'Humidity falling quickly';
    }

    // Light capability checks
    if (clean.includes('_has_bri') || clean.includes('brightness')) {
      return 'Light supports brightness';
    }
    if (clean.includes('_has_ct') || clean.includes('color_temp')) {
      return 'Light supports color temperature';
    }
    if (clean.includes('supported_color_modes')) {
      return 'Checking light capabilities';
    }

    // Boolean/logical checks
    if (clean.match(/^\s*(true|false)\s*$/i)) {
      return clean.toLowerCase().includes('true') ? 'Always true' : 'Always false';
    }

    // Entity domain checks
    if (clean.includes("split('.')[0]") || clean.includes('domain')) {
      return 'Checking entity type';
    }

    // Length/count checks
    if (clean.includes('| length') || clean.includes('| count')) {
      return 'Checking list size';
    }

    // Not in unavailable/unknown
    if (clean.includes("not in ['unknown','unavailable'") || clean.includes("not in ['unavailable','unknown'")) {
      return 'Entity is available';
    }

    // Float comparisons
    if (clean.includes('| float') && (clean.includes('>') || clean.includes('<'))) {
      return 'Numeric comparison';
    }

    // Generic: try to extract meaningful keywords
    const keywords = clean.match(
      /\b(motion|door|light|fan|humidity|temperature|presence|occupied|home|night|day|sun|time)\b/gi
    );
    if (keywords && keywords.length > 0) {
      const uniqueKeywords = [...new Set(keywords.map((k) => k.toLowerCase()))];
      return `Checking ${uniqueKeywords.slice(0, 2).join(' and ')} state`;
    }

    // If all else fails, try to make a short readable version
    if (clean.length > 50) {
      // Complex template - give a generic but accurate description
      return 'Complex condition evaluated';
    }

    // Return a cleaned up short version
    const shortClean = clean.length > 40 ? clean.substring(0, 37) + '...' : clean;
    // If it's just a variable name, format it nicely
    if (shortClean.match(/^[a-z_]+$/i)) {
      return formatTriggerId(shortClean) + ' is true';
    }
    return shortClean;
  }

  /**
   * Generate description for numeric_state conditions
   */
  function generateNumericConditionDescription(condition) {
    const entityName = formatEntityId(condition.entity_id);
    const entityDomain = getEntityDomain(condition.entity_id);

    // Smart descriptions based on entity type
    if (entityDomain === 'sensor') {
      if (entityName.toLowerCase().includes('humidity')) {
        if (condition.above !== undefined) {
          return `Humidity above ${condition.above}%`;
        }
        if (condition.below !== undefined) {
          return `Humidity below ${condition.below}%`;
        }
      }
      if (entityName.toLowerCase().includes('temp')) {
        if (condition.above !== undefined) {
          return `Temperature above ${condition.above}°`;
        }
        if (condition.below !== undefined) {
          return `Temperature below ${condition.below}°`;
        }
      }
      if (entityName.toLowerCase().includes('lux') || entityName.toLowerCase().includes('illumin')) {
        if (condition.above !== undefined) {
          return `Room is bright (>${condition.above} lux)`;
        }
        if (condition.below !== undefined) {
          return `Room is dark (<${condition.below} lux)`;
        }
      }
    }

    // Generic numeric description
    let desc = entityName;
    if (condition.above !== undefined && condition.below !== undefined) {
      desc += ` between ${condition.above} and ${condition.below}`;
    } else if (condition.above !== undefined) {
      desc += ` > ${condition.above}`;
    } else if (condition.below !== undefined) {
      desc += ` < ${condition.below}`;
    }
    return desc;
  }

  /**
   * Generate description for time conditions
   */
  function generateTimeConditionDescription(condition) {
    if (condition.after && condition.before) {
      return `Between ${formatTimeValue(condition.after)} and ${formatTimeValue(condition.before)}`;
    }
    if (condition.after) {
      return `After ${formatTimeValue(condition.after)}`;
    }
    if (condition.before) {
      return `Before ${formatTimeValue(condition.before)}`;
    }
    if (condition.weekday) {
      return `On ${Array.isArray(condition.weekday) ? condition.weekday.join(', ') : condition.weekday}`;
    }
    return 'Time condition';
  }

  /**
   * Generate description for sun conditions
   */
  function generateSunConditionDescription(condition) {
    if (condition.after === 'sunset') {
      return condition.after_offset ? `After sunset (${condition.after_offset})` : 'After sunset';
    }
    if (condition.after === 'sunrise') {
      return condition.after_offset ? `After sunrise (${condition.after_offset})` : 'After sunrise';
    }
    if (condition.before === 'sunset') {
      return 'Before sunset';
    }
    if (condition.before === 'sunrise') {
      return 'Before sunrise';
    }
    return 'Sun position check';
  }

  /**
   * Parse actions
   */
  function parseActions(actions) {
    if (!Array.isArray(actions)) {
      return actions ? [actions] : [];
    }
    return actions.map((action, index) => parseAction(action, index));
  }

  /**
   * Parse single action
   */
  function parseAction(action, index = 0) {
    if (!action) return null;

    // Handle choose blocks
    if (action.choose !== undefined) {
      return {
        type: 'choose',
        id: `choose_${index}`,
        choices: (action.choose || []).map((choice, i) => ({
          conditions: parseConditions(choice.conditions || []),
          sequence: parseActions(choice.sequence || []),
          description: generateChoiceDescription(choice, i),
        })),
        default: action.default ? parseActions(action.default) : null,
      };
    }

    // Handle if/then/else
    if (action.if !== undefined) {
      return {
        type: 'if',
        id: `if_${index}`,
        conditions: parseConditions(action.if || []),
        then: parseActions(action.then || []),
        else: action.else ? parseActions(action.else) : null,
      };
    }

    // Handle service calls
    if (action.service || action.action) {
      return {
        type: 'service',
        id: `service_${index}`,
        service: action.service || action.action,
        target: action.target,
        data: action.data,
        description: generateServiceDescription(action),
      };
    }

    // Handle delay
    if (action.delay !== undefined) {
      return {
        type: 'delay',
        id: `delay_${index}`,
        delay: action.delay,
        description: `Delay: ${formatDuration(action.delay)}`,
      };
    }

    // Handle stop
    if (action.stop !== undefined) {
      return {
        type: 'stop',
        id: `stop_${index}`,
        reason: action.stop,
        description: `Stop: ${action.stop}`,
      };
    }

    // Handle variables
    if (action.variables !== undefined) {
      return {
        type: 'variables',
        id: `vars_${index}`,
        variables: action.variables,
        description: 'Set variables',
      };
    }

    // Handle wait
    if (action.wait_template !== undefined) {
      return {
        type: 'wait',
        id: `wait_${index}`,
        template: action.wait_template,
        description: 'Wait for condition',
      };
    }

    // Default/unknown
    return {
      type: 'unknown',
      id: `action_${index}`,
      raw: action,
      description: 'Action',
    };
  }

  /**
   * Helper to extract string value from potentially HA-typed object
   */
  function getStringValue(value) {
    if (!value) return '';
    if (typeof value === 'object' && value.__ha_type) {
      return `input ${value.value || 'value'}`;
    }
    return String(value);
  }

  /**
   * Generate service call description
   */
  function generateServiceDescription(action) {
    let service = action.service || action.action || 'unknown';
    service = getStringValue(service);

    // Check if we have a friendly name for this service
    if (ACTION_FRIENDLY_NAMES[service]) {
      return ACTION_FRIENDLY_NAMES[service].name;
    }

    // Generate friendly name from service string
    const parts = service.split('.');
    if (parts.length > 1) {
      const domain = parts[0];
      const actionName = parts[1];

      // Special handling for common patterns
      if (actionName === 'turn_on') {
        return `Turn On ${capitalizeFirst(domain)}`;
      }
      if (actionName === 'turn_off') {
        return `Turn Off ${capitalizeFirst(domain)}`;
      }
      if (actionName === 'toggle') {
        return `Toggle ${capitalizeFirst(domain)}`;
      }
      if (actionName.includes('set_')) {
        const setting = actionName.replace('set_', '').replace(/_/g, ' ');
        return `Set ${domain} ${setting}`;
      }

      // Format the action name nicely
      const formatted = actionName.replace(/_/g, ' ').replace(/\b\w/g, (c) => c.toUpperCase());
      return `${capitalizeFirst(domain)}: ${formatted}`;
    }

    return service;
  }

  /**
   * Generate choice description for choose blocks
   */
  function generateChoiceDescription(choice, index) {
    const conditions = choice.conditions || [];
    if (conditions.length === 0) return `Default Path`;

    // Try to find a meaningful identifier from the conditions
    const firstCond = conditions[0];
    if (firstCond?.value_template) {
      // Look for trigger.id check
      const match = firstCond.value_template.match(/trigger\.id\s*[=!]=?\s*['"]([^'"]+)['"]/);
      if (match) {
        const triggerId = match[1];
        const friendlyName = TRIGGER_FRIENDLY_NAMES[triggerId]?.name || formatTriggerId(triggerId);
        return `When: ${friendlyName}`;
      }

      // Look for trigger.id in list
      const inMatch = firstCond.value_template.match(/trigger\.id\s+in\s+\[([^\]]+)\]/);
      if (inMatch) {
        const ids = inMatch[1].match(/['"]([^'"]+)['"]/g);
        if (ids && ids.length > 0) {
          const firstId = ids[0].replace(/['"]/g, '');
          const friendlyName = TRIGGER_FRIENDLY_NAMES[firstId]?.name || formatTriggerId(firstId);
          if (ids.length > 1) {
            return `When: ${friendlyName} (+${ids.length - 1} more)`;
          }
          return `When: ${friendlyName}`;
        }
      }
    }

    // Try to get description from condition parsing
    if (firstCond) {
      const condDesc = generateConditionDescription(firstCond);
      if (condDesc && condDesc !== 'Template check' && condDesc.length < 40) {
        return `If: ${condDesc}`;
      }
    }

    // Analyze the actions to infer what this branch does
    const actions = choice.sequence || [];
    const actionSummary = summarizeBranchActions(actions);
    if (actionSummary) {
      return actionSummary;
    }

    return `Branch ${index + 1}`;
  }

  /**
   * Summarize what a branch does based on its actions
   * Now uses deep extraction to get actual actions instead of stopping at nested blocks
   */
  function summarizeBranchActions(actions) {
    if (!actions || actions.length === 0) return null;

    // First, extract all meaningful actions recursively
    const meaningfulActions = extractMeaningfulActions(actions);

    // Use inferGroupNameFromActions for a human-readable summary
    const inferred = inferGroupNameFromActions(meaningfulActions);
    if (inferred) return inferred;

    // If inference didn't work, analyze the raw actions
    const hasLightOn = actions.some((a) => {
      const svc = a.service || a.action || '';
      return svc.includes('light.turn_on');
    });
    const hasLightOff = actions.some((a) => {
      const svc = a.service || a.action || '';
      return svc.includes('light.turn_off');
    });
    const hasFanOn = actions.some((a) => {
      const svc = a.service || a.action || '';
      return svc.includes('fan.turn_on') || svc.includes('switch.turn_on');
    });
    const hasFanOff = actions.some((a) => {
      const svc = a.service || a.action || '';
      return svc.includes('fan.turn_off') || svc.includes('switch.turn_off');
    });

    // Try to determine the main purpose
    if (hasLightOn && !hasLightOff) return 'Turn on lights';
    if (hasLightOff && !hasLightOn) return 'Turn off lights';
    if (hasFanOn && !hasFanOff) return 'Turn on fan';
    if (hasFanOff && !hasFanOn) return 'Turn off fan';

    // If there are meaningful actions, list the first few
    if (meaningfulActions.length > 0) {
      if (meaningfulActions.length === 1) {
        return meaningfulActions[0];
      }
      return meaningfulActions.slice(0, 2).join(', ');
    }

    return null;
  }

  /**
   * Count total conditions including nested
   */
  function countConditions(conditions) {
    let count = 0;
    for (const cond of conditions) {
      if (cond) {
        count++;
        if (cond.conditions) {
          count += countConditions(cond.conditions);
        }
      }
    }
    return count;
  }

  /**
   * Count conditions in action branches
   */
  function countActionConditions(actions) {
    let count = 0;
    for (const action of actions) {
      if (!action) continue;

      if (action.type === 'choose' && action.choices) {
        for (const choice of action.choices) {
          count += countConditions(choice.conditions || []);
          count += countActionConditions(choice.sequence || []);
        }
        if (action.default) {
          count += countActionConditions(action.default);
        }
      } else if (action.type === 'if') {
        count += countConditions(action.conditions || []);
        count += countActionConditions(action.then || []);
        if (action.else) {
          count += countActionConditions(action.else);
        }
      }
    }
    return count;
  }

  /**
   * Count total actions
   */
  function countActions(actions) {
    let count = 0;
    for (const action of actions) {
      if (!action) continue;

      if (action.type === 'choose' && action.choices) {
        for (const choice of action.choices) {
          count += countActions(choice.sequence || []);
        }
        if (action.default) {
          count += countActions(action.default);
        }
      } else if (action.type === 'if') {
        count += countActions(action.then || []);
        if (action.else) {
          count += countActions(action.else);
        }
      } else if (action.type === 'service') {
        count++;
      }
    }
    return count;
  }

  /**
   * Count decision branches
   */
  function countBranches(actions) {
    let count = 0;
    for (const action of actions) {
      if (!action) continue;

      if (action.type === 'choose') {
        count += action.choices?.length || 0;
        if (action.default) count++;
        for (const choice of action.choices || []) {
          count += countBranches(choice.sequence || []);
        }
        if (action.default) {
          count += countBranches(action.default);
        }
      } else if (action.type === 'if') {
        count += 2; // then and else branches
        count += countBranches(action.then || []);
        if (action.else) {
          count += countBranches(action.else);
        }
      }
    }
    return count;
  }

  /**
   * Show placeholder state
   */
  function showPlaceholder() {
    if (diagramPlaceholder) diagramPlaceholder.style.display = 'flex';
    if (diagramLoading) diagramLoading.style.display = 'none';
    if (matrixContent) matrixContent.style.display = 'none';
    if (vizInfo) vizInfo.style.display = 'none';
    parsedData = null;
  }

  /**
   * Show loading state
   */
  function showLoading() {
    if (diagramPlaceholder) diagramPlaceholder.style.display = 'none';
    if (diagramLoading) diagramLoading.style.display = 'flex';
    if (matrixContent) matrixContent.style.display = 'none';
  }

  /**
   * Show error state
   */
  function showError(message) {
    if (diagramPlaceholder) {
      diagramPlaceholder.innerHTML = `
                <svg class="placeholder-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" style="color: #f87171;">
                    <circle cx="12" cy="12" r="10"/>
                    <line x1="15" y1="9" x2="9" y2="15"/>
                    <line x1="9" y1="9" x2="15" y2="15"/>
                </svg>
                <h3>Error Loading Blueprint</h3>
                <p>${message}</p>
            `;
      diagramPlaceholder.style.display = 'flex';
    }
    if (diagramLoading) diagramLoading.style.display = 'none';
  }

  /**
   * Update info panel with parsed data
   */
  function updateInfoPanel(data) {
    if (!vizInfo) return;

    // Get blueprint summary for enhanced info
    const summary = BLUEPRINT_SUMMARIES[currentBlueprint];

    document.getElementById('info-name').textContent = data.name;
    document.getElementById('info-version').textContent = data.version;
    document.getElementById('info-domain').textContent = data.domain;
    document.getElementById('stat-triggers').textContent = data.stats.triggers;
    document.getElementById('stat-conditions').textContent = data.stats.conditions;
    document.getElementById('stat-actions').textContent = data.stats.actions;
    document.getElementById('stat-branches').textContent = data.stats.branches;

    // Add purpose description to the overview card if we have summary info
    const infoCardContent = document.querySelector('.info-card-content');
    if (infoCardContent && summary) {
      // Check if we already added a purpose row
      let purposeRow = document.getElementById('info-purpose-row');
      if (!purposeRow) {
        purposeRow = document.createElement('div');
        purposeRow.id = 'info-purpose-row';
        purposeRow.className = 'info-row info-purpose';
        infoCardContent.insertBefore(purposeRow, infoCardContent.firstChild);
      }
      purposeRow.innerHTML = `
                <span class="info-label">Purpose</span>
                <span class="info-value">${summary.userBenefit}</span>
            `;
    }

    vizInfo.style.display = 'block';
  }

  /**
   * Render the logic matrix (main and only view)
   */
  function renderVisualization() {
    if (!parsedData) return;

    if (diagramLoading) diagramLoading.style.display = 'none';
    if (matrixContent) matrixContent.style.display = 'block';

    const matrixHtml = generateDecisionMatrixHtml(parsedData);
    const matrixWrapper = document.getElementById('decision-matrix');
    if (matrixWrapper) {
      matrixWrapper.innerHTML = matrixHtml;
    }
  }

  /**
   * Generate Mermaid decision tree diagram - Human-Friendly Version
   * Uses the same per-trigger logic as the matrix view for consistency
   * Shows each trigger with its specific conditions and actions
   */
  function generateDecisionTreeMermaid(data) {
    const lines = ['graph TD'];

    // Get blueprint summary for context
    const summary = BLUEPRINT_SUMMARIES[currentBlueprint];

    // Add a title/purpose node if we have summary info
    if (summary) {
      lines.push(`    PURPOSE["🎯 ${escapeLabel(summary.userBenefit)}"]:::purpose`);
      lines.push('');
    }

    // Group triggers by category for better organization
    const triggerGroups = groupTriggersByCategory(data.triggers);
    let triggerIndex = 0;

    // For each trigger, show its complete path: trigger -> conditions -> actions
    Object.entries(triggerGroups).forEach(([category, triggers]) => {
      triggers.forEach((trigger, i) => {
        const triggerId = trigger.id || `trigger_${triggerIndex}`;
        const nodePrefix = `T${triggerIndex}`;

        // Get conditions and actions for this specific trigger
        const conditions = getConditionsForTrigger(triggerId, data);
        const actions = getActionsForTrigger(triggerId, data);

        // Get trigger icon
        const triggerIcon = getTriggerIcon(triggerId);

        // Create trigger node
        const triggerLabel = escapeLabel(`${triggerIcon} ${trigger.description}`);
        lines.push(`    ${nodePrefix}["${triggerLabel}"]:::trigger`);

        // Connect purpose to first few triggers if present
        if (summary && triggerIndex < 3) {
          lines.push(`    PURPOSE -.-> ${nodePrefix}`);
        }

        // Create condition node if there are conditions
        if (conditions.length > 0) {
          const condLabel = conditions
            .slice(0, 2)
            .map((c) => `✓ ${c}`)
            .join('\\n');
          const condNodeId = `${nodePrefix}C`;
          lines.push(`    ${condNodeId}{{"${escapeLabel(condLabel)}"}}:::condition`);
          lines.push(`    ${nodePrefix} --> ${condNodeId}`);

          // Create action nodes connected to conditions
          if (actions.length > 0) {
            actions.slice(0, 2).forEach((action, j) => {
              const actionIcon = getActionIcon(action);
              const actionNodeId = `${nodePrefix}A${j}`;
              lines.push(`    ${actionNodeId}["${actionIcon} ${escapeLabel(action)}"]:::action`);
              lines.push(`    ${condNodeId} --> ${actionNodeId}`);
            });
            if (actions.length > 2) {
              lines.push(`    ${nodePrefix}AM["+ ${actions.length - 2} more"]:::muted`);
              lines.push(`    ${condNodeId} --> ${nodePrefix}AM`);
            }
          }
        } else if (actions.length > 0) {
          // No conditions, connect trigger directly to actions
          actions.slice(0, 2).forEach((action, j) => {
            const actionIcon = getActionIcon(action);
            const actionNodeId = `${nodePrefix}A${j}`;
            lines.push(`    ${actionNodeId}["${actionIcon} ${escapeLabel(action)}"]:::action`);
            lines.push(`    ${nodePrefix} --> ${actionNodeId}`);
          });
          if (actions.length > 2) {
            lines.push(`    ${nodePrefix}AM["+ ${actions.length - 2} more"]:::muted`);
            lines.push(`    ${nodePrefix} --> ${nodePrefix}AM`);
          }
        }

        lines.push('');
        triggerIndex++;
      });
    });

    // Add styling classes with improved colors
    lines.push('');
    lines.push('    classDef purpose fill:#10b98120,stroke:#10b981,color:#e2e8f0,font-weight:bold');
    lines.push('    classDef trigger fill:#22d3ee20,stroke:#22d3ee,color:#e2e8f0');
    lines.push('    classDef condition fill:#fbbf2420,stroke:#fbbf24,color:#e2e8f0');
    lines.push('    classDef action fill:#a78bfa20,stroke:#a78bfa,color:#e2e8f0');
    lines.push('    classDef choose fill:#ec489920,stroke:#ec4899,color:#e2e8f0');
    lines.push('    classDef muted fill:#64748b20,stroke:#64748b,color:#94a3b8');

    return lines.join('\n');
  }

  /**
   * Generate Mermaid flow diagram - Human-Friendly Version
   * Uses the same per-trigger logic as the matrix view
   * Shows trigger categories with their specific conditions and actions
   */
  function generateFlowDiagramMermaid(data) {
    const lines = ['flowchart TB'];
    const summary = BLUEPRINT_SUMMARIES[currentBlueprint];

    // Add title with purpose
    const title = summary ? summary.userBenefit : 'Automation Flow';
    lines.push(`    TITLE["🎯 ${escapeLabel(title)}"]:::purpose`);
    lines.push('');

    // Group triggers by category
    const triggerGroups = groupTriggersByCategory(data.triggers);

    // Create subgraphs for each trigger category with their full logic paths
    let categoryIndex = 0;
    Object.entries(triggerGroups).forEach(([category, triggers]) => {
      const categoryName = getCategoryDisplayName(category);
      const subgraphId = `CAT${categoryIndex}`;

      lines.push(`    subgraph ${subgraphId}["${categoryName}"]`);
      lines.push('    direction TB');

      triggers.forEach((trigger, i) => {
        const triggerId = trigger.id || `trigger_${categoryIndex}_${i}`;
        const nodePrefix = `${subgraphId}T${i}`;

        // Get conditions and actions for this specific trigger
        const conditions = getConditionsForTrigger(triggerId, data);
        const actions = getActionsForTrigger(triggerId, data);

        // Get trigger icon
        const triggerIcon = getTriggerIcon(triggerId);

        // Trigger node
        lines.push(`    ${nodePrefix}["${triggerIcon} ${escapeLabel(trigger.description)}"]:::trigger`);

        // Show the flow for this trigger
        if (conditions.length > 0 && actions.length > 0) {
          // Condition check node
          const condText = conditions.slice(0, 2).join(', ');
          const condNodeId = `${nodePrefix}C`;
          lines.push(`    ${condNodeId}{{"If: ${escapeLabel(condText)}"}}:::condition`);
          lines.push(`    ${nodePrefix} --> ${condNodeId}`);

          // Action node
          const actionText = actions.slice(0, 2).join(', ');
          const actionIcon = getActionIcon(actions[0]);
          const actionNodeId = `${nodePrefix}A`;
          lines.push(`    ${actionNodeId}["${actionIcon} ${escapeLabel(actionText)}"]:::action`);
          lines.push(`    ${condNodeId} -->|Yes| ${actionNodeId}`);
        } else if (actions.length > 0) {
          // Direct action without conditions
          const actionText = actions.slice(0, 2).join(', ');
          const actionIcon = getActionIcon(actions[0]);
          const actionNodeId = `${nodePrefix}A`;
          lines.push(`    ${actionNodeId}["${actionIcon} ${escapeLabel(actionText)}"]:::action`);
          lines.push(`    ${nodePrefix} --> ${actionNodeId}`);
        }
      });

      lines.push('    end');
      lines.push('');
      categoryIndex++;
    });

    // Connect title to all category subgraphs
    Object.keys(triggerGroups).forEach((_, i) => {
      lines.push(`    TITLE --> CAT${i}`);
    });

    // Add Done node
    lines.push('');
    lines.push('    DONE(("✓ Done")):::done');

    // Connect all categories to Done
    Object.keys(triggerGroups).forEach((_, i) => {
      lines.push(`    CAT${i} --> DONE`);
    });

    // Add styling
    lines.push('');
    lines.push('    classDef purpose fill:#10b98120,stroke:#10b981,color:#e2e8f0,font-weight:bold');
    lines.push('    classDef trigger fill:#22d3ee20,stroke:#22d3ee,color:#e2e8f0');
    lines.push('    classDef condition fill:#fbbf2420,stroke:#fbbf24,color:#e2e8f0');
    lines.push('    classDef action fill:#a78bfa20,stroke:#a78bfa,color:#e2e8f0');
    lines.push('    classDef done fill:#10b98140,stroke:#10b981,color:#e2e8f0');

    return lines.join('\n');
  }

  /**
   * Group triggers by their functional category
   */
  function groupTriggersByCategory(triggers) {
    const groups = {
      motion: [],
      door: [],
      humidity: [],
      light: [],
      time: [],
      system: [],
      other: [],
    };

    triggers.forEach((trigger) => {
      const id = (trigger.id || '').toLowerCase();
      const desc = (trigger.description || '').toLowerCase();

      if (id.includes('motion') || desc.includes('motion')) {
        groups.motion.push(trigger);
      } else if (id.includes('door') || desc.includes('door')) {
        groups.door.push(trigger);
      } else if (id.includes('humidity') || desc.includes('humidity')) {
        groups.humidity.push(trigger);
      } else if (id.includes('light') || desc.includes('light')) {
        groups.light.push(trigger);
      } else if (id.includes('time') || desc.includes('time') || trigger.platform === 'time') {
        groups.time.push(trigger);
      } else if (id.includes('ha_start') || id.includes('homeassistant') || trigger.platform === 'homeassistant') {
        groups.system.push(trigger);
      } else {
        groups.other.push(trigger);
      }
    });

    // Remove empty groups
    return Object.fromEntries(Object.entries(groups).filter(([_, triggers]) => triggers.length > 0));
  }

  /**
   * Get display name for trigger category
   */
  function getCategoryDisplayName(category) {
    const names = {
      motion: 'Motion Events',
      door: 'Door Events',
      humidity: 'Humidity Changes',
      light: 'Light Events',
      time: 'Scheduled Times',
      system: 'System Events',
      other: 'Other Events',
    };
    return names[category] || 'Events';
  }

  /**
   * Analyze actions and group them by functionality
   * Creates human-readable groups based on what the actions actually do
   */
  function analyzeAndGroupActions(actions) {
    const groups = [];

    for (const action of actions) {
      if (!action) continue;

      if (action.type === 'choose' && action.choices) {
        // For choose blocks, create groups based on what each choice actually does
        action.choices.forEach((choice, i) => {
          // Get a meaningful name for this choice
          let groupName = choice.description || '';

          // If the description isn't helpful, try to infer from the actions
          if (!groupName || groupName.includes('Branch') || groupName.includes('Option')) {
            const outcomes = extractMeaningfulActions(choice.sequence || []);
            groupName = inferGroupNameFromActions(outcomes) || `Path ${i + 1}`;
          }

          // Get actual outcomes (not "Evaluate N sub-conditions")
          const outcomes = extractMeaningfulActions(choice.sequence || []);

          // Only add if there are meaningful outcomes
          if (outcomes.length > 0 || groupName) {
            groups.push({
              name: groupName,
              isDecision: true,
              outcomes: outcomes,
            });
          }
        });

        if (action.default && action.default.length > 0) {
          const outcomes = extractMeaningfulActions(action.default);
          if (outcomes.length > 0) {
            groups.push({
              name: 'Default behavior',
              isDecision: true,
              outcomes: outcomes,
            });
          }
        }
      } else if (action.type === 'if') {
        const thenOutcomes = extractMeaningfulActions(action.then || []);
        if (thenOutcomes.length > 0) {
          const groupName = inferGroupNameFromActions(thenOutcomes) || 'When condition met';
          groups.push({
            name: groupName,
            isDecision: true,
            outcomes: thenOutcomes,
          });
        }
        if (action.else) {
          const elseOutcomes = extractMeaningfulActions(action.else);
          if (elseOutcomes.length > 0) {
            const groupName = inferGroupNameFromActions(elseOutcomes) || 'Otherwise';
            groups.push({
              name: groupName,
              isDecision: true,
              outcomes: elseOutcomes,
            });
          }
        }
      } else if (action.type === 'service') {
        // Skip logging/internal actions in the main view
        const svc = action.service || action.action || '';
        if (!svc.includes('logbook') && !svc.includes('input_boolean') && !svc.includes('input_datetime')) {
          groups.push({
            name: action.description,
            isDecision: false,
            outcomes: [],
          });
        }
      }
    }

    // Consolidate similar groups
    return consolidateGroups(groups);
  }

  /**
   * Infer a human-readable group name from the actions it contains
   */
  function inferGroupNameFromActions(actions) {
    if (!actions || actions.length === 0) return null;

    const actionStr = actions.join(' ').toLowerCase();

    // Light control patterns
    if (actionStr.includes('turn on light') || actionStr.includes('light.turn_on')) {
      if (actionStr.includes('night') || actionStr.includes('dim')) {
        return 'Night mode lighting';
      }
      return 'Turn on lights';
    }
    if (actionStr.includes('turn off light') || actionStr.includes('light.turn_off')) {
      return 'Turn off lights';
    }

    // Fan control patterns
    if (
      actionStr.includes('turn on fan') ||
      actionStr.includes('fan.turn_on') ||
      actionStr.includes('switch.turn_on')
    ) {
      return 'Turn on fan';
    }
    if (
      actionStr.includes('turn off fan') ||
      actionStr.includes('fan.turn_off') ||
      actionStr.includes('switch.turn_off')
    ) {
      return 'Turn off fan';
    }

    // Climate control
    if (actionStr.includes('climate') || actionStr.includes('hvac') || actionStr.includes('temperature')) {
      return 'Adjust climate';
    }

    // Cover/shade control
    if (actionStr.includes('cover') || actionStr.includes('shade') || actionStr.includes('blind')) {
      return 'Adjust shades';
    }

    // Delay patterns
    if (actionStr.includes('delay') || actionStr.includes('wait')) {
      return 'Wait period';
    }

    // If there's exactly one action, use it as the name
    if (actions.length === 1) {
      return actions[0];
    }

    // Multiple actions - try to find a common theme
    const hasLight = actionStr.includes('light');
    const hasFan = actionStr.includes('fan');

    if (hasLight && hasFan) {
      return 'Control light and fan';
    }
    if (hasLight) {
      return 'Light control';
    }
    if (hasFan) {
      return 'Fan control';
    }

    return null;
  }

  /**
   * Extract meaningful (non-debug) actions from a sequence
   * Recursively traverses nested structures to get actual leaf actions
   */
  function extractMeaningfulActions(actions, depth = 0) {
    const meaningful = [];
    if (depth > 8) return meaningful; // Prevent infinite recursion

    for (const action of actions) {
      if (!action) continue;

      if (action.type === 'service') {
        const svc = action.service || action.action || '';
        // Skip debug/logging/internal actions
        if (
          !svc.includes('logbook') &&
          !svc.includes('input_boolean') &&
          !svc.includes('input_datetime') &&
          action.description
        ) {
          if (!meaningful.includes(action.description)) {
            meaningful.push(action.description);
          }
        }
      } else if (action.type === 'delay') {
        const delayDesc = action.description || 'Wait';
        if (!meaningful.includes(delayDesc)) {
          meaningful.push(delayDesc);
        }
      } else if (action.type === 'choose' && action.choices) {
        // Recurse into each choice to get actual actions
        for (const choice of action.choices) {
          const nestedActions = extractMeaningfulActions(choice.sequence || [], depth + 1);
          nestedActions.forEach((a) => {
            if (!meaningful.includes(a)) {
              meaningful.push(a);
            }
          });
        }
        // Also check default
        if (action.default) {
          const defaultActions = extractMeaningfulActions(action.default, depth + 1);
          defaultActions.forEach((a) => {
            if (!meaningful.includes(a)) {
              meaningful.push(a);
            }
          });
        }
      } else if (action.type === 'if') {
        // Recurse into then/else branches
        const thenActions = extractMeaningfulActions(action.then || [], depth + 1);
        thenActions.forEach((a) => {
          if (!meaningful.includes(a)) {
            meaningful.push(a);
          }
        });
        if (action.else) {
          const elseActions = extractMeaningfulActions(action.else, depth + 1);
          elseActions.forEach((a) => {
            if (!meaningful.includes(a)) {
              meaningful.push(a);
            }
          });
        }
      } else if (action.type === 'stop') {
        meaningful.push('Stop automation');
      } else if (action.type === 'variables') {
        // Skip variable assignments - internal
      }
    }

    return meaningful;
  }

  /**
   * Consolidate similar groups to reduce diagram complexity
   */
  function consolidateGroups(groups) {
    // If we have too many groups, consolidate by action type
    if (groups.length <= 8) {
      return groups;
    }

    const consolidated = [];
    const lightActions = [];
    const fanActions = [];
    const otherActions = [];

    groups.forEach((group) => {
      const name = (group.name || '').toLowerCase();
      if (name.includes('light')) {
        lightActions.push(group);
      } else if (name.includes('fan')) {
        fanActions.push(group);
      } else {
        otherActions.push(group);
      }
    });

    if (lightActions.length > 0) {
      consolidated.push({
        name: 'Control Lights',
        isDecision: true,
        outcomes: lightActions.map((g) => g.name),
      });
    }

    if (fanActions.length > 0) {
      consolidated.push({
        name: 'Control Fan',
        isDecision: true,
        outcomes: fanActions.map((g) => g.name),
      });
    }

    // Add other actions (limited)
    otherActions.slice(0, 4).forEach((group) => {
      consolidated.push(group);
    });

    return consolidated;
  }

  /**
   * Extract key conditions that are worth displaying
   */
  function extractKeyConditions(data) {
    const conditions = [];

    // Look for common condition patterns in the variables
    if (data.variables) {
      if (data.variables.presence_ok !== undefined) {
        conditions.push('Someone is home?');
      }
      if (data.variables.lux_ok !== undefined) {
        conditions.push('Room dark enough?');
      }
      if (data.variables.night_mode_active !== undefined) {
        conditions.push('Night mode?');
      }
      if (data.variables.humidity_delta !== undefined) {
        conditions.push('Humidity too high?');
      }
    }

    // Also extract from parsed conditions
    data.conditions.forEach((cond) => {
      if (cond && cond.description && cond.description.length < 30) {
        conditions.push(cond.description);
      }
    });

    return conditions.slice(0, 4);
  }

  /**
   * Extract top-level branches from actions
   */
  function extractTopLevelBranches(actions) {
    const branches = [];

    for (const action of actions) {
      if (!action) continue;

      if (action.type === 'choose' && action.choices) {
        for (const choice of action.choices) {
          branches.push({
            label: choice.description || 'Branch',
            isCondition: true,
            actions: extractActionLabels(choice.sequence || []),
          });
        }
        if (action.default) {
          branches.push({
            label: 'Default',
            isCondition: true,
            actions: extractActionLabels(action.default),
          });
        }
      } else if (action.type === 'if') {
        branches.push({
          label: 'If condition',
          isCondition: true,
          actions: extractActionLabels(action.then || []),
        });
        if (action.else) {
          branches.push({
            label: 'Else',
            isCondition: true,
            actions: extractActionLabels(action.else),
          });
        }
      } else if (action.description) {
        branches.push({
          label: action.description,
          isCondition: false,
          actions: [],
        });
      }
    }

    return branches;
  }

  /**
   * Extract action labels for display
   */
  function extractActionLabels(actions) {
    const labels = [];
    for (const action of actions) {
      if (!action) continue;
      if (action.type === 'service' && action.description) {
        labels.push(action.description);
      } else if (action.type === 'choose') {
        labels.push(`Choose (${action.choices?.length || 0} branches)`);
      } else if (action.description) {
        labels.push(action.description);
      }
    }
    return labels;
  }

  /**
   * Generate HTML decision matrix - Human-Friendly Version
   * Creates a comprehensive table showing triggers, conditions, actions, and why
   */
  function generateDecisionMatrixHtml(data) {
    const summary = BLUEPRINT_SUMMARIES[currentBlueprint];
    const blueprintLogic = BLUEPRINT_TRIGGER_LOGIC[currentBlueprint] || {};

    let html = '<div class="matrix-container">';

    // Add blueprint summary header with rich description
    if (summary) {
      html += '<div class="matrix-summary">';
      html += `<div class="summary-header">`;
      html += `<h3>${escapeHtml(summary.purpose)}</h3>`;
      html += `<p class="summary-benefit">✨ ${escapeHtml(summary.userBenefit)}</p>`;
      html += `</div>`;
      html += '<div class="feature-tags">';
      summary.mainFeatures.forEach((feature) => {
        html += `<span class="feature-tag">${escapeHtml(feature)}</span>`;
      });
      html += '</div>';
      html += '</div>';
    }

    // Create the main logic table with 4 columns
    html += '<table class="decision-matrix">';
    html += '<thead><tr>';
    html += '<th class="th-trigger"><span class="th-icon">🎯</span> When This Happens</th>';
    html += '<th class="th-conditions"><span class="th-icon">✓</span> If These Are True</th>';
    html += '<th class="th-actions"><span class="th-icon">⚡</span> Then Do This</th>';
    html += '<th class="th-why"><span class="th-icon">💡</span> Why</th>';
    html += '</tr></thead>';
    html += '<tbody>';

    // Group triggers by category for better organization
    const triggerGroups = groupTriggersByCategory(data.triggers);

    // Add category headers and rows
    let rowIndex = 0;
    Object.entries(triggerGroups).forEach(([category, triggers]) => {
      const categoryName = getCategoryDisplayName(category);

      // Add category header row
      html += `<tr class="category-row">`;
      html += `<td colspan="4" class="category-header">`;
      html += `<span class="category-icon">${getCategoryIcon(category)}</span>`;
      html += `<span class="category-name">${escapeHtml(categoryName)}</span>`;
      html += `<span class="category-count">${triggers.length} trigger${triggers.length > 1 ? 's' : ''}</span>`;
      html += `</td></tr>`;

      triggers.forEach((trigger) => {
        const rowClass = rowIndex % 2 === 0 ? 'row-even' : 'row-odd';
        const triggerId = trigger.id || '';
        const triggerLogic = blueprintLogic[triggerId];

        html += `<tr class="${rowClass}">`;

        // Column 1: Trigger with icon and description
        const triggerIcon = getTriggerIcon(triggerId);
        const triggerFriendly = TRIGGER_FRIENDLY_NAMES[triggerId];
        html += `<td class="trigger-cell">`;
        html += `<div class="trigger-content">`;
        html += `<div class="trigger-main">`;
        html += `<span class="trigger-icon">${triggerIcon}</span>`;
        html += `<span class="trigger-name">${escapeHtml(trigger.description)}</span>`;
        html += `</div>`;
        if (triggerFriendly && triggerFriendly.description) {
          html += `<div class="trigger-hint">${escapeHtml(triggerFriendly.description)}</div>`;
        }
        html += `</div>`;
        html += `</td>`;

        // Column 2: Conditions with checkmarks
        html += `<td class="condition-cell">`;
        html += `<div class="conditions-list">`;
        const relevantConditions = getConditionsForTrigger(triggerId, data);
        if (relevantConditions.length > 0) {
          relevantConditions.forEach((cond) => {
            html += `<div class="condition-item">`;
            html += `<span class="condition-check">✓</span>`;
            html += `<span class="condition-text">${escapeHtml(cond)}</span>`;
            html += `</div>`;
          });
        } else {
          html += `<div class="condition-item muted">`;
          html += `<span class="condition-check">—</span>`;
          html += `<span class="condition-text">Always executes</span>`;
          html += `</div>`;
        }
        html += `</div></td>`;

        // Column 3: Actions with icons
        html += `<td class="action-cell">`;
        html += `<div class="actions-list">`;
        const relevantActions = getActionsForTrigger(triggerId, data);
        if (relevantActions.length > 0) {
          relevantActions.forEach((action) => {
            const actionIcon = getActionIcon(action);
            html += `<div class="action-item">`;
            html += `<span class="action-icon">${actionIcon}</span>`;
            html += `<span class="action-text">${escapeHtml(action)}</span>`;
            html += `</div>`;
          });
        } else {
          html += `<div class="action-item muted">`;
          html += `<span class="action-icon">⚡</span>`;
          html += `<span class="action-text">Evaluate further conditions</span>`;
          html += `</div>`;
        }
        html += `</div></td>`;

        // Column 4: Why/Purpose description
        html += `<td class="why-cell">`;
        if (triggerLogic && triggerLogic.description) {
          html += `<div class="why-content">`;
          html += `<span class="why-text">${escapeHtml(triggerLogic.description)}</span>`;
          html += `</div>`;
        } else {
          // Generate a description based on the trigger type
          const inferredWhy = inferWhyForTrigger(triggerId, relevantActions);
          html += `<div class="why-content muted">`;
          html += `<span class="why-text">${escapeHtml(inferredWhy)}</span>`;
          html += `</div>`;
        }
        html += `</td>`;

        html += '</tr>';
        rowIndex++;
      });
    });

    html += '</tbody></table>';
    html += '</div>';

    return html;
  }

  /**
   * Get icon for a category
   */
  function getCategoryIcon(category) {
    const icons = {
      motion: '🚶',
      door: '🚪',
      humidity: '💧',
      light: '💡',
      fan: '💨',
      time: '⏰',
      system: '🏠',
      button: '🔘',
      switch: '🔘',
      temperature: '🌡️',
      presence: '👥',
      other: '📋',
    };
    return icons[category] || '📋';
  }

  /**
   * Infer a "why" description based on trigger and actions
   */
  function inferWhyForTrigger(triggerId, actions) {
    const id = (triggerId || '').toLowerCase();
    const actionStr = (actions || []).join(' ').toLowerCase();

    // Motion/presence triggers
    if (id.includes('motion') || id.includes('presence')) {
      if (actionStr.includes('light') && actionStr.includes('on')) {
        return 'Automatically light the room when someone enters';
      }
      if (actionStr.includes('light') && actionStr.includes('off')) {
        return 'Turn off lights when room becomes empty';
      }
      if (actionStr.includes('fan')) {
        return 'Control fan based on room occupancy';
      }
      return 'Respond to room occupancy changes';
    }

    // Door triggers
    if (id.includes('door')) {
      if (actionStr.includes('light')) {
        return 'Control lights when door opens/closes';
      }
      return 'Respond to door state changes';
    }

    // Humidity triggers
    if (id.includes('humidity')) {
      if (actionStr.includes('fan')) {
        return 'Control exhaust fan based on moisture levels';
      }
      return 'Respond to humidity changes';
    }

    // Temperature triggers
    if (id.includes('temp')) {
      if (actionStr.includes('fan')) {
        return 'Adjust fan based on temperature';
      }
      if (actionStr.includes('hvac') || actionStr.includes('climate')) {
        return 'Maintain comfort temperature';
      }
      return 'Respond to temperature changes';
    }

    // Button/switch triggers
    if (
      id.includes('single') ||
      id.includes('double') ||
      id.includes('triple') ||
      id.includes('hold') ||
      id.includes('release')
    ) {
      return 'Execute configured button action';
    }

    // Time triggers
    if (id.includes('time') || id.includes('tick') || id.includes('periodic')) {
      return 'Scheduled check and adjustment';
    }

    // System triggers
    if (id.includes('ha_start') || id.includes('start')) {
      return 'Initialize state after system restart';
    }

    // Manual override triggers
    if (id.includes('manual')) {
      return 'Respect manual control preferences';
    }

    // Max runtime/safety triggers
    if (id.includes('max_runtime') || id.includes('expired')) {
      return 'Safety timeout protection';
    }

    return 'Automation logic path';
  }

  /**
   * Get icon for a trigger based on its ID
   */
  function getTriggerIcon(triggerId) {
    const id = (triggerId || '').toLowerCase();
    if (id.includes('motion') || id.includes('wasp_motion')) return '🚶';
    if (id.includes('door')) return '🚪';
    if (id.includes('humidity')) return '💧';
    if (id.includes('light')) return '💡';
    if (id.includes('fan')) return '💨';
    if (id.includes('time') || id.includes('schedule')) return '⏰';
    if (id.includes('temp')) return '🌡️';
    if (id.includes('sun')) return '🌅';
    if (id.includes('ha_start') || id.includes('start')) return '🏠';
    return '▶️';
  }

  /**
   * Get icon for an action based on its description
   */
  function getActionIcon(actionDesc) {
    const desc = (actionDesc || '').toLowerCase();
    if (desc.includes('turn on light') || desc.includes('light on')) return '💡';
    if (desc.includes('turn off light') || desc.includes('light off')) return '🔌';
    if (desc.includes('turn on fan') || desc.includes('fan on')) return '💨';
    if (desc.includes('turn off fan') || desc.includes('fan off')) return '🔇';
    if (desc.includes('delay') || desc.includes('wait')) return '⏳';
    if (desc.includes('set')) return '⚙️';
    if (desc.includes('log')) return '📝';
    return '⚡';
  }

  /**
   * Get conditions that are relevant to a specific trigger
   * Uses explicit mappings first, then falls back to parsing
   */
  function getConditionsForTrigger(triggerId, data) {
    // First check if we have explicit mappings for this blueprint/trigger
    const blueprintLogic = BLUEPRINT_TRIGGER_LOGIC[currentBlueprint];
    if (blueprintLogic && blueprintLogic[triggerId]) {
      return blueprintLogic[triggerId].conditions.slice(0, 4);
    }

    // Fall back to parsing
    const conditions = [];
    const id = (triggerId || '').toLowerCase();

    // Deep-search for conditions related to this trigger
    function extractConditionsFromActions(actions, targetTriggerId, depth = 0) {
      if (depth > 5) return; // Prevent infinite recursion

      for (const action of actions) {
        if (!action) continue;

        if (action.type === 'choose' && action.choices) {
          for (const choice of action.choices) {
            // Check if this choice is for this trigger
            const isForTrigger = choice.conditions?.some((c) => {
              const template = c.value_template || '';
              return template.toLowerCase().includes(targetTriggerId);
            });

            if (isForTrigger) {
              // Extract conditions from this choice (skip trigger ID matches)
              choice.conditions?.forEach((c) => {
                const desc = generateConditionDescription(c);
                if (desc && !desc.includes('Triggered by') && !desc.includes('Template check') && desc.length < 60) {
                  if (!conditions.includes(desc)) {
                    conditions.push(desc);
                  }
                }
              });

              // Also look for nested conditions
              extractConditionsFromActions(choice.sequence || [], targetTriggerId, depth + 1);
            }
          }
        } else if (action.type === 'if' && action.conditions) {
          action.conditions.forEach((c) => {
            const desc = generateConditionDescription(c);
            if (desc && !desc.includes('Template check') && desc.length < 60) {
              if (!conditions.includes(desc)) {
                conditions.push(desc);
              }
            }
          });
          extractConditionsFromActions(action.then || [], targetTriggerId, depth + 1);
        }
      }
    }

    extractConditionsFromActions(data.actions, triggerId);

    // Also check top-level conditions
    data.conditions.forEach((c) => {
      if (c && c.description && c.description.length < 50 && !c.description.includes('Template check')) {
        if (!conditions.includes(c.description)) {
          conditions.push(c.description);
        }
      }
    });

    // If still no conditions, add inferred ones based on common patterns in variables
    if (conditions.length === 0 && data.variables) {
      if (id.includes('motion') || id.includes('door')) {
        if (data.variables.presence_ok !== undefined) conditions.push('Someone is home');
        if (data.variables.override_ok !== undefined) conditions.push('No manual override active');
        if (data.variables.lux_ok !== undefined) conditions.push('Room is dark enough');
        if (data.variables.should_turn_on_light !== undefined) conditions.push('Light should turn on');
      }
      if (id.includes('humidity')) {
        if (data.variables.humidity_sensors_ok !== undefined) conditions.push('Humidity sensors working');
        if (data.variables.humidity_delta !== undefined) conditions.push('Humidity delta calculated');
      }
    }

    return conditions.slice(0, 5);
  }

  /**
   * Get actions that are relevant to a specific trigger
   * Uses explicit mappings first, then deep-parses the action tree
   */
  function getActionsForTrigger(triggerId, data) {
    // First check if we have explicit mappings for this blueprint/trigger
    const blueprintLogic = BLUEPRINT_TRIGGER_LOGIC[currentBlueprint];
    if (blueprintLogic && blueprintLogic[triggerId]) {
      return blueprintLogic[triggerId].actions.slice(0, 5);
    }

    // Fall back to deep parsing
    const actions = [];
    const id = (triggerId || '').toLowerCase();

    /**
     * Recursively extract all service calls from nested structures
     */
    function deepExtractActions(actionList, depth = 0) {
      if (depth > 10 || !actionList) return; // Prevent infinite recursion

      for (const action of actionList) {
        if (!action) continue;

        if (action.type === 'service') {
          const svc = action.service || action.action || '';
          // Skip debug/logging/helper actions
          if (!svc.includes('logbook') && !svc.includes('input_boolean') && !svc.includes('input_datetime')) {
            const desc = action.description;
            if (desc && !actions.includes(desc)) {
              actions.push(desc);
            }
          }
        } else if (action.type === 'choose' && action.choices) {
          // Recurse into each choice
          for (const choice of action.choices) {
            deepExtractActions(choice.sequence || [], depth + 1);
          }
          // Also check default
          if (action.default) {
            deepExtractActions(action.default, depth + 1);
          }
        } else if (action.type === 'if') {
          deepExtractActions(action.then || [], depth + 1);
          if (action.else) {
            deepExtractActions(action.else, depth + 1);
          }
        } else if (action.type === 'delay' && action.description) {
          if (!actions.includes(action.description)) {
            actions.push(action.description);
          }
        }
      }
    }

    // Find the branch for this trigger and extract all nested actions
    for (const action of data.actions) {
      if (!action) continue;

      if (action.type === 'choose' && action.choices) {
        for (const choice of action.choices) {
          // Check if this choice is for this trigger
          const isForTrigger = choice.conditions?.some((c) => {
            const template = c.value_template || '';
            return template.toLowerCase().includes(triggerId);
          });

          if (isForTrigger) {
            deepExtractActions(choice.sequence || []);
          }
        }
      }
    }

    // Deduplicate and clean up
    const uniqueActions = [...new Set(actions)];

    // If no specific actions found, show inferred actions based on trigger type
    if (uniqueActions.length === 0) {
      if (id.includes('wasp_motion') || id.includes('wasp_door_opened')) {
        uniqueActions.push('Turn on light');
        uniqueActions.push('Apply night mode if active');
      } else if (id.includes('wasp_motion_clear') || id.includes('wasp_door_left_open')) {
        uniqueActions.push('Turn off light');
      } else if (id.includes('humidity')) {
        uniqueActions.push('Turn on/off fan based on humidity');
      } else if (id.includes('fan_max')) {
        uniqueActions.push('Turn off fan (safety limit)');
      } else if (id.includes('ha_start')) {
        uniqueActions.push('Check and normalize device states');
      } else if (id.includes('light_manual_off')) {
        uniqueActions.push('Set manual override timer');
      }
    }

    return uniqueActions.slice(0, 5);
  }

  // Utility functions
  function formatEntityId(entityId) {
    if (!entityId) return 'entity';
    // Handle custom HA type objects from YAML parsing
    if (typeof entityId === 'object' && entityId.__ha_type) {
      const inputName = entityId.value || 'entity';
      // Format input name to be more readable
      return formatEntityName(inputName);
    }
    if (typeof entityId === 'string') {
      // Handle !input references
      if (entityId.includes('!input')) {
        const inputName = entityId.replace('!input ', '');
        return formatEntityName(inputName);
      }
      // Get the entity name (part after the domain)
      const name = entityId.split('.').pop() || entityId;
      return formatEntityName(name);
    }
    return 'entity';
  }

  /**
   * Format an entity name to be human-readable
   * Converts underscores to spaces and capitalizes words
   */
  function formatEntityName(name) {
    if (!name) return 'entity';
    return name
      .replace(/_/g, ' ')
      .replace(/\b\w/g, (c) => c.toUpperCase())
      .trim();
  }

  function formatDuration(duration) {
    if (!duration) return '';
    if (typeof duration === 'string') return duration;
    if (typeof duration === 'object') {
      // Handle custom HA type objects from YAML parsing
      if (duration.__ha_type) {
        return `input ${duration.value || 'duration'}`;
      }
      const parts = [];
      if (duration.hours) parts.push(`${duration.hours}h`);
      if (duration.minutes) parts.push(`${duration.minutes}m`);
      if (duration.seconds) parts.push(`${duration.seconds}s`);
      if (duration.milliseconds) parts.push(`${duration.milliseconds}ms`);
      return parts.join(' ') || 'duration';
    }
    return String(duration);
  }

  function truncateTemplate(template) {
    if (!template) return '';
    if (typeof template !== 'string') return 'template';

    // Clean up template
    const clean = template.replace(/\{\{/g, '').replace(/\}\}/g, '').replace(/\s+/g, ' ').trim();

    // Look for meaningful patterns
    const triggerIdMatch = clean.match(/trigger\.id\s*[=!]=?\s*['"]([^'"]+)['"]/);
    if (triggerIdMatch) {
      return `trigger.id == '${triggerIdMatch[1]}'`;
    }

    // Truncate if too long
    return clean.length > 40 ? clean.substring(0, 37) + '...' : clean;
  }

  function escapeLabel(text) {
    if (!text) return '';
    // For Mermaid, we need to escape characters that have special meaning
    // Using Unicode escapes in format #decimal;
    return String(text)
      .substring(0, 45) // Truncate first to avoid issues
      .replace(/\n/g, ' ')
      .replace(/\r/g, '')
      .replace(/"/g, "'") // Replace double quotes with single
      .replace(/[<>]/g, '') // Remove angle brackets
      .replace(/[[\]]/g, '') // Remove square brackets
      .replace(/[{}]/g, '') // Remove curly braces
      .replace(/[()]/g, '') // Remove parentheses
      .replace(/\|/g, ' ') // Replace pipes with space
      .replace(/&/g, 'and') // Replace ampersand
      .replace(/;/g, ',') // Replace semicolons
      .replace(/#/g, '') // Remove hash
      .replace(/`/g, "'") // Replace backticks
      .replace(/\\/g, '') // Remove backslashes
      .trim();
  }

  function escapeHtml(text) {
    if (!text) return '';
    return String(text)
      .replace(/&/g, '&amp;')
      .replace(/</g, '&lt;')
      .replace(/>/g, '&gt;')
      .replace(/"/g, '&quot;')
      .replace(/'/g, '&#039;');
  }

  // Check for URL parameter to auto-load blueprint
  function checkUrlParams() {
    const params = new URLSearchParams(window.location.search);
    const blueprintId = params.get('blueprint');

    if (blueprintId && BLUEPRINT_URLS[blueprintId]) {
      // Set the select value
      if (blueprintSelect) {
        blueprintSelect.value = blueprintId;
        // Trigger the change event
        blueprintSelect.dispatchEvent(new Event('change'));
      }
    }
  }

  // Initialize
  showPlaceholder();
  checkUrlParams();

  // Expose global function for programmatic blueprint selection
  window.selectBlueprint = function (blueprintId) {
    if (blueprintId && BLUEPRINT_URLS[blueprintId] && blueprintSelect) {
      blueprintSelect.value = blueprintId;
      blueprintSelect.dispatchEvent(new Event('change'));
    }
  };
})();
