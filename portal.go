package gutter

// Portal teleports its Child out of the normal tree position and mounts it into
// a single body-level root (`#gutter-portal-root`) instead. This lets overlays
// (Popup/Drawer/BottomSheet) escape an ancestor's `overflow:hidden`, `transform`,
// or stacking context — a `position:fixed` child of a transformed ancestor is
// otherwise positioned relative to that ancestor, not the viewport.
//
// At the original tree position Portal leaves only a zero-size <template>
// anchor, so sibling layout and reconciliation positioning are unaffected.
//
// SSR: Portal renders just the placeholder anchor — its Child is NOT
// server-rendered (overlays are client-only and closed on first paint). On
// hydration the client adopts the anchor and mounts Child into the portal root.
type Portal struct {
	Child Widget
}

// Host makes Portal a HostWidget for SSR and host builds: it renders the
// zero-size placeholder anchor. The wasm runtime intercepts Portal in
// newElement (see element_wasm.go) before this is used for mounting, and
// teleports Child into the portal root instead.
func (p Portal) Host() *Host {
	return &Host{Tag: "template", Attrs: map[string]string{"data-gutter-portal": "1"}}
}
