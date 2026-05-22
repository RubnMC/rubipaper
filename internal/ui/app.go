package ui

import (
	"strings"

	"github.com/RubnMC/rubipaper/internal/domain"
	"github.com/RubnMC/rubipaper/internal/scanner"
	tea "github.com/charmbracelet/bubbletea"
)

type wallpaperResult struct {
	err error
}

type Model struct {
	items   []scanner.ImageFile
	cursor  int
	status  string
	backend domain.Backend
	mode    domain.WallpaperMode
}

func New(items []scanner.ImageFile, b domain.Backend, mode domain.WallpaperMode) Model {
	return Model{items: items, backend: b, mode: mode}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.items)-1 {
				m.cursor++
			}
		case "enter":
			if len(m.items) > 0 {
				return m, setWallpaperCmd(m.backend, m.items[m.cursor].Path, m.mode)
			}
		}
	case wallpaperResult:
		if msg.err != nil {
			m.status = "error: " + msg.err.Error()
		} else {
			m.status = "wallpaper set: " + m.items[m.cursor].Name
		}
	}
	return m, nil
}

func (m Model) View() string {
	var res strings.Builder
	for idx, item := range m.items {
		if idx == m.cursor {
			res.WriteString("-> ")
		} else {
			res.WriteString("   ")
		}
		res.WriteString(item.Name)
		res.WriteString("\n")
	}
	if m.status != "" {
		res.WriteString("\n" + m.status + "\n")
	}
	return res.String()
}

func setWallpaperCmd(b domain.Backend, path string, mode domain.WallpaperMode) tea.Cmd {
	return func() tea.Msg {
		return wallpaperResult{err: b.SetWallpaper(path, string(mode))}
	}
}
