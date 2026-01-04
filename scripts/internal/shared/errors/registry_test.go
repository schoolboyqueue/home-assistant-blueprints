package errors

import (
	"fmt"
	"testing"
)

func TestNewRegistry(t *testing.T) {
	r := NewRegistry()
	if r == nil {
		t.Fatal("NewRegistry should not return nil")
	}
	if r.definitions == nil {
		t.Error("NewRegistry should initialize definitions map")
	}
}

func TestRegistryRegisterAndGet(t *testing.T) {
	r := NewRegistry()

	def := ErrorDefinition{
		Code:    "test_error",
		Type:    ErrorTypeValidation,
		Message: "test message",
	}

	r.Register(def)

	got := r.Get("test_error")
	if got == nil {
		t.Fatal("Get should return registered definition")
	}
	if got.Code != def.Code || got.Type != def.Type || got.Message != def.Message {
		t.Error("Get should return matching definition")
	}

	// Test not found
	if r.Get("nonexistent") != nil {
		t.Error("Get should return nil for nonexistent code")
	}
}

func TestRegistryOverwrite(t *testing.T) {
	r := NewRegistry()

	r.Register(ErrorDefinition{
		Code:    "test",
		Type:    ErrorTypeValidation,
		Message: "first",
	})

	r.Register(ErrorDefinition{
		Code:    "test",
		Type:    ErrorTypeParsing,
		Message: "second",
	})

	got := r.Get("test")
	if got.Message != "second" || got.Type != ErrorTypeParsing {
		t.Error("Register should overwrite existing definition")
	}
}

func TestRegistryCreate(t *testing.T) {
	r := NewRegistry()
	r.Register(ErrorDefinition{
		Code:    "test_error",
		Type:    ErrorTypeValidation,
		Message: "test message",
	})

	err := r.Create("test_error")
	if err == nil {
		t.Fatal("Create should return error")
	}
	if err.Type != ErrorTypeValidation {
		t.Error("Create should set type from definition")
	}
	if err.Code != "test_error" {
		t.Error("Create should set code from definition")
	}
	if err.Message != "test message" {
		t.Error("Create should set message from definition")
	}
	if err.Details == nil {
		t.Error("Create should initialize Details map")
	}

	// Test not found
	if r.Create("nonexistent") != nil {
		t.Error("Create should return nil for nonexistent code")
	}
}

func TestRegistryCreateWithMessage(t *testing.T) {
	r := NewRegistry()
	r.Register(ErrorDefinition{
		Code:    "test_error",
		Type:    ErrorTypeValidation,
		Message: "default message",
	})

	err := r.CreateWithMessage("test_error", "custom message")
	if err == nil {
		t.Fatal("CreateWithMessage should return error")
	}
	if err.Message != "custom message" {
		t.Error("CreateWithMessage should use custom message")
	}
	if err.Type != ErrorTypeValidation {
		t.Error("CreateWithMessage should preserve type")
	}

	// Test not found
	if r.CreateWithMessage("nonexistent", "message") != nil {
		t.Error("CreateWithMessage should return nil for nonexistent code")
	}
}

func TestRegistryCreateWithCause(t *testing.T) {
	r := NewRegistry()
	r.Register(ErrorDefinition{
		Code:    "test_error",
		Type:    ErrorTypeValidation,
		Message: "test message",
	})

	cause := fmt.Errorf("underlying error")
	err := r.CreateWithCause("test_error", cause)
	if err == nil {
		t.Fatal("CreateWithCause should return error")
	}
	if err.Cause != cause {
		t.Error("CreateWithCause should set cause")
	}
	if err.Message != "test message" {
		t.Error("CreateWithCause should use definition message")
	}

	// Test not found
	if r.CreateWithCause("nonexistent", cause) != nil {
		t.Error("CreateWithCause should return nil for nonexistent code")
	}
}

func TestRegistryCreateWithPath(t *testing.T) {
	r := NewRegistry()
	r.Register(ErrorDefinition{
		Code:    "test_error",
		Type:    ErrorTypeValidation,
		Message: "test message",
	})

	err := r.CreateWithPath("test_error", "some.path")
	if err == nil {
		t.Fatal("CreateWithPath should return error")
	}
	if err.Path != "some.path" {
		t.Error("CreateWithPath should set path")
	}
	if err.Message != "test message" {
		t.Error("CreateWithPath should use definition message")
	}

	// Test not found
	if r.CreateWithPath("nonexistent", "path") != nil {
		t.Error("CreateWithPath should return nil for nonexistent code")
	}
}

