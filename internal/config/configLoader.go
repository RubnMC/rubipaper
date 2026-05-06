package config

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
	"github.com/RubnMC/rubipaper/internal/domain"
)

// tomlBase mirrors baseConfig with exported fields for TOML decoding.
type tomlBase struct {
	DefaultDir     string   `toml:"default_dir"`
	DefaultMode    string   `toml:"default_mode"`
	DefaultBackend string   `toml:"default_backend"`
	Backends       []string `toml:"backends"`
}

type tomlConfig struct {
	Base tomlBase `toml:"base"`
}

// Load reads a TOML file at path and returns a populated Config.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	var raw tomlConfig
	if err := toml.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("parsing config file: %w", err)
	}

	// TODO(human): validate and map raw into Config

	_ = domain.WallpaperMode("") // ensure domain import is used until TODO is filled
	return nil, nil
}
