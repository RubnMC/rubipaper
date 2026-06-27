package ui

import (
	"strings"
	"testing"

	"github.com/RubnMC/rubipaper/internal/domain"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

func makeWallpapers(names ...string) []domain.Wallpaper {
	ws := make([]domain.Wallpaper, len(names))
	for i, n := range names {
		ws[i] = domain.Wallpaper{Path: "/pics/" + n, FileName: n}
	}
	return ws
}

func TestFileExplorerStartsUnfocused(t *testing.T) {
	f := NewFileExplorer(makeWallpapers("a.jpg"), "/pics", stubBackend{}, domain.ModeFill)
	if f.IsFocused() {
		t.Error("new FileExplorer should not be focused")
	}
}

func TestFileExplorerFocusBlur(t *testing.T) {
	f := NewFileExplorer(makeWallpapers("a.jpg"), "/pics", stubBackend{}, domain.ModeFill)
	if !f.Focus().IsFocused() {
		t.Error("Focus() should return focused component")
	}
	if f.Focus().Blur().IsFocused() {
		t.Error("Blur() should return unfocused component")
	}
	if f.IsFocused() {
		t.Error("Focus() must not mutate the original")
	}
}

func TestFileExplorerIgnoresKeysWhenUnfocused(t *testing.T) {
	f := NewFileExplorer(makeWallpapers("a.jpg", "b.jpg"), "/pics", stubBackend{}, domain.ModeFill)
	updated, _ := f.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	if updated.(FileExplorerModel).cursor != 0 {
		t.Error("unfocused FileExplorer must not move cursor on key press")
	}
}

func TestFileExplorerNavigatesDownAndUp(t *testing.T) {
	f := NewFileExplorer(makeWallpapers("a.jpg", "b.jpg", "c.jpg"), "/pics", stubBackend{}, domain.ModeFill)
	f = f.Focus().(FileExplorerModel)

	updated, _ := f.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	f = updated.(FileExplorerModel)
	if f.cursor != 1 {
		t.Errorf("cursor = %d, want 1 after j", f.cursor)
	}

	updated, _ = f.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
	f = updated.(FileExplorerModel)
	if f.cursor != 0 {
		t.Errorf("cursor = %d, want 0 after k", f.cursor)
	}
}

func TestFileExplorerCursorDoesNotOverflow(t *testing.T) {
	f := NewFileExplorer(makeWallpapers("a.jpg"), "/pics", stubBackend{}, domain.ModeFill)
	f = f.Focus().(FileExplorerModel)

	updated, _ := f.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
	if updated.(FileExplorerModel).cursor != 0 {
		t.Error("cursor should not go below 0")
	}

	updated, _ = f.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	if updated.(FileExplorerModel).cursor != 0 {
		t.Error("cursor should not go past last item with one item")
	}
}

func TestFileExplorerRecursiveToggleUpdatesState(t *testing.T) {
	f := NewFileExplorer(makeWallpapers("a.jpg"), "/pics", stubBackend{}, domain.ModeFill)
	f = f.Focus().(FileExplorerModel)

	updated, cmd := f.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("r")})
	fe := updated.(FileExplorerModel)
	if !fe.recursive {
		t.Error("recursive should be true after first toggle")
	}
	if cmd == nil {
		t.Error("toggle should return a non-nil cmd for scan and broadcast")
	}
}

func TestFileExplorerViewShowsItems(t *testing.T) {
	f := NewFileExplorer(makeWallpapers("wall.jpg", "bg.png"), "/pics", stubBackend{}, domain.ModeFill)
	view := f.View()
	if !strings.Contains(view, "wall.jpg") || !strings.Contains(view, "bg.png") {
		t.Error("View() should render all file names")
	}
}

func TestFileExplorerCursorWithEmptyList(t *testing.T) {
	f := NewFileExplorer(makeWallpapers(), "/pics", stubBackend{}, domain.ModeFill)
	f = f.Focus().(FileExplorerModel)

	updated, _ := f.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	if updated.(FileExplorerModel).cursor != 0 {
		t.Error("cursor should stay at 0 with empty list")
	}

	updated, _ = f.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
	if updated.(FileExplorerModel).cursor != 0 {
		t.Error("cursor should stay at 0 with empty list")
	}
}

func TestFileExplorerRecursiveToggleTwiceRestores(t *testing.T) {
	f := NewFileExplorer(makeWallpapers("a.jpg"), "/pics", stubBackend{}, domain.ModeFill)
	f = f.Focus().(FileExplorerModel)

	updated, _ := f.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("r")})
	f = updated.(FileExplorerModel)
	if !f.recursive {
		t.Error("recursive should be true after first toggle")
	}

	updated, _ = f.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("r")})
	f = updated.(FileExplorerModel)
	if f.recursive {
		t.Error("recursive should be false after second toggle")
	}
}

func TestFileExplorerShortHelp(t *testing.T) {
	f := NewFileExplorer(makeWallpapers("a.jpg"), "/pics", stubBackend{}, domain.ModeFill)
	bindings := f.ShortHelp()
	if len(bindings) != 4 {
		t.Fatalf("ShortHelp() = %d bindings, want 4", len(bindings))
	}
	if !key.Matches(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")}, bindings[0]) {
		t.Error("bindings[0] should match 'k' (up)")
	}
	if !key.Matches(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")}, bindings[1]) {
		t.Error("bindings[1] should match 'j' (down)")
	}
	if !key.Matches(tea.KeyMsg{Type: tea.KeyEnter}, bindings[2]) {
		t.Error("bindings[2] should match enter")
	}
	if !key.Matches(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("r")}, bindings[3]) {
		t.Error("bindings[3] should match 'r' (toggle recursive search)")
	}
}

func TestFileExplorerEmitsSelectionOnCursorDown(t *testing.T) {
	f := NewFileExplorer(makeWallpapers("a.jpg", "b.jpg"), "/pics", stubBackend{}, domain.ModeFill)
	f = f.Focus().(FileExplorerModel)

	_, cmd := f.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	if cmd == nil {
		t.Fatal("cursor move must return a non-nil cmd")
	}
	msg := cmd()
	sel, ok := msg.(WallpaperSelectedMsg)
	if !ok {
		t.Fatalf("cmd() returned %T, want WallpaperSelectedMsg", msg)
	}
	if sel.Wallpaper == nil || sel.Wallpaper.FileName != "b.jpg" {
		t.Errorf("selected wallpaper = %v, want b.jpg", sel.Wallpaper)
	}
}

func TestFileExplorerEmitsNilSelectionOnEmptyList(t *testing.T) {
	f := NewFileExplorer(makeWallpapers(), "/pics", stubBackend{}, domain.ModeFill)
	f = f.Focus().(FileExplorerModel)

	_, cmd := f.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	if cmd == nil {
		t.Fatal("cursor move on empty list must return a non-nil cmd")
	}
	msg := cmd()
	sel, ok := msg.(WallpaperSelectedMsg)
	if !ok {
		t.Fatalf("cmd() returned %T, want WallpaperSelectedMsg", msg)
	}
	if sel.Wallpaper != nil {
		t.Error("Wallpaper should be nil when list is empty")
	}
}
