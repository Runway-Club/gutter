//go:build js && wasm

package gutter

import (
	"syscall/js"
	"testing"
)

// These tests exercise the real reconciler (element_wasm.go) against a live
// browser DOM. Run with wasmbrowsertest:
//
//	go install github.com/agnivade/wasmbrowsertest@latest
//	mv $(go env GOPATH)/bin/wasmbrowsertest $(go env GOPATH)/bin/go_js_wasm_exec
//	GOOS=js GOARCH=wasm go test ./...

var (
	doc        = js.Global().Get("document")
	testCtxVal = &BuildContext{}
)

// freshParent returns a detached <div> to mount into, isolating each test.
func freshParent() js.Value { return doc.Call("createElement", "div") }

// ---- test widgets ----

// testHost is a minimal HostWidget with configurable tag/text/attrs/style and
// children, used to drive the reconciler directly.
type testHost struct {
	tag      string
	text     string
	attrs    map[string]string
	style    map[string]string
	children []Widget
	events   map[string]func(Event)
}

func (h testHost) Host() *Host {
	return &Host{
		Tag:      h.tag,
		Text:     h.text,
		Attrs:    h.attrs,
		Style:    h.style,
		Children: h.children,
		Events:   h.events,
	}
}

// keyedHost participates in keyed reconciliation.
type keyedHost struct {
	k     string
	label string
}

func (h keyedHost) Host() *Host {
	return &Host{Tag: "div", Text: h.label, Attrs: map[string]string{"data-k": h.k}}
}
func (h keyedHost) WidgetKey() any { return h.k }

// ---- mount / update / unmount ----

func TestMountCreatesDOM(t *testing.T) {
	parent := freshParent()
	el := newElement(testHost{tag: "p", text: "hi", attrs: map[string]string{"id": "x"}, style: map[string]string{"color": "red"}})
	el.mount(parent, js.Null(), testCtxVal)

	node := parent.Get("firstChild")
	if node.Get("tagName").String() != "P" {
		t.Fatalf("tagName = %q, want P", node.Get("tagName").String())
	}
	if node.Get("textContent").String() != "hi" {
		t.Errorf("textContent = %q", node.Get("textContent").String())
	}
	if node.Call("getAttribute", "id").String() != "x" {
		t.Errorf("id attr = %q", node.Call("getAttribute", "id").String())
	}
	if got := node.Get("style").Get("color").String(); got != "red" {
		t.Errorf("style.color = %q, want red", got)
	}
}

func TestUpdateMutatesInPlace(t *testing.T) {
	parent := freshParent()
	el := newElement(testHost{tag: "div", text: "before", attrs: map[string]string{"id": "a"}})
	el.mount(parent, js.Null(), testCtxVal)
	node := parent.Get("firstChild")

	el.update(testHost{tag: "div", text: "after", attrs: map[string]string{"id": "b"}}, testCtxVal)

	// Same node object — updated in place, not replaced.
	if !parent.Get("firstChild").Equal(node) {
		t.Fatal("update replaced the node instead of mutating in place")
	}
	if node.Get("textContent").String() != "after" {
		t.Errorf("textContent = %q, want after", node.Get("textContent").String())
	}
	if node.Call("getAttribute", "id").String() != "b" {
		t.Errorf("id = %q, want b", node.Call("getAttribute", "id").String())
	}
}

func TestUpdateRemovesDroppedAttr(t *testing.T) {
	parent := freshParent()
	el := newElement(testHost{tag: "div", attrs: map[string]string{"id": "a", "title": "t"}})
	el.mount(parent, js.Null(), testCtxVal)
	node := parent.Get("firstChild")

	el.update(testHost{tag: "div", attrs: map[string]string{"id": "a"}}, testCtxVal)
	if node.Call("hasAttribute", "title").Bool() {
		t.Error("dropped attribute 'title' was not removed")
	}
}

func TestReconcileReplacesDifferentType(t *testing.T) {
	parent := freshParent()
	el := reconcile(parent, nil, testHost{tag: "div", text: "x"}, testCtxVal)
	first := el.dom()
	// A different Go type can't update in place — it must remount.
	el2 := reconcile(parent, el, keyedHost{k: "1", label: "y"}, testCtxVal)
	if el2.dom().Equal(first) {
		t.Error("different widget types should remount, not reuse the DOM node")
	}
	if parent.Get("childNodes").Get("length").Int() != 1 {
		t.Errorf("expected exactly 1 child after replace, got %d", parent.Get("childNodes").Get("length").Int())
	}
}

