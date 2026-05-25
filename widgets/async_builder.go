package widgets

import (
	"context"
	"reflect"

	"github.com/Runway-Club/gutter"
)

// AsyncState is the lifecycle stage of an AsyncBuilder operation.
type AsyncState int

const (
	// AsyncPending is the initial state and remains until Load returns.
	AsyncPending AsyncState = iota
	// AsyncDone means Load returned a value with no error.
	AsyncDone
	// AsyncFailed means Load returned a non-nil error.
	AsyncFailed
)

// AsyncSnapshot is the per-rebuild view AsyncBuilder hands to its Builder.
// Inspect State first; Data is the zero value unless State == AsyncDone, and
// Error is nil unless State == AsyncFailed.
type AsyncSnapshot[T any] struct {
	State AsyncState
	Data  T
	Error error
}

// AsyncBuilder runs Load in a goroutine on mount and rebuilds Builder with
// the resulting snapshot when Load returns. Load receives a context.Context
// that is canceled when the widget unmounts, so long-running operations can
// abort cleanly.
//
//	widgets.AsyncBuilder[User]{
//	    Load: func(ctx context.Context) (User, error) {
//	        return fetchUser(ctx, id)
//	    },
//	    Builder: func(ctx *gutter.BuildContext, snap widgets.AsyncSnapshot[User]) gutter.Widget {
//	        switch snap.State {
//	        case widgets.AsyncPending: return widgets.Body{Text: "Loading…"}
//	        case widgets.AsyncFailed:  return widgets.Body{Text: snap.Error.Error()}
//	        }
//	        return widgets.Heading{Level: widgets.H2, Text: snap.Data.Name}
//	    },
//	}
//
// Load is invoked on mount and again whenever Deps changes across a parent
// rebuild. Go function values cannot be compared, so the framework cannot tell
// when Load itself has changed — list the inputs Load depends on (e.g. a
// resource ID) in Deps, and AsyncBuilder cancels the in-flight call, resets to
// AsyncPending, and re-runs Load when any of them change (compared with
// reflect.DeepEqual). Leave Deps nil to load exactly once per mount. Wrapping
// in widgets.WithKey still works as a heavier alternative (it remounts the
// whole subtree, discarding child state).
type AsyncBuilder[T any] struct {
	Load    func(ctx context.Context) (T, error)
	Builder func(ctx *gutter.BuildContext, snapshot AsyncSnapshot[T]) gutter.Widget
	// Deps are re-run triggers: when they change (DeepEqual), Load runs again.
	Deps []any
}

func (a AsyncBuilder[T]) CreateState() gutter.State {
	return &asyncState[T]{}
}

type asyncState[T any] struct {
	gutter.StateObject
	snapshot AsyncSnapshot[T]
	cancel   context.CancelFunc
	deps     []any
	resolved bool // set when ResolveSSR loaded synchronously on the server
}

func (s *asyncState[T]) widget() AsyncBuilder[T] {
	return s.Widget().(AsyncBuilder[T])
}

// ResolveSSR (gutter.SSRResolver) runs Load synchronously during server-side
// rendering so SSR emits the resolved UI instead of the pending placeholder.
// Called before InitState, which then skips spawning the async load.
func (s *asyncState[T]) ResolveSSR(ctx context.Context) {
	s.resolved = true
	s.deps = s.widget().Deps
	load := s.widget().Load
	if load == nil {
		s.snapshot = AsyncSnapshot[T]{State: AsyncDone}
		return
	}
	data, err := load(ctx)
	if err != nil {
		s.snapshot = AsyncSnapshot[T]{State: AsyncFailed, Error: err}
	} else {
		s.snapshot = AsyncSnapshot[T]{State: AsyncDone, Data: data}
	}
}

func (s *asyncState[T]) InitState() {
	if s.resolved {
		return // server already resolved Load synchronously (see ResolveSSR)
	}
	s.deps = s.widget().Deps
	s.start()
}

// start cancels any in-flight Load, resets to Pending, and launches Load again.
// Called on mount (InitState) and when Deps change (DidUpdateWidget). It sets
// snapshot synchronously so the rebuild that follows shows Pending immediately;
// the goroutine SetStates the result when Load returns.
func (s *asyncState[T]) start() {
	if s.cancel != nil {
		s.cancel()
		s.cancel = nil
	}
	load := s.widget().Load
	if load == nil {
		s.snapshot = AsyncSnapshot[T]{State: AsyncDone}
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	s.cancel = cancel
	s.snapshot = AsyncSnapshot[T]{State: AsyncPending}
	go func() {
		data, err := load(ctx)
		if ctx.Err() != nil {
			return // canceled (unmounted or superseded by a newer Deps) — drop result
		}
		s.SetState(func() {
			s.cancel = nil
			if err != nil {
				s.snapshot = AsyncSnapshot[T]{State: AsyncFailed, Error: err}
			} else {
				s.snapshot = AsyncSnapshot[T]{State: AsyncDone, Data: data}
			}
		})
	}()
}

// DidUpdateWidget re-runs Load when Deps changes. Per the WidgetUpdater
// contract a rebuild follows unconditionally, so start()'s synchronous reset to
// Pending is enough — no SetState needed here.
func (s *asyncState[T]) DidUpdateWidget(gutter.Widget) {
	next := s.widget().Deps
	if depsEqual(s.deps, next) {
		return
	}
	s.deps = next
	s.start()
}

func (s *asyncState[T]) Dispose() {
	if s.cancel != nil {
		s.cancel()
		s.cancel = nil
	}
}

// depsEqual compares two dependency lists element-wise with reflect.DeepEqual,
// which tolerates non-comparable elements (slices, maps) without panicking.
func depsEqual(a, b []any) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if !reflect.DeepEqual(a[i], b[i]) {
			return false
		}
	}
	return true
}

func (s *asyncState[T]) Build(ctx *gutter.BuildContext) gutter.Widget {
	w := s.widget()
	if w.Builder == nil {
		return nil
	}
	return w.Builder(ctx, s.snapshot)
}
