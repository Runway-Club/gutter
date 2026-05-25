// ssrgen renders Root() to HTML on the host and assembles two pages:
//
//	dist-csr/  — empty #app, boots WASM (Gutter's current client-render flow)
//	dist-ssr/  — pre-rendered HTML inlined in #app
//
// Both pages boot the same app.wasm (built with WithHydrate): the CSR page has
// an empty #app so WASM mounts fresh; the SSR page has pre-rendered HTML so WASM
// hydrates it. FCP is recorded at HTML parse for SSR (instant) vs after WASM
// boot for CSR.
package main

import (
	"fmt"
	"os"
	"path/filepath"

	"benchssrdemo/app"

	"github.com/Runway-Club/gutter"
)

const tmpl = `<!DOCTYPE html><html><head><meta charset="utf-8"><title>SSR demo</title>` +
	`<style>html,body{margin:0;font-family:system-ui,sans-serif}#app{width:100%%}</style>` +
	`</head><body><div id="app">%s</div>%s</body></html>`

const boot = `<script src="wasm_exec.js"></script>` +
	`<script>const go=new Go();WebAssembly.instantiateStreaming(fetch("app.wasm"),go.importObject).then(r=>go.run(r.instance));</script>`

func main() {
	html, err := gutter.RenderToHTML(app.Root())
	if err != nil {
		panic(err)
	}
	write("dist-csr/index.html", fmt.Sprintf(tmpl, "", boot))
	write("dist-ssr/index.html", fmt.Sprintf(tmpl, html, boot))
	fmt.Printf("rendered %d bytes of SSR HTML\n", len(html))
}

func write(path, content string) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		panic(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		panic(err)
	}
}
