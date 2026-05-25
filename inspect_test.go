package gutter

import "testing"

func TestInspectNodeStringAndCount(t *testing.T) {
	n := InspectNode{Kind: "stateful", Type: "app.Counter", Children: []InspectNode{
		{Kind: "host", Type: "widgets.Styled", Tag: "div", Children: []InspectNode{
			{Kind: "host", Tag: "button", Key: "k1"},
		}},
	}}
	want := "stateful app.Counter\n" +
		"  host widgets.Styled <div>\n" +
		"    host <button> key=k1\n"
	if got := n.String(); got != want {
		t.Fatalf("String() =\n%q\nwant\n%q", got, want)
	}
	if got := n.Count(); got != 3 {
		t.Errorf("Count() = %d, want 3", got)
	}
}
