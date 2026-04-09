package domain

import "fmt"

// Backend defines the interface for setting wallpapers.
type Backend interface {
	IsAvaliable() bool
	SetWallpaper(imagePath string, mode string) error
	GetCurrentWallpaper() (string, error)
	SupportedModes() []WallpaperMode
	Name() string
}

// DetectBackend detects and returns the appropriate wallpaper backend.
func DetectBackend() (Backend, error) {
	return nil, fmt.Errorf("no backend detected")
}
