package client

import (
	"context"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/home-assistant-blueprints/testfixtures"

	"github.com/home-assistant-blueprints/ha-ws-client-go/internal/types"
)

// testServer creates a test WebSocket server that handles messages.
// This is a wrapper around testfixtures.TestServer for backward compatibility.
func testServer(t *testing.T, handler func(*websocket.Conn)) *httptest.Server {
	t.Helper()
	return testfixtures.TestServer(t, handler)
}

// dialServer connects to a test server and returns a Client.
// This wraps testfixtures.DialServer and creates a Client.
func dialServer(t *testing.T, server *httptest.Server) *Client {
	t.Helper()
	conn := testfixtures.DialServer(t, server)
	return New(conn)
}

// dialServerWithContext connects to a test server and returns a Client with context.
func dialServerWithContext(ctx context.Context, t *testing.T, server *httptest.Server) *Client {
	t.Helper()
	conn := testfixtures.DialServer(t, server)
	return NewWithContext(ctx, conn)
}

func TestHAClientError_Error(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		err      HAClientError
		expected string
	}{
		{
			name:     "with code",
			err:      HAClientError{Code: "not_found", Message: "Entity not found"},
			expected: "[not_found] Entity not found",
		},
		{
			name:     "without code",
			err:      HAClientError{Message: "Something went wrong"},
			expected: "Something went wrong",
		},
		{
			name:     "empty",
			err:      HAClientError{},
			expected: "",
		},
		{
			name:     "code only",
			err:      HAClientError{Code: "ERR001"},
			expected: "[ERR001] ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expected, tt.err.Error())
		})
	}
}

func TestHAClientError_Implements(t *testing.T) {
	t.Parallel()

	var err error = &HAClientError{Message: "test"}
	assert.Error(t, err)

	var clientErr *HAClientError
	assert.True(t, errors.As(err, &clientErr))
	assert.Equal(t, "test", clientErr.Message)
}

func TestNew(t *testing.T) {
	server := testServer(t, func(_ *websocket.Conn) {
		// Keep connection open briefly
		time.Sleep(100 * time.Millisecond)
	})
	defer server.Close()

	client := dialServer(t, server)
	defer client.Close()

	assert.NotNil(t, client)
	assert.NotNil(t, client.conn)
	assert.NotNil(t, client.pending)
	assert.NotNil(t, client.subscriptions)
	assert.NotNil(t, client.done)
}

func TestClient_NextID(t *testing.T) {
	server := testServer(t, func(_ *websocket.Conn) {
		time.Sleep(100 * time.Millisecond)
	})
	defer server.Close()

	client := dialServer(t, server)
	defer client.Close()

	// IDs should be sequential starting from 1
	assert.Equal(t, 1, client.NextID())
	assert.Equal(t, 2, client.NextID())
	assert.Equal(t, 3, client.NextID())
}

func TestClient_NextID_Concurrent(t *testing.T) {
	server := testServer(t, func(_ *websocket.Conn) {
		time.Sleep(200 * time.Millisecond)
	})
	defer server.Close()

	client := dialServer(t, server)
	defer client.Close()

	var wg sync.WaitGroup
	ids := make(chan int, 100)

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ids <- client.NextID()
		}()
	}

	wg.Wait()
	close(ids)

	// Collect all IDs
	seen := make(map[int]bool)
	for id := range ids {
		assert.False(t, seen[id], "duplicate ID: %d", id)
		seen[id] = true
	}
	assert.Len(t, seen, 100)
}

func TestClient_SendMessage_Success(t *testing.T) {
	server := testServer(t, func(conn *websocket.Conn) {
		// Read the request
		_, data, err := conn.ReadMessage()
		require.NoError(t, err)

		var req map[string]any
		err = json.Unmarshal(data, &req)
		require.NoError(t, err)

		// Send success response
		success := true
		reqID, _ := req["id"].(float64) //nolint:errcheck // test knows the format
		resp := types.HAMessage{
			ID:      int(reqID),
			Type:    "result",
			Success: &success,
			Result:  "pong",
		}
		err = conn.WriteJSON(resp)
		require.NoError(t, err)
		// Allow client time to read before handler returns and closes connection
		time.Sleep(100 * time.Millisecond)
	})
	defer server.Close()

	client := dialServer(t, server)
	defer client.Close()

	resp, err := client.SendMessage("ping", nil)
	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "pong", resp.Result)
}

