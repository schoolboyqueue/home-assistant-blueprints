package handlers

import (
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

// calculateTimeRange calculates the time range for history queries.
func calculateTimeRange(ctx *Context, defaultHours int) types.TimeRange {
	endTime := time.Now()
	if ctx.ToTime != nil {
		endTime = *ctx.ToTime
	}

	var startTime time.Time
	if ctx.FromTime != nil {
		startTime = *ctx.FromTime
	} else {
		hours := defaultHours
		if len(ctx.Args) > 2 {
			if h, err := strconv.Atoi(ctx.Args[2]); err == nil {
				hours = h
			}
		}
		startTime = endTime.Add(-time.Duration(hours) * time.Hour)
	}

	return types.TimeRange{StartTime: startTime, EndTime: endTime}
}

// HandleLogbook gets logbook entries for an entity.
func HandleLogbook(ctx *Context) error {
	entityID, err := RequireArg(ctx, 1, "Usage: logbook <entity_id> [hours]")
	if err != nil {
		return err
	}

	timeRange := calculateTimeRange(ctx, 24)

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
func HandleHistory(ctx *Context) error {
	entityID, err := RequireArg(ctx, 1, "Usage: history <entity_id> [hours]")
	if err != nil {
		return err
	}

	timeRange := calculateTimeRange(ctx, 24)

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
func HandleHistoryFull(ctx *Context) error {
	entityID, err := RequireArg(ctx, 1, "Usage: history-full <entity_id> [hours]")
	if err != nil {
		return err
	}

	timeRange := calculateTimeRange(ctx, 24)

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
func HandleAttrs(ctx *Context) error {
	entityID, err := RequireArg(ctx, 1, "Usage: attrs <entity_id> [hours]")
	if err != nil {
		return err
	}

	timeRange := calculateTimeRange(ctx, 24)

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
func HandleStats(ctx *Context) error {
	entityID, err := RequireArg(ctx, 1, "Usage: stats <entity_id> [hours]")
	if err != nil {
		return err
	}

	timeRange := calculateTimeRange(ctx, 24)

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

// HandleContext looks up what triggered a state change.
func HandleContext(ctx *Context) error {
	contextID, err := RequireArg(ctx, 1, "Usage: context <context_id>")
	if err != nil {
		return err
	}

	// Search for states with this context
	states, err := client.SendMessageTyped[[]types.HAState](ctx.Client, "get_states", nil)
	if err != nil {
		return err
	}

	var matches []types.HAState
	for _, s := range states {
		if s.Context != nil && (s.Context.ID == contextID || s.Context.ParentID == contextID) {
			matches = append(matches, s)
		}
	}

	if len(matches) == 0 {
		output.Message(fmt.Sprintf("No states found with context ID: %s", contextID))
		return nil
	}

	output.List(matches,
		output.ListTitle[types.HAState](fmt.Sprintf("States with context %s", contextID)),
		output.ListCommand[types.HAState]("context"),
		output.ListFormatter(func(s types.HAState, _ int) string {
			return fmt.Sprintf("%s: %s (context: %s)", s.EntityID, s.State, s.Context.ID)
		}),
	)
	return nil
}

// HandleWatch subscribes to live state changes.
func HandleWatch(ctx *Context) error {
	entityID, err := RequireArg(ctx, 1, "Usage: watch <entity_id> [seconds]")
	if err != nil {
		return err
	}

	seconds := 60
	if len(ctx.Args) > 2 {
		if parsed, parseErr := strconv.Atoi(ctx.Args[2]); parseErr == nil {
			seconds = parsed
		}
	}

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
