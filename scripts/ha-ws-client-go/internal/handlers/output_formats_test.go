package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/home-assistant-blueprints/ha-ws-client-go/internal/output"
	"github.com/home-assistant-blueprints/ha-ws-client-go/internal/types"
)

// captureStdout captures stdout during function execution.
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	old := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w

	fn()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	_, err = io.Copy(&buf, r)
	require.NoError(t, err)
	return buf.String()
}

// OutputFormat represents an output format for testing.
type OutputFormat struct {
	Name   string
	Format output.Format
}

// AllFormats returns all output formats to test.
func AllFormats() []OutputFormat {
	return []OutputFormat{
		{Name: "default", Format: output.FormatDefault},
		{Name: "compact", Format: output.FormatCompact},
		{Name: "json", Format: output.FormatJSON},
	}
}

// setFormat sets the output format and returns a cleanup function.
func setFormat(t *testing.T, format output.Format) func() {
	t.Helper()
	original := output.GetConfig()
	output.SetConfig(&output.Config{
		Format:         format,
		ShowTimestamps: true,
		ShowHeaders:    true,
	})
	return func() { output.SetConfig(original) }
}

// =============================================================================
// State Output Tests (single entity)
// =============================================================================

func TestOutputFormat_State(t *testing.T) {
	state := types.HAState{
		EntityID:   "light.kitchen",
		State:      "on",
		Attributes: map[string]any{"brightness": 255, "friendly_name": "Kitchen Light"},
	}

	for _, of := range AllFormats() {
		t.Run(of.Name, func(t *testing.T) {
			cleanup := setFormat(t, of.Format)
			defer cleanup()

			out := captureStdout(t, func() {
				output.Data(state, output.WithCommand("state"))
			})

			switch of.Format {
			case output.FormatJSON:
				var result output.Result
				err := json.Unmarshal([]byte(out), &result)
				require.NoError(t, err, "JSON should be valid: %s", out)
				assert.True(t, result.Success)
				assert.Equal(t, "state", result.Command)
				data, ok := result.Data.(map[string]any)
				require.True(t, ok)
				assert.Equal(t, "light.kitchen", data["entity_id"])
				assert.Equal(t, "on", data["state"])

			case output.FormatCompact:
				assert.Equal(t, "light.kitchen=on\n", out)

			case output.FormatDefault:
				assert.Contains(t, out, "light.kitchen")
				assert.Contains(t, out, "on")
			}
		})
	}
}

// =============================================================================
// States List Output Tests
// =============================================================================

func TestOutputFormat_StatesList(t *testing.T) {
	states := []types.HAState{
		{EntityID: "light.kitchen", State: "on"},
		{EntityID: "light.bedroom", State: "off"},
		{EntityID: "sensor.temperature", State: "23.5"},
	}

	for _, of := range AllFormats() {
		t.Run(of.Name, func(t *testing.T) {
			cleanup := setFormat(t, of.Format)
			defer cleanup()

			out := captureStdout(t, func() {
				output.List(states,
					output.ListTitle[types.HAState]("Matching entities"),
					output.ListCommand[types.HAState]("states-filter"),
					output.ListFormatter(func(s types.HAState, _ int) string {
						if output.IsCompact() {
							return fmt.Sprintf("%s=%s", s.EntityID, s.State)
						}
						return fmt.Sprintf("%s: %s", s.EntityID, s.State)
					}),
				)
			})

			switch of.Format {
			case output.FormatJSON:
				var result output.Result
				err := json.Unmarshal([]byte(out), &result)
				require.NoError(t, err, "JSON should be valid: %s", out)
				assert.True(t, result.Success)
				assert.Equal(t, 3, result.Count)
				assert.Equal(t, "states-filter", result.Command)

			case output.FormatCompact:
				assert.Contains(t, out, "light.kitchen=on")
				assert.Contains(t, out, "light.bedroom=off")
				assert.Contains(t, out, "sensor.temperature=23.5")

			case output.FormatDefault:
				assert.Contains(t, out, "Matching entities: 3")
				assert.Contains(t, out, "light.kitchen: on")
			}
		})
	}
}

