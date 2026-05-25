//go:build !js || !wasm

package gutter

import (
	"net/http/httptest"
	"strings"
	"testing"
)

type ssrSrvBox struct{ tag, text string }

func (b ssrSrvBox) Host() *Host { return &Host{Tag: b.tag, Text: b.text} }

func TestSSRHandlerRendersFullDoc(t *testing.T) {
	h, err := SSRHandler(SSRConfig{Root: func() Widget {
		return ssrSrvBox{tag: "main", text: "hello ssr"}
	}})
	if err != nil {
		t.Fatal(err)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("GET", "/", nil))
	if rec.Code != 200 {
		t.Fatalf("status = %d", rec.Code)
	}
	body := rec.Body.String()
	for _, want := range []string{
		"<!DOCTYPE html>", `<div id="app">`, "<main>hello ssr</main>",
		"wasm_exec.js", "instantiateStreaming",
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("response missing %q:\n%s", want, body)
		}
	}
	if ct := rec.Header().Get("Content-Type"); !strings.Contains(ct, "text/html") {
		t.Fatalf("content-type = %q", ct)
	}
}

func TestSSRHandlerRequiresRoot(t *testing.T) {
	if _, err := SSRHandler(SSRConfig{}); err == nil {
		t.Fatal("expected error when Root is nil")
	}
}
