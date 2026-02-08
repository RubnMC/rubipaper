package scanner

import "time"

// ImageFile represents an image file found during scanning.
type ImageFile struct {
	Path    string
	Name    string
	ModTime time.Time
}

// ScanDirectory scans the given path for image files.
func ScanDirectory(path string) ([]ImageFile, error) {
	return nil, nil
}