func TestClient_SendMessage_WithData(t *testing.T) {
	server := testServer(t, func(conn *websocket.Conn) {
		_, data, err := conn.ReadMessage()
		require.NoError(t, err)

		var req map[string]any
		err = json.Unmarshal(data, &req)
		require.NoError(t, err)

		// Verify the data was sent
		assert.Equal(t, "call_service", req["type"])
		assert.Equal(t, "light", req["domain"])
		assert.Equal(t, "turn_on", req["service"])

		success := true
		reqID, _ := req["id"].(float64) //nolint:errcheck // test knows the format
		resp := types.HAMessage{
			ID:      int(reqID),
			Type:    "result",
			Success: &success,
		}
		err = conn.WriteJSON(resp)
		require.NoError(t, err)
		// Allow client time to read before handler returns and closes connection
		time.Sleep(100 * time.Millisecond)
	})
	defer server.Close()

	client := dialServer(t, server)
	defer client.Close()

	resp, err := client.SendMessage("call_service", map[string]any{
		"domain":  "light",
		"service": "turn_on",
	})
	require.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestClient_SendMessage_Error(t *testing.T) {
	server := testServer(t, func(conn *websocket.Conn) {
		_, data, err := conn.ReadMessage()
		require.NoError(t, err)

		var req map[string]any
		err = json.Unmarshal(data, &req)
		require.NoError(t, err)

		// Send error response
		success := false
		reqID, _ := req["id"].(float64) //nolint:errcheck // test knows the format
		resp := types.HAMessage{
			ID:      int(reqID),
			Type:    "result",
			Success: &success,
			Error: &types.HAError{
				Code:    "not_found",
				Message: "Entity not found",
			},
		}
		err = conn.WriteJSON(resp)
		require.NoError(t, err)
		// Allow client time to read before handler returns and closes connection
		time.Sleep(100 * time.Millisecond)
	})
	defer server.Close()

	client := dialServer(t, server)
	defer client.Close()

	resp, err := client.SendMessage("get_states", nil)
	assert.Nil(t, resp)
	require.Error(t, err)

	var clientErr *HAClientError
	require.True(t, errors.As(err, &clientErr))
	assert.Equal(t, "not_found", clientErr.Code)
	assert.Equal(t, "Entity not found", clientErr.Message)
}

func TestClient_SendMessage_ErrorWithoutDetails(t *testing.T) {
	server := testServer(t, func(conn *websocket.Conn) {
		_, data, err := conn.ReadMessage()
		require.NoError(t, err)

		var req map[string]any
		err = json.Unmarshal(data, &req)
		require.NoError(t, err)

		// Send error response without Error field
		success := false
		reqID, _ := req["id"].(float64) //nolint:errcheck // test knows the format
		resp := types.HAMessage{
			ID:      int(reqID),
			Type:    "result",
			Success: &success,
		}
		err = conn.WriteJSON(resp)
		require.NoError(t, err)
		// Allow client time to read before handler returns and closes connection
		time.Sleep(100 * time.Millisecond)
	})
	defer server.Close()

	client := dialServer(t, server)
	defer client.Close()

	resp, err := client.SendMessage("test", nil)
	assert.Nil(t, resp)
	require.Error(t, err)

	var clientErr *HAClientError
	require.True(t, errors.As(err, &clientErr))
	assert.Equal(t, "unknown error", clientErr.Message)
}

func TestClient_SendMessage_ConnectionClosed(t *testing.T) {
	server := testServer(t, func(conn *websocket.Conn) {
		// Read the request then close
		_, _, _ = conn.ReadMessage() //nolint:errcheck // intentionally ignoring error
		conn.Close()
	})
	defer server.Close()

	client := dialServer(t, server)
	defer client.Close()

	// Give the server time to close
	time.Sleep(50 * time.Millisecond)

	resp, err := client.SendMessage("ping", nil)
	assert.Nil(t, resp)
	require.Error(t, err)
}

func TestSendMessageTyped(t *testing.T) {
	server := testServer(t, func(conn *websocket.Conn) {
		_, data, err := conn.ReadMessage()
		require.NoError(t, err)

		var req map[string]any
		err = json.Unmarshal(data, &req)
		require.NoError(t, err)

		// Send typed response
		success := true
		states := []types.HAState{
			{EntityID: "light.kitchen", State: "on"},
			{EntityID: "light.bedroom", State: "off"},
		}
		reqID, _ := req["id"].(float64) //nolint:errcheck // test knows the format
		resp := types.HAMessage{
			ID:      int(reqID),
			Type:    "result",
			Success: &success,
			Result:  states,
		}
		err = conn.WriteJSON(resp)
		require.NoError(t, err)
		// Allow client time to read before handler returns and closes connection
		time.Sleep(100 * time.Millisecond)
	})
	defer server.Close()

	client := dialServer(t, server)
	defer client.Close()

	states, err := SendMessageTyped[[]types.HAState](client, "get_states", nil)
	require.NoError(t, err)
	require.Len(t, states, 2)
	assert.Equal(t, "light.kitchen", states[0].EntityID)
	assert.Equal(t, "on", states[0].State)
}

func TestSendMessageTyped_Error(t *testing.T) {
	server := testServer(t, func(conn *websocket.Conn) {
		_, data, err := conn.ReadMessage()
		require.NoError(t, err)

		var req map[string]any
		err = json.Unmarshal(data, &req)
		require.NoError(t, err)

		success := false
		reqID, _ := req["id"].(float64) //nolint:errcheck // test knows the format
		resp := types.HAMessage{
			ID:      int(reqID),
			Type:    "result",
			Success: &success,
			Error:   &types.HAError{Code: "test", Message: "error"},
		}
		err = conn.WriteJSON(resp)
		require.NoError(t, err)
		// Allow client time to read before handler returns and closes connection
		time.Sleep(100 * time.Millisecond)
	})
	defer server.Close()

	client := dialServer(t, server)
	defer client.Close()

	states, err := SendMessageTyped[[]types.HAState](client, "get_states", nil)
	require.Error(t, err)
	assert.Nil(t, states)
}

func TestClient_HandleMessage_Result(t *testing.T) {
	server := testServer(t, func(conn *websocket.Conn) {
		// Read request
		_, data, err := conn.ReadMessage()
		require.NoError(t, err)

		var req map[string]any
		err = json.Unmarshal(data, &req)
		require.NoError(t, err)

		// Send pong response
		resp := map[string]any{
			"id":   req["id"],
			"type": "pong",
		}
		err = conn.WriteJSON(resp)
		require.NoError(t, err)
		// Allow client time to read before handler returns and closes connection
		time.Sleep(100 * time.Millisecond)
	})
	defer server.Close()

	client := dialServer(t, server)
	defer client.Close()

	resp, err := client.SendMessage("ping", nil)
	require.NoError(t, err)
	assert.Equal(t, "pong", resp.Type)
}

func TestClient_HandleMessage_Event(t *testing.T) {
	eventReceived := make(chan map[string]any, 1)

	server := testServer(t, func(conn *websocket.Conn) {
		// Read subscription request
		_, data, err := conn.ReadMessage()
		require.NoError(t, err)

		var req map[string]any
		err = json.Unmarshal(data, &req)
		require.NoError(t, err)

		reqID, _ := req["id"].(float64) //nolint:errcheck // test knows the format
		id := int(reqID)

		// Send subscription confirmation
		success := true
		resp := types.HAMessage{
			ID:      id,
			Type:    "result",
			Success: &success,
		}
		err = conn.WriteJSON(resp)
		require.NoError(t, err)

		// Send event
		time.Sleep(50 * time.Millisecond)
		event := types.HAMessage{
			ID:   id,
			Type: "event",
			Event: &types.HAEvent{
				Variables: map[string]any{
					"trigger": map[string]any{
						"platform":  "state",
						"entity_id": "binary_sensor.motion",
					},
				},
			},
		}
		err = conn.WriteJSON(event)
		require.NoError(t, err)

		// Keep connection open
		time.Sleep(200 * time.Millisecond)
	})
	defer server.Close()

	client := dialServer(t, server)
	defer client.Close()

	trigger := map[string]any{
		"platform":  "state",
		"entity_id": "binary_sensor.motion",
	}

	_, cleanup, err := client.SubscribeToTrigger(trigger, func(vars map[string]any) {
		eventReceived <- vars
	}, 5*time.Second)
	require.NoError(t, err)
	defer cleanup()

	select {
	case vars := <-eventReceived:
		assert.NotNil(t, vars["trigger"])
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for event")
	}
}

func TestClient_SubscribeToTemplate(t *testing.T) {
	resultReceived := make(chan string, 1)

	server := testServer(t, func(conn *websocket.Conn) {
		// Read subscription request
		_, data, err := conn.ReadMessage()
		require.NoError(t, err)

		var req map[string]any
		err = json.Unmarshal(data, &req)
		require.NoError(t, err)

		reqID, _ := req["id"].(float64) //nolint:errcheck // test knows the format
		id := int(reqID)
		assert.Equal(t, "render_template", req["type"])
		assert.Equal(t, "{{ states('sensor.temp') }}", req["template"])

		// Send subscription confirmation
		success := true
		resp := types.HAMessage{
			ID:      id,
			Type:    "result",
			Success: &success,
		}
		err = conn.WriteJSON(resp)
		require.NoError(t, err)

		// Send template result as event
		time.Sleep(50 * time.Millisecond)
		event := types.HAMessage{
			ID:   id,
			Type: "event",
			Event: &types.HAEvent{
				Result: "23.5",
			},
		}
		err = conn.WriteJSON(event)
		require.NoError(t, err)

		time.Sleep(200 * time.Millisecond)
	})
	defer server.Close()

	client := dialServer(t, server)
	defer client.Close()

	_, cleanup, err := client.SubscribeToTemplate("{{ states('sensor.temp') }}", func(result string) {
		resultReceived <- result
	}, 5*time.Second)
	require.NoError(t, err)
	defer cleanup()

	select {
	case result := <-resultReceived:
		assert.Equal(t, "23.5", result)
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for template result")
	}
}

func TestClient_SubscribeToTrigger_Timeout(t *testing.T) {
	server := testServer(t, func(conn *websocket.Conn) {
		// Read but don't respond
		_, _, _ = conn.ReadMessage() //nolint:errcheck // intentionally ignoring error
		time.Sleep(10 * time.Second)
	})
	defer server.Close()

	client := dialServer(t, server)
	defer client.Close()

	_, _, err := client.SubscribeToTrigger(map[string]any{}, func(map[string]any) {}, 100*time.Millisecond)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "timeout")
}