// =============================================================================
// History/Timeline Output Tests
// =============================================================================

func TestOutputFormat_History(t *testing.T) {
	refTime := time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC)
	history := []types.HistoryState{
		{S: "on", LU: float64(refTime.Unix())},
		{S: "off", LU: float64(refTime.Add(10 * time.Minute).Unix())},
		{S: "on", LU: float64(refTime.Add(20 * time.Minute).Unix())},
	}

	for _, of := range AllFormats() {
		t.Run(of.Name, func(t *testing.T) {
			cleanup := setFormat(t, of.Format)
			defer cleanup()

			out := captureStdout(t, func() {
				output.Timeline(history,
					output.TimelineTitle[types.HistoryState]("History for sensor.test"),
					output.TimelineCommand[types.HistoryState]("history"),
					output.TimelineFormatter(func(s types.HistoryState) string {
						ts := s.GetLastUpdated()
						state := s.GetState()
						if output.IsCompact() {
							return fmt.Sprintf("%s %s", ts.Format(time.RFC3339), state)
						}
						return fmt.Sprintf("%s: %s", output.FormatTime(ts), state)
					}),
				)
			})

			switch of.Format {
			case output.FormatJSON:
				var result output.Result
				err := json.Unmarshal([]byte(out), &result)
				require.NoError(t, err, "JSON should be valid: %s", out)
				assert.True(t, result.Success)
				assert.Equal(t, 3, result.Count)
				assert.Equal(t, "history", result.Command)

			case output.FormatCompact:
				assert.Contains(t, out, "on")
				assert.Contains(t, out, "off")
				// Should have RFC3339 timestamps
				assert.Contains(t, out, "T")

			case output.FormatDefault:
				assert.Contains(t, out, "History for sensor.test")
				assert.Contains(t, out, "Total: 3 entries")
			}
		})
	}
}

// =============================================================================
// Logbook Output Tests
// =============================================================================

func TestOutputFormat_Logbook(t *testing.T) {
	entries := []types.LogbookEntry{
		{When: 1718444400, EntityID: "light.kitchen", State: "on", Message: "turned on"},
		{When: 1718445000, EntityID: "light.kitchen", State: "off", Message: "turned off"},
	}

	for _, of := range AllFormats() {
		t.Run(of.Name, func(t *testing.T) {
			cleanup := setFormat(t, of.Format)
			defer cleanup()

			out := captureStdout(t, func() {
				output.Timeline(entries,
					output.TimelineTitle[types.LogbookEntry]("Logbook for light.kitchen"),
					output.TimelineCommand[types.LogbookEntry]("logbook"),
					output.TimelineFormatter(func(e types.LogbookEntry) string {
						ts := time.Unix(int64(e.When), 0)
						if output.IsCompact() {
							return fmt.Sprintf("%s %s %s", ts.Format(time.RFC3339), e.State, e.Message)
						}
						return fmt.Sprintf("%s: %s - %s", output.FormatTime(ts), e.State, e.Message)
					}),
				)
			})

			switch of.Format {
			case output.FormatJSON:
				var result output.Result
				err := json.Unmarshal([]byte(out), &result)
				require.NoError(t, err, "JSON should be valid: %s", out)
				assert.True(t, result.Success)
				assert.Equal(t, 2, result.Count)

			case output.FormatCompact:
				assert.Contains(t, out, "on")
				assert.Contains(t, out, "turned on")

			case output.FormatDefault:
				assert.Contains(t, out, "Logbook for light.kitchen")
			}
		})
	}
}

// =============================================================================
// Stats Output Tests
// =============================================================================

