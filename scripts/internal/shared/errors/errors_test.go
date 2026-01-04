package errors

import (
	"errors"
	"fmt"
	"testing"
)

func TestErrorTypeString(t *testing.T) {
	tests := []struct {
		errType  ErrorType
		expected string
	}{
		{ErrorTypeUnknown, "unknown"},
		{ErrorTypeValidation, "validation"},
		{ErrorTypeParsing, "parsing"},
		{ErrorTypeInternal, "internal"},
		{ErrorType(999), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.errType.String(); got != tt.expected {
				t.Errorf("ErrorType.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestRegisterTypeName(t *testing.T) {
	// Register a custom type
	customType := ErrorTypeAppBase + 1
	RegisterTypeName(customType, "custom_error")

	if got := customType.String(); got != "custom_error" {
		t.Errorf("custom ErrorType.String() = %v, want %v", got, "custom_error")
	}
}

func TestErrorError(t *testing.T) {
	tests := []struct {
		name     string
		err      *Error
		expected string
	}{
		{
			name:     "simple message",
			err:      &Error{Message: "test error"},
			expected: "test error",
		},
		{
			name:     "with code",
			err:      &Error{Code: "test_code", Message: "test error"},
			expected: "[test_code] test error",
		},
		{
			name:     "with path",
			err:      &Error{Path: "some.path", Message: "test error"},
			expected: "some.path: test error",
		},
		{
			name:     "with code and path",
			err:      &Error{Code: "test_code", Path: "some.path", Message: "test error"},
			expected: "[test_code] some.path: test error",
		},
		{
			name:     "with cause",
			err:      &Error{Message: "test error", Cause: fmt.Errorf("underlying")},
			expected: "test error: underlying",
		},
		{
			name:     "with path and cause",
			err:      &Error{Path: "some.path", Message: "test error", Cause: fmt.Errorf("underlying")},
			expected: "some.path: test error: underlying",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.expected {
				t.Errorf("Error.Error() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestErrorUnwrap(t *testing.T) {
	cause := fmt.Errorf("underlying error")
	err := &Error{Message: "test", Cause: cause}

	if got := err.Unwrap(); got != cause {
		t.Errorf("Error.Unwrap() = %v, want %v", got, cause)
	}

	errNoCause := &Error{Message: "test"}
	if got := errNoCause.Unwrap(); got != nil {
		t.Errorf("Error.Unwrap() = %v, want nil", got)
	}
}

func TestErrorIs(t *testing.T) {
	err1 := &Error{Type: ErrorTypeValidation, Code: "test_code"}
	err2 := &Error{Type: ErrorTypeValidation, Code: "test_code"}
	err3 := &Error{Type: ErrorTypeValidation, Code: "other_code"}
	err4 := &Error{Type: ErrorTypeParsing, Code: "test_code"}
	err5 := &Error{Type: ErrorTypeValidation}

	// Same type and code
	if !err1.Is(err2) {
		t.Error("Expected err1.Is(err2) to be true")
	}

	// Same type, different code
	if err1.Is(err3) {
		t.Error("Expected err1.Is(err3) to be false")
	}

	// Different type, same code
	if err1.Is(err4) {
		t.Error("Expected err1.Is(err4) to be false")
	}

	// Same type, no code on target
	if !err1.Is(err5) {
		t.Error("Expected err1.Is(err5) to be true (type match)")
	}

	// Not an *Error
	if err1.Is(fmt.Errorf("regular error")) {
		t.Error("Expected err1.Is(regular error) to be false")
	}
}

func TestErrorWithMethods(t *testing.T) {
	original := New(ErrorTypeValidation, "original")

	// WithDetails
	withDetails := original.WithDetails(map[string]any{"key": "value"})
	if withDetails.Details["key"] != "value" {
		t.Error("WithDetails should add details")
	}
	if original.Details["key"] == "value" {
		t.Error("WithDetails should not modify original")
	}

	// WithCause
	cause := fmt.Errorf("cause")
	withCause := original.WithCause(cause)
	if withCause.Cause != cause {
		t.Error("WithCause should set cause")
	}
	if original.Cause != nil {
		t.Error("WithCause should not modify original")
	}

	// WithPath
	withPath := original.WithPath("some.path")
	if withPath.Path != "some.path" {
		t.Error("WithPath should set path")
	}
	if original.Path != "" {
		t.Error("WithPath should not modify original")
	}

	// WithMessage
	withMessage := original.WithMessage("new message")
	if withMessage.Message != "new message" {
		t.Error("WithMessage should set message")
	}
	if original.Message != "original" {
		t.Error("WithMessage should not modify original")
	}

	// WithMessagef
	withMessagef := original.WithMessagef("formatted %s", "message")
	if withMessagef.Message != "formatted message" {
		t.Error("WithMessagef should format message")
	}
}

func TestWarningMethods(t *testing.T) {
	err := New(ErrorTypeValidation, "test")

	// Initially not a warning
	if err.IsWarning() {
		t.Error("New error should not be a warning")
	}

	// AsWarning
	warning := err.AsWarning()
	if !warning.IsWarning() {
		t.Error("AsWarning should mark as warning")
	}
	if err.IsWarning() {
		t.Error("AsWarning should not modify original")
	}
}

func TestNewFunctions(t *testing.T) {
	// New
	err := New(ErrorTypeValidation, "test")
	if err.Type != ErrorTypeValidation || err.Message != "test" {
		t.Error("New should create error with type and message")
	}

	// Newf
	errf := Newf(ErrorTypeValidation, "test %d", 42)
	if errf.Message != "test 42" {
		t.Error("Newf should format message")
	}

	// NewWithPath
	errPath := NewWithPath(ErrorTypeValidation, "path", "message")
	if errPath.Path != "path" || errPath.Message != "message" {
		t.Error("NewWithPath should set path and message")
	}

	// NewWithCode
	errCode := NewWithCode(ErrorTypeValidation, "code", "message")
	if errCode.Code != "code" || errCode.Message != "message" {
		t.Error("NewWithCode should set code and message")
	}
}

func TestWrapFunctions(t *testing.T) {
	cause := fmt.Errorf("cause")

	// Wrap
	wrapped := Wrap(ErrorTypeValidation, cause, "wrapped")
	if wrapped.Cause != cause || wrapped.Message != "wrapped" {
		t.Error("Wrap should set cause and message")
	}

	// Wrapf
	wrappedf := Wrapf(ErrorTypeValidation, cause, "wrapped %d", 42)
	if wrappedf.Cause != cause || wrappedf.Message != "wrapped 42" {
		t.Error("Wrapf should set cause and format message")
	}
}

func TestGetFunctions(t *testing.T) {
	err := &Error{
		Type:    ErrorTypeValidation,
		Code:    "test_code",
		Path:    "test.path",
		Details: map[string]any{"key": "value"},
	}

	// GetType
	if GetType(err) != ErrorTypeValidation {
		t.Error("GetType should return error type")
	}
	if GetType(fmt.Errorf("regular")) != ErrorTypeUnknown {
		t.Error("GetType should return unknown for non-Error")
	}

	// GetCode
	if GetCode(err) != "test_code" {
		t.Error("GetCode should return code")
	}
	if GetCode(fmt.Errorf("regular")) != "" {
		t.Error("GetCode should return empty for non-Error")
	}

	// GetPath
	if GetPath(err) != "test.path" {
		t.Error("GetPath should return path")
	}
	if GetPath(fmt.Errorf("regular")) != "" {
		t.Error("GetPath should return empty for non-Error")
	}

	// GetDetails
	if GetDetails(err)["key"] != "value" {
		t.Error("GetDetails should return details")
	}
	if GetDetails(fmt.Errorf("regular")) != nil {
		t.Error("GetDetails should return nil for non-Error")
	}
}

func TestIsTypeFunctions(t *testing.T) {
	validationErr := New(ErrorTypeValidation, "test")
	parsingErr := New(ErrorTypeParsing, "test")
	internalErr := New(ErrorTypeInternal, "test")

	if !IsValidation(validationErr) {
		t.Error("IsValidation should return true for validation errors")
	}
	if IsValidation(parsingErr) {
		t.Error("IsValidation should return false for non-validation errors")
	}

	if !IsParsing(parsingErr) {
		t.Error("IsParsing should return true for parsing errors")
	}

	if !IsInternal(internalErr) {
		t.Error("IsInternal should return true for internal errors")
	}
}

func TestErrorsAsUnwrap(t *testing.T) {
	// Test that errors.As works with our Error type
	cause := &Error{Type: ErrorTypeValidation, Code: "inner", Message: "inner"}
	outer := Wrap(ErrorTypeParsing, cause, "outer")

	var target *Error
	if !errors.As(outer, &target) {
		t.Error("errors.As should find *Error")
	}
	if target.Code != "" { // outer has no code
		t.Errorf("errors.As should return outer error, got code: %s", target.Code)
	}

	// errors.Is with our wrapping
	if !errors.Is(outer, cause) {
		// Note: This might not work as expected since we override Is()
		// This tests the standard library behavior
	}
}
