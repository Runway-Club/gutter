package widgets

import (
	"fmt"
	"sync"

	"github.com/Runway-Club/gutter"
)

// DragState is the snapshot of an in-flight drag. Observers receive it via
// Controller.Listen and Controller.Value. Active is false when no drag is
// happening; the other fields are meaningful only while Active is true.
type DragState[T any] struct {
	Active bool
	// Payload is the data carried by the dragged item — the Draggable.Data
	// of the source.
	Payload T
	// Pointer position in viewport coordinates.
	X, Y float64
	// OffsetX/Y is the pointer offset within the source card at drag start,
	// so DragOverlay can anchor the ghost where the user grabbed it instead
	// of jumping to the cursor's hot-spot.
	OffsetX, OffsetY float64
	// Ghost is the widget DragOverlay renders under the pointer. Set from
	// Draggable.Ghost at drag start.
	Ghost gutter.Widget
	// HoverID is the internal ID of the DropTarget currently under the
	// pointer, or 0 if the pointer isn't over any registered target.
	HoverID uint64
}

// Controller is the shared state that coordinates a set of Draggables and
// DropTargets. Construct one per drag domain (e.g. one for kanban cards),
// pass the pointer to every Draggable, DropTarget, and DragOverlay that
// participates.
//
// Implements gutter.Listenable[DragState[T]] so widgets can subscribe.
type Controller[T any] struct {
	mu           sync.RWMutex
	state        DragState[T]
	targets      map[uint64]*targetReg[T]
	hoverFns     map[uint64]func(bool)
	listeners    map[int]func(DragState[T])
	nextTargetID uint64
	nextListenID int
}

// NewController returns an empty Controller.
func NewController[T any]() *Controller[T] {
	return &Controller[T]{
		targets:   map[uint64]*targetReg[T]{},
		hoverFns:  map[uint64]func(bool){},
		listeners: map[int]func(DragState[T]){},
	}
}

// targetReg is the per-DropTarget registration the Controller keeps.
type targetReg[T any] struct {
	node    any // platform-native DOM handle (js.Value on WASM)
	accepts func(T) bool
	onDrop  func(T)
}

// Value returns the current drag state.
func (c *Controller[T]) Value() DragState[T] {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.state
}

// Listen subscribes fn to drag-state changes. Returns an idempotent cancel.
func (c *Controller[T]) Listen(fn func(DragState[T])) func() {
	c.mu.Lock()
	id := c.nextListenID
	c.nextListenID++
	c.listeners[id] = fn
	c.mu.Unlock()
	return func() {
		c.mu.Lock()
		delete(c.listeners, id)
		c.mu.Unlock()
	}
}

func (c *Controller[T]) snapshotListeners() []func(DragState[T]) {
	out := make([]func(DragState[T]), 0, len(c.listeners))
	for _, l := range c.listeners {
		out = append(out, l)
	}
	return out
}

func (c *Controller[T]) fire(snapshot DragState[T]) {
	c.mu.RLock()
	ls := c.snapshotListeners()
	c.mu.RUnlock()
	for _, l := range ls {
		l(snapshot)
	}
}

// startDrag is called by the WASM pointer wiring when a pointerdown on a
// Draggable begins a drag. payload + ghost are captured from the source;
// the offsets are pointer-relative-to-card so DragOverlay anchors right.
func (c *Controller[T]) startDrag(payload T, ghost gutter.Widget, x, y, offsetX, offsetY float64) {
	c.mu.Lock()
	c.state = DragState[T]{
		Active:  true,
		Payload: payload,
		Ghost:   ghost,
		X:       x, Y: y,
		OffsetX: offsetX, OffsetY: offsetY,
	}
	c.mu.Unlock()
	c.fire(c.Value())
}

// updateDrag is called on every pointermove during an active drag. hoverID
// is the ID of the smallest registered target whose bounding box contains
// the pointer, or 0 if none.
func (c *Controller[T]) updateDrag(x, y float64, hoverID uint64) {
	c.mu.Lock()
	oldHover := c.state.HoverID
	c.state.X = x
	c.state.Y = y
	c.state.HoverID = hoverID
	hoverIn := c.hoverFns[hoverID]
	hoverOut := c.hoverFns[oldHover]
	c.mu.Unlock()

	if oldHover != hoverID {
		if oldHover != 0 && hoverOut != nil {
			hoverOut(false)
		}
		if hoverID != 0 && hoverIn != nil {
			hoverIn(true)
		}
	}
	c.fire(c.Value())
}

