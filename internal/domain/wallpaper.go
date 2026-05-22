package domain

type WallpaperMode string
type FileSize int64
type Resolution struct {
	width  int
	height int
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
	fileName   string
	mode       WallpaperMode
	resolution Resolution
	fileSize   FileSize
}
