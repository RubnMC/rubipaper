package ui

import (
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"math"
	"os"
	"strings"

	"github.com/RubnMC/rubipaper/internal/domain"
	"github.com/dolmen-go/kittyimg"
	tea "github.com/charmbracelet/bubbletea"
	"golang.org/x/image/draw"
)

// cellPixelWidth is the assumed terminal cell width in pixels (typically 12px at default
// font sizes). Used to compute the target image pixel width from the available column count.
// cellPixelHeight is the assumed terminal cell height in pixels. Used to estimate how many
// rows the rendered Kitty image occupies so that Bubbletea's text layout leaves the correct
// vertical gap. 24px is the most common default but varies with font size and terminal config.
const (
	cellPixelWidth = 12
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
	reservedRows int // fixed row count for image area, locked in when selection starts loading
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
		p.reservedRows = 0
		if msg.Wallpaper == nil || !p.kittySupport || p.width == 0 {
			p.loading = false
			return p, nil
		}
		maxRows := max(p.height / 3, 1)
		p.reservedRows = maxRows
		p.loading = true
		innerCols := p.width - 2
		return p, renderImageCmd(msg.Wallpaper.Path, innerCols, maxRows)
	case imageRenderedMsg:
		p.loading = false
		if msg.err == nil {
			p.renderedImg = msg.rendered
			p.imgRows = msg.imgRows
		}
	case tea.WindowSizeMsg:
		p.width = int(float64(msg.Width) * previewRatio)
		p.height = msg.Height
		if p.selected != nil && p.kittySupport && p.renderedImg == "" && !p.loading {
			maxRows := max(p.height / 3, 1)
			p.reservedRows = maxRows
			p.loading = true
			innerCols := p.width - 2
			return p, renderImageCmd(p.selected.Path, innerCols, maxRows)
		}
	}
	return p, nil
}

func (p PreviewPanelModel) View() string {
	var sb strings.Builder

	switch {
	case p.loading:
		// Pad to reservedRows so height is stable — no jump when image arrives.
		sb.WriteString("loading…\n")
		if p.reservedRows > 1 {
			sb.WriteString(strings.Repeat("\n", p.reservedRows-1))
		}
	case p.renderedImg != "":
		sb.WriteString(p.renderedImg)
		// imgRows <= reservedRows (letterbox guarantee); pad the remainder.
		if pad := p.reservedRows - p.imgRows; pad > 0 {
			sb.WriteString(strings.Repeat("\n", pad))
		}
	case !p.kittySupport:
		sb.WriteString("[no preview — kitty terminal required]\n")
	case p.selected == nil:
		sb.WriteString("[no selection]\n")
	default:
		sb.WriteString("[failed to load image]\n")
		if p.reservedRows > 1 {
			sb.WriteString(strings.Repeat("\n", p.reservedRows-1))
		}
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

func renderImageCmd(path string, innerCols, maxRows int) tea.Cmd {
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

		// Letterbox: scale to fit within innerCols x maxRows cells, preserving aspect ratio.
		bounds := src.Bounds()
		maxPixelW := float64(innerCols * cellPixelWidth)
		maxPixelH := float64(maxRows * cellPixelHeight)
		scale := math.Min(maxPixelW/float64(bounds.Dx()), maxPixelH/float64(bounds.Dy()))
		targetW := int(float64(bounds.Dx()) * scale)
		targetH := int(float64(bounds.Dy()) * scale)
		if targetW < 1 {
			targetW = 1
		}
		if targetH < 1 {
			targetH = 1
		}

		dst := image.NewRGBA(image.Rect(0, 0, targetW, targetH))
		draw.BiLinear.Scale(dst, dst.Bounds(), src, bounds, draw.Over, nil)

		rows := max(targetH / cellPixelHeight, 1)

		var buf strings.Builder

		// Clear all previous Kitty image placements before writing the new one.
		// Without this, old image pixels persist in the terminal's graphics layer
		// even after Bubbletea overwrites the text content.
		buf.WriteString("\x1b_Ga=d,d=A\x1b\\")

		// Save cursor before the APC so we can restore it after. The Kitty APC
		// advances the terminal cursor by imgRows rows, but Bubbletea measures
		// frame height by counting \n characters and never sees that movement.
		// Restoring the cursor then emitting `rows` real newlines keeps both
		// Bubbletea's line count and the terminal cursor in sync.
		buf.WriteString("\x1b[s")

		if err := kittyimg.Fprint(&buf, dst); err != nil {
			return imageRenderedMsg{err: err}
		}
		buf.WriteString("\x1b[u")
		buf.WriteString(strings.Repeat("\n", rows))

		return imageRenderedMsg{rendered: buf.String(), imgRows: rows}
	}
}
