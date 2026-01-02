// Package handlers provides command handlers for the CLI.
package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/home-assistant-blueprints/ha-ws-client-go/internal/client"
	"github.com/home-assistant-blueprints/ha-ws-client-go/internal/output"
	"github.com/home-assistant-blueprints/ha-ws-client-go/internal/types"
)

// Context holds the execution context for handlers.
type Context struct {
	Client   *client.Client
	Args     []string
	FromTime *time.Time
	ToTime   *time.Time
}

// ErrEntityNotFound indicates an entity was not found.
var ErrEntityNotFound = errors.New("entity not found")

// Status constants for device health checks.
const (
	statusOK      = "ok"
	statusStale   = "stale"
	statusUnknown = "unknown"
)

// RequireArg returns the argument at index or errors with usage.
func RequireArg(ctx *Context, index int, usage string) (string, error) {
	if index >= len(ctx.Args) {
		return "", fmt.Errorf("missing argument: %s", usage)
	}
	return ctx.Args[index], nil
}

// HandlePing tests the WebSocket connection.
func HandlePing(ctx *Context) error {
	start := time.Now()
	_, err := ctx.Client.SendMessage("ping", nil)
	if err != nil {
		return err
	}
	latency := time.Since(start)

	if output.IsJSON() {
		output.Data(map[string]any{"latency_ms": latency.Milliseconds()}, output.WithCommand("ping"))
	} else {
		output.Message(fmt.Sprintf("Pong! (%dms)", latency.Milliseconds()))
	}
	return nil
}

// HandleState gets the state of a single entity.
func HandleState(ctx *Context) error {
	entityID, err := RequireArg(ctx, 1, "Usage: state <entity_id>")
	if err != nil {
		return err
	}

	states, err := client.SendMessageTyped[[]types.HAState](ctx.Client, "get_states", nil)
	if err != nil {
		return err
	}

	for _, state := range states {
		if state.EntityID == entityID {
			output.Data(state, output.WithCommand("state"))
			return nil
		}
	}

	return fmt.Errorf("%w: %s", ErrEntityNotFound, entityID)
}

// HandleStates gets a summary of all entity states.
func HandleStates(ctx *Context) error {
	states, err := client.SendMessageTyped[[]types.HAState](ctx.Client, "get_states", nil)
	if err != nil {
		return err
	}

	cfg := output.GetConfig()
	limit := 10
	if cfg.MaxItems > 0 {
		limit = cfg.MaxItems
	}

	sample := states
	if len(states) > limit {
		sample = states[:limit]
	}

	if output.IsJSON() {
		output.Data(map[string]any{
			"total":  len(states),
			"sample": sample,
		}, output.WithCommand("states"), output.WithCount(len(states)))
	} else {
		output.List(sample,
			output.ListTitle[types.HAState](fmt.Sprintf("Total entities: %d\nSample", len(states))),
			output.ListFormatter(func(s types.HAState, _ int) string {
				return fmt.Sprintf("  %s: %s", s.EntityID, s.State)
			}),
		)
	}
	return nil
}

// HandleStatesJSON gets all entity states as JSON.
func HandleStatesJSON(ctx *Context) error {
	states, err := client.SendMessageTyped[[]types.HAState](ctx.Client, "get_states", nil)
	if err != nil {
		return err
	}

	output.Data(states, output.WithCommand("states-json"), output.WithCount(len(states)))
	return nil
}

