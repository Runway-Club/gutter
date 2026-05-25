//go:build js && wasm

// Islands: two independent Gutter widget trees embedded in an otherwise static
// HTML page. The page's loader (see index.html) only fetches app.wasm once an
// island scrolls near the viewport, so the static content paints with zero
// WASM cost; MountInto then mounts each island into its placeholder.
package main

import (
	"fmt"

	"github.com/Runway-Club/gutter"
	"github.com/Runway-Club/gutter/themes"
	"github.com/Runway-Club/gutter/widgets"
)

type counter struct{ verb string }

func (c counter) CreateState() gutter.State { return &counterState{verb: c.verb} }

type counterState struct {
	gutter.StateObject
	verb string
	n    int
}

func (s *counterState) Build(ctx *gutter.BuildContext) gutter.Widget {
	return widgets.Button{
		Variant:   widgets.ButtonPrimary,
		Label:     fmt.Sprintf("%s (%d)", s.verb, s.n),
		OnPressed: func() { s.SetState(func() { s.n++ }) },
	}
}

func main() {
	// Each island is independent; WithHydrate adopts SSR markup if the
	// placeholder has any, else mounts fresh.
	gutter.MountInto("#island-counter", counter{verb: "Add to cart"}, gutter.WithTheme(themes.Apple), gutter.WithHydrate())
	gutter.MountInto("#island-likes", counter{verb: "Like"}, gutter.WithTheme(themes.Apple), gutter.WithHydrate())
	select {} // keep the WASM runtime alive for both islands' callbacks
}
