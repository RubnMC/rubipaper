# Component & Focus System Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Split the flat TUI model into composable, focusable components connected by a broadcast message bus, with configurable keyboard shortcuts and lipgloss borders as the focus indicator.

**Architecture:** A `Focusable` interface extending `tea.Model` is implemented by `OptionsBarModel` and `FileExplorerModel`. The parent `Model` in `app.go` intercepts focus-switch keys and broadcasts all other messages to every component; components only act on key messages when `IsFocused()` returns true. The parent wraps each component's `View()` output in a lipgloss border — active color when focused, dim when not.

**Tech Stack:** Go 1.25, Bubble Tea v1.3.10, Lipgloss v1.1.0, BurntSushi/toml v1.6.0

---

## File Structure

| File | Action | Responsibility |
|---|---|---|
| `internal/config/config.go` | Modify | Fill in `KeybindsConfig` and `AppearanceConfig`; add `Keybinds` field to `Config` |
| `internal/config/configLoader.go` | Modify | Add `tomlKeybinds`, `tomlAppearance`; extend `tomlConfig`, `merge`, `newConfig` |
| `internal/config/config_test.go` | Create | Test keybinds and appearance default + override loading |
| `cmd/rubipaper/configs/config.toml` | Modify | Add `[keybinds]` and `[appearance]` sections |
| `internal/ui/focusable.go` | Create | `Focusable` interface |
| `internal/ui/messages.go` | Create | Shared UI message types (`RecursiveToggledMsg`) |
| `internal/ui/options_bar.go` | Create | `OptionsBarModel` — renders dir/backend/mode/recursive; handles `RecursiveToggledMsg` |
| `internal/ui/options_bar_test.go` | Create | Tests for focus state and message handling |
| `internal/ui/file_explorer.go` | Create | `FileExplorerModel` — list navigation, wallpaper set, recursive toggle with broadcast |
| `internal/ui/file_explorer_test.go` | Create | Tests for focus-gated input and recursive toggle |
| `internal/ui/app.go` | Rewrite | Parent router — focus switching, broadcast, lipgloss borders, shared helpers |
| `internal/ui/app_test.go` | Create | Tests for focus cycling and broadcast delivery |
| `cmd/rubipaper/main.go` | Modify | Pass `cfg.Keybinds` and `cfg.Colors` to `ui.New()` |

---

### Task 1: Extend config for keybinds and appearance

**Files:**
- Modify: `internal/config/config.go`
- Modify: `internal/config/configLoader.go`
- Modify: `cmd/rubipaper/configs/config.toml`
- Create: `internal/config/config_test.go`

- [ ] **Step 1: Write the failing tests**

Create `internal/config/config_test.go`:

```go
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
```

- [ ] **Step 2: Run test to verify it fails**

```bash
go test ./internal/config/...
```

Expected: compile error — `cfg.Keybinds` undefined, `cfg.Colors.FocusBorderColor` undefined.

- [ ] **Step 3: Update `internal/config/config.go`**

Replace the entire file:

```go
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
```

- [ ] **Step 4: Update `internal/config/configLoader.go`**

Replace the entire file:

```go
package config

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
)

type tomlBase struct {
	DefaultDir         string   `toml:"default_dir"`
	DefaultMode        string   `toml:"default_mode"`
	DefaultBackend     string   `toml:"default_backend"`
	Backends           []string `toml:"backends"`
	RecursiveDirSearch *bool    `toml:"recursive_dir_search"`
}

type tomlKeybinds struct {
	FocusNext string `toml:"focus_next"`
	FocusPrev string `toml:"focus_prev"`
}

type tomlAppearance struct {
	FocusBorderColor string `toml:"focus_border_color"`
}

type tomlConfig struct {
	Base       tomlBase       `toml:"base"`
	Keybinds   tomlKeybinds   `toml:"keybinds"`
	Appearance tomlAppearance `toml:"appearance"`
}

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
	if override.Keybinds.FocusNext != "" {
		base.Keybinds.FocusNext = override.Keybinds.FocusNext
	}
	if override.Keybinds.FocusPrev != "" {
		base.Keybinds.FocusPrev = override.Keybinds.FocusPrev
	}
	if override.Appearance.FocusBorderColor != "" {
		base.Appearance.FocusBorderColor = override.Appearance.FocusBorderColor
	}
	return base
}

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
```

