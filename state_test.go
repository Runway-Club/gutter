package gutter

import "testing"

// fakeElem records how the State asked the framework to rebuild. SetState must
// go through the batched scheduleRebuild path, never the synchronous rebuild.
type fakeElem struct {
	rebuilds  int
	schedules int
}

func (f *fakeElem) rebuild()         { f.rebuilds++ }
func (f *fakeElem) scheduleRebuild() { f.schedules++ }

var _ stateElement = (*fakeElem)(nil)

func TestSetStateRunsMutationAndSchedules(t *testing.T) {
	var so StateObject
	fe := &fakeElem{}
	so.bindElement(fe)

	ran := false
	so.SetState(func() { ran = true })

	if !ran {
		t.Fatal("SetState did not run the mutation function")
	}
	if fe.schedules != 1 {
		t.Fatalf("scheduleRebuild called %d times, want 1", fe.schedules)
	}
	if fe.rebuilds != 0 {
		t.Fatalf("SetState must batch, not rebuild synchronously; rebuilds=%d", fe.rebuilds)
	}
}

func TestSetStateBeforeMountIsNoop(t *testing.T) {
	var so StateObject
	ran := false
	// No element bound yet (state created but not mounted). Must not panic.
	so.SetState(func() { ran = true })
	if !ran {
		t.Fatal("mutation should still run even before mount")
	}
}

func TestStateObjectWidgetBinding(t *testing.T) {
	var so StateObject
	if so.Widget() != nil {
		t.Fatalf("Widget() before bind = %v, want nil", so.Widget())
	}
	type myWidget struct{ n int }
	w := myWidget{n: 7}
	so.bindWidget(w)
	got, ok := so.Widget().(myWidget)
	if !ok || got.n != 7 {
		t.Fatalf("Widget() = %v, want %v", so.Widget(), w)
	}
}

func TestSetStateEachCallSchedules(t *testing.T) {
	// The scheduler dedups at the element level, but StateObject.SetState
	// itself should ask to be scheduled on every call — coalescing is the
	// runtime's job, not StateObject's.
	var so StateObject
	fe := &fakeElem{}
	so.bindElement(fe)
	so.SetState(func() {})
	so.SetState(func() {})
	so.SetState(func() {})
	if fe.schedules != 3 {
		t.Fatalf("scheduleRebuild called %d times across 3 SetState, want 3", fe.schedules)
	}
}
