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

// Test fixture entity IDs (from testdata/ha-config/configuration.yaml)
// These are available when running in CI or with the test config mounted.
const (
	testInputBoolean   = "input_boolean.test_switch"
	testInputNumber    = "input_number.test_temperature"
	testInputText      = "input_text.test_message"
	testInputSelect    = "input_select.test_mode"
	testTemplateSensor = "sensor.test_calculated_value"
	testAutomationID   = "test_automation_toggle"
	// Entity ID is derived from alias, not the id field
	testAutomationFull = "automation.test_toggle_on_temperature_change"
	testScript         = "script.test_script"
)

// hasTestFixtures checks if the test fixtures are available.
// Returns true if running in CI or against an HA instance with test config.
func hasTestFixtures(t *testing.T, c *client.Client) bool {
	t.Helper()
	states, err := client.SendMessageTyped[[]types.HAState](c, "get_states", nil)
	if err != nil {
		return false
	}
	for _, s := range states {
		if s.EntityID == testInputBoolean {
			return true
		}
	}
	return false
}

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

// =============================================================================
// Test Fixture Tests (require testdata/ha-config to be mounted)
// =============================================================================

func TestIntegration_Fixtures_InputHelpers(t *testing.T) {
	c := testClient(t)

	if !hasTestFixtures(t, c) {
		t.Skip("Test fixtures not available - skipping fixture tests")
	}

	states, err := client.SendMessageTyped[[]types.HAState](c, "get_states", nil)
	require.NoError(t, err)

	// Check that our test entities exist
	foundEntities := make(map[string]bool)
	testEntities := []string{testInputBoolean, testInputNumber, testInputText, testInputSelect}

	for _, s := range states {
		for _, te := range testEntities {
			if s.EntityID == te {
				foundEntities[te] = true
			}
		}
	}

	for _, te := range testEntities {
		assert.True(t, foundEntities[te], "test entity %s should exist", te)
	}
}

func TestIntegration_Fixtures_ServiceCall(t *testing.T) {
	c := testClient(t)

	if !hasTestFixtures(t, c) {
		t.Skip("Test fixtures not available - skipping fixture tests")
	}

	// Get current state
	states, err := client.SendMessageTyped[[]types.HAState](c, "get_states", nil)
	require.NoError(t, err)

	var initialState string
	for _, s := range states {
		if s.EntityID == testInputBoolean {
			initialState = s.State
			break
		}
	}

	// Toggle the input boolean
	_, err = c.SendMessage("call_service", map[string]any{
		"domain":  "input_boolean",
		"service": "toggle",
		"target": map[string]any{
			"entity_id": testInputBoolean,
		},
	})
	require.NoError(t, err)

	// Verify state changed
	time.Sleep(500 * time.Millisecond)
	states, err = client.SendMessageTyped[[]types.HAState](c, "get_states", nil)
	require.NoError(t, err)

	var newState string
	for _, s := range states {
		if s.EntityID == testInputBoolean {
			newState = s.State
			break
		}
	}

	assert.NotEqual(t, initialState, newState, "state should have changed after toggle")

	// Toggle back to restore original state
	_, err = c.SendMessage("call_service", map[string]any{
		"domain":  "input_boolean",
		"service": "toggle",
		"target": map[string]any{
			"entity_id": testInputBoolean,
		},
	})
	require.NoError(t, err)
}

func TestIntegration_Fixtures_InputNumberHistory(t *testing.T) {
	c := testClient(t)

	if !hasTestFixtures(t, c) {
		t.Skip("Test fixtures not available - skipping fixture tests")
	}

	// Set a value to ensure history exists
	_, err := c.SendMessage("call_service", map[string]any{
		"domain":  "input_number",
		"service": "set_value",
		"target": map[string]any{
			"entity_id": testInputNumber,
		},
		"service_data": map[string]any{
			"value": 22.5,
		},
	})
	require.NoError(t, err)

	time.Sleep(500 * time.Millisecond)

	// Query history
	endTime := time.Now()
	startTime := endTime.Add(-1 * time.Hour)

	result, err := client.SendMessageTyped[map[string][]types.HistoryState](c,
		"history/history_during_period",
		map[string]any{
			"start_time":               startTime.Format(time.RFC3339),
			"end_time":                 endTime.Format(time.RFC3339),
			"entity_ids":               []string{testInputNumber},
			"minimal_response":         true,
			"no_attributes":            true,
			"significant_changes_only": false,
		})
	require.NoError(t, err)

	history, ok := result[testInputNumber]
	assert.True(t, ok, "should have history for %s", testInputNumber)
	if ok {
		assert.Greater(t, len(history), 0, "should have at least one history entry")
	}
}

