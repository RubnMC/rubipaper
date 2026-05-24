package config

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
)

// tomlBase mirrors BaseConfig with exported fields for TOML decoding.
type tomlBase struct {
	DefaultDir         string   `toml:"default_dir"`
	DefaultMode        string   `toml:"default_mode"`
	DefaultBackend     string   `toml:"default_backend"`
	Backends           []string `toml:"backends"`
	RecursiveDirSearch *bool    `toml:"recursive_dir_search"`
}

type tomlConfig struct {
	Base tomlBase `toml:"base"`
}

// merge applies non-zero fields from override onto base.
func merge(base, override tomlConfig) tomlConfig {
	if override.Base.DefaultDir != "" {
		base.Base.DefaultDir = override.Base.DefaultDir
	}
	if override.Base.DefaultMode != "" {
		base.Base.DefaultMode = override.Base.DefaultMode
	}
	if override.Base.DefaultBackend != "" {
		base.Base.DefaultBackend = override.Base.DefaultBackend
	}
	if len(override.Base.Backends) > 0 {
		base.Base.Backends = override.Base.Backends
	}
	if override.Base.RecursiveDirSearch != nil {
		base.Base.RecursiveDirSearch = override.Base.RecursiveDirSearch
	}
	return base
}

// Load reads a TOML config from path, falling back to fallback for any missing fields.
// If path does not exist, fallback is used entirely.
func Load(path string, fallback []byte) (*Config, error) {
	var defaults tomlConfig
	if err := toml.Unmarshal(fallback, &defaults); err != nil {
		return nil, fmt.Errorf("parsing default config: %w", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return newConfig(defaults)
		}
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	var user tomlConfig
	if err := toml.Unmarshal(data, &user); err != nil {
		return nil, fmt.Errorf("parsing config file: %w", err)
	}

	return newConfig(merge(defaults, user))
}
