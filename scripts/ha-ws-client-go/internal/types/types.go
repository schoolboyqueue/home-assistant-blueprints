// Package types provides type definitions for the Home Assistant WebSocket API client.
package types

import "time"

// HAMessage represents a message from the Home Assistant WebSocket API.
type HAMessage struct {
	ID      int            `json:"id,omitempty"`
	Type    string         `json:"type"`
	Success *bool          `json:"success,omitempty"`
	Result  any            `json:"result,omitempty"`
	Error   *HAError       `json:"error,omitempty"`
	Message string         `json:"message,omitempty"`
	Event   *HAEvent       `json:"event,omitempty"`
	Data    map[string]any `json:"data,omitempty"`
}

// HAError represents an error response from Home Assistant.
type HAError struct {
	Code    string `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}

// HAEvent represents an event from a subscription.
type HAEvent struct {
	Result    any            `json:"result,omitempty"`
	Variables map[string]any `json:"variables,omitempty"`
}

// HAState represents a Home Assistant entity state.
type HAState struct {
	EntityID    string         `json:"entity_id"`
	State       string         `json:"state"`
	Attributes  map[string]any `json:"attributes,omitempty"`
	LastChanged string         `json:"last_changed,omitempty"`
	LastUpdated string         `json:"last_updated,omitempty"`
	Context     *HAContext     `json:"context,omitempty"`
}

// HAContext represents context information for tracking causation.
type HAContext struct {
	ID       string `json:"id"`
	ParentID string `json:"parent_id,omitempty"`
	UserID   string `json:"user_id,omitempty"`
}

// HAConfig represents Home Assistant configuration.
type HAConfig struct {
	Version      string            `json:"version"`
	LocationName string            `json:"location_name"`
	TimeZone     string            `json:"time_zone"`
	UnitSystem   map[string]string `json:"unit_system"`
	State        string            `json:"state"`
	Components   []string          `json:"components"`
}

// LogbookEntry represents a logbook entry.
type LogbookEntry struct {
	When      float64 `json:"when"`
	EntityID  string  `json:"entity_id,omitempty"`
	State     string  `json:"state,omitempty"`
	Message   string  `json:"message,omitempty"`
	ContextID string  `json:"context_id,omitempty"`
}

// HistoryState represents a historical state entry.
type HistoryState struct {
	// Compact format fields
	LU int64          `json:"lu,omitempty"` // Last updated (Unix timestamp)
	LC int64          `json:"lc,omitempty"` // Last changed (Unix timestamp)
	S  string         `json:"s,omitempty"`  // State
	A  map[string]any `json:"a,omitempty"`  // Attributes

	// Full format fields
	LastUpdated string         `json:"last_updated,omitempty"`
	LastChanged string         `json:"last_changed,omitempty"`
	State       string         `json:"state,omitempty"`
	Attributes  map[string]any `json:"attributes,omitempty"`
}

// GetState returns the state value (handles both compact and full format).
func (h *HistoryState) GetState() string {
	if h.S != "" {
		return h.S
	}
	return h.State
}

// GetLastUpdated returns the last updated time.
func (h *HistoryState) GetLastUpdated() time.Time {
	if h.LU > 0 {
		return time.Unix(h.LU, 0)
	}
	if h.LastUpdated != "" {
		t, _ := time.Parse(time.RFC3339, h.LastUpdated)
		return t
	}
	return time.Time{}
}

// TraceInfo represents summary information about an automation trace.
type TraceInfo struct {
	ItemID          string     `json:"item_id"`
	RunID           string     `json:"run_id"`
	State           string     `json:"state,omitempty"`
	ScriptExecution string     `json:"script_execution,omitempty"`
	Timestamp       *Timestamp `json:"timestamp,omitempty"`
	Context         *HAContext `json:"context,omitempty"`
}

// Timestamp represents start/finish times.
type Timestamp struct {
	Start  string `json:"start"`
	Finish string `json:"finish,omitempty"`
}

// TraceDetail represents detailed trace information for an automation run.
type TraceDetail struct {
	ScriptExecution string                 `json:"script_execution,omitempty"`
	Error           string                 `json:"error,omitempty"`
	Trace           map[string][]TraceStep `json:"trace,omitempty"`
	Config          *AutomationConfig      `json:"config,omitempty"`
	Context         *HAContext             `json:"context,omitempty"`
	Trigger         *TraceTrigger          `json:"trigger,omitempty"`
	RunID           string                 `json:"run_id,omitempty"`
	Domain          string                 `json:"domain,omitempty"`
	ItemID          string                 `json:"item_id,omitempty"`
	Timestamp       *Timestamp             `json:"timestamp,omitempty"`
}

// TraceStep represents a single step in an automation trace.
type TraceStep struct {
	Path             string         `json:"path,omitempty"`
	Error            any            `json:"error,omitempty"`
	Result           *TraceResult   `json:"result,omitempty"`
	Variables        map[string]any `json:"variables,omitempty"`
	ChangedVariables map[string]any `json:"changed_variables,omitempty"`
	Timestamp        string         `json:"timestamp,omitempty"`
}

// TraceResult represents the result of a trace step.
type TraceResult struct {
	Error         any            `json:"error,omitempty"`
	Response      any            `json:"response,omitempty"`
	Params        map[string]any `json:"params,omitempty"`
	RunningScript bool           `json:"running_script,omitempty"`
	Limit         int            `json:"limit,omitempty"`
	Enabled       bool           `json:"enabled,omitempty"`
}

// TraceTrigger represents trigger information from a trace.
type TraceTrigger struct {
	ID          string        `json:"id,omitempty"`
	Idx         string        `json:"idx,omitempty"`
	Alias       string        `json:"alias,omitempty"`
	Platform    string        `json:"platform,omitempty"`
	EntityID    string        `json:"entity_id,omitempty"`
	FromState   *TriggerState `json:"from_state,omitempty"`
	ToState     *TriggerState `json:"to_state,omitempty"`
	For         any           `json:"for,omitempty"`
	Description string        `json:"description,omitempty"`
}

// TriggerState represents state in a trigger.
type TriggerState struct {
	EntityID    string         `json:"entity_id,omitempty"`
	State       string         `json:"state"`
	Attributes  map[string]any `json:"attributes,omitempty"`
	LastChanged string         `json:"last_changed,omitempty"`
	LastUpdated string         `json:"last_updated,omitempty"`
}

// AutomationConfig represents automation configuration.
type AutomationConfig struct {
	ID           string        `json:"id,omitempty"`
	Alias        string        `json:"alias,omitempty"`
	UseBlueprint *BlueprintRef `json:"use_blueprint,omitempty"`
	Trigger      []any         `json:"trigger,omitempty"`
	Condition    []any         `json:"condition,omitempty"`
	Action       []any         `json:"action,omitempty"`
}

// BlueprintRef represents a reference to a blueprint.
type BlueprintRef struct {
	Path  string         `json:"path"`
	Input map[string]any `json:"input,omitempty"`
}

// EntityEntry represents an entity registry entry.
type EntityEntry struct {
	EntityID     string `json:"entity_id"`
	Name         string `json:"name,omitempty"`
	OriginalName string `json:"original_name,omitempty"`
	Platform     string `json:"platform,omitempty"`
	DisabledBy   string `json:"disabled_by,omitempty"`
}

// DeviceEntry represents a device registry entry.
type DeviceEntry struct {
	ID           string `json:"id"`
	Name         string `json:"name,omitempty"`
	NameByUser   string `json:"name_by_user,omitempty"`
	Manufacturer string `json:"manufacturer,omitempty"`
	Model        string `json:"model,omitempty"`
	AreaID       string `json:"area_id,omitempty"`
}

// AreaEntry represents an area registry entry.
type AreaEntry struct {
	AreaID  string   `json:"area_id"`
	Name    string   `json:"name"`
	Aliases []string `json:"aliases,omitempty"`
}

// SysLogEntry represents a system log entry.
type SysLogEntry struct {
	Level         string   `json:"level,omitempty"`
	Source        []string `json:"source,omitempty"`
	Message       any      `json:"message,omitempty"` // Can be string or []string
	Name          string   `json:"name,omitempty"`
	Timestamp     float64  `json:"timestamp,omitempty"`
	FirstOccurred float64  `json:"first_occurred,omitempty"`
	Count         int      `json:"count,omitempty"`
}

// GetMessage returns the message as a string.
func (e *SysLogEntry) GetMessage() string {
	switch m := e.Message.(type) {
	case string:
		return m
	case []any:
		if len(m) > 0 {
			if s, ok := m[0].(string); ok {
				return s
			}
		}
		return ""
	default:
		return ""
	}
}

// StatEntry represents a statistics entry for a sensor.
type StatEntry struct {
	Start any     `json:"start"` // Can be float64 (Unix timestamp) or string (ISO format)
	End   any     `json:"end,omitempty"`
	Min   float64 `json:"min"`
	Max   float64 `json:"max"`
	Mean  float64 `json:"mean,omitempty"`
	Sum   float64 `json:"sum,omitempty"`
	State float64 `json:"state,omitempty"`
}

// GetStartTime returns the start time as a formatted string.
func (s *StatEntry) GetStartTime() string {
	switch v := s.Start.(type) {
	case float64:
		return time.Unix(int64(v), 0).Format(time.RFC3339)
	case string:
		return v
	default:
		return ""
	}
}

// TimeRange represents a time range for history queries.
type TimeRange struct {
	StartTime time.Time
	EndTime   time.Time
}