func TestIntegration_Fixtures_TemplateSensor(t *testing.T) {
	c := testClient(t)

	if !hasTestFixtures(t, c) {
		t.Skip("Test fixtures not available - skipping fixture tests")
	}

	states, err := client.SendMessageTyped[[]types.HAState](c, "get_states", nil)
	require.NoError(t, err)

	var found bool
	for _, s := range states {
		if s.EntityID != testTemplateSensor {
			continue
		}

		found = true
		// Template sensor should have a numeric state (temperature in F)
		assert.NotEmpty(t, s.State)
		assert.NotEqual(t, "unavailable", s.State)
		assert.NotEqual(t, "unknown", s.State)
		t.Logf("Template sensor %s = %s", testTemplateSensor, s.State)
		break
	}
	assert.True(t, found, "template sensor %s should exist", testTemplateSensor)
}

func TestIntegration_Fixtures_AutomationTraces(t *testing.T) {
	c := testClient(t)

	if !hasTestFixtures(t, c) {
		t.Skip("Test fixtures not available - skipping fixture tests")
	}

	// Trigger the automation by changing the input number
	_, err := c.SendMessage("call_service", map[string]any{
		"domain":  "input_number",
		"service": "set_value",
		"target": map[string]any{
			"entity_id": testInputNumber,
		},
		"service_data": map[string]any{
			"value": 30.0,
		},
	})
	require.NoError(t, err)

	// Wait for automation to run
	time.Sleep(1 * time.Second)

	// Get traces
	traces, err := client.SendMessageTyped[[]types.TraceInfo](c,
		"trace/list",
		map[string]any{"domain": "automation"})
	require.NoError(t, err)

	// Look for our test automation trace
	// Note: trace/list returns item_id as the automation's id field, not entity_id
	var foundTrace bool
	var foundItemIDs []string
	for _, trace := range traces {
		foundItemIDs = append(foundItemIDs, trace.ItemID)
		if trace.ItemID == testAutomationID {
			foundTrace = true
			t.Logf("Found trace for %s: run_id=%s, state=%s",
				trace.ItemID, trace.RunID, trace.State)
			break
		}
	}

	if !foundTrace {
		t.Logf("Available trace item_ids: %v", foundItemIDs)
	}
	assert.True(t, foundTrace, "should have trace for test automation %s", testAutomationID)
}

func TestIntegration_Fixtures_StatesFilter(t *testing.T) {
	c := testClient(t)

	if !hasTestFixtures(t, c) {
		t.Skip("Test fixtures not available - skipping fixture tests")
	}

	states, err := client.SendMessageTyped[[]types.HAState](c, "get_states", nil)
	require.NoError(t, err)

	// Filter for input_boolean entities
	var inputBooleans []types.HAState
	for _, s := range states {
		if strings.HasPrefix(s.EntityID, "input_boolean.") {
			inputBooleans = append(inputBooleans, s)
		}
	}

	// Should have at least our test input booleans
	assert.GreaterOrEqual(t, len(inputBooleans), 3,
		"should have at least 3 input_boolean entities from fixtures")

	// Verify our specific test entities are present
	entityIDs := make([]string, len(inputBooleans))
	for i, s := range inputBooleans {
		entityIDs[i] = s.EntityID
	}
	assert.Contains(t, entityIDs, testInputBoolean)
}

func TestIntegration_Fixtures_Script(t *testing.T) {
	c := testClient(t)

	if !hasTestFixtures(t, c) {
		t.Skip("Test fixtures not available - skipping fixture tests")
	}

	// Call our test script
	_, err := c.SendMessage("call_service", map[string]any{
		"domain":  "script",
		"service": "test_script",
	})
	require.NoError(t, err)

	// Script runs asynchronously, just verify it was called successfully
	t.Log("Test script executed successfully")
}

