package widgets

import "github.com/Runway-Club/gutter"

// WithKey wraps any widget with a reconciliation key. It is purely a wrapper:
// the rendered tree is whatever Child produces, but the framework treats this
// node as the same instance across rebuilds iff Key compares equal.
//
// Prefer implementing gutter.Keyed directly on your own widgets when possible;
// WithKey exists for ad-hoc keying of widgets you do not own (e.g. wrapping a
// built-in widget inside a Column).
type WithKey struct {
	Key   any
	Child gutter.Widget
}

func (w WithKey) Build(ctx *gutter.BuildContext) gutter.Widget { return w.Child }
func (w WithKey) WidgetKey() any                               { return w.Key }
