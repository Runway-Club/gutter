package widgets

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Runway-Club/gutter"
)

// ListDirection picks the layout axis for [List] and [ListBuilder].
type ListDirection string

const (
	// ListVertical stacks children top-to-bottom and scrolls vertically.
	ListVertical ListDirection = "column"
	// ListHorizontal arranges children left-to-right and scrolls
	// horizontally.
	ListHorizontal ListDirection = "row"
)

// List is a scrollable flex container — Column/Row with an explicit viewport
// and overflow behavior. Use it when the children fit in memory and the
// content scrolls inside a bounded area.
//
// For very long lists where rendering every child would be wasteful, reach
// for [ListBuilder] instead — it only mounts the items in the visible
// window.
type List struct {
	Children  []gutter.Widget
	Direction ListDirection
	Spacing   float64
	Padding   EdgeInsets

	// Height/Width bound the viewport. Without them the list grows with its
	// content and never scrolls — that's usually a bug for List, the whole
	// point is to scroll.
	Height string
	Width  string

	// NoScroll disables overflow:auto. List scrolls by default — that's
	// what distinguishes it from Column/Row — so this is opt-out.
	NoScroll bool
}

func (l List) Host() *gutter.Host {
	direction := l.Direction
	if direction == "" {
		direction = ListVertical
	}
	style := map[string]string{
		"display":        "flex",
		"flex-direction": string(direction),
		"box-sizing":     "border-box",
	}
	if l.Height != "" {
		style["height"] = l.Height
	}
	if l.Width != "" {
		style["width"] = l.Width
	}
	if l.Spacing > 0 {
		style["gap"] = fmt.Sprintf("%gpx", l.Spacing)
	}
	if !l.Padding.IsZero() {
		style["padding"] = l.Padding.CSS()
	}
	if !l.NoScroll {
		switch direction {
		case ListHorizontal:
			style["overflow-x"] = "auto"
			style["overflow-y"] = "hidden"
		default:
			style["overflow-y"] = "auto"
			style["overflow-x"] = "hidden"
		}
	}
	return &gutter.Host{
		Tag:      "div",
		Style:    style,
		Children: l.Children,
	}
}

// ListBuilder renders a virtualized vertical list: only the items whose
// rows fall inside the viewport (plus an overscan band above/below) are
// mounted into the DOM. As the user scrolls, the visible window shifts and
// the reconciler updates the existing item DOM nodes in place — that's the
// "recycling" — instead of mounting and unmounting on every scroll tick.
//
//	widgets.ListBuilder{
//	    ItemCount:  10000,
//	    ItemHeight: 56,
//	    Height:     "480px",
//	    ItemBuilder: func(i int) gutter.Widget {
//	        return widgets.Container{
//	            Padding: widgets.EdgeInsetsAll(16),
//	            Child:   widgets.Body{Text: fmt.Sprintf("Row %d", i)},
//	        }
//	    },
//	}
//
// Recycling contract — for best performance and lowest DOM churn:
//
//   - ItemBuilder should return the same Go widget type for every index. The
//     reconciler reuses DOM nodes positionally when types match; type
//     changes force a remount of that slot.
//   - Do NOT key items with WithKey unless you actually need slot identity
//     to follow the data. Keying forces every shifted-by-one item to be a
//     different match, defeating recycling. The unkeyed default is what
//     enables in-place updates.
//   - Items are best implemented as StatelessWidgets. State belongs to the
//     slot, not the data — a StatefulWidget at row 3 keeps its state when
//     scrolling reveals new content into row 3, even though the underlying
//     data row changed.
//
// Sizing along the scroll axis ("extent" = height when vertical, width when
// horizontal) can be uniform or variable:
//
//   - Uniform: set ItemHeight to a fixed CSS-pixel extent (the fast path —
//     offsets are pure arithmetic).
//   - Variable: set ItemExtent(index) to return each item's extent. The state
//     builds a prefix-sum offset cache once (rebuilt when ItemCount changes)
//     and binary-searches it to find the visible window. ItemExtent wins over
//     ItemHeight when both are set.
//
// Direction selects the axis: ListVertical (default) scrolls vertically and
// bounds the viewport with Height; ListHorizontal scrolls horizontally and
// bounds it with Width (ItemHeight/ItemExtent then describe item WIDTH).
type ListBuilder struct {
	ItemCount   int
	ItemHeight  float64
	ItemBuilder func(index int) gutter.Widget

	// ItemExtent, when non-nil, gives each item's extent along the scroll axis
	// (variable-size rows). Takes precedence over ItemHeight.
	ItemExtent func(index int) float64

	// Direction is the scroll axis; ListVertical (default) or ListHorizontal.
	Direction ListDirection

	// Height/Width bound the viewport. The one along the scroll axis is
	// required (Height for vertical, Width for horizontal) — without it the
	// list grows with its (virtual) content and never scrolls.
	Height string
	Width  string

	// Overscan is the number of extra items rendered before and after the
	// visible window. Defaults to 3. Higher values reduce flashes during
	// fast scrolls at the cost of more DOM nodes.
	Overscan int
}

