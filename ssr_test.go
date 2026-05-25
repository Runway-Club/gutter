package gutter

import (
	"fmt"
	"strings"
	"testing"
)

// --- inline test widgets (package gutter can't import widgets) ---

type ssrBox struct {
	tag      string
	text     string
	attrs    map[string]string
	style    map[string]string
	children []Widget
}

func (b ssrBox) Host() *Host {
	return &Host{Tag: b.tag, Text: b.text, Attrs: b.attrs, Style: b.style, Children: b.children}
}

type ssrBtn struct{}

func (ssrBtn) Host() *Host {
	return &Host{Tag: "button", Text: "go", Events: map[string]func(Event){"click": func(Event) {}}}
}

type ssrKeyed struct{ k string }

func (ssrKeyed) Host() *Host          { return &Host{Tag: "li", Text: "item"} }
func (k ssrKeyed) WidgetKey() any     { return k.k }

type ssrWrap struct{ child Widget }

func (w ssrWrap) Build(*BuildContext) Widget { return w.child }

type ssrCounter struct{ start int }

func (c ssrCounter) CreateState() State { return &ssrCounterState{n: c.start} }

type ssrCounterState struct {
	StateObject
	n         int
	initFired bool
}

func (s *ssrCounterState) InitState()                 { s.initFired = true; s.n += 100 }
func (s *ssrCounterState) Build(*BuildContext) Widget { return ssrBox{tag: "span", text: fmt.Sprintf("n=%d", s.n)} }

func mustRender(t *testing.T, w Widget) string {
	t.Helper()
	out, err := RenderToHTML(w)
	if err != nil {
		t.Fatalf("RenderToHTML error: %v", err)
	}
	return out
}

func TestSSRHostTextAndNesting(t *testing.T) {
	w := ssrBox{tag: "div", children: []Widget{
		ssrBox{tag: "p", text: "hello"},
		ssrBox{tag: "p", text: "world"},
	}}
	got := mustRender(t, w)
	want := "<div><p>hello</p><p>world</p></div>"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestSSRAttrsAndStyleSorted(t *testing.T) {
	w := ssrBox{
		tag:   "div",
		attrs: map[string]string{"id": "x", "class": "c", "data-z": "1"},
		style: map[string]string{"color": "red", "background": "blue"},
	}
	got := mustRender(t, w)
	// attrs sorted: class, data-z, id ; style sorted: background, color
	want := `<div class="c" data-z="1" id="x" style="background: blue; color: red"></div>`
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestSSREscaping(t *testing.T) {
	w := ssrBox{tag: "div", text: `<script>&"`, attrs: map[string]string{"title": `a"b<c`}}
	got := mustRender(t, w)
	if strings.Contains(got, "<script>") {
		t.Fatalf("text not escaped: %q", got)
	}
	if !strings.Contains(got, "&lt;script&gt;") {
		t.Fatalf("expected escaped text, got %q", got)
	}
	if strings.Contains(got, `title="a"b<c"`) {
		t.Fatalf("attr not escaped: %q", got)
	}
}

func TestSSRVoidElement(t *testing.T) {
	w := ssrBox{tag: "img", attrs: map[string]string{"src": "a.png"}}
	got := mustRender(t, w)
	want := `<img src="a.png">`
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestSSRHydrationMarkers(t *testing.T) {
	got := mustRender(t, ssrBtn{})
	if !strings.Contains(got, `data-gutter-h="1"`) {
		t.Fatalf("expected hydration marker for events, got %q", got)
	}
	keyed := mustRender(t, ssrKeyed{k: "row-7"})
	if !strings.Contains(keyed, `data-gutter-key="row-7"`) {
		t.Fatalf("expected key marker, got %q", keyed)
	}
}

func TestSSRStatelessAndStateful(t *testing.T) {
	got := mustRender(t, ssrWrap{child: ssrBox{tag: "b", text: "x"}})
	if got != "<b>x</b>" {
		t.Fatalf("stateless: got %q", got)
	}
	// InitState must run before Build (n: 5 -> +100 = 105)
	got = mustRender(t, ssrCounter{start: 5})
	if got != "<span>n=105</span>" {
		t.Fatalf("stateful: got %q (InitState should have fired)", got)
	}
}

func TestSSRUnsupportedType(t *testing.T) {
	if _, err := RenderToHTML(struct{ X int }{42}); err == nil {
		t.Fatal("expected error for non-widget type")
	}
}
