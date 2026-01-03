package handlers

import (
	"fmt"
	"strings"
	"time"

	"github.com/home-assistant-blueprints/ha-ws-client-go/internal/client"
	"github.com/home-assistant-blueprints/ha-ws-client-go/internal/output"
	"github.com/home-assistant-blueprints/ha-ws-client-go/internal/types"
)

// HandleTraces lists automation traces.
// Wrapped with: WithOptionalAutomationID(1)
var HandleTraces = Apply(
	WithOptionalAutomationID(1),
	handleTraces,
)

func handleTraces(ctx *Context) error {
	automationID := ctx.Config.AutomationID

	data := map[string]any{"domain": "automation"}
	if automationID != "" {
		data["item_id"] = automationID
	}

	// HA returns traces as an array of TraceInfo, not a map
	allTraces, err := client.SendMessageTyped[[]types.TraceInfo](ctx.Client, "trace/list", data)
	if err != nil {
		return err
	}

	// Filter by --from time if specified
	var filteredTraces []types.TraceInfo
	if ctx.FromTime != nil {
		for _, t := range allTraces {
			if t.Timestamp != nil && t.Timestamp.Start != "" {
				traceTime, parseErr := time.Parse(time.RFC3339, t.Timestamp.Start)
				if parseErr == nil && traceTime.After(*ctx.FromTime) {
					filteredTraces = append(filteredTraces, t)
				}
			}
		}
		allTraces = filteredTraces
	}

	// If filtering by specific automation and no traces found, check if it has last_triggered
	if len(allTraces) == 0 && automationID != "" {
		entityID := EnsureAutomationPrefix(automationID)
		states, statesErr := client.SendMessageTyped[[]types.HAState](ctx.Client, "get_states", nil)
		if statesErr == nil {
			for _, s := range states {
				if s.EntityID == entityID {
					if lastTriggered, ok := s.Attributes["last_triggered"].(string); ok && lastTriggered != "" {
						if output.IsJSON() {
							// Output a single JSON object with all the info
							output.Data(map[string]any{
								"entity_id":      entityID,
								"traces":         []any{},
								"last_triggered": lastTriggered,
								"message":        "No stored traces. Traces may be disabled or cleared.",
							}, output.WithCommand("traces"), output.WithCount(0))
						} else {
							output.Message(fmt.Sprintf("No stored traces for %s", entityID))
							output.Message(fmt.Sprintf("However, last_triggered: %s", lastTriggered))
							output.Message("")
							output.Message("Traces may be disabled or cleared. Check:")
							output.Message("  - Settings > Automations > (automation) > Stored Traces")
							output.Message("  - Trace storage limit (default: 5)")
						}
						return nil
					}
					break
				}
			}
		}
	}

	output.List(allTraces,
		output.ListTitle[types.TraceInfo]("Automation traces"),
		output.ListCommand[types.TraceInfo]("traces"),
		output.ListFormatter(func(t types.TraceInfo, _ int) string {
			state := t.State
			if t.ScriptExecution != "" {
				state = t.ScriptExecution
			}
			timestamp := ""
			if t.Timestamp != nil {
				timestamp = t.Timestamp.Start
			}
			if output.IsCompact() {
				return fmt.Sprintf("%s %s %s %s", t.ItemID, t.RunID, state, timestamp)
			}
			return fmt.Sprintf("automation.%s\n  Run ID: %s\n  State: %s\n  Started: %s",
				t.ItemID, t.RunID, state, timestamp)
		}),
	)
	return nil
}

// HandleTrace gets detailed trace for a run.
// Wrapped with: Chain(WithAutomationID(1, ...), RequireArg2(...))
var HandleTrace = Apply(
	Chain(
		WithAutomationID(1, "Usage: trace <automation_id> <run_id>"),
		RequireArg2("Usage: trace <automation_id> <run_id>"),
	),
	handleTrace,
)

func handleTrace(ctx *Context) error {
	id := ctx.Config.AutomationID
	runID := ctx.Config.Args[1]

	trace, err := client.SendMessageTyped[types.TraceDetail](ctx.Client, "trace/get", map[string]any{
		"domain":  "automation",
		"item_id": id,
		"run_id":  runID,
	})
	if err != nil {
		return err
	}

	output.Data(trace, output.WithCommand("trace"))
	return nil
}

