package backend

import (
	"fmt"
	"github.com/RubnMC/rubipaper/internal/domain"
	"os"
	"os/exec"
)

type SwaybgBackend struct {
	currentWallpaper string
	currentMode      string
	socketPath       string
}

func NewSwaybgBackend() *SwaybgBackend {
	return &SwaybgBackend{
		socketPath: os.Getenv("SWAYSOCK"),
	}
}

// --- domain.Backend interface ---

func (s *SwaybgBackend) Name() string {
	return "swaybg"
}

func (s *SwaybgBackend) IsAvaliable() bool {
	_, errBg := exec.LookPath("swaybg")
	return errBg == nil
}

func (s *SwaybgBackend) SupportedModes() []domain.WallpaperMode {
	return []domain.WallpaperMode{
		domain.ModeStretch,
		domain.ModeFill,
		domain.ModeFit,
		domain.ModeCenter,
		domain.ModeTile,
	}
}

func (s *SwaybgBackend) SetWallpaper(imagePath string, mode string) error {
	if s.isSwayEnv() {
		cmd := exec.Command("swaymsg", "output", "*", "bg", imagePath, mode)
		if out, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("swaymsg failed: %w: %s", err, out)
		}
	} else {
		// Standalone swaybg (why would you do this?): kill existing instances then relaunch.
		pkill := exec.Command("pkill", "-x", "swaybg")
		_ = pkill.Run()

		cmd := exec.Command("swaybg", "-o", "*", "-i", imagePath, "-m", mode)
		if err := cmd.Start(); err != nil {
			return fmt.Errorf("swaybg failed to start: %w", err)
		}
		cmd.Process.Release()
	}
	s.currentWallpaper = imagePath
	s.currentMode = mode
	return nil
}

func (s *SwaybgBackend) GetCurrentWallpaper() (string, error) {
	if s.currentWallpaper == "" {
		return "", fmt.Errorf("no wallpaper set in this session")
	}
	return s.currentWallpaper, nil
}

// --- internal helpers ---

func (s *SwaybgBackend) isSwayEnv() bool {
	return s.socketPath != ""
}
