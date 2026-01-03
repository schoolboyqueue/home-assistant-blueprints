package handlers

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/home-assistant-blueprints/ha-ws-client-go/internal/output"
	"github.com/home-assistant-blueprints/ha-ws-client-go/internal/types"
)

// HandleMonitor monitors entity state changes.
// Wrapped with: Chain(RequireArg1(...), WithOptionalInt(60, 2))
var HandleMonitor = Apply(
	Chain(
		RequireArg1("Usage: monitor <entity_id> [seconds]"),
		WithOptionalInt(60, 2),
	),
	handleMonitor,
)

func handleMonitor(ctx *Context) error {
	entityID := ctx.Config.Args[0]
	seconds := ctx.Config.OptionalInt

	output.Message(fmt.Sprintf("Monitoring %s for %d seconds...", entityID, seconds))

	trigger := map[string]any{
		"platform":  "state",
		"entity_id": entityID,
	}

	changeCount := 0
	_, cleanup, err := ctx.Client.SubscribeToTrigger(trigger, func(vars map[string]any) {
		changeCount++
		if triggerData, ok := vars["trigger"].(map[string]any); ok {
			if toState, ok := triggerData["to_state"].(map[string]any); ok {
				state := toState["state"]
				fromState := ""
				if fs, ok := triggerData["from_state"].(map[string]any); ok {
					fromState = fmt.Sprintf("%v", fs["state"])
				}

				if output.IsJSON() {
					output.Data(map[string]any{
						"timestamp":  time.Now().Format(time.RFC3339),
						"entity_id":  entityID,
						"from_state": fromState,
						"to_state":   state,
						"change_num": changeCount,
					})
				} else {
					fmt.Printf("[%d] %s: %s -> %v\n", changeCount, output.FormatTime(time.Now()), fromState, state)
				}
			}
		}
	}, time.Duration(seconds)*time.Second)
	if err != nil {
		return err
	}

	defer cleanup()

	<-time.After(time.Duration(seconds) * time.Second)

	output.Message(fmt.Sprintf("Monitoring complete. %d state changes observed.", changeCount))
	return nil
}

// subscriptionResult holds the result of a subscription attempt.
type subscriptionResult struct {
	EntityID string
	Cleanup  func()
}

// HandleMonitorMulti monitors multiple entities concurrently.
// Uses batch execution to fan-out subscription requests with error collection.
func HandleMonitorMulti(ctx *Context) error {
	if len(ctx.Args) < 2 {
		return errors.New("usage: monitor-multi <entity>... [seconds]")
	}

	// Check if the last argument is a number (duration)
	args := ctx.Args[1:]
	seconds := 60
	var entities []string

	if len(args) > 1 {
		if parsed, parseErr := strconv.Atoi(args[len(args)-1]); parseErr == nil {
			seconds = parsed
			entities = args[:len(args)-1]
		} else {
			entities = args
		}
	} else {
		entities = args
	}

	if len(entities) == 0 {
		return errors.New("usage: monitor-multi <entity>... [seconds]")
	}

	output.Message(fmt.Sprintf("Monitoring %d entities for %d seconds (concurrent subscriptions)...", len(entities), seconds))

	// Thread-safe counter for state changes
	var changeCount atomic.Int64

	// Mutex for cleanups slice
	var cleanupsMu sync.Mutex
	cleanups := make([]func(), 0, len(entities))

	// Use batch executor to subscribe concurrently
	results := BatchExecutor(
		context.Background(),
		entities,
		func(e string) string { return e },
		func(_ context.Context, _ int, entityID string) (subscriptionResult, error) {
			trigger := map[string]any{
				"platform":  "state",
				"entity_id": entityID,
			}

			_, cleanup, err := ctx.Client.SubscribeToTrigger(trigger, func(vars map[string]any) {
				count := changeCount.Add(1)
				if triggerData, ok := vars["trigger"].(map[string]any); ok {
					if toState, ok := triggerData["to_state"].(map[string]any); ok {
						state := toState["state"]
						fromState := ""
						if fs, ok := triggerData["from_state"].(map[string]any); ok {
							fromState = fmt.Sprintf("%v", fs["state"])
						}

						if output.IsJSON() {
							output.Data(map[string]any{
								"timestamp":  time.Now().Format(time.RFC3339),
								"entity_id":  entityID,
								"from_state": fromState,
								"to_state":   state,
								"change_num": count,
							})
						} else {
							fmt.Printf("[%d] %s %s: %s -> %v\n", count, output.FormatTime(time.Now()), entityID, fromState, state)
						}
					}
				}
			}, time.Duration(seconds)*time.Second)
			if err != nil {
				return subscriptionResult{}, err
			}

			// Store cleanup function safely
			cleanupsMu.Lock()
			cleanups = append(cleanups, cleanup)
			cleanupsMu.Unlock()

			return subscriptionResult{EntityID: entityID, Cleanup: cleanup}, nil
		},
		BatchConfig{
			MaxConcurrency:  0, // No limit - subscribe to all concurrently
			ContinueOnError: true,
		},
	)

	// Defer cleanup of all successful subscriptions
	defer func() {
		cleanupsMu.Lock()
		defer cleanupsMu.Unlock()
		for _, cleanup := range cleanups {
			cleanup()
		}
	}()

	// Report any subscription failures
	failed := results.Failed()
	if len(failed) > 0 {
		for _, f := range failed {
			output.Message(fmt.Sprintf("Warning: failed to subscribe to %s: %v", f.Item, f.Err))
		}
	}

	successful := results.Successful()
	if len(successful) == 0 {
		return errors.New("failed to subscribe to any entities")
	}

	output.Message(fmt.Sprintf("Successfully subscribed to %d/%d entities", len(successful), len(entities)))

	// Wait for monitoring duration
	<-time.After(time.Duration(seconds) * time.Second)

	output.Message(fmt.Sprintf("Monitoring complete. %d state changes observed.", changeCount.Load()))
	return nil
}

