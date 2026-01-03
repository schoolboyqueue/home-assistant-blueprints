// Package testfixtures provides shared test fixtures and factory functions for
// Home Assistant related Go projects. It consolidates common testing patterns
// from ha-ws-client-go and validate-blueprint-go into a reusable library.
package testfixtures

// Map is a shorthand for map[string]interface{} used in YAML/JSON structures.
type Map = map[string]interface{}

// List is a shorthand for []interface{} used in YAML/JSON structures.
type List = []interface{}

// =====================================
// Home Assistant WebSocket Types
// These mirror the types from ha-ws-client-go for test fixtures
// =====================================

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
