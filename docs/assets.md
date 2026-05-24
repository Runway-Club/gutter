---
title: Assets
nav_order: 6
---

# Assets
{: .no_toc }

How to ship static files (images, icons, fonts you self-host) with your Gutter app — and how widgets resolve relative paths through `gutter.AssetURL`.
{: .fs-6 .fw-300 }

1. TOC
{:toc}

---

## The `./assets/` convention

`gutter build` (and `gutter run` / `gutter run dev`) copies a top-level `./assets/` directory in your project into `./dist/assets/`. Drop files there and reference them by their relative path:

```
my-app/
├── main.go
└── assets/
    ├── logo.svg
    └── icons/
        └── home.png
```

Files at `assets/logo.svg` are served at the URL `/assets/logo.svg`.

A `.gitkeep` file is created automatically by `gutter new` so the empty folder ships with your repo.

---

## `gutter.AssetURL`

```go
gutter.AssetURL("logo.svg")             // "assets/logo.svg"
gutter.AssetURL("icons/home.png")        // "assets/icons/home.png"
gutter.AssetURL("https://cdn.example/x") // returned as-is
gutter.AssetURL("/static/x")             // returned as-is (absolute path)
gutter.AssetURL("data:image/svg+xml;…")  // returned as-is
```

The helper prefixes a configurable base URL (defaulting to `"assets/"`). Absolute URLs and `data:` URIs pass through unchanged, so widget `Asset`/`Src` fields can take either form.

[`widgets.Image`](widgets/image.html) uses this automatically:

```go
widgets.Image{Asset: "logo.svg", Width: "128px"}
// Equivalent to:
widgets.Image{Src: gutter.AssetURL("logo.svg"), Width: "128px"}
```

---

## Pointing at a CDN

Override the base URL at startup, **before** any widget renders:

```go
func main() {
    gutter.SetAssetBase("https://cdn.example.com/myapp/v3/")
    gutter.RunApp(App{})
}
```

Now `Image{Asset: "logo.svg"}` resolves to `https://cdn.example.com/myapp/v3/logo.svg`. The CLI still copies `./assets/` into `./dist/assets/` (harmless if unused), so local dev keeps working.

`SetAssetBase("")` resets to the default `"assets/"`. The helper normalizes by appending a trailing `/` when missing.

---

## When NOT to use `Asset`

The `Asset` field is only for files the build pipeline produces under `./dist/assets/`. For one-off inline content (a small SVG icon, an avatar from your backend), use `Src` directly:

```go
widgets.Image{
    Src:    "data:image/svg+xml;utf8,<svg>…</svg>",
    Width:  "24px", Height: "24px",
}
widgets.Image{
    Src:    user.AvatarURL,  // arbitrary HTTPS URL
    Width:  "48px", Height: "48px",
    Rounded: "50%",
}
```

---

## Also see

- [Image](widgets/image.html) — the primary consumer of `Asset`.
- [CLI](cli.html) — `gutter build` / `gutter run` bundle internals.
