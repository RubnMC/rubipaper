package scanner

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type ImageFile struct {
	Path    string
	Name    string
	ModTime time.Time
}

var imageExts = map[string]struct{}{
	".jpg": {}, ".jpeg": {}, ".png": {}, ".gif": {},
	".webp": {}, ".bmp": {}, ".tif": {}, ".tiff": {},
}

func ScanDirectory(path string) ([]ImageFile, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open directory: %w", err)
	}
	defer f.Close()

	entries, err := f.Readdir(-1)
	if err != nil {
		return nil, fmt.Errorf("read directory: %w", err)
	}

	var images []ImageFile
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		ext := strings.ToLower(filepath.Ext(entry.Name()))
		if _, ok := imageExts[ext]; !ok {
			continue
		}
		abs, err := filepath.Abs(filepath.Join(path, entry.Name()))
		if err != nil {
			return nil, fmt.Errorf("resolve path for %q: %w", entry.Name(), err)
		}
		images = append(images, ImageFile{
			Path:    abs,
			Name:    entry.Name(),
			ModTime: entry.ModTime(),
		})
	}
	return images, nil
}

// Order of results is not guaranteed.
func ScanDirectoryRecursive(path string) ([]ImageFile, error) {
	var images []ImageFile
	err := filepath.WalkDir(path, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			found, err := ScanDirectory(p)
			if err != nil {
				return err
			}
			images = append(images, found...)
		}
		return nil
	})
	return images, err
}
