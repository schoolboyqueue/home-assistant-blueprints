package testfixtures

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/require"
)

// wsUpgrader is used to upgrade HTTP connections to WebSocket.
var wsUpgrader = websocket.Upgrader{
	CheckOrigin: func(_ *http.Request) bool { return true },
}

// WSHandler is a function that handles WebSocket connections in tests.
type WSHandler func(*websocket.Conn)

// TestServer creates a test WebSocket server that handles messages.
// The handler function is called with the WebSocket connection.
func TestServer(t *testing.T, handler WSHandler) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := wsUpgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Fatalf("failed to upgrade: %v", err)
		}
		defer conn.Close()
		handler(conn)
	}))
}

// DialServer connects to a test server and returns the WebSocket connection.
func DialServer(t *testing.T, server *httptest.Server) *websocket.Conn {
	t.Helper()
	url := "ws" + strings.TrimPrefix(server.URL, "http")
	conn, resp, err := websocket.DefaultDialer.Dial(url, nil)
	require.NoError(t, err)
	if resp != nil && resp.Body != nil {
		resp.Body.Close()
	}
	return conn
}

// =====================================
// Pre-built WebSocket Handlers
// =====================================

// EchoHandler returns a handler that echoes back received messages.
func EchoHandler(t *testing.T) WSHandler {
	return func(conn *websocket.Conn) {
		for {
			messageType, data, err := conn.ReadMessage()
			if err != nil {
				return
			}
			if err := conn.WriteMessage(messageType, data); err != nil {
				return
			}
		}
	}
}

// SuccessHandler returns a handler that responds to any message with success.
func SuccessHandler(t *testing.T, result any) WSHandler {
	return func(conn *websocket.Conn) {
		for {
			_, data, err := conn.ReadMessage()
			if err != nil {
				return
			}

			var req map[string]any
			if err := json.Unmarshal(data, &req); err != nil {
				return
			}

			reqID, _ := req["id"].(float64)
			resp := NewSuccessMessage(int(reqID), result)
			if err := conn.WriteJSON(resp); err != nil {
				return
			}
		}
	}
}

// ErrorHandler returns a handler that responds to any message with an error.
func ErrorHandler(t *testing.T, code, message string) WSHandler {
	return func(conn *websocket.Conn) {
		_, data, err := conn.ReadMessage()
		if err != nil {
			return
		}

		var req map[string]any
		if err := json.Unmarshal(data, &req); err != nil {
			return
		}

		reqID, _ := req["id"].(float64)
		resp := NewErrorMessage(int(reqID), code, message)
		if err := conn.WriteJSON(resp); err != nil {
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
}

// AuthFlowHandler returns a handler that simulates the HA auth flow.
// It sends auth_required, waits for auth, and sends auth_ok.
func AuthFlowHandler(t *testing.T, token string, afterAuth WSHandler) WSHandler {
	return func(conn *websocket.Conn) {
		// Send auth_required
		if err := conn.WriteJSON(NewAuthRequiredMessage()); err != nil {
			t.Errorf("failed to send auth_required: %v", err)
			return
		}

		// Read auth message
		_, data, err := conn.ReadMessage()
		if err != nil {
			return
		}

		var authMsg map[string]any
		if err := json.Unmarshal(data, &authMsg); err != nil {
			return
		}

		// Validate auth
		if authMsg["type"] == "auth" && authMsg["access_token"] == token {
			if err := conn.WriteJSON(NewAuthOKMessage()); err != nil {
				return
			}
			if afterAuth != nil {
				afterAuth(conn)
			}
		} else {
			if err := conn.WriteJSON(NewAuthInvalidMessage("Invalid token")); err != nil {
				return
			}
		}
	}
}

// SubscriptionHandler returns a handler that confirms subscriptions and sends events.
func SubscriptionHandler(t *testing.T, events ...HAMessage) WSHandler {
	return func(conn *websocket.Conn) {
		// Read subscription request
		_, data, err := conn.ReadMessage()
		if err != nil {
			return
		}

		var req map[string]any
		if err := json.Unmarshal(data, &req); err != nil {
			return
		}

		reqID, _ := req["id"].(float64)
		id := int(reqID)

		// Send subscription confirmation
		resp := NewSuccessMessage(id, nil)
		if err := conn.WriteJSON(resp); err != nil {
			return
		}

		// Send events
		time.Sleep(50 * time.Millisecond)
		for _, event := range events {
			event.ID = id
			if err := conn.WriteJSON(event); err != nil {
				return
			}
			time.Sleep(50 * time.Millisecond)
		}

		// Keep connection open
		time.Sleep(200 * time.Millisecond)
	}
}

// DelayHandler returns a handler that delays before doing anything.
// Useful for testing timeouts.
func DelayHandler(t *testing.T, delay time.Duration) WSHandler {
	return func(conn *websocket.Conn) {
		_, _, _ = conn.ReadMessage()
		time.Sleep(delay)
	}
}

// CloseHandler returns a handler that closes the connection immediately.
func CloseHandler(t *testing.T) WSHandler {
	return func(conn *websocket.Conn) {
		conn.Close()
	}
}

// ReadThenCloseHandler returns a handler that reads one message then closes.
func ReadThenCloseHandler(t *testing.T) WSHandler {
	return func(conn *websocket.Conn) {
		_, _, _ = conn.ReadMessage()
		conn.Close()
	}
}

// KeepAliveHandler returns a handler that keeps the connection open.
func KeepAliveHandler(t *testing.T, duration time.Duration) WSHandler {
	return func(_ *websocket.Conn) {
		time.Sleep(duration)
	}
}

// =====================================
// Request/Response Helpers
// =====================================

// ReadRequest reads and parses a JSON request from the connection.
func ReadRequest(t *testing.T, conn *websocket.Conn) map[string]any {
	t.Helper()
	_, data, err := conn.ReadMessage()
	require.NoError(t, err)

	var req map[string]any
	err = json.Unmarshal(data, &req)
	require.NoError(t, err)

	return req
}

// GetRequestID extracts the request ID from a parsed request.
func GetRequestID(req map[string]any) int {
	if id, ok := req["id"].(float64); ok {
		return int(id)
	}
	return 0
}

// WriteResponse writes a response to the connection.
func WriteResponse(t *testing.T, conn *websocket.Conn, resp any) {
	t.Helper()
	err := conn.WriteJSON(resp)
	require.NoError(t, err)
}

// SendSuccessResponse reads a request and sends a success response.
func SendSuccessResponse(t *testing.T, conn *websocket.Conn, result any) {
	t.Helper()
	req := ReadRequest(t, conn)
	resp := NewSuccessMessage(GetRequestID(req), result)
	WriteResponse(t, conn, resp)
	time.Sleep(100 * time.Millisecond)
}

// SendErrorResponse reads a request and sends an error response.
func SendErrorResponse(t *testing.T, conn *websocket.Conn, code, message string) {
	t.Helper()
	req := ReadRequest(t, conn)
	resp := NewErrorMessage(GetRequestID(req), code, message)
	WriteResponse(t, conn, resp)
	time.Sleep(100 * time.Millisecond)
}
