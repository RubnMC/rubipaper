package ui

import (
	"strings"
	"testing"

	"github.com/RubnMC/rubipaper/internal/domain"
	tea "github.com/charmbracelet/bubbletea"
)

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
