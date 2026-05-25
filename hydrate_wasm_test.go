//go:build js && wasm

package gutter

import (
	"strconv"
	"strings"
	"testing"
)

// hydCounter is a stateful button so we can prove hydration wires events: a
// click after hydrate must reach the handler and update the SAME DOM node.
type hydCounter struct{}

func (hydCounter) CreateState() State { return &hydCounterState{} }

type hydCounterState struct {
	StateObject
	n int
}

func (s *hydCounterState) Build(*BuildContext) Widget {
	return testHost{
		tag:    "button",
		text:   "count:" + strconv.Itoa(s.n),
		events: map[string]func(Event){"click": func(Event) { s.SetState(func() { s.n++ }) }},
	}
}

func TestHydrateAdoptsServerDOM(t *testing.T) {
	parent := freshParent()
	app := testHost{tag: "section", attrs: map[string]string{"id": "root"}, children: []Widget{
		testHost{tag: "h1", text: "Title"},
		hydCounter{},
	}}

	html, err := RenderToHTML(app)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(html, "data-gutter-h") {
		t.Fatalf("expected hydration marker in SSR html: %s", html)
	}
	parent.Set("innerHTML", html)
	serverRoot := parent.Get("firstChild")       // <section>
	serverButton := serverRoot.Get("lastChild")  // <button>
	if serverButton.Get("textContent").String() != "count:0" {
		t.Fatalf("server-rendered button = %q", serverButton.Get("textContent").String())
	}

	// Hydrate the existing DOM rather than mounting fresh.
	el := newElement(app)
	el.hydrate(parent.Get("children").Index(0), testCtxVal)

	// Node identity preserved (adopted, not recreated).
	if !parent.Get("firstChild").Equal(serverRoot) {
		t.Fatal("hydrate replaced the root node instead of adopting it")
	}
	// SSR-only markers stripped.
	if serverButton.Call("hasAttribute", "data-gutter-h").Bool() {
		t.Error("data-gutter-h marker not stripped during hydrate")
	}
	// Click must reach the handler wired by hydrate.
	serverButton.Call("click")
	flushRebuilds()
	if got := serverButton.Get("textContent").String(); got != "count:1" {
		t.Fatalf("after click textContent = %q, want count:1 (event not wired by hydrate)", got)
	}
	// And the update happened in place on the SAME node.
	if !serverRoot.Get("lastChild").Equal(serverButton) {
		t.Fatal("post-hydrate update replaced the button node")
	}
}

func TestHydrateTagMismatchFallback(t *testing.T) {
	parent := freshParent()
	parent.Set("innerHTML", "<p>old</p>")
	oldNode := parent.Get("firstChild")

	el := newElement(testHost{tag: "span", text: "new"})
	el.hydrate(parent.Get("children").Index(0), testCtxVal)

	node := parent.Get("firstChild")
	if node.Get("tagName").String() != "SPAN" {
		t.Fatalf("expected SPAN after mismatch fallback, got %s", node.Get("tagName").String())
	}
	if node.Equal(oldNode) {
		t.Fatal("expected a fresh node on tag mismatch, got the old one")
	}
	if node.Get("textContent").String() != "new" {
		t.Fatalf("textContent = %q, want new", node.Get("textContent").String())
	}
}