// HandleStatesFilter filters states by entity_id pattern.
func HandleStatesFilter(ctx *Context) error {
	pattern, err := RequireArg(ctx, 1, "Usage: states-filter <pattern> [--show-age]")
	if err != nil {
		return err
	}

	states, err := client.SendMessageTyped[[]types.HAState](ctx.Client, "get_states", nil)
	if err != nil {
		return err
	}

	// Convert glob pattern to regex
	regexPattern := regexp.QuoteMeta(pattern)
	regexPattern = regexp.MustCompile(`\\\*`).ReplaceAllString(regexPattern, ".*")
	re, err := regexp.Compile(regexPattern)
	if err != nil {
		return fmt.Errorf("invalid pattern: %w", err)
	}

	var filtered []types.HAState
	for _, s := range states {
		if re.MatchString(s.EntityID) {
			filtered = append(filtered, s)
		}
	}

	now := time.Now()
	staleThreshold := 1 * time.Hour

	output.List(filtered,
		output.ListTitle[types.HAState](fmt.Sprintf("Found %d matching entities", len(filtered))),
		output.ListCommand[types.HAState]("states-filter"),
		output.ListFormatter(func(s types.HAState, _ int) string {
			if output.ShowAge() {
				age := ""
				status := ""
				if s.LastUpdated != "" {
					if t, parseErr := time.Parse(time.RFC3339, s.LastUpdated); parseErr == nil {
						d := now.Sub(t)
						age = formatDuration(d)
						if d > staleThreshold {
							status = " ⚠️"
						} else {
							status = " ✓"
						}
					}
				}
				return fmt.Sprintf("%s: %s (%s ago)%s", s.EntityID, s.State, age, status)
			}
			return fmt.Sprintf("%s: %s", s.EntityID, s.State)
		}),
	)
	return nil
}

// HandleConfig gets Home Assistant configuration.
func HandleConfig(ctx *Context) error {
	config, err := client.SendMessageTyped[types.HAConfig](ctx.Client, "get_config", nil)
	if err != nil {
		return err
	}

	summary := map[string]any{
		"version":          config.Version,
		"location_name":    config.LocationName,
		"time_zone":        config.TimeZone,
		"unit_system":      config.UnitSystem,
		"state":            config.State,
		"components_count": len(config.Components),
	}

	output.Data(summary, output.WithCommand("config"))
	return nil
}

// HandleServices lists all available services.
func HandleServices(ctx *Context) error {
	services, err := client.SendMessageTyped[map[string]map[string]any](ctx.Client, "get_services", nil)
	if err != nil {
		return err
	}

	domains := make([]string, 0, len(services))
	for domain := range services {
		domains = append(domains, domain)
	}

	if output.IsJSON() {
		var data []map[string]any
		for _, domain := range domains {
			svcNames := make([]string, 0)
			for svcName := range services[domain] {
				svcNames = append(svcNames, svcName)
			}
			data = append(data, map[string]any{
				"domain":   domain,
				"services": svcNames,
			})
		}
		output.Data(data, output.WithCommand("services"), output.WithCount(len(domains)))
	} else {
		output.List(domains,
			output.ListTitle[string]("Domains"),
			output.ListCommand[string]("services"),
			output.ListFormatter(func(domain string, _ int) string {
				svcNames := make([]string, 0)
				for svcName := range services[domain] {
					svcNames = append(svcNames, svcName)
				}
				return fmt.Sprintf("  %s: %s", domain, strings.Join(svcNames, ", "))
			}),
		)
	}
	return nil
}

// HandleCall calls a Home Assistant service.
func HandleCall(ctx *Context) error {
	domain, err := RequireArg(ctx, 1, "Usage: call <domain> <service> [data]")
	if err != nil {
		return err
	}
	service, err := RequireArg(ctx, 2, "Usage: call <domain> <service> [data]")
	if err != nil {
		return err
	}

	var serviceData map[string]any
	if len(ctx.Args) > 3 {
		if unmarshalErr := json.Unmarshal([]byte(ctx.Args[3]), &serviceData); unmarshalErr != nil {
			return fmt.Errorf("invalid JSON data: %w", unmarshalErr)
		}
	}

	result, err := ctx.Client.SendMessage("call_service", map[string]any{
		"domain":       domain,
		"service":      service,
		"service_data": serviceData,
	})
	if err != nil {
		return err
	}

	if output.IsJSON() {
		output.Data(map[string]any{
			"domain":       domain,
			"service":      service,
			"service_data": serviceData,
			"response":     result.Result,
		}, output.WithCommand("call"), output.WithSummary("Service called successfully"))
	} else {
		output.Message("Service called successfully")
		if result.Result != nil {
			resultMap, ok := result.Result.(map[string]any)
			if ok && len(resultMap) > 0 {
				output.Data(result.Result, output.WithSummary("Response:"))
			}
		}
	}
	return nil
}

