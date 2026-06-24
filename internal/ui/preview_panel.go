package ui

import (
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"strings"

	"github.com/RubnMC/rubipaper/internal/domain"
	"github.com/dolmen-go/kittyimg"
	tea "github.com/charmbracelet/bubbletea"
	"golang.org/x/image/draw"
)

const (
	cellPixelWidth  = 12
	cellPixelHeight = 24
	previewRatio    = 0.6
)

type imageRenderedMsg struct {
	rendered string
	imgRows  int
	err      error
}

type PreviewPanelModel struct {
	selected     *domain.Wallpaper
	renderedImg  string
	imgRows      int
	loading      bool
	kittySupport bool
	width        int
	height       int
}

func NewPreviewPanel(kittySupport bool) PreviewPanelModel {
	return PreviewPanelModel{kittySupport: kittySupport}
}

func (p PreviewPanelModel) Update(msg tea.Msg) (PreviewPanelModel, tea.Cmd) {
	switch msg := msg.(type) {
	case WallpaperSelectedMsg:
		p.selected = msg.Wallpaper
		p.renderedImg = ""
		p.imgRows = 0
		if msg.Wallpaper == nil || !p.kittySupport || p.width == 0 {
			p.loading = false
			return p, nil
		}
		p.loading = true
		innerCols := p.width - 2 // subtract border
		return p, renderImageCmd(msg.Wallpaper.Path, innerCols)
	case imageRenderedMsg:
		p.loading = false
		if msg.err == nil {
			p.renderedImg = msg.rendered
			p.imgRows = msg.imgRows
		}
	case tea.WindowSizeMsg:
		p.width = int(float64(msg.Width) * previewRatio)
		p.height = msg.Height
	}
	return p, nil
}

func (p PreviewPanelModel) View() string {
	var sb strings.Builder

	switch {
	case p.loading:
		sb.WriteString("loading…\n")
	case p.renderedImg != "":
		sb.WriteString(p.renderedImg)
		sb.WriteString(strings.Repeat("\n", p.imgRows))
	case !p.kittySupport:
		sb.WriteString("[no preview — kitty terminal required]\n")
	case p.selected == nil:
		sb.WriteString("[no selection]\n")
	default:
		sb.WriteString("[failed to load image]\n")
	}

	if p.selected != nil {
		sb.WriteString("\n")
		sb.WriteString(p.selected.FileName)
		sb.WriteString("\n")
		sb.WriteString(p.selected.Path)
		sb.WriteString("\n")
		var meta []string
		if p.selected.Resolution.Width > 0 {
			meta = append(meta, fmt.Sprintf("%d×%d", p.selected.Resolution.Width, p.selected.Resolution.Height))
		}
		if p.selected.FileSize > 0 {
			meta = append(meta, formatFileSize(p.selected.FileSize))
		}
		if len(meta) > 0 {
			sb.WriteString(strings.Join(meta, " · "))
			sb.WriteString("\n")
		}
	}

	return sb.String()
}

func formatFileSize(size domain.FileSize) string {
	const (
		kb = 1024
		mb = 1024 * kb
	)
	switch {
	case int64(size) >= mb:
		return fmt.Sprintf("%.1f MB", float64(size)/mb)
	case int64(size) >= kb:
		return fmt.Sprintf("%.1f KB", float64(size)/kb)
	default:
		return fmt.Sprintf("%d B", size)
	}
}

func renderImageCmd(path string, innerCols int) tea.Cmd {
	return func() tea.Msg {
		f, err := os.Open(path)
		if err != nil {
			return imageRenderedMsg{err: err}
		}
		defer f.Close()

		src, _, err := image.Decode(f)
		if err != nil {
			return imageRenderedMsg{err: err}
		}

		bounds := src.Bounds()
		targetW := innerCols * cellPixelWidth
		targetH := int(float64(targetW) * float64(bounds.Dy()) / float64(bounds.Dx()))
		if targetH < 1 {
			targetH = 1
		}

		dst := image.NewRGBA(image.Rect(0, 0, targetW, targetH))
		draw.BiLinear.Scale(dst, dst.Bounds(), src, bounds, draw.Over, nil)

		rows := targetH / cellPixelHeight
		if rows < 1 {
			rows = 1
		}

		var buf strings.Builder
		if err := kittyimg.Fprint(&buf, dst); err != nil {
			return imageRenderedMsg{err: err}
		}

		return imageRenderedMsg{rendered: buf.String(), imgRows: rows}
	}
}
