package handlers

import (
	"fmt"
	"regexp"
	"slices"
	"strings"

	"github.com/home-assistant-blueprints/ha-ws-client-go/internal/client"
	"github.com/home-assistant-blueprints/ha-ws-client-go/internal/output"
	"github.com/home-assistant-blueprints/ha-ws-client-go/internal/types"
)

func init() {
	// Register entity/device/area registry commands
	RegisterAll(
		Cmd("entities", "List/search entity registry", "[pattern]", "registry", HandleEntities),
		Cmd("devices", "List/search device registry", "[pattern]", "registry", HandleDevices),
		Cmd("areas", "List all areas", "", "registry", HandleAreas),
	)
}

// RegistryConfig holds the configuration for a filtered registry handler.
type RegistryConfig[T any] struct {
	// MessageType is the WebSocket message type to fetch data (e.g., "config/entity_registry/list").
	MessageType string
	// Title is the list title shown in output.
	Title string
	// Command is the command name for JSON output.
	Command string
	// MatchFields returns the string fields to match against the pattern for a given item.
	MatchFields func(item T) []string
	// Formatter formats an item for display output.
	Formatter func(item T, index int) string
}

// FilteredRegistryHandler creates a handler that fetches registry data,
// optionally filters by pattern, and outputs the results.
// This is a generic helper that unifies the pattern-filtering logic used by
// entities, devices, and similar registry commands.
func FilteredRegistryHandler[T any](cfg RegistryConfig[T]) Handler {
	return func(ctx *Context) error {
		re := ctx.Config.Pattern

		items, err := client.SendMessageTyped[[]T](ctx.Client, cfg.MessageType, nil)
		if err != nil {
			return err
		}

		// Filter by pattern if provided
		if re != nil {
			items = filterByPattern(items, re, cfg.MatchFields)
		}

		output.List(items,
			output.ListTitle[T](cfg.Title),
			output.ListCommand[T](cfg.Command),
			output.ListFormatter(cfg.Formatter),
		)
		return nil
	}
}

// filterByPattern filters a slice of items, keeping only those where at least
// one of the match fields matches the given pattern.
func filterByPattern[T any](items []T, re *regexp.Regexp, matchFields func(T) []string) []T {
	var filtered []T
	for _, item := range items {
		if slices.ContainsFunc(matchFields(item), re.MatchString) {
			filtered = append(filtered, item)
		}
	}
	return filtered
}

// HandleEntities lists or searches the entity registry.
// Wrapped with: WithPattern(1)
var HandleEntities = Apply(
	WithPattern(1),
	FilteredRegistryHandler(RegistryConfig[types.EntityEntry]{
		MessageType: "config/entity_registry/list",
		Title:       "Entity registry",
		Command:     "entities",
		MatchFields: func(e types.EntityEntry) []string {
			return []string{e.EntityID, e.Name, e.OriginalName}
		},
		Formatter: func(e types.EntityEntry, _ int) string {
			name := e.Name
			if name == "" {
				name = e.OriginalName
			}
			disabled := ""
			if e.DisabledBy != "" {
				disabled = " [disabled]"
			}
			if output.IsCompact() {
				return fmt.Sprintf("%s (%s)%s", e.EntityID, name, disabled)
			}
			return fmt.Sprintf("%s\n  Name: %s\n  Platform: %s%s", e.EntityID, name, e.Platform, disabled)
		},
	}),
)

// HandleDevices lists or searches the device registry.
// Wrapped with: WithPattern(1)
var HandleDevices = Apply(
	WithPattern(1),
	FilteredRegistryHandler(RegistryConfig[types.DeviceEntry]{
		MessageType: "config/device_registry/list",
		Title:       "Device registry",
		Command:     "devices",
		MatchFields: func(d types.DeviceEntry) []string {
			name := d.Name
			if d.NameByUser != "" {
				name = d.NameByUser
			}
			return []string{d.ID, name, d.Manufacturer, d.Model}
		},
		Formatter: func(d types.DeviceEntry, _ int) string {
			name := d.Name
			if d.NameByUser != "" {
				name = d.NameByUser
			}
			if output.IsCompact() {
				return fmt.Sprintf("%s: %s (%s %s)", d.ID[:8], name, d.Manufacturer, d.Model)
			}
			return fmt.Sprintf("%s\n  Name: %s\n  Manufacturer: %s\n  Model: %s\n  Area: %s",
				d.ID, name, d.Manufacturer, d.Model, d.AreaID)
		},
	}),
)

// HandleAreas lists all areas.
func HandleAreas(ctx *Context) error {
	areas, err := client.SendMessageTyped[[]types.AreaEntry](ctx.Client, "config/area_registry/list", nil)
	if err != nil {
		return err
	}

	output.List(areas,
		output.ListTitle[types.AreaEntry]("Areas"),
		output.ListCommand[types.AreaEntry]("areas"),
		output.ListFormatter(func(a types.AreaEntry, _ int) string {
			aliases := ""
			if len(a.Aliases) > 0 {
				aliases = fmt.Sprintf(" (aliases: %s)", strings.Join(a.Aliases, ", "))
			}
			if output.IsCompact() {
				return fmt.Sprintf("%s: %s%s", a.AreaID, a.Name, aliases)
			}
			return fmt.Sprintf("%s: %s%s", a.AreaID, a.Name, aliases)
		}),
	)
	return nil
}