func TestClient_SubscribeToTrigger_Error(t *testing.T) {
	server := testServer(t, func(conn *websocket.Conn) {
		_, data, err := conn.ReadMessage()
		require.NoError(t, err)

		var req map[string]any
		err = json.Unmarshal(data, &req)
		require.NoError(t, err)

		// Send error response
		success := false
		reqID, _ := req["id"].(float64) //nolint:errcheck // test knows the format
		resp := types.HAMessage{
			ID:      int(reqID),
			Type:    "result",
			Success: &success,
			Error:   &types.HAError{Message: "invalid trigger"},
		}
		err = conn.WriteJSON(resp)
		require.NoError(t, err)
		// Allow client time to read before handler returns and closes connection
		time.Sleep(100 * time.Millisecond)
	})
	defer server.Close()

	client := dialServer(t, server)
	defer client.Close()

	_, _, err := client.SubscribeToTrigger(map[string]any{}, func(map[string]any) {}, time.Second)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid trigger")
}

func TestClient_SubscribeToTemplate_Error(t *testing.T) {
	server := testServer(t, func(conn *websocket.Conn) {
		_, data, err := conn.ReadMessage()
		require.NoError(t, err)

		var req map[string]any
		err = json.Unmarshal(data, &req)
		require.NoError(t, err)

		success := false
		reqID, _ := req["id"].(float64) //nolint:errcheck // test knows the format
		resp := types.HAMessage{
			ID:      int(reqID),
			Type:    "result",
			Success: &success,
			Error:   &types.HAError{Message: "template error"},
		}
		err = conn.WriteJSON(resp)
		require.NoError(t, err)
		// Allow client time to read before handler returns and closes connection
		time.Sleep(100 * time.Millisecond)
	})
	defer server.Close()

	client := dialServer(t, server)
	defer client.Close()

	_, _, err := client.SubscribeToTemplate("{{ invalid }}", func(string) {}, time.Second)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "template error")
}