// cancelDrag aborts an in-flight drag without firing any DropTarget.OnDrop.
// Used when the Draggable unmounts mid-drag, or when the pointer was
// cancelled by the browser (e.g. switched tabs).
func (c *Controller[T]) cancelDrag() {
	c.mu.Lock()
	hover := c.state.HoverID
	hoverOut := c.hoverFns[hover]
	c.state = DragState[T]{}
	c.mu.Unlock()
	if hover != 0 && hoverOut != nil {
		hoverOut(false)
	}
	c.fire(c.Value())
}

// endDrag is called on pointerup. If the pointer is over an accepting
// target, fires its OnDrop. Returns whether a drop landed and the source's
// payload, so the Draggable can call OnDragEnd accordingly.
func (c *Controller[T]) endDrag() (payload T, dropped bool) {
	c.mu.Lock()
	hover := c.state.HoverID
	payload = c.state.Payload
	reg := c.targets[hover]
	hoverOut := c.hoverFns[hover]
	if reg != nil && (reg.accepts == nil || reg.accepts(payload)) {
		dropped = true
	}
	var onDrop func(T)
	if dropped && reg != nil {
		onDrop = reg.onDrop
	}
	c.state = DragState[T]{}
	c.mu.Unlock()

	if hover != 0 && hoverOut != nil {
		hoverOut(false)
	}
	if onDrop != nil {
		onDrop(payload)
	}
	c.fire(c.Value())
	return
}

func (c *Controller[T]) register(node any, accepts func(T) bool, onDrop func(T), onHover func(bool)) uint64 {
	c.mu.Lock()
	c.nextTargetID++
	id := c.nextTargetID
	c.targets[id] = &targetReg[T]{node: node, accepts: accepts, onDrop: onDrop}
	if onHover != nil {
		c.hoverFns[id] = onHover
	}
	c.mu.Unlock()
	return id
}

func (c *Controller[T]) unregister(id uint64) {
	c.mu.Lock()
	delete(c.targets, id)
	delete(c.hoverFns, id)
	c.mu.Unlock()
}

// targetsSnapshot returns the registered targets for hit-testing. Internal —
// used by the WASM pointer code. Caller must not mutate.
func (c *Controller[T]) targetsSnapshot() []targetSnapshot[T] {
	c.mu.RLock()
	defer c.mu.RUnlock()
	out := make([]targetSnapshot[T], 0, len(c.targets))
	for id, reg := range c.targets {
		out = append(out, targetSnapshot[T]{ID: id, Node: reg.node})
	}
	return out
}

type targetSnapshot[T any] struct {
	ID   uint64
	Node any
}

// ============================ Draggable ============================

// Draggable wraps a child and makes it draggable via pointer. On
// pointerdown the widget captures the pointer, fires the Controller's
// startDrag with Data and Ghost, and hands off to window-level move/up
// listeners. Place a [DragOverlay] sharing the same Controller somewhere
// in the tree so the ghost is actually rendered.
//
//	ctrl := widgets.NewController[Task]()
//
//	widgets.Draggable[Task]{
//	    Controller: ctrl,
//	    Data:       task,
//	    Ghost:      cardWidget(task),
//	    Child:      cardWidget(task),
//	    OnDragEnd:  func(dropped bool) { /* clear local highlight */ },
//	}
type Draggable[T any] struct {
	Controller *Controller[T]
	Data       T
	Child      gutter.Widget
	// Ghost is the widget DragOverlay renders under the pointer while this
	// drag is active. Typically a duplicate of Child (or a stripped-down
	// preview). If nil, the overlay renders nothing — only hover highlights
	// guide the user.
	Ghost gutter.Widget
	// OnDragEnd fires when the pointer is released, with dropped=true when
	// a DropTarget accepted the payload.
	OnDragEnd func(dropped bool)
	// Disabled blocks the pointerdown handler. The child still renders.
	Disabled bool
}

func (d Draggable[T]) CreateState() gutter.State { return &draggableState[T]{} }

type draggableState[T any] struct {
	gutter.StateObject
	cleanup func()
	// active is true while this specific Draggable is the source of an
	// in-flight drag. Used to apply a lifted-card visual (reduced opacity)
	// so the source slot remains visible but visually different from
	// non-dragged cards.
	active bool
}

