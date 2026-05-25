//go:build !js || !wasm

// These tests spin a real HTTP server via httptest, which needs a TCP listener
// — unavailable on js/wasm. The client (Call) and server code compile on wasm
// (verified by `GOOS=js GOARCH=wasm go build ./rpc`); the round-trip logic is
// exercised here on the host, and end-to-end in browser via examples/fullstack.
package rpc_test

import (
	"context"
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/Runway-Club/gutter/rpc"
)

// Shared request/response types — in a real app these live in a package
// imported by both the wasm client and the host server.
type addReq struct{ A, B int }
type addRes struct{ Sum int }

type boomReq struct{ X int }
type boomRes struct{ Y int }

type unregReq struct{ Z int }
type unregRes struct{ W int }

type dupReq struct{ N int }
type dupRes struct{ N int }

func TestCallRoundTrip(t *testing.T) {
	rpc.Handle(func(_ context.Context, r addReq) (addRes, error) {
		return addRes{Sum: r.A + r.B}, nil
	})
	srv := httptest.NewServer(rpc.Handler())
	defer srv.Close()
	rpc.Endpoint = srv.URL

	res, err := rpc.Call[addReq, addRes](context.Background(), addReq{A: 2, B: 40})
	if err != nil {
		t.Fatalf("Call: %v", err)
	}
	if res.Sum != 42 {
		t.Fatalf("Sum = %d, want 42", res.Sum)
	}
}

func TestCallPropagatesHandlerError(t *testing.T) {
	rpc.Handle(func(_ context.Context, r boomReq) (boomRes, error) {
		return boomRes{}, errors.New("kaboom")
	})
	srv := httptest.NewServer(rpc.Handler())
	defer srv.Close()
	rpc.Endpoint = srv.URL

	_, err := rpc.Call[boomReq, boomRes](context.Background(), boomReq{X: 1})
	if err == nil || err.Error() != "kaboom" {
		t.Fatalf("err = %v, want \"kaboom\"", err)
	}
}

func TestCallUnknownProcedure(t *testing.T) {
	srv := httptest.NewServer(rpc.Handler())
	defer srv.Close()
	rpc.Endpoint = srv.URL

	// unregReq was never Handle'd → server returns 404 → Call errors.
	if _, err := rpc.Call[unregReq, unregRes](context.Background(), unregReq{Z: 9}); err == nil {
		t.Fatal("expected error for unregistered procedure")
	}
}

func TestDuplicateHandlerPanics(t *testing.T) {
	rpc.Handle(func(_ context.Context, r dupReq) (dupRes, error) { return dupRes{}, nil })
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic on duplicate Handle for the same request type")
		}
	}()
	rpc.Handle(func(_ context.Context, r dupReq) (dupRes, error) { return dupRes{}, nil })
}
