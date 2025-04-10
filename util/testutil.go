package util

import (
	"cmp"
	"fmt"
	"slices"
)

type TestingT interface {
	Helper()
	Fail()
	FailNow()
	Logf(string, ...any)
	Errorf(string, ...any)
}

func Equal[T comparable](t TestingT, expected T, got T, format string, args ...any) bool {
	t.Helper()
	if expected == got {
		return true
	}
	t.Logf(format, args...)
	t.Logf("\tExpected: %#v", expected)
	t.Logf("\tGot:      %#v", got)
	t.Fail()
	return false
}

func MustEqual[T comparable](t TestingT, expected T, got T, format string, args ...any) {
	t.Helper()
	if !Equal(t, expected, got, format, args...) {
		t.FailNow()
	}
}

func Recover(t TestingT, format string, args ...any) {
	MustEqual(t, fmt.Sprint(nil), fmt.Sprint(recover()), format, args...)
}

func TestOrderedList[T cmp.Ordered](t TestingT, rl OrderedList[T], vs ...T) {
	t.Helper()
	// we don't allow repetition and slices.IsSorted doesn't either because it
	// uses cmp.Less
	MustEqual(t, true, slices.IsSorted(vs), "input slice is not sorted")
	defer Recover(t, "list panicked")

	if len(vs) > 0 {
		Equal(t, vs[0], rl.Min(), "invalid Min")
		Equal(t, vs[len(vs)-1], rl.Max(), "invalid Max")
	} else {
		var zero T
		Equal(t, zero, rl.Min(), "invalid Min")
		Equal(t, zero, rl.Max(), "invalid Max")
	}
	Equal(t, len(vs), rl.Len(), "invalid Len")
	TestOrderedListIter[T](t, rl.Iter(), vs...)
}

func TestOrderedListIter[T cmp.Ordered](t TestingT, s OrderedListIter[T], vs ...T) {
	t.Helper()
	// we don't allow repetition and slices.IsSorted doesn't either because it
	// uses cmp.Less
	MustEqual(t, true, slices.IsSorted(vs), "input slice is not sorted")
	defer Recover(t, "iterator panicked")

	var zero T
	i := 0
	for v, ok := s(); ; v, ok = s() {
		if !ok {
			Equal(t, zero, v, "iterator exhausted and did not return zero value")
			break
		}
		MustEqual(t, true, i < len(vs), "index=%v; too many values in iterator", i)
		MustEqual(t, vs[i], v, "index=%v; values don't match")
		i++
	}
	Equal(t, len(vs), i, "too few values in iterator")

	for i := 0; i < 10; i++ {
		v, ok := s()
		MustEqual(t, zero, v, "index=%v; iterator returned non-zero value after exhaustion", i)
		MustEqual(t, false, ok, "index=%v; iterator returned true after exhaustion", i)
	}
}
