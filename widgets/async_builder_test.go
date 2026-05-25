package widgets

import "testing"

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
