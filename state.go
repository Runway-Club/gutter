package gutter

// State is the per-instance data of a StatefulWidget. Its Build is called on
// every rebuild of that subtree.
type State interface {
	Build(ctx *BuildContext) Widget
}

// StateInitializer is implemented by states that need a one-shot hook after
// being created and before their first build.
type StateInitializer interface {
	InitState()
}

// StateDisposer is implemented by states that need to clean up when their
// element is unmounted (close channels, cancel timers, etc.).
type StateDisposer interface {
	Dispose()
}

// WidgetUpdater is implemented by states that need to react when an ancestor
// rebuild produces a new instance of their StatefulWidget. The framework first
// swaps in the new widget pointer (so Widget() returns the new one), then
// invokes DidUpdateWidget with the previous instance so the State can diff
// fields and resubscribe/restart as needed. It is called BEFORE the rebuild,
// so calling SetState here is unnecessary — the rebuild happens unconditionally
// after the hook returns.
type WidgetUpdater interface {
	DidUpdateWidget(oldWidget Widget)
}

// stateElement is the slice of statefulElement that the State sees: just
// enough to request a rebuild of its own subtree. The concrete implementation
// lives in element_wasm.go.
type stateElement interface {
	rebuild()
}

// elementBinder is satisfied by StateObject (via embedding). The framework
// uses it to inject the per-instance element handle after CreateState.
type elementBinder interface {
	bindElement(stateElement)
}

// widgetBinder is satisfied by StateObject (via embedding). The framework
// uses it to inject the current widget instance on mount and on every update.
type widgetBinder interface {
	bindWidget(Widget)
}

// StateObject is the SetState mixin. Embed it (by value) in your concrete
// State struct, then return a pointer to that struct from CreateState. The
// framework binds the element and widget pointer automatically.
type StateObject struct {
	elem   stateElement
	widget Widget
}

// SetState mutates state and asks the framework to rebuild the subtree owned
// by this State. If the state has not yet been mounted, the call is a no-op.
func (s *StateObject) SetState(fn func()) {
	fn()
	if s.elem != nil {
		s.elem.rebuild()
	}
}

// Widget returns the current StatefulWidget instance that owns this State.
// The framework keeps this fresh: on mount it is set before InitState; on
// update it is replaced before DidUpdateWidget fires.
func (s *StateObject) Widget() Widget {
	return s.widget
}

func (s *StateObject) bindElement(el stateElement) {
	s.elem = el
}

func (s *StateObject) bindWidget(w Widget) {
	s.widget = w
}