func TestOutputFormat_Stats(t *testing.T) {
	stats := []types.StatEntry{
		{Start: float64(1718444400000), Min: 20.5, Max: 25.5, Mean: 23.0},
		{Start: float64(1718448000000), Min: 21.0, Max: 26.0, Mean: 23.5},
	}

	for _, of := range AllFormats() {
		t.Run(of.Name, func(t *testing.T) {
			cleanup := setFormat(t, of.Format)
			defer cleanup()

			out := captureStdout(t, func() {
				output.List(stats,
					output.ListTitle[types.StatEntry]("Statistics for sensor.temperature"),
					output.ListCommand[types.StatEntry]("stats"),
					output.ListFormatter(func(s types.StatEntry, _ int) string {
						ts := s.GetStartTime()
						if output.IsCompact() {
							return fmt.Sprintf("%s min=%.1f max=%.1f mean=%.1f", ts, s.Min, s.Max, s.Mean)
						}
						return fmt.Sprintf("%s\n  Min: %.1f, Max: %.1f, Mean: %.1f", ts, s.Min, s.Max, s.Mean)
					}),
				)
			})

			switch of.Format {
			case output.FormatJSON:
				var result output.Result
				err := json.Unmarshal([]byte(out), &result)
				require.NoError(t, err, "JSON should be valid: %s", out)
				assert.True(t, result.Success)
				assert.Equal(t, 2, result.Count)

			case output.FormatCompact:
				assert.Contains(t, out, "min=20.5")
				assert.Contains(t, out, "max=25.5")

			case output.FormatDefault:
				assert.Contains(t, out, "Statistics for sensor.temperature")
			}
		})
	}
}

// =============================================================================
// Syslog Output Tests
// =============================================================================

func TestOutputFormat_Syslog(t *testing.T) {
	entries := []types.SysLogEntry{
		{Level: "ERROR", Name: "homeassistant.core", Message: "Test error message", Timestamp: 1718444400},
		{Level: "WARNING", Name: "custom_component", Message: "Test warning", Timestamp: 1718445000},
	}

	for _, of := range AllFormats() {
		t.Run(of.Name, func(t *testing.T) {
			cleanup := setFormat(t, of.Format)
			defer cleanup()

			out := captureStdout(t, func() {
				output.List(entries,
					output.ListTitle[types.SysLogEntry]("System log"),
					output.ListCommand[types.SysLogEntry]("syslog"),
					output.ListFormatter(func(e types.SysLogEntry, _ int) string {
						ts := time.Unix(int64(e.Timestamp), 0)
						msg := e.GetMessage()
						if output.IsCompact() {
							return fmt.Sprintf("%s [%s] %s: %s", ts.Format(time.RFC3339), e.Level, e.Name, msg)
						}
						return fmt.Sprintf("[%s] %s\n  Source: %s\n  Time: %s", e.Level, msg, e.Name, output.FormatTime(ts))
					}),
				)
			})

			switch of.Format {
			case output.FormatJSON:
				var result output.Result
				err := json.Unmarshal([]byte(out), &result)
				require.NoError(t, err, "JSON should be valid: %s", out)
				assert.True(t, result.Success)
				assert.Equal(t, 2, result.Count)

			case output.FormatCompact:
				assert.Contains(t, out, "[ERROR]")
				assert.Contains(t, out, "Test error message")

			case output.FormatDefault:
				assert.Contains(t, out, "System log")
			}
		})
	}
}

// =============================================================================
// Entity Registry Output Tests
// =============================================================================

func TestOutputFormat_Entities(t *testing.T) {
	entities := []types.EntityEntry{
		{EntityID: "light.kitchen", Name: "Kitchen Light", Platform: "hue"},
		{EntityID: "sensor.temperature", Name: "Temperature", Platform: "esphome"},
	}

	for _, of := range AllFormats() {
		t.Run(of.Name, func(t *testing.T) {
			cleanup := setFormat(t, of.Format)
			defer cleanup()

			out := captureStdout(t, func() {
				output.List(entities,
					output.ListTitle[types.EntityEntry]("Entity registry"),
					output.ListCommand[types.EntityEntry]("entities"),
					output.ListFormatter(func(e types.EntityEntry, _ int) string {
						if output.IsCompact() {
							return fmt.Sprintf("%s %s [%s]", e.EntityID, e.Name, e.Platform)
						}
						return fmt.Sprintf("%s\n  Name: %s\n  Platform: %s", e.EntityID, e.Name, e.Platform)
					}),
				)
			})

			switch of.Format {
			case output.FormatJSON:
				var result output.Result
				err := json.Unmarshal([]byte(out), &result)
				require.NoError(t, err, "JSON should be valid: %s", out)
				assert.True(t, result.Success)
				assert.Equal(t, 2, result.Count)

			case output.FormatCompact:
				assert.Contains(t, out, "light.kitchen")
				assert.Contains(t, out, "[hue]")

			case output.FormatDefault:
				assert.Contains(t, out, "Entity registry: 2")
			}
		})
	}
}

