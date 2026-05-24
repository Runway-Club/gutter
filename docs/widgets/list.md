---
title: List & ListBuilder
parent: Widgets
nav_order: 50
---

# `List` and `ListBuilder`
{: .no_toc }

Scrollable containers. `List` renders everything eagerly — use it for short, known-length collections. `ListBuilder` is a virtualized vertical list: only the visible window is mounted, the rest of the rows live as a virtual height. Use it for 100s or 1000s of rows where DOM cost matters.
{: .fs-6 .fw-300 }

1. TOC
{:toc}

---

## `List` — eager scroll

```go
type ListDirection string
const ( ListVertical ListDirection = "column"; ListHorizontal = "row" )

type List struct {
    Children  []gutter.Widget
    Direction ListDirection
    Spacing   float64
    Padding   EdgeInsets
    Height    string
    Width     string
    NoScroll  bool
}
```

Bounded scroll container that wraps a flex Column or Row. Without `Height`/`Width`, it grows with content and never scrolls — which is usually a bug for `List`, the whole point is to scroll.

```go
widgets.List{
    Height:  "320px",
    Spacing: 8,
    Padding: widgets.EdgeInsetsAll(12),
    Children: []gutter.Widget{
        widgets.Body{Text: "Row 1"},
        widgets.Body{Text: "Row 2"},
        // …
    },
}
```

For horizontal scroll, set `Direction: widgets.ListHorizontal` and supply `Width`. Set `NoScroll: true` to render the layout without `overflow: auto` (e.g. when an ancestor already scrolls).

---

## `ListBuilder` — virtualized scroll
{: #listbuilder }

```go
type ListBuilder struct {
    ItemCount   int
    ItemHeight  float64                       // CSS pixels, must be fixed
    ItemBuilder func(index int) gutter.Widget
    Height      string                        // viewport size
    Overscan    int                           // extra items above/below; default 3
}
```

| Field         | What it does                                                                            |
| ------------- | --------------------------------------------------------------------------------------- |
| `ItemCount`   | Total number of rows in the data set.                                                   |
| `ItemHeight`  | Fixed CSS-pixel height per row. Required — variable height isn't supported.             |
| `ItemBuilder` | Returns the widget for a row by index. Called only for currently-visible rows.          |
| `Height`      | Viewport height bound. Required.                                                        |
| `Overscan`    | Rows rendered above and below the visible window. Defaults to 3.                        |

```go
widgets.ListBuilder{
    ItemCount:  10000,
    ItemHeight: 56,
    Height:     "480px",
    ItemBuilder: func(i int) gutter.Widget {
        return widgets.Container{
            Padding: widgets.EdgeInsetsAll(16),
            Child:   widgets.Body{Text: fmt.Sprintf("Row %d", i)},
        }
    },
}
```

### How recycling works

The DOM holds `ceil(Height/ItemHeight) + 2*Overscan + 1` row nodes regardless of `ItemCount`. As the user scrolls past a row boundary, the visible window shifts and the reconciler **updates the existing row DOM nodes in place** — that's the recycling. No mount or unmount on scroll, just attribute and text updates.

For this to work:

- `ItemBuilder` should return the **same Go widget type for every index**. Type changes force a remount of that slot.
- Do **not** key items with `WithKey` unless slot identity must follow data. Keying defeats positional reuse and forces unmount + remount on every scroll tick.
- Prefer **stateless** items. State belongs to the slot, not the data — a stateful row at position 3 keeps its state when scrolling reveals new content into row 3.

### Limitations

- `ItemHeight` is fixed. Variable-height rows would need a measurement and offset cache, which the current implementation doesn't do.
- Horizontal virtualization isn't supported. Use a regular `List` for horizontal scrollers; their row count is typically small enough that eager mounting is fine.

---

## See also

- [Notifier + ObserverBuilder](observerbuilder.html) — for lists driven by external state.
- The showcase example renders 10,000 rows under a `ListBuilder` to demonstrate the recycling.
