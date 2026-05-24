// Router demo — exercises Router + ObserverBuilder + AsyncBuilder together.
//
// Routes:
//
//	/         Home — shared counter (Notifier+ObserverBuilder demo)
//	/about    Static page
//	/user/:id User detail — exercises :param capture
//	/slow     AsyncBuilder demo (simulated 1s fetch)
//
// The NavBar lives in the Scaffold AppBar so navigation works from any route.
package main

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/Runway-Club/gutter"
	"github.com/Runway-Club/gutter/themes"
	"github.com/Runway-Club/gutter/widgets"
)

// One Notifier shared by Home + the AppBar badge proves cross-route state
// works without prop-drilling — both subtrees subscribe via ObserverBuilder.
var count = gutter.NewNotifier(0)

func main() {
	router := widgets.NewRouter(map[string]widgets.RouteBuilder{
		"/":         func(widgets.RouteParams) gutter.Widget { return Home{} },
		"/about":    func(widgets.RouteParams) gutter.Widget { return About{} },
		"/slow":     func(widgets.RouteParams) gutter.Widget { return Slow{} },
		"/user/:id": func(p widgets.RouteParams) gutter.Widget { return User{ID: p["id"]} },
	}, NotFound{})

	gutter.RunApp(Shell{Router: router})
}

// Shell wraps every route in a Scaffold with the NavBar so the chrome
// survives navigations (only the body subtree rebuilds).
type Shell struct {
	Router *widgets.Router
}

func (s Shell) Build(ctx *gutter.BuildContext) gutter.Widget {
	return widgets.Scaffold{
		Title: "Gutter Router Demo",
		Theme: themes.Apple,
		AppBar: widgets.AppBar{
			TitleWidget: widgets.Text{Data: "Router Demo", Style: &widgets.TextStyle{FontSize: "20px"}},
			Actions:     navActions(s.Router),
		},
		Body: widgets.Surface{
			Variant: widgets.SurfaceAlt,
			Child: widgets.Padding{
				Padding: widgets.EdgeInsetsAll(24),
				Child:   widgets.RouterView{Router: s.Router},
			},
		},
	}
}

func navActions(r *widgets.Router) []gutter.Widget {
	link := func(label, path string) gutter.Widget {
		return widgets.Button{
			Variant:   widgets.ButtonGhost,
			Label:     label,
			OnPressed: func() { r.Push(path) },
		}
	}
	return []gutter.Widget{
		link("Home", "/"),
		link("About", "/about"),
		link("User 42", "/user/42"),
		link("Slow", "/slow"),
		// Live counter in the chrome — observes the same Notifier as Home,
		// proving cross-route reactive updates.
		widgets.ObserverBuilder[int]{
			Source: count,
			Builder: func(_ *gutter.BuildContext, n int) gutter.Widget {
				return widgets.Badge{Text: fmt.Sprintf("count: %d", n)}
			},
		},
	}
}

// Home — reads the shared Notifier via ObserverBuilder. Buttons mutate the
// Notifier; the AppBar badge updates without Home telling it to.
type Home struct{}

func (Home) Build(ctx *gutter.BuildContext) gutter.Widget {
	return widgets.Card{
		Variant: widgets.CardFeature,
		Child: widgets.Column{
			CrossAxisAlign: widgets.CrossAxisCenter,
			Spacing:        16,
			Children: []gutter.Widget{
				widgets.Heading{Level: widgets.H1, Text: "Home"},
				widgets.Body{Text: "The badge in the AppBar observes the same Notifier as the buttons below."},
				widgets.ObserverBuilder[int]{
					Source: count,
					Builder: func(_ *gutter.BuildContext, n int) gutter.Widget {
						return widgets.Heading{Level: widgets.H2, Text: fmt.Sprintf("count = %d", n)}
					},
				},
				widgets.Row{
					Spacing: 8,
					Children: []gutter.Widget{
						widgets.Button{
							Variant:   widgets.ButtonPrimary,
							Label:     "−",
							OnPressed: func() { count.Update(func(n int) int { return n - 1 }) },
						},
						widgets.Button{
							Variant:   widgets.ButtonPrimary,
							Label:     "+",
							OnPressed: func() { count.Update(func(n int) int { return n + 1 }) },
						},
					},
				},
			},
		},
	}
}

