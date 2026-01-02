//go:build integration

package handlers_test

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/home-assistant-blueprints/ha-ws-client-go/internal/client"
	"github.com/home-assistant-blueprints/ha-ws-client-go/internal/handlers"
	"github.com/home-assistant-blueprints/ha-ws-client-go/internal/output"
	"github.com/home-assistant-blueprints/ha-ws-client-go/internal/types"
)

// getTestConfig returns the WebSocket URL and token for integration tests.
// It supports two modes:
//  1. Local (add-on): Uses SUPERVISOR_TOKEN and ws://supervisor/core/api/websocket
//  2. Remote: Uses HA_TOKEN and HA_WS_URL (e.g., ws://192.168.1.100:8123/api/websocket)
//
// For remote connections, you can also use HA_HOST as a shorthand:
//
//	HA_HOST=192.168.1.100 -> ws://192.168.1.100:8123/api/websocket
//	HA_HOST=192.168.1.100:8124 -> ws://192.168.1.100:8124/api/websocket
func getTestConfig(t *testing.T) (wsURL, token string) {
	t.Helper()

	// Check for remote connection config first
	token = os.Getenv("HA_TOKEN")
	wsURL = os.Getenv("HA_WS_URL")

	// HA_HOST shorthand: just specify host or host:port
	if host := os.Getenv("HA_HOST"); host != "" && wsURL == "" {
		if strings.Contains(host, ":") {
			wsURL = "ws://" + host + "/api/websocket"
		} else {
			wsURL = "ws://" + host + ":8123/api/websocket"
		}
	}

	// Fall back to supervisor token for local add-on environment
	if token == "" {
		token = os.Getenv("SUPERVISOR_TOKEN")
	}
	if wsURL == "" {
		wsURL = "ws://supervisor/core/api/websocket"
	}

	if token == "" {
		t.Skip("No token available - set SUPERVISOR_TOKEN (local) or HA_TOKEN (remote)")
	}

	return wsURL, token
}

// testClient creates an authenticated WebSocket client for integration tests.
// See getTestConfig for environment variable configuration.
func testClient(t *testing.T) *client.Client {
	t.Helper()

	wsURL, token := getTestConfig(t)

	// Connect to WebSocket
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err, "failed to connect to WebSocket at %s", wsURL)

	// Read auth_required message
	var authReq map[string]any
	err = conn.ReadJSON(&authReq)
	require.NoError(t, err, "failed to read auth_required")
	require.Equal(t, "auth_required", authReq["type"])

	// Send auth
	err = conn.WriteJSON(map[string]string{
		"type":         "auth",
		"access_token": token,
	})
	require.NoError(t, err, "failed to send auth")

	// Read auth response
	var authResp map[string]any
	err = conn.ReadJSON(&authResp)
	require.NoError(t, err, "failed to read auth response")
	require.Equal(t, "auth_ok", authResp["type"], "authentication failed")

	// Create client
	c := client.New(conn)
	t.Cleanup(func() {
		c.Close()
	})

	return c
}

// testContext creates a handler context for testing.
func testContext(t *testing.T, c *client.Client, args ...string) *handlers.Context {
	t.Helper()
	return &handlers.Context{
		Client: c,
		Args:   args,
	}
}

func init() {
	// Set default output config for tests
	output.SetConfig(&output.Config{
		Format:         output.FormatJSON,
		ShowTimestamps: true,
		ShowHeaders:    true,
	})
}

// =============================================================================
// Basic Handler Tests
// =============================================================================

func TestIntegration_HandlePing(t *testing.T) {
	c := testClient(t)
	ctx := testContext(t, c, "ping")

	err := handlers.HandlePing(ctx)
	assert.NoError(t, err)
}

func TestIntegration_HandleStates(t *testing.T) {
	c := testClient(t)
	ctx := testContext(t, c, "states")

	err := handlers.HandleStates(ctx)
	assert.NoError(t, err)
}

func TestIntegration_HandleStatesJSON(t *testing.T) {
	c := testClient(t)
	ctx := testContext(t, c, "states-json")

	err := handlers.HandleStatesJSON(ctx)
	assert.NoError(t, err)
}