func TestClient_Close(t *testing.T) {
	server := testServer(t, func(_ *websocket.Conn) {
		time.Sleep(500 * time.Millisecond)
	})
	defer server.Close()

	client := dialServer(t, server)

	err := client.Close()
	require.NoError(t, err)

	// Done channel should be closed after readLoop exits
	select {
	case <-client.Done():
		// Expected
	case <-time.After(1 * time.Second):
		t.Fatal("done channel not closed")
	}
}

func TestClient_Done(t *testing.T) {
	server := testServer(t, func(conn *websocket.Conn) {
		// Close immediately
		conn.Close()
	})
	defer server.Close()

	client := dialServer(t, server)

	select {
	case <-client.Done():
		// Connection closed, done signaled
	case <-time.After(1 * time.Second):
		t.Fatal("done channel not closed after connection closed")
	}
}

func TestClient_ReadLoop_InvalidJSON(t *testing.T) {
	server := testServer(t, func(conn *websocket.Conn) {
		// Send invalid JSON
		err := conn.WriteMessage(websocket.TextMessage, []byte("not json"))
		require.NoError(t, err)

		// Send valid message after
		success := true
		resp := types.HAMessage{
			ID:      1,
			Type:    "result",
			Success: &success,
			Result:  "ok",
		}
		err = conn.WriteJSON(resp)
		require.NoError(t, err)

		time.Sleep(200 * time.Millisecond)
	})
	defer server.Close()

	client := dialServer(t, server)
	defer client.Close()

	// The client should continue working after invalid JSON
	resp, err := client.SendMessage("test", nil)
	require.NoError(t, err)
	assert.Equal(t, "ok", resp.Result)
}

