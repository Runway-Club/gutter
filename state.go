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

// StateObject is the SetState mixin. Embed it (by value) in your concrete
// State struct, then return a pointer to that struct from CreateState. The
// framework binds the element automatically.
type StateObject struct {
	elem stateElement
}

// SetState mutates state and asks the framework to rebuild the subtree owned
// by this State. If the state has not yet been mounted, the call is a no-op.
func (s *StateObject) SetState(fn func()) {
	fn()
	if s.elem != nil {
		s.elem.rebuild()
	}
}

func (s *StateObject) bindElement(el stateElement) {
	s.elem = el
}
