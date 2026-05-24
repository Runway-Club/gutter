//go:build !js || !wasm

package widgets

// attachDragSource is a no-op on host builds — pointer events and DOM
// hit-testing only exist under GOOS=js GOARCH=wasm. The Draggable widget
// still renders its child correctly off WASM (just without the gesture).
func attachDragSource[T any](node any, s *draggableState[T]) func() {
	return func() {}
}
