package gutter

import (
	"testing"

	"github.com/Runway-Club/gutter/themes"
)

func TestNewRunConfigDefaults(t *testing.T) {
	cfg := newRunConfig(nil)
	if cfg.selector != "#app" {
		t.Fatalf("default selector = %q, want %q", cfg.selector, "#app")
	}
	if cfg.theme != themes.Apple {
		t.Fatalf("default theme = %v, want themes.Apple", cfg.theme)
	}
}

func TestWithThemeAndSelector(t *testing.T) {
	cfg := newRunConfig([]Option{
		WithTheme(themes.Meta),
		WithSelector("#root"),
	})
	if cfg.theme != themes.Meta {
		t.Fatalf("theme = %v, want themes.Meta", cfg.theme)
	}
	if cfg.selector != "#root" {
		t.Fatalf("selector = %q, want %q", cfg.selector, "#root")
	}
}

func TestWithThemeNilIgnored(t *testing.T) {
	cfg := newRunConfig([]Option{WithTheme(nil)})
	if cfg.theme != themes.Apple {
		t.Fatalf("WithTheme(nil) should keep default Apple, got %v", cfg.theme)
	}
}

func TestWithSelectorEmptyIgnored(t *testing.T) {
	cfg := newRunConfig([]Option{WithSelector("")})
	if cfg.selector != "#app" {
		t.Fatalf("WithSelector(\"\") should keep default, got %q", cfg.selector)
	}
}