// =============================================================================
// Device Registry Output Tests
// =============================================================================

func TestOutputFormat_Devices(t *testing.T) {
	devices := []types.DeviceEntry{
		{ID: "abc123", Name: "Kitchen Hue Bridge", Manufacturer: "Philips", Model: "BSB002"},
		{ID: "def456", Name: "Temperature Sensor", Manufacturer: "ESPHome", Model: "ESP32"},
	}

	for _, of := range AllFormats() {
		t.Run(of.Name, func(t *testing.T) {
			cleanup := setFormat(t, of.Format)
			defer cleanup()

			out := captureStdout(t, func() {
				output.List(devices,
					output.ListTitle[types.DeviceEntry]("Device registry"),
					output.ListCommand[types.DeviceEntry]("devices"),
					output.ListFormatter(func(d types.DeviceEntry, _ int) string {
						if output.IsCompact() {
							return fmt.Sprintf("%s %s %s/%s", d.ID, d.Name, d.Manufacturer, d.Model)
						}
						return fmt.Sprintf("%s\n  Name: %s\n  Manufacturer: %s\n  Model: %s", d.ID, d.Name, d.Manufacturer, d.Model)
					}),
				)
			})

			switch of.Format {
			case output.FormatJSON:
				var result output.Result
				err := json.Unmarshal([]byte(out), &result)
				require.NoError(t, err, "JSON should be valid: %s", out)
				assert.True(t, result.Success)
				assert.Equal(t, 2, result.Count)

			case output.FormatCompact:
				assert.Contains(t, out, "abc123")
				assert.Contains(t, out, "Philips/BSB002")

			case output.FormatDefault:
				assert.Contains(t, out, "Device registry: 2")
			}
		})
	}
}

// =============================================================================
// Area Registry Output Tests
// =============================================================================

func TestOutputFormat_Areas(t *testing.T) {
	areas := []types.AreaEntry{
		{AreaID: "kitchen", Name: "Kitchen"},
		{AreaID: "living_room", Name: "Living Room"},
		{AreaID: "bedroom", Name: "Bedroom"},
	}

	for _, of := range AllFormats() {
		t.Run(of.Name, func(t *testing.T) {
			cleanup := setFormat(t, of.Format)
			defer cleanup()

			out := captureStdout(t, func() {
				output.List(areas,
					output.ListTitle[types.AreaEntry]("Area registry"),
					output.ListCommand[types.AreaEntry]("areas"),
					output.ListFormatter(func(a types.AreaEntry, _ int) string {
						if output.IsCompact() {
							return fmt.Sprintf("%s: %s", a.AreaID, a.Name)
						}
						return fmt.Sprintf("%s (%s)", a.Name, a.AreaID)
					}),
				)
			})

			switch of.Format {
			case output.FormatJSON:
				var result output.Result
				err := json.Unmarshal([]byte(out), &result)
				require.NoError(t, err, "JSON should be valid: %s", out)
				assert.True(t, result.Success)
				assert.Equal(t, 3, result.Count)

			case output.FormatCompact:
				assert.Contains(t, out, "kitchen: Kitchen")
				assert.Contains(t, out, "living_room: Living Room")

			case output.FormatDefault:
				assert.Contains(t, out, "Area registry: 3")
			}
		})
	}
}

// =============================================================================
// Trace List Output Tests
// =============================================================================

