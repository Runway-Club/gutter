//go:build js && wasm

package widgets

import (
	"fmt"
	"syscall/js"
	"testing"

	"github.com/Runway-Club/gutter"
)

// mountListBuilder mounts lb into a fresh body container and returns the
// viewport node (the ListBuilder's root div). The render is synchronous; the
// scroll listener's viewport correction is a queued microtask that hasn't run
// yet, so assertions see the first-frame (fallback-viewport) window.
func mountListBuilder(t *testing.T, lb ListBuilder) js.Value {
	t.Helper()
	doc := js.Global().Get("document")
	id := fmt.Sprintf("lb-%d", listTestSeq)
	listTestSeq++
	c := doc.Call("createElement", "div")
	c.Call("setAttribute", "id", id)
	doc.Get("body").Call("appendChild", c)
	gutter.MountInto("#"+id, lb)
	return c.Get("firstChild")
}

var listTestSeq int

func TestListBuilderVirtualizesVertical(t *testing.T) {
	vp := mountListBuilder(t, ListBuilder{
		ItemCount:  1000,
		ItemHeight: 20,
		Height:     "100px",
		ItemBuilder: func(i int) gutter.Widget {
			return Text{Data: fmt.Sprintf("row-%d", i)}
		},
	})
	sizer := vp.Get("firstChild")
	if h := sizer.Get("style").Get("height").String(); h != "20000px" {
		t.Fatalf("sizer height = %q, want 20000px (1000*20)", h)
	}
	inner := sizer.Get("firstChild")
	n := inner.Get("children").Get("length").Int()
	if n == 0 || n >= 1000 {
		t.Fatalf("rendered %d slots; expected a small windowed subset of 1000", n)
	}
	if got := inner.Get("firstChild").Get("textContent").String(); got != "row-0" {
		t.Errorf("first rendered item = %q, want row-0", got)
	}
}

func TestListBuilderVariableExtents(t *testing.T) {
	// Extents alternate 10/30 over 100 items → total 50*10 + 50*30 = 2000.
	vp := mountListBuilder(t, ListBuilder{
		ItemCount:  100,
		ItemExtent: func(i int) float64 { return map[bool]float64{true: 10, false: 30}[i%2 == 0] },
		Height:     "120px",
		ItemBuilder: func(i int) gutter.Widget {
			return Text{Data: fmt.Sprintf("v-%d", i)}
		},
	})
	sizer := vp.Get("firstChild")
	if h := sizer.Get("style").Get("height").String(); h != "2000px" {
		t.Fatalf("variable sizer height = %q, want 2000px", h)
	}
	// First slot is index 0 with extent 10px.
	inner := sizer.Get("firstChild")
	slot0 := inner.Get("firstChild")
	if h := slot0.Get("style").Get("height").String(); h != "10px" {
		t.Errorf("slot 0 height = %q, want 10px", h)
	}
}

func TestListBuilderHorizontal(t *testing.T) {
	vp := mountListBuilder(t, ListBuilder{
		ItemCount:  500,
		ItemHeight: 40, // width along the scroll axis
		Direction:  ListHorizontal,
		Width:      "200px",
		ItemBuilder: func(i int) gutter.Widget {
			return Text{Data: fmt.Sprintf("col-%d", i)}
		},
	})
	if ox := vp.Get("style").Get("overflowX").String(); ox != "auto" {
		t.Errorf("viewport overflow-x = %q, want auto", ox)
	}
	sizer := vp.Get("firstChild")
	if w := sizer.Get("style").Get("width").String(); w != "20000px" {
		t.Fatalf("horizontal sizer width = %q, want 20000px (500*40)", w)
	}
	inner := sizer.Get("firstChild")
	if fd := inner.Get("style").Get("flexDirection").String(); fd != "row" {
		t.Errorf("inner flex-direction = %q, want row", fd)
	}
	slot0 := inner.Get("firstChild")
	if w := slot0.Get("style").Get("width").String(); w != "40px" {
		t.Errorf("horizontal slot 0 width = %q, want 40px", w)
	}
}
