package handlers

import (
	"fmt"
	"strings"

	"github.com/home-assistant-blueprints/ha-ws-client-go/internal/client"
	"github.com/home-assistant-blueprints/ha-ws-client-go/internal/output"
	"github.com/home-assistant-blueprints/ha-ws-client-go/internal/types"
)

// HandleTraces lists automation traces.
func HandleTraces(ctx *Context) error {
	var automationID string
	if len(ctx.Args) > 1 {
		automationID = ctx.Args[1]
	}

	data := map[string]any{"domain": "automation"}
	if automationID != "" {
		// Extract the automation ID from entity_id format
		id := strings.TrimPrefix(automationID, "automation.")
		data["item_id"] = id
	}

	traces, err := client.SendMessageTyped[map[string][]types.TraceInfo](ctx.Client, "trace/list", data)
	if err != nil {
		return err
	}

	// Flatten the traces map
	var allTraces []types.TraceInfo
	for _, traceList := range traces {
		allTraces = append(allTraces, traceList...)
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
func HandleTrace(ctx *Context) error {
	runID, err := RequireArg(ctx, 1, "Usage: trace <run_id> [automation_id]")
	if err != nil {
		return err
	}

	var automationID string
	if len(ctx.Args) > 2 {
		automationID = ctx.Args[2]
	}

	// If automation_id not provided, we need to find it
	if automationID == "" {
		// List all traces and find the one with matching run_id
		traces, lookupErr := client.SendMessageTyped[map[string][]types.TraceInfo](ctx.Client, "trace/list", map[string]any{"domain": "automation"})
		if lookupErr != nil {
			return lookupErr
		}

		for itemID, traceList := range traces {
			for _, t := range traceList {
				if t.RunID == runID {
					automationID = itemID
					break
				}
			}
			if automationID != "" {
				break
			}
		}

		if automationID == "" {
			return fmt.Errorf("trace not found: %s", runID)
		}
	}

	// Clean up automation ID
	id := strings.TrimPrefix(automationID, "automation.")

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

// HandleTraceVars shows evaluated variables from a trace.
func HandleTraceVars(ctx *Context) error {
	runID, err := RequireArg(ctx, 1, "Usage: trace-vars <run_id> [automation_id]")
	if err != nil {
		return err
	}

	trace, err := getTraceDetail(ctx, runID)
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
func HandleTraceTimeline(ctx *Context) error {
	runID, err := RequireArg(ctx, 1, "Usage: trace-timeline <run_id> [automation_id]")
	if err != nil {
		return err
	}

	trace, err := getTraceDetail(ctx, runID)
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
func HandleTraceTrigger(ctx *Context) error {
	runID, err := RequireArg(ctx, 1, "Usage: trace-trigger <run_id> [automation_id]")
	if err != nil {
		return err
	}

	trace, err := getTraceDetail(ctx, runID)
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
func HandleTraceActions(ctx *Context) error {
	runID, err := RequireArg(ctx, 1, "Usage: trace-actions <run_id> [automation_id]")
	if err != nil {
		return err
	}

	trace, err := getTraceDetail(ctx, runID)
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
func HandleTraceDebug(ctx *Context) error {
	runID, err := RequireArg(ctx, 1, "Usage: trace-debug <run_id> [automation_id]")
	if err != nil {
		return err
	}

	trace, err := getTraceDetail(ctx, runID)
	if err != nil {
		return err
	}

	// Output the complete trace data
	output.Data(trace, output.WithCommand("trace-debug"))
	return nil
}

// HandleAutomationConfig gets automation configuration.
func HandleAutomationConfig(ctx *Context) error {
	entityID, err := RequireArg(ctx, 1, "Usage: automation-config <entity_id>")
	if err != nil {
		return err
	}

	// Clean up entity ID
	id := strings.TrimPrefix(entityID, "automation.")

	result, err := client.SendMessageTyped[types.AutomationConfig](ctx.Client, "config/automation/config/"+id, nil)
	if err != nil {
		return err
	}

	output.Data(result, output.WithCommand("automation-config"))
	return nil
}

// HandleBlueprintInputs validates blueprint inputs.
func HandleBlueprintInputs(ctx *Context) error {
	entityID, err := RequireArg(ctx, 1, "Usage: blueprint-inputs <entity_id>")
	if err != nil {
		return err
	}

	// Clean up entity ID
	id := strings.TrimPrefix(entityID, "automation.")

	config, err := client.SendMessageTyped[types.AutomationConfig](ctx.Client, "config/automation/config/"+id, nil)
	if err != nil {
		return err
	}

	if config.UseBlueprint == nil {
		output.Message("This automation does not use a blueprint")
		return nil
	}

	output.Data(map[string]any{
		"blueprint_path": config.UseBlueprint.Path,
		"inputs":         config.UseBlueprint.Input,
	}, output.WithCommand("blueprint-inputs"))
	return nil
}

// getTraceDetail retrieves trace detail for a run.
func getTraceDetail(ctx *Context, runID string) (*types.TraceDetail, error) {
	var automationID string
	if len(ctx.Args) > 2 {
		automationID = ctx.Args[2]
	}

	// If automation_id not provided, we need to find it
	if automationID == "" {
		traces, lookupErr := client.SendMessageTyped[map[string][]types.TraceInfo](ctx.Client, "trace/list", map[string]any{"domain": "automation"})
		if lookupErr != nil {
			return nil, lookupErr
		}

		for itemID, traceList := range traces {
			for _, t := range traceList {
				if t.RunID == runID {
					automationID = itemID
					break
				}
			}
			if automationID != "" {
				break
			}
		}

		if automationID == "" {
			return nil, fmt.Errorf("trace not found: %s", runID)
		}
	}

	// Clean up automation ID
	id := strings.TrimPrefix(automationID, "automation.")

	trace, err := client.SendMessageTyped[types.TraceDetail](ctx.Client, "trace/get", map[string]any{
		"domain":  "automation",
		"item_id": id,
		"run_id":  runID,
	})
	if err != nil {
		return nil, err
	}

	return &trace, nil
}
