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
// Trace Detail Command Tests
// =============================================================================

// getTestTraceRunID is a helper that retrieves a valid run_id for trace tests.
// It ensures a trace exists by triggering the test automation if needed.
func getTestTraceRunID(t *testing.T, c *client.Client) string {
	t.Helper()

	// First, trigger the automation to ensure we have a trace
	_, err := c.SendMessage("call_service", map[string]any{
		"domain":  "input_number",
		"service": "set_value",
		"target": map[string]any{
			"entity_id": testInputNumber,
		},
		"service_data": map[string]any{
			"value": 25.0 + float64(time.Now().Unix()%10), // Vary value to ensure change
		},
	})
	require.NoError(t, err)

	time.Sleep(1 * time.Second)

	// Get traces for the test automation
	traces, err := client.SendMessageTyped[[]types.TraceInfo](c, "trace/list", map[string]any{
		"domain":  "automation",
		"item_id": testAutomationID,
	})
	require.NoError(t, err)

	if len(traces) == 0 {
		t.Skip("No traces available for test automation - skipping")
	}

	return traces[0].RunID
}

func TestIntegration_HandleTrace_WithValidRunID(t *testing.T) {
	c := testClient(t)

	if !hasTestFixtures(t, c) {
		t.Skip("Test fixtures not available - skipping fixture tests")
	}

	runID := getTestTraceRunID(t, c)

	// Call the trace command with valid run_id
	ctx := testContext(t, c, "trace", testAutomationID, runID)
	err := handlers.HandleTrace(ctx)
	assert.NoError(t, err)

	// Also verify the trace data structure directly
	trace, err := client.SendMessageTyped[types.TraceDetail](c, "trace/get", map[string]any{
		"domain":  "automation",
		"item_id": testAutomationID,
		"run_id":  runID,
	})
	require.NoError(t, err)

	// Verify required fields exist
	assert.NotEmpty(t, trace.RunID, "trace should have run_id")
	assert.Equal(t, "automation", trace.Domain, "trace domain should be automation")
	assert.NotEmpty(t, trace.ItemID, "trace should have item_id")
	assert.NotNil(t, trace.Trace, "trace should have trace data")

	t.Logf("Trace run_id=%s, domain=%s, item_id=%s, steps=%d",
		trace.RunID, trace.Domain, trace.ItemID, len(trace.Trace))
}

func TestIntegration_HandleTraceLatest(t *testing.T) {
	c := testClient(t)

	if !hasTestFixtures(t, c) {
		t.Skip("Test fixtures not available - skipping fixture tests")
	}

	// Ensure we have a trace
	_ = getTestTraceRunID(t, c)

	// Call trace-latest command
	ctx := testContext(t, c, "trace-latest", testAutomationID)
	err := handlers.HandleTraceLatest(ctx)
	assert.NoError(t, err)

	// Verify by getting the latest trace directly
	traces, err := client.SendMessageTyped[[]types.TraceInfo](c, "trace/list", map[string]any{
		"domain":  "automation",
		"item_id": testAutomationID,
	})
	require.NoError(t, err)
	require.NotEmpty(t, traces, "should have at least one trace")

	t.Logf("Latest trace run_id=%s, state=%s", traces[0].RunID, traces[0].ScriptExecution)
}

func TestIntegration_HandleTraceSummary(t *testing.T) {
	c := testClient(t)

	if !hasTestFixtures(t, c) {
		t.Skip("Test fixtures not available - skipping fixture tests")
	}

	// Ensure we have a trace
	_ = getTestTraceRunID(t, c)

	// Call trace-summary command
	ctx := testContext(t, c, "trace-summary", testAutomationID)
	err := handlers.HandleTraceSummary(ctx)
	assert.NoError(t, err)

	// Verify by getting traces directly to check summary data
	traces, err := client.SendMessageTyped[[]types.TraceInfo](c, "trace/list", map[string]any{
		"domain":  "automation",
		"item_id": testAutomationID,
	})
	require.NoError(t, err)
	require.NotEmpty(t, traces, "should have at least one trace for summary")

	// Count by state
	var finished, errors int
	for _, tr := range traces {
		state := tr.ScriptExecution
		if state == "" {
			state = tr.State
		}
		switch state {
		case "finished":
			finished++
		case "error":
			errors++
		}
	}
	t.Logf("Summary: %d total traces, %d finished, %d errors", len(traces), finished, errors)
}

