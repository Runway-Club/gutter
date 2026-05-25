package widgets

import (
	"testing"

	"github.com/Runway-Club/gutter/themes"
)

func TestResolveColorTokens(t *testing.T) {
	th := themes.Apple
	cases := map[string]string{
		ColorPrimary:      th.Colors.Primary,
		ColorOnPrimary:    th.Colors.OnPrimary,
		ColorAccent:       th.Colors.Accent,
		ColorCanvas:       th.Colors.Canvas,
		ColorCanvasAlt:    th.Colors.CanvasAlt,
		ColorSurfaceSoft:  th.Colors.SurfaceSoft,
		ColorSurfaceDark:  th.Colors.SurfaceDark,
		ColorOnDark:       th.Colors.OnDark,
		ColorInk:          th.Colors.Ink,
		ColorInkMuted:     th.Colors.InkMuted,
		ColorInkSubtle:    th.Colors.InkSubtle,
		ColorHairline:     th.Colors.Hairline,
		ColorHairlineSoft: th.Colors.HairlineSoft,
		ColorSuccess:      th.Colors.Success,
		ColorWarning:      th.Colors.Warning,
		ColorCritical:     th.Colors.Critical,
	}
	for token, want := range cases {
		if got := resolveColor(th, token); got != want {
			t.Errorf("resolveColor(%q) = %q, want %q", token, got, want)
		}
	}
}

func TestResolveColorPassThrough(t *testing.T) {
	th := themes.Apple
	for _, raw := range []string{"#fff", "rgb(1,2,3)", "tomato", "var(--x)", ""} {
		if got := resolveColor(th, raw); got != raw {
			t.Errorf("resolveColor(%q) = %q, want it unchanged", raw, got)
		}
	}
}

func TestResolveColorUnknownToken(t *testing.T) {
	if got := resolveColor(themes.Apple, "theme:does-not-exist"); got != "" {
		t.Errorf("unknown token resolved to %q, want \"\"", got)
	}
}

func TestContainerResolvesColorToken(t *testing.T) {
	h := hostOfCtx(t, Container{Color: ColorPrimary}, testCtx(themes.Apple))
	wantStyle(t, h, "background-color", themes.Apple.Colors.Primary)
}

func TestContainerRawColorUnchanged(t *testing.T) {
	h := hostOfCtx(t, Container{Color: "#abcdef"}, testCtx(themes.Apple))
	wantStyle(t, h, "background-color", "#abcdef")
}

func TestContainerBorderColorComposesBorder(t *testing.T) {
	h := hostOfCtx(t, Container{BorderColor: ColorHairline}, testCtx(themes.Apple))
	want := "1px solid " + themes.Apple.Colors.Hairline
	wantStyle(t, h, "border", want)
}

func TestContainerBorderShorthandWins(t *testing.T) {
	h := hostOfCtx(t, Container{Border: "2px dashed red", BorderColor: ColorHairline}, testCtx(themes.Apple))
	wantStyle(t, h, "border", "2px dashed red")
}

func TestTokenResolvesPerTheme(t *testing.T) {
	// The same token must resolve to each theme's own palette value.
	apple := hostOfCtx(t, Container{Color: ColorPrimary}, testCtx(themes.Apple))
	meta := hostOfCtx(t, Container{Color: ColorPrimary}, testCtx(themes.Meta))
	if apple.Style["background-color"] != themes.Apple.Colors.Primary {
		t.Error("Apple primary mismatch")
	}
	if meta.Style["background-color"] != themes.Meta.Colors.Primary {
		t.Error("Meta primary mismatch")
	}
	if themes.Apple.Colors.Primary != themes.Meta.Colors.Primary &&
		apple.Style["background-color"] == meta.Style["background-color"] {
		t.Error("token did not re-resolve when the theme changed")
	}
}

func TestHeadingResolvesColorToken(t *testing.T) {
	h := hostOfCtx(t, Heading{Level: H1, Text: "x", Color: ColorPrimary}, testCtx(themes.Apple))
	wantStyle(t, h, "color", themes.Apple.Colors.Primary)
}

func TestBodyResolvesColorToken(t *testing.T) {
	h := hostOfCtx(t, Body{Text: "x", Color: ColorOnDark}, testCtx(themes.Apple))
	wantStyle(t, h, "color", themes.Apple.Colors.OnDark)
}

func TestLinkResolvesColorToken(t *testing.T) {
	h := hostOfCtx(t, Link{Text: "x", Color: ColorAccent}, testCtx(themes.Apple))
	wantStyle(t, h, "color", themes.Apple.Colors.Accent)
}

func TestHeadingDefaultColorIsInk(t *testing.T) {
	h := hostOfCtx(t, Heading{Level: H2, Text: "x"}, testCtx(themes.Apple))
	wantStyle(t, h, "color", themes.Apple.Colors.Ink)
}