// HandleTemplate renders a Jinja2 template.
func HandleTemplate(ctx *Context) error {
	var template string

	if len(ctx.Args) > 1 {
		template = ctx.Args[1]
	}

	// Read from stdin if no argument or "-"
	if template == "" || template == "-" {
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("failed to read from stdin: %w", err)
		}
		template = string(data)
	}

	if template == "" {
		return errors.New("usage: template <template>\n  Or pipe template via stdin")
	}

	// render_template is subscription-based - we subscribe and wait for the first result
	resultChan := make(chan string, 1)
	errChan := make(chan error, 1)

	_, cleanup, err := ctx.Client.SubscribeToTemplate(template, func(result string) {
		select {
		case resultChan <- result:
		default:
		}
	}, 5*time.Second)
	if err != nil {
		return err
	}
	defer cleanup()

	// Wait for result or timeout
	select {
	case result := <-resultChan:
		if output.IsJSON() {
			output.Data(map[string]any{
				"template": template,
				"result":   result,
			}, output.WithCommand("template"))
		} else {
			output.Message(result)
		}
	case err := <-errChan:
		return err
	case <-time.After(5 * time.Second):
		return errors.New("template render timeout")
	}

	return nil
}

// HandleDeviceHealth checks if a device/entity is responsive by examining last_updated times.
func HandleDeviceHealth(ctx *Context) error {
	entityID, err := RequireArg(ctx, 1, "Usage: device-health <entity_id>")
	if err != nil {
		return err
	}

	states, err := client.SendMessageTyped[[]types.HAState](ctx.Client, "get_states", nil)
	if err != nil {
		return err
	}

	// Extract base device name from entity_id for finding related entities
	// e.g., "cover.guest_bedroom_window_shade" -> "guest_bedroom_window_shade"
	parts := strings.SplitN(entityID, ".", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid entity_id format: %s", entityID)
	}
	baseName := parts[1]

	// Find the main entity and related entities
	var mainEntity *types.HAState
	var relatedEntities []types.HAState

	for i := range states {
		s := &states[i]
		if s.EntityID == entityID {
			mainEntity = s
		} else if strings.Contains(s.EntityID, baseName) {
			relatedEntities = append(relatedEntities, *s)
		}
	}

	if mainEntity == nil {
		return fmt.Errorf("%w: %s", ErrEntityNotFound, entityID)
	}

	// Parse timestamps and calculate ages
	now := time.Now()
	staleThreshold := 1 * time.Hour

	type entityHealth struct {
		EntityID    string  `json:"entity_id"`
		State       string  `json:"state"`
		LastUpdated string  `json:"last_updated"`
		Age         string  `json:"age"`
		AgeSeconds  float64 `json:"age_seconds"`
		Status      string  `json:"status"`
	}

	parseHealth := func(s *types.HAState) entityHealth {
		h := entityHealth{
			EntityID:    s.EntityID,
			State:       s.State,
			LastUpdated: s.LastUpdated,
			Status:      statusUnknown,
		}

		if s.LastUpdated != "" {
			if t, parseErr := time.Parse(time.RFC3339, s.LastUpdated); parseErr == nil {
				age := now.Sub(t)
				h.AgeSeconds = age.Seconds()
				h.Age = formatDuration(age)
				if age > staleThreshold {
					h.Status = statusStale
				} else {
					h.Status = statusOK
				}
			}
		}
		return h
	}

	mainHealth := parseHealth(mainEntity)
	var relatedHealth []entityHealth
	for i := range relatedEntities {
		relatedHealth = append(relatedHealth, parseHealth(&relatedEntities[i]))
	}

	// Determine overall status
	overallStatus := statusOK
	if mainHealth.Status == statusStale {
		overallStatus = statusStale
	}

	result := map[string]any{
		"entity":          mainHealth,
		"related":         relatedHealth,
		"overall_status":  overallStatus,
		"stale_threshold": staleThreshold.String(),
	}

	if output.IsJSON() {
		output.Data(result, output.WithCommand("device-health"))
	} else {
		// Human-readable output
		statusIcon := "✓"
		if mainHealth.Status == statusStale {
			statusIcon = "⚠️"
		}
		fmt.Printf("Device Health for %s:\n\n", entityID)
		fmt.Printf("  %s %s: %s (%s ago) %s\n", statusIcon, mainHealth.EntityID, mainHealth.State, mainHealth.Age, strings.ToUpper(mainHealth.Status))

		if len(relatedHealth) > 0 {
			fmt.Printf("\nRelated entities:\n")
			for _, h := range relatedHealth {
				icon := "✓"
				if h.Status == statusStale {
					icon = "⚠️"
				}
				fmt.Printf("  %s %s: %s (%s ago)\n", icon, h.EntityID, h.State, h.Age)
			}
		}

		fmt.Printf("\nOverall: %s\n", strings.ToUpper(overallStatus))
	}

	return nil
}