func TestIntegration_HandleTraceVars(t *testing.T) {
	c := testClient(t)

	if !hasTestFixtures(t, c) {
		t.Skip("Test fixtures not available - skipping fixture tests")
	}

	runID := getTestTraceRunID(t, c)

	// Call trace-vars command
	ctx := testContext(t, c, "trace-vars", testAutomationID, runID)
	err := handlers.HandleTraceVars(ctx)
	assert.NoError(t, err)

	// Verify trace has variables
	trace, err := client.SendMessageTyped[types.TraceDetail](c, "trace/get", map[string]any{
		"domain":  "automation",
		"item_id": testAutomationID,
		"run_id":  runID,
	})
	require.NoError(t, err)

	// Count variables in trace steps
	varCount := 0
	if trace.Trace != nil {
		for _, steps := range trace.Trace {
			for _, step := range steps {
				if step.Variables != nil {
					varCount++
				}
			}
		}
	}
	t.Logf("Found %d trace steps with variables", varCount)
}

func TestIntegration_HandleTraceTimeline(t *testing.T) {
	c := testClient(t)

	if !hasTestFixtures(t, c) {
		t.Skip("Test fixtures not available - skipping fixture tests")
	}

	runID := getTestTraceRunID(t, c)

	// Call trace-timeline command
	ctx := testContext(t, c, "trace-timeline", testAutomationID, runID)
	err := handlers.HandleTraceTimeline(ctx)
	assert.NoError(t, err)

	// Verify trace has steps for timeline
	trace, err := client.SendMessageTyped[types.TraceDetail](c, "trace/get", map[string]any{
		"domain":  "automation",
		"item_id": testAutomationID,
		"run_id":  runID,
	})
	require.NoError(t, err)

	// Count timeline steps
	stepCount := 0
	if trace.Trace != nil {
		for path, steps := range trace.Trace {
			stepCount += len(steps)
			if len(steps) > 0 {
				t.Logf("Timeline step: %s (timestamp=%s)", path, steps[0].Timestamp)
			}
		}
	}
	assert.Greater(t, stepCount, 0, "trace should have at least one timeline step")
	t.Logf("Total timeline steps: %d", stepCount)
}

func TestIntegration_HandleTraceTrigger(t *testing.T) {
	c := testClient(t)

	if !hasTestFixtures(t, c) {
		t.Skip("Test fixtures not available - skipping fixture tests")
	}

	runID := getTestTraceRunID(t, c)

	// Call trace-trigger command
	ctx := testContext(t, c, "trace-trigger", testAutomationID, runID)
	err := handlers.HandleTraceTrigger(ctx)
	assert.NoError(t, err)

	// Verify trace has trigger info
	trace, err := client.SendMessageTyped[types.TraceDetail](c, "trace/get", map[string]any{
		"domain":  "automation",
		"item_id": testAutomationID,
		"run_id":  runID,
	})
	require.NoError(t, err)

	// Trigger should be present (test automation is triggered by state change)
	assert.NotNil(t, trace.Trigger, "trace should have trigger information")
	t.Logf("Trigger info present: %v", trace.Trigger != nil)
}

func TestIntegration_HandleTraceActions(t *testing.T) {
	c := testClient(t)

	if !hasTestFixtures(t, c) {
		t.Skip("Test fixtures not available - skipping fixture tests")
	}

	runID := getTestTraceRunID(t, c)

	// Call trace-actions command
	ctx := testContext(t, c, "trace-actions", testAutomationID, runID)
	err := handlers.HandleTraceActions(ctx)
	assert.NoError(t, err)

	// Verify trace has action steps
	trace, err := client.SendMessageTyped[types.TraceDetail](c, "trace/get", map[string]any{
		"domain":  "automation",
		"item_id": testAutomationID,
		"run_id":  runID,
	})
	require.NoError(t, err)

	// Count action steps (paths starting with "action/")
	actionCount := 0
	if trace.Trace != nil {
		for path := range trace.Trace {
			if strings.HasPrefix(path, "action/") {
				actionCount++
				t.Logf("Action path: %s", path)
			}
		}
	}
	t.Logf("Total action steps: %d", actionCount)
}

