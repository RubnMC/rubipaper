package scanner_test

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"slices"
	"testing"
	"time"

	"github.com/RubnMC/rubipaper/internal/scanner"
)

// makeDir creates a temp directory and populates it with the given filenames.
// It registers cleanup automatically via t.Cleanup.
func makeDir(t *testing.T, files []string) string {
	t.Helper()
	dir, err := os.MkdirTemp("", "rubipaper-test-*")
	if err != nil {
		t.Fatalf("create temp dir: %v", err)
	}
	t.Cleanup(func() { os.RemoveAll(dir) })

	for _, name := range files {
		path := filepath.Join(dir, name)
		if err := os.WriteFile(path, []byte{}, 0o644); err != nil {
			t.Fatalf("create file %q: %v", name, err)
		}
	}
	return dir
}

func TestScanDirectory_ReturnsImages(t *testing.T) {
	dir := makeDir(t, []string{"a.jpg", "b.PNG", "c.txt", "d.webp"})
	expected := []string{"a.jpg", "b.PNG", "d.webp"}

	got, err := scanner.ScanDirectory(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(got) != len(expected) {
		t.Fatalf("expected %d images, got %d", len(got), len(expected))
	}

	for i := range got {
		if !slices.Contains(expected, got[i].FileName) {
			t.Fatalf("differences found between the expected result set %v and the actual result %v", expected, got)
		}
	}
}

func TestScanDirectory_EmptyDir(t *testing.T) {
	dir := makeDir(t, []string{})

	got, err := scanner.ScanDirectory(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(got) != 0 {
		t.Fatalf("directory %v should be empty, but got %d image files", dir, len(got))
	}
}

func TestScanDirectory_NonExistentPath(t *testing.T) {
	got, err := scanner.ScanDirectory("/test_" + time.Now().Format("20060102") + "_b2c1-0d5e6f7a8b9c")
	_ = got
	if err == nil {
		t.Fatalf("expected an error after scanning non existent path but got %v", err)
	}
}

func TestScanDirectory_PopulatesMetadata(t *testing.T) {
	dir, err := os.MkdirTemp("", "rubipaper-test-*")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.RemoveAll(dir) })

	img := image.NewRGBA(image.Rect(0, 0, 100, 200))
	img.Set(0, 0, color.White)
	f, err := os.Create(filepath.Join(dir, "test.png"))
	if err != nil {
		t.Fatal(err)
	}
	if err := png.Encode(f, img); err != nil {
		t.Fatal(err)
	}
	f.Close()

	got, err := scanner.ScanDirectory(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 image, got %d", len(got))
	}
	w := got[0]
	if w.Resolution.Width != 100 || w.Resolution.Height != 200 {
		t.Errorf("Resolution = %dx%d, want 100x200", w.Resolution.Width, w.Resolution.Height)
	}
	if w.FileSize <= 0 {
		t.Errorf("FileSize = %d, want > 0", w.FileSize)
	}
}
