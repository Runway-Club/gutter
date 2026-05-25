package gutter

import "testing"

func TestNewWorkerTaskRegistersAndLooksUp(t *testing.T) {
	tok := NewWorkerTask("test-reverse", func(s string) string { return s + "!" })
	if tok.Name != "test-reverse" {
		t.Fatalf("token Name = %q, want %q", tok.Name, "test-reverse")
	}
	if tok.URL != "app.wasm" {
		t.Fatalf("token URL = %q, want default %q", tok.URL, "app.wasm")
	}
	h := lookupWorkerTask("test-reverse")
	if h == nil {
		t.Fatal("lookupWorkerTask returned nil for a registered task")
	}
	if got := h("hi"); got != "hi!" {
		t.Fatalf("handler(%q) = %q, want %q", "hi", got, "hi!")
	}
}

func TestLookupUnknownWorkerTaskIsNil(t *testing.T) {
	if lookupWorkerTask("does-not-exist-xyz") != nil {
		t.Fatal("lookupWorkerTask for unknown name should be nil")
	}
}

func TestNewWorkerTaskDuplicatePanics(t *testing.T) {
	NewWorkerTask("dup-task", func(s string) string { return s })
	assertPanics(t, "duplicate name", func() {
		NewWorkerTask("dup-task", func(s string) string { return s })
	})
}

func TestNewWorkerTaskEmptyNamePanics(t *testing.T) {
	assertPanics(t, "empty name", func() {
		NewWorkerTask("", func(s string) string { return s })
	})
}

func TestNewWorkerTaskNilHandlerPanics(t *testing.T) {
	assertPanics(t, "nil handler", func() {
		NewWorkerTask("nil-handler-task", nil)
	})
}

func assertPanics(t *testing.T, what string, fn func()) {
	t.Helper()
	defer func() {
		if recover() == nil {
			t.Fatalf("expected panic for %s, got none", what)
		}
	}()
	fn()
}