// HandleTraceLatest gets the most recent trace for an automation.
// Wrapped with: WithAutomationID(1, "Usage: trace-latest <automation_id>")
var HandleTraceLatest = Apply(
	WithAutomationID(1, "Usage: trace-latest <automation_id>"),
	handleTraceLatest,
)

func handleTraceLatest(ctx *Context) error {
	id := ctx.Config.AutomationID
	traces, err := client.SendMessageTyped[[]types.TraceInfo](ctx.Client, "trace/list", map[string]any{
		"domain":  "automation",
		"item_id": id,
	})
	if err != nil {
		return err
	}

	if len(traces) == 0 {
		return fmt.Errorf("no traces found for automation.%s", id)
	}

	// Get the most recent trace (first in list)
	latest := traces[0]
	trace, err := getTraceDetail(ctx.Client, id, latest.RunID)
	if err != nil {
		return err
	}

	output.Data(trace, output.WithCommand("trace-latest"))
	return nil
}

// HandleTraceSummary shows a quick overview of recent automation runs.
// Wrapped with: WithAutomationID(1, "Usage: trace-summary <automation_id>")
var HandleTraceSummary = Apply(
	WithAutomationID(1, "Usage: trace-summary <automation_id>"),
	handleTraceSummary,
)

func handleTraceSummary(ctx *Context) error {
	id := ctx.Config.AutomationID
	traces, err := client.SendMessageTyped[[]types.TraceInfo](ctx.Client, "trace/list", map[string]any{
		"domain":  "automation",
		"item_id": id,
	})
	if err != nil {
		return err
	}

	if len(traces) == 0 {
		output.Message(fmt.Sprintf("No traces found for automation.%s", id))
		return nil
	}

	// Count states
	finished, errors, other := 0, 0, 0
	for _, t := range traces {
		state := t.ScriptExecution
		if state == "" {
			state = t.State
		}
		switch state {
		case "finished":
			finished++
		case "error":
			errors++
		default:
			other++
		}
	}

	// Get details from the most recent trace
	latest := traces[0]
	trace, err := getTraceDetail(ctx.Client, id, latest.RunID)
	if err != nil {
		return err
	}

	// Build summary
	summary := map[string]any{
		"automation_id": "automation." + id,
		"total_traces":  len(traces),
		"finished":      finished,
		"errors":        errors,
		"other":         other,
		"last_run": map[string]any{
			"run_id":    latest.RunID,
			"state":     latest.ScriptExecution,
			"timestamp": latest.Timestamp,
		},
		"trigger": trace.GetTriggerDescription(),
	}

	// Add error info if present
	if trace.Error != "" {
		summary["last_error"] = trace.Error
	}

	output.Data(summary, output.WithCommand("trace-summary"))
	return nil
}

// HandleTraceVars shows evaluated variables from a trace.
// Wrapped with: Chain(WithAutomationID(1, ...), RequireArg2(...))
var HandleTraceVars = Apply(
	Chain(
		WithAutomationID(1, "Usage: trace-vars <automation_id> <run_id>"),
		RequireArg2("Usage: trace-vars <automation_id> <run_id>"),
	),
	handleTraceVars,
)

func handleTraceVars(ctx *Context) error {
	automationID := ctx.Config.AutomationID
	runID := ctx.Config.Args[1]

	trace, err := getTraceDetail(ctx.Client, automationID, runID)
	if err != nil {
		return err
	}

	// Collect all variables from trace steps
	var allVars []map[string]any
	if trace.Trace != nil {
		for path, steps := range trace.Trace {
			for _, step := range steps {
				if step.Variables != nil {
					allVars = append(allVars, map[string]any{
						"path":      path,
						"variables": step.Variables,
					})
				}
				if step.ChangedVariables != nil {
					allVars = append(allVars, map[string]any{
						"path":              path,
						"changed_variables": step.ChangedVariables,
					})
				}
			}
		}
	}

	output.List(allVars,
		output.ListTitle[map[string]any](fmt.Sprintf("Variables for trace %s", runID)),
		output.ListCommand[map[string]any]("trace-vars"),
	)
	return nil
}

