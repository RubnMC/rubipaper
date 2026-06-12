package ui

import (
	"os"
	"strconv"
	"strings"

	"github.com/RubnMC/rubipaper/internal/domain"
	tea "github.com/charmbracelet/bubbletea"
)

type OptionsBarModel struct {
	itemsPath string
	backend   domain.Backend
	mode      domain.WallpaperMode
	recursive bool
	focused   bool
}

func NewOptionsBar(itemsPath string, b domain.Backend, mode domain.WallpaperMode) OptionsBarModel {
	return OptionsBarModel{itemsPath: itemsPath, backend: b, mode: mode}
}

func (o OptionsBarModel) Init() tea.Cmd { return nil }

func (o OptionsBarModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if !o.focused {
			return o, nil
		}
		switch msg.String() {
		case "r":
			o.recursive = !o.recursive
			recursive := o.recursive
			return o, tea.Batch(
				scanWallpapersCmd(o.itemsPath, o.recursive),
				func() tea.Msg { return RecursiveToggledMsg{Recursive: recursive} },
			)
		}
	}
	return o, nil
}

func (o OptionsBarModel) View() string {
	var res strings.Builder
	res.WriteString("dir: ")
	res.WriteString(shortenHome(o.itemsPath))
	res.WriteString(" | backend: ")
	res.WriteString(o.backend.Name())
	res.WriteString(" | mode: ")
	res.WriteString(string(o.mode))
	res.WriteString(" | recursive search: ")
	res.WriteString(strconv.FormatBool(o.recursive))
	return res.String()
}

func (o OptionsBarModel) Focus() Focusable {
	o.focused = true
	return o
}

func (o OptionsBarModel) Blur() Focusable {
	o.focused = false
	return o
}

func (o OptionsBarModel) IsFocused() bool { return o.focused }

func shortenHome(path string) string {
	home, err := os.UserHomeDir()
	if err != nil || !strings.HasPrefix(path, home) {
		return path
	}
	return "~" + path[len(home):]
}
