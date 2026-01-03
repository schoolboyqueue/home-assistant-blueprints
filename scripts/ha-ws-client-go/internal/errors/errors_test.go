package errors

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestErrorType_String(t *testing.T) {
	tests := []struct {
		name     string
		errType  ErrorType
		expected string
	}{
		{"Unknown", ErrorTypeUnknown, "unknown"},
		{"Network", ErrorTypeNetwork, "network"},
		{"Validation", ErrorTypeValidation, "validation"},
		{"Parsing", ErrorTypeParsing, "parsing"},
		{"Timeout", ErrorTypeTimeout, "timeout"},
		{"NotFound", ErrorTypeNotFound, "not_found"},
		{"Auth", ErrorTypeAuth, "auth"},
		{"API", ErrorTypeAPI, "api"},
		{"Internal", ErrorTypeInternal, "internal"},
		{"Canceled", ErrorTypeCanceled, "canceled"},
		{"Subscription", ErrorTypeSubscription, "subscription"},
		{"Invalid", ErrorType(999), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.errType.String())
		})
	}
}

func TestError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *Error
		expected string
	}{
		{
			name:     "Simple message",
			err:      &Error{Type: ErrorTypeNetwork, Message: "connection failed"},
			expected: "connection failed",
		},
		{
			name:     "With code",
			err:      &Error{Type: ErrorTypeNetwork, Code: "conn_fail", Message: "connection failed"},
			expected: "[conn_fail] connection failed",
		},
		{
			name:     "With cause",
			err:      &Error{Type: ErrorTypeNetwork, Message: "connection failed", Cause: errors.New("timeout")},
			expected: "connection failed: timeout",
		},
		{
			name:     "With code and cause",
			err:      &Error{Type: ErrorTypeNetwork, Code: "conn_fail", Message: "connection failed", Cause: errors.New("timeout")},
			expected: "[conn_fail] connection failed: timeout",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.err.Error())
		})
	}
}

func TestError_Unwrap(t *testing.T) {
	cause := errors.New("underlying error")
	err := &Error{
		Type:    ErrorTypeNetwork,
		Message: "wrapped error",
		Cause:   cause,
	}

	assert.Equal(t, cause, err.Unwrap())
	assert.True(t, errors.Is(err, cause))
}

func TestError_Is(t *testing.T) {
	t.Run("Same type and code", func(t *testing.T) {
		err1 := &Error{Type: ErrorTypeNotFound, Code: "entity_not_found"}
		err2 := &Error{Type: ErrorTypeNotFound, Code: "entity_not_found"}
		assert.True(t, err1.Is(err2))
	})

	t.Run("Same type, different code", func(t *testing.T) {
		err1 := &Error{Type: ErrorTypeNotFound, Code: "entity_not_found"}
		err2 := &Error{Type: ErrorTypeNotFound, Code: "service_not_found"}
		assert.False(t, err1.Is(err2))
	})

	t.Run("Same type, no code", func(t *testing.T) {
		err1 := &Error{Type: ErrorTypeNotFound}
		err2 := &Error{Type: ErrorTypeNotFound}
		assert.True(t, err1.Is(err2))
	})

	t.Run("Different type", func(t *testing.T) {
		err1 := &Error{Type: ErrorTypeNotFound}
		err2 := &Error{Type: ErrorTypeNetwork}
		assert.False(t, err1.Is(err2))
	})

	t.Run("Non-Error target", func(t *testing.T) {
		err1 := &Error{Type: ErrorTypeNotFound}
		err2 := errors.New("standard error")
		assert.False(t, err1.Is(err2))
	})
}

func TestError_WithDetails(t *testing.T) {
	original := &Error{
		Type:    ErrorTypeNotFound,
		Message: "entity not found",
		Details: map[string]any{"key1": "value1"},
	}

	modified := original.WithDetails(map[string]any{"key2": "value2"})

	// Original should be unchanged
	assert.Equal(t, map[string]any{"key1": "value1"}, original.Details)

	// Modified should have both
	assert.Equal(t, map[string]any{"key1": "value1", "key2": "value2"}, modified.Details)
}

func TestError_WithCause(t *testing.T) {
	original := &Error{
		Type:    ErrorTypeNetwork,
		Message: "connection failed",
	}
	cause := errors.New("timeout")

	modified := original.WithCause(cause)

	// Original should be unchanged
	assert.Nil(t, original.Cause)

	// Modified should have cause
	assert.Equal(t, cause, modified.Cause)
}

func TestError_WithMessage(t *testing.T) {
	original := &Error{
		Type:    ErrorTypeNotFound,
		Message: "not found",
	}

	modified := original.WithMessage("entity light.living_room not found")

	// Original should be unchanged
	assert.Equal(t, "not found", original.Message)

	// Modified should have new message
	assert.Equal(t, "entity light.living_room not found", modified.Message)
}

func TestError_WithMessagef(t *testing.T) {
	original := &Error{
		Type:    ErrorTypeNotFound,
		Message: "not found",
	}

	modified := original.WithMessagef("entity %s not found", "light.living_room")

	assert.Equal(t, "entity light.living_room not found", modified.Message)
}

