package domain

type WallpaperMode string
type FileSize int64
type Resolution struct {
	Width  int
	Height int
}

const (
	ModeFill    WallpaperMode = "fill"
	ModeCenter  WallpaperMode = "center"
	ModeTile    WallpaperMode = "tile"
	ModeStretch WallpaperMode = "stretch"
	ModeFit     WallpaperMode = "fit"
)

var ValidModes = []WallpaperMode{ModeFill, ModeCenter, ModeTile, ModeStretch, ModeFit}

type Wallpaper struct {
	Path       string
	FileName   string
	Resolution Resolution
	FileSize   FileSize
}
