package gutter

// Server-side rendering: walk a widget tree to an HTML string in pure Go.
//
// This is the platform-neutral counterpart to the WASM Element tree. It never
// touches syscall/js — it only calls the same Build()/Host() methods the
// reconciler does (mirroring the type-switch in newElement) and serializes the
// resulting Host structs to HTML. State is built exactly once: there is no
// SetState loop, no Dispose, and Events/OnMount are not invoked — their
// presence is recorded as a `data-gutter-h` marker so a future hydration pass
// knows which nodes need wiring up.
//
// Because it builds on the host platform, a server can pre-render the initial
// HTML for instant first paint + SEO, then ship app.wasm to take over.

import (
	"fmt"
	"html"
	"sort"
	"strings"
)

// ssrVoidElements are HTML elements that have no closing tag or children.
var ssrVoidElements = map[string]bool{
	"area": true, "base": true, "br": true, "col": true, "embed": true,
	"hr": true, "img": true, "input": true, "link": true, "meta": true,
	"param": true, "source": true, "track": true, "wbr": true,
}

// RenderToHTML renders a widget tree to HTML on the server. It accepts the same
// Options as RunApp (only WithTheme is meaningful here) so themed widgets read
// the same tokens they will at runtime. Returns an error if the tree contains a
// value that implements none of HostWidget/StatelessWidget/StatefulWidget.
func RenderToHTML(root Widget, opts ...Option) (string, error) {
	_, body, err := RenderDocument(root, opts...)
	return body, err
}

// RenderDocument renders the body HTML plus the <head> HTML contributed by any
// gutter.Head widgets in the tree (title/meta/raw). ServeSSR uses this so
// server-rendered pages get a real <title> and meta tags; call it directly if
// you assemble your own document shell. The body is identical to RenderToHTML.
func RenderDocument(root Widget, opts ...Option) (head, body string, err error) {
	cfg := newRunConfig(opts)
	ctx := &BuildContext{Theme: cfg.theme}
	var bodyB, headB strings.Builder
	if err := ssrRender(&bodyB, &headB, root, ctx); err != nil {
		return "", "", err
	}
	return headB.String(), bodyB.String(), nil
}

func ssrRender(sb, head *strings.Builder, w Widget, ctx *BuildContext) error {
	if w == nil {
		return nil
	}
	// Widgets may contribute to the document <head> (gutter.Head) regardless of
	// what they render in the body.
	if hp, ok := w.(headProvider); ok {
		head.WriteString(hp.headHTML())
	}
	// Mirror newElement's dispatch order: Host, Stateful, Stateless.
	switch x := w.(type) {
	case HostWidget:
		return ssrRenderHost(sb, head, x.Host(), w, ctx)
	case StatefulWidget:
		st := x.CreateState()
		if b, ok := st.(widgetBinder); ok {
			b.bindWidget(w) // so State.Widget() works during Build
		}
		if init, ok := st.(StateInitializer); ok {
			init.InitState()
		}
		// No bindElement: s.elem stays nil, so any SetState during Build is a
		// no-op (see StateObject.SetState) — correct for a one-shot render.
		return ssrRender(sb, head, st.Build(ctx), ctx)
	case StatelessWidget:
		saved := ctx.inherited
		if p, ok := x.(inheritedProvider); ok {
			ctx.inherited = p.provideInto(ctx.inherited)
		}
		err := ssrRender(sb, head, x.Build(ctx), ctx)
		ctx.inherited = saved
		return err
	default:
		return fmt.Errorf("gutter: RenderToHTML: %T implements none of HostWidget/StatelessWidget/StatefulWidget", w)
	}
}

func ssrRenderHost(sb, head *strings.Builder, h *Host, w Widget, ctx *BuildContext) error {
	if h == nil {
		return nil
	}
	tag := h.Tag
	if tag == "" {
		tag = "div"
	}

	sb.WriteString("<")
	sb.WriteString(tag)
	ssrWriteAttrs(sb, h.Attrs)
	if len(h.Style) > 0 {
		sb.WriteString(` style="`)
		sb.WriteString(html.EscapeString(ssrStyleString(h.Style)))
		sb.WriteString(`"`)
	}
	// Hydration markers (consumed by the future hydrate pass).
	if k := ssrKeyOf(w); k != nil {
		sb.WriteString(` data-gutter-key="`)
		sb.WriteString(html.EscapeString(fmt.Sprint(k)))
		sb.WriteString(`"`)
	}
	if len(h.Events) > 0 || h.OnMount != nil {
		sb.WriteString(` data-gutter-h="1"`)
	}
	sb.WriteString(">")

	if ssrVoidElements[tag] {
		return nil // void elements take no text, children, or closing tag
	}

	// Text and Children are mutually exclusive in practice (see Host docs).
	if h.Text != "" {
		sb.WriteString(html.EscapeString(h.Text))
	}
	for _, child := range h.Children {
		if err := ssrRender(sb, head, child, ctx); err != nil {
			return err
		}
	}

	sb.WriteString("</")
	sb.WriteString(tag)
	sb.WriteString(">")
	return nil
}

func ssrWriteAttrs(sb *strings.Builder, attrs map[string]string) {
	if len(attrs) == 0 {
		return
	}
	keys := make([]string, 0, len(attrs))
	for k := range attrs {
		keys = append(keys, k)
	}
	sort.Strings(keys) // deterministic output for golden tests + stable hydration
	for _, k := range keys {
		sb.WriteString(" ")
		sb.WriteString(k)
		sb.WriteString(`="`)
		sb.WriteString(html.EscapeString(attrs[k]))
		sb.WriteString(`"`)
	}
}

func ssrStyleString(style map[string]string) string {
	keys := make([]string, 0, len(style))
	for k := range style {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var b strings.Builder
	for i, k := range keys {
		if i > 0 {
			b.WriteString("; ")
		}
		b.WriteString(k)
		b.WriteString(": ")
		b.WriteString(style[k])
	}
	return b.String()
}

func ssrKeyOf(w Widget) any {
	if k, ok := w.(Keyed); ok {
		return k.WidgetKey()
	}
	return nil
}
