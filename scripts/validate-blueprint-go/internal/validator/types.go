// Package validator provides validation for Home Assistant Blueprint YAML files.
// This file contains strongly-typed struct definitions for blueprint data structures,
// replacing the dynamic map[string]interface{} pattern with compile-time type safety.
package validator

// BlueprintData represents the complete parsed YAML structure of a Home Assistant blueprint.
// This is the root type that replaces map[string]interface{} for the main data field.
type BlueprintData struct {
	Blueprint Blueprint   `yaml:"blueprint"`
	Trigger   []Trigger   `yaml:"trigger,omitempty"`
	Condition []Condition `yaml:"condition,omitempty"`
	Action    []Action    `yaml:"action,omitempty"`
	Variables Variables   `yaml:"variables,omitempty"`
	Mode      string      `yaml:"mode,omitempty"`
	Max       int         `yaml:"max,omitempty"`
	Sequence  []Action    `yaml:"sequence,omitempty"` // For script blueprints
	Raw       RawData     `yaml:"-"`                  // Original untyped data for backward compatibility
}

// RawData is an alias for the original untyped map for cases where dynamic access is still needed.
// This provides a migration path while maintaining backward compatibility.
// Using an alias (=) ensures it's treated identically to map[string]interface{} in type comparisons.
type RawData = map[string]interface{}

// Blueprint represents the blueprint metadata section.
type Blueprint struct {
	Name                string              `yaml:"name"`
	Description         string              `yaml:"description"`
	Domain              string              `yaml:"domain"`
	Input               map[string]InputDef `yaml:"input,omitempty"`
	Author              string              `yaml:"author,omitempty"`
	HomeassistantConfig HomeassistantConfig `yaml:"homeassistant,omitempty"`
	SourceURL           string              `yaml:"source_url,omitempty"`
}

// HomeassistantConfig represents Home Assistant version requirements.
type HomeassistantConfig struct {
	MinVersion string `yaml:"min_version,omitempty"`
}

// InputDef represents a single input definition or an input group.
type InputDef struct {
	Name        string              `yaml:"name,omitempty"`
	Description string              `yaml:"description,omitempty"`
	Default     interface{}         `yaml:"default,omitempty"`
	Selector    Selector            `yaml:"selector,omitempty"`
	Input       map[string]InputDef `yaml:"input,omitempty"` // For input groups
	Collapsed   bool                `yaml:"collapsed,omitempty"`
	Icon        string              `yaml:"icon,omitempty"`
}

// Selector represents an input selector configuration.
// Only one selector type should be set at a time.
type Selector struct {
	Action            *ActionSelector            `yaml:"action,omitempty"`
	Addon             *AddonSelector             `yaml:"addon,omitempty"`
	Area              *AreaSelector              `yaml:"area,omitempty"`
	Attribute         *AttributeSelector         `yaml:"attribute,omitempty"`
	Boolean           *BooleanSelector           `yaml:"boolean,omitempty"`
	ColorRGB          *ColorRGBSelector          `yaml:"color_rgb,omitempty"`
	ColorTemp         *ColorTempSelector         `yaml:"color_temp,omitempty"`
	Condition         *ConditionSelector         `yaml:"condition,omitempty"`
	ConversationAgent *ConversationAgentSelector `yaml:"conversation_agent,omitempty"`
	Country           *CountrySelector           `yaml:"country,omitempty"`
	Date              *DateSelector              `yaml:"date,omitempty"`
	Datetime          *DatetimeSelector          `yaml:"datetime,omitempty"`
	Device            *DeviceSelector            `yaml:"device,omitempty"`
	Duration          *DurationSelector          `yaml:"duration,omitempty"`
	Entity            *EntitySelector            `yaml:"entity,omitempty"`
	File              *FileSelector              `yaml:"file,omitempty"`
	Floor             *FloorSelector             `yaml:"floor,omitempty"`
	Icon              *IconSelector              `yaml:"icon,omitempty"`
	Label             *LabelSelector             `yaml:"label,omitempty"`
	Language          *LanguageSelector          `yaml:"language,omitempty"`
	Location          *LocationSelector          `yaml:"location,omitempty"`
	Media             *MediaSelector             `yaml:"media,omitempty"`
	Navigation        *NavigationSelector        `yaml:"navigation,omitempty"`
	Number            *NumberSelector            `yaml:"number,omitempty"`
	Object            *ObjectSelector            `yaml:"object,omitempty"`
	Select            *SelectSelector            `yaml:"select,omitempty"`
	State             *StateSelector             `yaml:"state,omitempty"`
	Target            *TargetSelector            `yaml:"target,omitempty"`
	Template          *TemplateSelector          `yaml:"template,omitempty"`
	Text              *TextSelector              `yaml:"text,omitempty"`
	Theme             *ThemeSelector             `yaml:"theme,omitempty"`
	Time              *TimeSelector              `yaml:"time,omitempty"`
	TriggerSelector   *TriggerSelector           `yaml:"trigger,omitempty"`
	UIAction          *UIActionSelector          `yaml:"ui_action,omitempty"`
	UIColor           *UIColorSelector           `yaml:"ui_color,omitempty"`
}

