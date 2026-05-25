// Package app holds the shared widget tree so BOTH entry points use the same
// Root(): the wasm main (client render) and the host ssrgen (server render).
// This is the Root() convention from ROADMAP.md Phase 1.2.
package app

import (
	"fmt"

	"github.com/Runway-Club/gutter"
	"github.com/Runway-Club/gutter/themes"
	"github.com/Runway-Club/gutter/widgets"
)

// likeButton is a stateful counter used to prove hydration: the server renders
// "Likes: 0"; after the WASM hydrates the existing DOM, clicking increments it
// without the node being recreated.
type likeButton struct{}

func (likeButton) CreateState() gutter.State { return &likeState{} }

type likeState struct {
	gutter.StateObject
	n int
}

func (s *likeState) Build(ctx *gutter.BuildContext) gutter.Widget {
	return widgets.Button{
		Variant:   widgets.ButtonPrimary,
		Label:     fmt.Sprintf("Likes: %d", s.n),
		OnPressed: func() { s.SetState(func() { s.n++ }) },
	}
}

// Root builds a representative "product dashboard": an app bar plus a grid of
// content cards — enough real text/content that First Contentful Paint is
// meaningful — and one interactive counter to demonstrate hydration.
func Root() gutter.Widget {
	cards := make([]gutter.Widget, 0, 24)
	for i := 1; i <= 24; i++ {
		cards = append(cards, widgets.Card{
			Child: widgets.Column{
				Spacing: 6,
				Children: []gutter.Widget{
					widgets.Heading{Level: widgets.H3, Text: fmt.Sprintf("Feature %d", i)},
					widgets.Body{Text: "A representative product card with descriptive body text that paints as content on first render."},
					widgets.Button{Variant: widgets.ButtonPrimary, Label: "Open"},
				},
			},
		})
	}
	return widgets.Scaffold{
		Title: "Gutter SSR demo",
		Theme: themes.Apple,
		AppBar: widgets.AppBar{
			TitleWidget: widgets.Heading{Level: widgets.H2, Text: "Dashboard"},
		},
		Body: widgets.Column{
			Spacing: 16,
			Children: []gutter.Widget{
				likeButton{},
				widgets.Grid{
					MinColumnWidth: "240px",
					Gap:            12,
					Children:       cards,
				},
			},
		},
	}
}