func TestRegistryList(t *testing.T) {
	r := NewRegistry()
	r.Register(ErrorDefinition{Code: "error1", Type: ErrorTypeValidation, Message: "msg1"})
	r.Register(ErrorDefinition{Code: "error2", Type: ErrorTypeParsing, Message: "msg2"})

	list := r.List()
	if len(list) != 2 {
		t.Errorf("List should return %d definitions, got %d", 2, len(list))
	}

	// Check both definitions are present
	codes := make(map[string]bool)
	for _, def := range list {
		codes[def.Code] = true
	}
	if !codes["error1"] || !codes["error2"] {
		t.Error("List should return all registered definitions")
	}
}

func TestDefaultRegistry(t *testing.T) {
	// DefaultRegistry should have common errors registered via init()
	if DefaultRegistry.Get(CodeMissingArgument) == nil {
		t.Error("DefaultRegistry should have CodeMissingArgument registered")
	}
	if DefaultRegistry.Get(CodeInvalidArgument) == nil {
		t.Error("DefaultRegistry should have CodeInvalidArgument registered")
	}
	if DefaultRegistry.Get(CodeInvalidJSON) == nil {
		t.Error("DefaultRegistry should have CodeInvalidJSON registered")
	}
	if DefaultRegistry.Get(CodeFileNotFound) == nil {
		t.Error("DefaultRegistry should have CodeFileNotFound registered")
	}
}

func TestDefaultRegistryFunctions(t *testing.T) {
	// Test the convenience functions that use DefaultRegistry
	err := Create(CodeMissingArgument)
	if err == nil || err.Code != CodeMissingArgument {
		t.Error("Create should use DefaultRegistry")
	}

	err = CreateWithMessage(CodeInvalidArgument, "custom")
	if err == nil || err.Message != "custom" {
		t.Error("CreateWithMessage should use DefaultRegistry")
	}

	cause := fmt.Errorf("cause")
	err = CreateWithCause(CodeInvalidJSON, cause)
	if err == nil || err.Cause != cause {
		t.Error("CreateWithCause should use DefaultRegistry")
	}

	err = CreateWithPath(CodeFileNotFound, "path")
	if err == nil || err.Path != "path" {
		t.Error("CreateWithPath should use DefaultRegistry")
	}
}

func TestConvenienceFactoryFunctions(t *testing.T) {
	// ErrMissingArgument
	err := ErrMissingArgument("usage text")
	if err.Code != CodeMissingArgument {
		t.Error("ErrMissingArgument should use correct code")
	}
	if err.Message != "missing argument: usage text" {
		t.Errorf("ErrMissingArgument message = %q", err.Message)
	}

	// ErrInvalidArgument
	err = ErrInvalidArgument("bad value")
	if err.Code != CodeInvalidArgument {
		t.Error("ErrInvalidArgument should use correct code")
	}
	if err.Message != "bad value" {
		t.Error("ErrInvalidArgument should use provided message")
	}

	// ErrInvalidJSON
	cause := fmt.Errorf("json error")
	err = ErrInvalidJSON(cause)
	if err.Code != CodeInvalidJSON {
		t.Error("ErrInvalidJSON should use correct code")
	}
	if err.Cause != cause {
		t.Error("ErrInvalidJSON should wrap cause")
	}

	// ErrFileNotFound
	err = ErrFileNotFound("/path/to/file")
	if err.Code != CodeFileNotFound {
		t.Error("ErrFileNotFound should use correct code")
	}
	if err.Path != "/path/to/file" {
		t.Error("ErrFileNotFound should set path")
	}

	// ErrFileReadError
	cause = fmt.Errorf("io error")
	err = ErrFileReadError("/path/to/file", cause)
	if err.Code != CodeFileReadError {
		t.Error("ErrFileReadError should use correct code")
	}
	if err.Path != "/path/to/file" {
		t.Error("ErrFileReadError should set path")
	}
	if err.Cause != cause {
		t.Error("ErrFileReadError should wrap cause")
	}
}

func TestRegistryConcurrency(t *testing.T) {
	r := NewRegistry()
	done := make(chan bool)

	// Writer goroutine
	go func() {
		for i := range 100 {
			r.Register(ErrorDefinition{
				Code:    fmt.Sprintf("error_%d", i),
				Type:    ErrorTypeValidation,
				Message: "test",
			})
		}
		done <- true
	}()

	// Reader goroutine
	go func() {
		for i := range 100 {
			_ = r.Get(fmt.Sprintf("error_%d", i))
			_ = r.List()
		}
		done <- true
	}()

	// Wait for both
	<-done
	<-done
}
