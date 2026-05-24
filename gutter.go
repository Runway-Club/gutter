// Package gutter is a Flutter-inspired declarative UI library for Go. It
// targets WebAssembly and drives the browser DOM directly: widgets describe
// what should appear on screen, and the runtime takes care of materializing
// them into DOM elements.
//
// A Widget is any value that implements one of HostWidget, StatelessWidget,
// or StatefulWidget. HostWidgets are leaves that map to a single DOM element;
// the other two compose by producing a child widget at build time.
package gutter

// Widget is anything the framework knows how to mount. The concrete type must
// implement HostWidget, StatelessWidget, or StatefulWidget — the reconciler
// dispatches via a type switch.
type Widget = any

// HostWidget describes a DOM element directly. Use this for leaves like Text,
// Button, or layout primitives that map 1:1 to an HTML tag.
type HostWidget interface {
	Host() *Host
}

// StatelessWidget builds its UI by returning a child widget. The child is
// rebuilt from scratch whenever an ancestor rebuilds.
type StatelessWidget interface {
	Build(ctx *BuildContext) Widget
}

// StatefulWidget owns a State whose Build is invoked on every rebuild. Use it
// when the widget needs to remember anything between builds.
type StatefulWidget interface {
	CreateState() State
}

// Host is the data the framework needs to materialize a DOM element. Tag and
// Text are mutually exclusive in practice: a non-empty Text fills the element
// via textContent, while Children is recursively mounted by the framework.
//
// OnMount and OnUnmount are escape hatches for widgets that need to touch
// the live DOM node directly (e.g. Canvas calling getContext("2d"), or a
// widget wiring up a non-DOM API like a Web Worker tied to a placeholder
// element). The node argument is the platform's native handle — on WASM it
// is a syscall/js.Value; the parameter is typed as any so widget structs
// stay platform-neutral. OnMount fires after the element is inserted in the
// DOM; OnUnmount fires before it is removed.
type Host struct {
	Tag       string
	Text      string
	Attrs     map[string]string
	Style     map[string]string
	Events    map[string]func(Event)
	Children  []Widget
	OnMount   func(node any)
	OnUnmount func(node any)
}

// Event is what a registered handler receives. Value is populated from
// event.target.value for input-style events. X/Y are the viewport-relative
// pointer coordinates (clientX/clientY) for mouse/pointer events;
// OffsetX/OffsetY are the same coordinates expressed in the target
// element's local space (offsetX/offsetY) — handy for Canvas drawing.
// Key is the key identifier for keyboard events. Fields irrelevant to
// the source event are zero.
type Event struct {
	Type             string
	Value            string
	X, Y             float64
	OffsetX, OffsetY float64
	Key              string
}

// Keyed is implemented by widgets that participate in keyed reconciliation.
// During reconcileChildren, two widgets are considered the same instance only
// if their Go types match AND their WidgetKey values are equal. Widgets that
// do not implement Keyed fall back to positional matching among the unkeyed
// siblings of the same type. The widgets package ships a generic WithKey
// wrapper for ad-hoc keying.
type Keyed interface {
	WidgetKey() any
}
