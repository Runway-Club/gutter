package gutter

import "strings"

// InspectNode is a snapshot of one node in the live element tree, produced by
// Inspect() for devtools and debugging. It is a plain data tree (no DOM
// handles), so it is safe to log, diff, or assert on in tests.
type InspectNode struct {
	Kind     string // "host" | "stateless" | "stateful" | "portal"
	Type     string // the widget's Go type, e.g. "widgets.Button"
	Tag      string // DOM tag for host nodes; "" otherwise
	Key      string // reconciliation key, "" if unkeyed
	Children []InspectNode
}

// String renders the node and its descendants as an indented tree, one element
// per line — the text the devtools overlay displays.
func (n InspectNode) String() string {
	var b strings.Builder
	n.write(&b, 0)
	return b.String()
}

func (n InspectNode) write(b *strings.Builder, depth int) {
	b.WriteString(strings.Repeat("  ", depth))
	b.WriteString(n.Kind)
	if n.Type != "" {
		b.WriteByte(' ')
		b.WriteString(n.Type)
	}
	if n.Tag != "" {
		b.WriteString(" <")
		b.WriteString(n.Tag)
		b.WriteByte('>')
	}
	if n.Key != "" {
		b.WriteString(" key=")
		b.WriteString(n.Key)
	}
	b.WriteByte('\n')
	for _, c := range n.Children {
		c.write(b, depth+1)
	}
}

// Count returns the total number of nodes in the subtree (self + descendants).
func (n InspectNode) Count() int {
	total := 1
	for _, c := range n.Children {
		total += c.Count()
	}
	return total
}