func TestOutputFormat_Traces(t *testing.T) {
	traces := []types.TraceInfo{
		{
			ItemID:          "kitchen_lights",
			RunID:           "run123",
			ScriptExecution: "finished",
			Timestamp:       &types.Timestamp{Start: "2024-06-15T10:00:00Z"},
		},
		{
			ItemID:          "kitchen_lights",
			RunID:           "run456",
			ScriptExecution: "error",
			Timestamp:       &types.Timestamp{Start: "2024-06-15T09:00:00Z"},
		},
	}

	for _, of := range AllFormats() {
		t.Run(of.Name, func(t *testing.T) {
			cleanup := setFormat(t, of.Format)
			defer cleanup()

			out := captureStdout(t, func() {
				output.List(traces,
					output.ListTitle[types.TraceInfo]("Automation traces"),
					output.ListCommand[types.TraceInfo]("traces"),
					output.ListFormatter(func(tr types.TraceInfo, _ int) string {
						state := tr.ScriptExecution
						if state == "" {
							state = tr.State
						}
						timestamp := ""
						if tr.Timestamp != nil {
							timestamp = tr.Timestamp.Start
						}
						if output.IsCompact() {
							return fmt.Sprintf("%s %s %s %s", tr.ItemID, tr.RunID, state, timestamp)
						}
						return fmt.Sprintf("automation.%s\n  Run ID: %s\n  State: %s\n  Started: %s",
							tr.ItemID, tr.RunID, state, timestamp)
					}),
				)
			})

			switch of.Format {
			case output.FormatJSON:
				var result output.Result
				err := json.Unmarshal([]byte(out), &result)
				require.NoError(t, err, "JSON should be valid: %s", out)
				assert.True(t, result.Success)
				assert.Equal(t, 2, result.Count)

			case output.FormatCompact:
				assert.Contains(t, out, "kitchen_lights run123 finished")
				assert.Contains(t, out, "kitchen_lights run456 error")

			case output.FormatDefault:
				assert.Contains(t, out, "Automation traces: 2")
				assert.Contains(t, out, "automation.kitchen_lights")
			}
		})
	}
}

// =============================================================================
// Trace Detail Output Tests
// =============================================================================

func TestOutputFormat_TraceDetail(t *testing.T) {
	trace := types.TraceDetail{
		ItemID:          "kitchen_lights",
		RunID:           "run123",
		Domain:          "automation",
		ScriptExecution: "finished",
		Timestamp:       &types.Timestamp{Start: "2024-06-15T10:00:00Z", Finish: "2024-06-15T10:00:01Z"},
		Trigger:         "state of light.kitchen",
	}

	for _, of := range AllFormats() {
		t.Run(of.Name, func(t *testing.T) {
			cleanup := setFormat(t, of.Format)
			defer cleanup()

			out := captureStdout(t, func() {
				output.Data(trace, output.WithCommand("trace"))
			})

			switch of.Format {
			case output.FormatJSON:
				var result output.Result
				err := json.Unmarshal([]byte(out), &result)
				require.NoError(t, err, "JSON should be valid: %s", out)
				assert.True(t, result.Success)
				assert.Equal(t, "trace", result.Command)

			case output.FormatCompact:
				// Compact mode for struct uses JSON marshaling
				assert.Contains(t, out, "kitchen_lights")

			case output.FormatDefault:
				assert.Contains(t, out, "kitchen_lights")
				assert.Contains(t, out, "finished")
			}
		})
	}
}

// =============================================================================
// Config Output Tests
// =============================================================================

func TestOutputFormat_Config(t *testing.T) {
	config := types.HAConfig{
		Version:      "2024.6.0",
		LocationName: "Home",
		TimeZone:     "America/New_York",
		State:        "RUNNING",
	}

	for _, of := range AllFormats() {
		t.Run(of.Name, func(t *testing.T) {
			cleanup := setFormat(t, of.Format)
			defer cleanup()

			out := captureStdout(t, func() {
				output.Data(config, output.WithCommand("config"))
			})

			switch of.Format {
			case output.FormatJSON:
				var result output.Result
				err := json.Unmarshal([]byte(out), &result)
				require.NoError(t, err, "JSON should be valid: %s", out)
				assert.True(t, result.Success)
				assert.Equal(t, "config", result.Command)

			case output.FormatCompact:
				// Struct gets converted to map
				assert.Contains(t, out, "2024.6.0")

			case output.FormatDefault:
				assert.Contains(t, out, "2024.6.0")
				assert.Contains(t, out, "Home")
			}
		})
	}
}

