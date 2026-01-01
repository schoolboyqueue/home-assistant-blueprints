# Home Assistant WebSocket API Reference

This document provides the correct WebSocket message types, parameters, and response formats for the Home Assistant WebSocket API. Use this as context when developing WebSocket clients.

## Connection

- **URL**: `ws://supervisor/core/api/websocket` (from add-on) or `ws://<host>:8123/api/websocket`
- **Auth Flow**:
  1. Receive `{"type": "auth_required"}`
  2. Send `{"type": "auth", "access_token": "<token>"}`
  3. Receive `{"type": "auth_ok"}` or `{"type": "auth_invalid"}`

## Message Format

All messages after auth use this format:
```json
{
  "id": 1,
  "type": "<message_type>",
  ...parameters
}
```

Responses:
```json
{
  "id": 1,
  "type": "result",
  "success": true,
  "result": <data>
}
```

---

## Core Commands

### ping
Test connection.
```json
{"id": 1, "type": "ping"}
```
Response: `{"id": 1, "type": "pong"}`

### get_states
Get all entity states.
```json
{"id": 1, "type": "get_states"}
```
Response: Array of state objects
```json
{
  "result": [
    {
      "entity_id": "light.kitchen",
      "state": "on",
      "attributes": {...},
      "last_changed": "2024-01-01T00:00:00+00:00",
      "last_updated": "2024-01-01T00:00:00+00:00",
      "context": {"id": "...", "parent_id": null, "user_id": null}
    }
  ]
}
```

### get_config
Get Home Assistant configuration.
```json
{"id": 1, "type": "get_config"}
```

### get_services
Get all available services.
```json
{"id": 1, "type": "get_services"}
```
Response: `{"result": {"domain": {"service_name": {...}}}}`

### call_service
Call a service.
```json
{
  "id": 1,
  "type": "call_service",
  "domain": "light",
  "service": "turn_on",
  "service_data": {"entity_id": "light.kitchen", "brightness": 255}
}
```

### render_template (Subscription-based)
Render a Jinja2 template. **This is subscription-based** - it returns results via events.
```json
{
  "id": 1,
  "type": "render_template",
  "template": "{{ states('sun.sun') }}"
}
```
Initial response confirms subscription, then events with:
```json
{
  "id": 1,
  "type": "event",
  "event": {"result": "<rendered_template>"}
}
```

---

## History Commands

### history/history_during_period
Get state history for entities. **NOT** `history/period/<timestamp>`.
```json
{
  "id": 1,
  "type": "history/history_during_period",
  "start_time": "2024-01-01T00:00:00+00:00",
  "end_time": "2024-01-02T00:00:00+00:00",
  "entity_ids": ["sensor.temperature", "light.kitchen"],
  "minimal_response": true,
  "no_attributes": true,
  "significant_changes_only": false,
  "include_start_time_state": true
}
```
**Response format**: `map[entity_id][]HistoryState`
```json
{
  "result": {
    "sensor.temperature": [
      {"s": "72.5", "lu": 1704067200},
      {"s": "73.0", "lu": 1704070800}
    ],
    "light.kitchen": [
      {"s": "on", "lu": 1704067200}
    ]
  }
}
```

**HistoryState fields** (minimal_response=true):
- `s` - state value (string)
- `lu` - last_updated (Unix timestamp, number)
- `lc` - last_changed (Unix timestamp, number)
- `a` - attributes (object, if no_attributes=false)

**HistoryState fields** (minimal_response=false):
- `state` - state value
- `last_updated` - ISO timestamp string
- `last_changed` - ISO timestamp string
- `attributes` - full attributes object

### history/stream (Subscription-based)
Subscribe to live history stream.
```json
{
  "id": 1,
  "type": "history/stream",
  "start_time": "2024-01-01T00:00:00+00:00",
  "entity_ids": ["sensor.temperature"]
}
```

---

## Logbook Commands

