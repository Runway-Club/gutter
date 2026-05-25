package gutter

// Ambient dependency injection, à la Flutter's InheritedWidget. A Provider[T]
// makes a value of type T available to its whole subtree; any descendant reads
// it with DependOn[T](ctx) — no passing pointers down by hand. Typical uses: an
// app-wide store, the RPC client, the router, feature flags.
//
//	gutter.Provider[*Store]{Value: store, Child: ...app...}
//
//	// deep inside the subtree:
//	store, ok := gutter.DependOn[*Store](ctx)
//
// Lookups are by exact type. To provide two values of the same underlying type,
// define distinct named types. Providing the same type again deeper in the tree
// shadows the outer value for that subtree.
//
// Note: there is no fine-grained dependency invalidation yet — changing a
// Provider's Value updates descendants on the next top-down rebuild of that
// Provider (e.g. when its owner SetStates), not via targeted notifications.
// For values that change often, pair a Provider holding a Notifier[T] with an
// ObserverBuilder.

import "reflect"

// inheritedProvider is implemented by Provider[T]. The runtime and the SSR
// renderer detect it to push a richer scope while building the provider's
// subtree. Unexported so the scope representation stays internal.
type inheritedProvider interface {
	provideInto(parent map[reflect.Type]any) map[reflect.Type]any
}

// Provider makes Value (of type T) available to Child's subtree via DependOn[T].
// It is a StatelessWidget that simply renders Child; the runtime handles the
// scope.
type Provider[T any] struct {
	Value T
	Child Widget
}

// Build renders the child; the provided value is injected by the runtime.
func (p Provider[T]) Build(*BuildContext) Widget { return p.Child }

func (p Provider[T]) provideInto(parent map[reflect.Type]any) map[reflect.Type]any {
	m := make(map[reflect.Type]any, len(parent)+1)
	for k, v := range parent {
		m[k] = v
	}
	m[reflect.TypeFor[T]()] = p.Value
	return m
}

// DependOn returns the nearest ancestor Provider[T]'s value and true, or the
// zero value and false if no Provider[T] is in scope. It is safe to call at any
// time during a Build — including isolated SetState rebuilds — because the
// runtime restores each element's scope before rebuilding it.
func DependOn[T any](ctx *BuildContext) (T, bool) {
	var zero T
	if ctx == nil || ctx.inherited == nil {
		return zero, false
	}
	v, ok := ctx.inherited[reflect.TypeFor[T]()]
	if !ok {
		return zero, false
	}
	t, ok := v.(T)
	return t, ok
}
