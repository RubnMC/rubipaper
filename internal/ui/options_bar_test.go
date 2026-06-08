package ui

import (
	"strings"
	"testing"

	"github.com/RubnMC/rubipaper/internal/domain"
)

// stubBackend is shared across all ui tests in this package.
type stubBackend struct{}

func (s stubBackend) Name() string                            { return "stub" }
func (s stubBackend) IsAvaliable() bool                      { return true }
func (s stubBackend) SupportedModes() []domain.WallpaperMode { return nil }
func (s stubBackend) SetWallpaper(_, _ string) error         { return nil }
func (s stubBackend) GetCurrentWallpaper() (string, error)   { return "", nil }

func TestOptionsBarStartsUnfocused(t *testing.T) {
	o := NewOptionsBar("/home/user/pics", stubBackend{}, domain.ModeFill)
	if o.IsFocused() {
		t.Error("new OptionsBar should not be focused")
	}
}

func TestOptionsBarFocusReturnsFocusedCopy(t *testing.T) {
	o := NewOptionsBar("/home/user/pics", stubBackend{}, domain.ModeFill)
	focused := o.Focus()
	if !focused.IsFocused() {
		t.Error("Focus() should return a focused component")
	}
	if o.IsFocused() {
		t.Error("Focus() must not mutate the original")
	}
}

func TestOptionsBarBlurReturnsUnfocusedCopy(t *testing.T) {
	o := NewOptionsBar("/home/user/pics", stubBackend{}, domain.ModeFill)
	blurred := o.Focus().Blur()
	if blurred.IsFocused() {
		t.Error("Blur() should return an unfocused component")
	}
}

func TestOptionsBarRecursiveToggledMsg(t *testing.T) {
	o := NewOptionsBar("/home/user/pics", stubBackend{}, domain.ModeFill)
	updated, _ := o.Update(RecursiveToggledMsg{Recursive: true})
	ob := updated.(OptionsBarModel)
	if !ob.recursive {
		t.Error("OptionsBar should set recursive=true on RecursiveToggledMsg{Recursive:true}")
	}
}

func TestOptionsBarViewContainsFields(t *testing.T) {
	o := NewOptionsBar("/home/user/pics", stubBackend{}, domain.ModeFill)
	view := o.View()
	for _, want := range []string{"stub", "fill", "false"} {
		if !strings.Contains(view, want) {
			t.Errorf("View() missing %q", want)
		}
	}
}