- [ ] **Step 5: Update `cmd/rubipaper/configs/config.toml`**

Replace the entire file:

```toml
[base]
default_dir          = "~/Pictures/Wallpapers"
default_mode         = "fill"
default_backend      = "swaybg"
backends             = ["swaybg"]
recursive_dir_search = false

[keybinds]
focus_next = "shift+down"
focus_prev = "shift+up"

[appearance]
focus_border_color = ""
```

- [ ] **Step 6: Run tests to verify they pass**

```bash
go test ./internal/config/...
```

Expected: PASS (4 tests)

- [ ] **Step 7: Commit**

```bash
git add internal/config/config.go internal/config/configLoader.go internal/config/config_test.go cmd/rubipaper/configs/config.toml
git commit -m "feat: add keybinds and appearance config sections"
```

---

### Task 2: Define Focusable interface and shared message types

**Files:**
- Create: `internal/ui/focusable.go`
- Create: `internal/ui/messages.go`

- [ ] **Step 1: Create `internal/ui/focusable.go`**

```go
package ui

import tea "github.com/charmbracelet/bubbletea"

// Focusable extends tea.Model with focus state management.
// Every Update implementation must return a value that satisfies Focusable —
// the parent router asserts this; a failed assertion is a programmer error.
type Focusable interface {
	tea.Model
	Focus() Focusable
	Blur() Focusable
	IsFocused() bool
}
```

- [ ] **Step 2: Create `internal/ui/messages.go`**

```go
package ui

// RecursiveToggledMsg is broadcast when the file explorer toggles recursive
// directory search so other components can reflect the updated state.
type RecursiveToggledMsg struct {
	Recursive bool
}
```

- [ ] **Step 3: Verify it compiles**

```bash
go build ./internal/ui/...
```

Expected: success (no output)

- [ ] **Step 4: Commit**

```bash
git add internal/ui/focusable.go internal/ui/messages.go
git commit -m "feat: add Focusable interface and shared UI message types"
```

---

### Task 3: Implement OptionsBar component

**Files:**
- Create: `internal/ui/options_bar.go`
- Create: `internal/ui/options_bar_test.go`

- [ ] **Step 1: Write the failing tests**

Create `internal/ui/options_bar_test.go`:

```go
package ui

import (
	"strings"
	"testing"

	"github.com/RubnMC/rubipaper/internal/domain"
)

// stubBackend is shared across all ui tests in this package.
type stubBackend struct{}

func (s stubBackend) Name() string                           { return "stub" }
func (s stubBackend) IsAvaliable() bool                     { return true }
func (s stubBackend) SupportedModes() []domain.WallpaperMode { return nil }
func (s stubBackend) SetWallpaper(_, _ string) error        { return nil }
func (s stubBackend) GetCurrentWallpaper() (string, error)  { return "", nil }

func TestOptionsBarStartsUnfocused(t *testing.T) {
	o := NewOptionsBar("/home/user/pics", stubBackend{}, domain.ModeFill)
	if o.IsFocused() {
		t.Error("new OptionsBar should not be focused")
	}
}

func TestOptionsBarFocusReturnsFocusedCopy(t *testing.T) {
	o := NewOptionsBar("/home/user/pics", stubBackend{}, domain.ModeFill)
	focused := o.Focus()
	if !focused.IsFocused() {
		t.Error("Focus() should return a focused component")
	}
	if o.IsFocused() {
		t.Error("Focus() must not mutate the original")
	}
}

func TestOptionsBarBlurReturnsUnfocusedCopy(t *testing.T) {
	o := NewOptionsBar("/home/user/pics", stubBackend{}, domain.ModeFill)
	blurred := o.Focus().Blur()
	if blurred.IsFocused() {
		t.Error("Blur() should return an unfocused component")
	}
}

func TestOptionsBarRecursiveToggledMsg(t *testing.T) {
	o := NewOptionsBar("/home/user/pics", stubBackend{}, domain.ModeFill)
	updated, _ := o.Update(RecursiveToggledMsg{Recursive: true})
	ob := updated.(OptionsBarModel)
	if !ob.recursive {
		t.Error("OptionsBar should set recursive=true on RecursiveToggledMsg{Recursive:true}")
	}
}

func TestOptionsBarViewContainsFields(t *testing.T) {
	o := NewOptionsBar("/home/user/pics", stubBackend{}, domain.ModeFill)
	view := o.View()
	for _, want := range []string{"stub", "fill", "false"} {
		if !strings.Contains(view, want) {
			t.Errorf("View() missing %q", want)
		}
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
go test ./internal/ui/... -run TestOptionsBar
```

