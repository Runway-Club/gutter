package rpc

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
)

// Handle registers a typed handler for a procedure keyed by its request type.
// Call it once per procedure at server startup. Registering two handlers for
// the same request type panics — the route would be ambiguous.
func Handle[Req any, Res any](fn func(context.Context, Req) (Res, error)) {
	route := procName[Req]()
	if _, dup := registry[route]; dup {
		panic("gutter/rpc: duplicate handler for " + route)
	}
	registry[route] = func(ctx context.Context, payload []byte) ([]byte, error) {
		var req Req
		if len(payload) > 0 {
			if err := json.Unmarshal(payload, &req); err != nil {
				return nil, err
			}
		}
		res, err := fn(ctx, req)
		if err != nil {
			return nil, err
		}
		return json.Marshal(res)
	}
}

// Handler returns the http.Handler that serves every registered procedure.
// Mount it at Endpoint (default "/rpc"), e.g.:
//
//	mux.Handle(rpc.Endpoint, rpc.Handler())
func Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "gutter/rpc: POST required", http.StatusMethodNotAllowed)
			return
		}
		proc := r.Header.Get(procHeader)
		inv, ok := registry[proc]
		if !ok {
			writeErr(w, http.StatusNotFound, "unknown procedure: "+proc)
			return
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			writeErr(w, http.StatusBadRequest, "read body: "+err.Error())
			return
		}
		out, err := inv(r.Context(), body)
		if err != nil {
			// Handler-returned errors are surfaced to the client verbatim.
			writeErr(w, http.StatusInternalServerError, err.Error())
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(out)
	})
}

func writeErr(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(errEnvelope{Error: msg})
}
