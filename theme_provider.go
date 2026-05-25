package gutter

import (
	"reflect"

	"github.com/Runway-Club/gutter/themes"
)

// ThemeProvider overrides the active theme for its subtree. Unlike the single
// BuildContext.Theme field (set app-wide by RunApp's WithTheme or by Scaffold),
// ThemeProvider scopes a theme to Child only — nest it to theme a section
// (a dark card, a branded panel) differently from the rest of the app:
//
//	gutter.ThemeProvider{Theme: themes.Meta, Child: panel}
//
// It rides the same inherited-scope machinery as Provider, so it is correct
// under isolated SetState rebuilds (each element restores the scope it lives
// under) and on the SSR path. Themed widgets read it via the package-private
// activeTheme, which prefers a ThemeProvider over BuildContext.Theme.
type ThemeProvider struct {
	Theme *themes.Theme
	Child Widget
}

// Build renders the child; the theme is injected into the scope by the runtime.
func (p ThemeProvider) Build(*BuildContext) Widget { return p.Child }

func (p ThemeProvider) provideInto(parent map[reflect.Type]any) map[reflect.Type]any {
	m := make(map[reflect.Type]any, len(parent)+1)
	for k, v := range parent {
		m[k] = v
	}
	m[reflect.TypeFor[*themes.Theme]()] = p.Theme
	return m
}
