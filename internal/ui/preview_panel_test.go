package ui

import (
	"image"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/RubnMC/rubipaper/internal/domain"
	tea "github.com/charmbracelet/bubbletea"
)

// makeTempImage creates a temp PNG file and returns its path, following the
// same pattern as internal/scanner/scanner_test.go's real-decode tests.
func makeTempImage(t *testing.T, w, h int) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "test.png")
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err := png.Encode(f, image.NewRGBA(image.Rect(0, 0, w, h))); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestPreviewPanelStartsWithNoSelection(t *testing.T) {
	p := NewPreviewPanel(false)
	if p.selected != nil {
		t.Error("new PreviewPanel should have nil selection")
	}
}

func TestPreviewPanelUpdatesOnWallpaperSelected(t *testing.T) {
	p := NewPreviewPanel(false)
	w := &domain.Wallpaper{FileName: "test.jpg", Path: "/pics/test.jpg"}
	p, _ = p.Update(WallpaperSelectedMsg{Wallpaper: w})
	if p.selected != w {
		t.Error("PreviewPanel should store the selected wallpaper pointer")
	}
}

func TestPreviewPanelShowsMetadata(t *testing.T) {
	p := NewPreviewPanel(false)
	w := &domain.Wallpaper{
		FileName:   "bg.jpg",
		Path:       "/pics/bg.jpg",
		Resolution: domain.Resolution{Width: 1920, Height: 1080},
		FileSize:   domain.FileSize(2 * 1024 * 1024),
	}
	p, _ = p.Update(WallpaperSelectedMsg{Wallpaper: w})
	view := p.View()

	for _, want := range []string{"bg.jpg", "/pics/bg.jpg", "1920×1080", "2.0 MB"} {
		if !strings.Contains(view, want) {
			t.Errorf("View() missing %q", want)
		}
	}
}

func TestPreviewPanelShowsNoPreviewWithoutKitty(t *testing.T) {
	p := NewPreviewPanel(false)
	p, _ = p.Update(WallpaperSelectedMsg{Wallpaper: &domain.Wallpaper{FileName: "x.jpg", Path: "/x.jpg"}})
	if !strings.Contains(p.View(), "kitty") {
		t.Error("View() should explain that kitty terminal is required")
	}
}

func TestPreviewPanelHandlesNilSelection(t *testing.T) {
	p := NewPreviewPanel(false)
	p, _ = p.Update(WallpaperSelectedMsg{Wallpaper: nil})
	view := p.View()
	if strings.Contains(view, "1920") {
		t.Error("View() should not show resolution for nil selection")
	}
}

func TestPreviewPanelHandlesWindowResize(t *testing.T) {
	p := NewPreviewPanel(false)
	p, _ = p.Update(tea.WindowSizeMsg{Width: 200, Height: 50})
	if p.width == 0 {
		t.Error("PreviewPanel should update width on WindowSizeMsg")
	}
}

func TestPreviewPanelRerendersImageOnResize(t *testing.T) {
	imgPath := makeTempImage(t, 400, 200)
	p := NewPreviewPanel(true)

	p, _ = p.Update(tea.WindowSizeMsg{Width: 100, Height: 30})

	var cmd tea.Cmd
	p, cmd = p.Update(WallpaperSelectedMsg{Wallpaper: &domain.Wallpaper{FileName: "test.png", Path: imgPath}})
	if cmd == nil {
		t.Fatal("selecting a wallpaper with kitty support should return a render Cmd")
	}
	p, _ = p.Update(cmd())
	firstRendered := p.renderedImg
	if firstRendered == "" {
		t.Fatal("expected renderedImg to be populated after the initial render")
	}

	p, cmd = p.Update(tea.WindowSizeMsg{Width: 300, Height: 30})
	if cmd == nil {
		t.Fatal("resizing while a wallpaper is selected should return a render Cmd")
	}
	if !p.loading || p.renderedImg != "" {
		t.Error("resize should clear the stale renderedImg and show loading until the recentered image is ready")
	}

	p, _ = p.Update(cmd())
	if p.renderedImg == "" {
		t.Fatal("expected renderedImg to be populated after the resize re-render")
	}
	if p.renderedImg == firstRendered {
		t.Error("renderedImg should differ after resize — image should be recentered/rescaled for the new width")
	}
}

func TestPreviewPanelResizeSkipsRenderWithoutKittySupport(t *testing.T) {
	p := NewPreviewPanel(false)
	p, _ = p.Update(WallpaperSelectedMsg{Wallpaper: &domain.Wallpaper{FileName: "x.jpg", Path: "/x.jpg"}})

	_, cmd := p.Update(tea.WindowSizeMsg{Width: 200, Height: 50})
	if cmd != nil {
		t.Error("resize should not trigger a render Cmd without kitty support")
	}
}

func TestPreviewPanelResizeSkipsRenderWithoutSelection(t *testing.T) {
	p := NewPreviewPanel(true)

	_, cmd := p.Update(tea.WindowSizeMsg{Width: 200, Height: 50})
	if cmd != nil {
		t.Error("resize should not trigger a render Cmd with no wallpaper selected")
	}
}

func TestFormatFileSizeBytes(t *testing.T) {
	tests := []struct {
		size domain.FileSize
		want string
	}{
		{512, "512 B"},
		{1536, "1.5 KB"},
		{domain.FileSize(2 * 1024 * 1024), "2.0 MB"},
	}
	for _, tc := range tests {
		got := formatFileSize(tc.size)
		if got != tc.want {
			t.Errorf("formatFileSize(%d) = %q, want %q", tc.size, got, tc.want)
		}
	}
}
