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

func TestRouterGuardRedirectsPush(t *testing.T) {
	authed := false
	guard := func(to string) string {
		if to == "/dashboard" && !authed {
			return "/login"
		}
		return to
	}
	r := NewRouter(map[string]RouteBuilder{
		"/login":     func(RouteParams) gutter.Widget { return Text{Data: "login"} },
		"/dashboard": func(RouteParams) gutter.Widget { return Text{Data: "dash"} },
	}, Text{Data: "notfound"}, guard)

	r.Push("/dashboard")
	if r.Current() != "/login" {
		t.Fatalf("unauthenticated push to /dashboard went to %q, want /login", r.Current())
	}
	// Once authenticated the same push is allowed through.
	authed = true
	r.Push("/dashboard")
	if r.Current() != "/dashboard" {
		t.Fatalf("authenticated push went to %q, want /dashboard", r.Current())
	}
}

func TestRouterGuardSeedsInitialPath(t *testing.T) {
	// The seed path is run through guards at construction. Redirect everything
	// but "/welcome" so the assertion holds regardless of what initialPath()
	// returns (it's "/" on host, but the live browser shares window.location
	// across tests, so don't depend on a specific value).
	r := NewRouter(map[string]RouteBuilder{
		"/welcome": func(RouteParams) gutter.Widget { return Text{Data: "welcome"} },
	}, Text{Data: "notfound"}, func(to string) string {
		if to != "/welcome" {
			return "/welcome"
		}
		return to
	})
	if r.Current() != "/welcome" {
		t.Fatalf("guarded initial path = %q, want /welcome", r.Current())
	}
}

func TestRouterGuardChainStable(t *testing.T) {
	// Two guards: first sends /a→/b, second sends /b→/c. resolve must settle on
	// /c (re-checking after each redirect) without looping forever.
	r := NewRouter(map[string]RouteBuilder{}, Text{Data: "nf"},
		func(to string) string {
			if to == "/a" {
				return "/b"
			}
			return to
		},
		func(to string) string {
			if to == "/b" {
				return "/c"
			}
			return to
		},
	)
	if got := r.resolve("/a"); got != "/c" {
		t.Fatalf("resolve(/a) = %q, want /c", got)
	}
}

func TestRouterAddGuardAfterConstruction(t *testing.T) {
	r := NewRouter(map[string]RouteBuilder{}, Text{Data: "nf"})
	r.Guard(func(to string) string {
		if to == "/x" {
			return "/y"
		}
		return to
	})
	r.navigated("/x") // simulate a browser back/forward to /x
	if r.Current() != "/y" {
		t.Fatalf("guarded popstate = %q, want /y", r.Current())
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
