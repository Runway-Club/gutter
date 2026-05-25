package themes

import (
	"strings"
	"testing"
)

func allPresets() map[string]*Theme {
	return map[string]*Theme{
		"Apple":   Apple,
		"Meta":    Meta,
		"Neutral": Neutral,
	}
}

func TestPresetsAreComplete(t *testing.T) {
	for name, th := range allPresets() {
		t.Run(name, func(t *testing.T) {
			if th == nil {
				t.Fatal("preset is nil")
			}
			if th.Name == "" {
				t.Error("Name is empty")
			}
			// Core semantic colors every themed widget relies on.
			for field, v := range map[string]string{
				"Colors.Primary":   th.Colors.Primary,
				"Colors.OnPrimary": th.Colors.OnPrimary,
				"Colors.Canvas":    th.Colors.Canvas,
				"Colors.Ink":       th.Colors.Ink,
			} {
				if strings.TrimSpace(v) == "" {
					t.Errorf("%s is empty", field)
				}
			}
			// Body typography must be set, and per the design system every
			// built-in theme leads its font stack with Lexend.
			if th.Typography.Body.FontFamily == "" {
				t.Error("Typography.Body.FontFamily is empty")
			} else if !strings.Contains(th.Typography.Body.FontFamily, "Lexend") {
				t.Errorf("Typography.Body.FontFamily = %q, expected it to lead with Lexend", th.Typography.Body.FontFamily)
			}
			if th.Typography.Body.FontSize == "" {
				t.Error("Typography.Body.FontSize is empty")
			}
		})
	}
}

func TestAppleIsFrameworkDefaultColorShape(t *testing.T) {
	// Sanity: Apple and Meta are distinct presets (guards against a
	// copy-paste regression where one aliases the other).
	if Apple == Meta {
		t.Fatal("Apple and Meta point at the same Theme value")
	}
	if Apple.Name == Meta.Name {
		t.Fatalf("Apple.Name == Meta.Name == %q", Apple.Name)
	}
}
