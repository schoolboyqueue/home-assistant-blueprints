package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/home-assistant-blueprints/ha-ws-client-go/internal/client"
	"github.com/home-assistant-blueprints/ha-ws-client-go/internal/output"
	"github.com/home-assistant-blueprints/ha-ws-client-go/internal/types"
)

func init() {
	// Register history commands
	RegisterAll(
		Cmd("logbook", "Get logbook entries (default 24h)", "<entity_id> [hours]", "history", HandleLogbook),
		Cmd("history", "Get state history (default 24h)", "<entity_id> [hours]", "history", HandleHistory),
		Cmd("history-full", "Get history with full attributes", "<entity_id> [hours]", "history", HandleHistoryFull),
		Cmd("attrs", "Attribute change history (compact)", "<entity_id> [hours]", "history", HandleAttrs),
		Cmd("timeline", "Multi-entity chronological timeline", "<hours> <entity>...", "history", HandleTimeline),
		Cmd("syslog", "Get system log errors/warnings", "", "history", HandleSyslog),
		Cmd("stats", "Get sensor statistics (default 24h)", "<entity_id> [hours]", "history", HandleStats),
		Cmd("stats-multi", "Get statistics for multiple entities", "<entity>... [hours]", "history", HandleStatsMulti),
		Cmd("context", "Look up what triggered a state change", "<entity_id|context_id>", "history", HandleContext),
		Cmd("watch", "Live subscribe to state changes (default 60s)", "<entity_id> [seconds]", "history", HandleWatch),
	)
}

// HandleLogbook gets logbook entries for an entity.
// Wrapped with: Chain(RequireArg1(...), WithTimeRange(24, 2))
var HandleLogbook = Apply(
	Chain(
		RequireArg1("Usage: logbook <entity_id> [hours]"),
		WithTimeRange(24, 2),
	),
	handleLogbook,
)

func handleLogbook(ctx *Context) error {
	entityID := ctx.Config.Args[0]
	timeRange := *ctx.Config.TimeRange

	// Use logbook/get_events WebSocket message type
	entries, err := client.SendMessageTyped[[]types.LogbookEntry](ctx.Client, "logbook/get_events", map[string]any{
		"entity_ids": []string{entityID},
		"start_time": timeRange.StartTime.Format(time.RFC3339),
		"end_time":   timeRange.EndTime.Format(time.RFC3339),
	})
	if err != nil {
		return err
	}

	output.Timeline(entries,
		output.TimelineTitle[types.LogbookEntry](fmt.Sprintf("Logbook for %s", entityID)),
		output.TimelineCommand[types.LogbookEntry]("logbook"),
		output.TimelineFormatter(func(e types.LogbookEntry) string {
			t := time.Unix(int64(e.When), 0)
			if output.IsCompact() {
				return fmt.Sprintf("%s %s %s", t.Format(time.RFC3339), e.State, e.Message)
			}
			return fmt.Sprintf("%s: %s - %s", output.FormatTime(t), e.State, e.Message)
		}),
	)
	return nil
}

// HandleHistory gets state history for an entity.
// Wrapped with: Chain(RequireArg1(...), WithTimeRange(24, 2))
var HandleHistory = Apply(
	Chain(
		RequireArg1("Usage: history <entity_id> [hours]"),
		WithTimeRange(24, 2),
	),
	handleHistory,
)

