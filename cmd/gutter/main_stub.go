//go:build js && wasm

// The gutter CLI is a host-only developer tool. Under js/wasm we provide a
// no-op stub so 'GOOS=js GOARCH=wasm go build ./...' still succeeds.
package main

func main() {}
