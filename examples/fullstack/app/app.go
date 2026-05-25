// Package app is the shared UI: Root() is built by the wasm client (interactive)
// and by the host server (SSR). The button calls the Go server over typed RPC.
package app

import (
	"context"
	"fmt"

	"fullstackexample/api"

	"github.com/Runway-Club/gutter"
	"github.com/Runway-Club/gutter/rpc"
	"github.com/Runway-Club/gutter/themes"
	"github.com/Runway-Club/gutter/widgets"
)

func Root() gutter.Widget { return adder{} }

type adder struct{}

func (adder) CreateState() gutter.State { return &adderState{} }

type adderState struct {
	gutter.StateObject
	result string
	busy   bool
}

func (s *adderState) Build(ctx *gutter.BuildContext) gutter.Widget {
	label := "Compute 2 + 40 on the server"
	if s.busy {
		label = "…"
	}
	result := s.result
	if result == "" {
		result = "(not called yet)"
	}
	return widgets.Scaffold{
		Title: "Gutter full-stack",
		Theme: themes.Apple,
		AppBar: widgets.AppBar{
			TitleWidget: widgets.Heading{Level: widgets.H2, Text: "Full-stack Go"},
		},
		Body: widgets.Center{Child: widgets.Card{Variant: widgets.CardFeature, Child: widgets.Column{
			CrossAxisAlign: widgets.CrossAxisCenter,
			Spacing:        12,
			Children: []gutter.Widget{
				widgets.Heading{Level: widgets.H3, Text: "Typed RPC demo"},
				widgets.Body{Text: "Click to call the Go server with a struct shared by both sides."},
				widgets.Button{Variant: widgets.ButtonPrimary, Label: label, OnPressed: s.compute},
				widgets.Body{Text: "Result: " + result},
			},
		}}},
	}
}

// compute calls the server. rpc.Call blocks the goroutine on the fetch, so it
// runs in its own goroutine and SetStates the result back in.
func (s *adderState) compute() {
	if s.busy {
		return
	}
	s.SetState(func() { s.busy = true })
	go func() {
		res, err := rpc.Call[api.AddRequest, api.AddResponse](context.Background(), api.AddRequest{A: 2, B: 40})
		s.SetState(func() {
			s.busy = false
			if err != nil {
				s.result = "error: " + err.Error()
			} else {
				s.result = fmt.Sprintf("%d", res.Sum)
			}
		})
	}()
}