// =============================================================================
// Message Output Tests (ping, etc.)
// =============================================================================

func TestOutputFormat_Message(t *testing.T) {
	for _, of := range AllFormats() {
		t.Run(of.Name, func(t *testing.T) {
			cleanup := setFormat(t, of.Format)
			defer cleanup()

			out := captureStdout(t, func() {
				output.Message("pong")
			})

			switch of.Format {
			case output.FormatJSON:
				var result output.Result
				err := json.Unmarshal([]byte(out), &result)
				require.NoError(t, err, "JSON should be valid: %s", out)
				assert.True(t, result.Success)
				assert.Equal(t, "pong", result.Message)

			case output.FormatCompact:
				assert.Equal(t, "pong\n", out)

			case output.FormatDefault:
				assert.Equal(t, "pong\n", out)
			}
		})
	}
}

// =============================================================================
// Entity Output Tests (using output.Entity helper)
// =============================================================================

func TestOutputFormat_Entity(t *testing.T) {
	for _, of := range AllFormats() {
		t.Run(of.Name, func(t *testing.T) {
			cleanup := setFormat(t, of.Format)
			defer cleanup()

			out := captureStdout(t, func() {
				output.Entity("sensor.temperature", "23.5", map[string]any{"unit_of_measurement": "°C"})
			})

			switch of.Format {
			case output.FormatJSON:
				var result output.Result
				err := json.Unmarshal([]byte(out), &result)
				require.NoError(t, err, "JSON should be valid: %s", out)
				assert.True(t, result.Success)
				data, ok := result.Data.(map[string]any)
				require.True(t, ok)
				assert.Equal(t, "sensor.temperature", data["entity_id"])
				assert.Equal(t, "23.5", data["state"])

			case output.FormatCompact:
				assert.Equal(t, "sensor.temperature=23.5\n", out)

			case output.FormatDefault:
				assert.Contains(t, out, "sensor.temperature")
				assert.Contains(t, out, "23.5")
			}
		})
	}
}

// =============================================================================
// Empty Results Tests
// =============================================================================

func TestOutputFormat_EmptyList(t *testing.T) {
	for _, of := range AllFormats() {
		t.Run(of.Name, func(t *testing.T) {
			cleanup := setFormat(t, of.Format)
			defer cleanup()

			out := captureStdout(t, func() {
				output.List([]types.HAState{},
					output.ListTitle[types.HAState]("No matching entities"),
					output.ListCommand[types.HAState]("states-filter"),
				)
			})

			switch of.Format {
			case output.FormatJSON:
				var result output.Result
				err := json.Unmarshal([]byte(out), &result)
				require.NoError(t, err, "JSON should be valid: %s", out)
				assert.True(t, result.Success)
				assert.Equal(t, 0, result.Count)

			case output.FormatCompact:
				assert.Contains(t, out, "No matching entities: 0")

			case output.FormatDefault:
				assert.Contains(t, out, "No matching entities: 0")
			}
		})
	}
}

func TestOutputFormat_EmptyTimeline(t *testing.T) {
	for _, of := range AllFormats() {
		t.Run(of.Name, func(t *testing.T) {
			cleanup := setFormat(t, of.Format)
			defer cleanup()

			out := captureStdout(t, func() {
				output.Timeline([]types.HistoryState{},
					output.TimelineTitle[types.HistoryState]("No history"),
					output.TimelineCommand[types.HistoryState]("history"),
				)
			})

			switch of.Format {
			case output.FormatJSON:
				var result output.Result
				err := json.Unmarshal([]byte(out), &result)
				require.NoError(t, err, "JSON should be valid: %s", out)
				assert.True(t, result.Success)
				assert.Equal(t, 0, result.Count)

			case output.FormatCompact:
				assert.Contains(t, out, "No history: 0")

			case output.FormatDefault:
				assert.Contains(t, out, "Total: 0 entries")
			}
		})
	}
}

// =============================================================================
// MaxItems Tests
// =============================================================================