### logbook/get_events
Get logbook entries. **NOT** `logbook/period/<timestamp>`.
```json
{
  "id": 1,
  "type": "logbook/get_events",
  "start_time": "2024-01-01T00:00:00+00:00",
  "end_time": "2024-01-02T00:00:00+00:00",
  "entity_ids": ["light.kitchen"],
  "device_ids": [],
  "context_id": null
}
```
Response: Array of logbook entries
```json
{
  "result": [
    {
      "when": 1704067200.123,
      "state": "on",
      "entity_id": "light.kitchen",
      "message": "turned on"
    }
  ]
}
```

### logbook/event_stream (Subscription-based)
Subscribe to live logbook events.

---

## Recorder/Statistics Commands

### recorder/statistics_during_period
Get sensor statistics.
```json
{
  "id": 1,
  "type": "recorder/statistics_during_period",
  "start_time": "2024-01-01T00:00:00+00:00",
  "end_time": "2024-01-02T00:00:00+00:00",
  "statistic_ids": ["sensor:temperature"],
  "period": "hour"
}
```
**period options**: `5minute`, `hour`, `day`, `week`, `month`

Response: `map[statistic_id][]StatEntry`
```json
{
  "result": {
    "sensor:temperature": [
      {
        "start": 1704067200,
        "end": 1704070800,
        "min": 70.5,
        "max": 74.2,
        "mean": 72.3,
        "sum": null,
        "state": null
      }
    ]
  }
}
```

**Note**: `start` and `end` are Unix timestamps (numbers), not ISO strings.

### recorder/list_statistic_ids
List available statistic IDs.
```json
{
  "id": 1,
  "type": "recorder/list_statistic_ids",
  "statistic_type": "mean"
}
```

---

## System Log Commands

### system_log/list
Get system log entries.
```json
{"id": 1, "type": "system_log/list"}
```
Response:
```json
{
  "result": [
    {
      "level": "ERROR",
      "source": ["homeassistant.components.sensor", "123"],
      "message": ["Error message here"],
      "name": "homeassistant.components.sensor",
      "timestamp": 1704067200.123,
      "first_occurred": 1704067000.000,
      "count": 5
    }
  ]
}
```

**Note**: `message` and `source` can be arrays or strings depending on the log entry.

---

## Registry Commands

### config/entity_registry/list
List all entities in the registry.
```json
{"id": 1, "type": "config/entity_registry/list"}
```

### config/device_registry/list
List all devices.
```json
{"id": 1, "type": "config/device_registry/list"}
```

### config/area_registry/list
List all areas.
```json
{"id": 1, "type": "config/area_registry/list"}
```

---

## Automation Commands

### automation/config
Get automation configuration. **NOT** `config/automation/config/<id>`.
```json
{
  "id": 1,
  "type": "automation/config",
  "entity_id": "automation.my_automation"
}
```
Response:
```json
{
  "result": {
    "config": {
      "id": "1234567890",
      "alias": "My Automation",
      "trigger": [...],
      "condition": [...],
      "action": [...],
      "use_blueprint": {
        "path": "author/blueprint.yaml",
        "input": {...}
      }
    }
  }
}
```

---

## Trace Commands

### trace/list
List automation/script traces. Response is an **array**, not a map.
```json
{
  "id": 1,
  "type": "trace/list",
  "domain": "automation",
  "item_id": "my_automation"
}
```
Response:
```json
{
  "result": [
    {
      "item_id": "my_automation",
      "run_id": "01ABC123...",
      "state": "stopped",
      "script_execution": "finished",
      "timestamp": {"start": "2024-01-01T00:00:00+00:00", "finish": "2024-01-01T00:00:01+00:00"},
      "context": {"id": "..."}
    }
  ]
}
```

### trace/get
Get detailed trace for a specific run.
```json
{
  "id": 1,
  "type": "trace/get",
  "domain": "automation",
  "item_id": "my_automation",
  "run_id": "01ABC123..."
}
```