func (l ListBuilder) CreateState() gutter.State { return &listBuilderState{} }

type listBuilderState struct {
	gutter.StateObject
	cleanup      func()
	scrollOffset float64
	viewportSize float64
	winFirst     int // last-rendered window bounds, used to gate rebuilds
	winLast      int
	metricsCache *listMetrics
}

func (s *listBuilderState) currentWidget() ListBuilder { return s.Widget().(ListBuilder) }

func (s *listBuilderState) Dispose() {
	if s.cleanup != nil {
		s.cleanup()
		s.cleanup = nil
	}
}

// listMetrics maps between item indices and scroll-axis offsets. The uniform
// (fixed>0) path is arithmetic; the variable path uses a prefix-sum cache where
// offsets[i] is the start of item i and offsets[count] is the total extent.
type listMetrics struct {
	count   int
	fixed   float64
	offsets []float64
}

func (m *listMetrics) total() float64 {
	if m.fixed > 0 {
		return m.fixed * float64(m.count)
	}
	if m.count == 0 {
		return 0
	}
	return m.offsets[m.count]
}

func (m *listMetrics) offset(i int) float64 {
	if m.fixed > 0 {
		return m.fixed * float64(i)
	}
	return m.offsets[i]
}

func (m *listMetrics) extent(i int) float64 {
	if m.fixed > 0 {
		return m.fixed
	}
	return m.offsets[i+1] - m.offsets[i]
}

// indexAt returns the item index whose span contains offset, clamped to
// [0, count-1].
func (m *listMetrics) indexAt(offset float64) int {
	if m.count == 0 {
		return 0
	}
	if offset <= 0 {
		return 0
	}
	if m.fixed > 0 {
		return min(int(offset/m.fixed), m.count-1)
	}
	// Largest i with offsets[i] <= offset.
	lo, hi := 0, m.count
	for lo < hi {
		mid := (lo + hi) / 2
		if m.offsets[mid+1] <= offset {
			lo = mid + 1
		} else {
			hi = mid
		}
	}
	return min(lo, m.count-1)
}

// metrics returns the cached metrics for w, rebuilding when ItemCount or the
// fixed extent changes. (Changing ItemExtent's results without changing
// ItemCount won't invalidate the cache — remount via WithKey if extents change.)
func (s *listBuilderState) metrics(w ListBuilder) *listMetrics {
	stale := s.metricsCache == nil || s.metricsCache.count != w.ItemCount ||
		(w.ItemExtent == nil && s.metricsCache.fixed != w.ItemHeight) ||
		(w.ItemExtent != nil && s.metricsCache.fixed != 0)
	if !stale {
		return s.metricsCache
	}
	m := &listMetrics{count: w.ItemCount}
	if w.ItemExtent != nil {
		m.offsets = make([]float64, w.ItemCount+1)
		for i := range w.ItemCount {
			m.offsets[i+1] = m.offsets[i] + w.ItemExtent(i)
		}
	} else {
		m.fixed = w.ItemHeight
	}
	s.metricsCache = m
	return m
}

// virtualWindow returns the inclusive [first, last] item range to render for a
// given scroll offset and viewport size (with overscan padding).
func virtualWindow(m *listMetrics, scrollOffset, viewportSize float64, overscan int) (first, last int) {
	if m.count == 0 {
		return 0, -1
	}
	first = max(m.indexAt(scrollOffset)-overscan, 0)
	last = min(m.indexAt(scrollOffset+viewportSize)+overscan, m.count-1)
	return first, max(last, first)
}

