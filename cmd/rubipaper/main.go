package main

import (
	"fmt"
	"os"

	"github.com/RubnMC/rubipaper/internal/scanner"
)

const defaultWallpaperDir = "/home/ruben/Pictures/Wallpapers"

func main() {
	images, err := scanner.ScanDirectory(defaultWallpaperDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "rubipaper: failed to scan %q: %v\n", defaultWallpaperDir, err)
		os.Exit(1)
	}
	if len(images) == 0 {
		fmt.Printf("rubipaper: no images found in %s\n", defaultWallpaperDir)
		return
	}
	for _, img := range images {
		fmt.Println(img.Path)
	}
}
