package validator

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// ToFloat converts various types to float64
func ToFloat(val interface{}) (float64, error) {
	switch v := val.(type) {
	case float64:
		return v, nil
	case float32:
		return float64(v), nil
	case int:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case string:
		return strconv.ParseFloat(v, 64)
	default:
		return 0, fmt.Errorf("cannot convert to float")
	}
}

// Abs returns the absolute value of a float64
func Abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

// ContainsVariableRef checks if template contains variable references (not !input)
func ContainsVariableRef(template string) bool {
	varPattern := regexp.MustCompile(`\{\{[^}]*\b[a-z_][a-z0-9_]*\b[^}]*\}\}`)
	return varPattern.MatchString(template)
}

// ContainsInputRef checks if template contains !input references
func ContainsInputRef(template string) bool {
	return strings.Contains(template, "!input")
}

// CollectInputRefs collects !input references from a string
func (v *BlueprintValidator) CollectInputRefs(value string) {
	if after, ok := strings.CutPrefix(value, "!input "); ok {
		inputName := after
		inputName = strings.TrimSpace(inputName)
		v.UsedInputs[inputName] = true
	}
}

// CollectInputRefsFromMap recursively collects !input references from a map
func (v *BlueprintValidator) CollectInputRefsFromMap(m map[string]interface{}) {
	for _, value := range m {
		switch val := value.(type) {
		case string:
			v.CollectInputRefs(val)
		case map[string]interface{}:
			v.CollectInputRefsFromMap(val)
		case []interface{}:
			for _, item := range val {
				switch itemVal := item.(type) {
				case string:
					v.CollectInputRefs(itemVal)
				case map[string]interface{}:
					v.CollectInputRefsFromMap(itemVal)
				}
			}
		}
	}
}
