//go:build !js || !wasm

package gutter

// A minimal production-minded SSR server: it renders Root() to HTML on every
// request (instant first paint + SEO) and serves the static WASM assets that
// take over via hydration. The app's wasm main should call
// RunApp(Root(), WithHydrate()) so the same Root() drives both sides.

import (
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"

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
	`<meta name="viewport" content="width=device-width, initial-scale=1">` +
	`<style>html,body{margin:0;padding:0;width:100%%;height:100%%;font-family:Lexend,system-ui,sans-serif}#app{width:100%%;height:100%%}</style>%s</head>` +
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
				serveStaticAsset(w, r, p, st)
				return
			}
		}
		// Otherwise render the app. Unknown paths fall through to SSR so
		// client-side routes still get server-rendered HTML. Head hints from
		// gutter.Head widgets in the tree are appended after cfg.Head (so the
		// app's title/meta can override static fonts/favicon links).
		treeHead, body, err := RenderDocument(cfg.Root(), opts...)
		if err != nil {
			http.Error(w, "gutter SSR render error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		doc := fmt.Sprintf(ssrDocTmpl, cfg.Head+treeHead, body)
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Vary", "Accept-Encoding")
		if acceptsEncoding(r.Header.Get("Accept-Encoding"), "gzip") {
			w.Header().Set("Content-Encoding", "gzip")
			gz := gzip.NewWriter(w)
			io.WriteString(gz, doc)
			gz.Close()
			return
		}
		io.WriteString(w, doc)
	}), nil
}

// serveStaticAsset serves one file from Dist with content negotiation. The big
// payload is app.wasm — gzip cuts its transfer ~3-4x — so for compressible types
// we (1) prefer a pre-compressed sibling written at build time (best ratio, zero
// per-request CPU), (2) else gzip on the fly, (3) else serve plain. Stable
// filenames (no content hash) can't be marked immutable, so Cache-Control:
// no-cache asks the browser to revalidate — the Last-Modified conditional then
// returns 304 (no re-download) while unchanged, yet picks up a rebuild instantly.
func serveStaticAsset(w http.ResponseWriter, r *http.Request, path string, st os.FileInfo) {
	ext := strings.ToLower(filepath.Ext(path))
	w.Header().Set("Cache-Control", "no-cache")

	if !compressibleByExt(ext) {
		http.ServeFile(w, r, path)
		return
	}
	w.Header().Set("Vary", "Accept-Encoding")
	ae := r.Header.Get("Accept-Encoding")

	// (1) A sibling pre-compressed at build time. Brotli beats gzip, so prefer it.
	for _, enc := range []struct{ name, suffix string }{{"br", ".br"}, {"gzip", ".gz"}} {
		if !acceptsEncoding(ae, enc.name) {
			continue
		}
		if f, err := os.Open(path + enc.suffix); err == nil {
			defer f.Close()
			w.Header().Set("Content-Encoding", enc.name)
			w.Header().Set("Content-Type", contentTypeByExt(ext))
			// ServeContent keeps Range/304 working; modTime is the source file's
			// so revalidation tracks the asset the client actually sees.
			http.ServeContent(w, r, filepath.Base(path), st.ModTime(), f)
			return
		}
	}
	// (2) No sibling — gzip on the fly when the client accepts it.
	if acceptsEncoding(ae, "gzip") {
		f, err := os.Open(path)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer f.Close()
		w.Header().Set("Content-Encoding", "gzip")
		w.Header().Set("Content-Type", contentTypeByExt(ext))
		gz := gzip.NewWriter(w)
		_, _ = io.Copy(gz, f)
		gz.Close()
		return
	}
	// (3) Client can't take gzip — serve the raw bytes.
	http.ServeFile(w, r, path)
}

// compressibleByExt reports whether a file extension is worth gzipping. Already-
// compressed formats (png/jpg/woff2/…) are deliberately excluded.
func compressibleByExt(ext string) bool {
	switch ext {
	case ".wasm", ".js", ".mjs", ".css", ".html", ".htm", ".json", ".svg", ".xml", ".txt", ".map":
		return true
	}
	return false
}

func contentTypeByExt(ext string) string {
	if ct := mime.TypeByExtension(ext); ct != "" {
		return ct
	}
	return "application/octet-stream"
}

// acceptsEncoding reports whether an Accept-Encoding header offers enc with a
// non-zero quality. Crude but covers the real cases ("gzip, deflate, br" and
// "gzip;q=1.0"); an explicit "enc;q=0" disable is honored.
func acceptsEncoding(header, enc string) bool {
	for _, part := range strings.Split(header, ",") {
		part = strings.TrimSpace(part)
		name := part
		if i := strings.IndexByte(part, ';'); i >= 0 {
			name = strings.TrimSpace(part[:i])
			if q := part[i:]; strings.Contains(q, "q=0") && !strings.Contains(q, "q=0.") {
				continue // q=0 → explicitly not acceptable
			}
		}
		if name == enc || name == "*" {
			return true
		}
	}
	return false
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
