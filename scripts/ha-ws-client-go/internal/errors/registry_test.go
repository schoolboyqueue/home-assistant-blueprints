package errors

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegistry_RegisterAndGet(t *testing.T) {
	r := NewRegistry()

	def := ErrorDefinition{
		Code:    "test_error",
		Type:    ErrorTypeNetwork,
		Message: "test error message",
	}

	r.Register(def)

	retrieved := r.Get("test_error")
	require.NotNil(t, retrieved)
	assert.Equal(t, "test_error", retrieved.Code)
	assert.Equal(t, ErrorTypeNetwork, retrieved.Type)
	assert.Equal(t, "test error message", retrieved.Message)
}

func TestRegistry_GetNotFound(t *testing.T) {
	r := NewRegistry()
	retrieved := r.Get("nonexistent")
	assert.Nil(t, retrieved)
}

func TestRegistry_Create(t *testing.T) {
	r := NewRegistry()
	r.Register(ErrorDefinition{
		Code:    "test_error",
		Type:    ErrorTypeNetwork,
		Message: "test error",
	})

	err := r.Create("test_error")
	require.NotNil(t, err)
	assert.Equal(t, ErrorTypeNetwork, err.Type)
	assert.Equal(t, "test_error", err.Code)
	assert.Equal(t, "test error", err.Message)
	assert.Nil(t, err.Cause)
	assert.NotNil(t, err.Details)
}

func TestRegistry_CreateNotFound(t *testing.T) {
	r := NewRegistry()
	err := r.Create("nonexistent")
	assert.Nil(t, err)
}

func TestRegistry_CreateWithMessage(t *testing.T) {
	r := NewRegistry()
	r.Register(ErrorDefinition{
		Code:    "test_error",
		Type:    ErrorTypeNotFound,
		Message: "default message",
	})

	err := r.CreateWithMessage("test_error", "custom message")
	require.NotNil(t, err)
	assert.Equal(t, "custom message", err.Message)
	assert.Equal(t, "test_error", err.Code)
}

func TestRegistry_CreateWithMessageNotFound(t *testing.T) {
	r := NewRegistry()
	err := r.CreateWithMessage("nonexistent", "message")
	assert.Nil(t, err)
}

func TestRegistry_CreateWithCause(t *testing.T) {
	r := NewRegistry()
	r.Register(ErrorDefinition{
		Code:    "test_error",
		Type:    ErrorTypeNetwork,
		Message: "test error",
	})

	cause := errors.New("underlying cause")
	err := r.CreateWithCause("test_error", cause)
	require.NotNil(t, err)
	assert.Equal(t, cause, err.Cause)
	assert.True(t, errors.Is(err, cause))
}

func TestRegistry_CreateWithCauseNotFound(t *testing.T) {
	r := NewRegistry()
	err := r.CreateWithCause("nonexistent", errors.New("cause"))
	assert.Nil(t, err)
}

func TestRegistry_List(t *testing.T) {
	r := NewRegistry()
	r.Register(ErrorDefinition{Code: "error1", Type: ErrorTypeNetwork})
	r.Register(ErrorDefinition{Code: "error2", Type: ErrorTypeNotFound})

	list := r.List()
	assert.Len(t, list, 2)

	codes := make(map[string]bool)
	for _, def := range list {
		codes[def.Code] = true
	}
	assert.True(t, codes["error1"])
	assert.True(t, codes["error2"])
}

func TestRegistry_Overwrite(t *testing.T) {
	r := NewRegistry()

	r.Register(ErrorDefinition{
		Code:    "test_error",
		Type:    ErrorTypeNetwork,
		Message: "original",
	})

	r.Register(ErrorDefinition{
		Code:    "test_error",
		Type:    ErrorTypeNotFound,
		Message: "overwritten",
	})

	retrieved := r.Get("test_error")
	require.NotNil(t, retrieved)
	assert.Equal(t, ErrorTypeNotFound, retrieved.Type)
	assert.Equal(t, "overwritten", retrieved.Message)
}