func TestClient_Cleanup_OnError(t *testing.T) {
	server := testServer(t, func(conn *websocket.Conn) {
		// Read request
		_, _, _ = conn.ReadMessage() //nolint:errcheck // intentionally ignoring error
		// Close abruptly
		conn.Close()
	})
	defer server.Close()

	client := dialServer(t, server)
	defer client.Close()

	// This should fail because connection is closed
	_, err := client.SendMessage("test", nil)
	require.Error(t, err)

	// Pending map should be cleaned up
	client.pendingMu.RLock()
	assert.Empty(t, client.pending)
	client.pendingMu.RUnlock()
}

// Benchmark for SendMessage
func BenchmarkSendMessage(b *testing.B) {
	server := testServer(&testing.T{}, func(conn *websocket.Conn) {
		for {
			_, data, err := conn.ReadMessage()
			if err != nil {
				return
			}

			var req map[string]any
			if err := json.Unmarshal(data, &req); err != nil {
				return
			}

			success := true
			reqID, _ := req["id"].(float64) //nolint:errcheck // benchmark knows the format
			resp := types.HAMessage{
				ID:      int(reqID),
				Type:    "result",
				Success: &success,
				Result:  "pong",
			}
			if err := conn.WriteJSON(resp); err != nil {
				return
			}
		}
	})
	defer server.Close()

	url := "ws" + strings.TrimPrefix(server.URL, "http")
	conn, resp, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		b.Fatal(err)
	}
	if resp != nil && resp.Body != nil {
		resp.Body.Close()
	}
	client := New(conn)
	defer client.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := client.SendMessage("ping", nil); err != nil {
			b.Fatal(err)
		}
	}
}

// ============================================================================
// Context-aware tests for graceful shutdown functionality
// ============================================================================

func TestNewWithContext(t *testing.T) {
	server := testServer(t, func(_ *websocket.Conn) {
		time.Sleep(100 * time.Millisecond)
	})
	defer server.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	client := dialServerWithContext(ctx, t, server)
	defer client.Close()

	assert.NotNil(t, client)
	assert.NotNil(t, client.ctx)
	assert.NotNil(t, client.cancel)
	assert.NotNil(t, client.Context())
}

func TestClient_Context(t *testing.T) {
	server := testServer(t, func(_ *websocket.Conn) {
		time.Sleep(100 * time.Millisecond)
	})
	defer server.Close()

	ctx := context.Background()
	client := dialServerWithContext(ctx, t, server)
	defer client.Close()

	// Context() should return a child context (not the same as parent)
	clientCtx := client.Context()
	assert.NotNil(t, clientCtx)

	// Context should not be canceled initially
	select {
	case <-clientCtx.Done():
		t.Error("expected context to not be canceled initially")
	default:
		// Expected
	}
}

func TestClient_ContextCancellation(t *testing.T) {
	server := testServer(t, func(_ *websocket.Conn) {
		time.Sleep(500 * time.Millisecond)
	})
	defer server.Close()

	ctx, cancel := context.WithCancel(context.Background())
	client := dialServerWithContext(ctx, t, server)

	// Cancel the context
	cancel()

	// Wait for the done channel to be closed
	select {
	case <-client.Done():
		// Expected - client should stop when context is canceled
	case <-time.After(1 * time.Second):
		t.Fatal("done channel not closed after context cancellation")
	}
}

