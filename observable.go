package gutter

import "sync"

// Listenable is the read side of an observable value. Widgets subscribe via
// Listen and read the current snapshot via Value. The cancel function returned
// by Listen is idempotent — calling it more than once is safe.
//
// Listenable is the contract that ObserverBuilder consumes. Anything that can
// implement Value/Listen is observable, not just Notifier — adapters around
// channels, timers, or external sources work too.
type Listenable[T any] interface {
	Value() T
	Listen(fn func(T)) (cancel func())
}

// Notifier is the default Listenable: a value plus a set of listeners that
// fire on every Set or Update. Construct with NewNotifier; pass the pointer
// to widgets that need to react.
//
// Concurrent Set/Update/Listen are safe — the value is guarded by a mutex.
// Listener callbacks fire outside the lock so they can call Set on this or
// any other Notifier without deadlocking. WASM Go runs all goroutines on the
// single JS event loop, but the mutex still protects against reentrancy.
type Notifier[T any] struct {
	mu        sync.RWMutex
	value     T
	listeners map[int]func(T)
	nextID    int
}

// NewNotifier creates a Notifier seeded with initial.
func NewNotifier[T any](initial T) *Notifier[T] {
	return &Notifier[T]{value: initial}
}

// Value returns the current snapshot.
func (n *Notifier[T]) Value() T {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.value
}

// Set replaces the value and fires all listeners with the new value.
func (n *Notifier[T]) Set(v T) {
	n.mu.Lock()
	n.value = v
	listeners := n.snapshotListeners()
	n.mu.Unlock()
	for _, l := range listeners {
		l(v)
	}
}

// Update applies fn to the current value, stores the result, and fires all
// listeners with the new value. Useful for in-place mutation of collections.
func (n *Notifier[T]) Update(fn func(T) T) {
	n.mu.Lock()
	n.value = fn(n.value)
	v := n.value
	listeners := n.snapshotListeners()
	n.mu.Unlock()
	for _, l := range listeners {
		l(v)
	}
}

// Listen registers fn and returns an idempotent cancel function. Listeners
// are called in unspecified order; do not assume FIFO.
func (n *Notifier[T]) Listen(fn func(T)) func() {
	n.mu.Lock()
	if n.listeners == nil {
		n.listeners = map[int]func(T){}
	}
	id := n.nextID
	n.nextID++
	n.listeners[id] = fn
	n.mu.Unlock()
	return func() {
		n.mu.Lock()
		delete(n.listeners, id)
		n.mu.Unlock()
	}
}

// snapshotListeners copies the listener map into a slice so callbacks can
// fire without the lock held. Caller must hold n.mu.
func (n *Notifier[T]) snapshotListeners() []func(T) {
	out := make([]func(T), 0, len(n.listeners))
	for _, l := range n.listeners {
		out = append(out, l)
	}
	return out
}
