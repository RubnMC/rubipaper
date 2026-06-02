package ui

import (
	"os"
	"strconv"
	"strings"

	"github.com/RubnMC/rubipaper/internal/domain"
	"github.com/RubnMC/rubipaper/internal/scanner"
	tea "github.com/charmbracelet/bubbletea"
)

type wallpaperResult struct {
	err error
}

type scanResult struct {
	items []domain.Wallpaper
	err   error
}

type Model struct {
	items     []domain.Wallpaper
	itemsPath string
	cursor    int
	status    string
	recursive bool
	backend   domain.Backend
	mode      domain.WallpaperMode
}

func New(items []domain.Wallpaper, wallpaperDir string, b domain.Backend, mode domain.WallpaperMode) Model {
	return Model{items: items, itemsPath: wallpaperDir, backend: b, mode: mode}
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
		case "r":
			m.recursive = !m.recursive
			return m, scanWallpapersCmd(m.itemsPath, m.recursive)
		}

	case wallpaperResult:
		if msg.err != nil {
			m.status = "error: " + msg.err.Error()
		} else {
			m.status = "wallpaper set: " + m.items[m.cursor].FileName
		}

	case scanResult:
		if msg.err != nil {
			m.status = "error scanning: " + msg.err.Error()
		} else {
			m.items = msg.items
			m.cursor = 0
			m.status = ""
		}
	}
	return m, nil
}

func (m Model) View() string {
	var res strings.Builder
	res.WriteString("dir: ")
	res.WriteString(shortenHome(m.itemsPath))
	res.WriteString(" | ")

	res.WriteString("backend: ")
	res.WriteString(m.backend.Name())
	res.WriteString(" | ")

	res.WriteString("mode: ")
	res.WriteString(string(m.mode))
	res.WriteString(" | ")

	res.WriteString("recursive search: ")
	res.WriteString(strconv.FormatBool(m.recursive))
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

func shortenHome(path string) string {
	home, err := os.UserHomeDir()
	if err != nil || !strings.HasPrefix(path, home) {
		return path
	}
	return "~" + path[len(home):]
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
