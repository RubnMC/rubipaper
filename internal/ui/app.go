package ui

import tea "github.com/charmbracelet/bubbletea"

// Model is the main bubbletea model for rubipaper.
type Model struct{}

// New creates a new Model.
func New() Model {
	return Model{}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "q" || msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m Model) View() string {
	return "rubipaper v0.1.0\n\nPress q to quit.\n"
}
