package main

import (
	"fmt"
	"math"
	"strconv"

	"github.com/Runway-Club/gutter"
	"github.com/Runway-Club/gutter/widgets"
)

// PlaygroundApp exercises three new widgets:
//
//   - widgets.Canvas         — a 2D drawable surface
//   - widgets.GestureDetector — pointer events for child widgets
//   - widgets.Worker          — heavy work in a Web Worker (written in Go)
//
// Drag on the canvas to draw points. Press a button to ask a Go
// worker — defined inline below via gutter.NewWorkerTask — to count
// primes up to N without freezing the UI. No separate worker file, no
// JavaScript anywhere in this example.
type PlaygroundApp struct{}

// primesTask runs inside a Web Worker. The handler body is plain Go in
// this binary; gutter.RunApp dispatches to it when the worker thread
// reloads app.wasm. countPrimes is defined below — workers share the
// app's package so helpers are just function calls.
var primesTask = gutter.NewWorkerTask("primes", func(msg string) string {
	n, err := strconv.Atoi(msg)
	if err != nil || n < 0 {
		return `{"error":"expected a non-negative integer"}`
	}
	return fmt.Sprintf(`{"n":%d,"primes":%d}`, n, countPrimes(n))
})

func countPrimes(n int) int {
	if n < 2 {
		return 0
	}
	count := 0
	for i := 2; i <= n; i++ {
		if isPrime(i) {
			count++
		}
	}
	return count
}

func isPrime(n int) bool {
	if n < 2 {
		return false
	}
	for d := 2; d*d <= n; d++ {
		if n%d == 0 {
			return false
		}
	}
	return true
}

func (PlaygroundApp) CreateState() gutter.State { return &playgroundState{} }

type playgroundState struct {
	gutter.StateObject
	points  []point
	drawing bool
}

type point struct {
	x, y float64
}

func (s *playgroundState) Build(ctx *gutter.BuildContext) gutter.Widget {
	return widgets.Scaffold{
		Title: "Gutter Playground",
		AppBar: widgets.AppBar{
			TitleWidget: widgets.Text{
				Data:  "Playground",
				Style: &widgets.TextStyle{FontSize: "20px", FontWeight: "600"},
			},
			Actions: []gutter.Widget{
				widgets.Button{
					Variant: widgets.ButtonGhost,
					Label:   "Clear",
					OnPressed: func() {
						s.SetState(func() { s.points = nil })
					},
				},
			},
		},
		Body: widgets.Surface{
			Variant: widgets.SurfaceAlt,
			Child: widgets.Center{
				Child: widgets.Column{
					Spacing:        24,
					CrossAxisAlign: widgets.CrossAxisCenter,
					Children: []gutter.Widget{
						widgets.Heading{Level: widgets.H2, Text: "Canvas + Gestures + Worker"},
						widgets.Body{
							Text:  "Drag inside the box to paint. Trigger a worker to count primes without blocking the UI.",
							Small: true,
						},
						s.canvasCard(),
						s.workerCard(),
					},
				},
			},
		},
	}
}

func (s *playgroundState) canvasCard() gutter.Widget {
	return widgets.Card{
		Variant: widgets.CardFeature,
		Child: widgets.GestureDetector{
			OnPointerDown: func(e gutter.Event) {
				s.SetState(func() {
					s.drawing = true
					s.points = append(s.points, point{e.OffsetX, e.OffsetY})
				})
			},
			OnPointerMove: func(e gutter.Event) {
				if !s.drawing {
					return
				}
				s.SetState(func() {
					s.points = append(s.points, point{e.OffsetX, e.OffsetY})
				})
			},
			OnPointerUp: func(e gutter.Event) {
				s.SetState(func() { s.drawing = false })
			},
			Child: widgets.Canvas{
				Width:      400,
				Height:     280,
				Background: "#ffffff",
				Paint:      s.paint,
			},
		},
	}
}

func (s *playgroundState) paint(p *widgets.CanvasPainter) {
	w, h := p.Size()
	p.Clear()
	// 20px grid for visual reference.
	p.StrokeStyle("#e8e8ec")
	p.LineWidth(1)
	for x := 0.0; x < w; x += 20 {
		p.BeginPath()
		p.MoveTo(x, 0)
		p.LineTo(x, h)
		p.Stroke()
	}
	for y := 0.0; y < h; y += 20 {
		p.BeginPath()
		p.MoveTo(0, y)
		p.LineTo(w, y)
		p.Stroke()
	}
	// Connect dragged points with a stroked path, then dot the endpoints.
	if len(s.points) > 1 {
		p.StrokeStyle("#3478f6")
		p.LineWidth(3)
		p.LineCap("round")
		p.LineJoin("round")
		p.BeginPath()
		p.MoveTo(s.points[0].x, s.points[0].y)
		for _, pt := range s.points[1:] {
			p.LineTo(pt.x, pt.y)
		}
		p.Stroke()
	}
	p.FillStyle("#0a5cd6")
	for _, pt := range s.points {
		p.BeginPath()
		p.Arc(pt.x, pt.y, 3, 0, 2*math.Pi)
		p.Fill()
	}
}

func (s *playgroundState) workerCard() gutter.Widget {
	return widgets.Card{
		Variant: widgets.CardFeature,
		Child: widgets.Worker{
			Task:    primesTask,
			Message: "200000",
			Builder: func(snap widgets.WorkerSnapshot) gutter.Widget {
				status := "Counting primes up to 200,000…"
				if snap.Error != "" {
					status = "Error: " + snap.Error
				} else if !snap.Pending && snap.Message != "" {
					status = snap.Message
				}
				return widgets.Column{
					Spacing:        12,
					CrossAxisAlign: widgets.CrossAxisCenter,
					Children: []gutter.Widget{
						widgets.Heading{Level: widgets.H3, Text: "Heavy task in a Worker"},
						widgets.Body{Text: status},
						widgets.Row{
							Spacing: 8,
							Children: []gutter.Widget{
								widgets.Button{
									Variant: widgets.ButtonPrimary,
									Label:   "Count to 100k",
									OnPressed: func() {
										snap.Post("100000")
									},
								},
								widgets.Button{
									Variant: widgets.ButtonSecondary,
									Label:   "Count to 500k",
									OnPressed: func() {
										snap.Post("500000")
									},
								},
								widgets.Button{
									Variant: widgets.ButtonGhost,
									Label:   fmt.Sprintf("Points drawn: %d", len(s.points)),
								},
							},
						},
					},
				}
			},
		},
	}
}

func main() {
	gutter.RunApp(PlaygroundApp{})
}
