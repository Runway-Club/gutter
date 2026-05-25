//go:build js && wasm

package main

import (
	"benchssrdemo/app"

	"github.com/Runway-Club/gutter"
)

func main() {
	// WithHydrate adopts the server-rendered DOM when present (the SSR page) and
	// falls back to a fresh client mount when absent (the CSR page) — one main
	// for both.
	gutter.RunApp(app.Root(), gutter.WithHydrate())
}