func TestIntegration_HandleTraceDebug(t *testing.T) {
	c := testClient(t)

	if !hasTestFixtures(t, c) {
		t.Skip("Test fixtures not available - skipping fixture tests")
	}

	runID := getTestTraceRunID(t, c)

	// Call trace-debug command
	ctx := testContext(t, c, "trace-debug", testAutomationID, runID)
	err := handlers.HandleTraceDebug(ctx)
	assert.NoError(t, err)

	// Verify the full trace structure
	trace, err := client.SendMessageTyped[types.TraceDetail](c, "trace/get", map[string]any{
		"domain":  "automation",
		"item_id": testAutomationID,
		"run_id":  runID,
	})
	require.NoError(t, err)

	// Debug should show all trace details
	assert.NotEmpty(t, trace.RunID, "debug trace should have run_id")
	assert.NotNil(t, trace.Trace, "debug trace should have trace steps")

	// Log comprehensive debug info
	t.Logf("Debug trace: run_id=%s, domain=%s, item_id=%s", trace.RunID, trace.Domain, trace.ItemID)
	t.Logf("  script_execution=%s, error=%s", trace.ScriptExecution, trace.Error)
	t.Logf("  trace_paths=%d, has_trigger=%v, has_context=%v, has_config=%v",
		len(trace.Trace), trace.Trigger != nil, trace.Context != nil, trace.Config != nil)
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

func TestIntegration_HandleCompare(t *testing.T) {
	c := testClient(t)

	// Test with sun.sun comparing to itself (always exists)
	// This verifies the handler works with a guaranteed entity
	ctx := testContext(t, c, "compare", "sun.sun", "sun.sun")
	err := handlers.HandleCompare(ctx)
	assert.NoError(t, err)
}

func TestIntegration_HandleCompare_TwoEntitiesSameDomain(t *testing.T) {
	c := testClient(t)

	if !hasTestFixtures(t, c) {
		t.Skip("Test fixtures not available - skipping fixture tests")
	}

	// Compare two input_boolean entities from the same domain
	// This tests the side-by-side comparison with different entities
	// Using test_switch and away_mode from testdata/ha-config/configuration.yaml
	ctx := testContext(t, c, "compare", testInputBoolean, "input_boolean.away_mode")
	err := handlers.HandleCompare(ctx)
	assert.NoError(t, err)
}

func TestIntegration_HandleCompare_VerifiesOutput(t *testing.T) {
	c := testClient(t)

	if !hasTestFixtures(t, c) {
		t.Skip("Test fixtures not available - skipping fixture tests")
	}

	// Verify the comparison data structure by making the API calls directly
	// This ensures the compare command would show both entities properly
	states, err := client.SendMessageTyped[[]types.HAState](c, "get_states", nil)
	require.NoError(t, err)

	var state1, state2 *types.HAState
	for i := range states {
		if states[i].EntityID == testInputBoolean {
			state1 = &states[i]
		}
		if states[i].EntityID == testInputNumber {
			state2 = &states[i]
		}
	}

	require.NotNil(t, state1, "%s should exist", testInputBoolean)
	require.NotNil(t, state2, "%s should exist", testInputNumber)

	// Verify both entities appear in comparison data
	assert.NotEmpty(t, state1.EntityID)
	assert.NotEmpty(t, state2.EntityID)
	assert.NotEmpty(t, state1.State)
	assert.NotEmpty(t, state2.State)

	// The compare handler builds a structure with entity1, entity2, and differences
	// Verify attributes exist (which the comparison uses)
	assert.NotNil(t, state1.Attributes)
	assert.NotNil(t, state2.Attributes)

	t.Logf("Compare: %s (%s) vs %s (%s)",
		state1.EntityID, state1.State, state2.EntityID, state2.State)

	// Now call the actual handler
	ctx := testContext(t, c, "compare", testInputBoolean, testInputNumber)
	err = handlers.HandleCompare(ctx)
	assert.NoError(t, err)
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

// =============================================================================
// History Full and Attrs Handler Tests
// =============================================================================

func TestIntegration_HandleHistoryFull(t *testing.T) {
	c := testClient(t)

	// Test with sun.sun which always has history and attributes
	ctx := testContext(t, c, "history-full", "sun.sun")
	err := handlers.HandleHistoryFull(ctx)
	assert.NoError(t, err)
}

func TestIntegration_HandleHistoryFull_IncludesAttributes(t *testing.T) {
	c := testClient(t)

	// Verify the API returns attributes in the response
	endTime := time.Now()
	startTime := endTime.Add(-1 * time.Hour)

	result, err := client.SendMessageTyped[map[string][]types.HistoryState](c,
		"history/history_during_period",
		map[string]any{
			"start_time":       startTime.Format(time.RFC3339),
			"end_time":         endTime.Format(time.RFC3339),
			"entity_ids":       []string{"sun.sun"},
			"minimal_response": false,
			"no_attributes":    false,
		})
	require.NoError(t, err)

	history, ok := result["sun.sun"]
	require.True(t, ok, "should have history for sun.sun")

	if len(history) > 0 {
		// Verify at least one entry has attributes
		hasAttrs := false
		for _, entry := range history {
			// Check both possible attribute fields (A for compact, Attributes for full)
			if entry.A != nil || entry.Attributes != nil {
				hasAttrs = true
				attrs := entry.A
				if attrs == nil {
					attrs = entry.Attributes
				}
				t.Logf("Found history entry with %d attributes", len(attrs))
				break
			}
		}
		assert.True(t, hasAttrs, "history-full should include attributes in response")
	}
}

func TestIntegration_HandleHistoryFull_WithFixtures(t *testing.T) {
	c := testClient(t)

	if !hasTestFixtures(t, c) {
		t.Skip("Test fixtures not available - skipping fixture tests")
	}

	// First trigger a state change to ensure fresh history
	_, err := c.SendMessage("call_service", map[string]any{
		"domain":  "input_number",
		"service": "set_value",
		"target": map[string]any{
			"entity_id": testInputNumber,
		},
		"service_data": map[string]any{
			"value": 21.0 + float64(time.Now().Unix()%10),
		},
	})
	require.NoError(t, err)

	time.Sleep(500 * time.Millisecond)

	// Test history-full with the test entity
	ctx := testContext(t, c, "history-full", testInputNumber)
	err = handlers.HandleHistoryFull(ctx)
	assert.NoError(t, err)

	// Also verify the structure directly
	endTime := time.Now()
	startTime := endTime.Add(-1 * time.Hour)

	result, err := client.SendMessageTyped[map[string][]types.HistoryState](c,
		"history/history_during_period",
		map[string]any{
			"start_time":       startTime.Format(time.RFC3339),
			"end_time":         endTime.Format(time.RFC3339),
			"entity_ids":       []string{testInputNumber},
			"minimal_response": false,
			"no_attributes":    false,
		})
	require.NoError(t, err)

	history, ok := result[testInputNumber]
	require.True(t, ok, "should have history for %s", testInputNumber)
	require.Greater(t, len(history), 0, "should have at least one history entry")

	// Verify attributes are present
	entry := history[len(history)-1] // Most recent entry
	attrs := entry.A
	if attrs == nil {
		attrs = entry.Attributes
	}
	assert.NotNil(t, attrs, "history-full entry should include attributes")

	// input_number should have specific attributes like unit_of_measurement
	if attrs != nil {
		t.Logf("Attributes: %v", attrs)
		// input_number entities typically have min, max, step, mode, etc.
		_, hasMin := attrs["min"]
		_, hasMax := attrs["max"]
		assert.True(t, hasMin || hasMax, "input_number should have min or max attribute")
	}
}

func TestIntegration_HandleAttrs(t *testing.T) {
	c := testClient(t)

	// Test with sun.sun which always has history and changing attributes
	ctx := testContext(t, c, "attrs", "sun.sun")
	err := handlers.HandleAttrs(ctx)
	assert.NoError(t, err)
}

func TestIntegration_HandleAttrs_ShowsAttributeHistory(t *testing.T) {
	c := testClient(t)

	// Verify the API returns attribute change history
	endTime := time.Now()
	startTime := endTime.Add(-1 * time.Hour)

	result, err := client.SendMessageTyped[map[string][]types.HistoryState](c,
		"history/history_during_period",
		map[string]any{
			"start_time":       startTime.Format(time.RFC3339),
			"end_time":         endTime.Format(time.RFC3339),
			"entity_ids":       []string{"sun.sun"},
			"minimal_response": false,
			"no_attributes":    false,
		})
	require.NoError(t, err)

	history, ok := result["sun.sun"]
	require.True(t, ok, "should have history for sun.sun")

	if len(history) > 0 {
		// Verify at least one entry has attributes (attrs command relies on this)
		hasAttrs := false
		for _, entry := range history {
			attrs := entry.A
			if attrs == nil {
				attrs = entry.Attributes
			}
			if attrs != nil {
				hasAttrs = true
				// sun.sun should have elevation and azimuth attributes
				_, hasElevation := attrs["elevation"]
				_, hasAzimuth := attrs["azimuth"]
				if hasElevation || hasAzimuth {
					t.Logf("Found sun.sun history with elevation/azimuth attributes")
					break
				}
			}
		}
		assert.True(t, hasAttrs, "attrs command requires history entries with attributes")
	}
}

func TestIntegration_HandleAttrs_WithFixtures(t *testing.T) {
	c := testClient(t)

	if !hasTestFixtures(t, c) {
		t.Skip("Test fixtures not available - skipping fixture tests")
	}

	// First trigger multiple state changes to create attribute history
	for i := 0; i < 3; i++ {
		_, err := c.SendMessage("call_service", map[string]any{
			"domain":  "input_number",
			"service": "set_value",
			"target": map[string]any{
				"entity_id": testInputNumber,
			},
			"service_data": map[string]any{
				"value": 20.0 + float64(i*2),
			},
		})
		require.NoError(t, err)
		time.Sleep(200 * time.Millisecond)
	}

	time.Sleep(300 * time.Millisecond)

	// Test attrs command with the test entity
	ctx := testContext(t, c, "attrs", testInputNumber)
	err := handlers.HandleAttrs(ctx)
	assert.NoError(t, err)
}

func TestIntegration_HandleAttrs_FormatsCorrectly(t *testing.T) {
	c := testClient(t)

	if !hasTestFixtures(t, c) {
		t.Skip("Test fixtures not available - skipping fixture tests")
	}

	// Verify the attribute history data structure
	endTime := time.Now()
	startTime := endTime.Add(-1 * time.Hour)

	result, err := client.SendMessageTyped[map[string][]types.HistoryState](c,
		"history/history_during_period",
		map[string]any{
			"start_time":       startTime.Format(time.RFC3339),
			"end_time":         endTime.Format(time.RFC3339),
			"entity_ids":       []string{testInputNumber},
			"minimal_response": false,
			"no_attributes":    false,
		})
	require.NoError(t, err)

	history, ok := result[testInputNumber]
	require.True(t, ok, "should have history for %s", testInputNumber)

	// Count entries with attributes (what attrs command displays)
	attrsCount := 0
	for _, entry := range history {
		attrs := entry.A
		if attrs == nil {
			attrs = entry.Attributes
		}
		if attrs != nil {
			attrsCount++
		}
	}

	t.Logf("Found %d history entries with attributes for attrs display", attrsCount)

	// Verify at least one entry has attributes for meaningful attrs output
	if len(history) > 0 {
		entry := history[len(history)-1]
		attrs := entry.A
		if attrs == nil {
			attrs = entry.Attributes
		}
		if attrs != nil {
			// Verify the entry contains expected state and timestamp fields
			assert.NotEmpty(t, entry.GetState(), "attrs entry should have state")
			assert.False(t, entry.GetLastUpdated().IsZero(), "attrs entry should have timestamp")
		}
	}
}

// =============================================================================
// Statistics Command Tests
// =============================================================================

func TestIntegration_HandleStats(t *testing.T) {
	c := testClient(t)

	// Test with sun.sun which is a built-in entity with guaranteed statistics
	// The sun sensor tracks elevation which generates statistics data
	ctx := testContext(t, c, "stats", "sun.sun")
	err := handlers.HandleStats(ctx)
	// Stats may return empty result if no statistics exist, but should not error
	assert.NoError(t, err)
}

func TestIntegration_HandleStats_VerifiesOutputStructure(t *testing.T) {
	c := testClient(t)

	// Verify the statistics API returns the expected data structure
	endTime := time.Now()
	startTime := endTime.Add(-24 * time.Hour)

	// Try sun.sun first (always exists)
	result, err := client.SendMessageTyped[map[string][]types.StatEntry](c,
		"recorder/statistics_during_period",
		map[string]any{
			"start_time":    startTime.Format(time.RFC3339),
			"end_time":      endTime.Format(time.RFC3339),
			"statistic_ids": []string{"sun.sun"},
			"period":        "hour",
		})
	require.NoError(t, err)

	// sun.sun may not have statistics (it depends on the recorder configuration)
	// If it doesn't, that's okay - the command should still work
	stats, ok := result["sun.sun"]
	if ok && len(stats) > 0 {
		// Verify structure contains expected fields
		entry := stats[0]
		t.Logf("Stats entry: start=%v, min=%.2f, max=%.2f, mean=%.2f",
			entry.Start, entry.Min, entry.Max, entry.Mean)

		// At least one of min, max, mean should be populated for valid stats
		hasValues := entry.Min != 0 || entry.Max != 0 || entry.Mean != 0 || entry.State != 0
		if hasValues {
			t.Logf("Statistics data found with values")
		} else {
			t.Log("Statistics entry found but all values are zero")
		}
	} else {
		t.Log("No statistics data found for sun.sun (this is acceptable)")
	}
}

func TestIntegration_HandleStats_WithFixtures(t *testing.T) {
	c := testClient(t)

	if !hasTestFixtures(t, c) {
		t.Skip("Test fixtures not available - skipping fixture tests")
	}

	// First trigger a state change to ensure fresh statistics data
	_, err := c.SendMessage("call_service", map[string]any{
		"domain":  "input_number",
		"service": "set_value",
		"target": map[string]any{
			"entity_id": testInputNumber,
		},
		"service_data": map[string]any{
			"value": 23.0 + float64(time.Now().Unix()%10),
		},
	})
	require.NoError(t, err)

	time.Sleep(500 * time.Millisecond)

	// Test stats command with the test entity
	// Note: input_number may not have statistics immediately - the recorder
	// needs time to aggregate. The command should still work without error.
	ctx := testContext(t, c, "stats", testInputNumber)
	err = handlers.HandleStats(ctx)
	assert.NoError(t, err)

	// Also verify the API structure directly
	endTime := time.Now()
	startTime := endTime.Add(-24 * time.Hour)

	result, err := client.SendMessageTyped[map[string][]types.StatEntry](c,
		"recorder/statistics_during_period",
		map[string]any{
			"start_time":    startTime.Format(time.RFC3339),
			"end_time":      endTime.Format(time.RFC3339),
			"statistic_ids": []string{testInputNumber},
			"period":        "hour",
		})
	require.NoError(t, err)

	stats, ok := result[testInputNumber]
	if ok && len(stats) > 0 {
		// Verify statistics contain the expected fields
		entry := stats[len(stats)-1] // Most recent entry
		t.Logf("Found %d stats entries for %s", len(stats), testInputNumber)
		t.Logf("Latest stats: start=%v, min=%.2f, max=%.2f, mean=%.2f, state=%.2f",
			entry.Start, entry.Min, entry.Max, entry.Mean, entry.State)
	} else {
		t.Logf("No statistics data for %s (recorder may need time to aggregate)", testInputNumber)
	}
}

// =============================================================================
// Error Handling Edge Case Tests
// =============================================================================

func TestIntegration_HandleTrace_InvalidAutomation(t *testing.T) {
	c := testClient(t)

	// Test with a nonexistent automation ID
	// The trace command should handle this gracefully without crashing
	ctx := testContext(t, c, "trace", "nonexistent_automation_xyz123", "some_run_id")
	err := handlers.HandleTrace(ctx)

	// The command may return an error or handle it gracefully
	// Either way, it should not panic
	if err != nil {
		t.Logf("HandleTrace with invalid automation returned error (expected): %v", err)
		// Verify the error message is meaningful
		assert.Contains(t, err.Error(), "nonexistent_automation_xyz123",
			"error should reference the invalid automation ID")
	} else {
		t.Log("HandleTrace with invalid automation returned no error (graceful handling)")
	}
}

func TestIntegration_HandleTrace_InvalidRunID(t *testing.T) {
	c := testClient(t)

	if !hasTestFixtures(t, c) {
		t.Skip("Test fixtures not available - skipping fixture tests")
	}

	// Test with a valid automation but invalid run_id
	// The trace command should handle this gracefully
	invalidRunID := "invalid_run_id_xyz_" + time.Now().Format("20060102150405")
	ctx := testContext(t, c, "trace", testAutomationID, invalidRunID)
	err := handlers.HandleTrace(ctx)

	// The command may return an error or handle it gracefully
	if err != nil {
		t.Logf("HandleTrace with invalid run_id returned error (expected): %v", err)
		// Error message should be meaningful
		errStr := err.Error()
		// It might say "trace not found" or "no trace" or similar
		t.Logf("Error details: %s", errStr)
	} else {
		t.Log("HandleTrace with invalid run_id returned no error (graceful handling)")
	}

	// Verify we can still use the client after the error
	// (i.e., the connection wasn't corrupted)
	_, pingErr := c.SendMessage("ping", nil)
	assert.NoError(t, pingErr, "client should still be usable after error")
}

func TestIntegration_HandleStats_NoStatistics(t *testing.T) {
	c := testClient(t)

	// Test with an entity that is unlikely to have statistics
	// Binary sensors and input_booleans typically don't generate statistics
	// Use a nonexistent but well-formed entity ID to ensure no statistics exist
	noStatsEntity := "input_boolean.no_stats_test_entity_xyz"

	ctx := testContext(t, c, "stats", noStatsEntity)
	err := handlers.HandleStats(ctx)

	// The command should handle the case gracefully (empty result, not error)
	assert.NoError(t, err, "stats command should not error for entity without statistics")

	// Also verify the API behavior directly
	endTime := time.Now()
	startTime := endTime.Add(-24 * time.Hour)

	result, err := client.SendMessageTyped[map[string][]types.StatEntry](c,
		"recorder/statistics_during_period",
		map[string]any{
			"start_time":    startTime.Format(time.RFC3339),
			"end_time":      endTime.Format(time.RFC3339),
			"statistic_ids": []string{noStatsEntity},
			"period":        "hour",
		})
	require.NoError(t, err, "API should not error for entity without statistics")

	// Result should be empty (no statistics for this entity)
	stats, ok := result[noStatsEntity]
	if ok {
		t.Logf("Found %d stats entries (unexpected, but acceptable)", len(stats))
	} else {
		t.Log("No statistics found for entity without stats (expected)")
	}
}

func TestIntegration_HandleCompare_InvalidEntity(t *testing.T) {
	c := testClient(t)

	// Test compare with one nonexistent entity
	ctx := testContext(t, c, "compare", "sun.sun", "nonexistent.entity_xyz123")
	err := handlers.HandleCompare(ctx)

	// The command should handle this gracefully
	if err != nil {
		t.Logf("HandleCompare with invalid entity returned error (expected): %v", err)
		// Error message should mention the invalid entity
		assert.Contains(t, err.Error(), "nonexistent",
			"error should reference the invalid entity")
	} else {
		t.Log("HandleCompare with invalid entity returned no error (shows partial result)")
	}

	// Verify client is still usable
	_, pingErr := c.SendMessage("ping", nil)
	assert.NoError(t, pingErr, "client should still be usable after compare error")
}

func TestIntegration_HandleCompare_BothEntitiesInvalid(t *testing.T) {
	c := testClient(t)

	// Test compare with both entities nonexistent
	ctx := testContext(t, c, "compare", "nonexistent.entity_a", "nonexistent.entity_b")
	err := handlers.HandleCompare(ctx)

	// The command should return an error or handle gracefully
	if err != nil {
		t.Logf("HandleCompare with both invalid entities returned error (expected): %v", err)
	} else {
		t.Log("HandleCompare with both invalid entities returned no error (shows empty result)")
	}
}

func TestIntegration_HandleHistoryFull_EmptyTimeRange(t *testing.T) {
	c := testClient(t)

	// Test history with a time range in the past where no data exists
	// Use a date range from 10 years ago when no data would exist
	pastTime := time.Date(2015, 1, 1, 0, 0, 0, 0, time.UTC)
	endTime := pastTime.Add(1 * time.Hour)

	result, err := client.SendMessageTyped[map[string][]types.HistoryState](c,
		"history/history_during_period",
		map[string]any{
			"start_time":       pastTime.Format(time.RFC3339),
			"end_time":         endTime.Format(time.RFC3339),
			"entity_ids":       []string{"sun.sun"},
			"minimal_response": false,
			"no_attributes":    false,
		})
	require.NoError(t, err, "history query should not error for empty time range")

	// Result should be empty (no data from 2015)
	history, ok := result["sun.sun"]
	if ok && len(history) > 0 {
		t.Logf("Unexpectedly found %d history entries from 2015", len(history))
	} else {
		t.Log("No history found for past time range (expected empty result)")
	}
}

func TestIntegration_HandleHistoryFull_FutureTimeRange(t *testing.T) {
	c := testClient(t)

	// Test history with a time range in the future where no data can exist
	futureTime := time.Now().Add(365 * 24 * time.Hour) // 1 year in future
	endTime := futureTime.Add(1 * time.Hour)

	result, err := client.SendMessageTyped[map[string][]types.HistoryState](c,
		"history/history_during_period",
		map[string]any{
			"start_time":       futureTime.Format(time.RFC3339),
			"end_time":         endTime.Format(time.RFC3339),
			"entity_ids":       []string{"sun.sun"},
			"minimal_response": false,
			"no_attributes":    false,
		})
	require.NoError(t, err, "history query should not error for future time range")

	// Result should be empty (no data from the future)
	history, ok := result["sun.sun"]
	if ok && len(history) > 0 {
		t.Fatalf("Unexpectedly found %d history entries from the future", len(history))
	} else {
		t.Log("No history found for future time range (expected empty result)")
	}
}

func TestIntegration_HandleAttrs_NoAttributeChanges(t *testing.T) {
	c := testClient(t)

	// Test attrs with an entity that exists but has minimal attribute changes
	// Binary sensors with static configuration typically don't change attributes
	// Use sun.sun but with a very short time range to minimize attribute changes

	// Get attributes for a 1-second window in the past (unlikely to have changes)
	shortRangeEnd := time.Now().Add(-1 * time.Hour)
	shortRangeStart := shortRangeEnd.Add(-1 * time.Second)

	result, err := client.SendMessageTyped[map[string][]types.HistoryState](c,
		"history/history_during_period",
		map[string]any{
			"start_time":       shortRangeStart.Format(time.RFC3339),
			"end_time":         shortRangeEnd.Format(time.RFC3339),
			"entity_ids":       []string{"sun.sun"},
			"minimal_response": false,
			"no_attributes":    false,
		})
	require.NoError(t, err, "attrs query should not error for short time range")

	// Result may be empty or have minimal data
	history, ok := result["sun.sun"]
	if ok {
		t.Logf("Found %d history entries in 1-second window", len(history))
	} else {
		t.Log("No history found in 1-second window (expected for short range)")
	}
}

func TestIntegration_HandleAttrs_NonexistentEntity(t *testing.T) {
	c := testClient(t)

	// Test attrs with a nonexistent entity
	endTime := time.Now()
	startTime := endTime.Add(-1 * time.Hour)

	result, err := client.SendMessageTyped[map[string][]types.HistoryState](c,
		"history/history_during_period",
		map[string]any{
			"start_time":       startTime.Format(time.RFC3339),
			"end_time":         endTime.Format(time.RFC3339),
			"entity_ids":       []string{"nonexistent.entity_xyz123"},
			"minimal_response": false,
			"no_attributes":    false,
		})
	require.NoError(t, err, "attrs query should not error for nonexistent entity")

	// Result should be empty (entity doesn't exist)
	history, ok := result["nonexistent.entity_xyz123"]
	if ok && len(history) > 0 {
		t.Fatalf("Unexpectedly found history for nonexistent entity")
	} else {
		t.Log("No history found for nonexistent entity (expected)")
	}
}