Expected: compile error — `NewOptionsBar` and `OptionsBarModel` undefined.

- [ ] **Step 3: Create `internal/ui/options_bar.go`**

```go
package ui

import (
	"os"
	"strconv"
	"strings"

	"github.com/RubnMC/rubipaper/internal/domain"
	tea "github.com/charmbracelet/bubbletea"
)

type OptionsBarModel struct {
	itemsPath string
	backend   domain.Backend
	mode      domain.WallpaperMode
	recursive bool
	focused   bool
}

func NewOptionsBar(itemsPath string, b domain.Backend, mode domain.WallpaperMode) OptionsBarModel {
	return OptionsBarModel{itemsPath: itemsPath, backend: b, mode: mode}
}

func (o OptionsBarModel) Init() tea.Cmd { return nil }

func (o OptionsBarModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m, ok := msg.(RecursiveToggledMsg); ok {
		o.recursive = m.Recursive
	}
	return o, nil
}

func (o OptionsBarModel) View() string {
	var res strings.Builder
	res.WriteString("dir: ")
	res.WriteString(shortenHome(o.itemsPath))
	res.WriteString(" | backend: ")
	res.WriteString(o.backend.Name())
	res.WriteString(" | mode: ")
	res.WriteString(string(o.mode))
	res.WriteString(" | recursive search: ")
	res.WriteString(strconv.FormatBool(o.recursive))
	return res.String()
}

func (o OptionsBarModel) Focus() Focusable {
	o.focused = true
	return o
}

func (o OptionsBarModel) Blur() Focusable {
	o.focused = false
	return o
}

func (o OptionsBarModel) IsFocused() bool { return o.focused }

func shortenHome(path string) string {
	home, err := os.UserHomeDir()
	if err != nil || !strings.HasPrefix(path, home) {
		return path
	}
	return "~" + path[len(home):]
}
```

- [ ] **Step 4: Run tests to verify they pass**

```bash
go test ./internal/ui/... -run TestOptionsBar
```

Expected: PASS (5 tests)

- [ ] **Step 5: Commit**

```bash
git add internal/ui/options_bar.go internal/ui/options_bar_test.go
git commit -m "feat: add OptionsBar component"
```

---

### Task 4: Implement FileExplorer component

**Files:**
- Create: `internal/ui/file_explorer.go`
- Create: `internal/ui/file_explorer_test.go`

- [ ] **Step 1: Write the failing tests**

Create `internal/ui/file_explorer_test.go`:

