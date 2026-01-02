// Package validator provides comprehensive validation for Home Assistant Blueprint files.
package validator

import "regexp"

// ValidModes are the valid automation modes
var ValidModes = []string{"single", "restart", "queued", "parallel"}

// ValidConditionTypes are the valid condition types
var ValidConditionTypes = []string{
	"and", "or", "not", "state", "numeric_state", "template",
	"time", "zone", "trigger", "sun", "device",
}

// ValidSelectorTypes are the valid input selector types
var ValidSelectorTypes = map[string]bool{
	"action": true, "addon": true, "area": true, "attribute": true,
	"boolean": true, "color_rgb": true, "color_temp": true, "condition": true,
	"conversation_agent": true, "country": true, "date": true, "datetime": true,
	"device": true, "duration": true, "entity": true, "file": true,
	"floor": true, "icon": true, "label": true, "language": true,
	"location": true, "media": true, "navigation": true, "number": true,
	"object": true, "select": true, "state": true, "target": true,
	"template": true, "text": true, "theme": true, "time": true,
	"trigger": true, "ui_action": true, "ui_color": true,
}

// RequiredBlueprintKeys are required in the blueprint section
var RequiredBlueprintKeys = []string{"name", "description", "domain", "input"}

// RequiredRootKeys are required at the root level
var RequiredRootKeys = []string{"blueprint", "trigger", "action"}

// HysteresisPattern defines a pattern pair for hysteresis validation
type HysteresisPattern struct {
	OnPattern   *regexp.Regexp
	OffPattern  string
	Description string
}

// HysteresisPatterns for detecting hysteresis configuration issues
var HysteresisPatterns = []HysteresisPattern{
	{regexp.MustCompile(`(.*)_on$`), "${1}_off", "threshold"},
	{regexp.MustCompile(`(.*)_high$`), "${1}_low", "boundary"},
	{regexp.MustCompile(`(.*)_upper$`), "${1}_lower", "limit"},
	{regexp.MustCompile(`(.*)_start$`), "${1}_stop", "trigger point"},
	{regexp.MustCompile(`(.*)_enable$`), "${1}_disable", "activation point"},
	{regexp.MustCompile(`delta_on$`), "delta_off", "delta threshold"},
}

// Jinja2Builtins are built-in Jinja2/HA template functions that shouldn't trigger undefined warnings
var Jinja2Builtins = map[string]bool{
	// Python/Jinja2 built-in constants
	"true": true, "false": true, "none": true,
	"True": true, "False": true, "None": true,
	// Jinja2 control keywords
	"if": true, "else": true, "elif": true, "endif": true,
	"for": true, "endfor": true, "in": true, "not": true,
	"and": true, "or": true, "is": true, "set": true, "endset": true,
	"macro": true, "endmacro": true, "call": true, "endcall": true,
	"filter": true, "endfilter": true, "block": true, "endblock": true,
	"extends": true, "include": true, "import": true, "from": true,
	"as": true, "with": true, "endwith": true, "do": true,
	"continue": true, "break": true,
	// Jinja2 tests
	"defined": true, "undefined": true, "number": true, "string": true,
	"mapping": true, "iterable": true, "callable": true, "sequence": true,
	"sameas": true, "escaped": true, "even": true, "odd": true,
	"divisibleby": true, "lower": true, "upper": true,
	// Jinja2 built-in filters
	"abs": true, "attr": true, "batch": true, "capitalize": true,
	"center": true, "count": true, "default": true, "dictsort": true,
	"escape": true, "filesizeformat": true, "first": true, "float": true,
	"forceescape": true, "format": true, "groupby": true, "indent": true,
	"int": true, "items": true, "join": true, "last": true, "length": true,
	"list": true, "map": true, "max": true, "min": true, "pprint": true,
	"random": true, "reject": true, "rejectattr": true, "replace": true,
	"reverse": true, "round": true, "safe": true, "select": true,
	"selectattr": true, "slice": true, "sort": true, "split": true,
	"striptags": true, "sum": true, "title": true, "tojson": true,
	"trim": true, "truncate": true, "unique": true, "urlencode": true,
	"urlize": true, "wordcount": true, "wordwrap": true, "xmlattr": true,
	// Home Assistant specific functions
	"states": true, "is_state": true, "state_attr": true, "is_state_attr": true,
	"has_value": true, "expand": true, "device_entities": true, "area_entities": true,
	"integration_entities": true, "device_attr": true, "device_id": true,
	"area_name": true, "area_id": true, "floor_id": true, "floor_name": true,
	"label_id": true, "label_name": true, "labels": true, "relative_time": true,
	"time_since": true, "timedelta": true, "strptime": true, "strftime": true,
	"as_timestamp": true, "as_datetime": true, "as_local": true, "as_timedelta": true,
	"today_at": true, "now": true, "utcnow": true, "distance": true,
	"closest": true, "iif": true, "log": true, "sin": true, "cos": true,
	"tan": true, "asin": true, "acos": true, "atan": true, "atan2": true,
	"sqrt": true, "e": true, "pi": true, "tau": true, "inf": true,
	"average": true, "median": true, "statistical_mode": true, "pack": true,
	"unpack": true, "ord": true, "base64_encode": true, "base64_decode": true,
	"slugify": true, "regex_match": true, "regex_search": true, "regex_replace": true,
	"regex_findall": true, "regex_findall_index": true, "from_json": true,
	"to_json": true, "value_json": true, "trigger": true, "this": true,
	"context": true, "repeat": true, "wait": true, "namespace": true,
	// Common loop variables
	"item": true, "loop": true, "index": true, "index0": true,
	"cycle": true, "depth": true, "depth0": true,
	"previtem": true, "nextitem": true, "changed": true,
	// Datetime attributes
	"year": true, "month": true, "day": true, "hour": true, "minute": true,
	"second": true, "microsecond": true, "weekday": true, "isoweekday": true,
	"isocalendar": true, "isoformat": true, "date": true, "time": true,
	"timestamp": true, "tzinfo": true, "tzname": true, "utcoffset": true,
	"dst": true, "timetuple": true,
	// State object attributes
	"state": true, "attributes": true, "entity_id": true, "domain": true,
	"object_id": true, "name": true, "last_changed": true, "last_updated": true,
	"last_reported": true, "context_id": true,
	// Trigger object attributes
	"platform": true, "event": true, "to_state": true, "from_state": true,
	"idx": true, "id": true, "description": true, "alias": true,
	// Additional common attributes
	"friendly_name": true, "icon": true, "unit_of_measurement": true,
	"device_class": true, "brightness": true, "color_temp": true, "hs_color": true,
	"rgb_color": true, "xy_color": true, "temperature": true, "humidity": true,
	"pressure": true, "position": true, "current_position": true,
	"current_temperature": true, "target_temperature": true, "hvac_mode": true,
	"hvac_action": true, "fan_mode": true, "swing_mode": true, "preset_mode": true,
	"speed": true, "percentage": true, "battery_level": true, "battery": true,
	"power": true, "voltage": true, "current": true, "energy": true,
	"elevation": true, "azimuth": true, "rising": true, "setting": true,
	"next_rising": true, "next_setting": true,
}