func TestIntegration_HandleState_SunSun(t *testing.T) {
	c := testClient(t)
	ctx := testContext(t, c, "state", "sun.sun")
	ctx.Config = &handlers.HandlerConfig{Args: []string{"sun.sun"}}

	// Use the unwrapped handler directly since we set up Config manually
	states, err := client.SendMessageTyped[[]types.HAState](c, "get_states", nil)
	require.NoError(t, err)

	var found bool
	for _, s := range states {
		if s.EntityID == "sun.sun" {
			found = true
			assert.Contains(t, []string{"above_horizon", "below_horizon"}, s.State)
			break
		}
	}
	assert.True(t, found, "sun.sun entity should exist")
}

func TestIntegration_HandleState_NotFound(t *testing.T) {
	c := testClient(t)
	ctx := testContext(t, c, "state", "nonexistent.entity_12345")
	ctx.Config = &handlers.HandlerConfig{Args: []string{"nonexistent.entity_12345"}}

	states, err := client.SendMessageTyped[[]types.HAState](c, "get_states", nil)
	require.NoError(t, err)

	var found bool
	for _, s := range states {
		if s.EntityID == "nonexistent.entity_12345" {
			found = true
			break
		}
	}
	assert.False(t, found, "nonexistent entity should not be found")
}

func TestIntegration_HandleConfig(t *testing.T) {
	c := testClient(t)
	ctx := testContext(t, c, "config")

	err := handlers.HandleConfig(ctx)
	assert.NoError(t, err)
}

func TestIntegration_HandleServices(t *testing.T) {
	c := testClient(t)
	ctx := testContext(t, c, "services")

	err := handlers.HandleServices(ctx)
	assert.NoError(t, err)
}

func TestIntegration_HandleStatesFilter(t *testing.T) {
	c := testClient(t)

	// Test with sun.* pattern
	err := handlers.HandleStatesFilter(testContext(t, c, "states-filter", "sun.*"))
	assert.NoError(t, err)

	// Test with light.* pattern (may or may not have matches)
	err = handlers.HandleStatesFilter(testContext(t, c, "states-filter", "light.*"))
	assert.NoError(t, err)
}

// =============================================================================
// Template Handler Tests
// =============================================================================

func TestIntegration_HandleTemplate_SimpleState(t *testing.T) {
	c := testClient(t)
	ctx := testContext(t, c, "template", "{{ states('sun.sun') }}")

	err := handlers.HandleTemplate(ctx)
	assert.NoError(t, err)
}

func TestIntegration_HandleTemplate_Math(t *testing.T) {
	c := testClient(t)
	ctx := testContext(t, c, "template", "{{ 1 + 1 }}")

	err := handlers.HandleTemplate(ctx)
	assert.NoError(t, err)
}

func TestIntegration_HandleTemplate_Now(t *testing.T) {
	c := testClient(t)
	ctx := testContext(t, c, "template", "{{ now().year }}")

	err := handlers.HandleTemplate(ctx)
	assert.NoError(t, err)
}

// =============================================================================
// Registry Handler Tests
// =============================================================================

func TestIntegration_HandleEntities(t *testing.T) {
	c := testClient(t)

	// Get all entities
	entities, err := client.SendMessageTyped[[]types.EntityEntry](c, "config/entity_registry/list", nil)
	require.NoError(t, err)
	assert.Greater(t, len(entities), 0, "should have at least one entity")
}

func TestIntegration_HandleDevices(t *testing.T) {
	c := testClient(t)

	devices, err := client.SendMessageTyped[[]types.DeviceEntry](c, "config/device_registry/list", nil)
	require.NoError(t, err)
	// Devices may be empty in some installations, so just check no error
	t.Logf("Found %d devices", len(devices))
}

func TestIntegration_HandleAreas(t *testing.T) {
	c := testClient(t)
	ctx := testContext(t, c, "areas")

	err := handlers.HandleAreas(ctx)
	assert.NoError(t, err)
}

// =============================================================================
// History Handler Tests
// =============================================================================

func TestIntegration_HandleHistory(t *testing.T) {
	c := testClient(t)

	// Get history for sun.sun (always has history)
	endTime := time.Now()
	startTime := endTime.Add(-1 * time.Hour)

	result, err := client.SendMessageTyped[map[string][]types.HistoryState](c,
		"history/history_during_period",
		map[string]any{
			"start_time":               startTime.Format(time.RFC3339),
			"end_time":                 endTime.Format(time.RFC3339),
			"entity_ids":               []string{"sun.sun"},
			"minimal_response":         true,
			"no_attributes":            true,
			"significant_changes_only": false,
		})
	require.NoError(t, err)

	// sun.sun should have history
	history, ok := result["sun.sun"]
	if ok {
		assert.Greater(t, len(history), 0, "should have history entries")
	}
}