```go
package ui

import (
	"strings"
	"testing"

	"github.com/RubnMC/rubipaper/internal/domain"
	tea "github.com/charmbracelet/bubbletea"
)

func makeWallpapers(names ...string) []domain.Wallpaper {
	ws := make([]domain.Wallpaper, len(names))
	for i, n := range names {
		ws[i] = domain.Wallpaper{Path: "/pics/" + n, FileName: n}
	}
	return ws
}

func TestFileExplorerStartsUnfocused(t *testing.T) {
	f := NewFileExplorer(makeWallpapers("a.jpg"), "/pics", stubBackend{}, domain.ModeFill)
	if f.IsFocused() {
		t.Error("new FileExplorer should not be focused")
	}
}

func TestFileExplorerFocusBlur(t *testing.T) {
	f := NewFileExplorer(makeWallpapers("a.jpg"), "/pics", stubBackend{}, domain.ModeFill)
	if !f.Focus().IsFocused() {
		t.Error("Focus() should return focused component")
	}
	if f.Focus().Blur().IsFocused() {
		t.Error("Blur() should return unfocused component")
	}
	if f.IsFocused() {
		t.Error("Focus() must not mutate the original")
	}
}

func TestFileExplorerIgnoresKeysWhenUnfocused(t *testing.T) {
	f := NewFileExplorer(makeWallpapers("a.jpg", "b.jpg"), "/pics", stubBackend{}, domain.ModeFill)
	updated, _ := f.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	if updated.(FileExplorerModel).cursor != 0 {
		t.Error("unfocused FileExplorer must not move cursor on key press")
	}
}

func TestFileExplorerNavigatesDownAndUp(t *testing.T) {
	f := NewFileExplorer(makeWallpapers("a.jpg", "b.jpg", "c.jpg"), "/pics", stubBackend{}, domain.ModeFill)
	f = f.Focus().(FileExplorerModel)

	updated, _ := f.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	f = updated.(FileExplorerModel)
	if f.cursor != 1 {
		t.Errorf("cursor = %d, want 1 after j", f.cursor)
	}

	updated, _ = f.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
	f = updated.(FileExplorerModel)
	if f.cursor != 0 {
		t.Errorf("cursor = %d, want 0 after k", f.cursor)
	}
}

func TestFileExplorerCursorDoesNotOverflow(t *testing.T) {
	f := NewFileExplorer(makeWallpapers("a.jpg"), "/pics", stubBackend{}, domain.ModeFill)
	f = f.Focus().(FileExplorerModel)

	updated, _ := f.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
	if updated.(FileExplorerModel).cursor != 0 {
		t.Error("cursor should not go below 0")
	}

	updated, _ = f.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	if updated.(FileExplorerModel).cursor != 0 {
		t.Error("cursor should not go past last item with one item")
	}
}

func TestFileExplorerRecursiveToggleUpdatesState(t *testing.T) {
	f := NewFileExplorer(makeWallpapers("a.jpg"), "/pics", stubBackend{}, domain.ModeFill)
	f = f.Focus().(FileExplorerModel)

	updated, cmd := f.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("r")})
	fe := updated.(FileExplorerModel)
	if !fe.recursive {
		t.Error("recursive should be true after first toggle")
	}
	if cmd == nil {
		t.Error("toggle should return a non-nil cmd for scan and broadcast")
	}
}

func TestFileExplorerViewShowsItems(t *testing.T) {
	f := NewFileExplorer(makeWallpapers("wall.jpg", "bg.png"), "/pics", stubBackend{}, domain.ModeFill)
	view := f.View()
	if !strings.Contains(view, "wall.jpg") || !strings.Contains(view, "bg.png") {
		t.Error("View() should render all file names")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
go test ./internal/ui/... -run TestFileExplorer
```

Expected: compile error — `NewFileExplorer` and `FileExplorerModel` undefined.

- [ ] **Step 3: Create `internal/ui/file_explorer.go`**

```go
package ui

import (
	"strings"

	"github.com/RubnMC/rubipaper/internal/domain"
	"github.com/RubnMC/rubipaper/internal/scanner"
	tea "github.com/charmbracelet/bubbletea"
)

type wallpaperResult struct {
	err error
}

type scanResult struct {
	items []domain.Wallpaper
	err   error
}

type FileExplorerModel struct {
	items     []domain.Wallpaper
	itemsPath string
	cursor    int
	status    string
	recursive bool
	backend   domain.Backend
	mode      domain.WallpaperMode
	focused   bool
}

func NewFileExplorer(items []domain.Wallpaper, itemsPath string, b domain.Backend, mode domain.WallpaperMode) FileExplorerModel {
	return FileExplorerModel{items: items, itemsPath: itemsPath, backend: b, mode: mode}
}

func (f FileExplorerModel) Init() tea.Cmd { return nil }

func (f FileExplorerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if !f.focused {
			return f, nil
		}
		switch msg.String() {
		case "up", "k":
			if f.cursor > 0 {
				f.cursor--
			}
		case "down", "j":
			if f.cursor < len(f.items)-1 {
				f.cursor++
			}
		case "enter":
			if len(f.items) > 0 {
				return f, setWallpaperCmd(f.backend, f.items[f.cursor].Path, f.mode)
			}
		case "r":
			f.recursive = !f.recursive
			recursive := f.recursive
			return f, tea.Batch(
				scanWallpapersCmd(f.itemsPath, f.recursive),
				func() tea.Msg { return RecursiveToggledMsg{Recursive: recursive} },
			)
		}
	case wallpaperResult:
		if msg.err != nil {
			f.status = "error: " + msg.err.Error()
		} else if len(f.items) > 0 {
			f.status = "wallpaper set: " + f.items[f.cursor].FileName
		}
	case scanResult:
		if msg.err != nil {
			f.status = "error scanning: " + msg.err.Error()
		} else {
			f.items = msg.items
			f.cursor = 0
			f.status = ""
		}
	}
	return f, nil
}

func (f FileExplorerModel) View() string {
	var res strings.Builder
	for idx, item := range f.items {
		if idx == f.cursor {
			res.WriteString("-> ")
		} else {
			res.WriteString("   ")
		}
		res.WriteString(item.FileName)
		res.WriteString("\n")
	}
	if f.status != "" {
		res.WriteString("\n")
		res.WriteString(f.status)
		res.WriteString("\n")
	}
	return res.String()
}

func (f FileExplorerModel) Focus() Focusable {
	f.focused = true
	return f
}

func (f FileExplorerModel) Blur() Focusable {
	f.focused = false
	return f
}

func (f FileExplorerModel) IsFocused() bool { return f.focused }

func scanWallpapersCmd(path string, recursive bool) tea.Cmd {
	return func() tea.Msg {
		var res scanResult
		if recursive {
			res.items, res.err = scanner.ScanDirectoryRecursive(path)
		} else {
			res.items, res.err = scanner.ScanDirectory(path)
		}
		return res
	}
}

func setWallpaperCmd(b domain.Backend, path string, mode domain.WallpaperMode) tea.Cmd {
	return func() tea.Msg {
		return wallpaperResult{err: b.SetWallpaper(path, string(mode))}
	}
}
```