func TestNew(t *testing.T) {
	err := New(ErrorTypeNetwork, "connection failed")

	assert.Equal(t, ErrorTypeNetwork, err.Type)
	assert.Equal(t, "connection failed", err.Message)
	assert.Empty(t, err.Code)
	assert.Nil(t, err.Cause)
	assert.NotNil(t, err.Details)
}

func TestNewf(t *testing.T) {
	err := Newf(ErrorTypeNotFound, "entity %s not found", "light.living_room")

	assert.Equal(t, ErrorTypeNotFound, err.Type)
	assert.Equal(t, "entity light.living_room not found", err.Message)
}

func TestWrap(t *testing.T) {
	cause := errors.New("underlying")
	err := Wrap(ErrorTypeNetwork, cause, "connection failed")

	assert.Equal(t, ErrorTypeNetwork, err.Type)
	assert.Equal(t, "connection failed", err.Message)
	assert.Equal(t, cause, err.Cause)
}

func TestWrapf(t *testing.T) {
	cause := errors.New("underlying")
	err := Wrapf(ErrorTypeNotFound, cause, "entity %s not found", "light.living_room")

	assert.Equal(t, ErrorTypeNotFound, err.Type)
	assert.Equal(t, "entity light.living_room not found", err.Message)
	assert.Equal(t, cause, err.Cause)
}

func TestGetType(t *testing.T) {
	t.Run("Typed error", func(t *testing.T) {
		err := &Error{Type: ErrorTypeNetwork}
		assert.Equal(t, ErrorTypeNetwork, GetType(err))
	})

	t.Run("Wrapped typed error", func(t *testing.T) {
		inner := &Error{Type: ErrorTypeNotFound}
		outer := Wrap(ErrorTypeNetwork, inner, "outer")
		// GetType returns the outermost type
		assert.Equal(t, ErrorTypeNetwork, GetType(outer))
	})

	t.Run("Standard error", func(t *testing.T) {
		err := errors.New("standard error")
		assert.Equal(t, ErrorTypeUnknown, GetType(err))
	})

	t.Run("Nil error", func(t *testing.T) {
		assert.Equal(t, ErrorTypeUnknown, GetType(nil))
	})
}

func TestGetCode(t *testing.T) {
	t.Run("Error with code", func(t *testing.T) {
		err := &Error{Type: ErrorTypeNotFound, Code: "entity_not_found"}
		assert.Equal(t, "entity_not_found", GetCode(err))
	})

	t.Run("Error without code", func(t *testing.T) {
		err := &Error{Type: ErrorTypeNotFound}
		assert.Equal(t, "", GetCode(err))
	})

	t.Run("Standard error", func(t *testing.T) {
		err := errors.New("standard")
		assert.Equal(t, "", GetCode(err))
	})
}

func TestGetDetails(t *testing.T) {
	t.Run("Error with details", func(t *testing.T) {
		err := &Error{
			Type:    ErrorTypeNotFound,
			Details: map[string]any{"entity_id": "light.test"},
		}
		assert.Equal(t, map[string]any{"entity_id": "light.test"}, GetDetails(err))
	})

	t.Run("Error without details", func(t *testing.T) {
		err := &Error{Type: ErrorTypeNotFound}
		assert.Nil(t, GetDetails(err))
	})

	t.Run("Standard error", func(t *testing.T) {
		err := errors.New("standard")
		assert.Nil(t, GetDetails(err))
	})
}

func TestIsType(t *testing.T) {
	err := &Error{Type: ErrorTypeNetwork}

	assert.True(t, IsType(err, ErrorTypeNetwork))
	assert.False(t, IsType(err, ErrorTypeNotFound))
}

func TestTypeCheckers(t *testing.T) {
	tests := []struct {
		name    string
		errType ErrorType
		checker func(error) bool
	}{
		{"IsNetwork", ErrorTypeNetwork, IsNetwork},
		{"IsValidation", ErrorTypeValidation, IsValidation},
		{"IsParsing", ErrorTypeParsing, IsParsing},
		{"IsTimeout", ErrorTypeTimeout, IsTimeout},
		{"IsNotFound", ErrorTypeNotFound, IsNotFound},
		{"IsAuth", ErrorTypeAuth, IsAuth},
		{"IsAPI", ErrorTypeAPI, IsAPI},
		{"IsInternal", ErrorTypeInternal, IsInternal},
		{"IsCanceled", ErrorTypeCanceled, IsCanceled},
		{"IsSubscription", ErrorTypeSubscription, IsSubscription},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := &Error{Type: tt.errType}
			assert.True(t, tt.checker(err))

			// Check that other types don't match
			otherErr := &Error{Type: ErrorTypeUnknown}
			assert.False(t, tt.checker(otherErr))
		})
	}
}

func TestErrorsAs(t *testing.T) {
	typedErr := &Error{
		Type:    ErrorTypeNotFound,
		Code:    "entity_not_found",
		Message: "entity not found",
	}

	var target *Error
	require.True(t, errors.As(typedErr, &target))
	assert.Equal(t, ErrorTypeNotFound, target.Type)
	assert.Equal(t, "entity_not_found", target.Code)
}