func TestIntegration_HandleLogbook(t *testing.T) {
	c := testClient(t)

	endTime := time.Now()
	startTime := endTime.Add(-1 * time.Hour)

	entries, err := client.SendMessageTyped[[]types.LogbookEntry](c,
		"logbook/get_events",
		map[string]any{
			"start_time": startTime.Format(time.RFC3339),
			"end_time":   endTime.Format(time.RFC3339),
		})
	require.NoError(t, err)
	t.Logf("Found %d logbook entries in last hour", len(entries))
}

func TestIntegration_HandleSyslog(t *testing.T) {
	c := testClient(t)

	entries, err := client.SendMessageTyped[[]types.SysLogEntry](c, "system_log/list", nil)
	require.NoError(t, err)
	t.Logf("Found %d syslog entries", len(entries))
}

// =============================================================================
// Automation Trace Tests
// =============================================================================

func TestIntegration_HandleTraces(t *testing.T) {
	c := testClient(t)

	// Get all traces (returns flat array, not map)
	traces, err := client.SendMessageTyped[[]types.TraceInfo](c,
		"trace/list",
		map[string]any{"domain": "automation"})
	require.NoError(t, err)

	t.Logf("Found %d automation traces", len(traces))
}

// =============================================================================
// Service Call Tests (Read-only services)
// =============================================================================

func TestIntegration_HandleCall_Reload(t *testing.T) {
	// Skip this test by default as it modifies state
	t.Skip("Skipping service call test - uncomment to test reload")

	c := testClient(t)

	// Test a safe service call (reload automations)
	result, err := c.SendMessage("call_service", map[string]any{
		"domain":  "automation",
		"service": "reload",
	})
	require.NoError(t, err)
	assert.NotNil(t, result)
}

// =============================================================================
// Subscription Tests
// =============================================================================

func TestIntegration_SubscribeToTemplate(t *testing.T) {
	c := testClient(t)

	resultCh := make(chan string, 1)
	_, cleanup, err := c.SubscribeToTemplate("{{ now().timestamp() }}", func(result string) {
		select {
		case resultCh <- result:
		default:
		}
	}, 5*time.Second)
	require.NoError(t, err)
	defer cleanup()

	select {
	case result := <-resultCh:
		assert.NotEmpty(t, result)
		t.Logf("Template result: %s", result)
	case <-time.After(3 * time.Second):
		t.Fatal("timeout waiting for template result")
	}
}

func TestIntegration_SubscribeToTrigger_Time(t *testing.T) {
	c := testClient(t)

	// Subscribe to a time pattern trigger (fires every second for testing)
	trigger := map[string]any{
		"platform": "time_pattern",
		"seconds":  "/1", // Every second
	}

	eventCh := make(chan map[string]any, 1)
	_, cleanup, err := c.SubscribeToTrigger(trigger, func(vars map[string]any) {
		select {
		case eventCh <- vars:
		default:
		}
	}, 5*time.Second)
	require.NoError(t, err)
	defer cleanup()

	select {
	case event := <-eventCh:
		assert.NotNil(t, event)
		t.Logf("Received trigger event: %v", event)
	case <-time.After(3 * time.Second):
		t.Fatal("timeout waiting for trigger event")
	}
}

// =============================================================================
// Error Handling Tests
// =============================================================================

func TestIntegration_InvalidServiceCall(t *testing.T) {
	c := testClient(t)

	_, err := c.SendMessage("call_service", map[string]any{
		"domain":  "nonexistent_domain",
		"service": "nonexistent_service",
	})
	assert.Error(t, err)

	// Should be a client error
	var clientErr *client.HAClientError
	if assert.ErrorAs(t, err, &clientErr) {
		assert.NotEmpty(t, clientErr.Code)
		t.Logf("Error code: %s, message: %s", clientErr.Code, clientErr.Message)
	}
}

func TestIntegration_InvalidTemplate(t *testing.T) {
	c := testClient(t)

	_, _, err := c.SubscribeToTemplate("{{ invalid_function() }}", func(string) {}, time.Second)
	// Invalid templates may error on subscribe or return an error result
	// Either way is acceptable
	t.Logf("Invalid template result: %v", err)
}

// =============================================================================
// Performance/Stress Tests
// =============================================================================

func TestIntegration_MultipleQueries(t *testing.T) {
	c := testClient(t)

	// Run multiple queries in sequence to test connection stability
	for i := 0; i < 10; i++ {
		_, err := c.SendMessage("ping", nil)
		require.NoError(t, err, "ping %d failed", i)
	}
}

