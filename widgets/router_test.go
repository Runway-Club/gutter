package widgets

import (
	"testing"

	"github.com/Runway-Club/gutter"
)

func TestRouterStripsQueryWhenMatching(t *testing.T) {
	r := NewRouter(map[string]RouteBuilder{
		"/about":    func(RouteParams) gutter.Widget { return Text{Data: "about"} },
		"/user/:id": func(p RouteParams) gutter.Widget { return Text{Data: "user-" + p["id"]} },
	}, Text{Data: "notfound"})

	// Param route with a trailing query must still match and capture :id.
	r.Replace("/user/42?tab=settings")
	if h := hostOf(t, r.match(r.Current())); h.Text != "user-42" {
		t.Errorf("param route with query matched to %q, want user-42", h.Text)
	}

	// Static route with a query.
	r.Replace("/about?ref=home")
	if h := hostOf(t, r.match(r.Current())); h.Text != "about" {
		t.Errorf("static route with query matched to %q, want about", h.Text)
	}
}

func TestRouterQueryParsing(t *testing.T) {
	r := NewRouter(map[string]RouteBuilder{
		"/search": func(RouteParams) gutter.Widget { return Text{Data: "search"} },
	}, Text{Data: "notfound"})

	r.Replace("/search?q=go+lang&page=2")
	q := r.Query()
	if q.Get("q") != "go lang" {
		t.Errorf("q = %q, want %q", q.Get("q"), "go lang")
	}
	if q.Get("page") != "2" {
		t.Errorf("page = %q, want 2", q.Get("page"))
	}

	r.Replace("/search")
	if len(r.Query()) != 0 {
		t.Errorf("expected empty query for path without ?, got %v", r.Query())
	}
}