func TestDefaultRegistry_CommonErrors(t *testing.T) {
	// Test that common errors are registered in the default registry
	codes := []string{
		CodeConnectionClosed,
		CodeConnectionFailed,
		CodeConnectionTimeout,
		CodeAuthFailed,
		CodeEntityNotFound,
		CodeEntityInvalidID,
		CodeSubscriptionFailed,
		CodeSubscriptionTimeout,
		CodeRequestCanceled,
		CodeRequestTimeout,
		CodeMissingArgument,
		CodeInvalidArgument,
		CodeInvalidPattern,
		CodeInvalidJSON,
		CodeAPIError,
		CodeNoDataFound,
	}

	for _, code := range codes {
		t.Run(code, func(t *testing.T) {
			def := DefaultRegistry.Get(code)
			require.NotNil(t, def, "Error code %s should be registered", code)
		})
	}
}

func TestDefaultRegistryFunctions(t *testing.T) {
	// Test that the package-level functions work with DefaultRegistry
	t.Run("Create", func(t *testing.T) {
		err := Create(CodeConnectionClosed)
		require.NotNil(t, err)
		assert.Equal(t, CodeConnectionClosed, err.Code)
	})

	t.Run("CreateWithMessage", func(t *testing.T) {
		err := CreateWithMessage(CodeConnectionClosed, "custom message")
		require.NotNil(t, err)
		assert.Equal(t, "custom message", err.Message)
	})

	t.Run("CreateWithCause", func(t *testing.T) {
		cause := errors.New("cause")
		err := CreateWithCause(CodeConnectionClosed, cause)
		require.NotNil(t, err)
		assert.Equal(t, cause, err.Cause)
	})
}

