//go:build js && wasm

package gutter

import (
	"syscall/js"
	"testing"
)

func TestInspectWalksMountedTree(t *testing.T) {
	doc := js.Global().Get("document")
	c := doc.Call("createElement", "div")
	c.Call("setAttribute", "id", "inspect-host")
	doc.Get("body").Call("appendChild", c)

	MountInto("#inspect-host", testHost{tag: "section", children: []Widget{
		batchWidget{}, // a StatefulWidget
		testHost{tag: "span", text: "x"},
	}})

	trees := Inspect()
	if len(trees) == 0 {
		t.Fatal("Inspect returned no roots")
	}
	root := trees[len(trees)-1] // the one we just mounted
	if root.Kind != "host" || root.Tag != "section" {
		t.Fatalf("root = %+v, want host <section>", root)
	}
	if len(root.Children) != 2 {
		t.Fatalf("root has %d children, want 2", len(root.Children))
	}
	if root.Children[0].Kind != "stateful" {
		t.Errorf("child 0 kind = %q, want stateful", root.Children[0].Kind)
	}
	if root.Children[1].Tag != "span" {
		t.Errorf("child 1 tag = %q, want span", root.Children[1].Tag)
	}
}

func TestEnableDevtoolsTogglesPanel(t *testing.T) {
	EnableDevtools()
	doc := js.Global().Get("document")

	init := js.Global().Get("Object").New()
	init.Set("key", "g")
	init.Set("ctrlKey", true)
	init.Set("shiftKey", true)
	init.Set("bubbles", true)
	ev := js.Global().Get("KeyboardEvent").New("keydown", init)
	doc.Call("dispatchEvent", ev)

	panel := doc.Call("getElementById", "gutter-devtools")
	if panel.IsNull() || panel.IsUndefined() {
		t.Fatal("devtools panel not created on Ctrl+Shift+G")
	}
	if panel.Get("style").Get("display").String() != "block" {
		t.Errorf("panel display = %q, want block after first toggle", panel.Get("style").Get("display").String())
	}
	if panel.Get("textContent").String() == "" {
		t.Error("panel should show the inspected tree text")
	}
}
