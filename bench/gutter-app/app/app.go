// Package app holds the benchmark grid widget tree, host-safe (no syscall/js)
// so it can be both compiled to WASM (client) and rendered to HTML (SSR).
package app

import (
	"fmt"
	"strconv"

	"github.com/Runway-Club/gutter"
	"github.com/Runway-Club/gutter/widgets"
)

// Root builds the grid of n independent stateful buttons.
func Root(n int) gutter.Widget { return grid{n: n} }

type grid struct{ n int }

func (g grid) Build(ctx *gutter.BuildContext) gutter.Widget {
	items := make([]gutter.Widget, g.n)
	for i := 0; i < g.n; i++ {
		items[i] = benchItem{i: i}
	}
	return widgets.Styled{
		Tag:   "div",
		Attrs: map[string]string{"id": "grid"},
		Style: map[string]string{
			"display":   "flex",
			"flex-wrap": "wrap",
			"gap":       "4px",
			"padding":   "8px",
		},
		Children: items,
	}
}

type benchItem struct{ i int }

func (b benchItem) CreateState() gutter.State { return &itemState{i: b.i} }

type itemState struct {
	gutter.StateObject
	i int
	c int
}

func (s *itemState) Build(ctx *gutter.BuildContext) gutter.Widget {
	return widgets.Styled{
		Tag:   "button",
		Text:  fmt.Sprintf("Item %d — %d", s.i, s.c),
		Attrs: map[string]string{"data-bench-item": strconv.Itoa(s.i)},
		Style: map[string]string{"padding": "4px 8px"},
		Events: map[string]func(gutter.Event){
			"click": func(gutter.Event) { s.SetState(func() { s.c++ }) },
		},
	}
}