// GetType returns the selector type name, or empty string if none is set.
//
//nolint:gocyclo // Complexity is inherent to checking all selector types
func (s *Selector) GetType() string {
	switch {
	case s.Action != nil:
		return "action"
	case s.Addon != nil:
		return "addon"
	case s.Area != nil:
		return "area"
	case s.Attribute != nil:
		return "attribute"
	case s.Boolean != nil:
		return "boolean"
	case s.ColorRGB != nil:
		return "color_rgb"
	case s.ColorTemp != nil:
		return "color_temp"
	case s.Condition != nil:
		return "condition"
	case s.ConversationAgent != nil:
		return "conversation_agent"
	case s.Country != nil:
		return "country"
	case s.Date != nil:
		return "date"
	case s.Datetime != nil:
		return "datetime"
	case s.Device != nil:
		return "device"
	case s.Duration != nil:
		return "duration"
	case s.Entity != nil:
		return "entity"
	case s.File != nil:
		return "file"
	case s.Floor != nil:
		return "floor"
	case s.Icon != nil:
		return "icon"
	case s.Label != nil:
		return "label"
	case s.Language != nil:
		return "language"
	case s.Location != nil:
		return "location"
	case s.Media != nil:
		return "media"
	case s.Navigation != nil:
		return "navigation"
	case s.Number != nil:
		return "number"
	case s.Object != nil:
		return "object"
	case s.Select != nil:
		return "select"
	case s.State != nil:
		return "state"
	case s.Target != nil:
		return "target"
	case s.Template != nil:
		return "template"
	case s.Text != nil:
		return "text"
	case s.Theme != nil:
		return "theme"
	case s.Time != nil:
		return "time"
	case s.TriggerSelector != nil:
		return "trigger"
	case s.UIAction != nil:
		return "ui_action"
	case s.UIColor != nil:
		return "ui_color"
	default:
		return ""
	}
}

// Selector type definitions

// ActionSelector for action inputs
type ActionSelector struct{}

// AddonSelector for add-on selection
type AddonSelector struct {
	Name string `yaml:"name,omitempty"`
	Slug string `yaml:"slug,omitempty"`
}

// AreaSelector for area selection
type AreaSelector struct {
	Multiple bool           `yaml:"multiple,omitempty"`
	Entity   []EntityFilter `yaml:"entity,omitempty"`
	Device   []DeviceFilter `yaml:"device,omitempty"`
}

// AttributeSelector for attribute selection
type AttributeSelector struct {
	EntityID       string   `yaml:"entity_id,omitempty"`
	HideAttributes []string `yaml:"hide_attributes,omitempty"`
}

// BooleanSelector for boolean inputs
type BooleanSelector struct{}

// ColorRGBSelector for RGB color selection
type ColorRGBSelector struct{}

// ColorTempSelector for color temperature selection
type ColorTempSelector struct {
	Min       int    `yaml:"min,omitempty"`
	Max       int    `yaml:"max,omitempty"`
	MinMireds int    `yaml:"min_mireds,omitempty"`
	MaxMireds int    `yaml:"max_mireds,omitempty"`
	Unit      string `yaml:"unit,omitempty"`
}

// ConditionSelector for condition inputs
type ConditionSelector struct{}

// ConversationAgentSelector for conversation agent selection
type ConversationAgentSelector struct {
	Language string `yaml:"language,omitempty"`
}

// CountrySelector for country selection
type CountrySelector struct {
	Countries []string `yaml:"countries,omitempty"`
	NoSort    bool     `yaml:"no_sort,omitempty"`
}

// DateSelector for date selection
type DateSelector struct{}

// DatetimeSelector for datetime selection
type DatetimeSelector struct{}

