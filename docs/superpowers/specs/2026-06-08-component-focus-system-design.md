# Component & Focus System Design

**Date:** 2026-06-08  
**Status:** Approved  
**Scope:** TUI component architecture and keyboard-driven focus routing. Does not extend the functionality of individual components.

---

## Overview

Split the current flat `Model` into composable, independently-focused components. The user navigates between components using configurable keyboard shortcuts. Only the focused component handles input. The parent model acts as a message bus and focus router.

---

## Component Interface

A `Focusable` interface is defined in `internal/ui/focusable.go`:

```go
type Focusable interface {
    tea.Model
    Focus() Focusable
    Blur()  Focusable
    IsFocused() bool
}
```

`Focus()` and `Blur()` return an updated copy of the component, following Bubble Tea's immutable update convention. Each component owns its focused state and uses it solely for rendering decisions (e.g., border color). The parent never inspects internal component state directly.

Because `tea.Model.Update` returns `(tea.Model, tea.Cmd)`, the parent retrieves updated components via type assertion after each `Update` call: `updated.(Focusable)`. Components must ensure their `Update` always returns a value that satisfies `Focusable` — failing the assertion is a programmer error and may panic.

---

## Parent Model

`internal/ui/app.go` is refactored into a pure router. The parent `Model` holds:

- `components []Focusable` — ordered list of all components
- `focused int` — index of the currently focused component
- `keybinds config.KeybindsConfig` — focus-switch key names
- `appearance config.AppearanceConfig` — border color config

### Focus Switching

On `tea.KeyMsg`, the parent checks if the key string matches `keybinds.FocusNext` or `keybinds.FocusPrev` before any other handling:

- Call `Blur()` on `components[focused]`
- Increment or decrement `focused` with wraparound (`mod len(components)`)
- Call `Focus()` on `components[focused]`

### Message Routing

All non-focus-switch messages are broadcast to every component in order. Each component's `Update` returns `(Focusable, tea.Cmd)`. The parent collects all returned `tea.Cmd`s into a `tea.Batch`. Components that do not handle a message return themselves unchanged.

This broadcast model means inter-component communication requires no parent-side routing logic: a child emits a custom `tea.Msg` via a `tea.Cmd`; the Bubble Tea runtime delivers it back to the root `Update`; the parent broadcasts it to all components; the intended recipient handles it.

### View

The parent calls `View()` on each component in slice order and joins the results vertically with `lipgloss.JoinVertical`. Each component's output is wrapped in a lipgloss border:

- **Focused:** `lipgloss.RoundedBorder()` with color from `appearance.FocusBorderColor` (empty string = terminal default foreground)
- **Unfocused:** `lipgloss.RoundedBorder()` with `lipgloss.Color("240")` (muted gray)

Border rendering is entirely the parent's responsibility. Components return plain content strings from `View()`.

---

## Components

### OptionsBar (`internal/ui/options_bar.go`)

Displays the current directory, backend, mode, and recursive-search toggle. Extracted verbatim from the existing top line of `app.go View()`. Holds the subset of state needed to render those fields.

### FileExplorer (`internal/ui/file_explorer.go`)

Displays the scrollable wallpaper list with cursor. Extracted from the existing list-rendering and navigation logic in `app.go`. Holds `items`, `cursor`, `status`, and `recursive` state.

---

## Configuration

### Keybinds (`internal/config/config.go`)

```go
type KeybindsConfig struct {
    FocusNext string
    FocusPrev string
}
```

### Appearance (`internal/config/config.go`)

```go
type AppearanceConfig struct {
    FocusBorderColor string
}
```

### Default TOML (`cmd/rubipaper/configs/config.toml`)

```toml
[keybinds]
focus_next = "shift+down"
focus_prev = "shift+up"

[appearance]
focus_border_color = ""   # empty = terminal default foreground
```

### TOML Loader (`internal/config/configLoader.go`)

Two new TOML mirror structs are added (`tomlKeybinds`, `tomlAppearance`) alongside the existing `tomlBase`. The `merge` function is extended to cover both. `newConfig` maps them into `KeybindsConfig` and `AppearanceConfig`.

---

## File Changes

| File | Change |
|---|---|
| `internal/ui/focusable.go` | New — `Focusable` interface |
| `internal/ui/options_bar.go` | New — `OptionsBarModel` implementing `Focusable` |
| `internal/ui/file_explorer.go` | New — `FileExplorerModel` implementing `Focusable` |
| `internal/ui/app.go` | Refactored — becomes parent router; existing logic moved to child models |
| `internal/config/config.go` | Fill in `KeybindsConfig` and `AppearanceConfig` structs |
| `internal/config/configLoader.go` | Add `tomlKeybinds`, `tomlAppearance`, extend merge and newConfig |
| `cmd/rubipaper/configs/config.toml` | Add `[keybinds]` and `[appearance]` sections |

---

## Out of Scope

- Extending the functionality of `OptionsBar` or `FileExplorer` (e.g., editable fields, new actions)
- More than two components — the slice-based design supports them, but none are added here
- Focus indicator beyond the lipgloss border (e.g., header labels, icons)