func handleHistory(ctx *Context) error {
	entityID := ctx.Config.Args[0]
	timeRange := *ctx.Config.TimeRange

	// Use history/history_during_period WebSocket message type
	// Returns map[entity_id][]HistoryState
	result, err := client.SendMessageTyped[map[string][]types.HistoryState](ctx.Client, "history/history_during_period", map[string]any{
		"entity_ids":               []string{entityID},
		"start_time":               timeRange.StartTime.Format(time.RFC3339),
		"end_time":                 timeRange.EndTime.Format(time.RFC3339),
		"minimal_response":         true,
		"no_attributes":            true,
		"significant_changes_only": false,
	})
	if err != nil {
		return err
	}

	states := result[entityID]

	output.Timeline(states,
		output.TimelineTitle[types.HistoryState](fmt.Sprintf("History for %s", entityID)),
		output.TimelineCommand[types.HistoryState]("history"),
		output.TimelineFormatter(func(s types.HistoryState) string {
			t := s.GetLastUpdated()
			state := s.GetState()
			if output.IsCompact() {
				return fmt.Sprintf("%s %s", t.Format(time.RFC3339), state)
			}
			return fmt.Sprintf("%s: %s", output.FormatTime(t), state)
		}),
	)
	return nil
}

// HandleHistoryFull gets full history with attributes.
// Wrapped with: Chain(RequireArg1(...), WithTimeRange(24, 2))
var HandleHistoryFull = Apply(
	Chain(
		RequireArg1("Usage: history-full <entity_id> [hours]"),
		WithTimeRange(24, 2),
	),
	handleHistoryFull,
)

func handleHistoryFull(ctx *Context) error {
	entityID := ctx.Config.Args[0]
	timeRange := *ctx.Config.TimeRange

	// Use history/history_during_period WebSocket message type
	result, err := client.SendMessageTyped[map[string][]types.HistoryState](ctx.Client, "history/history_during_period", map[string]any{
		"entity_ids":       []string{entityID},
		"start_time":       timeRange.StartTime.Format(time.RFC3339),
		"end_time":         timeRange.EndTime.Format(time.RFC3339),
		"minimal_response": false,
		"no_attributes":    false,
	})
	if err != nil {
		return err
	}

	states := result[entityID]

	output.Timeline(states,
		output.TimelineTitle[types.HistoryState](fmt.Sprintf("Full history for %s", entityID)),
		output.TimelineCommand[types.HistoryState]("history-full"),
		output.TimelineFormatter(func(s types.HistoryState) string {
			t := s.GetLastUpdated()
			state := s.GetState()

			// Get attributes (handle both compact 'a' and full 'Attributes' formats)
			attrs := s.A
			if attrs == nil {
				attrs = s.Attributes
			}

			if output.IsCompact() {
				// Compact: timestamp state {key1:val1, key2:val2, ...}
				if len(attrs) == 0 {
					return fmt.Sprintf("%s %s", t.Format(time.RFC3339), state)
				}
				// Build abbreviated attrs string
				attrParts := make([]string, 0, len(attrs))
				for k, v := range attrs {
					attrParts = append(attrParts, fmt.Sprintf("%s:%v", k, v))
				}
				return fmt.Sprintf("%s %s {%s}", t.Format(time.RFC3339), state, strings.Join(attrParts, ", "))
			}

			// Default format: formatted time, state, and attributes on separate lines
			if len(attrs) == 0 {
				return fmt.Sprintf("%s: %s", output.FormatTime(t), state)
			}
			attrJSON, err := json.MarshalIndent(attrs, "    ", "  ")
			if err != nil {
				return fmt.Sprintf("%s: %s (attrs: %v)", output.FormatTime(t), state, attrs)
			}
			return fmt.Sprintf("%s: %s\n    %s", output.FormatTime(t), state, string(attrJSON))
		}),
	)
	return nil
}

// HandleAttrs gets attribute change history.
// Wrapped with: Chain(RequireArg1(...), WithTimeRange(24, 2))
var HandleAttrs = Apply(
	Chain(
		RequireArg1("Usage: attrs <entity_id> [hours]"),
		WithTimeRange(24, 2),
	),
	handleAttrs,
)