// HandleTraceTimeline shows step-by-step execution timeline.
// Wrapped with: Chain(WithAutomationID(1, ...), RequireArg2(...))
var HandleTraceTimeline = Apply(
	Chain(
		WithAutomationID(1, "Usage: trace-timeline <automation_id> <run_id>"),
		RequireArg2("Usage: trace-timeline <automation_id> <run_id>"),
	),
	handleTraceTimeline,
)

func handleTraceTimeline(ctx *Context) error {
	automationID := ctx.Config.AutomationID
	runID := ctx.Config.Args[1]

	trace, err := getTraceDetail(ctx.Client, automationID, runID)
	if err != nil {
		return err
	}

	type TimelineStep struct {
		Timestamp string
		Path      string
		Error     any
	}

	var steps []TimelineStep
	if trace.Trace != nil {
		for path, traceSteps := range trace.Trace {
			for _, step := range traceSteps {
				steps = append(steps, TimelineStep{
					Timestamp: step.Timestamp,
					Path:      path,
					Error:     step.Error,
				})
			}
		}
	}

	output.Timeline(steps,
		output.TimelineTitle[TimelineStep](fmt.Sprintf("Execution timeline for %s", runID)),
		output.TimelineCommand[TimelineStep]("trace-timeline"),
		output.TimelineFormatter(func(s TimelineStep) string {
			errStr := ""
			if s.Error != nil {
				errStr = fmt.Sprintf(" ERROR: %v", s.Error)
			}
			return fmt.Sprintf("%s: %s%s", s.Timestamp, s.Path, errStr)
		}),
	)
	return nil
}

// HandleTraceTrigger shows trigger context details.
// Wrapped with: Chain(WithAutomationID(1, ...), RequireArg2(...))
var HandleTraceTrigger = Apply(
	Chain(
		WithAutomationID(1, "Usage: trace-trigger <automation_id> <run_id>"),
		RequireArg2("Usage: trace-trigger <automation_id> <run_id>"),
	),
	handleTraceTrigger,
)

func handleTraceTrigger(ctx *Context) error {
	automationID := ctx.Config.AutomationID
	runID := ctx.Config.Args[1]

	trace, err := getTraceDetail(ctx.Client, automationID, runID)
	if err != nil {
		return err
	}

	if trace.Trigger == nil {
		output.Message("No trigger information available")
		return nil
	}

	output.Data(trace.Trigger, output.WithCommand("trace-trigger"))
	return nil
}

// HandleTraceActions shows action results.
// Wrapped with: Chain(WithAutomationID(1, ...), RequireArg2(...))
var HandleTraceActions = Apply(
	Chain(
		WithAutomationID(1, "Usage: trace-actions <automation_id> <run_id>"),
		RequireArg2("Usage: trace-actions <automation_id> <run_id>"),
	),
	handleTraceActions,
)

func handleTraceActions(ctx *Context) error {
	automationID := ctx.Config.AutomationID
	runID := ctx.Config.Args[1]

	trace, err := getTraceDetail(ctx.Client, automationID, runID)
	if err != nil {
		return err
	}

	type ActionResult struct {
		Path   string
		Result *types.TraceResult
	}

	var actions []ActionResult
	if trace.Trace != nil {
		for path, steps := range trace.Trace {
			if strings.HasPrefix(path, "action/") {
				for _, step := range steps {
					actions = append(actions, ActionResult{
						Path:   path,
						Result: step.Result,
					})
				}
			}
		}
	}

	output.List(actions,
		output.ListTitle[ActionResult](fmt.Sprintf("Actions for trace %s", runID)),
		output.ListCommand[ActionResult]("trace-actions"),
		output.ListFormatter(func(a ActionResult, _ int) string {
			if output.IsCompact() {
				return a.Path
			}
			result := "no result"
			if a.Result != nil && a.Result.Response != nil {
				result = fmt.Sprintf("%v", a.Result.Response)
			}
			return fmt.Sprintf("%s: %s", a.Path, result)
		}),
	)
	return nil
}

// HandleTraceDebug shows comprehensive debug view.
// Wrapped with: Chain(WithAutomationID(1, ...), RequireArg2(...))
var HandleTraceDebug = Apply(
	Chain(
		WithAutomationID(1, "Usage: trace-debug <automation_id> <run_id>"),
		RequireArg2("Usage: trace-debug <automation_id> <run_id>"),
	),
	handleTraceDebug,
)

