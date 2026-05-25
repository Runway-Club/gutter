//go:build !js || !wasm

package gutter

import (
	"log"
	"net/http"
	"os"

	"github.com/Runway-Club/gutter/rpc"
)

// Serve runs the Config as an SSR server: it registers the RPC handlers, then
// serves server-rendered HTML at "/", the RPC endpoint at rpc.Endpoint, and the
// wasm assets from Dist (for hydration). It blocks. Addr/Dist fall back to the
// GUTTER_ADDR/GUTTER_DIST env vars (set by `gutter run --ssr`) then Config defaults.
func Serve(a Config) {
	if a.Root == nil {
		panic("gutter: Config.Root is required")
	}
	h, err := serveHandler(a)
	if err != nil {
		log.Fatal(err)
	}
	addr := firstNonEmpty(os.Getenv("GUTTER_ADDR"), a.Addr, ":8080")
	log.Printf("gutter SSR server on http://localhost%s", addr)
	log.Fatal(http.ListenAndServe(addr, h))
}

// serveHandler builds the combined RPC + SSR + static handler. Split out so it
// can be exercised in tests without binding a port.
func serveHandler(a Config) (http.Handler, error) {
	if a.RPC != nil {
		a.RPC()
	}
	ssr, err := SSRHandler(SSRConfig{
		Root:  a.Root,
		Theme: a.Theme,
		Head:  a.Head,
		Dist:  firstNonEmpty(os.Getenv("GUTTER_DIST"), a.Dist, "dist"),
	})
	if err != nil {
		return nil, err
	}
	mux := http.NewServeMux()
	mux.Handle(rpc.Endpoint, rpc.Handler())
	mux.Handle("/", ssr)
	return mux, nil
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}