// TestReconcileRemountsOnTagChange covers the tag-stability rule: a HostWidget
// of the SAME Go type but a different rendered tag must remount (you can't morph
// a <div> into a <span> by attribute diffing), while an unchanged tag updates in
// place and keeps its DOM node.
func TestReconcileRemountsOnTagChange(t *testing.T) {
	parent := freshParent()
	el := reconcile(parent, nil, testHost{tag: "div", text: "x"}, testCtxVal)
	divNode := el.dom()
	if got := divNode.Get("tagName").String(); got != "DIV" {
		t.Fatalf("initial tag = %q, want DIV", got)
	}

	// Same type, same tag → update in place, same node.
	elSame := reconcile(parent, el, testHost{tag: "div", text: "y"}, testCtxVal)
	if !elSame.dom().Equal(divNode) {
		t.Error("same tag should update in place, not remount")
	}
	if divNode.Get("textContent").String() != "y" {
		t.Errorf("in-place update text = %q, want y", divNode.Get("textContent").String())
	}

	// Same type, different tag → remount into a <span>.
	elDiff := reconcile(parent, elSame, testHost{tag: "span", text: "z"}, testCtxVal)
	if elDiff.dom().Equal(divNode) {
		t.Error("changed tag should remount, not reuse the <div> node")
	}
	if got := elDiff.dom().Get("tagName").String(); got != "SPAN" {
		t.Fatalf("remounted tag = %q, want SPAN", got)
	}
	if parent.Get("childNodes").Get("length").Int() != 1 {
		t.Errorf("expected exactly 1 child after remount, got %d", parent.Get("childNodes").Get("length").Int())
	}
}

// portalRootHas reports whether any child of #gutter-portal-root has the given
// textContent (the portal root is shared across tests).
func portalRootHas(text string) bool {
	root := js.Global().Get("document").Call("getElementById", "gutter-portal-root")
	if root.IsNull() {
		return false
	}
	kids := root.Get("childNodes")
	for i := range kids.Get("length").Int() {
		if kids.Index(i).Get("textContent").String() == text {
			return true
		}
	}
	return false
}

func TestPortalTeleportsChild(t *testing.T) {
	parent := freshParent()
	el := newElement(Portal{Child: testHost{tag: "p", text: "ported"}})
	el.mount(parent, js.Null(), testCtxVal)

	// The tree position holds only the zero-size <template> anchor.
	if n := parent.Get("childNodes").Get("length").Int(); n != 1 {
		t.Fatalf("parent child count = %d, want 1 (the anchor)", n)
	}
	anchor := parent.Get("firstChild")
	if got := anchor.Get("tagName").String(); got != "TEMPLATE" {
		t.Fatalf("anchor tag = %q, want TEMPLATE", got)
	}
	if !el.dom().Equal(anchor) {
		t.Error("portal dom() should be the anchor node")
	}
	// The child is teleported into the body-level portal root, not the parent.
	if parent.Get("textContent").String() == "ported" {
		t.Error("child should not be in the parent subtree")
	}
	if !portalRootHas("ported") {
		t.Error("child not found in #gutter-portal-root")
	}

	// Update reconciles the child in place (in the portal root).
	el.update(Portal{Child: testHost{tag: "p", text: "updated"}}, testCtxVal)
	if !portalRootHas("updated") {
		t.Error("updated child not found in portal root")
	}

	// Unmount removes the anchor and the teleported child.
	el.unmount()
	if n := parent.Get("childNodes").Get("length").Int(); n != 0 {
		t.Errorf("anchor not removed on unmount: %d children remain", n)
	}
	if portalRootHas("updated") {
		t.Error("teleported child not removed from portal root on unmount")
	}
}

// ---- keyed reconciliation ----

func childTexts(parent js.Value) []string {
	n := parent.Get("childNodes").Get("length").Int()
	out := make([]string, n)
	for i := range n {
		out[i] = parent.Get("childNodes").Index(i).Get("textContent").String()
	}
	return out
}

