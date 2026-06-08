package config

import (
	"fmt"
	"slices"
	"strings"

	"github.com/RubnMC/rubipaper/internal/domain"
)

type BaseConfig struct {
	DefaultDir     string
	DefaultMode    domain.WallpaperMode
	DefaultBackend string
	Backends       []string
}

type AppearanceConfig struct {
	FocusBorderColor string
}

type KeybindsConfig struct {
	FocusNext string
	FocusPrev string
}

type Config struct {
	Base     BaseConfig
	Colors   AppearanceConfig
	Keybinds KeybindsConfig
}

type validator func(raw tomlConfig) error

var baseValidators = []validator{
	func(raw tomlConfig) error {
		if raw.Base.DefaultDir == "" {
			return fmt.Errorf("base.default_dir is required")
		}
		return nil
	},
	func(raw tomlConfig) error {
		if raw.Base.DefaultMode == "" {
			return fmt.Errorf("base.default_mode is required")
		}
		if slices.Contains(domain.ValidModes, domain.WallpaperMode(raw.Base.DefaultMode)) {
			return nil
		}
		names := make([]string, len(domain.ValidModes))
		for i, m := range domain.ValidModes {
			names[i] = string(m)
		}
		return fmt.Errorf("base.default_mode %q must be one of: %s", raw.Base.DefaultMode, strings.Join(names, ", "))
	},
	func(raw tomlConfig) error {
		if raw.Base.DefaultBackend == "" {
			return fmt.Errorf("base.default_backend is required")
		}
		return nil
	},
	func(raw tomlConfig) error {
		if len(raw.Base.Backends) == 0 {
			return fmt.Errorf("base.backends is required and must not be empty")
		}
		return nil
	},
}

func newConfig(raw tomlConfig) (*Config, error) {
	for _, v := range baseValidators {
		if err := v(raw); err != nil {
			return nil, err
		}
	}

	p := new(Config)
	p.Base.DefaultDir = raw.Base.DefaultDir
	p.Base.DefaultMode = domain.WallpaperMode(raw.Base.DefaultMode)
	p.Base.DefaultBackend = raw.Base.DefaultBackend
	p.Base.Backends = raw.Base.Backends
	p.Keybinds.FocusNext = raw.Keybinds.FocusNext
	p.Keybinds.FocusPrev = raw.Keybinds.FocusPrev
	p.Colors.FocusBorderColor = raw.Appearance.FocusBorderColor

	return p, nil
}
