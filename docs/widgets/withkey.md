---
title: WithKey
parent: Widgets
nav_order: 19
---

# `WithKey`
{: .no_toc }

Wraps a child with a reconciliation key. Use it when the wrapped widget doesn't implement [`gutter.Keyed`](../architecture.html#keys) itself.
{: .fs-6 .fw-300 }

1. TOC
{:toc}

---

## Signature

```go
type WithKey struct {
    Key   any
    Child gutter.Widget
}
```

`WithKey` is a `StatelessWidget` whose `Build` returns `Child` and which implements `Keyed.WidgetKey() any` to return `Key`. The rendered tree is whatever Child produces; `WithKey` only adds an identity to the element tree.

---

## When to use it

Use `WithKey` for **lists whose order can change** (or whose middle can be inserted into or deleted from) when the child widgets don't already implement `Keyed`. Without a key, the reconciler matches siblings of the same Go type positionally, which assigns State to the wrong items after a reorder.

```go
widgets.Column{
    Children: []gutter.Widget{
        widgets.WithKey{Key: todo.ID, Child: TodoItem{Todo: todo}},
        // … one per todo …
    },
}
```

See [Architecture → Keys](../architecture.html#keys) and [State Management → Keyed reconciliation](../state-management.html#keyed-reconciliation) for the full story.

---

## When NOT to use it

- The list is **stable** (no reorder/insert/delete in the middle) — positional matching works fine, keys add nothing.
- You **own the widget**; implement `gutter.Keyed` directly so callers don't have to remember to wrap:

  ```go
  type TodoItem struct{ Todo Todo }
  func (t TodoItem) WidgetKey() any { return t.Todo.ID }
  ```

`WithKey` is the ad-hoc tool for widgets you don't own (e.g. wrapping a built-in inside a `Column`).

---

## Notes

- The `Key` is compared with `==` — use a hashable type. Strings, ints, and structs of those work; slices/maps/funcs don't.
- `WithKey` adds one extra `statelessElement` layer to the tree. Cheap, but it's there.

---

## See also

- [Architecture → Keys](../architecture.html#keys) — what keys do.
- [State Management → Keyed reconciliation](../state-management.html#keyed-reconciliation) — when (and why) to key.
