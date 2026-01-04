package validators

import (
	"testing"
)

func TestGetMap(t *testing.T) {
	tests := []struct {
		name      string
		value     interface{}
		path      string
		wantOk    bool
		wantError string
	}{
		{
			name:   "valid map",
			value:  map[string]interface{}{"key": "value"},
			path:   "test",
			wantOk: true,
		},
		{
			name:      "nil value",
			value:     nil,
			path:      "test",
			wantOk:    false,
			wantError: "test: value is nil",
		},
		{
			name:      "wrong type",
			value:     "not a map",
			path:      "test",
			wantOk:    false,
			wantError: "test: must be a dictionary, got string",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, ok, errMsg := GetMap(tt.value, tt.path)
			if ok != tt.wantOk {
				t.Errorf("GetMap() ok = %v, want %v", ok, tt.wantOk)
			}
			if tt.wantOk && result == nil {
				t.Error("GetMap() result should not be nil when ok is true")
			}
			if !tt.wantOk && errMsg != tt.wantError {
				t.Errorf("GetMap() errMsg = %q, want %q", errMsg, tt.wantError)
			}
		})
	}
}

func TestGetString(t *testing.T) {
	tests := []struct {
		name      string
		value     interface{}
		path      string
		wantOk    bool
		wantValue string
	}{
		{
			name:      "valid string",
			value:     "hello",
			path:      "test",
			wantOk:    true,
			wantValue: "hello",
		},
		{
			name:   "nil value",
			value:  nil,
			path:   "test",
			wantOk: false,
		},
		{
			name:   "wrong type",
			value:  123,
			path:   "test",
			wantOk: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, ok, _ := GetString(tt.value, tt.path)
			if ok != tt.wantOk {
				t.Errorf("GetString() ok = %v, want %v", ok, tt.wantOk)
			}
			if ok && result != tt.wantValue {
				t.Errorf("GetString() = %q, want %q", result, tt.wantValue)
			}
		})
	}
}

func TestGetList(t *testing.T) {
	result, ok, _ := GetList([]interface{}{"a", "b"}, "test")
	if !ok || len(result) != 2 {
		t.Error("GetList should return valid list")
	}

	_, ok, errMsg := GetList(nil, "test")
	if ok || errMsg == "" {
		t.Error("GetList should fail on nil")
	}
}

func TestGetBool(t *testing.T) {
	result, ok, _ := GetBool(true, "test")
	if !ok || !result {
		t.Error("GetBool should return true")
	}

	result, ok, _ = GetBool(false, "test")
	if !ok || result {
		t.Error("GetBool should return false")
	}

	_, ok, _ = GetBool("not a bool", "test")
	if ok {
		t.Error("GetBool should fail on string")
	}
}

func TestGetInt(t *testing.T) {
	result, ok, _ := GetInt(42, "test")
	if !ok || result != 42 {
		t.Error("GetInt should return 42")
	}

	_, ok, _ = GetInt("not an int", "test")
	if ok {
		t.Error("GetInt should fail on string")
	}
}

func TestTryGetFunctions(t *testing.T) {
	parent := map[string]interface{}{
		"str":  "value",
		"num":  42,
		"bool": true,
		"list": []interface{}{"a"},
		"map":  map[string]interface{}{"nested": "value"},
	}

	// TryGetString
	if s, ok := TryGetString(parent, "str"); !ok || s != "value" {
		t.Error("TryGetString failed")
	}
	if _, ok := TryGetString(parent, "missing"); ok {
		t.Error("TryGetString should return false for missing key")
	}
	if _, ok := TryGetString(parent, "num"); ok {
		t.Error("TryGetString should return false for wrong type")
	}

	// TryGetMap
	if m, ok := TryGetMap(parent, "map"); !ok || m["nested"] != "value" {
		t.Error("TryGetMap failed")
	}

	// TryGetList
	if l, ok := TryGetList(parent, "list"); !ok || len(l) != 1 {
		t.Error("TryGetList failed")
	}

	// TryGetBool
	if b, ok := TryGetBool(parent, "bool"); !ok || !b {
		t.Error("TryGetBool failed")
	}

	// TryGetInt
	if i, ok := TryGetInt(parent, "num"); !ok || i != 42 {
		t.Error("TryGetInt failed")
	}
}

