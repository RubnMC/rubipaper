package config

import "github.com/RubnMC/rubipaper/internal/domain"

type baseConfig struct {
	dir            string
	mode           domain.WallpaperMode
	defaultBackend string
	backends       []string
}

type colorsConfig struct {
}

type keybindsConfig struct {
}

type Config struct {
	base     baseConfig
	colors   colorsConfig
	keybinds keybindsConfig
}