func handleAttrs(ctx *Context) error {
	entityID := ctx.Config.Args[0]
	timeRange := *ctx.Config.TimeRange

	// Use history/history_during_period WebSocket message type
	result, err := client.SendMessageTyped[map[string][]types.HistoryState](ctx.Client, "history/history_during_period", map[string]any{
		"entity_ids":       []string{entityID},
		"start_time":       timeRange.StartTime.Format(time.RFC3339),
		"end_time":         timeRange.EndTime.Format(time.RFC3339),
		"minimal_response": false,
		"no_attributes":    false,
	})
	if err != nil {
		return err
	}

	states := result[entityID]

	// Format as attribute changes
	type AttrChange struct {
		Time  time.Time
		State string
		Attrs map[string]any
	}

	changes := make([]AttrChange, 0, len(states))
	for _, s := range states {
		attrs := s.A
		if attrs == nil {
			attrs = s.Attributes
		}
		changes = append(changes, AttrChange{
			Time:  s.GetLastUpdated(),
			State: s.GetState(),
			Attrs: attrs,
		})
	}

	output.Timeline(changes,
		output.TimelineTitle[AttrChange](fmt.Sprintf("Attribute history for %s", entityID)),
		output.TimelineCommand[AttrChange]("attrs"),
		output.TimelineFormatter(func(c AttrChange) string {
			if output.IsCompact() {
				return fmt.Sprintf("%s %s", c.Time.Format(time.RFC3339), c.State)
			}
			return fmt.Sprintf("%s: %s %v", output.FormatTime(c.Time), c.State, c.Attrs)
		}),
	)
	return nil
}

// HandleTimeline shows multi-entity chronological timeline.
func HandleTimeline(ctx *Context) error {
	if len(ctx.Args) < 3 {
		return errors.New("missing arguments: timeline <hours> <entity>")
	}

	hours, err := strconv.Atoi(ctx.Args[1])
	if err != nil {
		return fmt.Errorf("invalid hours: %w", err)
	}

	entities := ctx.Args[2:]
	endTime := time.Now()
	startTime := endTime.Add(-time.Duration(hours) * time.Hour)

	type TimelineEntry struct {
		Time     time.Time
		EntityID string
		State    string
	}

	var allEntries []TimelineEntry

	// Use history/history_during_period with multiple entity_ids
	result, err := client.SendMessageTyped[map[string][]types.HistoryState](ctx.Client, "history/history_during_period", map[string]any{
		"entity_ids":       entities,
		"start_time":       startTime.Format(time.RFC3339),
		"end_time":         endTime.Format(time.RFC3339),
		"minimal_response": true,
		"no_attributes":    true,
	})
	if err != nil {
		return err
	}

	for entityID, states := range result {
		for _, s := range states {
			allEntries = append(allEntries, TimelineEntry{
				Time:     s.GetLastUpdated(),
				EntityID: entityID,
				State:    s.GetState(),
			})
		}
	}

	// Sort by time
	sort.Slice(allEntries, func(i, j int) bool {
		return allEntries[i].Time.Before(allEntries[j].Time)
	})

	output.Timeline(allEntries,
		output.TimelineTitle[TimelineEntry](fmt.Sprintf("Timeline for %d entities", len(entities))),
		output.TimelineCommand[TimelineEntry]("timeline"),
		output.TimelineFormatter(func(e TimelineEntry) string {
			if output.IsCompact() {
				return fmt.Sprintf("%s %s %s", e.Time.Format(time.RFC3339), e.EntityID, e.State)
			}
			return fmt.Sprintf("%s: %s = %s", output.FormatTime(e.Time), e.EntityID, e.State)
		}),
	)
	return nil
}

// HandleSyslog gets system log errors/warnings.
func HandleSyslog(ctx *Context) error {
	entries, err := client.SendMessageTyped[[]types.SysLogEntry](ctx.Client, "system_log/list", nil)
	if err != nil {
		return err
	}

	output.List(entries,
		output.ListTitle[types.SysLogEntry]("System log entries"),
		output.ListCommand[types.SysLogEntry]("syslog"),
		output.ListFormatter(func(e types.SysLogEntry, _ int) string {
			source := e.GetSource()
			msg := e.GetMessage()
			if output.IsCompact() {
				return fmt.Sprintf("[%s] %s: %s", e.Level, source, msg)
			}
			return fmt.Sprintf("[%s] %s\n  %s", e.Level, source, msg)
		}),
	)
	return nil
}

