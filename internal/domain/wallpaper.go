package domain

type WallpaperMode string

const (
	ModeFill    WallpaperMode = "fill"
	ModeCenter  WallpaperMode = "center"
	ModeTile    WallpaperMode = "tile"
	ModeStretch WallpaperMode = "stretch"
	ModeFit     WallpaperMode = "fit"
)
