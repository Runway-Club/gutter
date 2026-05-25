//go:build js && wasm

package gutter

// Serve runs the Config in the browser: it mounts (or hydrates, when the page was
// server-rendered) Root into the configured selector. Config.RPC is ignored here —
// those procedures run on the server; the client reaches them via rpc.Call.
func Serve(a Config) {
	if a.Root == nil {
		panic("gutter: Config.Root is required")
	}
	opts := []Option{WithHydrate()}
	if a.Theme != nil {
		opts = append(opts, WithTheme(a.Theme))
	}
	if a.Selector != "" {
		opts = append(opts, WithSelector(a.Selector))
	}
	RunApp(a.Root(), opts...)
}
