package main

import (
	"fmt"

	"github.com/Runway-Club/gutter"
	"github.com/Runway-Club/gutter/themes"
	"github.com/Runway-Club/gutter/widgets"
)

type CounterApp struct{}

func (CounterApp) CreateState() gutter.State { return &counterState{} }

type counterState struct {
	gutter.StateObject
	count int
}

func (s *counterState) Build(ctx *gutter.BuildContext) gutter.Widget {
	return widgets.Scaffold{
		Title: "Gutter Counter",
		Theme: themes.Meta,
		AppBar: widgets.AppBar{
			TitleWidget: widgets.Text{Data: "Counter", Style: &widgets.TextStyle{FontSize: "20px"}},
			Actions: []gutter.Widget{
				widgets.Button{
					Variant: widgets.ButtonGhost,
					Label:   "Reset",
					OnPressed: func() {
						s.SetState(func() { s.count = 0 })
					},
				},
			},
		},
		Body: widgets.Surface{
			Variant: widgets.SurfaceAlt,
			Child: widgets.Center{
				Child: widgets.Card{
					Variant: widgets.CardFeature,
					Child: widgets.Column{
						CrossAxisAlign: widgets.CrossAxisCenter,
						Spacing:        16,
						Children: []gutter.Widget{
							widgets.Heading{Level: widgets.H2, Text: fmt.Sprintf("Count: %d", s.count)},
							widgets.Body{Text: "Tap the buttons. No CSS in this file.", Small: true},
							widgets.Row{
								Spacing: 8,
								Children: []gutter.Widget{
									widgets.Button{
										Variant:   widgets.ButtonPrimary,
										Label:     "−",
										OnPressed: func() { s.SetState(func() { s.count-- }) },
									},
									widgets.Button{
										Variant:   widgets.ButtonPrimary,
										Label:     "+",
										OnPressed: func() { s.SetState(func() { s.count++ }) },
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func main() {
	// Scaffold drives the theme — no WithTheme needed at RunApp.
	gutter.RunApp(CounterApp{})
}
