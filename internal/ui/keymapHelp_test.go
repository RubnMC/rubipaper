package ui

import (
	"testing"

	"github.com/RubnMC/rubipaper/internal/config"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

func TestGlobalKeyMapAlwaysIncludesQuit(t *testing.T) {
	keys := globalKeyMap(config.KeybindsConfig{})
	if len(keys) != 1 {
		t.Fatalf("globalKeyMap with no focus keys = %d bindings, want 1 (quit only)", len(keys))
	}
	if !key.Matches(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")}, keys[0]) {
		t.Error("first binding should match 'q'")
	}
}

func TestGlobalKeyMapIncludesConfiguredFocusKeys(t *testing.T) {
	kb := config.KeybindsConfig{FocusNext: "shift+down", FocusPrev: "shift+up"}
	keys := globalKeyMap(kb)
	if len(keys) != 3 {
		t.Fatalf("globalKeyMap with both focus keys = %d bindings, want 3", len(keys))
	}
	if !key.Matches(tea.KeyMsg{Type: tea.KeyShiftDown}, keys[1]) {
		t.Error("second binding should match shift+down")
	}
	if !key.Matches(tea.KeyMsg{Type: tea.KeyShiftUp}, keys[2]) {
		t.Error("third binding should match shift+up")
	}
}

func TestGlobalKeyMapOmitsUnconfiguredFocusKeys(t *testing.T) {
	keys := globalKeyMap(config.KeybindsConfig{FocusNext: "shift+down"})
	if len(keys) != 2 {
		t.Fatalf("globalKeyMap with only FocusNext = %d bindings, want 2", len(keys))
	}
}
