---
title: CLI
nav_order: 6
---

# The `gutter` CLI
{: .no_toc }

`gutter` scaffolds projects, builds them to WebAssembly, serves them locally with optional live reload, and packages them for deployment.
{: .fs-6 .fw-300 }

1. TOC
{:toc}

---

## Install

```sh
go install github.com/Runway-Club/gutter/cmd/gutter@latest
```

This drops a `gutter` binary into `$GOBIN` (or `$GOPATH/bin`). Check it's on your `PATH`:

```sh
gutter --version
# gutter version 0.2.0
```

From a local checkout:

```sh
go build -o bin/gutter ./cmd/gutter
```

---

## `gutter new` — scaffold a project

```text
gutter new [name] [--module github.com/you/name]
```

Creates a directory `name/` containing `main.go`, `index.html`, `go.mod`, and a `.gitignore` (covering `/dist/` and any stray top-level `app.wasm` / `wasm_exec.js`). Without arguments, it walks you through a short interactive prompt for the project name and Go module path. After writing the files it runs `go get github.com/Runway-Club/gutter@latest` inside the new project so `go.mod` is pinned to the current published version.

```sh
# Interactive:
gutter new

# Both args inline:
gutter new myapp --module github.com/me/myapp
```

The scaffold renders a complete "Hello, Gutter!" app — `Scaffold` with `AppBar`, centered `Card`, `Heading`, `Body`, and primary `Button` — so you can `gutter run` it immediately and see something working.

> **Local checkout?** `gutter new` does **not** emit a `replace` directive. If you want your scaffolded `go.mod` to point at your local Gutter source, add it yourself:
>
> ```
> replace github.com/Runway-Club/gutter => ../path/to/gutter
> ```

---

## `gutter run` — build + serve

```text
gutter run [--addr :8080]
```

The fastest path to seeing pixels:

1. Bundles your project into `./dist/` with `GOOS=js GOARCH=wasm`:
   - Compiles to `./dist/app.wasm`.
   - Copies `wasm_exec.js` from `$GOROOT/lib/wasm/` (Go 1.24+) or `$GOROOT/misc/wasm/` (older) into `./dist/`.
   - Copies `index.html` and everything in `./public/` (if present) into `./dist/`.
2. Registers `application/wasm` as the MIME type for `.wasm` (required — browsers refuse to instantiate WASM served with the wrong MIME).
3. Serves `./dist/` over HTTP on `:8080`.

Your project root stays clean — every artifact lives under `./dist/`, the same directory `gutter build` writes to. So what you see running locally is byte-identical to what you'd deploy.

```sh
gutter run            # serves at http://localhost:8080
gutter run --addr :3000   # custom port
```

### `gutter run dev` — live reload

```text
gutter run dev [--addr :8080]
```

Same as `gutter run`, plus:

- Watches `.go`, `.html`, and `.css` files (skipping `.git`, `node_modules`, `dist`, `vendor`, and dotfiles — so writes into `./dist/` never cause an infinite rebuild loop).
- Re-bundles into `./dist/` on save, debouncing rapid changes by ~150ms.
- Increments an internal build counter on every successful rebuild.
- Injects a tiny `<script>` snippet into `index.html` responses that polls `/__gutter/build` every 500ms; when the counter changes, the browser reloads.

```sh
gutter run dev
```

Save a `.go` file. The CLI prints `change detected — rebuilding`, then `rebuilt in 240ms — browser reloading`, and your tab refreshes. No browser extension needed.

You can opt into watcher debug logs:

```sh
GUTTER_WATCH_DEBUG=1 gutter run dev
```

---

## `gutter build` — production bundle

```text
gutter build
```

Writes a self-contained static bundle into `./dist/`:

```text
dist/
├── app.wasm
├── index.html       # copied from current dir
├── wasm_exec.js     # from $GOROOT
└── …                # plus everything in ./public/, if it exists
```

Drop `dist/` behind any static file server, as long as it serves `.wasm` files with `Content-Type: application/wasm`:

