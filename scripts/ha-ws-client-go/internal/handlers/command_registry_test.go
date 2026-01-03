package handlers

import (
	"testing"
)

func TestCmd(t *testing.T) {
	handler := func(_ *Context) error { return nil }
	cmd := Cmd("test", "Test command", "<arg1>", "testing", handler)

	if cmd.Name != "test" {
		t.Errorf("expected name 'test', got %q", cmd.Name)
	}
	if cmd.Usage != "Test command" {
		t.Errorf("expected usage 'Test command', got %q", cmd.Usage)
	}
	if cmd.ArgsUsage != "<arg1>" {
		t.Errorf("expected argsUsage '<arg1>', got %q", cmd.ArgsUsage)
	}
	if cmd.Category != "testing" {
		t.Errorf("expected category 'testing', got %q", cmd.Category)
	}
	if cmd.Handler == nil {
		t.Error("expected handler to be set")
	}
}

func TestCommandRegistry_Register(t *testing.T) {
	registry := &CommandRegistry{
		commands: make(map[string]*CommandDefinition),
	}

	cmd1 := &CommandDefinition{Name: "cmd1", Usage: "Command 1", Category: "cat1"}
	cmd2 := &CommandDefinition{Name: "cmd2", Usage: "Command 2", Category: "cat1"}

	registry.Register(cmd1)
	registry.Register(cmd2)

	if len(registry.commands) != 2 {
		t.Errorf("expected 2 commands, got %d", len(registry.commands))
	}

	// Verify order is preserved
	if registry.order[0] != "cmd1" || registry.order[1] != "cmd2" {
		t.Errorf("order not preserved: %v", registry.order)
	}
}

func TestCommandRegistry_Register_Overwrite(t *testing.T) {
	registry := &CommandRegistry{
		commands: make(map[string]*CommandDefinition),
	}

	cmd1 := &CommandDefinition{Name: "cmd1", Usage: "Original"}
	cmd2 := &CommandDefinition{Name: "cmd1", Usage: "Updated"}

	registry.Register(cmd1)
	registry.Register(cmd2)

	// Should still have only 1 command
	if len(registry.commands) != 1 {
		t.Errorf("expected 1 command, got %d", len(registry.commands))
	}

	// Order should not have duplicate
	if len(registry.order) != 1 {
		t.Errorf("expected 1 in order, got %d", len(registry.order))
	}

	// Should have the updated command
	if cmd, ok := registry.commands["cmd1"]; ok {
		if cmd.Usage != "Updated" {
			t.Errorf("expected usage 'Updated', got %q", cmd.Usage)
		}
	} else {
		t.Error("command cmd1 not found")
	}
}

func TestCommandRegistry_GetCommand(t *testing.T) {
	registry := &CommandRegistry{
		commands: make(map[string]*CommandDefinition),
	}

	cmd := &CommandDefinition{Name: "test", Usage: "Test"}
	registry.Register(cmd)

	// Test found
	found, ok := registry.GetCommand("test")
	if !ok {
		t.Error("expected to find command 'test'")
	}
	if found.Usage != "Test" {
		t.Errorf("expected usage 'Test', got %q", found.Usage)
	}

	// Test not found
	_, ok = registry.GetCommand("nonexistent")
	if ok {
		t.Error("expected to not find command 'nonexistent'")
	}
}

func TestCommandRegistry_GetAllCommands(t *testing.T) {
	registry := &CommandRegistry{
		commands: make(map[string]*CommandDefinition),
	}

	registry.Register(&CommandDefinition{Name: "first", Category: "cat1"})
	registry.Register(&CommandDefinition{Name: "second", Category: "cat2"})
	registry.Register(&CommandDefinition{Name: "third", Category: "cat1"})

	cmds := registry.GetAllCommands()

	if len(cmds) != 3 {
		t.Errorf("expected 3 commands, got %d", len(cmds))
	}

	// Verify order is preserved
	expectedOrder := []string{"first", "second", "third"}
	for i, cmd := range cmds {
		if cmd.Name != expectedOrder[i] {
			t.Errorf("expected command %d to be %q, got %q", i, expectedOrder[i], cmd.Name)
		}
	}
}

func TestCommandRegistry_GetCommandsByCategory(t *testing.T) {
	registry := &CommandRegistry{
		commands: make(map[string]*CommandDefinition),
	}

	registry.Register(&CommandDefinition{Name: "b-cmd", Category: "cat1"})
	registry.Register(&CommandDefinition{Name: "other", Category: "cat2"})
	registry.Register(&CommandDefinition{Name: "a-cmd", Category: "cat1"})

	cmds := registry.GetCommandsByCategory("cat1")

	if len(cmds) != 2 {
		t.Errorf("expected 2 commands in cat1, got %d", len(cmds))
	}

	// Should be sorted by name
	if cmds[0].Name != "a-cmd" || cmds[1].Name != "b-cmd" {
		t.Errorf("expected sorted order [a-cmd, b-cmd], got [%s, %s]", cmds[0].Name, cmds[1].Name)
	}
}

func TestCommandRegistry_GetCategories(t *testing.T) {
	registry := &CommandRegistry{
		commands: make(map[string]*CommandDefinition),
	}

	registry.Register(&CommandDefinition{Name: "cmd1", Category: "first"})
	registry.Register(&CommandDefinition{Name: "cmd2", Category: "second"})
	registry.Register(&CommandDefinition{Name: "cmd3", Category: "first"}) // Duplicate category
	registry.Register(&CommandDefinition{Name: "cmd4", Category: "third"})

	categories := registry.GetCategories()

	if len(categories) != 3 {
		t.Errorf("expected 3 categories, got %d: %v", len(categories), categories)
	}

	// Verify order matches first appearance
	expected := []string{"first", "second", "third"}
	for i, cat := range categories {
		if cat != expected[i] {
			t.Errorf("expected category %d to be %q, got %q", i, expected[i], cat)
		}
	}
}

func TestGlobalRegistry_HasCommands(t *testing.T) {
	// This test verifies that the global registry is populated by init() functions
	cmds := GetAllCommands()

	if len(cmds) == 0 {
		t.Error("expected global registry to have commands from init() functions")
	}

	// Verify some known commands exist
	knownCommands := []string{"ping", "state", "history", "traces", "monitor"}
	for _, name := range knownCommands {
		if _, ok := GetCommand(name); !ok {
			t.Errorf("expected command %q to be registered", name)
		}
	}
}

func TestGlobalRegistry_Categories(t *testing.T) {
	categories := GetCategories()

	if len(categories) == 0 {
		t.Error("expected global registry to have categories")
	}

	// Verify some known categories exist
	categorySet := make(map[string]bool)
	for _, cat := range categories {
		categorySet[cat] = true
	}

	expectedCategories := []string{"basic", "history", "automation", "monitoring", "registry"}
	for _, cat := range expectedCategories {
		if !categorySet[cat] {
			t.Errorf("expected category %q to exist, available: %v", cat, categories)
		}
	}
}

func TestGlobalRegistry_CommandCounts(t *testing.T) {
	// Count commands by category
	categoryCounts := make(map[string]int)
	for _, cmd := range GetAllCommands() {
		categoryCounts[cmd.Category]++
	}

	// Verify we have multiple commands per category
	for cat, count := range categoryCounts {
		if count == 0 {
			t.Errorf("category %q has no commands", cat)
		}
	}

	// Total should be around 30+ commands
	total := len(GetAllCommands())
	if total < 25 {
		t.Errorf("expected at least 25 commands, got %d", total)
	}
}
