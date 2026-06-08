package ui

import tea "github.com/charmbracelet/bubbletea"

// Focusable extends tea.Model with focus state management.
// Every Update implementation must return a value that satisfies Focusable —
// the parent router asserts this; a failed assertion is a programmer error.
type Focusable interface {
	tea.Model
	Focus() Focusable
	Blur() Focusable
	IsFocused() bool
}