func TestClient_SendMessageWithContext_Cancellation(t *testing.T) {
	server := testServer(t, func(conn *websocket.Conn) {
		// Read request but don't respond, simulating a slow server
		_, _, _ = conn.ReadMessage() //nolint:errcheck // intentionally ignoring error
		time.Sleep(5 * time.Second)
	})
	defer server.Close()

	client := dialServer(t, server)
	defer client.Close()

	ctx, cancel := context.WithCancel(context.Background())

	// Start a goroutine to cancel the context after a short delay
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	// This should be canceled
	resp, err := client.SendMessageWithContext(ctx, "ping", nil)
	assert.Nil(t, resp)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "canceled")
}

func TestClient_SendMessageWithContext_Timeout(t *testing.T) {
	server := testServer(t, func(conn *websocket.Conn) {
		// Read request but don't respond, simulating a slow server
		_, _, _ = conn.ReadMessage() //nolint:errcheck // intentionally ignoring error
		time.Sleep(5 * time.Second)
	})
	defer server.Close()

	client := dialServer(t, server)
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	// This should timeout
	resp, err := client.SendMessageWithContext(ctx, "ping", nil)
	assert.Nil(t, resp)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "canceled")
}

func TestClient_SubscribeToTriggerWithContext_Cancellation(t *testing.T) {
	server := testServer(t, func(conn *websocket.Conn) {
		// Read request but don't respond, simulating a slow server
		_, _, _ = conn.ReadMessage() //nolint:errcheck // intentionally ignoring error
		time.Sleep(5 * time.Second)
	})
	defer server.Close()

	client := dialServer(t, server)
	defer client.Close()

	ctx, cancel := context.WithCancel(context.Background())

	// Start a goroutine to cancel the context after a short delay
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	// This should be canceled
	_, _, err := client.SubscribeToTriggerWithContext(ctx, map[string]any{
		"platform":  "state",
		"entity_id": "light.test",
	}, func(map[string]any) {}, 5*time.Second)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "canceled")
}

func TestClient_SubscribeToTemplateWithContext_Cancellation(t *testing.T) {
	server := testServer(t, func(conn *websocket.Conn) {
		// Read request but don't respond, simulating a slow server
		_, _, _ = conn.ReadMessage() //nolint:errcheck // intentionally ignoring error
		time.Sleep(5 * time.Second)
	})
	defer server.Close()

	client := dialServer(t, server)
	defer client.Close()

	ctx, cancel := context.WithCancel(context.Background())

	// Start a goroutine to cancel the context after a short delay
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	// This should be canceled
	_, _, err := client.SubscribeToTemplateWithContext(ctx, "{{ states('sensor.temp') }}", func(string) {}, 5*time.Second)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "canceled")
}

func TestClient_ClearSubscriptions(t *testing.T) {
	eventReceived := make(chan bool, 1)

	server := testServer(t, func(conn *websocket.Conn) {
		// Read subscription request
		_, data, err := conn.ReadMessage()
		require.NoError(t, err)

		var req map[string]any
		err = json.Unmarshal(data, &req)
		require.NoError(t, err)

		reqID, _ := req["id"].(float64) //nolint:errcheck // test knows the format
		id := int(reqID)

		// Send subscription confirmation
		success := true
		resp := types.HAMessage{
			ID:      id,
			Type:    "result",
			Success: &success,
		}
		err = conn.WriteJSON(resp)
		require.NoError(t, err)

		// Send event after a delay
		time.Sleep(50 * time.Millisecond)
		event := types.HAMessage{
			ID:   id,
			Type: "event",
			Event: &types.HAEvent{
				Variables: map[string]any{
					"trigger": map[string]any{
						"platform": "state",
					},
				},
			},
		}
		err = conn.WriteJSON(event)
		require.NoError(t, err)

		time.Sleep(200 * time.Millisecond)
	})
	defer server.Close()

	client := dialServer(t, server)
	defer client.Close()

	// Subscribe
	_, _, err := client.SubscribeToTrigger(map[string]any{
		"platform":  "state",
		"entity_id": "light.test",
	}, func(map[string]any) {
		eventReceived <- true
	}, 5*time.Second)
	require.NoError(t, err)

	// Verify subscription count
	assert.Equal(t, 1, client.SubscriptionCount())

	// Clear subscriptions
	client.ClearSubscriptions()

	// Verify subscriptions are cleared
	assert.Equal(t, 0, client.SubscriptionCount())

	// Event should not be received since subscription was cleared
	select {
	case <-eventReceived:
		t.Error("should not receive event after ClearSubscriptions")
	case <-time.After(150 * time.Millisecond):
		// Expected - no event received
	}
}