func TestConvenienceFactories(t *testing.T) {
	t.Run("ErrConnectionClosed", func(t *testing.T) {
		err := ErrConnectionClosed()
		require.NotNil(t, err)
		assert.Equal(t, CodeConnectionClosed, err.Code)
		assert.Equal(t, ErrorTypeNetwork, err.Type)
	})

	t.Run("ErrConnectionFailed", func(t *testing.T) {
		cause := errors.New("timeout")
		err := ErrConnectionFailed(cause)
		require.NotNil(t, err)
		assert.Equal(t, CodeConnectionFailed, err.Code)
		assert.Equal(t, cause, err.Cause)
	})

	t.Run("ErrAuthFailed", func(t *testing.T) {
		err := ErrAuthFailed("invalid token")
		require.NotNil(t, err)
		assert.Equal(t, CodeAuthFailed, err.Code)
		assert.Equal(t, "invalid token", err.Message)
	})

	t.Run("ErrEntityNotFound", func(t *testing.T) {
		err := ErrEntityNotFound("light.living_room")
		require.NotNil(t, err)
		assert.Equal(t, CodeEntityNotFound, err.Code)
		assert.Contains(t, err.Message, "light.living_room")
		assert.Equal(t, "light.living_room", err.Details["entity_id"])
	})

	t.Run("ErrEntityInvalidID", func(t *testing.T) {
		err := ErrEntityInvalidID("invalid")
		require.NotNil(t, err)
		assert.Equal(t, CodeEntityInvalidID, err.Code)
		assert.Contains(t, err.Message, "invalid")
	})

	t.Run("ErrSubscriptionFailed with cause", func(t *testing.T) {
		cause := errors.New("failed")
		err := ErrSubscriptionFailed(cause)
		require.NotNil(t, err)
		assert.Equal(t, CodeSubscriptionFailed, err.Code)
		assert.Equal(t, cause, err.Cause)
	})

	t.Run("ErrSubscriptionFailed without cause", func(t *testing.T) {
		err := ErrSubscriptionFailed(nil)
		require.NotNil(t, err)
		assert.Equal(t, CodeSubscriptionFailed, err.Code)
		assert.Nil(t, err.Cause)
	})

	t.Run("ErrSubscriptionTimeout", func(t *testing.T) {
		err := ErrSubscriptionTimeout()
		require.NotNil(t, err)
		assert.Equal(t, CodeSubscriptionTimeout, err.Code)
		assert.Equal(t, ErrorTypeTimeout, err.Type)
	})

	t.Run("ErrRequestCanceled", func(t *testing.T) {
		cause := errors.New("context canceled")
		err := ErrRequestCanceled(cause)
		require.NotNil(t, err)
		assert.Equal(t, CodeRequestCanceled, err.Code)
		assert.Equal(t, ErrorTypeCanceled, err.Type)
	})

	t.Run("ErrRequestTimeout", func(t *testing.T) {
		err := ErrRequestTimeout()
		require.NotNil(t, err)
		assert.Equal(t, CodeRequestTimeout, err.Code)
		assert.Equal(t, ErrorTypeTimeout, err.Type)
	})

	t.Run("ErrMissingArgument", func(t *testing.T) {
		err := ErrMissingArgument("Usage: command <arg>")
		require.NotNil(t, err)
		assert.Equal(t, CodeMissingArgument, err.Code)
		assert.Contains(t, err.Message, "Usage: command <arg>")
	})

	t.Run("ErrInvalidArgument", func(t *testing.T) {
		err := ErrInvalidArgument("value must be positive")
		require.NotNil(t, err)
		assert.Equal(t, CodeInvalidArgument, err.Code)
		assert.Equal(t, "value must be positive", err.Message)
	})

	t.Run("ErrInvalidPattern", func(t *testing.T) {
		cause := errors.New("invalid regex")
		err := ErrInvalidPattern(cause)
		require.NotNil(t, err)
		assert.Equal(t, CodeInvalidPattern, err.Code)
		assert.Equal(t, cause, err.Cause)
	})

	t.Run("ErrInvalidJSON", func(t *testing.T) {
		cause := errors.New("unexpected token")
		err := ErrInvalidJSON(cause)
		require.NotNil(t, err)
		assert.Equal(t, CodeInvalidJSON, err.Code)
		assert.Equal(t, cause, err.Cause)
	})

	t.Run("ErrMessageMarshalFailed", func(t *testing.T) {
		cause := errors.New("marshal error")
		err := ErrMessageMarshalFailed(cause)
		require.NotNil(t, err)
		assert.Equal(t, CodeMessageMarshalError, err.Code)
	})

	t.Run("ErrMessageSendFailed", func(t *testing.T) {
		cause := errors.New("write error")
		err := ErrMessageSendFailed(cause)
		require.NotNil(t, err)
		assert.Equal(t, CodeMessageSendFailed, err.Code)
	})

	t.Run("ErrAPIError", func(t *testing.T) {
		err := ErrAPIError("not_found", "Entity not found")
		require.NotNil(t, err)
		assert.Equal(t, "not_found", err.Code)
		assert.Equal(t, "Entity not found", err.Message)
		assert.Equal(t, ErrorTypeAPI, err.Type)
	})

	t.Run("ErrNoTracesFound", func(t *testing.T) {
		err := ErrNoTracesFound("my_automation")
		require.NotNil(t, err)
		assert.Equal(t, CodeNoTracesFound, err.Code)
		assert.Contains(t, err.Message, "my_automation")
		assert.Equal(t, "my_automation", err.Details["automation_id"])
	})

	t.Run("ErrTemplateTimeout", func(t *testing.T) {
		err := ErrTemplateTimeout()
		require.NotNil(t, err)
		assert.Equal(t, CodeTemplateError, err.Code)
		assert.Contains(t, err.Message, "timeout")
	})

	t.Run("ErrNoDataFound", func(t *testing.T) {
		err := ErrNoDataFound("no statistics")
		require.NotNil(t, err)
		assert.Equal(t, CodeNoDataFound, err.Code)
		assert.Equal(t, "no statistics", err.Message)
	})
}

func TestRegistry_ConcurrentAccess(_ *testing.T) {
	r := NewRegistry()

	// Register errors concurrently
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(n int) {
			r.Register(ErrorDefinition{
				Code:    string(rune('a' + n)),
				Type:    ErrorTypeNetwork,
				Message: "test",
			})
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Read concurrently
	for i := 0; i < 10; i++ {
		go func(n int) {
			_ = r.Get(string(rune('a' + n)))
			_ = r.List()
			done <- true
		}(i)
	}

	for i := 0; i < 10; i++ {
		<-done
	}
}
