// Package handlers provides command handlers for the CLI.
package handlers

import (
	"sort"
	"sync"
)

// CommandDefinition describes a CLI command and its handler.
type CommandDefinition struct {
	// Name is the command name (e.g., "state", "history").
	Name string

	// Usage is a short description of what the command does.
	Usage string

	// ArgsUsage describes the expected arguments (e.g., "<entity_id> [hours]").
	ArgsUsage string

	// Category groups related commands (e.g., "basic", "history", "automation").
	Category string

	// Handler is the function that executes the command.
	Handler Handler
}

// CommandRegistry stores all registered commands.
type CommandRegistry struct {
	mu       sync.RWMutex
	commands map[string]*CommandDefinition
	order    []string // Track insertion order for consistent iteration
}

// globalRegistry is the default registry used by the package.
var globalRegistry = &CommandRegistry{
	commands: make(map[string]*CommandDefinition),
}

// Register adds a command to the global registry.
// If a command with the same name already exists, it will be overwritten.
func Register(cmd *CommandDefinition) {
	globalRegistry.Register(cmd)
}

// RegisterAll adds multiple commands to the global registry.
func RegisterAll(cmds ...*CommandDefinition) {
	for _, cmd := range cmds {
		Register(cmd)
	}
}

// GetCommand retrieves a command by name from the global registry.
func GetCommand(name string) (*CommandDefinition, bool) {
	return globalRegistry.GetCommand(name)
}

// GetAllCommands returns all registered commands in registration order.
func GetAllCommands() []*CommandDefinition {
	return globalRegistry.GetAllCommands()
}

// GetCommandsByCategory returns all commands in a category.
func GetCommandsByCategory(category string) []*CommandDefinition {
	return globalRegistry.GetCommandsByCategory(category)
}

// GetCategories returns all unique categories in registration order.
func GetCategories() []string {
	return globalRegistry.GetCategories()
}

// Register adds a command to the registry.
func (r *CommandRegistry) Register(cmd *CommandDefinition) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.commands[cmd.Name]; !exists {
		r.order = append(r.order, cmd.Name)
	}
	r.commands[cmd.Name] = cmd
}

// GetCommand retrieves a command by name.
func (r *CommandRegistry) GetCommand(name string) (*CommandDefinition, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	cmd, ok := r.commands[name]
	return cmd, ok
}

// GetAllCommands returns all registered commands in registration order.
func (r *CommandRegistry) GetAllCommands() []*CommandDefinition {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]*CommandDefinition, 0, len(r.commands))
	for _, name := range r.order {
		if cmd, ok := r.commands[name]; ok {
			result = append(result, cmd)
		}
	}
	return result
}

// GetCommandsByCategory returns all commands in a category, sorted by name.
func (r *CommandRegistry) GetCommandsByCategory(category string) []*CommandDefinition {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*CommandDefinition
	for _, name := range r.order {
		if cmd, ok := r.commands[name]; ok && cmd.Category == category {
			result = append(result, cmd)
		}
	}

	// Sort by name within category
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})

	return result
}

// GetCategories returns all unique categories in the order they first appear.
func (r *CommandRegistry) GetCategories() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	seen := make(map[string]bool)
	var categories []string

	for _, name := range r.order {
		if cmd, ok := r.commands[name]; ok {
			if !seen[cmd.Category] {
				seen[cmd.Category] = true
				categories = append(categories, cmd.Category)
			}
		}
	}

	return categories
}

// Cmd is a convenience function for creating a CommandDefinition.
// It makes command registration more concise.
//
// Example:
//
//	handlers.Register(handlers.Cmd("ping", "Test connection", "", "basic", HandlePing))
func Cmd(name, usage, argsUsage, category string, handler Handler) *CommandDefinition {
	return &CommandDefinition{
		Name:      name,
		Usage:     usage,
		ArgsUsage: argsUsage,
		Category:  category,
		Handler:   handler,
	}
}