// HandleStats gets sensor statistics.
// Wrapped with: Chain(RequireArg1(...), WithTimeRange(24, 2))
var HandleStats = Apply(
	Chain(
		RequireArg1("Usage: stats <entity_id> [hours]"),
		WithTimeRange(24, 2),
	),
	handleStats,
)

func handleStats(ctx *Context) error {
	entityID := ctx.Config.Args[0]
	timeRange := *ctx.Config.TimeRange

	result, err := client.SendMessageTyped[map[string][]types.StatEntry](ctx.Client, "recorder/statistics_during_period", map[string]any{
		"start_time":    timeRange.StartTime.Format(time.RFC3339),
		"end_time":      timeRange.EndTime.Format(time.RFC3339),
		"statistic_ids": []string{entityID},
		"period":        "hour",
	})
	if err != nil {
		return err
	}

	stats, ok := result[entityID]
	if !ok || len(stats) == 0 {
		output.Message(fmt.Sprintf("No statistics found for %s", entityID))
		return nil
	}

	output.Timeline(stats,
		output.TimelineTitle[types.StatEntry](fmt.Sprintf("Statistics for %s", entityID)),
		output.TimelineCommand[types.StatEntry]("stats"),
		output.TimelineFormatter(func(s types.StatEntry) string {
			startTime := s.GetStartTime()
			if output.IsCompact() {
				return fmt.Sprintf("%s min=%.2f max=%.2f mean=%.2f", startTime, s.Min, s.Max, s.Mean)
			}
			return fmt.Sprintf("%s: min=%.2f, max=%.2f, mean=%.2f", startTime, s.Min, s.Max, s.Mean)
		}),
	)
	return nil
}

// EntityStatsSummary holds aggregated statistics for a single entity.
type EntityStatsSummary struct {
	EntityID   string  `json:"entity_id"`
	Min        float64 `json:"min"`
	Max        float64 `json:"max"`
	Mean       float64 `json:"mean"`
	Sum        float64 `json:"sum"`
	DataPoints int     `json:"data_points"`
	Error      string  `json:"error,omitempty"`
}