func TestKeyedReorderPreservesNodeIdentity(t *testing.T) {
	parent := freshParent()
	el := newElement(testHost{tag: "div", children: []Widget{
		keyedHost{k: "A", label: "A"}, keyedHost{k: "B", label: "B"}, keyedHost{k: "C", label: "C"},
	}})
	el.mount(parent, js.Null(), testCtxVal)
	host := parent.Get("firstChild")

	// Grab the node for key "A" before reordering.
	nodeA := host.Get("childNodes").Index(0)
	if nodeA.Call("getAttribute", "data-k").String() != "A" {
		t.Fatal("setup: first child is not key A")
	}

	// Reorder to C, A, B.
	el.update(testHost{tag: "div", children: []Widget{
		keyedHost{k: "C", label: "C"}, keyedHost{k: "A", label: "A"}, keyedHost{k: "B", label: "B"},
	}}, testCtxVal)

	got := childTexts(host)
	if !(got[0] == "C" && got[1] == "A" && got[2] == "B") {
		t.Fatalf("after reorder, child texts = %v, want [C A B]", got)
	}
	// Key A must be the SAME DOM node (moved, not recreated).
	movedA := host.Get("childNodes").Index(1)
	if !movedA.Equal(nodeA) {
		t.Error("keyed reorder recreated node A instead of moving it")
	}
}

func TestUnkeyedPositionalReuse(t *testing.T) {
	parent := freshParent()
	el := newElement(testHost{tag: "div", children: []Widget{
		testHost{tag: "span", text: "1"}, testHost{tag: "span", text: "2"},
	}})
	el.mount(parent, js.Null(), testCtxVal)
	host := parent.Get("firstChild")
	firstSpan := host.Get("childNodes").Index(0)

	el.update(testHost{tag: "div", children: []Widget{
		testHost{tag: "span", text: "one"}, testHost{tag: "span", text: "two"},
	}}, testCtxVal)

	if !host.Get("childNodes").Index(0).Equal(firstSpan) {
		t.Error("unkeyed same-type child should be reused in place")
	}
	if got := childTexts(host); got[0] != "one" || got[1] != "two" {
		t.Errorf("child texts = %v, want [one two]", got)
	}
}

func TestReconcileChildrenGrowsAndShrinks(t *testing.T) {
	parent := freshParent()
	el := newElement(testHost{tag: "div", children: []Widget{testHost{tag: "i", text: "1"}}})
	el.mount(parent, js.Null(), testCtxVal)
	host := parent.Get("firstChild")

	el.update(testHost{tag: "div", children: []Widget{
		testHost{tag: "i", text: "1"}, testHost{tag: "i", text: "2"}, testHost{tag: "i", text: "3"},
	}}, testCtxVal)
	if n := host.Get("childNodes").Get("length").Int(); n != 3 {
		t.Fatalf("after grow, children = %d, want 3", n)
	}
	el.update(testHost{tag: "div", children: []Widget{testHost{tag: "i", text: "1"}}}, testCtxVal)
	if n := host.Get("childNodes").Get("length").Int(); n != 1 {
		t.Fatalf("after shrink, children = %d, want 1", n)
	}
}

// ---- events ----

func TestEventDispatchPayloadPointer(t *testing.T) {
	parent := freshParent()
	var got Event
	el := newElement(testHost{tag: "button", events: map[string]func(Event){
		"click": func(e Event) { got = e },
	}})
	el.mount(parent, js.Null(), testCtxVal)
	node := parent.Get("firstChild")

	evt := js.Global().Get("MouseEvent").New("click", map[string]any{"clientX": 12, "clientY": 34, "bubbles": true})
	node.Call("dispatchEvent", evt)

	if got.Type != "click" {
		t.Errorf("event type = %q, want click", got.Type)
	}
	if got.X != 12 || got.Y != 34 {
		t.Errorf("pointer coords = (%v,%v), want (12,34)", got.X, got.Y)
	}
}

