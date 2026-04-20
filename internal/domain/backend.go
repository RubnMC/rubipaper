package domain

// Backend defines the interface for setting wallpapers.
type Backend interface {
	IsAvaliable() bool
	SetWallpaper(imagePath string, mode string) error
	GetCurrentWallpaper() (string, error)
	SupportedModes() []WallpaperMode
	Name() string
}