// HandleStatsMulti gets sensor statistics for multiple entities concurrently.
// Uses batch execution to fan-out statistics requests with error collection.
func HandleStatsMulti(ctx *Context) error {
	if len(ctx.Args) < 2 {
		return errors.New("usage: stats-multi <entity>... [hours]")
	}

	// Check if the last argument is a number (hours)
	args := ctx.Args[1:]
	hours := 24
	var entities []string

	if len(args) > 1 {
		if parsed, parseErr := strconv.Atoi(args[len(args)-1]); parseErr == nil {
			hours = parsed
			entities = args[:len(args)-1]
		} else {
			entities = args
		}
	} else {
		entities = args
	}

	if len(entities) == 0 {
		return errors.New("usage: stats-multi <entity>... [hours]")
	}

	endTime := time.Now()
	startTime := endTime.Add(-time.Duration(hours) * time.Hour)

	output.Message(fmt.Sprintf("Fetching statistics for %d entities (last %d hours, concurrent requests)...", len(entities), hours))

	// Use batch executor to fetch stats concurrently
	results := BatchExecutor(
		context.Background(),
		entities,
		func(e string) string { return e },
		func(_ context.Context, _ int, entityID string) (EntityStatsSummary, error) {
			result, err := client.SendMessageTyped[map[string][]types.StatEntry](ctx.Client, "recorder/statistics_during_period", map[string]any{
				"start_time":    startTime.Format(time.RFC3339),
				"end_time":      endTime.Format(time.RFC3339),
				"statistic_ids": []string{entityID},
				"period":        "hour",
			})
			if err != nil {
				return EntityStatsSummary{EntityID: entityID}, err
			}

			stats, ok := result[entityID]
			if !ok || len(stats) == 0 {
				return EntityStatsSummary{EntityID: entityID}, fmt.Errorf("no statistics found")
			}

			// Calculate aggregated statistics
			summary := EntityStatsSummary{
				EntityID:   entityID,
				Min:        stats[0].Min,
				Max:        stats[0].Max,
				DataPoints: len(stats),
			}

			var sum float64
			for _, s := range stats {
				if s.Min < summary.Min {
					summary.Min = s.Min
				}
				if s.Max > summary.Max {
					summary.Max = s.Max
				}
				sum += s.Mean
				summary.Sum += s.Sum
			}
			summary.Mean = sum / float64(len(stats))

			return summary, nil
		},
		BatchConfig{
			MaxConcurrency:  10, // Limit concurrent requests to avoid overwhelming the server
			ContinueOnError: true,
		},
	)

	// Collect successful results
	successful := results.Successful()
	failed := results.Failed()

	// Build output data
	var summaries []EntityStatsSummary
	for _, r := range successful {
		summaries = append(summaries, r.Result)
	}

	// Add failed entities with error messages
	for _, f := range failed {
		summaries = append(summaries, EntityStatsSummary{
			EntityID: f.Item,
			Error:    f.Err.Error(),
		})
	}

	// Sort by entity ID for consistent output
	sort.Slice(summaries, func(i, j int) bool {
		return summaries[i].EntityID < summaries[j].EntityID
	})

	if output.IsJSON() {
		output.Data(map[string]any{
			"entities":   summaries,
			"hours":      hours,
			"successful": len(successful),
			"failed":     len(failed),
		}, output.WithCommand("stats-multi"), output.WithCount(len(entities)))
	} else {
		fmt.Printf("\nStatistics Summary (%d hours):\n", hours)
		fmt.Printf("Successfully fetched: %d/%d entities\n\n", len(successful), len(entities))

		for _, s := range summaries {
			if s.Error != "" {
				fmt.Printf("  %s: ERROR - %s\n", s.EntityID, s.Error)
			} else {
				fmt.Printf("  %s:\n", s.EntityID)
				fmt.Printf("    min=%.2f, max=%.2f, mean=%.2f, sum=%.2f (%d data points)\n",
					s.Min, s.Max, s.Mean, s.Sum, s.DataPoints)
			}
		}
	}

	return nil
}

// HandleContext looks up what triggered a state change.
// Accepts either an entity_id or a context_id.
// Wrapped with: RequireArg1("Usage: context <entity_id|context_id>")
var HandleContext = Apply(
	RequireArg1("Usage: context <entity_id|context_id>"),
	handleContext,
)

// findEntityByID finds an entity in states by entity_id.
func findEntityByID(states []types.HAState, entityID string) *types.HAState {
	for i := range states {
		if states[i].EntityID == entityID {
			return &states[i]
		}
	}
	return nil
}

// findStatesByContext finds all states matching a context ID or having it as parent.
func findStatesByContext(states []types.HAState, contextID string) []types.HAState {
	var matches []types.HAState
	for _, s := range states {
		if s.Context != nil && (s.Context.ID == contextID || s.Context.ParentID == contextID) {
			matches = append(matches, s)
		}
	}
	return matches
}

// addParentContextMatches adds states matching the parent context if not already present.
func addParentContextMatches(matches, states []types.HAState, parentID string) []types.HAState {
	for _, s := range states {
		if s.Context == nil || s.Context.ID != parentID {
			continue
		}
		found := false
		for _, m := range matches {
			if m.EntityID == s.EntityID {
				found = true
				break
			}
		}
		if !found {
			matches = append(matches, s)
		}
	}
	return matches
}

// formatContextInfo formats context information for display.
func formatContextInfo(ctx *types.HAContext) string {
	if ctx == nil {
		return ""
	}
	if ctx.ParentID != "" {
		return fmt.Sprintf(" (context: %s, parent: %s)", ctx.ID, ctx.ParentID)
	}
	return fmt.Sprintf(" (context: %s)", ctx.ID)
}

