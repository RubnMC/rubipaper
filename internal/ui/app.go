package ui

import (
	"github.com/RubnMC/rubipaper/internal/config"
	"github.com/RubnMC/rubipaper/internal/domain"
	"github.com/charmbracelet/lipgloss"
	tea "github.com/charmbracelet/bubbletea"
)

type Model struct {
	components []Focusable
	focused    int
	keybinds   config.KeybindsConfig
	appearance config.AppearanceConfig
}

func New(items []domain.Wallpaper, wallpaperDir string, b domain.Backend, mode domain.WallpaperMode, keybinds config.KeybindsConfig, appearance config.AppearanceConfig) Model {
	optBar := NewOptionsBar(wallpaperDir, b, mode)
	fileExp := NewFileExplorer(items, wallpaperDir, b, mode)
	return Model{
		components: []Focusable{optBar, fileExp.Focus()},
		focused:    1,
		keybinds:   keybinds,
		appearance: appearance,
	}
}

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if key, ok := msg.(tea.KeyMsg); ok {
		switch key.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case m.keybinds.FocusNext:
			m.components[m.focused] = m.components[m.focused].Blur()
			m.focused = (m.focused + 1) % len(m.components)
			m.components[m.focused] = m.components[m.focused].Focus()
			return m, nil
		case m.keybinds.FocusPrev:
			m.components[m.focused] = m.components[m.focused].Blur()
			m.focused = (m.focused - 1 + len(m.components)) % len(m.components)
			m.components[m.focused] = m.components[m.focused].Focus()
			return m, nil
		}
	}

	var cmds []tea.Cmd
	for i, c := range m.components {
		updated, cmd := c.Update(msg)
		m.components[i] = updated.(Focusable)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}
	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	views := make([]string, len(m.components))
	for i, c := range m.components {
		style := lipgloss.NewStyle().Border(lipgloss.RoundedBorder())
		if i == m.focused {
			if m.appearance.FocusBorderColor != "" {
				style = style.BorderForeground(lipgloss.Color(m.appearance.FocusBorderColor))
			}
		} else {
			style = style.BorderForeground(lipgloss.Color("240"))
		}
		views[i] = style.Render(c.View())
	}
	return lipgloss.JoinVertical(lipgloss.Left, views...)
}
