package ui

import (
	"strings"
	"testing"

	"github.com/RubnMC/rubipaper/internal/config"
	"github.com/RubnMC/rubipaper/internal/domain"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
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

func TestViewReturnsNonEmpty(t *testing.T) {
	kb, ap := testConfigs()
	m := New(makeWallpapers("a.jpg"), "/pics", stubBackend{}, domain.ModeFill, kb, ap)
	view := m.View()
	if view == "" {
		t.Error("View() should return a non-empty string")
	}
}

func TestViewWithFocusBorderColor(t *testing.T) {
	// Force lipgloss to output colors for testing
	oldProfile := lipgloss.ColorProfile()
	defer lipgloss.SetColorProfile(oldProfile)
	lipgloss.SetColorProfile(termenv.TrueColor)

	kb := config.KeybindsConfig{FocusNext: "shift+down", FocusPrev: "shift+up"}

	apColor := config.AppearanceConfig{FocusBorderColor: "#ff0000"}
	mColor := New(makeWallpapers("a.jpg"), "/pics", stubBackend{}, domain.ModeFill, kb, apColor)
	viewWithColor := mColor.View()

	apNone := config.AppearanceConfig{}
	mNone := New(makeWallpapers("a.jpg"), "/pics", stubBackend{}, domain.ModeFill, kb, apNone)
	viewWithout := mNone.View()

	if viewWithColor == viewWithout {
		t.Error("View() output should differ when FocusBorderColor is set vs not set")
	}
}

func TestQuitKeyReturnsQuitCmd(t *testing.T) {
	kb, ap := testConfigs()
	m := New(makeWallpapers("a.jpg"), "/pics", stubBackend{}, domain.ModeFill, kb, ap)

	for _, key := range []string{"q", "ctrl+c"} {
		_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)})
		if key == "ctrl+c" {
			_, cmd = m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
		}
		if cmd == nil {
			t.Errorf("Update(%q) should return a non-nil cmd", key)
		}
	}
}

func TestViewFooterShowsFocusedComponentKeys(t *testing.T) {
	kb, ap := testConfigs()
	m := New(makeWallpapers("a.jpg"), "/pics", stubBackend{}, domain.ModeFill, kb, ap)

	view := m.View()
	for _, want := range []string{"set wallpaper", "toggle recursive search"} {
		if !strings.Contains(view, want) {
			t.Errorf("footer missing FileExplorer key hint %q when it is focused", want)
		}
	}
}

func TestViewFooterSwitchesWithFocus(t *testing.T) {
	kb, ap := testConfigs()
	m := New(makeWallpapers("a.jpg"), "/pics", stubBackend{}, domain.ModeFill, kb, ap)

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyShiftUp})
	m = updated.(Model)

	view := m.View()
	if strings.Contains(view, "set wallpaper") || strings.Contains(view, "toggle recursive search") {
		t.Error("footer should not show FileExplorer's keybinds when it is not focused")
	}
	for _, want := range []string{"quit", "next panel", "prev panel"} {
		if !strings.Contains(view, want) {
			t.Errorf("footer missing global key hint %q when OptionsBar is focused", want)
		}
	}
}

func TestViewFooterAlwaysShowsGlobalKeys(t *testing.T) {
	kb, ap := testConfigs()
	m := New(makeWallpapers("a.jpg"), "/pics", stubBackend{}, domain.ModeFill, kb, ap)

	view := m.View()
	for _, want := range []string{"quit", "next panel", "prev panel"} {
		if !strings.Contains(view, want) {
			t.Errorf("footer missing global key hint %q", want)
		}
	}
}

func TestViewFooterOmitsEmptyFocusKeys(t *testing.T) {
	m := New(makeWallpapers("a.jpg"), "/pics", stubBackend{}, domain.ModeFill, config.KeybindsConfig{}, config.AppearanceConfig{})

	view := m.View()
	if strings.Contains(view, "next panel") || strings.Contains(view, "prev panel") {
		t.Error("footer should omit focus-switch hints when keybinds are unconfigured")
	}
	if !strings.Contains(view, "quit") {
		t.Error("footer should always show quit")
	}
}

func TestViewContainsPreviewPanel(t *testing.T) {
	kb, ap := testConfigs()
	m := New(makeWallpapers("a.jpg"), "/pics", stubBackend{}, domain.ModeFill, kb, ap)
	view := m.View()
	// Preview panel should show "no selection" or "kitty" placeholder
	if !strings.Contains(view, "no selection") && !strings.Contains(view, "kitty") {
		t.Error("View() should contain preview panel content")
	}
}

func TestOptionsBarSpansFullTerminalWidth(t *testing.T) {
	kb, ap := testConfigs()
	m := New(makeWallpapers("a.jpg"), "/pics", stubBackend{}, domain.ModeFill, kb, ap)

	updated, _ := m.Update(tea.WindowSizeMsg{Width: 200, Height: 40})
	m = updated.(Model)

	topLine := strings.Split(m.View(), "\n")[0]
	// TrimRight strips any padding JoinVertical adds to match a wider sibling
	// block, so this measures optStyle's own width, not JoinVertical's.
	if got := lipgloss.Width(strings.TrimRight(topLine, " ")); got != 200 {
		t.Errorf("optbar top border width = %d, want 200", got)
	}
}

func TestOptionsBarResizesWithTerminal(t *testing.T) {
	kb, ap := testConfigs()
	m := New(makeWallpapers("a.jpg"), "/pics", stubBackend{}, domain.ModeFill, kb, ap)

	for _, w := range []int{80, 160} {
		updated, _ := m.Update(tea.WindowSizeMsg{Width: w, Height: 40})
		mw := updated.(Model)
		topLine := strings.Split(mw.View(), "\n")[0]
		if got := lipgloss.Width(strings.TrimRight(topLine, " ")); got != w {
			t.Errorf("width %d: optbar top border width = %d, want %d", w, got, w)
		}
	}
}

func TestOptionsBarClampsOnNarrowTerminal(t *testing.T) {
	kb, ap := testConfigs()
	m := New(makeWallpapers("a.jpg"), "/pics", stubBackend{}, domain.ModeFill, kb, ap)

	updated, _ := m.Update(tea.WindowSizeMsg{Width: 10, Height: 40})
	m = updated.(Model)

	for _, line := range strings.Split(m.View(), "\n")[:3] { // top border, content, bottom border
		// TrimRight strips padding JoinVertical adds to match the footer
		// line (which isn't width-bounded) — without it every line here
		// would read as footer-width regardless of optStyle's own clamp.
		if got := lipgloss.Width(strings.TrimRight(line, " ")); got > 10 {
			t.Errorf("optbar line width = %d, want <= 10 on a narrow terminal: %q", got, line)
		}
	}
}

func TestPreviewPanelReceivesWallpaperSelectedMsg(t *testing.T) {
	kb, ap := testConfigs()
	m := New(makeWallpapers("a.jpg", "b.jpg"), "/pics", stubBackend{}, domain.ModeFill, kb, ap)

	w := &domain.Wallpaper{FileName: "picked.jpg", Path: "/pics/picked.jpg"}
	updated, _ := m.Update(WallpaperSelectedMsg{Wallpaper: w})
	m = updated.(Model)

	if m.previewPanel.selected == nil || m.previewPanel.selected.FileName != "picked.jpg" {
		t.Error("previewPanel should store the wallpaper from WallpaperSelectedMsg")
	}
}