// DeviceSelector for device selection
type DeviceSelector struct {
	Multiple     bool           `yaml:"multiple,omitempty"`
	Entity       []EntityFilter `yaml:"entity,omitempty"`
	Filter       []DeviceFilter `yaml:"filter,omitempty"`
	Integration  string         `yaml:"integration,omitempty"`
	Manufacturer string         `yaml:"manufacturer,omitempty"`
	Model        string         `yaml:"model,omitempty"`
}

// DeviceFilter for filtering devices
type DeviceFilter struct {
	Integration  string `yaml:"integration,omitempty"`
	Manufacturer string `yaml:"manufacturer,omitempty"`
	Model        string `yaml:"model,omitempty"`
}

// DurationSelector for duration selection
type DurationSelector struct {
	EnableDay bool `yaml:"enable_day,omitempty"`
}

// EntitySelector for entity selection
type EntitySelector struct {
	Multiple        bool           `yaml:"multiple,omitempty"`
	Filter          []EntityFilter `yaml:"filter,omitempty"`
	Domain          interface{}    `yaml:"domain,omitempty"`       // string or []string
	DeviceClass     interface{}    `yaml:"device_class,omitempty"` // string or []string
	Integration     string         `yaml:"integration,omitempty"`
	ExcludeEntities []string       `yaml:"exclude_entities,omitempty"`
	IncludeEntities []string       `yaml:"include_entities,omitempty"`
}

// EntityFilter for filtering entities
type EntityFilter struct {
	Domain            interface{} `yaml:"domain,omitempty"`       // string or []string
	DeviceClass       interface{} `yaml:"device_class,omitempty"` // string or []string
	Integration       string      `yaml:"integration,omitempty"`
	SupportedFeatures []int       `yaml:"supported_features,omitempty"`
}

// FileSelector for file selection
type FileSelector struct {
	Accept string `yaml:"accept,omitempty"`
}

// FloorSelector for floor selection
type FloorSelector struct {
	Multiple bool `yaml:"multiple,omitempty"`
}

// IconSelector for icon selection
type IconSelector struct {
	Placeholder string `yaml:"placeholder,omitempty"`
}

// LabelSelector for label selection
type LabelSelector struct {
	Multiple bool `yaml:"multiple,omitempty"`
}

// LanguageSelector for language selection
type LanguageSelector struct {
	Languages  []string `yaml:"languages,omitempty"`
	NativeName bool     `yaml:"native_name,omitempty"`
	NoSort     bool     `yaml:"no_sort,omitempty"`
}

// LocationSelector for location selection
type LocationSelector struct {
	Radius bool   `yaml:"radius,omitempty"`
	Icon   string `yaml:"icon,omitempty"`
}

// MediaSelector for media selection
type MediaSelector struct{}

// NavigationSelector for navigation selection
type NavigationSelector struct{}

// NumberSelector for number inputs
type NumberSelector struct {
	Min               float64     `yaml:"min,omitempty"`
	Max               float64     `yaml:"max,omitempty"`
	Step              interface{} `yaml:"step,omitempty"` // float64 or "any"
	Mode              string      `yaml:"mode,omitempty"` // "box" or "slider"
	UnitOfMeasurement string      `yaml:"unit_of_measurement,omitempty"`
}

// ObjectSelector for object/dictionary inputs
type ObjectSelector struct{}

// SelectSelector for select/dropdown inputs
type SelectSelector struct {
	Options        []SelectOption `yaml:"options,omitempty"`
	Multiple       bool           `yaml:"multiple,omitempty"`
	Mode           string         `yaml:"mode,omitempty"` // "list" or "dropdown"
	CustomValue    bool           `yaml:"custom_value,omitempty"`
	Sort           bool           `yaml:"sort,omitempty"`
	TranslationKey string         `yaml:"translation_key,omitempty"`
}

// SelectOption represents a select option (can be string or label/value pair)
type SelectOption struct {
	Label       string      `yaml:"label,omitempty"`
	Value       string      `yaml:"value,omitempty"`
	Description string      `yaml:"description,omitempty"`
	RawValue    interface{} `yaml:"-"` // For simple string options
}

// StateSelector for state selection
type StateSelector struct {
	EntityID string `yaml:"entity_id,omitempty"`
}

// TargetSelector for target (entity/device/area) selection
type TargetSelector struct {
	Entity []EntityFilter `yaml:"entity,omitempty"`
	Device []DeviceFilter `yaml:"device,omitempty"`
}

