// Package rpc is a typed, codegen-free RPC layer for Gutter's "write the whole
// web app in Go" model. Request/response types are plain Go structs defined
// ONCE in a package shared by client and server:
//
//	// package api (imported by both sides)
//	type AddRequest  struct{ A, B int }
//	type AddResponse struct{ Sum int }
//
//	// server (host): register a typed handler
//	rpc.Handle(func(ctx context.Context, r api.AddRequest) (api.AddResponse, error) {
//	    return api.AddResponse{Sum: r.A + r.B}, nil
//	})
//	mux.Handle(rpc.Endpoint, rpc.Handler())
//
//	// client (wasm): call it with full type safety
//	res, err := rpc.Call[api.AddRequest, api.AddResponse](ctx, api.AddRequest{A: 2, B: 3})
//
// The procedure route is derived from the request type (its package path +
// name), so the two sides agree automatically — there is no string to keep in
// sync, and changing a field in the shared struct is a compile error on BOTH
// sides. The client uses net/http, which on js/wasm transparently routes
// through the browser Fetch API, so the same Call code runs in the browser and
// on the host (and in tests).
//
// The package imports only the standard library; it does not depend on the
// gutter core, so it can be used independently.
package rpc

import (
	"context"
	"reflect"
)

// Endpoint is the URL the client POSTs to and where the server's Handler is
// expected to be mounted. The default is relative so it resolves against the
// page origin in the browser; override it (e.g. to an absolute URL) in tests or
// cross-origin setups.
var Endpoint = "/rpc"

// procHeader carries the derived procedure route on each request.
const procHeader = "X-Gutter-Proc"

// errEnvelope is the JSON body returned for a handler error.
type errEnvelope struct {
	Error string `json:"error"`
}

// invoker decodes a request payload, runs the handler, and encodes the result.
// Registered handlers are stored type-erased so the registry can be a plain map.
type invoker func(ctx context.Context, payload []byte) ([]byte, error)

// registry maps a procedure route to its invoker. It is populated by Handle on
// the server and read by Handler. The client never touches it.
var registry = map[string]invoker{}

// procName derives the stable route for a request type: its package path plus
// type name (e.g. "myapp/api.AddRequest"). Using (*T)(nil) means it works for
// any named type without needing a value.
func procName[T any]() string {
	t := reflect.TypeFor[T]()
	return t.PkgPath() + "." + t.Name()
}
