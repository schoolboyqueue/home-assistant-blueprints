// Package client provides WebSocket client utilities for communicating with Home Assistant.
package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"

	"github.com/home-assistant-blueprints/ha-ws-client-go/internal/types"
)

// Client represents a WebSocket client for Home Assistant.
type Client struct {
	conn           *websocket.Conn
	messageID      atomic.Int32
	pendingMu      sync.RWMutex
	pending        map[int]chan *types.HAMessage
	subscriptions  map[int]func(map[string]any)
	subscriptionMu sync.RWMutex
	done           chan struct{}
	readErr        error
}

// HAClientError represents an error from the Home Assistant API.
type HAClientError struct {
	Message string
	Code    string
}

func (e *HAClientError) Error() string {
	if e.Code != "" {
		return fmt.Sprintf("[%s] %s", e.Code, e.Message)
	}
	return e.Message
}

// New creates a new client connected to the given WebSocket connection.
func New(conn *websocket.Conn) *Client {
	c := &Client{
		conn:          conn,
		pending:       make(map[int]chan *types.HAMessage),
		subscriptions: make(map[int]func(map[string]any)),
		done:          make(chan struct{}),
	}
	go c.readLoop()
	return c
}

// NextID generates the next unique message ID.
func (c *Client) NextID() int {
	return int(c.messageID.Add(1))
}

// readLoop continuously reads messages from the WebSocket.
func (c *Client) readLoop() {
	defer close(c.done)
	for {
		_, data, err := c.conn.ReadMessage()
		if err != nil {
			c.readErr = err
			c.pendingMu.Lock()
			for _, ch := range c.pending {
				close(ch)
			}
			c.pending = make(map[int]chan *types.HAMessage)
			c.pendingMu.Unlock()
			return
		}

		var msg types.HAMessage
		if err := json.Unmarshal(data, &msg); err != nil {
			continue
		}

		c.handleMessage(&msg)
	}
}

// handleMessage routes incoming messages to their handlers.
func (c *Client) handleMessage(msg *types.HAMessage) {
	switch msg.Type {
	case "result", "pong":
		c.pendingMu.RLock()
		ch, ok := c.pending[msg.ID]
		c.pendingMu.RUnlock()
		if ok {
			ch <- msg
		}
	case "event":
		c.subscriptionMu.RLock()
		handler, ok := c.subscriptions[msg.ID]
		c.subscriptionMu.RUnlock()
		if ok && msg.Event != nil && msg.Event.Variables != nil {
			handler(msg.Event.Variables)
		}
	}
}

// SendMessage sends a message and waits for the response.
func (c *Client) SendMessage(msgType string, data map[string]any) (*types.HAMessage, error) {
	id := c.NextID()

	msg := map[string]any{
		"id":   id,
		"type": msgType,
	}
	for k, v := range data {
		msg[k] = v
	}

	respCh := make(chan *types.HAMessage, 1)
	c.pendingMu.Lock()
	c.pending[id] = respCh
	c.pendingMu.Unlock()

	defer func() {
		c.pendingMu.Lock()
		delete(c.pending, id)
		c.pendingMu.Unlock()
	}()

	msgBytes, err := json.Marshal(msg)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal message: %w", err)
	}

	if err := c.conn.WriteMessage(websocket.TextMessage, msgBytes); err != nil {
		return nil, fmt.Errorf("failed to send message: %w", err)
	}

	select {
	case resp, ok := <-respCh:
		if !ok {
			return nil, errors.New("connection closed")
		}
		if resp.Success != nil && !*resp.Success {
			errMsg := "unknown error"
			code := ""
			if resp.Error != nil {
				errMsg = resp.Error.Message
				code = resp.Error.Code
			}
			return nil, &HAClientError{Message: errMsg, Code: code}
		}
		return resp, nil
	case <-c.done:
		return nil, fmt.Errorf("connection closed: %w", c.readErr)
	}
}

// SendMessageTyped sends a message and unmarshals the result into the provided type.
func SendMessageTyped[T any](c *Client, msgType string, data map[string]any) (T, error) {
	var result T
	resp, err := c.SendMessage(msgType, data)
	if err != nil {
		return result, err
	}

	resultBytes, err := json.Marshal(resp.Result)
	if err != nil {
		return result, fmt.Errorf("failed to marshal result: %w", err)
	}

	if err := json.Unmarshal(resultBytes, &result); err != nil {
		return result, fmt.Errorf("failed to unmarshal result: %w", err)
	}

	return result, nil
}

