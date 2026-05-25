package gutter

import (
	"html"
	"sort"
	"strings"
)

// Head declares document-<head> metadata for the subtree it wraps. During SSR
// (RenderToHTML/RenderDocument/ServeSSR) the title, meta, and raw entries are
// collected and injected into the page <head> — giving server-rendered pages a
// real <title> and meta tags for SEO and social previews. On the client it sets
// document.title via SetTitle (the rest of the head is already present from the
// server render). Head is transparent in layout: it renders exactly its Child,
// so it can wrap the app root or any subtree without affecting the DOM tree.
//
//	gutter.Head{
//	    Title:    "Products — Acme",
//	    Meta:     map[string]string{"description": "Everything Acme makes."},
//	    Property: map[string]string{"og:title": "Acme Products"},
//	    Child:    app,
//	}
type Head struct {
	// Title sets <title> (SSR) and document.title (client).
	Title string
	// Meta maps a <meta name=..> to its content, e.g. "description": "...".
	Meta map[string]string
	// Property maps a <meta property=..> to its content for Open Graph/Twitter,
	// e.g. "og:title": "...".
	Property map[string]string
	// Raw is extra head HTML appended verbatim, e.g. a <link rel="canonical">.
	Raw []string
	// Child is the wrapped subtree, rendered in place of Head.
	Child Widget
}

// Build makes Head a StatelessWidget: it sets the title on the client and
// renders its Child unchanged. Head HTML collection happens in the SSR walk.
func (h Head) Build(ctx *BuildContext) Widget {
	if h.Title != "" {
		SetTitle(h.Title)
	}
	return h.Child
}

// headProvider is implemented by widgets that contribute to the document head
// during SSR. The walk in ssr.go checks for it and accumulates the result.
type headProvider interface {
	headHTML() string
}

func (h Head) headHTML() string {
	var b strings.Builder
	if h.Title != "" {
		b.WriteString("<title>")
		b.WriteString(html.EscapeString(h.Title))
		b.WriteString("</title>")
	}
	writeMetaTags(&b, "name", h.Meta)
	writeMetaTags(&b, "property", h.Property)
	for _, raw := range h.Raw {
		b.WriteString(raw)
	}
	return b.String()
}

// writeMetaTags emits <meta {attr}="key" content="val"> for each entry, sorted
// for deterministic output (golden tests, stable diffs).
func writeMetaTags(b *strings.Builder, attr string, m map[string]string) {
	if len(m) == 0 {
		return
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		b.WriteString("<meta ")
		b.WriteString(attr)
		b.WriteString(`="`)
		b.WriteString(html.EscapeString(k))
		b.WriteString(`" content="`)
		b.WriteString(html.EscapeString(m[k]))
		b.WriteString(`">`)
	}
}