func (s *listBuilderState) Build(ctx *gutter.BuildContext) gutter.Widget {
	w := s.currentWidget()
	if (w.ItemHeight <= 0 && w.ItemExtent == nil) || w.ItemCount <= 0 || w.ItemBuilder == nil {
		return Styled{}
	}
	horizontal := w.Direction == ListHorizontal

	overscan := w.Overscan
	if overscan == 0 {
		overscan = 3
	}
	m := s.metrics(w)

	// First-mount fallback: the viewport size isn't known until the scroll
	// listener fires once after OnMount. Guess from the bounding dimension when
	// it's a px literal, else assume 400px; the first scroll callback corrects it.
	viewportSize := s.viewportSize
	if viewportSize <= 0 {
		bound := w.Height
		if horizontal {
			bound = w.Width
		}
		viewportSize = parsePxFallback(bound, 400)
	}

	first, last := virtualWindow(m, s.scrollOffset, viewportSize, overscan)

	items := make([]gutter.Widget, 0, last-first+1)
	for i := first; i <= last; i++ {
		// Wrap each item in a slot sized to its extent so row geometry matches
		// the offset math regardless of the child's own sizing.
		slot := map[string]string{"flex": "none", "box-sizing": "border-box"}
		if horizontal {
			slot["width"] = fmt.Sprintf("%gpx", m.extent(i))
		} else {
			slot["height"] = fmt.Sprintf("%gpx", m.extent(i))
		}
		items = append(items, Styled{Style: slot, Children: []gutter.Widget{w.ItemBuilder(i)}})
	}

	// The virtual sizer reserves the full content extent; the visible items sit
	// in an absolutely-positioned wrapper offset by offset(first) along the axis
	// so scroll math lines up.
	total := fmt.Sprintf("%gpx", m.total())
	startOffset := fmt.Sprintf("%gpx", m.offset(first))
	sizerStyle := map[string]string{"position": "relative"}
	innerStyle := map[string]string{"position": "absolute", "display": "flex"}
	viewportStyle := map[string]string{"position": "relative", "box-sizing": "border-box"}
	if horizontal {
		sizerStyle["width"], sizerStyle["height"] = total, "100%"
		innerStyle["top"], innerStyle["bottom"], innerStyle["left"] = "0", "0", startOffset
		innerStyle["flex-direction"] = "row"
		viewportStyle["overflow-x"], viewportStyle["overflow-y"] = "auto", "hidden"
		if w.Width != "" {
			viewportStyle["width"] = w.Width
		}
	} else {
		sizerStyle["width"], sizerStyle["height"] = "100%", total
		innerStyle["left"], innerStyle["right"], innerStyle["top"] = "0", "0", startOffset
		innerStyle["flex-direction"] = "column"
		viewportStyle["overflow-y"], viewportStyle["overflow-x"] = "auto", "hidden"
		if w.Height != "" {
			viewportStyle["height"] = w.Height
		}
	}
	sizer := Styled{Style: sizerStyle, Children: []gutter.Widget{
		Styled{Style: innerStyle, Children: items},
	}}

	return propSyncHost{
		tag:      "div",
		style:    viewportStyle,
		children: []gutter.Widget{sizer},
		onMount: func(node any) {
			if s.cleanup != nil {
				return
			}
			s.cleanup = attachScrollListener(node, horizontal, func(scrollOffset, viewportSize float64) {
				w := s.currentWidget()
				m := s.metrics(w)
				nf, nl := virtualWindow(m, scrollOffset, viewportSize, overscan)
				// scrollOffset changes every tick; only rebuild when the window
				// (or the viewport size) actually changed.
				if nf == s.winFirst && nl == s.winLast && viewportSize == s.viewportSize {
					s.scrollOffset = scrollOffset
					return
				}
				s.SetState(func() {
					s.scrollOffset = scrollOffset
					s.viewportSize = viewportSize
					s.winFirst = nf
					s.winLast = nl
				})
			})
		},
	}
}

// parsePxFallback returns the numeric pixel value of a "<n>px" CSS string,
// or fallback if the input doesn't parse. ListBuilder uses this for the
// first-render viewport guess; the real viewport size arrives via the
// scroll listener on mount.
func parsePxFallback(size string, fallback float64) float64 {
	s := strings.TrimSuffix(strings.TrimSpace(size), "px")
	if v, err := strconv.ParseFloat(strings.TrimSpace(s), 64); err == nil && v > 0 {
		return v
	}
	return fallback
}