// SubscribeToTrigger subscribes to a trigger and calls the callback for each event.
// Returns the subscription ID, a cleanup function, and any error.
func (c *Client) SubscribeToTrigger(trigger map[string]any, callback func(map[string]any), timeout time.Duration) (subscriptionID int, cleanup func(), err error) {
	id := c.NextID()

	msg := map[string]any{
		"id":      id,
		"type":    "subscribe_trigger",
		"trigger": trigger,
	}

	respCh := make(chan *types.HAMessage, 1)
	c.pendingMu.Lock()
	c.pending[id] = respCh
	c.pendingMu.Unlock()

	msgBytes, err := json.Marshal(msg)
	if err != nil {
		c.pendingMu.Lock()
		delete(c.pending, id)
		c.pendingMu.Unlock()
		return 0, nil, fmt.Errorf("failed to marshal message: %w", err)
	}

	if err := c.conn.WriteMessage(websocket.TextMessage, msgBytes); err != nil {
		c.pendingMu.Lock()
		delete(c.pending, id)
		c.pendingMu.Unlock()
		return 0, nil, fmt.Errorf("failed to send message: %w", err)
	}

	// Wait for subscription confirmation
	select {
	case resp, ok := <-respCh:
		if !ok {
			return 0, nil, errors.New("connection closed")
		}
		if resp.Success != nil && !*resp.Success {
			errMsg := "subscription failed"
			if resp.Error != nil {
				errMsg = resp.Error.Message
			}
			return 0, nil, errors.New(errMsg)
		}
	case <-time.After(5 * time.Second):
		c.pendingMu.Lock()
		delete(c.pending, id)
		c.pendingMu.Unlock()
		return 0, nil, errors.New("subscription timeout")
	}

	// Clean up pending request channel
	c.pendingMu.Lock()
	delete(c.pending, id)
	c.pendingMu.Unlock()

	// Register the subscription handler
	c.subscriptionMu.Lock()
	c.subscriptions[id] = callback
	c.subscriptionMu.Unlock()

	// Create cleanup function
	cleanupFn := func() {
		c.subscriptionMu.Lock()
		delete(c.subscriptions, id)
		c.subscriptionMu.Unlock()
	}

	// Auto-cleanup after timeout if specified
	if timeout > 0 {
		go func() {
			time.Sleep(timeout)
			cleanupFn()
		}()
	}

	return id, cleanupFn, nil
}

// SubscribeToTemplate subscribes to template rendering and calls the callback with results.
func (c *Client) SubscribeToTemplate(template string, callback func(string), timeout time.Duration) (subscriptionID int, cleanup func(), err error) {
	id := c.NextID()

	// Create cleanup function
	cleanupFn := func() {
		c.subscriptionMu.Lock()
		delete(c.subscriptions, id)
		c.subscriptionMu.Unlock()
	}

	// Register event handler BEFORE sending to avoid race condition
	c.subscriptionMu.Lock()
	c.subscriptions[id] = func(vars map[string]any) {
		if result, ok := vars["result"].(string); ok {
			callback(result)
		}
	}
	c.subscriptionMu.Unlock()

	msg := map[string]any{
		"id":       id,
		"type":     "render_template",
		"template": template,
	}

	respCh := make(chan *types.HAMessage, 1)
	c.pendingMu.Lock()
	c.pending[id] = respCh
	c.pendingMu.Unlock()

	msgBytes, err := json.Marshal(msg)
	if err != nil {
		cleanupFn()
		c.pendingMu.Lock()
		delete(c.pending, id)
		c.pendingMu.Unlock()
		return 0, nil, fmt.Errorf("failed to marshal message: %w", err)
	}

	if err := c.conn.WriteMessage(websocket.TextMessage, msgBytes); err != nil {
		cleanupFn()
		c.pendingMu.Lock()
		delete(c.pending, id)
		c.pendingMu.Unlock()
		return 0, nil, fmt.Errorf("failed to send message: %w", err)
	}

	// Wait for subscription confirmation
	select {
	case resp, ok := <-respCh:
		if !ok {
			cleanupFn()
			return 0, nil, errors.New("connection closed")
		}
		if resp.Success != nil && !*resp.Success {
			cleanupFn()
			errMsg := "subscription failed"
			if resp.Error != nil {
				errMsg = resp.Error.Message
			}
			return 0, nil, errors.New(errMsg)
		}
	case <-time.After(5 * time.Second):
		cleanupFn()
		c.pendingMu.Lock()
		delete(c.pending, id)
		c.pendingMu.Unlock()
		return 0, nil, errors.New("subscription timeout")
	}

	// Clean up pending request channel
	c.pendingMu.Lock()
	delete(c.pending, id)
	c.pendingMu.Unlock()

	// Auto-cleanup after timeout if specified
	if timeout > 0 {
		go func() {
			time.Sleep(timeout)
			cleanupFn()
		}()
	}

	return id, cleanupFn, nil
}

// Close closes the WebSocket connection.
func (c *Client) Close() error {
	return c.conn.Close()
}

// Done returns a channel that's closed when the client is done.
func (c *Client) Done() <-chan struct{} {
	return c.done
}
