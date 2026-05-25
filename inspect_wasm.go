//go:build js && wasm

package gutter

import (
	"fmt"
	"reflect"
	"strings"
	"syscall/js"
)

// mountedApps tracks every root mounted by MountInto so Inspect/EnableDevtools
// can walk the live element trees. It only grows (apps rarely unmount), which
// is fine for a debug facility.
var mountedApps []*App

func registerApp(a *App) { mountedApps = append(mountedApps, a) }

// Inspect returns the live element tree of every mounted app root as a plain
// data tree, for devtools and debugging. Empty before anything mounts.
func Inspect() []InspectNode {
	out := make([]InspectNode, 0, len(mountedApps))
	for _, a := range mountedApps {
		if a != nil && a.root != nil {
			out = append(out, inspectElement(a.root))
		}
	}
	return out
}

func inspectElement(e Element) InspectNode {
	switch x := e.(type) {
	case *hostElement:
		n := InspectNode{Kind: "host", Type: typeNameOf(x.wt), Tag: normTag(x.host), Key: keyString(x.key())}
		for _, c := range x.children {
			n.Children = append(n.Children, inspectElement(c))
		}
		return n
	case *statelessElement:
		n := InspectNode{Kind: "stateless", Type: typeNameOf(x.wt), Key: keyString(x.key())}
		if x.child != nil {
			n.Children = append(n.Children, inspectElement(x.child))
		}
		return n
	case *statefulElement:
		n := InspectNode{Kind: "stateful", Type: typeNameOf(x.wt), Key: keyString(x.key())}
		if x.child != nil {
			n.Children = append(n.Children, inspectElement(x.child))
		}
		return n
	case *portalElement:
		n := InspectNode{Kind: "portal", Type: typeNameOf(x.wt)}
		if x.child != nil {
			n.Children = append(n.Children, inspectElement(x.child))
		}
		return n
	}
	return InspectNode{Kind: "unknown"}
}

func typeNameOf(t reflect.Type) string {
	if t == nil {
		return ""
	}
	return t.String()
}

func keyString(k any) string {
	if k == nil {
		return ""
	}
	return fmt.Sprint(k)
}

// EnableDevtools installs a Ctrl+Shift+G keydown toggle that overlays the live
// element tree (Inspect rendered as text) in a fixed panel. Call it once after
// mounting (e.g. behind a build flag or env check) to inspect structure, types,
// tags, and keys without external tooling.
func EnableDevtools() {
	doc := js.Global().Get("document")
	var panel js.Value
	visible := false

	render := func() {
		var b strings.Builder
		for _, n := range Inspect() {
			b.WriteString(n.String())
		}
		if b.Len() == 0 {
			b.WriteString("(no mounted roots)")
		}
		panel.Set("textContent", b.String())
	}

	cb := js.FuncOf(func(_ js.Value, args []js.Value) any {
		ev := args[0]
		if !ev.Get("ctrlKey").Bool() || !ev.Get("shiftKey").Bool() || !strings.EqualFold(ev.Get("key").String(), "g") {
			return nil
		}
		if panel.IsUndefined() || panel.IsNull() {
			panel = doc.Call("createElement", "pre")
			panel.Call("setAttribute", "id", "gutter-devtools")
			panel.Get("style").Set("cssText",
				"position:fixed;top:8px;right:8px;max-width:42vw;max-height:90vh;overflow:auto;"+
					"z-index:2147483647;background:rgba(0,0,0,.85);color:#0f0;"+
					"font:11px/1.4 ui-monospace,monospace;padding:8px;margin:0;border-radius:8px;white-space:pre")
			doc.Get("body").Call("appendChild", panel)
		}
		visible = !visible
		if visible {
			render()
			panel.Get("style").Set("display", "block")
		} else {
			panel.Get("style").Set("display", "none")
		}
		return nil
	})
	doc.Call("addEventListener", "keydown", cb)
}
