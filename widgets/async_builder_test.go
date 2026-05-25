package widgets

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/Runway-Club/gutter"
)

func TestAsyncBuilderResolvesDuringSSR(t *testing.T) {
	w := AsyncBuilder[string]{
		Load: func(context.Context) (string, error) { return "loaded!", nil },
		Builder: func(_ *gutter.BuildContext, snap AsyncSnapshot[string]) gutter.Widget {
			if snap.State == AsyncDone {
				return Text{Data: snap.Data}
			}
			return Text{Data: "pending"}
		},
	}
	out, err := gutter.RenderToHTML(w)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "loaded!") || strings.Contains(out, "pending") {
		t.Fatalf("SSR should render the resolved value, got: %s", out)
	}
}

func TestAsyncBuilderSSRRendersError(t *testing.T) {
	w := AsyncBuilder[string]{
		Load: func(context.Context) (string, error) { return "", errors.New("boom") },
		Builder: func(_ *gutter.BuildContext, snap AsyncSnapshot[string]) gutter.Widget {
			if snap.State == AsyncFailed {
				return Text{Data: "err:" + snap.Error.Error()}
			}
			return Text{Data: "other"}
		},
	}
	out, _ := gutter.RenderToHTML(w)
	if !strings.Contains(out, "err:boom") {
		t.Fatalf("SSR should render the failed snapshot, got: %s", out)
	}
}

func TestAsyncSSRUsesRenderContext(t *testing.T) {
	type ctxKey struct{}
	ctx := context.WithValue(context.Background(), ctxKey{}, "from-request")
	var seen string
	w := AsyncBuilder[string]{
		Load: func(c context.Context) (string, error) {
			seen, _ = c.Value(ctxKey{}).(string)
			return "ok", nil
		},
		Builder: func(_ *gutter.BuildContext, snap AsyncSnapshot[string]) gutter.Widget {
			return Text{Data: snap.Data}
		},
	}
	if _, _, err := gutter.RenderDocumentCtx(ctx, w); err != nil {
		t.Fatal(err)
	}
	if seen != "from-request" {
		t.Errorf("Load received context value %q, want from-request", seen)
	}
}

func TestDepsEqual(t *testing.T) {
	cases := []struct {
		name string
		a, b []any
		want bool
	}{
		{"both nil", nil, nil, true},
		{"same scalars", []any{1, "x"}, []any{1, "x"}, true},
		{"different len", []any{1}, []any{1, 2}, false},
		{"different value", []any{1, "x"}, []any{1, "y"}, false},
		{"nil vs empty", nil, []any{}, true},
		{"non-comparable slices equal", []any{[]int{1, 2}}, []any{[]int{1, 2}}, true},
		{"non-comparable slices differ", []any{[]int{1, 2}}, []any{[]int{1, 3}}, false},
	}
	for _, c := range cases {
		if got := depsEqual(c.a, c.b); got != c.want {
			t.Errorf("%s: depsEqual = %v, want %v", c.name, got, c.want)
		}
	}
}