func handleContext(ctx *Context) error {
	arg := ctx.Config.Args[0]

	states, err := client.SendMessageTyped[[]types.HAState](ctx.Client, "get_states", nil)
	if err != nil {
		return err
	}

	// Check if arg is an entity_id
	targetEntity := findEntityByID(states, arg)
	contextID := arg
	if targetEntity != nil && targetEntity.Context != nil {
		contextID = targetEntity.Context.ID
	}

	// Find matching states
	matches := findStatesByContext(states, contextID)

	// Add parent context matches if applicable
	if targetEntity != nil && targetEntity.Context != nil && targetEntity.Context.ParentID != "" {
		matches = addParentContextMatches(matches, states, targetEntity.Context.ParentID)
	}

	if len(matches) == 0 {
		return outputNoContextMatches(arg, contextID, targetEntity)
	}

	title := fmt.Sprintf("States with context %s", contextID)
	if targetEntity != nil {
		title = fmt.Sprintf("Related state changes for %s", arg)
	}

	output.List(matches,
		output.ListTitle[types.HAState](title),
		output.ListCommand[types.HAState]("context"),
		output.ListFormatter(func(s types.HAState, _ int) string {
			contextInfo := formatContextInfo(s.Context)
			if output.IsCompact() {
				return fmt.Sprintf("%s %s%s", s.EntityID, s.State, contextInfo)
			}
			return fmt.Sprintf("%s: %s%s", s.EntityID, s.State, contextInfo)
		}),
	)
	return nil
}

// outputNoContextMatches outputs appropriate message when no matches found.
func outputNoContextMatches(arg, contextID string, targetEntity *types.HAState) error {
	if targetEntity != nil {
		output.Message(fmt.Sprintf("No related state changes found for %s", arg))
		output.Message(fmt.Sprintf("Context ID: %s", contextID))
		if targetEntity.Context != nil && targetEntity.Context.ParentID != "" {
			output.Message(fmt.Sprintf("Parent context: %s", targetEntity.Context.ParentID))
		}
		output.Message("")
		output.Message("The triggering event may have occurred before the current state snapshot.")
		output.Message("Try 'logbook <entity_id>' to see recent history with context.")
	} else {
		output.Message(fmt.Sprintf("No states found with context ID: %s", contextID))
	}
	return nil
}

// HandleWatch subscribes to live state changes.
// Wrapped with: Chain(RequireArg1(...), WithOptionalInt(60, 2))
var HandleWatch = Apply(
	Chain(
		RequireArg1("Usage: watch <entity_id> [seconds]"),
		WithOptionalInt(60, 2),
	),
	handleWatch,
)

func handleWatch(ctx *Context) error {
	entityID := ctx.Config.Args[0]
	seconds := ctx.Config.OptionalInt

	output.Message(fmt.Sprintf("Watching %s for %d seconds...", entityID, seconds))

	trigger := map[string]any{
		"platform":  "state",
		"entity_id": entityID,
	}

	done := make(chan struct{})

	_, cleanup, err := ctx.Client.SubscribeToTrigger(trigger, func(vars map[string]any) {
		if triggerData, ok := vars["trigger"].(map[string]any); ok {
			if toState, ok := triggerData["to_state"].(map[string]any); ok {
				state := toState["state"]
				now := time.Now().Format(time.RFC3339)
				if output.IsJSON() {
					output.Data(map[string]any{
						"timestamp": now,
						"entity_id": entityID,
						"state":     state,
					})
				} else {
					fmt.Printf("%s: %s -> %v\n", output.FormatTime(time.Now()), entityID, state)
				}
			}
		}
	}, time.Duration(seconds)*time.Second)
	if err != nil {
		return err
	}

	defer cleanup()

	select {
	case <-done:
	case <-time.After(time.Duration(seconds) * time.Second):
	}

	output.Message("Watch complete")
	return nil
}
