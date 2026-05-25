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
// No wildcards, no nested routers, no guards — wrap the route builder if you
// need those.
type Router struct {
	routes   map[string]RouteBuilder
	notFound gutter.Widget
	current  *gutter.Notifier[string]
}

// NewRouter creates the router, seeds its current path from the browser's
// location (on WASM) or "/" (on host), and installs the popstate listener so
// browser back/forward updates the tree. notFound is rendered when no route
// pattern matches.
func NewRouter(routes map[string]RouteBuilder, notFound gutter.Widget) *Router {
	r := &Router{
		routes:   routes,
		notFound: notFound,
		current:  gutter.NewNotifier(initialPath()),
	}
	r.installHistoryListener()
	return r
}

// Current returns the currently active path.
func (r *Router) Current() string { return r.current.Value() }

// Listenable exposes the router as a Listenable[string] so it can be observed
// directly (e.g. by an external ObserverBuilder for breadcrumbs or analytics).
func (r *Router) Listenable() gutter.Listenable[string] { return r.current }

// Push navigates to path, pushing a new history entry.
func (r *Router) Push(path string) {
	r.pushHistory(path)
	r.current.Set(path)
}

// Replace navigates to path without growing the history stack.
func (r *Router) Replace(path string) {
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