// TemplateSelector for template inputs
type TemplateSelector struct{}

// TextSelector for text inputs
type TextSelector struct {
	Multiline    bool   `yaml:"multiline,omitempty"`
	Prefix       string `yaml:"prefix,omitempty"`
	Suffix       string `yaml:"suffix,omitempty"`
	Type         string `yaml:"type,omitempty"` // "text", "password", "email", "url", "tel", "search"
	Autocomplete string `yaml:"autocomplete,omitempty"`
	Multiple     bool   `yaml:"multiple,omitempty"`
}

// ThemeSelector for theme selection
type ThemeSelector struct {
	IncludeDefault bool `yaml:"include_default,omitempty"`
}

// TimeSelector for time selection
type TimeSelector struct{}

// TriggerSelector for trigger inputs
type TriggerSelector struct{}

// UIActionSelector for UI action selection
type UIActionSelector struct{}

// UIColorSelector for UI color selection
type UIColorSelector struct{}

// Trigger represents a trigger definition in a blueprint.
type Trigger struct {
	Platform      string      `yaml:"platform,omitempty"`
	TriggerType   string      `yaml:"trigger,omitempty"`
	ID            string      `yaml:"id,omitempty"`
	EntityID      interface{} `yaml:"entity_id,omitempty"` // string or []string or !input
	From          interface{} `yaml:"from,omitempty"`
	To            interface{} `yaml:"to,omitempty"`
	For           interface{} `yaml:"for,omitempty"` // string, Duration, or template
	Attribute     string      `yaml:"attribute,omitempty"`
	ValueTemplate string      `yaml:"value_template,omitempty"`
	At            interface{} `yaml:"at,omitempty"` // string or []string
	Above         interface{} `yaml:"above,omitempty"`
	Below         interface{} `yaml:"below,omitempty"`
	DeviceID      interface{} `yaml:"device_id,omitempty"`
	Domain        string      `yaml:"domain,omitempty"`
	Type          string      `yaml:"type,omitempty"`
	Subtype       string      `yaml:"subtype,omitempty"`
	Zone          string      `yaml:"zone,omitempty"`
	Event         string      `yaml:"event,omitempty"`
	Offset        string      `yaml:"offset,omitempty"`
	EventType     string      `yaml:"event_type,omitempty"`
	EventData     RawData     `yaml:"event_data,omitempty"`
	Enabled       *bool       `yaml:"enabled,omitempty"`
	Alias         string      `yaml:"alias,omitempty"`
	Variables     RawData     `yaml:"variables,omitempty"`
	Raw           RawData     `yaml:"-"` // Original map for extensions
}

// Condition represents a condition in a blueprint.
type Condition struct {
	Condition     string      `yaml:"condition,omitempty"`
	Conditions    []Condition `yaml:"conditions,omitempty"` // For and/or/not
	EntityID      interface{} `yaml:"entity_id,omitempty"`
	State         interface{} `yaml:"state,omitempty"`
	Attribute     string      `yaml:"attribute,omitempty"`
	Above         interface{} `yaml:"above,omitempty"`
	Below         interface{} `yaml:"below,omitempty"`
	ValueTemplate string      `yaml:"value_template,omitempty"`
	After         interface{} `yaml:"after,omitempty"`
	Before        interface{} `yaml:"before,omitempty"`
	Weekday       interface{} `yaml:"weekday,omitempty"` // string or []string
	Zone          interface{} `yaml:"zone,omitempty"`
	DeviceID      string      `yaml:"device_id,omitempty"`
	Domain        string      `yaml:"domain,omitempty"`
	Type          string      `yaml:"type,omitempty"`
	ID            interface{} `yaml:"id,omitempty"`    // string or []string
	Match         string      `yaml:"match,omitempty"` // "any" or "all"
	AfterOffset   string      `yaml:"after_offset,omitempty"`
	BeforeOffset  string      `yaml:"before_offset,omitempty"`
	Enabled       *bool       `yaml:"enabled,omitempty"`
	Alias         string      `yaml:"alias,omitempty"`
	Raw           RawData     `yaml:"-"` // Original map for extensions
}

