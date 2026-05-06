package config

import "github.com/RubnMC/rubipaper/internal/domain"

type BaseConfig struct {
	defaultDir     string
	defaultMode    domain.WallpaperMode
	defaultBackend string
	backends       []string
}

type ThemeConfig struct {
}

type KeybindsConfig struct {
}

type Config struct {
	base     BaseConfig
	colors   ThemeConfig
	keybinds KeybindsConfig
}
