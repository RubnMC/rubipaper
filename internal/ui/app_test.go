package ui

import (
	"testing"

	"github.com/RubnMC/rubipaper/internal/config"
	"github.com/RubnMC/rubipaper/internal/domain"
	tea "github.com/charmbracelet/bubbletea"
)

func testConfigs() (config.KeybindsConfig, config.AppearanceConfig) {
	return config.KeybindsConfig{FocusNext: "shift+down", FocusPrev: "shift+up"},
		config.AppearanceConfig{}
}

func TestInitialFocusOnFileExplorer(t *testing.T) {
	kb, ap := testConfigs()
	m := New(makeWallpapers("a.jpg"), "/pics", stubBackend{}, domain.ModeFill, kb, ap)
	if m.components[0].IsFocused() {
		t.Error("OptionsBar should not be focused on init")
	}
	if !m.components[1].IsFocused() {
		t.Error("FileExplorer should be focused on init")
	}
}

func TestFocusNextCyclesForward(t *testing.T) {
	kb, ap := testConfigs()
	m := New(makeWallpapers("a.jpg"), "/pics", stubBackend{}, domain.ModeFill, kb, ap)

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyShiftDown})
	m = updated.(Model)
	if m.focused != 0 {
		t.Errorf("focused = %d, want 0 after shift+down from index 1", m.focused)
	}
	if !m.components[0].IsFocused() {
		t.Error("component 0 should now be focused")
	}
	if m.components[1].IsFocused() {
		t.Error("component 1 should now be blurred")
	}
}

func TestFocusNextWrapsAround(t *testing.T) {
	kb, ap := testConfigs()
	m := New(makeWallpapers("a.jpg"), "/pics", stubBackend{}, domain.ModeFill, kb, ap)

	// Two shift+downs from index 1 should land back at index 1
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyShiftDown})
	updated, _ = updated.(Model).Update(tea.KeyMsg{Type: tea.KeyShiftDown})
	m = updated.(Model)
	if m.focused != 1 {
		t.Errorf("focused = %d, want 1 after full wrap", m.focused)
	}
}

func TestFocusPrevCyclesBackward(t *testing.T) {
	kb, ap := testConfigs()
	m := New(makeWallpapers("a.jpg"), "/pics", stubBackend{}, domain.ModeFill, kb, ap)

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyShiftUp})
	m = updated.(Model)
	if m.focused != 0 {
		t.Errorf("focused = %d, want 0 after shift+up from index 1", m.focused)
	}
}

func TestBroadcastDeliveredToAllComponents(t *testing.T) {
	kb, ap := testConfigs()
	m := New(makeWallpapers("a.jpg"), "/pics", stubBackend{}, domain.ModeFill, kb, ap)

	updated, _ := m.Update(RecursiveToggledMsg{Recursive: true})
	m = updated.(Model)
	ob := m.components[0].(OptionsBarModel)
	if !ob.recursive {
		t.Error("OptionsBar should receive RecursiveToggledMsg via broadcast")
	}
}
