package config

import (
	"os"
	"path/filepath"
	"testing"
)

var testDefaults = []byte(`
[base]
default_dir     = "~/Pictures"
default_mode    = "fill"
default_backend = "swaybg"
backends        = ["swaybg"]

[keybinds]
focus_next = "shift+down"
focus_prev = "shift+up"

[appearance]
focus_border_color = ""
`)

func TestKeybindsDefaults(t *testing.T) {
	cfg, err := Load("nonexistent_path_xyz", testDefaults)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Keybinds.FocusNext != "shift+down" {
		t.Errorf("FocusNext = %q, want %q", cfg.Keybinds.FocusNext, "shift+down")
	}
	if cfg.Keybinds.FocusPrev != "shift+up" {
		t.Errorf("FocusPrev = %q, want %q", cfg.Keybinds.FocusPrev, "shift+up")
	}
}

func TestAppearanceDefaults(t *testing.T) {
	cfg, err := Load("nonexistent_path_xyz", testDefaults)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Colors.FocusBorderColor != "" {
		t.Errorf("FocusBorderColor = %q, want empty", cfg.Colors.FocusBorderColor)
	}
}

func TestKeybindsUserOverride(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.toml")
	if err := os.WriteFile(path, []byte(`
[base]
default_dir     = "~/Pictures"
default_mode    = "fill"
default_backend = "swaybg"
backends        = ["swaybg"]

[keybinds]
focus_next = "tab"
focus_prev = "shift+tab"
`), 0600); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(path, testDefaults)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Keybinds.FocusNext != "tab" {
		t.Errorf("FocusNext = %q, want %q", cfg.Keybinds.FocusNext, "tab")
	}
	if cfg.Keybinds.FocusPrev != "shift+tab" {
		t.Errorf("FocusPrev = %q, want %q", cfg.Keybinds.FocusPrev, "shift+tab")
	}
}

func TestAppearanceUserOverride(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.toml")
	if err := os.WriteFile(path, []byte(`
[base]
default_dir     = "~/Pictures"
default_mode    = "fill"
default_backend = "swaybg"
backends        = ["swaybg"]

[appearance]
focus_border_color = "#cba6f7"
`), 0600); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(path, testDefaults)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Colors.FocusBorderColor != "#cba6f7" {
		t.Errorf("FocusBorderColor = %q, want %q", cfg.Colors.FocusBorderColor, "#cba6f7")
	}
}
