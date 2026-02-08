package backend

import "fmt"

// Backend defines the interface for setting wallpapers.
type Backend interface {
	SetWallpaper(imagePath string, mode string) error
	Name() string
}

// DetectBackend detects and returns the appropriate wallpaper backend.
func DetectBackend() (Backend, error) {
	return nil, fmt.Errorf("no backend detected")
}