func handleTraceDebug(ctx *Context) error {
	automationID := ctx.Config.AutomationID
	runID := ctx.Config.Args[1]

	trace, err := getTraceDetail(ctx.Client, automationID, runID)
	if err != nil {
		return err
	}

	// Output the complete trace data
	output.Data(trace, output.WithCommand("trace-debug"))
	return nil
}

// HandleAutomationConfig gets automation configuration.
// Wrapped with: RequireArg1("Usage: automation-config <entity_id>")
var HandleAutomationConfig = Apply(
	RequireArg1("Usage: automation-config <entity_id>"),
	handleAutomationConfig,
)

func handleAutomationConfig(ctx *Context) error {
	entityID := EnsureAutomationPrefix(ctx.Config.Args[0])
	automationID := strings.TrimPrefix(entityID, "automation.")

	// First try to get config from a trace (more complete for blueprint automations)
	traces, err := client.SendMessageTyped[[]types.TraceInfo](ctx.Client, "trace/list", map[string]any{
		"domain":  "automation",
		"item_id": automationID,
	})
	if err == nil && len(traces) > 0 {
		// Get the most recent trace which contains the resolved config
		trace, traceErr := getTraceDetail(ctx.Client, automationID, traces[0].RunID)
		if traceErr == nil && trace.Config != nil {
			// Check if trace config has actual content (trigger/action)
			if len(trace.Config.Trigger) > 0 || len(trace.Config.Action) > 0 {
				output.Data(trace.Config, output.WithCommand("automation-config"))
				return nil
			}
		}
	}

	// Fall back to automation/config API
	type configResponse struct {
		Config types.AutomationConfig `json:"config"`
	}
	result, apiErr := client.SendMessageTyped[configResponse](ctx.Client, "automation/config", map[string]any{
		"entity_id": entityID,
	})
	if apiErr != nil {
		return apiErr
	}

	// Check if we got a meaningful config
	if len(result.Config.Trigger) == 0 && len(result.Config.Action) == 0 && result.Config.UseBlueprint == nil {
		// Minimal config returned - likely a blueprint automation with no traces
		output.Message(fmt.Sprintf("Limited config available for %s (blueprint automation with no stored traces).", entityID))
		output.Message("Run the automation once to generate a trace, then use 'trace-latest' or 'trace-debug' to see the resolved config.")
		output.Message("")
	}

	output.Data(result.Config, output.WithCommand("automation-config"))
	return nil
}

// HandleBlueprintInputs validates blueprint inputs.
// Wrapped with: RequireArg1("Usage: blueprint-inputs <entity_id>")
var HandleBlueprintInputs = Apply(
	RequireArg1("Usage: blueprint-inputs <entity_id>"),
	handleBlueprintInputs,
)

func handleBlueprintInputs(ctx *Context) error {
	entityID := EnsureAutomationPrefix(ctx.Config.Args[0])

	// Use automation/config WebSocket message type
	type configResponse struct {
		Config types.AutomationConfig `json:"config"`
	}
	result, err := client.SendMessageTyped[configResponse](ctx.Client, "automation/config", map[string]any{
		"entity_id": entityID,
	})
	if err != nil {
		return err
	}

	if result.Config.UseBlueprint == nil {
		output.Message("Blueprint info not available (HA API limitation for blueprint automations). Check automations.yaml directly or use trace-vars to see resolved values.")
		return nil
	}

	output.Data(map[string]any{
		"blueprint_path": result.Config.UseBlueprint.Path,
		"inputs":         result.Config.UseBlueprint.Input,
	}, output.WithCommand("blueprint-inputs"))
	return nil
}

// getTraceDetail retrieves trace detail for a run.
func getTraceDetail(c *client.Client, automationID, runID string) (*types.TraceDetail, error) {
	// Clean up automation ID
	id := strings.TrimPrefix(automationID, "automation.")

	trace, err := client.SendMessageTyped[types.TraceDetail](c, "trace/get", map[string]any{
		"domain":  "automation",
		"item_id": id,
		"run_id":  runID,
	})
	if err != nil {
		return nil, err
	}

	return &trace, nil
}