func TestOutputFormat_MaxItems(t *testing.T) {
	states := make([]types.HAState, 10)
	for i := range 10 {
		states[i] = types.HAState{
			EntityID: fmt.Sprintf("sensor.test_%d", i),
			State:    strconv.Itoa(i),
		}
	}

	for _, of := range AllFormats() {
		t.Run(of.Name, func(t *testing.T) {
			original := output.GetConfig()
			output.SetConfig(&output.Config{
				Format:      of.Format,
				ShowHeaders: true,
				MaxItems:    3,
			})
			defer output.SetConfig(original)

			out := captureStdout(t, func() {
				output.List(states,
					output.ListTitle[types.HAState]("Test entities"),
					output.ListCommand[types.HAState]("states-filter"),
					output.ListFormatter(func(s types.HAState, _ int) string {
						return s.EntityID + "=" + s.State
					}),
				)
			})

			switch of.Format {
			case output.FormatJSON:
				// JSON still contains all items
				var result output.Result
				err := json.Unmarshal([]byte(out), &result)
				require.NoError(t, err)
				assert.Equal(t, 10, result.Count)

			case output.FormatCompact:
				assert.Contains(t, out, "+7 more")

			case output.FormatDefault:
				assert.Contains(t, out, "... and 7 more")
			}
		})
	}
}

// =============================================================================
// Traces No-Traces JSON Output Test (regression test for multi-line fix)
// =============================================================================

func TestOutputFormat_TracesNoTracesJSON(t *testing.T) {
	// This tests the fix for traces --json outputting multiple JSON lines
	cleanup := setFormat(t, output.FormatJSON)
	defer cleanup()

	out := captureStdout(t, func() {
		// Simulate the fixed traces handler behavior when no traces exist
		output.Data(map[string]any{
			"entity_id":      "automation.test",
			"traces":         []any{},
			"last_triggered": "2024-06-15T10:00:00Z",
			"message":        "No stored traces. Traces may be disabled or cleared.",
		}, output.WithCommand("traces"), output.WithCount(0))
	})

	// Should be a single JSON line, not multiple
	var result output.Result
	err := json.Unmarshal([]byte(out), &result)
	require.NoError(t, err, "should be valid JSON: %s", out)
	assert.True(t, result.Success)
	assert.Equal(t, "traces", result.Command)

	data, ok := result.Data.(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "automation.test", data["entity_id"])
	assert.Equal(t, "2024-06-15T10:00:00Z", data["last_triggered"])
}

// =============================================================================
// Typed Struct Compact Output Test (regression test for struct output fix)
// =============================================================================

func TestOutputFormat_TypedStructCompact(t *testing.T) {
	// This tests the fix for state --compact outputting raw Go struct
	cleanup := setFormat(t, output.FormatCompact)
	defer cleanup()

	state := types.HAState{
		EntityID:    "sun.sun",
		State:       "below_horizon",
		Attributes:  map[string]any{"azimuth": 285.4},
		LastChanged: "2024-06-15T10:00:00Z",
	}

	out := captureStdout(t, func() {
		output.Data(state, output.WithCommand("state"))
	})

	// Should be formatted as entity_id=state, not raw Go struct
	assert.Equal(t, "sun.sun=below_horizon\n", out)
	assert.NotContains(t, out, "map[")
	assert.NotContains(t, out, "0x")
}

// =============================================================================
// Compare Output Tests
// =============================================================================

func TestOutputFormat_Compare(t *testing.T) {
	comparison := map[string]any{
		"entity1": map[string]any{
			"entity_id": "light.kitchen",
			"state":     "on",
		},
		"entity2": map[string]any{
			"entity_id": "light.bedroom",
			"state":     "off",
		},
		"differences": []string{"state"},
	}

	for _, of := range AllFormats() {
		t.Run(of.Name, func(t *testing.T) {
			cleanup := setFormat(t, of.Format)
			defer cleanup()

			out := captureStdout(t, func() {
				output.Data(comparison, output.WithCommand("compare"))
			})

			switch of.Format {
			case output.FormatJSON:
				var result output.Result
				err := json.Unmarshal([]byte(out), &result)
				require.NoError(t, err, "JSON should be valid: %s", out)
				assert.True(t, result.Success)
				assert.Equal(t, "compare", result.Command)

			case output.FormatCompact:
				assert.Contains(t, out, "entity1")
				assert.Contains(t, out, "entity2")

			case output.FormatDefault:
				assert.Contains(t, out, "light.kitchen")
				assert.Contains(t, out, "light.bedroom")
			}
		})
	}
}