- [ ] **Step 4: Run tests to verify they pass**

```bash
go test ./internal/ui/... -run TestFileExplorer
```

Expected: PASS (7 tests)

- [ ] **Step 5: Commit**

```bash
git add internal/ui/file_explorer.go internal/ui/file_explorer_test.go
git commit -m "feat: add FileExplorer component"
```

---

### Task 5: Refactor app.go into parent router

**Files:**
- Rewrite: `internal/ui/app.go`
- Create: `internal/ui/app_test.go`

- [ ] **Step 1: Add lipgloss as a direct dependency**

```bash
go get github.com/charmbracelet/lipgloss@v1.1.0
```

Expected: the `go.mod` entry for lipgloss changes from `// indirect` to a direct dependency.

- [ ] **Step 2: Write the failing tests**

Create `internal/ui/app_test.go`:

```go
package ui

import (
	"testing"

	"github.com/RubnMC/rubipaper/internal/config"
	"github.com/RubnMC/rubipaper/internal/domain"
	tea "github.com/charmbracelet/bubbletea"
)

func testConfigs() (config.KeybindsConfig, config.AppearanceConfig) {
	return config.KeybindsConfig{FocusNext: "shift+down", FocusPrev: "shift+up"},
		config.AppearanceConfig{}
}

func TestInitialFocusOnFileExplorer(t *testing.T) {
	kb, ap := testConfigs()
	m := New(makeWallpapers("a.jpg"), "/pics", stubBackend{}, domain.ModeFill, kb, ap)
	if m.components[0].IsFocused() {
		t.Error("OptionsBar should not be focused on init")
	}
	if !m.components[1].IsFocused() {
		t.Error("FileExplorer should be focused on init")
	}
}

func TestFocusNextCyclesForward(t *testing.T) {
	kb, ap := testConfigs()
	m := New(makeWallpapers("a.jpg"), "/pics", stubBackend{}, domain.ModeFill, kb, ap)

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyShiftDown})
	m = updated.(Model)
	if m.focused != 0 {
		t.Errorf("focused = %d, want 0 after shift+down from index 1", m.focused)
	}
	if !m.components[0].IsFocused() {
		t.Error("component 0 should now be focused")
	}
	if m.components[1].IsFocused() {
		t.Error("component 1 should now be blurred")
	}
}

func TestFocusNextWrapsAround(t *testing.T) {
	kb, ap := testConfigs()
	m := New(makeWallpapers("a.jpg"), "/pics", stubBackend{}, domain.ModeFill, kb, ap)

	// Two shift+downs from index 1 should land back at index 1
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyShiftDown})
	updated, _ = updated.(Model).Update(tea.KeyMsg{Type: tea.KeyShiftDown})
	m = updated.(Model)
	if m.focused != 1 {
		t.Errorf("focused = %d, want 1 after full wrap", m.focused)
	}
}

func TestFocusPrevCyclesBackward(t *testing.T) {
	kb, ap := testConfigs()
	m := New(makeWallpapers("a.jpg"), "/pics", stubBackend{}, domain.ModeFill, kb, ap)

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyShiftUp})
	m = updated.(Model)
	if m.focused != 0 {
		t.Errorf("focused = %d, want 0 after shift+up from index 1", m.focused)
	}
}

func TestBroadcastDeliveredToAllComponents(t *testing.T) {
	kb, ap := testConfigs()
	m := New(makeWallpapers("a.jpg"), "/pics", stubBackend{}, domain.ModeFill, kb, ap)

	updated, _ := m.Update(RecursiveToggledMsg{Recursive: true})
	m = updated.(Model)
	ob := m.components[0].(OptionsBarModel)
	if !ob.recursive {
		t.Error("OptionsBar should receive RecursiveToggledMsg via broadcast")
	}
}
```

