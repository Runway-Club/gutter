//go:build js && wasm

package gutter

import (
	"strconv"
	"syscall/js"
	"testing"
)

type islandCounter struct{ label string }

func (c islandCounter) CreateState() State { return &islandCounterState{label: c.label} }

type islandCounterState struct {
	StateObject
	label string
	n     int
}

func (s *islandCounterState) Build(*BuildContext) Widget {
	return testHost{
		tag:    "button",
		text:   s.label + ":" + strconv.Itoa(s.n),
		events: map[string]func(Event){"click": func(Event) { s.SetState(func() { s.n++ }) }},
	}
}

func appendIsland(id string) (el, body js.Value) {
	body = js.Global().Get("document").Get("body")
	el = js.Global().Get("document").Call("createElement", "div")
	el.Set("id", id)
	body.Call("appendChild", el)
	return el, body
}

func TestMountIntoIndependentIslands(t *testing.T) {
	a, body := appendIsland("isl-a")
	b, _ := appendIsland("isl-b")
	defer func() { body.Call("removeChild", a); body.Call("removeChild", b) }()

	MountInto("#isl-a", islandCounter{label: "A"})
	MountInto("#isl-b", islandCounter{label: "B"})

	btnA := a.Get("firstChild")
	btnB := b.Get("firstChild")
	if btnA.Get("textContent").String() != "A:0" || btnB.Get("textContent").String() != "B:0" {
		t.Fatalf("mount: A=%q B=%q", btnA.Get("textContent").String(), btnB.Get("textContent").String())
	}

	// Clicking island A must not affect island B — they are independent trees.
	btnA.Call("click")
	flushRebuilds()
	if got := btnA.Get("textContent").String(); got != "A:1" {
		t.Fatalf("A after click = %q, want A:1", got)
	}
	if got := btnB.Get("textContent").String(); got != "B:0" {
		t.Fatalf("island B leaked to %q, want B:0", got)
	}
}

func TestMountIntoHydratesIsland(t *testing.T) {
	c, body := appendIsland("isl-c")
	defer func() { body.Call("removeChild", c) }()

	html, err := RenderToHTML(islandCounter{label: "C"})
	if err != nil {
		t.Fatal(err)
	}
	c.Set("innerHTML", html)
	server := c.Get("firstChild")

	MountInto("#isl-c", islandCounter{label: "C"}, WithHydrate())

	if !c.Get("firstChild").Equal(server) {
		t.Fatal("island was re-rendered instead of hydrated (node replaced)")
	}
	server.Call("click")
	flushRebuilds()
	if got := server.Get("textContent").String(); got != "C:1" {
		t.Fatalf("hydrated island click = %q, want C:1", got)
	}
}
