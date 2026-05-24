package widgets

import (
	"context"

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
// Load is invoked exactly once per mount. Go function values cannot be
// compared, so the framework cannot tell when Load itself has changed across
// a parent rebuild. To force a fresh invocation (e.g. when the resource ID
// changes), wrap the AsyncBuilder in widgets.WithKey with a key derived from
// the inputs — that causes the old subtree to unmount and a new one to mount.
type AsyncBuilder[T any] struct {
	Load    func(ctx context.Context) (T, error)
	Builder func(ctx *gutter.BuildContext, snapshot AsyncSnapshot[T]) gutter.Widget
}

func (a AsyncBuilder[T]) CreateState() gutter.State {
	return &asyncState[T]{}
}

type asyncState[T any] struct {
	gutter.StateObject
	snapshot AsyncSnapshot[T]
	cancel   context.CancelFunc
}

func (s *asyncState[T]) widget() AsyncBuilder[T] {
	return s.Widget().(AsyncBuilder[T])
}

func (s *asyncState[T]) InitState() {
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
			return
		}
		s.SetState(func() {
			if err != nil {
				s.snapshot = AsyncSnapshot[T]{State: AsyncFailed, Error: err}
			} else {
				s.snapshot = AsyncSnapshot[T]{State: AsyncDone, Data: data}
			}
		})
	}()
}

func (s *asyncState[T]) Dispose() {
	if s.cancel != nil {
		s.cancel()
		s.cancel = nil
	}
}

func (s *asyncState[T]) Build(ctx *gutter.BuildContext) gutter.Widget {
	w := s.widget()
	if w.Builder == nil {
		return nil
	}
	return w.Builder(ctx, s.snapshot)
}
