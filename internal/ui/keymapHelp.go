package ui

import (
	"github.com/RubnMC/rubipaper/internal/config"
	"github.com/charmbracelet/bubbles/key"
)

// ComponentKeyMap is implemented by components that expose keybindings
// for the footer help bar.
type ComponentKeyMap interface {
	ShortHelp() []key.Binding
}

var keyQuit = key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "quit"))

// globalKeyMap returns the always-visible bindings: quit, plus focus-switch
// keys for any non-empty configured keybind.
func globalKeyMap(kb config.KeybindsConfig) []key.Binding {
	keys := []key.Binding{keyQuit}
	if kb.FocusNext != "" {
		keys = append(keys, key.NewBinding(key.WithKeys(kb.FocusNext), key.WithHelp(kb.FocusNext, "next panel")))
	}
	if kb.FocusPrev != "" {
		keys = append(keys, key.NewBinding(key.WithKeys(kb.FocusPrev), key.WithHelp(kb.FocusPrev, "prev panel")))
	}
	return keys
}
