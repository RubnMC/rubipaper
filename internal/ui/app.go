package ui

import (
	"strings"

	"github.com/RubnMC/rubipaper/internal/domain"
	tea "github.com/charmbracelet/bubbletea"
)

type wallpaperResult struct {
	err error
}

type Model struct {
	items   []domain.Wallpaper
	cursor  int
	status  string
	backend domain.Backend
	mode    domain.WallpaperMode
}

func New(items []domain.Wallpaper, b domain.Backend, mode domain.WallpaperMode) Model {
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
			m.status = "wallpaper set: " + m.items[m.cursor].FileName
		}
	}
	return m, nil
}

func (m Model) View() string {
	var res strings.Builder
	res.WriteString("Backend: ")
	res.WriteString(m.backend.Name())
	res.WriteString(" | ")
	res.WriteString("Mode: ")
	res.WriteString(string(m.mode))
	res.WriteString("\n")

	for idx, item := range m.items {
		if idx == m.cursor {
			res.WriteString("-> ")
		} else {
			res.WriteString("   ")
		}
		res.WriteString(item.FileName)
		res.WriteString("\n")
	}
	if m.status != "" {
		res.WriteString("\n")
		res.WriteString(m.status)
		res.WriteString("\n")
	}
	return res.String()
}

func setWallpaperCmd(b domain.Backend, path string, mode domain.WallpaperMode) tea.Cmd {
	return func() tea.Msg {
		return wallpaperResult{err: b.SetWallpaper(path, string(mode))}
	}
}
