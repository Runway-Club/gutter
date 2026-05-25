// One main for the whole full-stack app — no build tags, no separate server.
// `gutter run` serves it client-side; `gutter run --ssr` builds the wasm and
// runs this same program as the SSR server (rendering app.Root, mounting the
// RPC handler, serving assets for hydration). gutter.Serve does the right thing
// for whichever target the binary was compiled into.
package main

import (
	"context"

	"fullstackexample/api"
	"fullstackexample/app"

	"github.com/Runway-Club/gutter"
	"github.com/Runway-Club/gutter/rpc"
	"github.com/Runway-Club/gutter/themes"
)

func main() {
	gutter.Serve(gutter.Config{
		Root:  app.Root,
		Theme: themes.Apple,
		// Registered once, runs only on the server. The shared api.AddRequest
		// type keys it; the client's rpc.Call reaches it with no extra wiring.
		RPC: func() {
			rpc.Handle(func(_ context.Context, r api.AddRequest) (api.AddResponse, error) {
				return api.AddResponse{Sum: r.A + r.B}, nil
			})
		},
	})
}
