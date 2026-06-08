package ui

// RecursiveToggledMsg is broadcast when the file explorer toggles recursive
// directory search so other components can reflect the updated state.
type RecursiveToggledMsg struct {
	Recursive bool
}
