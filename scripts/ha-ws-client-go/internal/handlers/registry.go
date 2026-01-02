package handlers

import (
	"fmt"
	"strings"

	"github.com/home-assistant-blueprints/ha-ws-client-go/internal/client"
	"github.com/home-assistant-blueprints/ha-ws-client-go/internal/output"
	"github.com/home-assistant-blueprints/ha-ws-client-go/internal/types"
)

// HandleEntities lists or searches the entity registry.
// Wrapped with: WithPattern(1)
var HandleEntities = Apply(
	WithPattern(1),
	handleEntities,
)

func handleEntities(ctx *Context) error {
	re := ctx.Config.Pattern

	entries, err := client.SendMessageTyped[[]types.EntityEntry](ctx.Client, "config/entity_registry/list", nil)
	if err != nil {
		return err
	}

	// Filter by pattern if provided
	if re != nil {
		var filtered []types.EntityEntry
		for _, e := range entries {
			if re.MatchString(e.EntityID) || re.MatchString(e.Name) || re.MatchString(e.OriginalName) {
				filtered = append(filtered, e)
			}
		}
		entries = filtered
	}

	output.List(entries,
		output.ListTitle[types.EntityEntry]("Entity registry"),
		output.ListCommand[types.EntityEntry]("entities"),
		output.ListFormatter(func(e types.EntityEntry, _ int) string {
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
		}),
	)
	return nil
}

// HandleDevices lists or searches the device registry.
// Wrapped with: WithPattern(1)
var HandleDevices = Apply(
	WithPattern(1),
	handleDevices,
)

func handleDevices(ctx *Context) error {
	re := ctx.Config.Pattern

	devices, err := client.SendMessageTyped[[]types.DeviceEntry](ctx.Client, "config/device_registry/list", nil)
	if err != nil {
		return err
	}

	// Filter by pattern if provided
	if re != nil {
		var filtered []types.DeviceEntry
		for _, d := range devices {
			name := d.Name
			if d.NameByUser != "" {
				name = d.NameByUser
			}
			if re.MatchString(d.ID) || re.MatchString(name) || re.MatchString(d.Manufacturer) || re.MatchString(d.Model) {
				filtered = append(filtered, d)
			}
		}
		devices = filtered
	}

	output.List(devices,
		output.ListTitle[types.DeviceEntry]("Device registry"),
		output.ListCommand[types.DeviceEntry]("devices"),
		output.ListFormatter(func(d types.DeviceEntry, _ int) string {
			name := d.Name
			if d.NameByUser != "" {
				name = d.NameByUser
			}
			if output.IsCompact() {
				return fmt.Sprintf("%s: %s (%s %s)", d.ID[:8], name, d.Manufacturer, d.Model)
			}
			return fmt.Sprintf("%s\n  Name: %s\n  Manufacturer: %s\n  Model: %s\n  Area: %s",
				d.ID, name, d.Manufacturer, d.Model, d.AreaID)
		}),
	)
	return nil
}

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