type About struct{}

func (About) Build(ctx *gutter.BuildContext) gutter.Widget {
	return widgets.Card{
		Child: widgets.Column{
			Spacing: 12,
			Children: []gutter.Widget{
				widgets.Heading{Level: widgets.H1, Text: "About"},
				widgets.Body{Text: "Static route. Click Home to return — browser back/forward also works."},
			},
		},
	}
}

// User — exercises :param capture. Switching from /user/42 to /user/99
// re-runs the route builder with the new param; AsyncBuilder is wrapped in
// WithKey on ID so the simulated fetch re-runs on each ID change.
type User struct{ ID string }

func (u User) Build(ctx *gutter.BuildContext) gutter.Widget {
	return widgets.Card{
		Variant: widgets.CardFeature,
		Child: widgets.Column{
			Spacing: 12,
			Children: []gutter.Widget{
				widgets.Heading{Level: widgets.H1, Text: "User " + u.ID},
				widgets.WithKey{
					Key: u.ID,
					Child: widgets.AsyncBuilder[string]{
						Load: func(ctx context.Context) (string, error) {
							return fakeFetchUser(ctx, u.ID)
						},
						Builder: func(_ *gutter.BuildContext, snap widgets.AsyncSnapshot[string]) gutter.Widget {
							switch snap.State {
							case widgets.AsyncPending:
								return widgets.Body{Text: "Loading user " + u.ID + "…"}
							case widgets.AsyncFailed:
								return widgets.Body{Text: "Error: " + snap.Error.Error()}
							}
							return widgets.Body{Text: snap.Data}
						},
					},
				},
			},
		},
	}
}

type Slow struct{}

func (Slow) Build(ctx *gutter.BuildContext) gutter.Widget {
	return widgets.Card{
		Child: widgets.Column{
			Spacing: 12,
			Children: []gutter.Widget{
				widgets.Heading{Level: widgets.H1, Text: "Slow"},
				widgets.AsyncBuilder[string]{
					Load: func(ctx context.Context) (string, error) {
						select {
						case <-time.After(1 * time.Second):
							return "Done after 1 second", nil
						case <-ctx.Done():
							return "", ctx.Err()
						}
					},
					Builder: func(_ *gutter.BuildContext, snap widgets.AsyncSnapshot[string]) gutter.Widget {
						switch snap.State {
						case widgets.AsyncPending:
							return widgets.Body{Text: "Fetching… (navigate away to cancel)"}
						case widgets.AsyncFailed:
							return widgets.Body{Text: "Canceled: " + snap.Error.Error()}
						}
						return widgets.Body{Text: snap.Data}
					},
				},
			},
		},
	}
}

type NotFound struct{}

func (NotFound) Build(ctx *gutter.BuildContext) gutter.Widget {
	return widgets.Card{
		Child: widgets.Column{
			Spacing: 12,
			Children: []gutter.Widget{
				widgets.Heading{Level: widgets.H1, Text: "404"},
				widgets.Body{Text: "No route matched."},
			},
		},
	}
}

// fakeFetchUser simulates a backend call. Returns an error for IDs that
// don't parse as integers, so the AsyncFailed branch has something to show.
func fakeFetchUser(ctx context.Context, id string) (string, error) {
	select {
	case <-time.After(400 * time.Millisecond):
	case <-ctx.Done():
		return "", ctx.Err()
	}
	n, err := strconv.Atoi(id)
	if err != nil {
		return "", errors.New("user id must be numeric")
	}
	return fmt.Sprintf("User #%d — Alice the %dth", n, n), nil
}
