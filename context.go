package gutter

import "github.com/Runway-Club/gutter/themes"

// BuildContext is threaded through every Build call. It carries the active
// theme so themed widgets can read tokens (colors, typography, component
// styles) without the app having to pass the theme down by hand. Future
// versions may expose additional inherited data (locale, router state).
type BuildContext struct {
	// Theme is the active theme. Set by RunApp from the WithTheme option, or
	// themes.Default if none was provided. Never nil during normal mounting.
	Theme *themes.Theme
}
