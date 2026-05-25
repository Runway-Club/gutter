package gutter

import "testing"

type diSvc struct{ name string }

// diReader renders the provided *diSvc's name, or "none" if absent.
type diReader struct{}

func (diReader) Build(ctx *BuildContext) Widget {
	if s, ok := DependOn[*diSvc](ctx); ok {
		return ssrBox{tag: "span", text: s.name}
	}
	return ssrBox{tag: "span", text: "none"}
}

func TestProviderScopesToSubtree(t *testing.T) {
	tree := ssrBox{tag: "div", children: []Widget{
		Provider[*diSvc]{Value: &diSvc{name: "alpha"}, Child: diReader{}},
		diReader{}, // sibling OUTSIDE the provider must NOT see the value
	}}
	out, err := RenderToHTML(tree)
	if err != nil {
		t.Fatal(err)
	}
	want := "<div><span>alpha</span><span>none</span></div>"
	if out != want {
		t.Fatalf("got %q\nwant %q", out, want)
	}
}

func TestProviderNestedShadowing(t *testing.T) {
	tree := Provider[*diSvc]{Value: &diSvc{name: "outer"}, Child: ssrBox{tag: "div", children: []Widget{
		diReader{}, // sees outer
		Provider[*diSvc]{Value: &diSvc{name: "inner"}, Child: diReader{}}, // shadows with inner
	}}}
	out, err := RenderToHTML(tree)
	if err != nil {
		t.Fatal(err)
	}
	want := "<div><span>outer</span><span>inner</span></div>"
	if out != want {
		t.Fatalf("got %q\nwant %q", out, want)
	}
}

func TestDependOnAbsentAndNilContext(t *testing.T) {
	if _, ok := DependOn[*diSvc](&BuildContext{}); ok {
		t.Fatal("expected not-found on empty context")
	}
	if _, ok := DependOn[*diSvc](nil); ok {
		t.Fatal("expected not-found on nil context")
	}
}
