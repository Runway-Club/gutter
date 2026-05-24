---
title: File
parent: Widgets
nav_order: 27
---

# `File`
{: .no_toc }

A themed file picker. The trigger looks like a Button; clicking it opens the native file dialog. Once the user picks files, their bytes are read into memory via `FileReader` and handed to `OnSelect`.
{: .fs-6 .fw-300 }

1. TOC
{:toc}

---

## Signature

```go
type File struct {
    Label    string
    Child    gutter.Widget
    Accept   string
    Multiple bool
    OnSelect func([]FilePick)
    Variant  ButtonVariant
}

type FilePick struct {
    Name     string
    Size     int64
    MimeType string
    Data     []byte
}
```

| Field      | What it does                                                                       |
| ---------- | ---------------------------------------------------------------------------------- |
| `Label`    | Trigger text. Ignored when `Child` is set.                                         |
| `Child`    | Override the trigger content (e.g. an [`Icon`](icon.html)).                        |
| `Accept`   | MIME filter forwarded to the input — e.g. `"image/*,application/pdf"`.             |
| `Multiple` | Allow selecting more than one file.                                                |
| `OnSelect` | Fires once *all* picked files have been read. Slice order matches the user's pick. |
| `Variant`  | Picks the button style for the trigger (`ButtonPrimary`, `ButtonGhost`, …).        |

The widget renders a Button-styled `<label>` wrapping a hidden `<input type="file">`. The label catches clicks (native browser behavior) and triggers the file dialog.

---

## Basic usage

```go
widgets.File{
    Label:    "Upload images",
    Accept:   "image/*",
    Multiple: true,
    OnSelect: func(files []widgets.FilePick) {
        for _, f := range files {
            log.Printf("got %s (%d bytes, %s)", f.Name, len(f.Data), f.MimeType)
        }
    },
}
```

---

## With an Icon trigger

```go
widgets.File{
    Child:    widgets.Icon{Name: "upload"},
    Variant:  widgets.ButtonGhost,
    OnSelect: func(files []widgets.FilePick) { /* … */ },
}
```

---

## Notes

- On the host (non-WASM) build, the widget renders the same label but never invokes `OnSelect` — file reading needs the browser's FileReader API. The split lives in `file_wasm.go` / `file_stub.go`.
- The FileReader read is async; `OnSelect` fires once *all* picks complete. For very large files, expect the callback to be delayed.
- Files are read as `ArrayBuffer` → `Uint8Array` → Go `[]byte`. For text content, wrap with `string(pick.Data)`. For images, base64-encode or use a blob URL.

---

## See also

- [Assets](../assets.html) — the cross-platform asset URL helper for files you ship with the app.
- [Button](button.html) — the visual style this widget reuses.
