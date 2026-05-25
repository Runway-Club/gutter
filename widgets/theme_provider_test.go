package widgets

import (
	"strings"
	"testing"

	"github.com/Runway-Club/gutter"
	"github.com/Runway-Club/gutter/themes"
)

// RenderToHTML pushes inherited scopes, so a ThemeProvider over a themed widget
// must make activeTheme resolve to the provided theme — end to end.
func TestThemeProviderOverridesTheme(t *testing.T) {
	metaBg := themes.Meta.Components.ButtonPrimary.Background
	appleBg := themes.Apple.Components.ButtonPrimary.Background
	if metaBg == appleBg {
		t.Skip("Meta and Apple primary backgrounds coincide; test can't distinguish")
	}

	out, err := gutter.RenderToHTML(gutter.ThemeProvider{
		Theme: themes.Meta,
		Child: Button{Variant: ButtonPrimary, Label: "x"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, metaBg) {
		t.Errorf("button did not use the provided Meta theme background %q:\n%s", metaBg, out)
	}
	if strings.Contains(out, appleBg) {
		t.Errorf("button used the Apple background %q despite the ThemeProvider", appleBg)
	}
}

// A nested ThemeProvider shadows an outer one for its subtree only.
func TestThemeProviderNestedShadowing(t *testing.T) {
	out, err := gutter.RenderToHTML(gutter.ThemeProvider{
		Theme: themes.Apple,
		Child: Row{Children: []gutter.Widget{
			Button{Variant: ButtonPrimary, Label: "outer"},
			gutter.ThemeProvider{Theme: themes.Meta, Child: Button{Variant: ButtonPrimary, Label: "inner"}},
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, themes.Apple.Components.ButtonPrimary.Background) {
		t.Error("outer button should use Apple")
	}
	if !strings.Contains(out, themes.Meta.Components.ButtonPrimary.Background) {
		t.Error("inner button should use Meta")
	}
}