// =============================================================================
// Context Handler Tests
// =============================================================================

func TestIntegration_HandleContext_WithEntityID(t *testing.T) {
	c := testClient(t)

	if !hasTestFixtures(t, c) {
		t.Skip("Test fixtures not available - skipping fixture tests")
	}

	// Trigger the automation to create a causal chain:
	// input_number.test_temperature change -> automation -> input_boolean.test_switch toggle
	_, err := c.SendMessage("call_service", map[string]any{
		"domain":  "input_number",
		"service": "set_value",
		"target": map[string]any{
			"entity_id": testInputNumber,
		},
		"service_data": map[string]any{
			"value": 42.0,
		},
	})
	require.NoError(t, err)

	// Wait for automation to run
	time.Sleep(1 * time.Second)

	// Now test the context handler with an entity_id
	ctx := testContext(t, c, "context", testInputBoolean)
	err = handlers.HandleContext(ctx)
	assert.NoError(t, err)
}

func TestIntegration_HandleContext_WithNonexistentEntity(t *testing.T) {
	c := testClient(t)

	// Test with a non-existent entity (should treat as context_id)
	ctx := testContext(t, c, "context", "nonexistent.entity_xyz")
	err := handlers.HandleContext(ctx)
	// Should not error, just show "no states found" message
	assert.NoError(t, err)
}

func TestIntegration_HandleContext_FindsRelatedStates(t *testing.T) {
	c := testClient(t)

	if !hasTestFixtures(t, c) {
		t.Skip("Test fixtures not available - skipping fixture tests")
	}

	// Trigger an automation to create related state changes
	_, err := c.SendMessage("call_service", map[string]any{
		"domain":  "input_number",
		"service": "set_value",
		"target": map[string]any{
			"entity_id": testInputNumber,
		},
		"service_data": map[string]any{
			"value": 35.0,
		},
	})
	require.NoError(t, err)

	time.Sleep(1 * time.Second)

	// Get the input_boolean state and verify it has a context
	states, err := client.SendMessageTyped[[]types.HAState](c, "get_states", nil)
	require.NoError(t, err)

	var switchState *types.HAState
	for i, s := range states {
		if s.EntityID == testInputBoolean {
			switchState = &states[i]
			break
		}
	}
	require.NotNil(t, switchState, "test switch should exist")
	require.NotNil(t, switchState.Context, "test switch should have context")

	t.Logf("Test switch context: ID=%s, ParentID=%s",
		switchState.Context.ID, switchState.Context.ParentID)
}

// =============================================================================
// Traces Handler Tests (with last_triggered discrepancy)
// =============================================================================

const testAutomationNoTraces = "test_automation_no_traces"

func TestIntegration_HandleTraces_ShowsLastTriggered(t *testing.T) {
	c := testClient(t)

	if !hasTestFixtures(t, c) {
		t.Skip("Test fixtures not available - skipping fixture tests")
	}

	// Trigger the automation that has stored_traces: 0
	_, err := c.SendMessage("call_service", map[string]any{
		"domain":  "input_number",
		"service": "set_value",
		"target": map[string]any{
			"entity_id": "input_number.target_humidity",
		},
		"service_data": map[string]any{
			"value": 55.0,
		},
	})
	require.NoError(t, err)

	time.Sleep(1 * time.Second)

	// Verify the automation has last_triggered but no traces
	states, err := client.SendMessageTyped[[]types.HAState](c, "get_states", nil)
	require.NoError(t, err)

	var automationState *types.HAState
	for i, s := range states {
		if s.EntityID == "automation.test_no_traces_stored" {
			automationState = &states[i]
			break
		}
	}

	if automationState != nil {
		lastTriggered, ok := automationState.Attributes["last_triggered"].(string)
		if ok && lastTriggered != "" {
			t.Logf("Automation last_triggered: %s", lastTriggered)

			// Now verify traces command shows the discrepancy
			ctx := testContext(t, c, "traces", testAutomationNoTraces)
			err = handlers.HandleTraces(ctx)
			assert.NoError(t, err)
		}
	}
}

