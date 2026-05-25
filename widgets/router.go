package widgets

import (
	"net/url"
	"strings"

	"github.com/Runway-Club/gutter"
)

// RouteParams is the bag of path parameters extracted from a matched pattern.
// For the pattern "/user/:id" matched against "/user/42", RouteParams holds
// {"id": "42"}. Static routes pass a nil map.
type RouteParams map[string]string

// RouteBuilder produces the widget for a matched route. params is nil for
// routes with no :placeholders.
type RouteBuilder func(params RouteParams) gutter.Widget

// Router owns the application's current path and the route table. Construct
// once at app startup (NewRouter installs the browser history listener) and
// pass the same pointer to both RouterView (to render) and any widget that
// needs to navigate (so it can call Push/Replace/Pop).
//
//	router := widgets.NewRouter(map[string]widgets.RouteBuilder{
//	    "/":          func(_ widgets.RouteParams) gutter.Widget { return HomePage{} },
//	    "/about":     func(_ widgets.RouteParams) gutter.Widget { return AboutPage{} },
//	    "/user/:id":  func(p widgets.RouteParams) gutter.Widget { return UserPage{ID: p["id"]} },
//	}, NotFoundPage{})
//
//	gutter.RunApp(widgets.Scaffold{
//	    Title: "My App",
//	    Body:  widgets.RouterView{Router: router},
//	})
//
// Pattern syntax is intentionally minimal: literal segments must match
// exactly, segments prefixed with ":" capture the corresponding path segment.
// No wildcards and no nested routers — wrap the route builder if you need those.
//
// Guards/redirects ARE supported: pass NavGuards to NewRouter (or add them with
// Guard). Every navigation — Push/Replace, browser back/forward, and the
// initial load — is routed through the guards, which can rewrite the
// destination (e.g. send "/dashboard" to "/login" when unauthenticated).
type Router struct {
	routes   map[string]RouteBuilder
	notFound gutter.Widget
	current  *gutter.Notifier[string]
	guards   []NavGuard
}

// NavGuard inspects an intended destination path and returns the path to
// actually navigate to. Return the path unchanged to allow the navigation;
// return a different path to redirect; return the current path to effectively
// block it. Guards run in order and the result is re-checked, so one guard's
// redirect is itself guarded (capped to avoid an infinite redirect loop).
type NavGuard func(to string) string

// NewRouter creates the router, seeds its current path from the browser's
// location (on WASM) or "/" (on host) — run through any guards — and installs
// the popstate listener so browser back/forward updates the tree. notFound is
// rendered when no route pattern matches.
func NewRouter(routes map[string]RouteBuilder, notFound gutter.Widget, guards ...NavGuard) *Router {
	r := &Router{
		routes:   routes,
		notFound: notFound,
		guards:   guards,
	}
	start := initialPath()
	resolved := r.resolve(start)
	r.current = gutter.NewNotifier(resolved)
	r.installHistoryListener()
	if resolved != start {
		// Land on a guarded redirect: fix the URL bar without a history entry.
		r.replaceHistory(resolved)
	}
	return r
}

// Guard appends a NavGuard after construction (e.g. once an auth store exists).
// Returns the router for chaining.
func (r *Router) Guard(g NavGuard) *Router {
	r.guards = append(r.guards, g)
	return r
}

// resolve runs path through every guard, re-checking until the result is stable
// (so a redirect target is itself guarded). The iteration cap prevents an
// infinite loop if two guards bounce a path back and forth.
func (r *Router) resolve(to string) string {
	for range 10 {
		next := to
		for _, g := range r.guards {
			next = g(next)
		}
		if next == to {
			return to
		}
		to = next
	}
	return to
}

// navigated guards a path the browser restored (back/forward) and, if a guard
// redirected, rewrites history before updating current. Called from the wasm
// popstate listener.
func (r *Router) navigated(path string) {
	resolved := r.resolve(path)
	if resolved != path {
		r.replaceHistory(resolved)
	}
	r.current.Set(resolved)
}

// Current returns the currently active path.
func (r *Router) Current() string { return r.current.Value() }

// Listenable exposes the router as a Listenable[string] so it can be observed
// directly (e.g. by an external ObserverBuilder for breadcrumbs or analytics).
func (r *Router) Listenable() gutter.Listenable[string] { return r.current }

// Push navigates to path (after guards), pushing a new history entry.
func (r *Router) Push(path string) {
	path = r.resolve(path)
	r.pushHistory(path)
	r.current.Set(path)
}

// Replace navigates to path (after guards) without growing the history stack.
func (r *Router) Replace(path string) {
	path = r.resolve(path)
	r.replaceHistory(path)
	r.current.Set(path)
}

// Pop asks the browser to go back one entry. The popstate listener picks up
// the resulting URL and updates current.
func (r *Router) Pop() { r.popHistory() }

// match returns the widget for path, or notFound if no pattern matches.
// Query parses the query string of the current path into url.Values. For
// "/search?q=go&page=2" it returns {"q": ["go"], "page": ["2"]}. Empty when the
// path has no query string.
func (r *Router) Query() url.Values {
	_, q := splitPathQuery(r.current.Value())
	v, _ := url.ParseQuery(q)
	return v
}

// splitPathQuery separates "/path?a=1" into ("/path", "a=1").
func splitPathQuery(p string) (path, query string) {
	path, query, _ = strings.Cut(p, "?")
	return path, query
}

func (r *Router) match(path string) gutter.Widget {
	// Match on the path only; the query string is read via Router.Query.
	path, _ = splitPathQuery(path)
	if b, ok := r.routes[path]; ok {
		return b(nil)
	}
	for pat, b := range r.routes {
		if !strings.ContainsRune(pat, ':') {
			continue
		}
		if params, ok := matchPattern(pat, path); ok {
			return b(params)
		}
	}
	return r.notFound
}

// matchPattern matches a pattern like "/user/:id" against a concrete path.
// Returns the captured params and true on success.
func matchPattern(pattern, path string) (RouteParams, bool) {
	patSeg := strings.Split(strings.Trim(pattern, "/"), "/")
	pathSeg := strings.Split(strings.Trim(path, "/"), "/")
	if len(patSeg) != len(pathSeg) {
		return nil, false
	}
	params := RouteParams{}
	for i, p := range patSeg {
		if strings.HasPrefix(p, ":") {
			params[p[1:]] = pathSeg[i]
			continue
		}
		if p != pathSeg[i] {
			return nil, false
		}
	}
	return params, true
}

// RouterView renders the route that Router.match returns for the current
// path. It rebuilds via ObserverBuilder whenever the router's current path
// changes, so navigation does not need to touch any ancestor.
type RouterView struct {
	Router *Router
}

func (rv RouterView) Build(ctx *gutter.BuildContext) gutter.Widget {
	r := rv.Router
	return ObserverBuilder[string]{
		Source: r.current,
		Builder: func(_ *gutter.BuildContext, path string) gutter.Widget {
			return r.match(path)
		},
	}
}
