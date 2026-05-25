package gutter

import (
	"sync"
	"testing"
)

func TestNotifierValue(t *testing.T) {
	n := NewNotifier(42)
	if got := n.Value(); got != 42 {
		t.Fatalf("Value() = %d, want 42", got)
	}
}

func TestNotifierSetFiresListeners(t *testing.T) {
	n := NewNotifier("a")
	var got string
	var calls int
	n.Listen(func(v string) { got = v; calls++ })
	n.Set("b")
	if n.Value() != "b" {
		t.Fatalf("Value() = %q, want %q", n.Value(), "b")
	}
	if got != "b" || calls != 1 {
		t.Fatalf("listener got %q after %d calls, want %q after 1", got, calls, "b")
	}
}

func TestNotifierUpdate(t *testing.T) {
	n := NewNotifier(10)
	var seen int
	n.Listen(func(v int) { seen = v })
	n.Update(func(v int) int { return v + 5 })
	if n.Value() != 15 || seen != 15 {
		t.Fatalf("after Update: value=%d seen=%d, want 15/15", n.Value(), seen)
	}
}

func TestNotifierMultipleListeners(t *testing.T) {
	n := NewNotifier(0)
	var a, b int
	n.Listen(func(v int) { a = v })
	n.Listen(func(v int) { b = v })
	n.Set(7)
	if a != 7 || b != 7 {
		t.Fatalf("listeners a=%d b=%d, want both 7", a, b)
	}
}

func TestNotifierCancelStopsListener(t *testing.T) {
	n := NewNotifier(0)
	var calls int
	cancel := n.Listen(func(int) { calls++ })
	n.Set(1)
	cancel()
	n.Set(2)
	if calls != 1 {
		t.Fatalf("listener fired %d times, want 1 (cancel should stop it)", calls)
	}
}

func TestNotifierCancelIdempotent(t *testing.T) {
	n := NewNotifier(0)
	cancel := n.Listen(func(int) {})
	cancel()
	// A second cancel must be a safe no-op, not a panic.
	cancel()
}

func TestNotifierConcurrentSet(t *testing.T) {
	// The mutex must let concurrent Set/Listen run without a data race
	// (run with -race to be meaningful). We only assert it doesn't deadlock
	// or panic and that the final value is one of the written values.
	n := NewNotifier(0)
	n.Listen(func(int) {})
	var wg sync.WaitGroup
	for i := 1; i <= 50; i++ {
		wg.Add(1)
		go func(v int) { defer wg.Done(); n.Set(v) }(i)
	}
	wg.Wait()
	if v := n.Value(); v < 1 || v > 50 {
		t.Fatalf("final value %d out of written range [1,50]", v)
	}
}

// Notifier must satisfy the Listenable interface it claims to implement.
var _ Listenable[int] = (*Notifier[int])(nil)
