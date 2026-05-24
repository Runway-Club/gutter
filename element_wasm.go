//go:build js && wasm

package gutter

import (
	"reflect"
	"strings"
	"syscall/js"
)

// Element is one node in the persistent reconciliation tree. Each Element
// holds its current Widget, the DOM node it owns (transitively, for non-host
// elements), and its parent DOM so it can be moved or replaced. Mount creates
// the DOM and inserts it; update diffs the new widget against the current one
// and edits the DOM in place; unmount tears down listeners and removes the
// DOM.
type Element interface {
	// mount creates the DOM for this element and inserts it into parent. If
	// before is non-null, the new DOM is inserted before that sibling;
	// otherwise it is appended.
	mount(parent, before js.Value, ctx *BuildContext)
	// update reconciles this element with newW, mutating DOM in place. The
	// caller must already have determined via canUpdate that newW is
	// compatible with the current widget.
	update(newW Widget, ctx *BuildContext)
	// unmount removes the element's DOM and releases any listeners. It also
	// recurses into children/child.
	unmount()
	// dom returns the root DOM node owned by this element. For composite
	// elements (Stateless/Stateful) this delegates to the child.
	dom() js.Value
	// widget returns the current Widget instance.
	widget() Widget
	// key returns the widget's reconciliation key (nil if unkeyed).
	key() any
}

// newElement creates the right Element type for a given Widget without
// mounting it. The element has no DOM until mount is called.
func newElement(w Widget) Element {
	switch x := w.(type) {
	case HostWidget:
		return &hostElement{w: x}
	case StatefulWidget:
		return &statefulElement{w: x}
	case StatelessWidget:
		return &statelessElement{w: x}
	}
	panic("gutter: value does not implement HostWidget, StatelessWidget, or StatefulWidget")
}

func widgetKey(w Widget) any {
	if k, ok := w.(Keyed); ok {
		return k.WidgetKey()
	}
	return nil
}

// canUpdate reports whether an existing Element can be reused in place to
// represent newW. We require both the Go type and the key to match.
func canUpdate(old Element, newW Widget) bool {
	if reflect.TypeOf(old.widget()) != reflect.TypeOf(newW) {
		return false
	}
	return old.key() == widgetKey(newW)
}

// reconcile is the single-child counterpart of reconcileChildren. It either
// updates oldEl in place or unmounts it and mounts a fresh element at the same
// DOM position.
func reconcile(parent js.Value, oldEl Element, newW Widget, ctx *BuildContext) Element {
	if oldEl != nil && canUpdate(oldEl, newW) {
		oldEl.update(newW, ctx)
		return oldEl
	}
	newEl := newElement(newW)
	if oldEl == nil {
		newEl.mount(parent, js.Null(), ctx)
		return newEl
	}
	oldDom := oldEl.dom()
	newEl.mount(parent, oldDom, ctx)
	oldEl.unmount()
	return newEl
}

// reconcileChildren matches new widgets against existing child elements using
// keys (when present) or positional matching (for unkeyed siblings of the
// same Go type), updates the matches, mounts the new ones, unmounts the
// leftover ones, then walks the result list backwards and uses insertBefore
// to land every DOM node in the correct position.
func reconcileChildren(parent js.Value, oldChildren []Element, newWidgets []Widget, ctx *BuildContext) []Element {
	// Index keyed old children.
	keyedOld := map[any]int{}
	for i, oldEl := range oldChildren {
		if k := oldEl.key(); k != nil {
			keyedOld[k] = i
		}
	}
	used := make([]bool, len(oldChildren))
	result := make([]Element, len(newWidgets))

	// Pass 1: match and update. Mark which old children we kept.
	for i, newW := range newWidgets {
		newKey := widgetKey(newW)
		var reuse Element
		if newKey != nil {
			if idx, ok := keyedOld[newKey]; ok && !used[idx] && canUpdate(oldChildren[idx], newW) {
				reuse = oldChildren[idx]
				used[idx] = true
			}
		} else {
			for j, oldEl := range oldChildren {
				if used[j] || oldEl.key() != nil {
					continue
				}
				if canUpdate(oldEl, newW) {
					reuse = oldEl
					used[j] = true
					break
				}
			}
		}
		if reuse != nil {
			reuse.update(newW, ctx)
			result[i] = reuse
		} else {
			result[i] = newElement(newW)
		}
	}

	// Pass 2: unmount unmatched old children.
	for i, u := range used {
		if !u {
			oldChildren[i].unmount()
		}
	}

	// Pass 3: position. Walk backwards so each insertBefore uses the
	// already-correctly-placed nextDom as its anchor. Freshly created
	// elements have an undefined dom() until they are mounted in this loop.
	//
	// Skip the insertBefore call when the node is already in the right
	// spot (its nextSibling === nextDom): per the DOM spec it's a no-op,
	// but in practice moving an already-focused <input> through
	// insertBefore on every reconcile blurs and refocuses it, and the
	// browser's focus restore triggers scrollIntoView — which on a long
	// page snaps scroll back to where the focused element lives.
	nextDom := js.Null()
	for i := len(result) - 1; i >= 0; i-- {
		el := result[i]
		current := el.dom()
		if current.IsUndefined() || current.IsNull() {
			el.mount(parent, nextDom, ctx)
		} else if !current.Get("nextSibling").Equal(nextDom) {
			parent.Call("insertBefore", current, nextDom)
		}
		nextDom = el.dom()
	}

	return result
}

