package gutter

import "github.com/Runway-Club/gutter/themes"

// Config is the single declarative entry point for a Gutter application. The SAME
// Config value drives both the WebAssembly client and (in SSR mode) the host
// server — gutter.Serve does the right thing for whichever target the binary
// was compiled into, so you write ONE main with no build tags:
//
//	func main() {
//	    gutter.Serve(gutter.Config{Root: Root})
//	}
//
// `gutter run` serves it client-side (CSR); `gutter run --ssr` builds the wasm,
// then runs this same program on the host as an SSR server that renders Root()
// per request, mounts the RPC handlers, and serves the wasm assets for
// hydration. Add server procedures once via RPC:
//
//	func main() {
//	    gutter.Serve(gutter.Config{Root: Root, RPC: func() {
//	        rpc.Handle(func(ctx context.Context, r AddReq) (AddRes, error) { ... })
//	    }})
//	}
//
// RPC only runs on the server. (Handlers referenced here are still linked into
// the wasm client; keep heavy server-only logic in host-tagged files if bundle
// size matters.)
type Config struct {
	// Root builds the widget tree. Required. Called on the client to mount/
	// hydrate, and on the server to render HTML.
	Root func() Widget
	// RPC registers rpc handlers (typically a few rpc.Handle calls). Optional;
	// invoked once on the server before it starts, never on the client.
	RPC func()
	// Theme is forwarded to both the client and SSR rendering.
	Theme *themes.Theme
	// Selector is the client mount point; defaults to "#app".
	Selector string
	// Addr is the SSR server listen address; defaults to ":8080". The
	// GUTTER_ADDR env var (set by `gutter run --ssr`) takes precedence.
	Addr string
	// Dist is the SSR server's static asset dir; defaults to "dist". The
	// GUTTER_DIST env var takes precedence.
	Dist string
	// Head is extra raw HTML injected into the <head> of SSR pages.
	Head string
}
