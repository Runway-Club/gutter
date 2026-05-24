---
title: Router & RouterView
parent: Widgets
nav_order: 73
---

# `Router` + `RouterView`
{: .no_toc }

Path-based routing with `:param` captures and full browser history integration. The router itself is a `Listenable[string]` so any widget can observe the current path; `RouterView` renders the route currently matched.
{: .fs-6 .fw-300 }

1. TOC
{:toc}

---

## Signatures

```go
type RouteParams map[string]string

type RouteBuilder func(params RouteParams) gutter.Widget

type Router struct { /* unexported */ }

func NewRouter(routes map[string]RouteBuilder, notFound gutter.Widget) *Router

// Methods on *Router:
//   Current() string
//   Listenable() gutter.Listenable[string]
//   Push(path string)        // grows history
//   Replace(path string)     // replaces top of history
//   Pop()                    // browser back

type RouterView struct { Router *Router }
```

Pattern syntax: literal segments must match exactly; segments prefixed with `:` capture the corresponding path segment. No wildcards, no nested routers, no guards — wrap a `RouteBuilder` if you need those.

---

## Basic usage

```go
router := widgets.NewRouter(map[string]widgets.RouteBuilder{
    "/":          func(_ widgets.RouteParams) gutter.Widget { return HomePage{} },
    "/about":     func(_ widgets.RouteParams) gutter.Widget { return AboutPage{} },
    "/user/:id":  func(p widgets.RouteParams) gutter.Widget { return UserPage{ID: p["id"]} },
}, NotFoundPage{})

gutter.RunApp(widgets.Scaffold{
    Title: "My App",
    Body:  widgets.RouterView{Router: router},
})
```

Navigate from a button:

```go
widgets.Button{
    Label:     "About",
    OnPressed: func() { router.Push("/about") },
}
```

The router hooks the browser's `popstate` event, so back/forward update the rendered route automatically.

---

## Initial path normalization

The page can be opened at `/index.html` on most static servers. If that doesn't match any route, the matcher renders `notFound`. To force a sensible first paint, normalize in your owner's `InitState`:

```go
if _, ok := routes[router.Current()]; !ok {
    router.Replace("/")
}
```

---

## Serving SPA-style routes

When the user reloads on `/about`, the server must return your `index.html` rather than 404. The `gutter run` CLI does this automatically (SPA fallback for any extensionless path); for a production server, configure the equivalent — nginx `try_files $uri /index.html`, Caddy `try_files`, etc.

---

## See also

- [State Management](../state-management.html) — for the `Notifier`-style observer pattern Router builds on.
- [Examples](../examples.html) — `examples/router` for a runnable demo.
