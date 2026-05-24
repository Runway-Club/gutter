---
title: Image
parent: Widgets
nav_order: 28
---

# `Image`
{: .no_toc }

An HTML `<img>` element. Source can be a declared asset (resolved through `gutter.AssetURL`) or an absolute URL — including `data:` URIs for inline content.
{: .fs-6 .fw-300 }

1. TOC
{:toc}

---

## Signature

```go
type Image struct {
    Asset   string
    Src     string
    Alt     string
    Width   string
    Height  string
    Fit     ImageFit
    Rounded string
}
```

| Field    | What it does                                                                          |
| -------- | ------------------------------------------------------------------------------------- |
| `Asset`  | Relative path; resolved through `gutter.AssetURL` (default base `"assets/"`).         |
| `Src`    | Absolute URL or `data:` URI. Ignored when `Asset` is set.                             |
| `Alt`    | Accessibility text.                                                                   |
| `Width`  | CSS length: `"64px"`, `"100%"`, etc.                                                  |
| `Height` | CSS length.                                                                           |
| `Fit`    | One of `ImageFitCover/Contain/Fill/None/ScaleDown` — maps to CSS `object-fit`.        |
| `Rounded`| Optional CSS border-radius — e.g. `"50%"` for a circular avatar.                      |

---

## With a project asset

Drop a file into `./assets/` (the CLI copies it into `dist/assets/`):

```
my-app/
├── main.go
└── assets/
    └── logo.svg
```

Then reference it:

```go
widgets.Image{Asset: "logo.svg", Width: "128px"}
```

`gutter.AssetURL("logo.svg")` → `"assets/logo.svg"`. Override the base for a CDN: `gutter.SetAssetBase("https://cdn.example.com/v3/")`. See the [Assets](../assets.html) page.

---

## With an absolute URL or data URI

```go
widgets.Image{
    Src:    "https://example.com/portrait.jpg",
    Width:  "240px",
    Height: "240px",
    Fit:    widgets.ImageFitCover,
    Rounded: "50%",
}
```

`data:` URIs are useful for tiny inline icons:

```go
widgets.Image{Src: "data:image/svg+xml;utf8,<svg…/>", Width: "24px", Height: "24px"}
```

---

## Notes

- `Asset` wins over `Src` when both are set.
- Without `Width` or `Height`, the image renders at its intrinsic size.
- `Fit` only matters when both `Width` and `Height` constrain the box.

---

## See also

- [Assets](../assets.html) — the asset-resolution pipeline.
- [Icon](icon.html) — vector glyphs without an asset file.
