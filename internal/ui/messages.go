package ui

import "github.com/RubnMC/rubipaper/internal/domain"

// RecursiveToggledMsg is broadcast when the file explorer toggles recursive
// directory search so other components can reflect the updated state.
type RecursiveToggledMsg struct {
	Recursive bool
}

// WallpaperSelectedMsg is broadcast when the cursor moves in the file explorer.
// Wallpaper is nil when the list is empty.
type WallpaperSelectedMsg struct {
	Wallpaper *domain.Wallpaper
}
