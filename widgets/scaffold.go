package widgets

import (
	"github.com/Runway-Club/gutter"
	"github.com/Runway-Club/gutter/themes"
)

// Scaffold is the app shell. It's typically the root widget your app's
// Build returns, and it ties together the four big pieces of a real app:
//
//   - Title       — pushed to document.title
//   - Theme       — switches the active theme for this subtree (and, since
//                   gutter has one BuildContext per app, effectively the
//                   whole app)
//   - AppBar      — the top navigation strip (use widgets.AppBar)
//   - Body        — your main content; takes the remaining vertical space
//   - Footer      — an optional bottom strip (legal, build info, etc.)
//
// Background and ink come from the active theme's canvas/ink. Body sits in
// a flex column between AppBar and Footer, so a Center inside Body fills
// the viewport minus the chrome — exactly what you want for landing pages,
// dialogs, and most app screens.
type Scaffold struct {
	Title  string
	Theme  *themes.Theme
	AppBar gutter.Widget
	Body   gutter.Widget
	Footer gutter.Widget
}

func (s Scaffold) Build(ctx *gutter.BuildContext) gutter.Widget {
	// Theme on Scaffold wins over the framework default / WithTheme. We
	// mutate ctx because the same BuildContext is threaded through every
	// descendant; this is the simplest "theme provider" without an
	// InheritedWidget mechanism.
	if s.Theme != nil {
		ctx.Theme = s.Theme
	}
	if s.Title != "" {
		gutter.SetTitle(s.Title)
	}
	t := activeTheme(ctx)

	var children []gutter.Widget
	if s.AppBar != nil {
		children = append(children, s.AppBar)
	}
	if s.Body != nil {
		// Body takes the remaining space (flex: 1) and is itself a flex
		// column so a Center inside it can use height: 100% reliably.
		children = append(children, Styled{
			Style: map[string]string{
				"flex":           "1",
				"display":        "flex",
				"flex-direction": "column",
				"min-height":     "0",
			},
			Children: []gutter.Widget{s.Body},
		})
	}
	if s.Footer != nil {
		children = append(children, s.Footer)
	}

	return Styled{
		Style: map[string]string{
			"background-color": t.Colors.Canvas,
			"color":            t.Colors.Ink,
			"display":          "flex",
			"flex-direction":   "column",
			"min-height":       "100%",
			"height":           "100%",
			"width":            "100%",
			"box-sizing":       "border-box",
		},
		Children: children,
	}
}
