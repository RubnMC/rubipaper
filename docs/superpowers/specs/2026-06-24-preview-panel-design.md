# Preview Panel Design

**Date:** 2026-06-24  
**Issue:** #12 — Add a preview container next to image explorer  
**Milestone:** v0.1 - MVP

---

## Context

The file explorer currently lists wallpapers as plain text with no additional information. Issue #12 asks for a preview panel next to the file explorer that shows the selected image (if the terminal supports it) and its metadata. The user runs Kitty terminal, so the Kitty graphics protocol is the primary rendering target, with a graceful fallback to metadata-only for other terminals.

---

## Layout

The current vertical stack gains a horizontal split in the middle row:

```
┌─────────────────────────────────────────────────┐
│  OptionsBar  (full width, unchanged)            │
├──────────────────┬──────────────────────────────┤
│  FileExplorer    │  PreviewPanel                │
│  (40%)           │  (60%)                       │
│                  │  [image or placeholder]      │
│                  │                              │
│                  │  filename.jpg                │
│                  │  ~/Pictures/Wallpapers/...   │
│                  │  1920×1080 · 2.3 MB          │
├──────────────────┴──────────────────────────────┤
│  Footer (key hints, unchanged)                  │
└─────────────────────────────────────────────────┘
```

`app.go` composes this as:
```go
middle := lipgloss.JoinHorizontal(lipgloss.Top, fileExplorerView, previewView)
lipgloss.JoinVertical(lipgloss.Left, optionsView, middle, footerView)
```

Width is derived from `tea.WindowSizeMsg`: FileExplorer = 40%, PreviewPanel = 60%, each minus their border widths. Both components receive their width/height via `tea.WindowSizeMsg` routing.

---

## Components & Data Flow

### PreviewPanelModel

A **passive, non-focusable** component held as a plain field in `Model` (not in `components []Focusable`). The router dispatches messages to it explicitly.

```go
type PreviewPanelModel struct {
    selected     *domain.Wallpaper  // nil = nothing selected
    renderedImg  string             // pre-rendered Kitty escape sequence string
    loading      bool
    kittySupport bool               // set once at startup via $KITTY_WINDOW_ID
    width        int                // in terminal cells
    height       int
}
```

### Message flow

1. FileExplorer emits `WallpaperSelectedMsg` on every cursor move and on scan completion.
2. `app.go` routes it to `previewPanel.Update()`.
3. PreviewPanel triggers an async `tea.Cmd` that decodes the image and renders it.
4. An internal `imageRenderedMsg` carries the rendered string back to update state.

New message in `messages.go`:
```go
type WallpaperSelectedMsg struct {
    Wallpaper *domain.Wallpaper  // nil if list is empty
}
```

### Scanner enrichment

`Wallpaper.Resolution` and `Wallpaper.FileSize` are currently always zero-valued. The scanner must populate them:
- `image.DecodeConfig()` (stdlib) — reads dimensions without decoding the full image
- `os.Stat()` — file size

This is a prerequisite for the metadata display in the preview panel.

---

## Kitty Image Rendering

### Library

Use `github.com/dolmen-go/kittyimg` (Apache 2.0).  
Vendor via `go get` + `go mod vendor` — the vendor directory includes the LICENSE file automatically.

### Detection

```go
kittySupport := os.Getenv("KITTY_WINDOW_ID") != ""
```

Checked once at startup in `New()`, stored in `PreviewPanelModel`.

### Async rendering

Image loading is async to avoid blocking the UI:

```go
func renderImageCmd(path string) tea.Cmd {
    return func() tea.Msg {
        f, err := os.Open(path)
        if err != nil { return imageRenderedMsg{err: err} }
        defer f.Close()
        img, _, err := image.Decode(f)
        if err != nil { return imageRenderedMsg{err: err} }
        var buf strings.Builder
        kittyimg.Fprint(&buf, img)
        return imageRenderedMsg{rendered: buf.String()}
    }
}
```

### View composition

```
[renderedImg string — escape sequence, image renders here]
[N blank lines reserved for image height in layout]
filename.jpg
~/Pictures/Wallpapers/filename.jpg
1920×1080 · 2.3 MB
```

If `kittySupport == false` or `loading == true` or `selected == nil`, the image area shows a placeholder string (`[no preview]`, `loading…`, or `no selection`).

---

## Files Changed

| File | Change |
|------|--------|
| `internal/scanner/scanner.go` | Populate `Resolution` via `image.DecodeConfig`, `FileSize` via `os.Stat` |
| `internal/ui/messages.go` | Add `WallpaperSelectedMsg` |
| `internal/ui/file_explorer.go` | Emit `WallpaperSelectedMsg` on cursor move + scan complete |
| `internal/ui/preview_panel.go` | New passive component |
| `internal/ui/app.go` | Add `previewPanel` field, `width` tracking, horizontal layout, message routing |
| `go.mod` / `go.sum` / `vendor/` | Add `dolmen-go/kittyimg` dependency |

---

## Verification

1. `go build ./...` — clean compile
2. `go test ./...` — existing tests pass; new tests for scanner enrichment and preview panel
3. Run the app in Kitty: select different images with `j`/`k`, confirm preview updates and image renders
4. Run the app in a non-Kitty terminal: confirm metadata displays and no escape sequence garbage appears
5. Resize the terminal window: confirm layout reflows correctly
