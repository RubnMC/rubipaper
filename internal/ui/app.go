package ui

import (
	"os"

	"github.com/RubnMC/rubipaper/internal/config"
	"github.com/RubnMC/rubipaper/internal/domain"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/lipgloss"
	tea "github.com/charmbracelet/bubbletea"
)

type Model struct {
	components   []Focusable
	previewPanel PreviewPanelModel
	focused      int
	keybinds     config.KeybindsConfig
	appearance   config.AppearanceConfig
	help         help.Model
	globalKeys   []key.Binding
	termWidth    int
	termHeight   int
}

func New(items []domain.Wallpaper, wallpaperDir string, b domain.Backend, mode domain.WallpaperMode, keybinds config.KeybindsConfig, appearance config.AppearanceConfig) Model {
	optBar := NewOptionsBar(wallpaperDir, b, mode)
	fileExp := NewFileExplorer(items, wallpaperDir, b, mode)
	return Model{
		components:   []Focusable{optBar, fileExp.Focus()},
		previewPanel: NewPreviewPanel(os.Getenv("KITTY_WINDOW_ID") != ""),
		focused:      1,
		keybinds:     keybinds,
		appearance:   appearance,
		help:         help.New(),
		globalKeys:   globalKeyMap(keybinds),
	}
}

func (m Model) Init() tea.Cmd {
	var cmds []tea.Cmd
	for _, c := range m.components {
		if cmd := c.Init(); cmd != nil {
			cmds = append(cmds, cmd)
		}
	}
	return tea.Batch(cmds...)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Copy components slice to avoid mutating the backing array of any previous Model.
	components := make([]Focusable, len(m.components))
	copy(components, m.components)
	m.components = components

	if ws, ok := msg.(tea.WindowSizeMsg); ok {
		m.termWidth = ws.Width
		m.termHeight = ws.Height
	}

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

	var ppCmd tea.Cmd
	m.previewPanel, ppCmd = m.previewPanel.Update(msg)
	if ppCmd != nil {
		cmds = append(cmds, ppCmd)
	}

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	// OptionsBar — full width
	optStyle := lipgloss.NewStyle().Border(lipgloss.NormalBorder())
	if m.focused == 0 && m.appearance.FocusBorderColor != "" {
		optStyle = optStyle.BorderForeground(lipgloss.Color(m.appearance.FocusBorderColor))
	} else {
		optStyle = optStyle.BorderForeground(lipgloss.Color("240"))
	}
	optView := optStyle.Render(m.components[0].View())

	// FileExplorer (40%) + PreviewPanel (60%) — horizontal split
	feStyle := lipgloss.NewStyle().Border(lipgloss.NormalBorder())
	// PreviewPanel is passive (non-focusable) — its border color never changes.
	ppStyle := lipgloss.NewStyle().Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240"))

	if m.termWidth > 0 {
		feTotal := int(float64(m.termWidth) * (1 - previewRatio))
		ppTotal := m.termWidth - feTotal
		feStyle = feStyle.Width(feTotal - 2)
		ppStyle = ppStyle.Width(ppTotal - 2)
	}

	if m.focused == 1 && m.appearance.FocusBorderColor != "" {
		feStyle = feStyle.BorderForeground(lipgloss.Color(m.appearance.FocusBorderColor))
	} else {
		feStyle = feStyle.BorderForeground(lipgloss.Color("240"))
	}

	middle := lipgloss.JoinHorizontal(lipgloss.Top,
		feStyle.Render(m.components[1].View()),
		ppStyle.Render(m.previewPanel.View()),
	)

	// Footer
	keys := append([]key.Binding{}, m.globalKeys...)
	if km, ok := m.components[m.focused].(ComponentKeyMap); ok {
		keys = append(keys, km.ShortHelp()...)
	}
	footer := m.help.ShortHelpView(keys)

	return lipgloss.JoinVertical(lipgloss.Left, optView, middle, footer)
}