func TestEventDispatchInputValue(t *testing.T) {
	parent := freshParent()
	var got Event
	el := newElement(testHost{tag: "input", events: map[string]func(Event){
		"input": func(e Event) { got = e },
	}})
	el.mount(parent, js.Null(), testCtxVal)
	node := parent.Get("firstChild")
	node.Set("value", "typed text")

	evt := js.Global().Get("Event").New("input")
	node.Call("dispatchEvent", evt)
	if got.Value != "typed text" {
		t.Errorf("input event value = %q, want %q", got.Value, "typed text")
	}
}

func TestEventHandlerSwapsAcrossUpdate(t *testing.T) {
	parent := freshParent()
	var which string
	el := newElement(testHost{tag: "button", events: map[string]func(Event){
		"click": func(Event) { which = "first" },
	}})
	el.mount(parent, js.Null(), testCtxVal)
	node := parent.Get("firstChild")

	// Update with a new handler closure for the same event name. The
	// persistent per-name listener must now dispatch to the new handler.
	el.update(testHost{tag: "button", events: map[string]func(Event){
		"click": func(Event) { which = "second" },
	}}, testCtxVal)

	node.Call("dispatchEvent", js.Global().Get("MouseEvent").New("click"))
	if which != "second" {
		t.Errorf("after update, click routed to %q, want second", which)
	}
}

func TestEventRemovedOnUpdate(t *testing.T) {
	parent := freshParent()
	fired := false
	el := newElement(testHost{tag: "button", events: map[string]func(Event){
		"click": func(Event) { fired = true },
	}})
	el.mount(parent, js.Null(), testCtxVal)
	node := parent.Get("firstChild")

	// Remove the click handler entirely.
	el.update(testHost{tag: "button"}, testCtxVal)
	node.Call("dispatchEvent", js.Global().Get("MouseEvent").New("click"))
	if fired {
		t.Error("click handler fired after it was removed on update")
	}
}

// ---- batched SetState ----

type batchWidget struct{}

func (batchWidget) CreateState() State { return &batchState{} }

type batchState struct {
	StateObject
	builds int
}

func (s *batchState) Build(*BuildContext) Widget {
	s.builds++
	return testHost{tag: "div"}
}

func TestSetStateBatchesIntoOneRebuild(t *testing.T) {
	parent := freshParent()
	el := newElement(batchWidget{}).(*statefulElement)
	el.mount(parent, js.Null(), testCtxVal)
	st := el.state.(*batchState)

	if st.builds != 1 {
		t.Fatalf("after mount builds = %d, want 1", st.builds)
	}

	st.SetState(func() {})
	st.SetState(func() {})
	st.SetState(func() {})
	// Still batched — no synchronous rebuild yet.
	if st.builds != 1 {
		t.Fatalf("SetState rebuilt synchronously (builds=%d); batching is broken", st.builds)
	}

	flushRebuilds() // drain the microtask queue deterministically
	if st.builds != 2 {
		t.Fatalf("after flush builds = %d, want 2 (three SetStates coalesced into one rebuild)", st.builds)
	}
}

func TestUnmountedElementNotRebuilt(t *testing.T) {
	parent := freshParent()
	el := newElement(batchWidget{}).(*statefulElement)
	el.mount(parent, js.Null(), testCtxVal)
	st := el.state.(*batchState)

	st.SetState(func() {}) // enqueue
	el.unmount()           // unmount before the flush
	flushRebuilds()
	if st.builds != 1 {
		t.Errorf("unmounted element was rebuilt (builds=%d); mounted-guard failed", st.builds)
	}
}

// ---- dispose lifecycle ----

var disposeFlag bool

type dispWidget struct{}

func (dispWidget) CreateState() State { return &dispState{} }

type dispState struct{ StateObject }

func (s *dispState) Build(*BuildContext) Widget { return testHost{tag: "div"} }
func (s *dispState) Dispose()                   { disposeFlag = true }

func TestUnmountRemovesNodeAndDisposes(t *testing.T) {
	parent := freshParent()
	disposeFlag = false
	el := newElement(dispWidget{})
	el.mount(parent, js.Null(), testCtxVal)
	if parent.Get("childNodes").Get("length").Int() != 1 {
		t.Fatal("setup: expected one child after mount")
	}
	el.unmount()
	if parent.Get("childNodes").Get("length").Int() != 0 {
		t.Error("unmount did not remove the DOM node")
	}
	if !disposeFlag {
		t.Error("State.Dispose was not called on unmount")
	}
}
