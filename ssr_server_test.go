//go:build !js || !wasm

package gutter

import (
	"compress/gzip"
	"io"
	"net/http/httptest"
	"os"
	"path/filepath"
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

func TestSSRHandlerInjectsTreeHead(t *testing.T) {
	h, _ := SSRHandler(SSRConfig{Root: func() Widget {
		return Head{Title: "Injected", Meta: map[string]string{"description": "d"}, Child: ssrSrvBox{tag: "main", text: "x"}}
	}})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("GET", "/", nil))
	body := rec.Body.String()
	if !strings.Contains(body, "<title>Injected</title>") || !strings.Contains(body, `content="d"`) {
		t.Fatalf("doc missing injected head:\n%s", body)
	}
}

func TestSSRHandlerRequiresRoot(t *testing.T) {
	if _, err := SSRHandler(SSRConfig{}); err == nil {
		t.Fatal("expected error when Root is nil")
	}
}

// The SSR document must carry the same CSS reset the CSR index.html ships, or
// the browser's default 8px <body> margin shows up as stray padding.
func TestSSRDocHasMarginReset(t *testing.T) {
	h, _ := SSRHandler(SSRConfig{Root: func() Widget { return ssrSrvBox{tag: "main"} }})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("GET", "/", nil))
	if body := rec.Body.String(); !strings.Contains(body, "margin:0") || !strings.Contains(body, "100%;height:100%") {
		t.Fatalf("doc missing CSS reset:\n%s", body)
	}
}

func TestSSRHandlerGzipsHTML(t *testing.T) {
	h, _ := SSRHandler(SSRConfig{Root: func() Widget { return ssrSrvBox{tag: "main", text: "hi"} }})
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if enc := rec.Header().Get("Content-Encoding"); enc != "gzip" {
		t.Fatalf("Content-Encoding = %q, want gzip", enc)
	}
	if v := rec.Header().Get("Vary"); !strings.Contains(v, "Accept-Encoding") {
		t.Fatalf("Vary = %q, want Accept-Encoding", v)
	}
	if got := gunzip(t, rec.Body.Bytes()); !strings.Contains(got, "<main>hi</main>") {
		t.Fatalf("decompressed body missing content:\n%s", got)
	}
}

func TestServeStaticAssetGzipsWasmOnTheFly(t *testing.T) {
	dir := t.TempDir()
	want := strings.Repeat("wasmbytes\x00", 4096) // >1KB, compressible
	if err := os.WriteFile(filepath.Join(dir, "app.wasm"), []byte(want), 0o644); err != nil {
		t.Fatal(err)
	}
	h, _ := SSRHandler(SSRConfig{Dist: dir, Root: func() Widget { return ssrSrvBox{tag: "main"} }})

	req := httptest.NewRequest("GET", "/app.wasm", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if enc := rec.Header().Get("Content-Encoding"); enc != "gzip" {
		t.Fatalf("Content-Encoding = %q, want gzip", enc)
	}
	if ct := rec.Header().Get("Content-Type"); !strings.Contains(ct, "wasm") {
		t.Fatalf("Content-Type = %q, want application/wasm", ct)
	}
	if got := gunzip(t, rec.Body.Bytes()); got != want {
		t.Fatalf("decompressed wasm mismatch (%d vs %d bytes)", len(got), len(want))
	}
}

// A pre-compressed sibling (app.wasm.gz) must be served verbatim in preference
// to gzipping on the fly. We prove it by giving the sibling distinguishable
// contents and asserting those bytes come back.
func TestServeStaticAssetPrefersPrecompressedSibling(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "app.wasm"), []byte("RAW-uncompressed-source"), 0o644); err != nil {
		t.Fatal(err)
	}
	sentinel := "SIBLING-PRECOMPRESSED-PAYLOAD"
	var buf strings.Builder
	gz := gzip.NewWriter(&buf)
	io.WriteString(gz, sentinel)
	gz.Close()
	if err := os.WriteFile(filepath.Join(dir, "app.wasm.gz"), []byte(buf.String()), 0o644); err != nil {
		t.Fatal(err)
	}
	h, _ := SSRHandler(SSRConfig{Dist: dir, Root: func() Widget { return ssrSrvBox{tag: "main"} }})

	req := httptest.NewRequest("GET", "/app.wasm", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if enc := rec.Header().Get("Content-Encoding"); enc != "gzip" {
		t.Fatalf("Content-Encoding = %q, want gzip", enc)
	}
	if got := gunzip(t, rec.Body.Bytes()); got != sentinel {
		t.Fatalf("served %q, want the pre-compressed sibling %q", got, sentinel)
	}
}

func gunzip(t *testing.T, b []byte) string {
	t.Helper()
	zr, err := gzip.NewReader(strings.NewReader(string(b)))
	if err != nil {
		t.Fatalf("gzip.NewReader: %v", err)
	}
	out, err := io.ReadAll(zr)
	if err != nil {
		t.Fatalf("gunzip: %v", err)
	}
	return string(out)
}
