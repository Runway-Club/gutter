//go:build js && wasm

package widgets

import "syscall/js"

// attachDragSource wires the source DOM node's pointerdown to begin a drag
// for the Draggable owning state s. Returns an idempotent cleanup that
// detaches the pointerdown listener (and, while a drag is in-flight, the
// window-level move/up listeners).
//
// The flow:
//
//  1. pointerdown → preventDefault, capture pointer to the source, record
//     pointer offset within the card so the ghost stays anchored where the
//     user grabbed.
//  2. controller.startDrag(payload, ghost, x, y, offsetX, offsetY).
//  3. Install window-level pointermove listener that hit-tests all
//     registered DropTargets and feeds controller.updateDrag.
//  4. Install window-level pointerup listener that fires controller.endDrag
//     and detaches itself.
//
// Hit-test policy: pick the **smallest** target rect containing the
// pointer. That way nested targets behave naturally — the inner target
// wins, the outer one only catches the pointer outside the inner.
func attachDragSource[T any](node any, s *draggableState[T]) func() {
	n, ok := node.(js.Value)
	if !ok {
		return func() {}
	}

	released := false
	var moveCB, upCB js.Func
	pointerActive := false

	finishDrag := func(viaPointerUp bool) {
		if !pointerActive {
			return
		}
		pointerActive = false
		js.Global().Call("removeEventListener", "pointermove", moveCB)
		js.Global().Call("removeEventListener", "pointerup", upCB)
		moveCB.Release()
		upCB.Release()
		w := s.widget()
		var dropped bool
		if viaPointerUp {
			_, dropped = w.Controller.endDrag()
		} else {
			w.Controller.cancelDrag()
		}
		s.SetState(func() { s.active = false })
		if w.OnDragEnd != nil {
			w.OnDragEnd(dropped)
		}
	}

	pointerDown := js.FuncOf(func(this js.Value, args []js.Value) any {
		if len(args) == 0 || pointerActive {
			return nil
		}
		evt := args[0]
		// Only react to the primary pointer (mouse left button / first touch).
		// `isPrimary` is true for the first contact and the mouse.
		if isPrimary := evt.Get("isPrimary"); !isPrimary.IsUndefined() && !isPrimary.Bool() {
			return nil
		}
		evt.Call("preventDefault")

		w := s.widget()
		if w.Controller == nil {
			return nil
		}

		x := evt.Get("clientX").Float()
		y := evt.Get("clientY").Float()
		rect := n.Call("getBoundingClientRect")
		offsetX := x - rect.Get("left").Float()
		offsetY := y - rect.Get("top").Float()

		// Capture so subsequent move/up keep coming to the source even if
		// the pointer leaves its box. Window listeners also catch them, but
		// capture stops the source's parent scroller from stealing.
		pointerID := evt.Get("pointerId")
		if !pointerID.IsUndefined() && !pointerID.IsNull() {
			n.Call("setPointerCapture", pointerID.Int())
		}

		s.SetState(func() { s.active = true })
		w.Controller.startDrag(w.Data, w.Ghost, x, y, offsetX, offsetY)
		pointerActive = true

		moveCB = js.FuncOf(func(this js.Value, margs []js.Value) any {
			if len(margs) == 0 {
				return nil
			}
			me := margs[0]
			mx := me.Get("clientX").Float()
			my := me.Get("clientY").Float()
			hover := hitTestTargets(w.Controller, mx, my)
			w.Controller.updateDrag(mx, my, hover)
			return nil
		})
		upCB = js.FuncOf(func(this js.Value, uargs []js.Value) any {
			finishDrag(true)
			return nil
		})
		js.Global().Call("addEventListener", "pointermove", moveCB)
		js.Global().Call("addEventListener", "pointerup", upCB)
		return nil
	})

	n.Call("addEventListener", "pointerdown", pointerDown)

	return func() {
		if released {
			return
		}
		released = true
		finishDrag(false)
		n.Call("removeEventListener", "pointerdown", pointerDown)
		pointerDown.Release()
	}
}

// hitTestTargets returns the ID of the smallest registered DropTarget whose
// bounding-rect contains (x, y). 0 means no target matched.
//
// "Smallest" so nested targets (a column inside another column, a card
// slot inside a column) behave intuitively — the innermost match wins.
func hitTestTargets[T any](c *Controller[T], x, y float64) uint64 {
	if c == nil {
		return 0
	}
	targets := c.targetsSnapshot()
	var bestID uint64
	bestArea := -1.0
	for _, t := range targets {
		node, ok := t.Node.(js.Value)
		if !ok {
			continue
		}
		rect := node.Call("getBoundingClientRect")
		left := rect.Get("left").Float()
		top := rect.Get("top").Float()
		right := rect.Get("right").Float()
		bottom := rect.Get("bottom").Float()
		if x < left || x > right || y < top || y > bottom {
			continue
		}
		area := (right - left) * (bottom - top)
		if bestArea < 0 || area < bestArea {
			bestID = t.ID
			bestArea = area
		}
	}
	return bestID
}
