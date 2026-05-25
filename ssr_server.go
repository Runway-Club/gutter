//go:build !js || !wasm

package gutter

// A minimal production-minded SSR server: it renders Root() to HTML on every
// request (instant first paint + SEO) and serves the static WASM assets that
// take over via hydration. The app's wasm main should call
// RunApp(Root(), WithHydrate()) so the same Root() drives both sides.

import (
	"errors"
	"fmt"
	"mime"
	"net/http"
	"os"
	"path/filepath"

	"github.com/Runway-Club/gutter/themes"
)

// SSRConfig configures an SSR server. Root is required; the rest have defaults.
type SSRConfig struct {
	Addr  string        // listen address; default ":8080"
	Root  func() Widget // builds a fresh widget tree per request (required)
	Dist  string        // dir holding app.wasm + wasm_exec.js; default "dist"
	Head  string        // extra raw HTML injected into <head> (fonts, meta, …)
	Theme *themes.Theme // optional; forwarded to RenderToHTML
}

const ssrDocTmpl = `<!DOCTYPE html><html><head><meta charset="utf-8">` +
	`<meta name="viewport" content="width=device-width, initial-scale=1">%s</head>` +
	`<body><div id="app">%s</div>` +
	`<script src="wasm_exec.js"></script>` +
	`<script>const go=new Go();WebAssembly.instantiateStreaming(fetch("app.wasm"),go.importObject).then(r=>go.run(r.instance));</script>` +
	`</body></html>`

// SSRHandler returns the http.Handler used by ServeSSR. Exposed so it can be
// mounted in an existing mux or exercised in tests.
func SSRHandler(cfg SSRConfig) (http.Handler, error) {
	if cfg.Root == nil {
		return nil, errors.New("gutter: SSRHandler requires SSRConfig.Root")
	}
	if cfg.Dist == "" {
		cfg.Dist = "dist"
	}
	mime.AddExtensionType(".wasm", "application/wasm")

	var opts []Option
	if cfg.Theme != nil {
		opts = append(opts, WithTheme(cfg.Theme))
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Serve a real static asset if one exists at this path (app.wasm,
		// wasm_exec.js, css, …). filepath.Clean on a rooted path can't escape Dist.
		if r.URL.Path != "/" {
			p := filepath.Join(cfg.Dist, filepath.Clean("/"+r.URL.Path))
			if st, err := os.Stat(p); err == nil && !st.IsDir() {
				http.ServeFile(w, r, p)
				return
			}
		}
		// Otherwise render the app. Unknown paths fall through to SSR so
		// client-side routes still get server-rendered HTML.
		html, err := RenderToHTML(cfg.Root(), opts...)
		if err != nil {
			http.Error(w, "gutter SSR render error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprintf(w, ssrDocTmpl, cfg.Head, html)
	}), nil
}

// ServeSSR builds the handler and blocks serving it on cfg.Addr (default
// ":8080"). Returns the ListenAndServe error.
func ServeSSR(cfg SSRConfig) error {
	if cfg.Addr == "" {
		cfg.Addr = ":8080"
	}
	h, err := SSRHandler(cfg)
	if err != nil {
		return err
	}
	return http.ListenAndServe(cfg.Addr, h)
}
