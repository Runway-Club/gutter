package widgets

import "testing"

func TestListMetricsFixed(t *testing.T) {
	m := &listMetrics{count: 100, fixed: 20}
	if m.total() != 2000 {
		t.Errorf("total = %g, want 2000", m.total())
	}
	if m.offset(5) != 100 {
		t.Errorf("offset(5) = %g, want 100", m.offset(5))
	}
	if m.extent(5) != 20 {
		t.Errorf("extent(5) = %g, want 20", m.extent(5))
	}
	if got := m.indexAt(105); got != 5 { // 105/20 = 5.25 → 5
		t.Errorf("indexAt(105) = %d, want 5", got)
	}
	if got := m.indexAt(1e9); got != 99 { // clamps to last
		t.Errorf("indexAt(huge) = %d, want 99", got)
	}
}

func TestListMetricsVariable(t *testing.T) {
	// Extents 10, 30, 20, 40 → offsets 0,10,40,60,100.
	m := &listMetrics{count: 4, offsets: []float64{0, 10, 40, 60, 100}}
	if m.total() != 100 {
		t.Errorf("total = %g, want 100", m.total())
	}
	if m.offset(2) != 40 {
		t.Errorf("offset(2) = %g, want 40", m.offset(2))
	}
	if m.extent(1) != 30 {
		t.Errorf("extent(1) = %g, want 30", m.extent(1))
	}
	cases := map[float64]int{0: 0, 5: 0, 10: 1, 39: 1, 40: 2, 59: 2, 60: 3, 99: 3, 200: 3}
	for off, want := range cases {
		if got := m.indexAt(off); got != want {
			t.Errorf("indexAt(%g) = %d, want %d", off, got, want)
		}
	}
}

func TestVirtualWindow(t *testing.T) {
	m := &listMetrics{count: 100, fixed: 20}
	// Offset 200, viewport 200: indexAt(200)=10, indexAt(400)=20 (item 20 starts
	// exactly at 400); ±overscan 2 → [8, 22].
	first, last := virtualWindow(m, 200, 200, 2)
	if first != 8 || last != 22 {
		t.Fatalf("window = [%d,%d], want [8,22]", first, last)
	}
	// At the top, first clamps to 0.
	first, _ = virtualWindow(m, 0, 200, 3)
	if first != 0 {
		t.Errorf("first at top = %d, want 0", first)
	}
	// Empty list.
	if f, l := virtualWindow(&listMetrics{count: 0, fixed: 20}, 0, 200, 3); f != 0 || l != -1 {
		t.Errorf("empty window = [%d,%d], want [0,-1]", f, l)
	}
}