func TestPathFunctions(t *testing.T) {
	// JoinPath
	if JoinPath("", "child") != "child" {
		t.Error("JoinPath with empty parent failed")
	}
	if JoinPath("parent", "child") != "parent.child" {
		t.Error("JoinPath failed")
	}

	// IndexPath
	if IndexPath("arr", 0) != "arr[0]" {
		t.Error("IndexPath failed")
	}

	// KeyPath
	if KeyPath("obj", "key") != "obj.key" {
		t.Error("KeyPath failed")
	}
}

func TestValidationIssue(t *testing.T) {
	// With path
	issue := ValidationIssue{Severity: SeverityError, Path: "some.path", Message: "error message"}
	if issue.String() != "some.path: error message" {
		t.Errorf("String() = %q, want %q", issue.String(), "some.path: error message")
	}

	// Without path
	issue2 := ValidationIssue{Severity: SeverityWarning, Message: "warning message"}
	if issue2.String() != "warning message" {
		t.Errorf("String() = %q, want %q", issue2.String(), "warning message")
	}

	// IsError/IsWarning
	if !issue.IsError() || issue.IsWarning() {
		t.Error("Error issue should be error, not warning")
	}
	if issue2.IsError() || !issue2.IsWarning() {
		t.Error("Warning issue should be warning, not error")
	}
}

func TestValidationResult(t *testing.T) {
	result := NewValidationResult()

	if result.HasErrors() || result.HasWarnings() {
		t.Error("New result should have no issues")
	}

	result.AddError("path1", "error 1")
	result.AddErrorf("path2", "error %d", 2)
	result.AddWarning("path3", "warning 1")
	result.AddWarningf("path4", "warning %d", 2)

	if len(result.Issues) != 4 {
		t.Errorf("Issues count = %d, want 4", len(result.Issues))
	}

	if !result.HasErrors() {
		t.Error("Should have errors")
	}
	if !result.HasWarnings() {
		t.Error("Should have warnings")
	}

	errors := result.Errors()
	if len(errors) != 2 {
		t.Errorf("Error count = %d, want 2", len(errors))
	}

	warnings := result.Warnings()
	if len(warnings) != 2 {
		t.Errorf("Warning count = %d, want 2", len(warnings))
	}

	// Test Merge
	other := NewValidationResult()
	other.AddError("other", "other error")
	result.Merge(other)

	if len(result.Issues) != 5 {
		t.Errorf("After merge, issues count = %d, want 5", len(result.Issues))
	}

	// Test nil merge
	result.Merge(nil)
	if len(result.Issues) != 5 {
		t.Error("Merge nil should not add issues")
	}

	// Test ErrorStrings/WarningStrings
	errorStrings := result.ErrorStrings()
	if len(errorStrings) != 3 {
		t.Errorf("ErrorStrings count = %d, want 3", len(errorStrings))
	}

	warningStrings := result.WarningStrings()
	if len(warningStrings) != 2 {
		t.Errorf("WarningStrings count = %d, want 2", len(warningStrings))
	}
}

func TestValidateRequired(t *testing.T) {
	m := map[string]interface{}{"key": "value"}

	if err := ValidateRequired(m, "key", "test"); err != "" {
		t.Error("ValidateRequired should pass for existing key")
	}

	if err := ValidateRequired(m, "missing", "test"); err == "" {
		t.Error("ValidateRequired should fail for missing key")
	}
}

func TestValidateRequiredKeys(t *testing.T) {
	m := map[string]interface{}{"a": 1, "b": 2}

	errors := ValidateRequiredKeys(m, []string{"a", "b"}, "test")
	if len(errors) != 0 {
		t.Error("ValidateRequiredKeys should pass when all keys present")
	}

	errors = ValidateRequiredKeys(m, []string{"a", "c"}, "test")
	if len(errors) != 1 {
		t.Errorf("ValidateRequiredKeys error count = %d, want 1", len(errors))
	}
}

func TestValidateEnumValue(t *testing.T) {
	allowed := []string{"a", "b", "c"}

	if err := ValidateEnumValue("a", allowed, "test", "field"); err != "" {
		t.Error("ValidateEnumValue should pass for valid value")
	}

	if err := ValidateEnumValue("x", allowed, "test", "field"); err == "" {
		t.Error("ValidateEnumValue should fail for invalid value")
	}
}

