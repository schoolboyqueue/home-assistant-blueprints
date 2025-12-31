# Home Assistant WebSocket API Reference

## Overview

The WebSocket API provides real-time bidirectional communication with Home Assistant. It supports streaming, subscriptions, and all the commands needed for automation debugging and entity management.

**Endpoint:** `ws://supervisor/core/api/websocket` (from add-on environment)

## Authentication Flow

```
1. Client connects → Server sends auth_required
2. Client sends auth with access_token
3. Server sends auth_ok or auth_invalid
4. Command phase begins (all messages need unique id)
```

### Authentication Messages

```json
// Server sends first:
{"type": "auth_required", "ha_version": "2025.12.5"}

// Client responds:
{"type": "auth", "access_token": "YOUR_TOKEN"}

// Server confirms:
{"type": "auth_ok", "ha_version": "2025.12.5"}
// or
{"type": "auth_invalid", "message": "Invalid access token"}
```

## Message Format

All post-auth messages require:
- `id`: Unique integer for request-response correlation
- `type`: The command type

```json
{"id": 1, "type": "get_states"}
```

### Response Format

```json
// Success
{"id": 1, "type": "result", "success": true, "result": {...}}

// Error
{"id": 1, "type": "result", "success": false, "error": {"code": "...", "message": "..."}}
```

---

## Available Commands

### Data Retrieval

| Command | Description | Parameters |
|---------|-------------|------------|
| `get_states` | Get all entity states | None |
| `get_config` | Get HA configuration | None |
| `get_services` | List all services | None |
| `get_panels` | Get UI panels | None |

### Service Calls

```json
{
  "id": 5,
  "type": "call_service",
  "domain": "light",
  "service": "turn_on",
  "target": {"entity_id": "light.living_room"},
  "service_data": {"brightness": 255}
}
```

Optional: Add `"return_response": true` for services that return data.

### Event Operations

**Fire an event:**
```json
{
  "id": 6,
  "type": "fire_event",
  "event_type": "my_custom_event",
  "event_data": {"key": "value"}
}
```

### Subscriptions

**Subscribe to all events:**
```json
{"id": 10, "type": "subscribe_events"}
```

**Subscribe to specific event type:**
```json
{"id": 11, "type": "subscribe_events", "event_type": "state_changed"}
```

**Subscribe to trigger:**
```json
{
  "id": 12,
  "type": "subscribe_trigger",
  "trigger": {
    "platform": "state",
    "entity_id": "binary_sensor.motion",
    "to": "on"
  }
}
```

**Unsubscribe:**
```json
{"id": 13, "type": "unsubscribe_events", "subscription": 10}
```

### Advanced Commands

| Command | Description | Parameters |
|---------|-------------|------------|
| `ping` | Heartbeat check | None (returns `pong`) |
| `validate_config` | Validate automation config | `trigger`, `condition`, `action` |
| `render_template` | Render Jinja template | `template`, `variables` (optional) |
| `extract_from_target` | Resolve target entities | `target`, `expand_group` (optional) |

### Render Template Example

```json
{
  "id": 20,
  "type": "render_template",
  "template": "{{ states('sun.sun') }} - {{ states.light | selectattr('state', 'eq', 'on') | list | count }} lights on"
}
```

---

## Log & History Commands (Undocumented)

These commands work but aren't in the official docs:

### Logbook Events

```json
{
  "id": 30,
  "type": "logbook/get_events",
  "start_time": "2025-01-01T00:00:00Z",
  "end_time": "2025-01-01T01:00:00Z",
  "entity_ids": ["binary_sensor.motion"]
}
```

### State History

```json
{
  "id": 31,
  "type": "history/history_during_period",
  "start_time": "2025-01-01T00:00:00Z",
  "end_time": "2025-01-01T01:00:00Z",
  "entity_ids": ["sensor.temperature"],
  "minimal_response": true,
  "significant_changes_only": true
}
```

### System Log

```json
{"id": 32, "type": "system_log/list"}
```

Returns array of error/warning entries with `level`, `source`, `message`, `timestamp`.

### Recorder Statistics

```json
{
  "id": 33,
  "type": "recorder/statistics_during_period",
  "start_time": "2025-01-01T00:00:00Z",
  "statistic_ids": ["sensor.temperature"],
  "period": "hour"
}
```

Returns hourly/daily statistics with `min`, `max`, `mean` values.

---

## Registry Commands (Undocumented)

### Entity Registry

```json
{"id": 40, "type": "config/entity_registry/list"}
```

Returns all entities with metadata: `entity_id`, `name`, `platform`, `device_id`, `area_id`, `disabled_by`.

### Device Registry

```json
{"id": 41, "type": "config/device_registry/list"}
```

Returns devices with: `id`, `name`, `manufacturer`, `model`, `area_id`, `identifiers`.

### Area Registry

```json
{"id": 42, "type": "config/area_registry/list"}
```