func TestClient_SubscriptionCount(t *testing.T) {
	server := testServer(t, func(conn *websocket.Conn) {
		for i := 0; i < 3; i++ {
			// Read subscription request
			_, data, err := conn.ReadMessage()
			if err != nil {
				return
			}

			var req map[string]any
			if err := json.Unmarshal(data, &req); err != nil {
				return
			}

			reqID, _ := req["id"].(float64) //nolint:errcheck // test knows the format

			// Send subscription confirmation
			success := true
			resp := types.HAMessage{
				ID:      int(reqID),
				Type:    "result",
				Success: &success,
			}
			if err := conn.WriteJSON(resp); err != nil {
				return
			}
		}

		time.Sleep(500 * time.Millisecond)
	})
	defer server.Close()

	client := dialServer(t, server)
	defer client.Close()

	// Initially zero subscriptions
	assert.Equal(t, 0, client.SubscriptionCount())

	// Create first subscription
	_, cleanup1, err := client.SubscribeToTrigger(map[string]any{
		"platform":  "state",
		"entity_id": "light.test1",
	}, func(map[string]any) {}, 0)
	require.NoError(t, err)
	assert.Equal(t, 1, client.SubscriptionCount())

	// Create second subscription
	_, cleanup2, err := client.SubscribeToTrigger(map[string]any{
		"platform":  "state",
		"entity_id": "light.test2",
	}, func(map[string]any) {}, 0)
	require.NoError(t, err)
	assert.Equal(t, 2, client.SubscriptionCount())

	// Cleanup first subscription
	cleanup1()
	assert.Equal(t, 1, client.SubscriptionCount())

	// Cleanup is idempotent
	cleanup1()
	assert.Equal(t, 1, client.SubscriptionCount())

	// Cleanup second subscription
	cleanup2()
	assert.Equal(t, 0, client.SubscriptionCount())
}

func TestClient_CloseWithContext(t *testing.T) {
	server := testServer(t, func(_ *websocket.Conn) {
		time.Sleep(500 * time.Millisecond)
	})
	defer server.Close()

	ctx := context.Background()
	client := dialServerWithContext(ctx, t, server)

	// Close should cancel the context
	err := client.Close()
	require.NoError(t, err)

	// Client's context should be canceled after close
	select {
	case <-client.Context().Done():
		// Expected
	case <-time.After(100 * time.Millisecond):
		t.Error("expected client context to be canceled after close")
	}
}

func TestClient_SubscriptionAutoCleanupOnContextCancel(t *testing.T) {
	server := testServer(t, func(conn *websocket.Conn) {
		// Read subscription request
		_, data, err := conn.ReadMessage()
		require.NoError(t, err)

		var req map[string]any
		err = json.Unmarshal(data, &req)
		require.NoError(t, err)

		reqID, _ := req["id"].(float64) //nolint:errcheck // test knows the format

		// Send subscription confirmation
		success := true
		resp := types.HAMessage{
			ID:      int(reqID),
			Type:    "result",
			Success: &success,
		}
		err = conn.WriteJSON(resp)
		require.NoError(t, err)

		time.Sleep(500 * time.Millisecond)
	})
	defer server.Close()

	ctx, cancel := context.WithCancel(context.Background())
	client := dialServerWithContext(ctx, t, server)
	defer client.Close()

	// Create subscription with timeout (which enables auto-cleanup on context cancel)
	_, _, err := client.SubscribeToTriggerWithContext(ctx, map[string]any{
		"platform":  "state",
		"entity_id": "light.test",
	}, func(map[string]any) {}, 10*time.Second) // Long timeout, but should cleanup on cancel
	require.NoError(t, err)

	assert.Equal(t, 1, client.SubscriptionCount())

	// Cancel context
	cancel()

	// Wait for cleanup goroutine to run
	time.Sleep(50 * time.Millisecond)

	// Subscription should be cleaned up
	assert.Equal(t, 0, client.SubscriptionCount())
}