func TestValidateEnumMap(t *testing.T) {
	allowed := map[string]bool{"a": true, "b": true}

	if err := ValidateEnumMap("a", allowed, "test", "field"); err != "" {
		t.Error("ValidateEnumMap should pass for valid value")
	}

	if err := ValidateEnumMap("x", allowed, "test", "field"); err == "" {
		t.Error("ValidateEnumMap should fail for invalid value")
	}
}

func TestValidatePositiveInt(t *testing.T) {
	if err := ValidatePositiveInt(1, "test", "field"); err != "" {
		t.Error("ValidatePositiveInt should pass for 1")
	}

	if err := ValidatePositiveInt(0, "test", "field"); err == "" {
		t.Error("ValidatePositiveInt should fail for 0")
	}

	if err := ValidatePositiveInt(-1, "test", "field"); err == "" {
		t.Error("ValidatePositiveInt should fail for -1")
	}
}

func TestValidateNotNil(t *testing.T) {
	if err := ValidateNotNil("value", "test", "field"); err != "" {
		t.Error("ValidateNotNil should pass for non-nil")
	}

	if err := ValidateNotNil(nil, "test", "field"); err == "" {
		t.Error("ValidateNotNil should fail for nil")
	}
}

func TestValidateNotEmpty(t *testing.T) {
	if err := ValidateNotEmpty("value", "test", "field"); err != "" {
		t.Error("ValidateNotEmpty should pass for non-empty")
	}

	if err := ValidateNotEmpty("", "test", "field"); err == "" {
		t.Error("ValidateNotEmpty should fail for empty")
	}
}

func TestContainsTemplate(t *testing.T) {
	if !ContainsTemplate("{{ value }}") {
		t.Error("Should detect {{ }}")
	}
	if !ContainsTemplate("{% if true %}") {
		t.Error("Should detect {% %}")
	}
	if ContainsTemplate("plain text") {
		t.Error("Should not detect in plain text")
	}
}

func TestContainsInputRef(t *testing.T) {
	if !ContainsInputRef("!input my_input") {
		t.Error("Should detect !input")
	}
	if ContainsInputRef("plain text") {
		t.Error("Should not detect in plain text")
	}
}

func TestValidateBalancedDelimiters(t *testing.T) {
	// Balanced
	errors := ValidateBalancedDelimiters("{{ value }}", "test")
	if len(errors) != 0 {
		t.Error("Should pass for balanced {{ }}")
	}

	// Unbalanced {{
	errors = ValidateBalancedDelimiters("{{ value", "test")
	if len(errors) != 1 {
		t.Error("Should fail for unbalanced {{ }}")
	}

	// Unbalanced {%
	errors = ValidateBalancedDelimiters("{% if true %}", "test")
	if len(errors) != 0 {
		t.Error("Should pass for balanced {% %}")
	}

	errors = ValidateBalancedDelimiters("{% if true", "test")
	if len(errors) != 1 {
		t.Error("Should fail for unbalanced {% %}")
	}
}

func TestValidateNoInputInTemplate(t *testing.T) {
	if err := ValidateNoInputInTemplate("{{ !input my_input }}", "test"); err == "" {
		t.Error("Should fail for !input inside {{ }}")
	}

	if err := ValidateNoInputInTemplate("!input my_input", "test"); err != "" {
		t.Error("Should pass for !input outside {{ }}")
	}
}

func TestValidateNoTemplateInField(t *testing.T) {
	if err := ValidateNoTemplateInField("{{ value }}", "test", "field"); err == "" {
		t.Error("Should fail for template in field")
	}

	if err := ValidateNoTemplateInField("static value", "test", "field"); err != "" {
		t.Error("Should pass for static value")
	}
}

func TestTraverseValue(t *testing.T) {
	data := map[string]interface{}{
		"str": "value",
		"nested": map[string]interface{}{
			"inner": "inner_value",
		},
		"list": []interface{}{"a", "b"},
	}

	var visited []string
	TraverseValue(data, "root", func(v interface{}, path string) bool {
		visited = append(visited, path)
		return true
	})

	// Should have visited: root, root.str, root.nested, root.nested.inner, root.list, root.list[0], root.list[1]
	if len(visited) < 7 {
		t.Errorf("Should have visited at least 7 nodes, got %d", len(visited))
	}
}

func TestCollectStrings(t *testing.T) {
	data := map[string]interface{}{
		"str": "value1",
		"nested": map[string]interface{}{
			"inner": "value2",
		},
		"list":   []interface{}{"value3", 123},
		"number": 42,
	}

	strings := CollectStrings(data)
	if len(strings) != 3 {
		t.Errorf("CollectStrings count = %d, want 3", len(strings))
	}
}