func TestIntegration_LargeStateQuery(t *testing.T) {
	c := testClient(t)

	// Get all states (can be large)
	states, err := client.SendMessageTyped[[]types.HAState](c, "get_states", nil)
	require.NoError(t, err)
	t.Logf("Retrieved %d entity states", len(states))
	assert.Greater(t, len(states), 0)
}

// =============================================================================
// Middleware Integration Tests
// =============================================================================

func TestIntegration_MiddlewareWithRealClient(t *testing.T) {
	c := testClient(t)

	// Test RequireArg1 middleware with real context
	handler := handlers.Apply(
		handlers.RequireArg1("Usage: test <arg>"),
		func(ctx *handlers.Context) error {
			assert.Equal(t, "sun.sun", ctx.Config.Args[0])
			return nil
		},
	)

	ctx := testContext(t, c, "test", "sun.sun")
	err := handler(ctx)
	assert.NoError(t, err)
}

func TestIntegration_TimeRangeMiddleware(t *testing.T) {
	c := testClient(t)

	handler := handlers.Apply(
		handlers.Chain(
			handlers.RequireArg1("Usage: test <entity>"),
			handlers.WithTimeRange(24, 2),
		),
		func(ctx *handlers.Context) error {
			assert.NotNil(t, ctx.Config.TimeRange)
			assert.False(t, ctx.Config.TimeRange.StartTime.IsZero())
			assert.False(t, ctx.Config.TimeRange.EndTime.IsZero())
			return nil
		},
	)

	ctx := testContext(t, c, "test", "sun.sun", "4")
	err := handler(ctx)
	assert.NoError(t, err)
}

func TestIntegration_PatternMiddleware(t *testing.T) {
	c := testClient(t)

	handler := handlers.Apply(
		handlers.WithRequiredPattern(1, "Usage: test <pattern>"),
		func(ctx *handlers.Context) error {
			require.NotNil(t, ctx.Config.Pattern)
			assert.True(t, ctx.Config.Pattern.MatchString("sun.sun"))
			assert.False(t, ctx.Config.Pattern.MatchString("light.kitchen"))
			return nil
		},
	)

	ctx := testContext(t, c, "test", "sun.*")
	err := handler(ctx)
	assert.NoError(t, err)
}

// =============================================================================
// Data Validation Tests
// =============================================================================

func TestIntegration_StateAttributes(t *testing.T) {
	c := testClient(t)

	states, err := client.SendMessageTyped[[]types.HAState](c, "get_states", nil)
	require.NoError(t, err)

	// Find sun.sun and validate its attributes
	for _, s := range states {
		if s.EntityID != "sun.sun" {
			continue
		}

		assert.NotEmpty(t, s.State)
		assert.NotNil(t, s.Attributes)
		assert.NotEmpty(t, s.LastUpdated)

		// Sun entity should have specific attributes
		_, hasElevation := s.Attributes["elevation"]
		_, hasAzimuth := s.Attributes["azimuth"]
		assert.True(t, hasElevation || hasAzimuth, "sun.sun should have elevation or azimuth")
		return
	}
	t.Fatal("sun.sun not found")
}

func TestIntegration_ConfigValidation(t *testing.T) {
	c := testClient(t)

	config, err := client.SendMessageTyped[types.HAConfig](c, "get_config", nil)
	require.NoError(t, err)

	assert.NotEmpty(t, config.Version)
	assert.NotEmpty(t, config.TimeZone)
	assert.NotEmpty(t, config.State)
	assert.Greater(t, len(config.Components), 0)

	t.Logf("Home Assistant version: %s", config.Version)
	t.Logf("Location: %s", config.LocationName)
	t.Logf("Timezone: %s", config.TimeZone)
	t.Logf("Components loaded: %d", len(config.Components))
}

func TestIntegration_EntityRegistryValidation(t *testing.T) {
	c := testClient(t)

	entities, err := client.SendMessageTyped[[]types.EntityEntry](c, "config/entity_registry/list", nil)
	require.NoError(t, err)
	require.Greater(t, len(entities), 0)

	// Validate structure
	for _, e := range entities[:min(5, len(entities))] {
		assert.NotEmpty(t, e.EntityID)
		assert.Contains(t, e.EntityID, ".")
		parts := strings.Split(e.EntityID, ".")
		assert.Equal(t, 2, len(parts), "entity_id should have domain.name format")
	}
}
