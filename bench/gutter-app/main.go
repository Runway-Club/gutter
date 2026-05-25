//go:build js && wasm

// WASM client for the grid benchmark. Reads the item count from ?n= and mounts
// (CSR) or hydrates (SSR) the shared app.Root tree. WithHydrate makes one binary
// serve both the CSR pages (empty #app → mount) and the SSR pages (pre-rendered
// #app → hydrate).
package main

import (
	"strconv"
	"syscall/js"

	"benchgutter/app"

	"github.com/Runway-Club/gutter"
)

func itemCount() int {
	params := js.Global().Get("URLSearchParams").New(js.Global().Get("location").Get("search"))
	v := params.Call("get", "n")
	if v.IsNull() {
		return 100
	}
	n, err := strconv.Atoi(v.String())
	if err != nil || n < 0 {
		return 100
	}
	return n
}

func main() {
	gutter.RunApp(app.Root(itemCount()), gutter.WithHydrate())
}