// HandleAnalyze analyzes entity state patterns.
// Wrapped with: Chain(RequireArg1(...), WithTimeRange(24, 0))
var HandleAnalyze = Apply(
	Chain(
		RequireArg1("Usage: analyze <entity_id>"),
		WithTimeRange(24, 0), // No hours arg for analyze
	),
	handleAnalyze,
)

func handleAnalyze(ctx *Context) error {
	entityID := ctx.Config.Args[0]

	// Get current state
	states, err := getStates(ctx)
	if err != nil {
		return err
	}

	var currentState *struct {
		EntityID   string
		State      string
		Attributes map[string]any
	}

	for _, s := range states {
		if s.EntityID == entityID {
			currentState = &struct {
				EntityID   string
				State      string
				Attributes map[string]any
			}{
				EntityID:   s.EntityID,
				State:      s.State,
				Attributes: s.Attributes,
			}
			break
		}
	}

	if currentState == nil {
		return fmt.Errorf("entity not found: %s", entityID)
	}

	// Get recent history
	timeRange := *ctx.Config.TimeRange

	result, err := getHistory(ctx, entityID, timeRange)
	if err != nil {
		return err
	}

	// Calculate statistics
	stateChanges := len(result)
	stateCounts := make(map[string]int)
	for _, s := range result {
		state := s.GetState()
		stateCounts[state]++
	}

	analysis := map[string]any{
		"entity_id":          entityID,
		"current_state":      currentState.State,
		"attributes":         currentState.Attributes,
		"history_hours":      24,
		"state_changes":      stateChanges,
		"state_distribution": stateCounts,
	}

	output.Data(analysis, output.WithCommand("analyze"))
	return nil
}

// Helper functions

func getStates(ctx *Context) ([]struct {
	EntityID   string         `json:"entity_id"`
	State      string         `json:"state"`
	Attributes map[string]any `json:"attributes,omitempty"`
}, error,
) {
	resp, err := ctx.Client.SendMessage("get_states", nil)
	if err != nil {
		return nil, err
	}

	// Convert result to states
	statesData, ok := resp.Result.([]any)
	if !ok {
		return nil, errors.New("unexpected response type")
	}

	states := make([]struct {
		EntityID   string         `json:"entity_id"`
		State      string         `json:"state"`
		Attributes map[string]any `json:"attributes,omitempty"`
	}, 0, len(statesData))

	for _, s := range statesData {
		stateMap, ok := s.(map[string]any)
		if !ok {
			continue
		}

		entityID, _ := stateMap["entity_id"].(string)
		state, _ := stateMap["state"].(string)
		attrs, _ := stateMap["attributes"].(map[string]any)

		states = append(states, struct {
			EntityID   string         `json:"entity_id"`
			State      string         `json:"state"`
			Attributes map[string]any `json:"attributes,omitempty"`
		}{
			EntityID:   entityID,
			State:      state,
			Attributes: attrs,
		})
	}

	return states, nil
}

func getHistory(ctx *Context, entityID string, timeRange types.TimeRange) ([]types.HistoryState, error) {
	// Use history/history_during_period WebSocket message type
	resp, err := ctx.Client.SendMessage("history/history_during_period", map[string]any{
		"entity_ids":       []string{entityID},
		"start_time":       timeRange.StartTime.Format(time.RFC3339),
		"end_time":         timeRange.EndTime.Format(time.RFC3339),
		"minimal_response": true,
		"no_attributes":    true,
	})
	if err != nil {
		return nil, err
	}

	// Convert result to history states - response is map[entity_id][]state
	resultMap, ok := resp.Result.(map[string]any)
	if !ok {
		return nil, nil
	}

	statesArr, ok := resultMap[entityID].([]any)
	if !ok {
		return nil, nil
	}

	states := make([]types.HistoryState, 0, len(statesArr))
	for _, s := range statesArr {
		stateMap, ok := s.(map[string]any)
		if !ok {
			continue
		}

		hs := types.HistoryState{}
		if lu, ok := stateMap["lu"].(float64); ok {
			hs.LU = lu
		}
		if s, ok := stateMap["s"].(string); ok {
			hs.S = s
		}
		if state, ok := stateMap["state"].(string); ok {
			hs.State = state
		}
		if lastUpdated, ok := stateMap["last_updated"].(string); ok {
			hs.LastUpdated = lastUpdated
		}

		states = append(states, hs)
	}

	return states, nil
}
