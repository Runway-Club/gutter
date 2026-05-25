// Command testapp is a deterministic gutter app whose only purpose is to give
// the Playwright end-to-end suite stable surfaces to drive. Every interactive
// element is reachable by a stable selector (a data-testid attribute or a
// known placeholder/label) so the specs in ../tests don't depend on layout.
//
// Build + serve via the gutter CLI (see ../serve.sh); the specs talk to it on
// http://localhost:8080.
package main

import (
	"strconv"

	"github.com/Runway-Club/gutter"
	"github.com/Runway-Club/gutter/widgets"
)

func main() {
	gutter.RunApp(App{})
}

// App is the root StatefulWidget. One state object holds every surface so the
// specs can exercise SetState, controlled inputs, keyed reordering, and
// conditional mount/unmount through a single tree.
type App struct{}

func (App) CreateState() gutter.State { return &appState{} }

type appState struct {
	gutter.StateObject
	count      int
	echo       string
	order      []string
	itemValues map[string]string
	dialogOpen bool
}

func (s *appState) InitState() {
	s.order = []string{"A", "B", "C"}
	s.itemValues = map[string]string{}
}

// testID wraps a child in a div carrying a data-testid attribute so Playwright
// can select it. Uses display:contents so it doesn't perturb layout.
func testID(id string, child gutter.Widget) gutter.Widget {
	return widgets.Styled{
		Attrs:    map[string]string{"data-testid": id},
		Style:    map[string]string{"display": "contents"},
		Children: []gutter.Widget{child},
	}
}

func (s *appState) Build(ctx *gutter.BuildContext) gutter.Widget {
	return widgets.Scaffold{
		Title: "Gutter E2E",
		Body: widgets.Padding{
			Padding: widgets.EdgeInsetsAll(24),
			Child: widgets.Column{
				Spacing: 16,
				Children: []gutter.Widget{
					widgets.Heading{Level: widgets.H3, Text: "Gutter E2E"},
					s.counterSection(),
					s.echoSection(),
					s.keyedListSection(),
					s.dialogSection(),
				},
			},
		},
	}
}

// counterSection: a count display plus an increment and a "burst" button. Burst
// calls SetState five times in one turn — correctness here proves batched
// SetState applies every mutation while coalescing the rebuilds.
func (s *appState) counterSection() gutter.Widget {
	return widgets.Row{
		Spacing: 12,
		Children: []gutter.Widget{
			testID("count", widgets.Body{Text: "count: " + strconv.Itoa(s.count)}),
			widgets.Button{Label: "increment", OnPressed: func() {
				s.SetState(func() { s.count++ })
			}},
			widgets.Button{Label: "burst", OnPressed: func() {
				for range 5 {
					s.SetState(func() { s.count++ })
				}
			}},
		},
	}
}

// echoSection: a controlled input mirrored into a label. Tests caret-preserving
// controlled-input sync under batched rebuilds.
func (s *appState) echoSection() gutter.Widget {
	return widgets.Row{
		Spacing: 12,
		Children: []gutter.Widget{
			widgets.Input{
				Placeholder: "echo",
				Value:       s.echo,
				OnChanged:   func(v string) { s.SetState(func() { s.echo = v }) },
			},
			testID("echo", widgets.Body{Text: s.echo}),
		},
	}
}

// keyedListSection: keyed rows, each with its own input, plus a reverse button.
// Reversing must move the existing DOM nodes (keyed reconcile), so an input's
// typed value and focus survive the reorder.
func (s *appState) keyedListSection() gutter.Widget {
	rows := make([]gutter.Widget, 0, len(s.order)+1)
	for _, label := range s.order {
		rows = append(rows, widgets.WithKey{Key: label, Child: widgets.Row{
			Spacing: 8,
			Children: []gutter.Widget{
				widgets.Body{Text: label},
				widgets.Input{
					Placeholder: "input-" + label,
					Value:       s.itemValues[label],
					OnChanged:   func(v string) { s.SetState(func() { s.itemValues[label] = v }) },
				},
			},
		}})
	}
	rows = append(rows, widgets.Button{Label: "reverse", OnPressed: func() {
		s.SetState(func() {
			for i, j := 0, len(s.order)-1; i < j; i, j = i+1, j-1 {
				s.order[i], s.order[j] = s.order[j], s.order[i]
			}
		})
	}})
	return testID("keyed-list", widgets.Column{Spacing: 8, Children: rows})
}

// dialogSection: a conditionally-mounted panel. Toggling exercises reconcile
// mount/unmount of a subtree.
func (s *appState) dialogSection() gutter.Widget {
	children := []gutter.Widget{
		widgets.Button{Label: "open dialog", OnPressed: func() {
			s.SetState(func() { s.dialogOpen = true })
		}},
	}
	if s.dialogOpen {
		children = append(children, testID("dialog", widgets.Card{
			Child: widgets.Column{
				Spacing: 8,
				Children: []gutter.Widget{
					widgets.Body{Text: "Dialog is open"},
					widgets.Button{Label: "close dialog", OnPressed: func() {
						s.SetState(func() { s.dialogOpen = false })
					}},
				},
			},
		}))
	}
	return widgets.Column{Spacing: 8, Children: children}
}
