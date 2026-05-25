// ssrgen pre-renders the grid to HTML for each benchmark tier, so the SSR
// variant can be served statically. Each n{N}.html inlines the rendered grid in
// #app plus the WASM bootstrap; the runner loads it with ?n=N so the wasm builds
// the matching tree and hydrates.
package main

import (
	"fmt"
	"os"
	"path/filepath"

	"benchgutter/app"

	"github.com/Runway-Club/gutter"
)

const tmpl = `<!DOCTYPE html><html><head><meta charset="utf-8"><title>Gutter SSR grid</title>` +
	`<style>html,body{margin:0;font-family:system-ui,sans-serif}button{font:inherit}</style></head>` +
	`<body><div id="app">%s</div>` +
	`<script src="wasm_exec.js"></script>` +
	`<script>const go=new Go();WebAssembly.instantiateStreaming(fetch("app.wasm"),go.importObject).then(r=>go.run(r.instance));</script>` +
	`</body></html>`

func main() {
	tiers := []int{10, 100, 1000, 10000}
	for _, n := range tiers {
		html, err := gutter.RenderToHTML(app.Root(n))
		if err != nil {
			panic(err)
		}
		out := filepath.Join("dist-ssr", fmt.Sprintf("n%d.html", n))
		if err := os.WriteFile(out, []byte(fmt.Sprintf(tmpl, html)), 0o644); err != nil {
			panic(err)
		}
		fmt.Printf("wrote %s (%d bytes html)\n", out, len(html))
	}
}