func TestIntegration_HandleTraces_WithTraces(t *testing.T) {
	c := testClient(t)

	if !hasTestFixtures(t, c) {
		t.Skip("Test fixtures not available - skipping fixture tests")
	}

	// Trigger an automation that stores traces
	_, err := c.SendMessage("call_service", map[string]any{
		"domain":  "input_number",
		"service": "set_value",
		"target": map[string]any{
			"entity_id": testInputNumber,
		},
		"service_data": map[string]any{
			"value": 28.0,
		},
	})
	require.NoError(t, err)

	time.Sleep(1 * time.Second)

	// Get traces for the test automation
	ctx := testContext(t, c, "traces", testAutomationID)
	err = handlers.HandleTraces(ctx)
	assert.NoError(t, err)
}

// =============================================================================
// Automation Config Handler Tests
// =============================================================================

func TestIntegration_HandleAutomationConfig_WithTraces(t *testing.T) {
	c := testClient(t)

	if !hasTestFixtures(t, c) {
		t.Skip("Test fixtures not available - skipping fixture tests")
	}

	// First ensure we have a trace
	_, err := c.SendMessage("call_service", map[string]any{
		"domain":  "input_number",
		"service": "set_value",
		"target": map[string]any{
			"entity_id": testInputNumber,
		},
		"service_data": map[string]any{
			"value": 33.0,
		},
	})
	require.NoError(t, err)

	time.Sleep(1 * time.Second)

	// Test automation-config command
	ctx := testContext(t, c, "automation-config", testAutomationFull)
	err = handlers.HandleAutomationConfig(ctx)
	assert.NoError(t, err)
}

func TestIntegration_HandleAutomationConfig_DirectAPI(t *testing.T) {
	c := testClient(t)

	if !hasTestFixtures(t, c) {
		t.Skip("Test fixtures not available - skipping fixture tests")
	}

	// Test the automation/config API directly
	type configResponse struct {
		Config types.AutomationConfig `json:"config"`
	}
	result, err := client.SendMessageTyped[configResponse](c, "automation/config", map[string]any{
		"entity_id": testAutomationFull,
	})
	require.NoError(t, err)

	t.Logf("Automation config: ID=%s, Alias=%s", result.Config.ID, result.Config.Alias)
	assert.NotEmpty(t, result.Config.ID)
	assert.NotEmpty(t, result.Config.Alias)
}

func TestIntegration_HandleAutomationConfig_NonBlueprintAutomation(t *testing.T) {
	c := testClient(t)

	if !hasTestFixtures(t, c) {
		t.Skip("Test fixtures not available - skipping fixture tests")
	}

	// Test with our non-blueprint test automation
	// Should return full config since it's defined in YAML
	type configResponse struct {
		Config types.AutomationConfig `json:"config"`
	}
	result, err := client.SendMessageTyped[configResponse](c, "automation/config", map[string]any{
		"entity_id": testAutomationFull,
	})
	require.NoError(t, err)

	// Non-blueprint automations should have trigger and action
	t.Logf("Config has %d triggers, %d actions",
		len(result.Config.Trigger), len(result.Config.Action))
}

// =============================================================================
// Monitor Handler Tests
// =============================================================================

func TestIntegration_WatchSubscription(t *testing.T) {
	c := testClient(t)

	// Test subscription to state trigger (similar to watch command)
	trigger := map[string]any{
		"platform":  "state",
		"entity_id": "sun.sun",
	}

	eventCh := make(chan map[string]any, 1)
	_, cleanup, err := c.SubscribeToTrigger(trigger, func(vars map[string]any) {
		select {
		case eventCh <- vars:
		default:
		}
	}, 2*time.Second)
	require.NoError(t, err)
	defer cleanup()

	// sun.sun changes slowly, so we may or may not get an event
	// Just verify the subscription was set up correctly
	select {
	case event := <-eventCh:
		assert.NotNil(t, event)
		t.Logf("Received state change event: %v", event)
	case <-time.After(2 * time.Second):
		// Timeout is acceptable - sun.sun doesn't change frequently
		t.Log("No state change within timeout (expected for slow-changing entity)")
	}
}

