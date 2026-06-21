package ui

import (
	"github.com/RubnMC/rubipaper/internal/config"
	"github.com/RubnMC/rubipaper/internal/domain"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/lipgloss"
	tea "github.com/charmbracelet/bubbletea"
)

type Model struct {
	components []Focusable
	focused    int
	keybinds   config.KeybindsConfig
	appearance config.AppearanceConfig
	help       help.Model
	globalKeys []key.Binding
}

func New(items []domain.Wallpaper, wallpaperDir string, b domain.Backend, mode domain.WallpaperMode, keybinds config.KeybindsConfig, appearance config.AppearanceConfig) Model {
	optBar := NewOptionsBar(wallpaperDir, b, mode)
	fileExp := NewFileExplorer(items, wallpaperDir, b, mode)
	return Model{
		components: []Focusable{optBar, fileExp.Focus()},
		focused:    1,
		keybinds:   keybinds,
		appearance: appearance,
		help:       help.New(),
		globalKeys: globalKeyMap(keybinds),
	}
}

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Copy components slice to avoid mutating the backing array of any previous Model.
	components := make([]Focusable, len(m.components))
	copy(components, m.components)
	m.components = components

	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		if key.Matches(keyMsg, keyQuit) {
			return m, tea.Quit
		}
		if m.keybinds.FocusNext != "" && key.Matches(keyMsg, key.NewBinding(key.WithKeys(m.keybinds.FocusNext))) {
			m.components[m.focused] = m.components[m.focused].Blur()
			m.focused = (m.focused + 1) % len(m.components)
			m.components[m.focused] = m.components[m.focused].Focus()
			return m, nil
		}
		if m.keybinds.FocusPrev != "" && key.Matches(keyMsg, key.NewBinding(key.WithKeys(m.keybinds.FocusPrev))) {
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

	keys := append([]key.Binding{}, m.globalKeys...)
	if km, ok := m.components[m.focused].(ComponentKeyMap); ok {
		keys = append(keys, km.ShortHelp()...)
	}
	footer := m.help.ShortHelpView(keys)

	return lipgloss.JoinVertical(lipgloss.Left, append(views, footer)...)
}
