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
// ItemHeight must be a fixed CSS-pixel value; variable-height rows would
// require measurement and an offset cache, which this implementation does
// not do. Horizontal virtualization is not supported — use a plain [List]
// for horizontal scrolling.
type ListBuilder struct {
	ItemCount   int
	ItemHeight  float64
	ItemBuilder func(index int) gutter.Widget

	// Height bounds the viewport. Required — without it the list grows
	// with its (virtual) content and never scrolls.
	Height string

	// Overscan is the number of extra items rendered above and below the
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
	firstVisible int
}

func (s *listBuilderState) currentWidget() ListBuilder { return s.Widget().(ListBuilder) }

func (s *listBuilderState) Dispose() {
	if s.cleanup != nil {
		s.cleanup()
		s.cleanup = nil
	}
}

func (s *listBuilderState) Build(ctx *gutter.BuildContext) gutter.Widget {
	w := s.currentWidget()
	if w.ItemHeight <= 0 || w.ItemCount <= 0 || w.ItemBuilder == nil {
		return Styled{}
	}

	overscan := w.Overscan
	if overscan == 0 {
		overscan = 3
	}

	// First-mount fallback: we don't know the viewport size until the
	// scroll listener fires once after OnMount. Render a sensible window
	// based on Height-as-string when it's a px literal, otherwise assume
	// a 400px viewport. The first scroll callback will SetState with the
	// real clientHeight and trigger a corrective rebuild.
	viewportSize := s.viewportSize
	if viewportSize <= 0 {
		viewportSize = parsePxFallback(w.Height, 400)
	}

	visibleCount := int(viewportSize/w.ItemHeight) + 1 + overscan*2
	if visibleCount > w.ItemCount {
		visibleCount = w.ItemCount
	}
	firstVisible := s.firstVisible - overscan
	if firstVisible < 0 {
		firstVisible = 0
	}
	if firstVisible+visibleCount > w.ItemCount {
		firstVisible = w.ItemCount - visibleCount
		if firstVisible < 0 {
			firstVisible = 0
		}
	}

	items := make([]gutter.Widget, 0, visibleCount)
	for i := firstVisible; i < firstVisible+visibleCount && i < w.ItemCount; i++ {
		// Wrap each item in a fixed-height slot so the row geometry is
		// predictable regardless of the child's own sizing.
		items = append(items, Styled{
			Style: map[string]string{
				"height":     fmt.Sprintf("%gpx", w.ItemHeight),
				"flex":       "none",
				"box-sizing": "border-box",
			},
			Children: []gutter.Widget{w.ItemBuilder(i)},
		})
	}

	// The virtual sizer fills the viewport with itemCount*itemHeight of
	// empty space; the visible items sit in an absolutely-positioned
	// wrapper offset by firstVisible*itemHeight so scroll math lines up.
	sizer := Styled{
		Style: map[string]string{
			"position": "relative",
			"width":    "100%",
			"height":   fmt.Sprintf("%gpx", float64(w.ItemCount)*w.ItemHeight),
		},
		Children: []gutter.Widget{
			Styled{
				Style: map[string]string{
					"position":       "absolute",
					"left":           "0",
					"right":          "0",
					"top":            fmt.Sprintf("%gpx", float64(firstVisible)*w.ItemHeight),
					"display":        "flex",
					"flex-direction": "column",
				},
				Children: items,
			},
		},
	}

	viewportStyle := map[string]string{
		"position":   "relative",
		"box-sizing": "border-box",
		"overflow-y": "auto",
		"overflow-x": "hidden",
	}
	if w.Height != "" {
		viewportStyle["height"] = w.Height
	}

	return propSyncHost{
		tag:      "div",
		style:    viewportStyle,
		children: []gutter.Widget{sizer},
		onMount: func(node any) {
			if s.cleanup != nil {
				return
			}
			s.cleanup = attachScrollListener(node, func(scrollOffset, viewportSize float64) {
				w := s.currentWidget()
				if w.ItemHeight <= 0 {
					return
				}
				newFirst := int(scrollOffset / w.ItemHeight)
				// Gate rebuilds on the values that actually affect the
				// rendered window. scrollOffset itself changes every
				// scroll tick — we only care when we've crossed an item
				// boundary or the viewport resized.
				if newFirst == s.firstVisible && viewportSize == s.viewportSize {
					s.scrollOffset = scrollOffset
					return
				}
				s.SetState(func() {
					s.scrollOffset = scrollOffset
					s.viewportSize = viewportSize
					s.firstVisible = newFirst
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
