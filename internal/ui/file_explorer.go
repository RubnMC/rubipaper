package ui

import (
	"strings"

	"github.com/RubnMC/rubipaper/internal/domain"
	"github.com/RubnMC/rubipaper/internal/scanner"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

var (
	keyUp              = key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "up"))
	keyDown            = key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "down"))
	keyEnter           = key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "set wallpaper"))
	keyToggleRecursive = key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "toggle recursive search"))
)

type wallpaperResult struct {
	err error
}

type scanResult struct {
	items []domain.Wallpaper
	err   error
}

type FileExplorerModel struct {
	items     []domain.Wallpaper
	itemsPath string
	cursor    int
	status    string
	recursive bool
	backend   domain.Backend
	mode      domain.WallpaperMode
	focused   bool
}

func NewFileExplorer(items []domain.Wallpaper, itemsPath string, b domain.Backend, mode domain.WallpaperMode) FileExplorerModel {
	return FileExplorerModel{items: items, itemsPath: itemsPath, backend: b, mode: mode}
}

func (f FileExplorerModel) Init() tea.Cmd { return nil }

func (f FileExplorerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if !f.focused {
			return f, nil
		}
		switch {
		case key.Matches(msg, keyUp):
			if f.cursor > 0 {
				f.cursor--
			}
			return f, f.selectedCmd()
		case key.Matches(msg, keyDown):
			if f.cursor < len(f.items)-1 {
				f.cursor++
			}
			return f, f.selectedCmd()
		case key.Matches(msg, keyEnter):
			if len(f.items) > 0 {
				return f, setWallpaperCmd(f.backend, f.items[f.cursor].Path, f.mode)
			}
		case key.Matches(msg, keyToggleRecursive):
			f.recursive = !f.recursive
			recursive := f.recursive
			return f, tea.Batch(
				scanWallpapersCmd(f.itemsPath, f.recursive),
				func() tea.Msg { return RecursiveToggledMsg{Recursive: recursive} },
			)
		}
	case wallpaperResult:
		if msg.err != nil {
			f.status = "error: " + msg.err.Error()
		} else if len(f.items) > 0 {
			f.status = "wallpaper set: " + f.items[f.cursor].FileName
		}
	case scanResult:
		if msg.err != nil {
			f.status = "error scanning: " + msg.err.Error()
		} else {
			f.items = msg.items
			f.cursor = 0
			f.status = ""
			return f, f.selectedCmd()
		}
	}
	return f, nil
}

func (f FileExplorerModel) View() string {
	var res strings.Builder
	for idx, item := range f.items {
		if idx == f.cursor {
			res.WriteString("-> ")
		} else {
			res.WriteString("   ")
		}
		res.WriteString(item.FileName)
		res.WriteString("\n")
	}
	if f.status != "" {
		res.WriteString("\n")
		res.WriteString(f.status)
		res.WriteString("\n")
	}
	return res.String()
}

func (f FileExplorerModel) Focus() Focusable {
	f.focused = true
	return f
}

func (f FileExplorerModel) Blur() Focusable {
	f.focused = false
	return f
}

func (f FileExplorerModel) IsFocused() bool { return f.focused }

func (f FileExplorerModel) ShortHelp() []key.Binding {
	return []key.Binding{keyUp, keyDown, keyEnter, keyToggleRecursive}
}

func (f FileExplorerModel) selectedCmd() tea.Cmd {
	if len(f.items) == 0 {
		return func() tea.Msg { return WallpaperSelectedMsg{Wallpaper: nil} }
	}
	w := f.items[f.cursor]
	return func() tea.Msg { return WallpaperSelectedMsg{Wallpaper: &w} }
}

func scanWallpapersCmd(path string, recursive bool) tea.Cmd {
	return func() tea.Msg {
		var res scanResult
		if recursive {
			res.items, res.err = scanner.ScanDirectoryRecursive(path)
		} else {
			res.items, res.err = scanner.ScanDirectory(path)
		}
		return res
	}
}

func setWallpaperCmd(b domain.Backend, path string, mode domain.WallpaperMode) tea.Cmd {
	return func() tea.Msg {
		return wallpaperResult{err: b.SetWallpaper(path, string(mode))}
	}
}