func TestExtractInputRef(t *testing.T) {
	if ref := ExtractInputRef("!input my_input"); ref != "my_input" {
		t.Errorf("ExtractInputRef = %q, want %q", ref, "my_input")
	}

	if ref := ExtractInputRef("!input   spaced  "); ref != "spaced" {
		t.Errorf("ExtractInputRef with spaces = %q, want %q", ref, "spaced")
	}

	if ref := ExtractInputRef("not an input"); ref != "" {
		t.Error("ExtractInputRef should return empty for non-input")
	}
}

func TestCollectInputRefsFromValue(t *testing.T) {
	data := map[string]interface{}{
		"input1": "!input first",
		"nested": map[string]interface{}{
			"input2": "!input second",
		},
		"list": []interface{}{"!input third", "not input"},
	}

	refs := CollectInputRefsFromValue(data)
	if len(refs) != 3 {
		t.Errorf("CollectInputRefsFromValue count = %d, want 3", len(refs))
	}
	if !refs["first"] || !refs["second"] || !refs["third"] {
		t.Error("Missing expected input references")
	}
}

func TestValidateListItems(t *testing.T) {
	list := []interface{}{"valid", "invalid", "valid"}

	result := ValidateListItems(list, "test", func(item interface{}, index int, path string) *ValidationResult {
		r := NewValidationResult()
		if item == "invalid" {
			r.AddError(path, "invalid item")
		}
		return r
	})

	if len(result.Errors()) != 1 {
		t.Errorf("Error count = %d, want 1", len(result.Errors()))
	}
}

func TestValidateMapEntries(t *testing.T) {
	m := map[string]interface{}{
		"valid":   "ok",
		"invalid": "bad",
	}

	result := ValidateMapEntries(m, "test", func(key string, value interface{}, path string) *ValidationResult {
		r := NewValidationResult()
		if key == "invalid" {
			r.AddError(path, "invalid entry")
		}
		return r
	})

	if len(result.Errors()) != 1 {
		t.Errorf("Error count = %d, want 1", len(result.Errors()))
	}
}

func TestValidateIf(t *testing.T) {
	if err := ValidateIf(true, func() string { return "error" }); err != "error" {
		t.Error("ValidateIf should call validator when condition is true")
	}

	if err := ValidateIf(false, func() string { return "error" }); err != "" {
		t.Error("ValidateIf should not call validator when condition is false")
	}
}

func TestValidateIfPresent(t *testing.T) {
	m := map[string]interface{}{"key": "value"}

	if err := ValidateIfPresent(m, "key", func(v interface{}) string {
		return "error"
	}); err != "error" {
		t.Error("ValidateIfPresent should call validator when key present")
	}

	if err := ValidateIfPresent(m, "missing", func(v interface{}) string {
		return "error"
	}); err != "" {
		t.Error("ValidateIfPresent should not call validator when key missing")
	}
}

func TestValidateServiceFormat(t *testing.T) {
	// Valid formats
	validCases := []string{
		"light.turn_on",
		"!input service_name",
		"{{ service }}",
		"{% if true %}light.turn_on{% endif %}",
	}

	for _, c := range validCases {
		if err := ValidateServiceFormat(c, "test"); err != "" {
			t.Errorf("ValidateServiceFormat(%q) should pass", c)
		}
	}

	// Invalid format
	if err := ValidateServiceFormat("turn_on", "test"); err == "" {
		t.Error("ValidateServiceFormat should fail for 'turn_on'")
	}
}

func TestValidateSelector(t *testing.T) {
	validTypes := map[string]bool{"boolean": true, "text": true}

	if err := ValidateSelector("boolean", validTypes, "test"); err != "" {
		t.Error("ValidateSelector should pass for valid type")
	}

	if err := ValidateSelector("unknown", validTypes, "test"); err == "" {
		t.Error("ValidateSelector should fail for unknown type")
	}
}

func TestValidateEntityDomain(t *testing.T) {
	domains := []string{"light", "switch"}

	if err := ValidateEntityDomain("light", domains, "test"); err != "" {
		t.Error("ValidateEntityDomain should pass for valid domain")
	}

	if err := ValidateEntityDomain("sensor", domains, "test"); err == "" {
		t.Error("ValidateEntityDomain should fail for invalid domain")
	}
}
