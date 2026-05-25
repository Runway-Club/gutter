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
	// hydrate adopts an existing server-rendered DOM node at this tree position
	// instead of creating one: it wires events/lifecycle and recurses into the
	// existing children, preserving node identity. A structural mismatch falls
	// back to a fresh mount of that subtree.
	hydrate(node js.Value, ctx *BuildContext)
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
	// widgetType returns the cached reflect.Type of the current widget. It is
	// fixed for the element's lifetime: reconcile only reuses an element for a
	// widget of the same Go type, so the type never changes across update.
	// Caching it here avoids reflecting on the old widget every reconcile.
	widgetType() reflect.Type
	// key returns the widget's reconciliation key (nil if unkeyed).
	key() any
}

// newElement creates the right Element type for a given Widget without
// mounting it. The element has no DOM until mount is called.
func newElement(w Widget) Element {
	wt := reflect.TypeOf(w)
	switch x := w.(type) {
	case HostWidget:
		return &hostElement{w: x, wt: wt}
	case StatefulWidget:
		return &statefulElement{w: x, wt: wt}
	case StatelessWidget:
		return &statelessElement{w: x, wt: wt}
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
// represent newW. We require the Go type and key to match, and — for
// HostWidgets — the rendered tag too.
func canUpdate(old Element, newW Widget) bool {
	if old.widgetType() != reflect.TypeOf(newW) {
		return false
	}
	if old.key() != widgetKey(newW) {
		return false
	}
	// Same Go type but a different rendered tag (a HostWidget that varies its
	// Host().Tag by its fields) can't be updated in place: attribute-diffing a
	// <div> into a <span> leaves the wrong element in the DOM. Force a remount.
	if he, ok := old.(*hostElement); ok {
		if hw, ok := newW.(HostWidget); ok && normTag(he.host) != normTag(hw.Host()) {
			return false
		}
	}
	return true
}

// normTag is the effective tag of a Host, normalizing the empty default to the
// "div" that mount's createElement uses.
func normTag(h *Host) string {
	if h == nil || h.Tag == "" {
		return "div"
	}
	return h.Tag
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
	wt        reflect.Type
	host      *Host
	children  []Element
	listeners map[string]js.Func
}

func (e *hostElement) widget() Widget           { return e.w }
func (e *hostElement) widgetType() reflect.Type { return e.wt }
func (e *hostElement) dom() js.Value            { return e.node }
func (e *hostElement) key() any                 { return widgetKey(e.w) }

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
	e.syncEvents(e.host.Events)
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

func (e *hostElement) hydrate(node js.Value, ctx *BuildContext) {
	e.host = e.w.Host()
	e.parent = node.Get("parentNode")
	// Tag mismatch → the client built a different element than the server.
	// Can't adopt: mount a fresh node in place and drop the stale SSR one.
	if !sameTag(node, e.host.Tag) {
		e.mount(e.parent, node, ctx)
		e.parent.Call("removeChild", node)
		return
	}
	e.node = node
	// Re-apply attrs/style/text idempotently so the client is authoritative
	// even if the server markup drifted; then strip the SSR-only markers.
	applyAttrs(e.node, nil, e.host.Attrs)
	applyStyle(e.node, nil, e.host.Style)
	e.node.Call("removeAttribute", "data-gutter-h")
	e.node.Call("removeAttribute", "data-gutter-key")
	if e.host.Text != "" && e.node.Get("textContent").String() != e.host.Text {
		e.node.Set("textContent", e.host.Text)
	}
	e.syncEvents(e.host.Events)
	// Adopt children positionally: every widget resolves to exactly one DOM
	// element, so host children line up 1:1 with node.children (element-only,
	// so SSR whitespace text nodes — of which our renderer emits none — are
	// ignored regardless).
	existing := e.node.Get("children")
	count := existing.Get("length").Int()
	e.children = make([]Element, 0, len(e.host.Children))
	for i, childW := range e.host.Children {
		childEl := newElement(childW)
		if i < count {
			childEl.hydrate(existing.Index(i), ctx)
		} else {
			childEl.mount(e.node, js.Null(), ctx) // server had fewer children
		}
		e.children = append(e.children, childEl)
	}
	for i := count - 1; i >= len(e.host.Children); i-- { // server had extras
		e.node.Call("removeChild", existing.Index(i))
	}
	if e.host.OnMount != nil {
		e.host.OnMount(e.node)
	}
}

func sameTag(node js.Value, tag string) bool {
	if tag == "" {
		tag = "div"
	}
	tn := node.Get("tagName")
	if tn.IsUndefined() || tn.IsNull() {
		return false
	}
	return strings.EqualFold(tn.String(), tag)
}

func (e *hostElement) update(newW Widget, ctx *BuildContext) {
	newHost := newW.(HostWidget).Host()
	oldHost := e.host
	// Assign w/host before reconciling children or syncing events: the
	// persistent event listeners dispatch through e.host.Events at fire time,
	// so e.host must already point at the new handlers.
	e.w = newW.(HostWidget)
	e.host = newHost
	if newHost.Text != oldHost.Text {
		// textContent rewrites all children; do this BEFORE children
		// reconciliation if the widget actually uses text. In practice a
		// HostWidget either has text or children, not both.
		e.node.Set("textContent", newHost.Text)
	}
	applyAttrs(e.node, oldHost.Attrs, newHost.Attrs)
	applyStyle(e.node, oldHost.Style, newHost.Style)
	e.syncEvents(newHost.Events)
	e.children = reconcileChildren(e.node, e.children, newHost.Children, ctx)
	if newHost.OnMount != nil {
		// Treat update as a remount signal for hook-style widgets (Canvas
		// re-paints when its size or paint callback changes). The hook
		// fires after the DOM is reconciled so handlers see the new state.
		newHost.OnMount(e.node)
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
	e.releaseEvents()
	if !e.parent.IsUndefined() && !e.parent.IsNull() {
		e.parent.Call("removeChild", e.node)
	}
}

// syncEvents reconciles the set of attached DOM listeners against newEvents by
// NAME only. For each event name we register exactly one persistent js.Func
// that, when fired, looks up the current handler via e.host.Events[name] — so
// a rebuild that swaps the handler closure (the common case, since handlers are
// fresh closures every Build) requires no DOM work at all: the name set is
// unchanged and the listener already dispatches to the latest handler. Only
// added or removed event NAMES touch the DOM. This avoids the
// release-all/recreate-all churn that made every reconcile pay for
// js.FuncOf + addEventListener on every handler.
func (e *hostElement) syncEvents(newEvents map[string]func(Event)) {
	// Remove listeners whose event name is gone.
	for name, cb := range e.listeners {
		if _, ok := newEvents[name]; !ok {
			e.node.Call("removeEventListener", name, cb)
			cb.Release()
			delete(e.listeners, name)
		}
	}
	if len(newEvents) == 0 {
		return
	}
	if e.listeners == nil {
		e.listeners = make(map[string]js.Func, len(newEvents))
	}
	// Add a persistent dispatcher for each newly-seen event name.
	for name := range newEvents {
		if _, ok := e.listeners[name]; ok {
			continue
		}
		n := name
		cb := js.FuncOf(func(this js.Value, args []js.Value) any {
			h := e.host.Events[n]
			if h == nil {
				return nil
			}
			ev := Event{Type: n}
			if len(args) > 0 {
				fillEvent(&ev, n, args[0])
			}
			h(ev)
			return nil
		})
		e.node.Call("addEventListener", n, cb)
		e.listeners[n] = cb
	}
}

func (e *hostElement) releaseEvents() {
	for name, cb := range e.listeners {
		e.node.Call("removeEventListener", name, cb)
		cb.Release()
	}
	e.listeners = nil
}

// fillEvent reads only the fields of the raw DOM event that are meaningful for
// the given event name, instead of probing all six on every fire. Each
// raw.Get crosses the Go↔JS boundary, so a click that used to cost six reads
// (value + 4 coords + key) now costs four, and a text "input" event costs one.
func fillEvent(ev *Event, name string, raw js.Value) {
	if eventCarriesValue(name) {
		if target := raw.Get("target"); !target.IsUndefined() && !target.IsNull() {
			if v := target.Get("value"); !v.IsUndefined() && !v.IsNull() {
				ev.Value = v.String()
			}
		}
	}
	if eventCarriesPointer(name) {
		ev.X = floatProp(raw, "clientX")
		ev.Y = floatProp(raw, "clientY")
		ev.OffsetX = floatProp(raw, "offsetX")
		ev.OffsetY = floatProp(raw, "offsetY")
	}
	if strings.HasPrefix(name, "key") {
		if k := raw.Get("key"); !k.IsUndefined() && !k.IsNull() {
			ev.Key = k.String()
		}
	}
}

func floatProp(raw js.Value, name string) float64 {
	if v := raw.Get(name); !v.IsUndefined() && !v.IsNull() {
		return v.Float()
	}
	return 0
}

// eventCarriesValue reports whether target.value is worth reading for this
// event. Form/typing events carry it; pure pointer events don't.
func eventCarriesValue(name string) bool {
	switch name {
	case "input", "change", "blur", "focus", "submit", "search", "paste":
		return true
	}
	return strings.HasPrefix(name, "key")
}

// eventCarriesPointer reports whether clientX/clientY/offsetX/offsetY are
// meaningful for this event.
func eventCarriesPointer(name string) bool {
	switch name {
	case "click", "dblclick", "contextmenu", "wheel", "auxclick":
		return true
	}
	return strings.HasPrefix(name, "pointer") ||
		strings.HasPrefix(name, "mouse") ||
		strings.HasPrefix(name, "drag") ||
		strings.HasPrefix(name, "touch") ||
		name == "drop"
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
	wt     reflect.Type
	child  Element
}

func (e *statelessElement) widget() Widget           { return e.w }
func (e *statelessElement) widgetType() reflect.Type { return e.wt }
func (e *statelessElement) dom() js.Value {
	if e.child == nil {
		return js.Undefined()
	}
	return e.child.dom()
}
func (e *statelessElement) key() any { return widgetKey(e.w) }

func (e *statelessElement) mount(parent, before js.Value, ctx *BuildContext) {
	e.parent = parent
	saved := ctx.inherited
	if p, ok := e.w.(inheritedProvider); ok {
		ctx.inherited = p.provideInto(ctx.inherited)
	}
	childW := e.w.Build(ctx)
	e.child = newElement(childW)
	e.child.mount(parent, before, ctx)
	ctx.inherited = saved
}

func (e *statelessElement) hydrate(node js.Value, ctx *BuildContext) {
	e.parent = node.Get("parentNode")
	saved := ctx.inherited
	if p, ok := e.w.(inheritedProvider); ok {
		ctx.inherited = p.provideInto(ctx.inherited)
	}
	childW := e.w.Build(ctx)
	e.child = newElement(childW)
	e.child.hydrate(node, ctx) // DOM-transparent: the child owns this node
	ctx.inherited = saved
}

func (e *statelessElement) update(newW Widget, ctx *BuildContext) {
	e.w = newW.(StatelessWidget)
	saved := ctx.inherited
	if p, ok := e.w.(inheritedProvider); ok {
		ctx.inherited = p.provideInto(ctx.inherited)
	}
	newChildW := e.w.Build(ctx)
	e.child = reconcile(e.parent, e.child, newChildW, ctx)
	ctx.inherited = saved
}

func (e *statelessElement) unmount() {
	if e.child != nil {
		e.child.unmount()
		e.child = nil
	}
}

// =========== statefulElement ===========

type statefulElement struct {
	parent  js.Value
	ctx     *BuildContext
	w       StatefulWidget
	wt      reflect.Type
	state   State
	child   Element
	mounted bool
	// scope is the ambient dependency scope (Provider values) visible at this
	// element's position, captured during top-down passes so an isolated
	// SetState rebuild — which doesn't re-run ancestor Providers — can restore
	// it before calling Build. nil when no Provider is in scope.
	scope map[reflect.Type]any
}

func (e *statefulElement) widget() Widget           { return e.w }
func (e *statefulElement) widgetType() reflect.Type { return e.wt }
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
	e.scope = ctx.inherited // capture ambient scope for later isolated rebuilds
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
	e.mounted = true
}

func (e *statefulElement) hydrate(node js.Value, ctx *BuildContext) {
	e.parent = node.Get("parentNode")
	e.ctx = ctx
	e.scope = ctx.inherited
	e.state = e.w.CreateState()
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
	e.child.hydrate(node, ctx)
	e.mounted = true
}

// update is called when an ancestor rebuild produces a new widget instance of
// the same Go type. State is preserved; only the widget reference is updated.
// The subtree is then rebuilt against the new widget's state.
func (e *statefulElement) update(newW Widget, ctx *BuildContext) {
	oldW := e.w
	e.w = newW.(StatefulWidget)
	e.ctx = ctx
	e.scope = ctx.inherited // refresh: a top-down pass carries the current scope
	if binder, ok := e.state.(widgetBinder); ok {
		binder.bindWidget(e.w)
	}
	if upd, ok := e.state.(WidgetUpdater); ok {
		upd.DidUpdateWidget(oldW)
	}
	e.rebuild()
}

// rebuild rebuilds only this subtree, reusing the parent DOM and the current
// Element. It is the synchronous path taken by an ancestor-driven update.
func (e *statefulElement) rebuild() {
	if !e.mounted {
		return
	}
	// Restore the ambient scope this element lives under: an isolated SetState
	// rebuild doesn't re-run ancestor Providers, so without this DependOn would
	// see an empty (or wrong) scope.
	saved := e.ctx.inherited
	e.ctx.inherited = e.scope
	newChildW := e.state.Build(e.ctx)
	e.child = reconcile(e.parent, e.child, newChildW, e.ctx)
	e.ctx.inherited = saved
}

// scheduleRebuild is the SetState path: rather than rebuilding synchronously,
// it enqueues this element and lets the microtask flush coalesce repeated
// SetState calls (and SetStates across sibling elements) into a single rebuild
// pass per element. Mirrors React's batched updates.
func (e *statefulElement) scheduleRebuild() {
	enqueueRebuild(e)
}

func (e *statefulElement) unmount() {
	e.mounted = false
	if disp, ok := e.state.(StateDisposer); ok {
		disp.Dispose()
	}
	if e.child != nil {
		e.child.unmount()
		e.child = nil
	}
}

// =========== batched rebuild scheduler ===========
//
// SetState enqueues its element here instead of rebuilding inline. The first
// enqueue of a flush cycle schedules a microtask (queueMicrotask) that drains
// the queue. Repeated SetState on the same element within one tick collapses to
// one rebuild; SetStates on different elements all flush together. Elements
// unmounted before the flush are skipped via their mounted flag.
//
// WASM Go runs on the single JS event loop, so no locking is needed: enqueue
// and flush never interleave.
var (
	rebuildQueue   []*statefulElement
	rebuildQueued  = map[*statefulElement]bool{}
	flushScheduled bool
	flushFn        js.Func
)

func enqueueRebuild(e *statefulElement) {
	if rebuildQueued[e] {
		return
	}
	rebuildQueued[e] = true
	rebuildQueue = append(rebuildQueue, e)
	if flushScheduled {
		return
	}
	flushScheduled = true
	if flushFn.IsUndefined() {
		flushFn = js.FuncOf(func(js.Value, []js.Value) any {
			flushRebuilds()
			return nil
		})
	}
	js.Global().Call("queueMicrotask", flushFn)
}

func flushRebuilds() {
	// Snapshot and reset before rebuilding: a rebuild may itself call SetState,
	// which must land in the next cycle's queue, not this drain.
	queue := rebuildQueue
	rebuildQueue = nil
	rebuildQueued = map[*statefulElement]bool{}
	flushScheduled = false
	for _, e := range queue {
		e.rebuild()
	}
}
