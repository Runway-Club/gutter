# Testing Gutter

Gutter's tests are layered to match where the code actually runs. A change is
"safe to ship" only when the layer that exercises it is green.

| Layer | What it covers | Tooling | Command |
|-------|----------------|---------|---------|
| 1. Host unit | Platform-neutral logic: every widget's rendered `*gutter.Host` (tags/styles/attrs/children), theme color tokens, typography mapping, `Notifier`, `AssetURL`, options, `SetState` batching contract, theme presets | `go test` (no browser) | `go test ./...` |
| 2. WASM runtime | The reconciler in `element_wasm.go` against a **real DOM**: mount/update/unmount, attribute/style diffing, keyed + positional `reconcileChildren`, event dispatch + payload, **batched `SetState` coalescing**, dispose lifecycle | `wasmbrowsertest` (headless Chrome) | `GOOS=js GOARCH=wasm go test ./...` |
| 3. End-to-end | Full user flows through a built WASM app served by the gutter CLI: render, batched counter, controlled-input caret, keyed reorder identity, conditional mount/unmount | Playwright (real browser) | `cd e2e && npm test` |

Why three layers: most widget logic is pure CSS generation and is fastest and
most exhaustively tested on the host (layer 1). But the reconciler only exists
under `//go:build js && wasm` and needs a DOM, so it's tested in a browser
(layers 2 and 3). Layer 2 pokes the runtime's internals directly in Go; layer 3
proves the whole stack works the way a product built on Gutter would use it.

## Layer 1 — host unit tests

```sh
go test ./...              # fast
go test -race -cover ./... # what CI runs
```

No setup. These run anywhere `go` runs.

## Layer 2 — WASM runtime tests

These are normal `_test.go` files tagged `//go:build js && wasm` (e.g.
`element_wasm_test.go`). The Go toolchain runs a `GOOS=js GOARCH=wasm` test
binary through an exec wrapper named `go_js_wasm_exec`; we point that at
[`wasmbrowsertest`](https://github.com/agnivade/wasmbrowsertest), which loads
the binary into headless Chrome.

One-time setup:

```sh
go install github.com/agnivade/wasmbrowsertest@latest
cp "$(go env GOPATH)/bin/wasmbrowsertest" "$(go env GOPATH)/bin/go_js_wasm_exec"
# ensure $(go env GOPATH)/bin is on PATH
```

Then:

```sh
GOOS=js GOARCH=wasm go test -count=1 ./...
```

Requires a Chrome/Chromium binary on the machine (chromedp finds it
automatically). The harmless `Error: Go program has already exited` line printed
after a passing run is a wasmbrowsertest artifact, not a failure — trust the
`ok` / `PASS`.

## Layer 3 — end-to-end (Playwright)

The app under test is [`e2e/testapp`](e2e/testapp): a deterministic gutter app
whose every interactive surface has a stable selector. Playwright's config
builds the gutter CLI, has it build + serve the testapp on `:8080`
(`e2e/serve.sh`), then drives it.

```sh
cd e2e
npm install
npx playwright install chromium   # first time only
npm test
```

To watch it run: `npm run test:headed`.

## Adding tests

- **New widget?** Add a layer-1 test asserting its rendered `*gutter.Host`
  (see `widgets/*_test.go` and the `hostOf` helper). If it's a `StatefulWidget`
  or imperative (`_wasm.go`), cover it in layer 2 or 3 instead.
- **Reconciler/runtime change?** Add a layer-2 test in `element_wasm_test.go`.
- **New user-facing behavior?** Add a surface to `e2e/testapp` (with a
  `data-testid` via the `testID` helper) and a spec in `e2e/tests`.

CI (`.github/workflows/test.yml`) runs all three layers on every push and PR.