// =========== hostElement ===========

type hostElement struct {
	parent    js.Value
	node      js.Value
	w         HostWidget
	host      *Host
	children  []Element
	listeners map[string]js.Func
}

func (e *hostElement) widget() Widget { return e.w }
func (e *hostElement) dom() js.Value  { return e.node }
func (e *hostElement) key() any       { return widgetKey(e.w) }

func (e *hostElement) mount(parent, before js.Value, ctx *BuildContext) {
	e.parent = parent
	e.host = e.w.Host()
	doc := js.Global().Get("document")
	e.node = doc.Call("createElement", e.host.Tag)
	applyAttrs(e.node, nil, e.host.Attrs)
	applyStyle(e.node, nil, e.host.Style)
	if e.host.Text != "" {
		e.node.Set("textContent", e.host.Text)
	}
	e.attachEvents(e.host.Events)
	parent.Call("insertBefore", e.node, before)
	e.children = make([]Element, 0, len(e.host.Children))
	for _, childW := range e.host.Children {
		childEl := newElement(childW)
		childEl.mount(e.node, js.Null(), ctx)
		e.children = append(e.children, childEl)
	}
	if e.host.OnMount != nil {
		e.host.OnMount(e.node)
	}
}

func (e *hostElement) update(newW Widget, ctx *BuildContext) {
	newHost := newW.(HostWidget).Host()
	if newHost.Text != e.host.Text {
		// textContent rewrites all children; do this BEFORE children
		// reconciliation if the widget actually uses text. In practice a
		// HostWidget either has text or children, not both.
		e.node.Set("textContent", newHost.Text)
	}
	applyAttrs(e.node, e.host.Attrs, newHost.Attrs)
	applyStyle(e.node, e.host.Style, newHost.Style)
	e.detachEvents()
	e.attachEvents(newHost.Events)
	e.children = reconcileChildren(e.node, e.children, newHost.Children, ctx)
	e.w = newW.(HostWidget)
	e.host = newHost
	if e.host.OnMount != nil {
		// Treat update as a remount signal for hook-style widgets (Canvas
		// re-paints when its size or paint callback changes). The hook
		// fires after the DOM is reconciled so handlers see the new state.
		e.host.OnMount(e.node)
	}
}

func (e *hostElement) unmount() {
	if e.host != nil && e.host.OnUnmount != nil {
		e.host.OnUnmount(e.node)
	}
	for _, child := range e.children {
		child.unmount()
	}
	e.children = nil
	e.detachEvents()
	if !e.parent.IsUndefined() && !e.parent.IsNull() {
		e.parent.Call("removeChild", e.node)
	}
}

func (e *hostElement) attachEvents(events map[string]func(Event)) {
	if len(events) == 0 {
		return
	}
	if e.listeners == nil {
		e.listeners = make(map[string]js.Func, len(events))
	}
	for name, handler := range events {
		n := name
		h := handler
		cb := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			ev := Event{Type: n}
			if len(args) > 0 {
				raw := args[0]
				target := raw.Get("target")
				if !target.IsUndefined() && !target.IsNull() {
					v := target.Get("value")
					if !v.IsUndefined() && !v.IsNull() {
						ev.Value = v.String()
					}
				}
				// Pointer/mouse coordinates. clientX/clientY exist on
				// MouseEvent and PointerEvent; offsetX/offsetY would be
				// element-local but aren't part of the PointerEvent spec.
				if x := raw.Get("clientX"); !x.IsUndefined() && !x.IsNull() {
					ev.X = x.Float()
				}
				if y := raw.Get("clientY"); !y.IsUndefined() && !y.IsNull() {
					ev.Y = y.Float()
				}
				if x := raw.Get("offsetX"); !x.IsUndefined() && !x.IsNull() {
					ev.OffsetX = x.Float()
				}
				if y := raw.Get("offsetY"); !y.IsUndefined() && !y.IsNull() {
					ev.OffsetY = y.Float()
				}
				if k := raw.Get("key"); !k.IsUndefined() && !k.IsNull() {
					ev.Key = k.String()
				}
			}
			h(ev)
			return nil
		})
		e.node.Call("addEventListener", n, cb)
		e.listeners[n] = cb
	}
}

