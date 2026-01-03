// Package client provides WebSocket client utilities for communicating with Home Assistant.
package client

import (
	"context"
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
	ctx            context.Context
	cancel         context.CancelFunc
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
// Deprecated: Use NewWithContext instead for proper shutdown handling.
func New(conn *websocket.Conn) *Client {
	return NewWithContext(context.Background(), conn)
}

// NewWithContext creates a new client with context support for graceful shutdown.
// When the context is canceled, the client will stop processing messages and close.
func NewWithContext(ctx context.Context, conn *websocket.Conn) *Client {
	clientCtx, cancel := context.WithCancel(ctx)
	c := &Client{
		conn:          conn,
		pending:       make(map[int]chan *types.HAMessage),
		subscriptions: make(map[int]func(map[string]any)),
		done:          make(chan struct{}),
		ctx:           clientCtx,
		cancel:        cancel,
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
		// Check if context is canceled before reading
		select {
		case <-c.ctx.Done():
			c.readErr = c.ctx.Err()
			c.cleanupPending()
			return
		default:
		}

		_, data, err := c.conn.ReadMessage()
		if err != nil {
			c.readErr = err
			c.cleanupPending()
			return
		}

		var msg types.HAMessage
		if err := json.Unmarshal(data, &msg); err != nil {
			continue
		}

		c.handleMessage(&msg)
	}
}

// cleanupPending closes all pending response channels and clears the map.
func (c *Client) cleanupPending() {
	c.pendingMu.Lock()
	for _, ch := range c.pending {
		close(ch)
	}
	c.pending = make(map[int]chan *types.HAMessage)
	c.pendingMu.Unlock()
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
		if ok && msg.Event != nil {
			vars := msg.Event.Variables
			// render_template uses Event.Result, not Event.Variables
			if vars == nil && msg.Event.Result != nil {
				vars = map[string]any{"result": msg.Event.Result}
			}
			if vars != nil {
				handler(vars)
			}
		}
	}
}

// SendMessage sends a message and waits for the response.
// It uses the client's context for cancellation.
func (c *Client) SendMessage(msgType string, data map[string]any) (*types.HAMessage, error) {
	return c.SendMessageWithContext(c.ctx, msgType, data)
}

// SendMessageWithContext sends a message with explicit context for cancellation/timeout.
func (c *Client) SendMessageWithContext(ctx context.Context, msgType string, data map[string]any) (*types.HAMessage, error) {
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
	case <-ctx.Done():
		return nil, fmt.Errorf("request canceled: %w", ctx.Err())
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
// The timeout parameter controls auto-cleanup (0 means no auto-cleanup).
// For context-aware subscriptions, use SubscribeToTriggerWithContext.
func (c *Client) SubscribeToTrigger(trigger map[string]any, callback func(map[string]any), timeout time.Duration) (subscriptionID int, cleanup func(), err error) {
	return c.SubscribeToTriggerWithContext(c.ctx, trigger, callback, timeout)
}

// SubscribeToTriggerWithContext subscribes to a trigger with explicit context support.
// The context is used for the subscription setup. For cleanup on context cancellation,
// the caller should monitor ctx.Done() and call the cleanup function.
func (c *Client) SubscribeToTriggerWithContext(ctx context.Context, trigger map[string]any, callback func(map[string]any), timeout time.Duration) (subscriptionID int, cleanup func(), err error) {
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

	// Wait for subscription confirmation with context support
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
	case <-ctx.Done():
		c.pendingMu.Lock()
		delete(c.pending, id)
		c.pendingMu.Unlock()
		return 0, nil, fmt.Errorf("subscription canceled: %w", ctx.Err())
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

	// Create idempotent cleanup function
	var cleanupOnce sync.Once
	cleanupFn := func() {
		cleanupOnce.Do(func() {
			c.subscriptionMu.Lock()
			delete(c.subscriptions, id)
			c.subscriptionMu.Unlock()
		})
	}

	// Auto-cleanup after timeout if specified
	if timeout > 0 {
		go func() {
			select {
			case <-time.After(timeout):
				cleanupFn()
			case <-ctx.Done():
				cleanupFn()
			case <-c.done:
				// Connection closed, cleanup already handled
			}
		}()
	}

	return id, cleanupFn, nil
}

// SubscribeToTemplate subscribes to template rendering and calls the callback with results.
func (c *Client) SubscribeToTemplate(template string, callback func(string), timeout time.Duration) (subscriptionID int, cleanup func(), err error) {
	return c.SubscribeToTemplateWithContext(c.ctx, template, callback, timeout)
}

// SubscribeToTemplateWithContext subscribes to template rendering with explicit context support.
func (c *Client) SubscribeToTemplateWithContext(ctx context.Context, template string, callback func(string), timeout time.Duration) (subscriptionID int, cleanup func(), err error) {
	id := c.NextID()

	// Create idempotent cleanup function
	var cleanupOnce sync.Once
	cleanupFn := func() {
		cleanupOnce.Do(func() {
			c.subscriptionMu.Lock()
			delete(c.subscriptions, id)
			c.subscriptionMu.Unlock()
		})
	}

	// Register event handler BEFORE sending to avoid race condition
	c.subscriptionMu.Lock()
	c.subscriptions[id] = func(vars map[string]any) {
		if result, ok := vars["result"]; ok {
			// Convert result to string (could be string, int, float, etc.)
			callback(fmt.Sprintf("%v", result))
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

	// Wait for subscription confirmation with context support
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
	case <-ctx.Done():
		cleanupFn()
		c.pendingMu.Lock()
		delete(c.pending, id)
		c.pendingMu.Unlock()
		return 0, nil, fmt.Errorf("subscription canceled: %w", ctx.Err())
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
			select {
			case <-time.After(timeout):
				cleanupFn()
			case <-ctx.Done():
				cleanupFn()
			case <-c.done:
				// Connection closed, cleanup already handled
			}
		}()
	}

	return id, cleanupFn, nil
}

// Close closes the WebSocket connection.
func (c *Client) Close() error {
	if c.cancel != nil {
		c.cancel()
	}
	return c.conn.Close()
}

// Done returns a channel that's closed when the client is done.
func (c *Client) Done() <-chan struct{} {
	return c.done
}

// Context returns the client's context.
func (c *Client) Context() context.Context {
	return c.ctx
}

// ClearSubscriptions removes all active subscriptions.
// This is useful during graceful shutdown.
func (c *Client) ClearSubscriptions() {
	c.subscriptionMu.Lock()
	c.subscriptions = make(map[int]func(map[string]any))
	c.subscriptionMu.Unlock()
}

// SubscriptionCount returns the number of active subscriptions.
func (c *Client) SubscriptionCount() int {
	c.subscriptionMu.RLock()
	defer c.subscriptionMu.RUnlock()
	return len(c.subscriptions)
}
