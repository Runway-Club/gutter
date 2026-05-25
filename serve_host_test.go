//go:build !js || !wasm

package gutter

import (
	"context"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Runway-Club/gutter/rpc"
)

type serveBox struct{ tag, text string }

func (b serveBox) Host() *Host { return &Host{Tag: b.tag, Text: b.text} }

type serveReq struct{ N int }
type serveRes struct{ Double int }

func TestServeHandlerServesSSRAndRPC(t *testing.T) {
	h, err := serveHandler(Config{
		Root: func() Widget { return serveBox{tag: "main", text: "served"} },
		RPC: func() {
			rpc.Handle(func(_ context.Context, r serveReq) (serveRes, error) {
				return serveRes{Double: r.N * 2}, nil
			})
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	// "/" → server-rendered HTML.
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("GET", "/", nil))
	if !strings.Contains(rec.Body.String(), "<main>served</main>") {
		t.Fatalf("SSR body missing rendered content:\n%s", rec.Body.String())
	}

	// rpc.Endpoint → the registered handler, reachable on the same server.
	srv := httptest.NewServer(h)
	defer srv.Close()
	rpc.Endpoint = srv.URL + "/rpc"
	res, err := rpc.Call[serveReq, serveRes](context.Background(), serveReq{N: 21})
	if err != nil {
		t.Fatalf("rpc.Call: %v", err)
	}
	if res.Double != 42 {
		t.Fatalf("Double = %d, want 42", res.Double)
	}
}
