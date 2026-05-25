// Reference SSR server for the demo — the project layout `gutter run --ssr`
// expects. It renders the shared app.Root() per request and serves the wasm
// assets from dist/ for hydration. GUTTER_ADDR / GUTTER_DIST are set by the
// CLI; defaults make `go run ./server` work standalone too.
package main

import (
	"log"
	"os"

	"benchssrdemo/app"

	"github.com/Runway-Club/gutter"
	"github.com/Runway-Club/gutter/themes"
)

func main() {
	addr := os.Getenv("GUTTER_ADDR")
	if addr == "" {
		addr = ":8080"
	}
	dist := os.Getenv("GUTTER_DIST")
	if dist == "" {
		dist = "dist"
	}
	log.Printf("gutter SSR demo on http://localhost%s (assets from %s)", addr, dist)
	err := gutter.ServeSSR(gutter.SSRConfig{
		Addr:  addr,
		Dist:  dist,
		Theme: themes.Apple,
		Root:  app.Root,
	})
	if err != nil {
		log.Fatal(err)
	}
}