### trace/contexts
Get trace contexts.
```json
{
  "id": 1,
  "type": "trace/contexts",
  "domain": "automation",
  "item_id": "my_automation"
}
```

---

## Subscription Commands

### subscribe_trigger
Subscribe to a trigger.
```json
{
  "id": 1,
  "type": "subscribe_trigger",
  "trigger": {
    "platform": "state",
    "entity_id": "light.kitchen"
  }
}
```
Events:
```json
{
  "id": 1,
  "type": "event",
  "event": {
    "variables": {
      "trigger": {
        "platform": "state",
        "entity_id": "light.kitchen",
        "from_state": {...},
        "to_state": {...}
      }
    }
  }
}
```

### subscribe_events
Subscribe to events by type.
```json
{
  "id": 1,
  "type": "subscribe_events",
  "event_type": "state_changed"
}
```

### unsubscribe_events
Unsubscribe from events.
```json
{
  "id": 1,
  "type": "unsubscribe_events",
  "subscription": 1
}
```

---

## Common Mistakes to Avoid

1. **History endpoint**: Use `history/history_during_period`, NOT `history/period/<timestamp>`

2. **Logbook endpoint**: Use `logbook/get_events`, NOT `logbook/period/<timestamp>`

3. **Automation config**: Use `automation/config` with `entity_id` parameter, NOT `config/automation/config/<id>`

4. **Trace list response**: Returns `[]TraceInfo` array, NOT `map[string][]TraceInfo`

5. **History response**: Returns `map[entity_id][]HistoryState`, NOT `[][]HistoryState`

6. **Statistics timestamps**: `start` field is a Unix timestamp (number), NOT an ISO string

7. **System log message**: Can be a string OR an array, handle both cases

8. **Template rendering**: Is subscription-based, result comes via event, not direct response

9. **Entity ID parameters**: Some commands want `entity_id` (string), others want `entity_ids` (array)

10. **Minimal response format**: Uses short field names (`s`, `lu`, `lc`, `a`) vs full names

---

## Type Definitions

### HAState
```typescript
interface HAState {
  entity_id: string;
  state: string;
  attributes: Record<string, any>;
  last_changed: string;  // ISO timestamp
  last_updated: string;  // ISO timestamp
  context: {
    id: string;
    parent_id: string | null;
    user_id: string | null;
  };
}
```

### HistoryState (minimal_response=true)
```typescript
interface HistoryStateMinimal {
  s: string;           // state
  lu: number;          // last_updated (Unix timestamp)
  lc?: number;         // last_changed (Unix timestamp)
  a?: Record<string, any>;  // attributes
}
```

### HistoryState (minimal_response=false)
```typescript
interface HistoryStateFull {
  state: string;
  last_updated: string;  // ISO timestamp
  last_changed: string;  // ISO timestamp
  attributes: Record<string, any>;
}
```

### LogbookEntry
```typescript
interface LogbookEntry {
  when: number;         // Unix timestamp with fractional seconds
  state?: string;
  entity_id?: string;
  message?: string;
  context_id?: string;
}
```

### StatEntry
```typescript
interface StatEntry {
  start: number;        // Unix timestamp (NOT string)
  end?: number;         // Unix timestamp
  min: number;
  max: number;
  mean?: number;
  sum?: number;
  state?: number;
}
```

### SysLogEntry
```typescript
interface SysLogEntry {
  level: string;
  source: string[];      // Array of strings
  message: string | string[];  // Can be string OR array
  name?: string;
  timestamp?: number;
  first_occurred?: number;
  count?: number;
}
```

### TraceInfo
```typescript
interface TraceInfo {
  item_id: string;
  run_id: string;
  state?: string;
  script_execution?: string;
  timestamp?: {
    start: string;
    finish?: string;
  };
  context?: {
    id: string;
    parent_id?: string;
    user_id?: string;
  };
}
```
