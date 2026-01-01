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
	pattern, err := RequireArg(ctx, 1, "Usage: states-filter <pattern>")
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

	output.List(filtered,
		output.ListTitle[types.HAState](fmt.Sprintf("Found %d matching entities", len(filtered))),
		output.ListCommand[types.HAState]("states-filter"),
		output.ListFormatter(func(s types.HAState, _ int) string {
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
				return fmt.Sprintf("  %s: %s", domain, join(svcNames, ", "))
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

func join(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	var result strings.Builder
	result.WriteString(strs[0])
	for i := 1; i < len(strs); i++ {
		result.WriteString(sep + strs[i])
	}
	return result.String()
}