Returns areas with: `area_id`, `name`, `aliases`.

---

## Automation Debugging Commands

### List Traces

```json
{"id": 50, "type": "trace/list", "domain": "automation"}
```

Returns recent automation runs with: `item_id`, `run_id`, `state`, `timestamp`, `script_execution`.

### Get Trace Details

```json
{"id": 51, "type": "trace/get", "domain": "automation", "item_id": "my_automation_id", "run_id": "abc123..."}
```

Returns detailed trace for a specific run including: `script_execution`, `error`, `trace` (step-by-step with variables and errors), `config`, `context`.

### Get Automation Config

```json
{"id": 51, "type": "automation/config", "entity_id": "automation.my_automation"}
```

Returns the full YAML configuration for an automation.

---

## Event Types for Subscriptions

| Event Type | Description |
|------------|-------------|
| `state_changed` | Entity state changes |
| `call_service` | Service calls |
| `automation_triggered` | Automation triggers |
| `script_started` | Script executions |
| `homeassistant_start` | HA startup |
| `homeassistant_stop` | HA shutdown |

---

## Using from CLI (TypeScript)

A helper script is available at `.claude/ha-ws-client.ts`:

```bash
# Two ways to run:
npx tsx .claude/ha-ws-client.ts <command>    # Direct
npm run ha -- <command>                       # Via npm script (from .claude dir)
```

### Basic Commands

```bash
# Entity states
npx tsx .claude/ha-ws-client.ts state sun.sun                    # Single entity
npx tsx .claude/ha-ws-client.ts states                           # Summary of all entities
npx tsx .claude/ha-ws-client.ts states-json                      # All states as JSON
npx tsx .claude/ha-ws-client.ts states-filter "light.*"          # Filter by pattern

# Service calls
npx tsx .claude/ha-ws-client.ts call light turn_on '{"entity_id":"light.kitchen"}'
npx tsx .claude/ha-ws-client.ts call automation reload

# System info
npx tsx .claude/ha-ws-client.ts config                           # HA configuration
npx tsx .claude/ha-ws-client.ts services                         # List all services
npx tsx .claude/ha-ws-client.ts ping                             # Test connection

# Templates
npx tsx .claude/ha-ws-client.ts template "{{ states('sun.sun') }}"
echo "{{ now() }}" | npx tsx .claude/ha-ws-client.ts template -   # Via stdin
```

### Log & History Commands

```bash
# Logbook and history
npx tsx .claude/ha-ws-client.ts logbook binary_sensor.motion 2   # Last 2 hours
npx tsx .claude/ha-ws-client.ts history sensor.temperature 4     # Last 4 hours
npx tsx .claude/ha-ws-client.ts history-full climate.thermostat 4  # With attributes
npx tsx .claude/ha-ws-client.ts attrs climate.thermostat 4       # Compact attribute changes
npx tsx .claude/ha-ws-client.ts timeline 2 climate.thermostat binary_sensor.door  # Multi-entity

# System logs and stats
npx tsx .claude/ha-ws-client.ts syslog                           # System errors/warnings
npx tsx .claude/ha-ws-client.ts stats sensor.temperature 24      # Hourly statistics

# Context lookup (find what triggered a state change)
npx tsx .claude/ha-ws-client.ts context 01KDQS4E2WHMYJYYXKC7K28XFG

# Live monitoring
npx tsx .claude/ha-ws-client.ts watch light.kitchen 30           # Watch state changes for 30s
```

### Time Filtering Options

For `logbook`, `history`, `history-full`, `attrs`, and `timeline` commands:

```bash
# Use --from and --to instead of hours
npx tsx .claude/ha-ws-client.ts history sensor.temp --from "2025-12-29 06:00" --to "2025-12-29 12:00"
npx tsx .claude/ha-ws-client.ts logbook binary_sensor.door --from "2025-12-30 07:00"
```

### Registry Commands

```bash
npx tsx .claude/ha-ws-client.ts entities                         # List all entities
npx tsx .claude/ha-ws-client.ts entities "occupancy"             # Search entities
npx tsx .claude/ha-ws-client.ts devices                          # List all devices
npx tsx .claude/ha-ws-client.ts devices "bedroom"                # Search devices
npx tsx .claude/ha-ws-client.ts areas                            # List all areas
```

### Automation Debugging

```bash
# List traces (accepts entity_id or numeric item_id)
npx tsx .claude/ha-ws-client.ts traces                           # All automation traces
npx tsx .claude/ha-ws-client.ts traces automation.my_automation  # Filter by automation

# Get detailed trace
npx tsx .claude/ha-ws-client.ts trace <run_id> [automation_id]   # Detailed trace for a run

# Show evaluated variables from a trace
npx tsx .claude/ha-ws-client.ts trace-vars <run_id>              # Variables at each step

# Get automation config
npx tsx .claude/ha-ws-client.ts automation-config automation.my_automation

# Validate blueprint inputs vs expected (catches misnamed inputs!)
npx tsx .claude/ha-ws-client.ts blueprint-inputs automation.bathroom_lights
```

