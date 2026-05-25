//go:build js && wasm

package gutter

import (
	"strconv"
	"syscall/js"
	"testing"
)

// diCounter reads the provided *diSvc on every Build and shows it next to a
// counter, so we can prove DependOn still resolves after an ISOLATED rebuild
// (a SetState that does not re-run the ancestor Provider top-down).
type diCounter struct{}

func (diCounter) CreateState() State { return &diCounterState{} }

type diCounterState struct {
	StateObject
	n int
}

func (s *diCounterState) Build(ctx *BuildContext) Widget {
	name := "none"
	if svc, ok := DependOn[*diSvc](ctx); ok {
		name = svc.name
	}
	return testHost{
		tag:    "button",
		text:   name + ":" + strconv.Itoa(s.n),
		events: map[string]func(Event){"click": func(Event) { s.SetState(func() { s.n++ }) }},
	}
}

func TestProviderSurvivesIsolatedRebuild(t *testing.T) {
	parent := freshParent()
	tree := Provider[*diSvc]{Value: &diSvc{name: "live"}, Child: diCounter{}}
	el := newElement(tree)
	el.mount(parent, js.Null(), &BuildContext{})

	btn := parent.Get("firstChild")
	if got := btn.Get("textContent").String(); got != "live:0" {
		t.Fatalf("after mount = %q, want live:0", got)
	}

	// Isolated SetState rebuild: the Provider is NOT re-run top-down. DependOn
	// must still resolve via the element's captured scope.
	btn.Call("click")
	flushRebuilds()
	if got := btn.Get("textContent").String(); got != "live:1" {
		t.Fatalf("after isolated rebuild = %q, want live:1 (Provider value lost on rebuild)", got)
	}
}
