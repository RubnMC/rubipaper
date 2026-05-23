package main

import (
	_ "embed"
	"flag"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/RubnMC/rubipaper/internal/backend"
	"github.com/RubnMC/rubipaper/internal/config"
	"github.com/RubnMC/rubipaper/internal/scanner"
	"github.com/RubnMC/rubipaper/internal/ui"
	tea "github.com/charmbracelet/bubbletea"
)

//go:embed configs/config.toml
var defaultConfigBytes []byte

func defaultConfigPath() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		return ""
	}
	return filepath.Join(dir, "rubipaper", "config.toml")
}

func expandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return filepath.Join(home, path[2:])
	}
	return path
}

func main() {
	configPath := flag.String("config", defaultConfigPath(), "path to config file")
	flag.Parse()

	cfg, err := config.Load(expandPath(*configPath), defaultConfigBytes)
	if err != nil {
		slog.Error("failed to load config", "path", *configPath, "err", err)
		os.Exit(1)
	}

	wallpaperDir := expandPath(cfg.Base.DefaultDir)
	images, err := scanner.ScanDirectory(wallpaperDir)
	if err != nil {
		slog.Error("failed to scan directory", "dir", wallpaperDir, "err", err)
		os.Exit(1)
	}
	if len(images) == 0 {
		slog.Warn("no images found", "dir", wallpaperDir)
		return
	}

	b, err := backend.NewBackend(cfg.Base.DefaultBackend)
	if err != nil {
		slog.Error("failed to initialize default backend", "backend", cfg.Base.DefaultBackend)
		os.Exit(1)
	}

	p := tea.NewProgram(ui.New(images, b, cfg.Base.DefaultMode))
	if _, err := p.Run(); err != nil {
		slog.Error("TUI error", "err", err)
		os.Exit(1)
	}
}
