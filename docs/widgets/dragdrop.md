---
title: Drag & drop
parent: Widgets
nav_order: 35
---

# `Controller[T]` + `Draggable[T]` + `DropTarget[T]` + `DragOverlay[T]`
{: .no_toc }

A small pointer-based drag-and-drop kit. One `Controller[T]` coordinates a set of `Draggable[T]` sources and `DropTarget[T]` sinks; a `DragOverlay[T]` renders the ghost following the pointer. Generic over the payload type — strings, structs, IDs — so the drop handler receives a typed value, not an `any`.
{: .fs-6 .fw-300 }

1. TOC
{:toc}

---

## How the pieces fit

```
Controller[Task]
   │
   ├── Draggable[Task] (one per draggable card)
   │        ↑ on pointerdown: startDrag(payload, ghost)
   │        ↑ on pointerup: endDrag → DropTarget.OnDrop
   │
   ├── DropTarget[Task] (one per drop zone)
   │        ↑ on mount: register(node, accepts, onDrop, onHover)
   │
   └── DragOverlay[Task] (one near the app root)
            ↑ observes Controller; renders ghost while drag is Active
```

Construct the controller once (typically in your State's `InitState`) and pass the same pointer to every participating widget.

---

## Minimal usage

```go
type Task struct { ID int; Title string }

ctrl := widgets.NewController[Task]()

widgets.Column{Children: []gutter.Widget{
    widgets.Row{Children: []gutter.Widget{
        // A source
        widgets.Draggable[Task]{
            Controller: ctrl,
            Data:       Task{ID: 1, Title: "Hello"},
            Ghost:      widgets.Card{Variant: widgets.CardFeature, Child: widgets.Body{Text: "Hello"}},
            Child:      widgets.Card{Variant: widgets.CardFeature, Child: widgets.Body{Text: "Hello"}},
        },
        // A drop target
        widgets.DropTarget[Task]{
            Controller: ctrl,
            OnDrop:     func(t Task) { /* state.move(t.ID, "done") */ },
            Child:      widgets.Container{Width: "240px", Height: "120px", Color: "#fafafa"},
        },
    }},
    // Mounted once; renders the ghost during drag.
    widgets.DragOverlay[Task]{Controller: ctrl},
}}
```

---

## `Controller[T]`

```go
type DragState[T any] struct {
    Active           bool
    Payload          T
    X, Y             float64
    OffsetX, OffsetY float64       // pointer offset within the source card at drag start
    Ghost            gutter.Widget // captured from Draggable.Ghost
    HoverID          uint64        // internal target ID under pointer, 0 if none
}

type Controller[T any] struct { /* unexported */ }

func NewController[T any]() *Controller[T]

// Implements gutter.Listenable[DragState[T]]
func (c *Controller[T]) Value() DragState[T]
func (c *Controller[T]) Listen(fn func(DragState[T])) (cancel func())
```

The Controller is the single source of truth during a drag. It is itself a `Listenable[DragState[T]]`, so you can wrap any subtree in an `ObserverBuilder[DragState[T]]` to react to drag events globally (for example, applying a class to the body while a drag is active).

---

## `Draggable[T]`

```go
type Draggable[T any] struct {
    Controller *Controller[T]
    Data       T
    Child      gutter.Widget
    Ghost      gutter.Widget // optional; rendered under the pointer by DragOverlay
    OnDragEnd  func(dropped bool)
    Disabled   bool
}
```

| Field         | What it does                                                                       |
| ------------- | ---------------------------------------------------------------------------------- |
| `Controller`  | Required. Same instance as the matching targets.                                   |
| `Data`        | Typed payload delivered to `DropTarget.OnDrop` on a successful drop.               |
| `Child`       | Visible content. While dragging, the source dims to opacity 0.4 in place.          |
| `Ghost`       | Floats under the pointer. Pass a copy of `Child` for a Trello-style lift, or a simpler preview. |
| `OnDragEnd`   | Fires on pointerup; `dropped` is true if any target accepted the payload.           |
| `Disabled`    | Skip the pointer wiring; the child still renders.                                  |

The wrapper sets `cursor: grab` and `touch-action: none` so dragging on touch devices doesn't scroll the page.

---

## `DropTarget[T]`

```go
type DropTarget[T any] struct {
    Controller    *Controller[T]
    Child         gutter.Widget
    Accepts       func(T) bool // optional; default accepts everything
    OnDrop        func(T)
    OnHoverChange func(over bool) // optional
}
```

| Field            | What it does                                                                          |
| ---------------- | ------------------------------------------------------------------------------------- |
| `Accepts`        | Optional gate. Return false to reject the payload — the drop won't fire.              |
| `OnDrop`         | Fires when the pointer is released over this target and `Accepts` allows the payload. |
| `OnHoverChange`  | Fires when the pointer enters/leaves this target during a drag. Use it to highlight the target via `SetState`. |

The wrapper renders as a plain block-level `<div>` (not `display: contents`) so `getBoundingClientRect` returns the child's bounds — that's how the hit-tester finds the target.

---

## `DragOverlay[T]`

```go
type DragOverlay[T any] struct {
    Controller *Controller[T]
    ZIndex     string // defaults to "1500"
}
```

Mount one DragOverlay near the root of your app. It uses `ObserverBuilder` internally so only this widget rebuilds while a drag is in flight. The ghost renders `position: fixed; pointer-events: none` so it follows the cursor without catching its own events (pointer hit-tests pass through to the targets behind it).

---

## Hit-test policy

On every `pointermove`, the wasm bridge iterates every registered DropTarget, reads `getBoundingClientRect`, and picks the **smallest** rect containing the pointer. That way nested targets behave the way you'd expect — an inner target wins over the outer one when the pointer is inside both.

---

## Limitations

- `getBoundingClientRect` is read on every pointermove during the drag. For a few dozen targets this is cheap; for hundreds you'd want to cache rects at drag start.
- No keyboard-driven drag (yet). Accessibility cost — apps that need accessible reorder should expose alternative controls (up/down buttons, `aria-grabbed`, …) until this is added.
- No drag-between-frames or drag-from-file-system. For file drops use [`File`](file.html).

---

## End-to-end example: kanban

[`examples/kanban`](https://github.com/Runway-Club/gutter/tree/main/examples/kanban) builds a three-column board with seven tasks. Drag any card to any column; the source dims and a ghost follows the pointer; on drop the OnDrop callback rewrites the task's column and the app rebuilds.

---

## See also

- [Notifier + ObserverBuilder](observerbuilder.html) — the reactive plumbing Controller and DragOverlay use.
- [GestureDetector](gesturedetector.html) — the simpler pointer-events primitive for tap/swipe handling.