- nginx — `types { application/wasm wasm; }` in the server block.
- Caddy — `header /*.wasm Content-Type application/wasm`.
- GitHub Pages — works out of the box.
- Cloudflare Pages — works out of the box.
- S3 + CloudFront — set the metadata on `app.wasm` to `Content-Type: application/wasm`.

### Including extra static assets

Anything you drop in `./public/` is copied into `dist/` recursively. Good for favicons, images, fonts you serve yourself, JSON data files:

```text
myapp/
├── main.go
├── index.html
├── go.mod
└── public/
    ├── favicon.ico
    └── img/
        └── hero.png
```

After `gutter build`:

```text
dist/
├── app.wasm
├── index.html
├── wasm_exec.js
├── favicon.ico
└── img/
    └── hero.png
```

---

## `gutter build deploy` — Docker image

```text
gutter build deploy [--image registry.example.com/app:tag] [--no-build]
```

Builds the project (you can skip this with `--no-build` if `./dist` already exists), writes `Dockerfile`, `nginx.conf`, and `.dockerignore` **only if they're missing**, then runs `docker build`. The generated nginx config already maps `application/wasm` and SPA-style `try_files`.

```sh
# Interactive — prompts for the image name:
gutter build deploy

# Image name inline:
gutter build deploy --image registry.example.com/myapp:v1
```

When it finishes, it prints the push and run commands:

```text
  docker push registry.example.com/myapp:v1
  # or run locally:
  docker run --rm -p 8080:80 registry.example.com/myapp:v1
```

The files are intentionally not overwritten if you've customized them — the CLI logs `kept existing Dockerfile` and moves on.

---

## Behind the scenes

| Step                    | What the CLI does                                                                            |
| ----------------------- | -------------------------------------------------------------------------------------------- |
| Building WASM           | `exec.Command("go", "build", "-o", "dist/app.wasm")` with `GOOS=js`, `GOARCH=wasm` set in the env. |
| Locating `wasm_exec.js` | Tries `$GOROOT/lib/wasm/wasm_exec.js` first (Go 1.24+), falls back to `$GOROOT/misc/wasm/`.  |
| Bundling                | Copies `index.html`, `wasm_exec.js`, and `./public/` (if present) into `./dist/`.            |
| Serving                 | `http.FileServer(http.Dir("dist"))`. The wasm MIME type is added via `mime.AddExtensionType`. |
| Watching                | `fsnotify.Watcher` recursive over your project; skips `.git`, `node_modules`, `dist`, `vendor`, dotfiles. |
| Live-reload injection   | Inserts a `<script>` before `</body>` of the served `index.html` that polls `/__gutter/build`. |

You can replicate any of this by hand if you'd rather drive the toolchain yourself:

```sh
mkdir -p dist
GOOS=js GOARCH=wasm go build -o dist/app.wasm .
cp "$(go env GOROOT)/lib/wasm/wasm_exec.js" dist/   # or misc/wasm for older Go
cp index.html dist/
cd dist && python3 -m http.server 8080              # serve, but doesn't set wasm MIME
```

The CLI's main value is consistent behavior across Go versions, the right MIME type, and live reload.

---

## Reference

| Command                | What it does                                                       |
| ---------------------- | ------------------------------------------------------------------ |
| `gutter new <name>`    | Scaffold a project: `main.go`, `index.html`, `go.mod`, `.gitignore`. |
| `gutter run`           | Bundle into `./dist/` and serve it on `:8080`.                     |
| `gutter run dev`       | Same as `run`, plus re-bundle + browser reload on `.go` / `.html` / `.css` changes. |
| `gutter build`         | Build a production-ready bundle into `./dist`.                     |
| `gutter build deploy`  | Build, generate Dockerfile + nginx.conf, run `docker build`.       |
| `gutter --version`     | Print the CLI version.                                             |
| `gutter --help`        | Print top-level help, or per-subcommand with `gutter <cmd> --help`. |