func (e *hostElement) detachEvents() {
	for name, cb := range e.listeners {
		e.node.Call("removeEventListener", name, cb)
		cb.Release()
	}
	e.listeners = nil
}

func applyAttrs(node js.Value, oldAttrs, newAttrs map[string]string) {
	for k, v := range newAttrs {
		if oldAttrs == nil || oldAttrs[k] != v {
			node.Call("setAttribute", k, v)
		}
	}
	for k := range oldAttrs {
		if _, ok := newAttrs[k]; !ok {
			node.Call("removeAttribute", k)
		}
	}
}

func applyStyle(node js.Value, oldStyle, newStyle map[string]string) {
	if styleEqual(oldStyle, newStyle) {
		return
	}
	if len(newStyle) == 0 {
		node.Call("removeAttribute", "style")
		return
	}
	var b strings.Builder
	for k, v := range newStyle {
		b.WriteString(k)
		b.WriteString(":")
		b.WriteString(v)
		b.WriteString(";")
	}
	node.Call("setAttribute", "style", b.String())
}

func styleEqual(a, b map[string]string) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if b[k] != v {
			return false
		}
	}
	return true
}

// =========== statelessElement ===========

type statelessElement struct {
	parent js.Value
	w      StatelessWidget
	child  Element
}

func (e *statelessElement) widget() Widget { return e.w }
func (e *statelessElement) dom() js.Value {
	if e.child == nil {
		return js.Undefined()
	}
	return e.child.dom()
}
func (e *statelessElement) key() any { return widgetKey(e.w) }

func (e *statelessElement) mount(parent, before js.Value, ctx *BuildContext) {
	e.parent = parent
	childW := e.w.Build(ctx)
	e.child = newElement(childW)
	e.child.mount(parent, before, ctx)
}

func (e *statelessElement) update(newW Widget, ctx *BuildContext) {
	e.w = newW.(StatelessWidget)
	newChildW := e.w.Build(ctx)
	e.child = reconcile(e.parent, e.child, newChildW, ctx)
}

func (e *statelessElement) unmount() {
	if e.child != nil {
		e.child.unmount()
		e.child = nil
	}
}

// =========== statefulElement ===========

type statefulElement struct {
	parent js.Value
	ctx    *BuildContext
	w      StatefulWidget
	state  State
	child  Element
}

func (e *statefulElement) widget() Widget { return e.w }
func (e *statefulElement) dom() js.Value {
	if e.child == nil {
		return js.Undefined()
	}
	return e.child.dom()
}
func (e *statefulElement) key() any { return widgetKey(e.w) }

func (e *statefulElement) mount(parent, before js.Value, ctx *BuildContext) {
	e.parent = parent
	e.ctx = ctx
	e.state = e.w.CreateState()
	// Bind element and widget BEFORE InitState so the State can call SetState
	// from inside InitState (e.g. to flip into a loading state after spawning
	// a goroutine) and read its widget via Widget().
	if binder, ok := e.state.(elementBinder); ok {
		binder.bindElement(e)
	}
	if binder, ok := e.state.(widgetBinder); ok {
		binder.bindWidget(e.w)
	}
	if init, ok := e.state.(StateInitializer); ok {
		init.InitState()
	}
	childW := e.state.Build(ctx)
	e.child = newElement(childW)
	e.child.mount(parent, before, ctx)
}

// update is called when an ancestor rebuild produces a new widget instance of
// the same Go type. State is preserved; only the widget reference is updated.
// The subtree is then rebuilt against the new widget's state.
func (e *statefulElement) update(newW Widget, ctx *BuildContext) {
	oldW := e.w
	e.w = newW.(StatefulWidget)
	e.ctx = ctx
	if binder, ok := e.state.(widgetBinder); ok {
		binder.bindWidget(e.w)
	}
	if upd, ok := e.state.(WidgetUpdater); ok {
		upd.DidUpdateWidget(oldW)
	}
	e.rebuild()
}

// rebuild is invoked by the State when SetState fires. It rebuilds only this
// subtree, reusing the parent DOM and the current Element.
func (e *statefulElement) rebuild() {
	newChildW := e.state.Build(e.ctx)
	e.child = reconcile(e.parent, e.child, newChildW, e.ctx)
}

func (e *statefulElement) unmount() {
	if disp, ok := e.state.(StateDisposer); ok {
		disp.Dispose()
	}
	if e.child != nil {
		e.child.unmount()
		e.child = nil
	}
}
