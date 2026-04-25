package main

import (
	"log/slog"
	"os"

	"github.com/RubnMC/rubipaper/internal/backend"
	"github.com/RubnMC/rubipaper/internal/scanner"
	"github.com/RubnMC/rubipaper/internal/ui"
	tea "github.com/charmbracelet/bubbletea"
)

const defaultWallpaperDir = "/home/ruben/Pictures/Wallpapers"

func main() {
	b := backend.NewSwaybgBackend()
	images, err := scanner.ScanDirectory(defaultWallpaperDir)
	if err != nil {
		slog.Error("failed to scan directory", "dir", defaultWallpaperDir, "err", err)
		os.Exit(1)
	}
	if len(images) == 0 {
		slog.Warn("no images found", "dir", defaultWallpaperDir)
		return
	}

	p := tea.NewProgram(ui.New(images, b))
	if _, err := p.Run(); err != nil {
		slog.Error("TUI error", "err", err)
		os.Exit(1)
	}
}
