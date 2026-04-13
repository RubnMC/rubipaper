package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/RubnMC/rubipaper/internal/backend"
	"github.com/RubnMC/rubipaper/internal/scanner"
)

const defaultWallpaperDir = "/home/ruben/Pictures/Wallpapers"

func main() {
	b := backend.NewSwaybgBackend()
	_ = b
	images, err := scanner.ScanDirectory(defaultWallpaperDir)
	if err != nil {
		slog.Error("failed to scan directory", "dir", defaultWallpaperDir, "err", err)
		os.Exit(1)
	}
	if len(images) == 0 {
		slog.Warn("no images found", "dir", defaultWallpaperDir)
		return
	}
	for _, img := range images {
		fmt.Println(img.Path)
	}
	// if err := b.SetWallpaper(images[0].Path, "fill"); err != nil {
	// 	slog.Error("failed to set wallpaper", "err", err)
	// 	os.Exit(1)
	// }
}
