package widgets

import "github.com/Runway-Club/gutter"

// ObserverBuilder rebuilds its child whenever the Source Listenable fires.
// Use it to bind a leaf subtree to a Notifier so the rest of the tree does
// not have to know about the observable.
//
//	count := gutter.NewNotifier(0)
//	widgets.ObserverBuilder[int]{
//	    Source: count,
//	    Builder: func(ctx *gutter.BuildContext, n int) gutter.Widget {
//	        return widgets.Heading{Level: widgets.H1, Text: fmt.Sprintf("%d", n)}
//	    },
//	}
//
// If an ancestor rebuild replaces Source with a different Listenable, the
// subscription is swapped over to the new one (the old one is unsubscribed).
type ObserverBuilder[T any] struct {
	Source  gutter.Listenable[T]
	Builder func(ctx *gutter.BuildContext, value T) gutter.Widget
}

func (o ObserverBuilder[T]) CreateState() gutter.State {
	return &observerState[T]{}
}

type observerState[T any] struct {
	gutter.StateObject
	cancel func()
	source gutter.Listenable[T]
}

func (s *observerState[T]) widget() ObserverBuilder[T] {
	return s.Widget().(ObserverBuilder[T])
}

func (s *observerState[T]) subscribe(src gutter.Listenable[T]) {
	if src == nil {
		return
	}
	s.source = src
	s.cancel = src.Listen(func(_ T) {
		s.SetState(func() {})
	})
}

func (s *observerState[T]) InitState() {
	s.subscribe(s.widget().Source)
}

func (s *observerState[T]) DidUpdateWidget(_ gutter.Widget) {
	next := s.widget().Source
	if next == s.source {
		return
	}
	if s.cancel != nil {
		s.cancel()
		s.cancel = nil
	}
	s.subscribe(next)
}

func (s *observerState[T]) Dispose() {
	if s.cancel != nil {
		s.cancel()
		s.cancel = nil
	}
}

func (s *observerState[T]) Build(ctx *gutter.BuildContext) gutter.Widget {
	w := s.widget()
	if w.Source == nil || w.Builder == nil {
		return nil
	}
	return w.Builder(ctx, w.Source.Value())
}