- [ ] **Step 3: Run test to verify it fails**

```bash
go test ./internal/ui/... -run "TestInitialFocus|TestFocusNext|TestFocusPrev|TestBroadcast"
```

Expected: compile error — `New` signature mismatch (`ui.New` still has old signature in `app.go`).

- [ ] **Step 4: Rewrite `internal/ui/app.go`**

```go
package ui

import (
	"github.com/RubnMC/rubipaper/internal/config"
	"github.com/RubnMC/rubipaper/internal/domain"
	"github.com/charmbracelet/lipgloss"
	tea "github.com/charmbracelet/bubbletea"
)

type Model struct {
	components []Focusable
	focused    int
	keybinds   config.KeybindsConfig
	appearance config.AppearanceConfig
}

func New(items []domain.Wallpaper, wallpaperDir string, b domain.Backend, mode domain.WallpaperMode, keybinds config.KeybindsConfig, appearance config.AppearanceConfig) Model {
	optBar := NewOptionsBar(wallpaperDir, b, mode)
	fileExp := NewFileExplorer(items, wallpaperDir, b, mode)
	return Model{
		components: []Focusable{optBar, fileExp.Focus()},
		focused:    1,
		keybinds:   keybinds,
		appearance: appearance,
	}
}

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if key, ok := msg.(tea.KeyMsg); ok {
		switch key.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case m.keybinds.FocusNext:
			m.components[m.focused] = m.components[m.focused].Blur()
			m.focused = (m.focused + 1) % len(m.components)
			m.components[m.focused] = m.components[m.focused].Focus()
			return m, nil
		case m.keybinds.FocusPrev:
			m.components[m.focused] = m.components[m.focused].Blur()
			m.focused = (m.focused - 1 + len(m.components)) % len(m.components)
			m.components[m.focused] = m.components[m.focused].Focus()
			return m, nil
		}
	}

	var cmds []tea.Cmd
	for i, c := range m.components {
		updated, cmd := c.Update(msg)
		m.components[i] = updated.(Focusable)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}
	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	views := make([]string, len(m.components))
	for i, c := range m.components {
		style := lipgloss.NewStyle().Border(lipgloss.RoundedBorder())
		if i == m.focused {
			if m.appearance.FocusBorderColor != "" {
				style = style.BorderForeground(lipgloss.Color(m.appearance.FocusBorderColor))
			}
		} else {
			style = style.BorderForeground(lipgloss.Color("240"))
		}
		views[i] = style.Render(c.View())
	}
	return lipgloss.JoinVertical(lipgloss.Left, views...)
}
```

- [ ] **Step 5: Run all ui tests**

```bash
go test ./internal/ui/...
```

Expected: PASS (all ui tests)

- [ ] **Step 6: Commit**

```bash
git add internal/ui/app.go internal/ui/app_test.go go.mod go.sum
git commit -m "feat: refactor app into parent router with focus system and lipgloss borders"
```

---

### Task 6: Wire up main.go

**Files:**
- Modify: `cmd/rubipaper/main.go`

- [ ] **Step 1: Update the `ui.New()` call**

In `cmd/rubipaper/main.go`, find:

```go
p := tea.NewProgram(ui.New(images, wallpaperDir, b, cfg.Base.DefaultMode))
```

Replace with:

```go
p := tea.NewProgram(ui.New(images, wallpaperDir, b, cfg.Base.DefaultMode, cfg.Keybinds, cfg.Colors))
```

- [ ] **Step 2: Build to verify**

```bash
go build ./cmd/rubipaper/...
```

Expected: success (no output)

- [ ] **Step 3: Run full test suite**

```bash
go test ./...
```

Expected: PASS

- [ ] **Step 4: Commit**

```bash
git add cmd/rubipaper/main.go
git commit -m "feat: wire keybinds and appearance config to TUI"
```
