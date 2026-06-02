# rubipaper

> **Work in Progress** — the project is under active development. APIs, configuration keys, and supported backends will change without notice until a stable release is tagged.

A terminal UI wallpaper manager for Linux (Sway / Wayland).

---

## Table of Contents

1. [Overview](#overview)
2. [Architecture](#architecture)
3. [Prerequisites](#prerequisites)
4. [Installation & Local Development](#installation--local-development)
5. [CLI Usage](#cli-usage)
6. [Configuration](#configuration)
7. [Testing](#testing)
8. [Contributing](#contributing)

---

## Overview

`rubipaper` is a keyboard-driven TUI for browsing and applying desktop wallpapers on Wayland compositors. It scans a directory for images, displays them in a list, and sets the selected image as the wallpaper via `swaybg` or `swaymsg` (when running inside Sway).

**This is an early-stage project.** Features are incomplete, the config schema may evolve, and breaking changes can land on `main` at any time. Contributions and feedback are very welcome.

Key capabilities:

- Browse `.jpg`, `.jpeg`, `.png`, `.gif`, `.webp`, `.bmp`, `.tif`, and `.tiff` files in any local directory.
- Optional recursive directory scan (toggle with `r` at runtime).
- Pluggable backend interface — `swaybg` is the only implementation today, with more compositor backends planned.
- TOML configuration with embedded defaults; the binary works out of the box without a config file.

---

## Architecture

```
┌─────────────────────────────────────┐
│           cmd/rubipaper             │
│  main.go — wires config, scanner,   │
│            backend, and TUI         │
└────────────────┬────────────────────┘
                 │
     ┌───────────┼────────────────┐
     │           │                │
     ▼           ▼                ▼
┌─────────┐ ┌─────────┐  ┌────────────┐
│ config  │ │ scanner │  │  backend   │
│         │ │         │  │            │
│ Loads   │ │ Walks   │  │ Calls      │
│ TOML    │ │ dir for │  │ swaybg /   │
│ config  │ │ images  │  │ swaymsg    │
└─────────┘ └────┬────┘  └──────┬─────┘
                 │              │
                 └──────┬───────┘
                        ▼
                  ┌──────────┐
                  │    ui    │
                  │          │
                  │ Bubble   │
                  │ Tea TUI  │
                  └──────────┘

internal/domain  ←  shared types (Wallpaper, WallpaperMode, Backend interface)
                     no external dependencies
```

**Package responsibilities:**

| Package | Responsibility |
|---|---|
| `internal/domain` | Pure types: `Wallpaper`, `WallpaperMode`, `Backend` interface |
| `internal/config` | TOML loading with embedded defaults and user-override merging |
| `internal/scanner` | Flat and recursive directory scanning for image files |
| `internal/backend` | Concrete backends; `SwaybgBackend` auto-detects Sway via `swaymsg`. New backends are added here. |
| `internal/ui` | Bubble Tea model: list navigation, wallpaper application, recursive toggle |

---

## Prerequisites

| Requirement | Notes |
|---|---|
| Go 1.25.6+ | Matches `go.mod` |
| `swaybg` | Must be on `$PATH`; the only supported backend currently |
| `swaymsg` | Required when running inside a Sway session |
| Sway / Wayland | Only tested compositor environment for now; more targets are planned |

---

## Installation & Local Development

```bash
# Clone the repository
git clone https://github.com/RubnMC/rubipaper.git
cd rubipaper

# Download dependencies
go mod download

# Build the binary
go build -o rubipaper ./cmd/rubipaper/...

# Run directly (no build step needed)
go run ./cmd/rubipaper/...

# Lint (requires golangci-lint)
golangci-lint run
```

To install the binary to your Go bin directory:

```bash
go install ./cmd/rubipaper/...
```

---

## CLI Usage

```
rubipaper [flags]

Flags:
  -config <path>   Path to the TOML configuration file.
                   Default: $XDG_CONFIG_HOME/rubipaper/config.toml
                            (usually ~/.config/rubipaper/config.toml)
```

### TUI keybindings

| Key | Action |
|---|---|
| `↑` / `k` | Move cursor up |
| `↓` / `j` | Move cursor down |
| `Enter` | Apply selected wallpaper |
| `r` | Toggle recursive directory scan |
| `q` / `Ctrl+C` | Quit |

The status bar at the top of the TUI shows the current directory, active backend, wallpaper mode, and whether recursive search is enabled.

---

## Configuration

rubipaper reads a TOML file. If the file does not exist, the embedded defaults (shown below) are used without error.

**Default location:** `~/.config/rubipaper/config.toml`

```toml
[base]
default_dir          = "~/Pictures/Wallpapers"
default_mode         = "fill"
default_backend      = "swaybg"
backends             = ["swaybg"]
recursive_dir_search = false
```

### Configuration reference

| Key | Type | Description | Default | Required |
|---|---|---|---|---|
| `base.default_dir` | string | Directory to scan for wallpapers. Supports `~` expansion. | `~/Pictures/Wallpapers` | Yes |
| `base.default_mode` | string | Wallpaper scaling mode: `fill`, `center`, `tile`, `stretch`, `fit` | `fill` | Yes |
| `base.default_backend` | string | Backend to use for setting wallpapers. Must be one of `base.backends`. | `swaybg` | Yes |
| `base.backends` | string list | Backends to enable. Currently only `swaybg` is implemented; more will be added. | `["swaybg"]` | Yes |
| `base.recursive_dir_search` | bool | Whether to scan subdirectories on startup. Can also be toggled at runtime with `r`. | `false` | No |

To use a custom config file:

```bash
rubipaper -config /path/to/myconfig.toml
```

---

## Testing

```bash
# Run all tests
go test ./...

# Run tests for a specific package
go test ./internal/scanner/...

# Run tests with verbose output
go test -v ./...

# Run tests with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

---

## Contributing

The project is in early development and the scope is intentionally expanding — new backends, richer config, and TUI improvements are all on the horizon. Contributions are very welcome.

1. Fork the repository and create a feature branch from `main`.
2. Keep commits focused; one logical change per commit.
3. Run `go test ./...` and `golangci-lint run` before opening a pull request.
4. Open a pull request against `main` with a clear description of what changed and why.
5. If you're planning a larger change (new backend, config schema addition), open an issue first to align on the approach before investing time in an implementation.