func TestIntegration_WatchWithTestFixtures(t *testing.T) {
	c := testClient(t)

	if !hasTestFixtures(t, c) {
		t.Skip("Test fixtures not available - skipping fixture tests")
	}

	// Set up subscription to watch test_switch
	trigger := map[string]any{
		"platform":  "state",
		"entity_id": testInputBoolean,
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

	// Toggle the input boolean to trigger an event
	time.Sleep(100 * time.Millisecond) // Brief pause to ensure subscription is active
	_, err = c.SendMessage("call_service", map[string]any{
		"domain":  "input_boolean",
		"service": "toggle",
		"target": map[string]any{
			"entity_id": testInputBoolean,
		},
	})
	require.NoError(t, err)

	// Wait for event
	select {
	case event := <-eventCh:
		assert.NotNil(t, event)
		// Verify event structure
		if trigger, ok := event["trigger"].(map[string]any); ok {
			t.Logf("Trigger platform: %v", trigger["platform"])
			if toState, ok := trigger["to_state"].(map[string]any); ok {
				t.Logf("New state: %v", toState["state"])
			}
		}
	case <-time.After(3 * time.Second):
		t.Fatal("timeout waiting for state change event")
	}

	// Toggle back to restore state
	_, err = c.SendMessage("call_service", map[string]any{
		"domain":  "input_boolean",
		"service": "toggle",
		"target": map[string]any{
			"entity_id": testInputBoolean,
		},
	})
	require.NoError(t, err)
}

func TestIntegration_MonitorMultipleEntities(t *testing.T) {
	c := testClient(t)

	if !hasTestFixtures(t, c) {
		t.Skip("Test fixtures not available - skipping fixture tests")
	}

	// Get initial states
	states, err := client.SendMessageTyped[[]types.HAState](c, "get_states", nil)
	require.NoError(t, err)

	// Find our test entities
	var testEntities []string
	for _, s := range states {
		if s.EntityID == testInputBoolean || s.EntityID == testInputNumber {
			testEntities = append(testEntities, s.EntityID)
		}
	}
	require.GreaterOrEqual(t, len(testEntities), 2, "should have at least 2 test entities")

	t.Logf("Found test entities for monitoring: %v", testEntities)

	// Set up subscription for both entities (simulating monitor-multi)
	trigger := map[string]any{
		"platform":  "state",
		"entity_id": testEntities,
	}

	eventCh := make(chan map[string]any, 2)
	_, cleanup, err := c.SubscribeToTrigger(trigger, func(vars map[string]any) {
		select {
		case eventCh <- vars:
		default:
		}
	}, 5*time.Second)
	require.NoError(t, err)
	defer cleanup()

	// Trigger a change
	time.Sleep(100 * time.Millisecond)
	_, err = c.SendMessage("call_service", map[string]any{
		"domain":  "input_boolean",
		"service": "toggle",
		"target": map[string]any{
			"entity_id": testInputBoolean,
		},
	})
	require.NoError(t, err)

	// Wait for event
	select {
	case event := <-eventCh:
		assert.NotNil(t, event)
		t.Logf("Received multi-entity event: %+v", event)
	case <-time.After(3 * time.Second):
		t.Fatal("timeout waiting for multi-entity event")
	}

	// Restore state
	_, err = c.SendMessage("call_service", map[string]any{
		"domain":  "input_boolean",
		"service": "toggle",
		"target": map[string]any{
			"entity_id": testInputBoolean,
		},
	})
	require.NoError(t, err)
}

func TestIntegration_AnalyzePattern(t *testing.T) {
	c := testClient(t)

	// Test pattern-based entity filtering (used by analyze command)
	states, err := client.SendMessageTyped[[]types.HAState](c, "get_states", nil)
	require.NoError(t, err)

	// Filter by pattern
	var sunEntities []types.HAState
	for _, s := range states {
		if strings.HasPrefix(s.EntityID, "sun.") {
			sunEntities = append(sunEntities, s)
		}
	}

	assert.Greater(t, len(sunEntities), 0, "should have at least one sun.* entity")

	for _, e := range sunEntities {
		t.Logf("Found entity: %s = %s", e.EntityID, e.State)
	}
}

func TestIntegration_AnalyzeInputPattern(t *testing.T) {
	c := testClient(t)

	if !hasTestFixtures(t, c) {
		t.Skip("Test fixtures not available - skipping fixture tests")
	}

	states, err := client.SendMessageTyped[[]types.HAState](c, "get_states", nil)
	require.NoError(t, err)

	// Filter for input_* entities
	var inputEntities []types.HAState
	for _, s := range states {
		if strings.HasPrefix(s.EntityID, "input_") {
			inputEntities = append(inputEntities, s)
		}
	}

	assert.GreaterOrEqual(t, len(inputEntities), 4,
		"should have at least 4 input_* entities from fixtures")

	t.Logf("Found %d input_* entities", len(inputEntities))
	for _, e := range inputEntities {
		t.Logf("  %s = %s", e.EntityID, e.State)
	}
}

// =============================================================================
// Compare and Device Health Handler Tests
// =============================================================================

func TestIntegration_CompareEntities(t *testing.T) {
	c := testClient(t)

	states, err := client.SendMessageTyped[[]types.HAState](c, "get_states", nil)
	require.NoError(t, err)

	// Find two entities to compare
	var entity1, entity2 *types.HAState
	for i := range states {
		if strings.HasPrefix(states[i].EntityID, "sensor.") {
			if entity1 == nil {
				entity1 = &states[i]
			} else if entity2 == nil {
				entity2 = &states[i]
				break
			}
		}
	}

	if entity1 == nil || entity2 == nil {
		// Fall back to sun entities
		for i := range states {
			if states[i].EntityID == "sun.sun" {
				entity1 = &states[i]
				entity2 = &states[i] // Compare to itself
				break
			}
		}
	}

	require.NotNil(t, entity1, "should find at least one entity to compare")
	require.NotNil(t, entity2, "should find a second entity to compare")

	t.Logf("Comparing %s vs %s", entity1.EntityID, entity2.EntityID)

	// Verify both entities have expected fields
	assert.NotEmpty(t, entity1.State)
	assert.NotEmpty(t, entity2.State)
}

func TestIntegration_DeviceHealth(t *testing.T) {
	c := testClient(t)

	states, err := client.SendMessageTyped[[]types.HAState](c, "get_states", nil)
	require.NoError(t, err)

	// Find sun.sun for health check (always exists)
	var sunEntity *types.HAState
	for i := range states {
		if states[i].EntityID == "sun.sun" {
			sunEntity = &states[i]
			break
		}
	}

	require.NotNil(t, sunEntity, "sun.sun should exist")
	assert.NotEmpty(t, sunEntity.LastUpdated, "should have last_updated")

	// Parse the timestamp to verify it's valid
	_, err = time.Parse(time.RFC3339, sunEntity.LastUpdated)
	assert.NoError(t, err, "last_updated should be valid RFC3339")

	t.Logf("sun.sun last updated: %s", sunEntity.LastUpdated)
}

func TestIntegration_DeviceHealthWithFixtures(t *testing.T) {
	c := testClient(t)

	if !hasTestFixtures(t, c) {
		t.Skip("Test fixtures not available - skipping fixture tests")
	}

	states, err := client.SendMessageTyped[[]types.HAState](c, "get_states", nil)
	require.NoError(t, err)

	// Find test input boolean
	var inputBoolEntity *types.HAState
	for i := range states {
		if states[i].EntityID == testInputBoolean {
			inputBoolEntity = &states[i]
			break
		}
	}

	require.NotNil(t, inputBoolEntity, "%s should exist", testInputBoolean)
	assert.NotEmpty(t, inputBoolEntity.LastUpdated)

	// Calculate age
	lastUpdated, err := time.Parse(time.RFC3339, inputBoolEntity.LastUpdated)
	require.NoError(t, err)

	age := time.Since(lastUpdated)
	t.Logf("%s last updated %s ago", testInputBoolean, age.Round(time.Second))

	// Recently triggered entity should be fresh
	assert.Less(t, age, 24*time.Hour, "test entity should have been updated in last 24 hours")
}
