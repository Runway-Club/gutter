package gutter

import (
	"reflect"

	"github.com/Runway-Club/gutter/themes"
)

// BuildContext is threaded through every Build call. It carries the active
// theme so themed widgets can read tokens (colors, typography, component
// styles) without the app having to pass the theme down by hand, plus the
// ambient dependency scope read by DependOn (see Provider).
type BuildContext struct {
	// Theme is the active theme. Set by RunApp from the WithTheme option, or
	// themes.Default if none was provided. Never nil during normal mounting.
	Theme *themes.Theme

	// inherited is the ambient dependency scope visible at the current position
	// in the tree, keyed by provided type. Providers push a superset of it for
	// their subtree; the runtime restores the right scope before an isolated
	// rebuild. Nil until a Provider is used, so apps that don't use DI pay
	// nothing. Read via DependOn, never directly.
	inherited map[reflect.Type]any
}