func (s *draggableState[T]) widget() Draggable[T] { return s.Widget().(Draggable[T]) }

func (s *draggableState[T]) Build(ctx *gutter.BuildContext) gutter.Widget {
	w := s.widget()
	style := map[string]string{
		"cursor":       "grab",
		"touch-action": "none", // prevent the browser from scrolling on drag
	}
	if w.Disabled {
		style["cursor"] = "default"
	}
	if s.active {
		style["opacity"] = "0.4"
	}
	return propSyncHost{
		tag:      "div",
		style:    style,
		children: []gutter.Widget{w.Child},
		onMount: func(node any) {
			if s.cleanup != nil {
				return
			}
			if w.Disabled || w.Controller == nil {
				return
			}
			s.cleanup = attachDragSource[T](node, s)
		},
	}
}

func (s *draggableState[T]) Dispose() {
	if s.cleanup != nil {
		s.cleanup()
		s.cleanup = nil
	}
}

// ============================ DropTarget ============================

// DropTarget wraps a child and registers itself with the Controller as a
// drop site. When a drag ends with the pointer over this target and
// Accepts returns true (or is nil), OnDrop fires with the payload.
//
// OnHoverChange optionally fires whenever this specific target enters or
// leaves the dragged pointer's hover area, letting the caller toggle a
// visual highlight via SetState.
type DropTarget[T any] struct {
	Controller    *Controller[T]
	Child         gutter.Widget
	Accepts       func(T) bool
	OnDrop        func(T)
	OnHoverChange func(over bool)
}

func (d DropTarget[T]) CreateState() gutter.State { return &dropTargetState[T]{} }

type dropTargetState[T any] struct {
	gutter.StateObject
	id      uint64
	node    any
	cleanup func()
}

func (s *dropTargetState[T]) widget() DropTarget[T] { return s.Widget().(DropTarget[T]) }

func (s *dropTargetState[T]) Build(ctx *gutter.BuildContext) gutter.Widget {
	w := s.widget()
	// The wrapper deliberately renders as a normal flow div (not
	// display:contents): hit-testing reads getBoundingClientRect, which
	// returns a zero-size rect on display:contents elements. Letting the
	// wrapper take the child's natural box keeps the layout the same and
	// gives the hit-tester something to measure.
	return propSyncHost{
		tag:      "div",
		children: []gutter.Widget{w.Child},
		onMount: func(node any) {
			if s.cleanup != nil {
				return
			}
			if w.Controller == nil {
				return
			}
			s.node = node
			s.id = w.Controller.register(node, w.Accepts, w.OnDrop, w.OnHoverChange)
			ctrl := w.Controller
			id := s.id
			s.cleanup = func() { ctrl.unregister(id) }
		},
	}
}

func (s *dropTargetState[T]) Dispose() {
	if s.cleanup != nil {
		s.cleanup()
		s.cleanup = nil
	}
}

// ============================ DragOverlay ============================

// DragOverlay renders the ghost from an active drag. Mount it once near
// the root of your app (after the main body, before overlays like Popup is
// fine — DragOverlay uses a higher z-index). It uses ObserverBuilder under
// the hood so only this widget rebuilds while a drag is in flight.
type DragOverlay[T any] struct {
	Controller *Controller[T]
	// ZIndex defaults to "1500" so the ghost sits above Popup/Drawer
	// (z-index 1000) but below absolutely-pinned debugging UI.
	ZIndex string
}

func (d DragOverlay[T]) Build(ctx *gutter.BuildContext) gutter.Widget {
	if d.Controller == nil {
		return Styled{}
	}
	z := d.ZIndex
	if z == "" {
		z = "1500"
	}
	return ObserverBuilder[DragState[T]]{
		Source: d.Controller,
		Builder: func(_ *gutter.BuildContext, s DragState[T]) gutter.Widget {
			if !s.Active || s.Ghost == nil {
				return Styled{}
			}
			return Styled{
				Style: map[string]string{
					"position":       "fixed",
					"left":           fmt.Sprintf("%gpx", s.X-s.OffsetX),
					"top":            fmt.Sprintf("%gpx", s.Y-s.OffsetY),
					"z-index":        z,
					"pointer-events": "none",
					"opacity":        "0.85",
					// The ghost shouldn't catch its own events; pointer-events
					// none guarantees pointer hit-tests pass through to the
					// drop targets behind it.
				},
				Children: []gutter.Widget{s.Ghost},
			}
		},
	}
}