### Advanced Trace Debugging

```bash
# Step-by-step execution timeline with timestamps and durations
npx tsx .claude/ha-ws-client.ts trace-timeline <run_id>          # Chronological trace steps

# Detailed trigger context showing what triggered the automation
npx tsx .claude/ha-ws-client.ts trace-trigger <run_id>           # Trigger entity, state changes

# Action results showing service call params and responses
npx tsx .claude/ha-ws-client.ts trace-actions <run_id>           # Action params, results, errors

# Comprehensive debug view combining all information
npx tsx .claude/ha-ws-client.ts trace-debug <run_id>             # Full debug trace
```

**Example trace-debug output:**
```
============================================================
AUTOMATION DEBUG TRACE
============================================================

Automation: automation.bathroom_lights
Run ID: 01KDQS4E2WHMYJYYXKC7K28XFG
Status: finished
Started: 1/15/2025, 10:30:45 AM
Duration: 234ms

------------------------------------------------------------
TRIGGER CONTEXT
------------------------------------------------------------
Platform: state
Entity: binary_sensor.bathroom_motion
From: off
To: on

------------------------------------------------------------
EVALUATED VARIABLES
------------------------------------------------------------
[boolean] is_occupied: true
[number] light_level: 45
[string] time_of_day: morning

------------------------------------------------------------
EXECUTION TIMELINE
------------------------------------------------------------

[ok] [10:30:45] Trigger 0
[ok] [10:30:45] Condition 0
    humidity_above_threshold: true
[ok] [10:30:45] Action 0
    domain: light
    service: turn_on
    entity_id: light.bathroom

------------------------------------------------------------
CONTEXT
------------------------------------------------------------
Context ID: 01KDQS4E2WHMYJYYXKC7K28XFG

============================================================
END OF DEBUG TRACE
============================================================
```

### Live Monitoring

```bash
# Basic watch (simpler output)
npx tsx .claude/ha-ws-client.ts watch binary_sensor.motion 30    # Watch for 30 seconds
npx tsx .claude/ha-ws-client.ts watch light.kitchen 60           # Shows brightness/temp changes

# Advanced monitoring with anomaly detection
npx tsx .claude/ha-ws-client.ts monitor sensor.temperature 300   # 5 minutes with full analysis
npx tsx .claude/ha-ws-client.ts monitor light.kitchen 60         # Track state + attributes

# Monitor multiple entities simultaneously
npx tsx .claude/ha-ws-client.ts monitor-multi 120 sensor.temp1 sensor.temp2 light.living_room

# Analyze historical data for anomalies
npx tsx .claude/ha-ws-client.ts analyze sensor.temperature 24    # Last 24 hours
npx tsx .claude/ha-ws-client.ts analyze binary_sensor.motion 4   # Detect oscillation patterns
```

**Monitor features:**
- Live state change tracking with timestamps
- Rate-of-change detection for numeric sensors
- Anomaly detection and highlighting:
  - Values outside normal range (3+ std deviations)
  - Rapid rate of change
  - Oscillation/flapping detection
  - Unavailability events
- Summary statistics at end of monitoring
- Attribute change tracking

**Example monitor output:**
```
Monitoring sensor.temperature for 300 seconds...

Initial state: 21.5
  Attributes: temperature=21.5, humidity=45

─── Live State Changes ───

[10:30:45] (initial) → 21.6 (+0.033/s)
[10:31:15] 21.6 → 21.8 (+0.007/s)
[10:32:00] 21.8 → 25.2 (+0.076/s)
         ⚠ Rapid change: +0.076/s; Value 25.20 is 3.2 std devs from mean

═══ Monitoring Summary ═══
Entity: sensor.temperature
Duration: 5m 0.0s
State changes: 3

Numeric Statistics:
  Min: 21.50
  Max: 25.20
  Mean: 22.77
  Std Dev: 1.72

Anomalies Detected: 1
  10:32:00: 25.2 - Value 25.20 is 3.2 std devs from mean
```

### Attribute History

```bash
# Compact attribute change history (shows only changed attributes)
npx tsx .claude/ha-ws-client.ts attrs climate.thermostat 12      # Last 12 hours
npx tsx .claude/ha-ws-client.ts attrs light.kitchen 4            # Brightness/color_temp changes
npx tsx .claude/ha-ws-client.ts attrs cover.blinds 24            # Position changes
```

---

## Sources

- [WebSocket API Developer Docs](https://developers.home-assistant.io/docs/api/websocket/)
- [WebSocket API Integration](https://www.home-assistant.io/integrations/websocket_api/)
- [Official JS Library](https://github.com/home-assistant/home-assistant-js-websocket)