// Action represents an action in a blueprint.
type Action struct {
	// Service call fields
	Service      string  `yaml:"service,omitempty"`
	Action       string  `yaml:"action,omitempty"` // Alternative to service
	Target       *Target `yaml:"target,omitempty"`
	Data         RawData `yaml:"data,omitempty"`
	DataTemplate RawData `yaml:"data_template,omitempty"` // Deprecated but still used
	Response     string  `yaml:"response_variable,omitempty"`

	// Control flow
	Choose   []ChooseOption `yaml:"choose,omitempty"`
	Default  []Action       `yaml:"default,omitempty"`
	If       interface{}    `yaml:"if,omitempty"` // []Condition or shorthand
	Then     []Action       `yaml:"then,omitempty"`
	Else     []Action       `yaml:"else,omitempty"`
	Repeat   *RepeatAction  `yaml:"repeat,omitempty"`
	Parallel []Action       `yaml:"parallel,omitempty"`
	Sequence []Action       `yaml:"sequence,omitempty"`

	// Wait actions
	Wait              *WaitAction `yaml:"wait_for_trigger,omitempty"`
	WaitTemplate      string      `yaml:"wait_template,omitempty"`
	Timeout           interface{} `yaml:"timeout,omitempty"`
	ContinueOnTimeout *bool       `yaml:"continue_on_timeout,omitempty"`

	// Other actions
	Delay       interface{} `yaml:"delay,omitempty"` // string, Duration, or template
	Event       string      `yaml:"event,omitempty"`
	EventData   RawData     `yaml:"event_data,omitempty"`
	Scene       string      `yaml:"scene,omitempty"`
	Stop        string      `yaml:"stop,omitempty"`
	Error       *bool       `yaml:"error,omitempty"`
	Variables   RawData     `yaml:"variables,omitempty"`
	SetVariable *bool       `yaml:"set_conversation_response,omitempty"`
	Enabled     *bool       `yaml:"enabled,omitempty"`
	Alias       string      `yaml:"alias,omitempty"`

	Raw RawData `yaml:"-"` // Original map for extensions
}

// Target represents the target of a service call.
type Target struct {
	EntityID interface{} `yaml:"entity_id,omitempty"` // string, []string, or !input
	DeviceID interface{} `yaml:"device_id,omitempty"` // string, []string, or !input
	AreaID   interface{} `yaml:"area_id,omitempty"`   // string, []string, or !input
	FloorID  interface{} `yaml:"floor_id,omitempty"`  // string, []string, or !input
	LabelID  interface{} `yaml:"label_id,omitempty"`  // string, []string, or !input
}

// ChooseOption represents an option in a choose action.
type ChooseOption struct {
	Conditions interface{} `yaml:"conditions,omitempty"` // []Condition or shorthand
	Sequence   []Action    `yaml:"sequence,omitempty"`
	Alias      string      `yaml:"alias,omitempty"`
}

// RepeatAction represents a repeat action configuration.
type RepeatAction struct {
	Count    interface{} `yaml:"count,omitempty"`
	While    interface{} `yaml:"while,omitempty"` // []Condition or shorthand
	Until    interface{} `yaml:"until,omitempty"` // []Condition or shorthand
	ForEach  interface{} `yaml:"for_each,omitempty"`
	Sequence []Action    `yaml:"sequence,omitempty"`
}

// WaitAction represents a wait_for_trigger action.
type WaitAction struct {
	Trigger           []Trigger   `yaml:"trigger,omitempty"`
	Timeout           interface{} `yaml:"timeout,omitempty"`
	ContinueOnTimeout *bool       `yaml:"continue_on_timeout,omitempty"`
}

// Duration represents a time duration as used in Home Assistant.
type Duration struct {
	Days         int `yaml:"days,omitempty"`
	Hours        int `yaml:"hours,omitempty"`
	Minutes      int `yaml:"minutes,omitempty"`
	Seconds      int `yaml:"seconds,omitempty"`
	Milliseconds int `yaml:"milliseconds,omitempty"`
}

// Variables represents the variables section of a blueprint.
type Variables map[string]interface{}

// --- Helper Types for Maps ---

// StringMap provides a type alias for map[string]interface{} to make
// transitions easier and improve code readability.
type StringMap = map[string]interface{}

// AnyList provides a type alias for []interface{} to make transitions easier.
type AnyList = []interface{}

// --- Conversion Functions ---

// ToRawData converts a strongly-typed BlueprintData back to RawData for
// backward compatibility with existing validation code.
func (b *BlueprintData) ToRawData() RawData {
	return b.Raw
}

// FromRawData creates a BlueprintData from RawData, preserving the original
// raw data for backward compatibility.
func FromRawData(raw RawData) *BlueprintData {
	return &BlueprintData{
		Raw: raw,
	}
}