// =============================================================================
// Device Health Output Tests
// =============================================================================

func TestOutputFormat_DeviceHealth(t *testing.T) {
	health := map[string]any{
		"entity_id":    "sensor.temperature",
		"status":       "ok",
		"last_updated": "2024-06-15T10:00:00Z",
		"age_seconds":  120,
	}

	for _, of := range AllFormats() {
		t.Run(of.Name, func(t *testing.T) {
			cleanup := setFormat(t, of.Format)
			defer cleanup()

			out := captureStdout(t, func() {
				output.Data(health, output.WithCommand("device-health"))
			})

			switch of.Format {
			case output.FormatJSON:
				var result output.Result
				err := json.Unmarshal([]byte(out), &result)
				require.NoError(t, err, "JSON should be valid: %s", out)
				assert.True(t, result.Success)
				assert.Equal(t, "device-health", result.Command)

			case output.FormatCompact:
				assert.Contains(t, out, "status")
				assert.Contains(t, out, "ok")

			case output.FormatDefault:
				assert.Contains(t, out, "sensor.temperature")
				assert.Contains(t, out, "ok")
			}
		})
	}
}

// =============================================================================
// Automation Config Output Tests
// =============================================================================

func TestOutputFormat_AutomationConfig(t *testing.T) {
	config := types.AutomationConfig{
		ID:    "kitchen_lights",
		Alias: "Kitchen Lights",
		Trigger: []any{
			map[string]any{"platform": "state", "entity_id": "binary_sensor.motion"},
		},
		Action: []any{
			map[string]any{"service": "light.turn_on", "target": map[string]any{"entity_id": "light.kitchen"}},
		},
	}

	for _, of := range AllFormats() {
		t.Run(of.Name, func(t *testing.T) {
			cleanup := setFormat(t, of.Format)
			defer cleanup()

			out := captureStdout(t, func() {
				output.Data(config, output.WithCommand("automation-config"))
			})

			switch of.Format {
			case output.FormatJSON:
				var result output.Result
				err := json.Unmarshal([]byte(out), &result)
				require.NoError(t, err, "JSON should be valid: %s", out)
				assert.True(t, result.Success)
				assert.Equal(t, "automation-config", result.Command)

			case output.FormatCompact:
				assert.Contains(t, out, "kitchen_lights")

			case output.FormatDefault:
				assert.Contains(t, out, "Kitchen Lights")
			}
		})
	}
}

// =============================================================================
// Analyze Output Tests
// =============================================================================

func TestOutputFormat_Analyze(t *testing.T) {
	analysis := map[string]any{
		"entity_id":            "sensor.temperature",
		"total_changes":        42,
		"avg_changes_per_hour": 1.75,
		"most_common_state":    "23.5",
		"state_distribution": map[string]int{
			"22.0": 5,
			"23.0": 20,
			"23.5": 17,
		},
	}

	for _, of := range AllFormats() {
		t.Run(of.Name, func(t *testing.T) {
			cleanup := setFormat(t, of.Format)
			defer cleanup()

			out := captureStdout(t, func() {
				output.Data(analysis, output.WithCommand("analyze"))
			})

			switch of.Format {
			case output.FormatJSON:
				var result output.Result
				err := json.Unmarshal([]byte(out), &result)
				require.NoError(t, err, "JSON should be valid: %s", out)
				assert.True(t, result.Success)
				assert.Equal(t, "analyze", result.Command)

			case output.FormatCompact:
				assert.Contains(t, out, "total_changes")

			case output.FormatDefault:
				assert.Contains(t, out, "sensor.temperature")
			}
		})
	}
}