// HandleCompare compares two entities side-by-side.
func HandleCompare(ctx *Context) error {
	entity1, err := RequireArg(ctx, 1, "Usage: compare <entity_id1> <entity_id2>")
	if err != nil {
		return err
	}
	entity2, err := RequireArg(ctx, 2, "Usage: compare <entity_id1> <entity_id2>")
	if err != nil {
		return err
	}

	states, err := client.SendMessageTyped[[]types.HAState](ctx.Client, "get_states", nil)
	if err != nil {
		return err
	}

	var state1, state2 *types.HAState
	for i := range states {
		if states[i].EntityID == entity1 {
			state1 = &states[i]
		}
		if states[i].EntityID == entity2 {
			state2 = &states[i]
		}
	}

	if state1 == nil {
		return fmt.Errorf("%w: %s", ErrEntityNotFound, entity1)
	}
	if state2 == nil {
		return fmt.Errorf("%w: %s", ErrEntityNotFound, entity2)
	}

	now := time.Now()
	parseAge := func(ts string) string {
		if t, parseErr := time.Parse(time.RFC3339, ts); parseErr == nil {
			return formatDuration(now.Sub(t))
		}
		return "unknown"
	}

	comparison := map[string]any{
		"entity1": map[string]any{
			"entity_id":    state1.EntityID,
			"state":        state1.State,
			"last_updated": state1.LastUpdated,
			"age":          parseAge(state1.LastUpdated),
			"attributes":   state1.Attributes,
		},
		"entity2": map[string]any{
			"entity_id":    state2.EntityID,
			"state":        state2.State,
			"last_updated": state2.LastUpdated,
			"age":          parseAge(state2.LastUpdated),
			"attributes":   state2.Attributes,
		},
		"differences": map[string]any{
			"state_match": state1.State == state2.State,
		},
	}

	// Find attribute differences
	attrDiffs := make(map[string]any)
	for k, v1 := range state1.Attributes {
		if v2, ok := state2.Attributes[k]; ok {
			if fmt.Sprintf("%v", v1) != fmt.Sprintf("%v", v2) {
				attrDiffs[k] = map[string]any{"entity1": v1, "entity2": v2}
			}
		} else {
			attrDiffs[k] = map[string]any{"entity1": v1, "entity2": nil}
		}
	}
	for k, v2 := range state2.Attributes {
		if _, ok := state1.Attributes[k]; !ok {
			attrDiffs[k] = map[string]any{"entity1": nil, "entity2": v2}
		}
	}
	if diffs, ok := comparison["differences"].(map[string]any); ok {
		diffs["attributes"] = attrDiffs
	}

	if output.IsJSON() {
		output.Data(comparison, output.WithCommand("compare"))
	} else {
		fmt.Printf("Comparison: %s vs %s\n\n", entity1, entity2)
		fmt.Printf("%-30s | %-25s | %-25s\n", "Property", entity1, entity2)
		fmt.Printf("%s\n", strings.Repeat("-", 85))
		fmt.Printf("%-30s | %-25s | %-25s\n", "State", state1.State, state2.State)
		fmt.Printf("%-30s | %-25s | %-25s\n", "Last Updated", parseAge(state1.LastUpdated)+" ago", parseAge(state2.LastUpdated)+" ago")

		if len(attrDiffs) > 0 {
			fmt.Printf("\nAttribute Differences:\n")
			for k, v := range attrDiffs {
				if diff, ok := v.(map[string]any); ok {
					fmt.Printf("  %s: %v vs %v\n", k, diff["entity1"], diff["entity2"])
				}
			}
		}
	}

	return nil
}

// formatDuration formats a duration in a human-readable way.
func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm", int(d.Minutes()))
	}
	if d < 24*time.Hour {
		return fmt.Sprintf("%dh", int(d.Hours()))
	}
	return fmt.Sprintf("%dd", int(d.Hours()/24))
}
