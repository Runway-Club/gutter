package widgets

import (
	"testing"

	"github.com/Runway-Club/gutter"
	"github.com/Runway-Club/gutter/themes"
)

// testCtx is the BuildContext threaded into Build() during unit tests. Default
// theme is Apple; tests that need a specific theme build their own ctx.
func testCtx(th *themes.Theme) *gutter.BuildContext {
	return &gutter.BuildContext{Theme: th}
}

// hostOf resolves a widget down to the single *gutter.Host it ultimately
// renders, recursing through StatelessWidget.Build layers (Container → Styled,
// Heading → Text, etc.). It fails the test for StatefulWidgets, which need the
// full runtime and are covered by the WASM/e2e layers instead.
func hostOf(t *testing.T, w gutter.Widget) *gutter.Host {
	t.Helper()
	return hostOfCtx(t, w, testCtx(themes.Apple))
}

func hostOfCtx(t *testing.T, w gutter.Widget, ctx *gutter.BuildContext) *gutter.Host {
	t.Helper()
	switch x := w.(type) {
	case gutter.HostWidget:
		return x.Host()
	case gutter.StatelessWidget:
		return hostOfCtx(t, x.Build(ctx), ctx)
	case gutter.StatefulWidget:
		t.Fatalf("hostOf: %T is a StatefulWidget; test it via the runtime/e2e layer", w)
	}
	t.Fatalf("hostOf: %T implements no widget interface", w)
	return nil
}

// wantStyle asserts a single CSS property equals want.
func wantStyle(t *testing.T, h *gutter.Host, prop, want string) {
	t.Helper()
	if got := h.Style[prop]; got != want {
		t.Errorf("style[%q] = %q, want %q", prop, got, want)
	}
}

// wantNoStyle asserts a CSS property is absent (zero value).
func wantNoStyle(t *testing.T, h *gutter.Host, prop string) {
	t.Helper()
	if got, ok := h.Style[prop]; ok {
		t.Errorf("style[%q] = %q, want it to be unset", prop, got)
	}
}

func wantTag(t *testing.T, h *gutter.Host, want string) {
	t.Helper()
	if h.Tag != want {
		t.Errorf("tag = %q, want %q", h.Tag, want)
	}
}

func wantChildren(t *testing.T, h *gutter.Host, n int) {
	t.Helper()
	if len(h.Children) != n {
		t.Errorf("child count = %d, want %d", len(h.Children), n)
	}
}
