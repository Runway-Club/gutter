// Package api holds the request/response types shared by the wasm client and
// the host server. Change a field here and BOTH sides fail to compile — that is
// the whole point of the typed RPC layer.
package api

type AddRequest struct {
	A int `json:"a"`
	B int `json:"b"`
}

type AddResponse struct {
	Sum int `json:"sum"`
}
