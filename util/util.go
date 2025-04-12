package util

import (
	"iter"
	"strconv"
	"unicode"
)

// FormatSizeEstimation formats the output of [Sizeof].
func FormatSizeEstimation(s uintptr, ok bool) string {
	if !ok {
		return "unknown"
	}
	return strconv.FormatUint(uint64(s), 10)
}

// ContainsFunc is an adapter for the `Contains` method of `Set`. This is mostly
// useful for testing or in PoCs.
type ContainsFunc func(rune) bool

func (f ContainsFunc) Contains(r rune) bool {
	return f(r)
}

// RangeTableIter returns an iterator over the runes of the given range table.
func RangeTableIter(rt *unicode.RangeTable) iter.Seq[rune] {
	ss := make([]iter.Seq[rune], 0, len(rt.R16)+len(rt.R32))
	for _, r := range rt.R16 {
		ss = append(ss, Seq(rune(r.Lo), rune(r.Hi), rune(r.Stride)))
	}
	for _, r := range rt.R32 {
		ss = append(ss, Seq(rune(r.Lo), rune(r.Hi), rune(r.Stride)))
	}
	return Concat(ss...)
}

// Seq returns an iterator over the runes starting at `lo` and ending in `hi`,
// `stride` apart from each other. It panics if the stride is invalid. It is
// valid to have hi<lo, in which case the iterator will not yield values.
func Seq(lo, hi, stride rune) iter.Seq[rune] {
	if stride < 1 || (hi-lo)%stride != 0 {
		panic("invalid stride")
	}
	return func(yield func(rune) bool) {
		for v := lo; v <= hi; v += stride {
			yield(v)
		}
	}
}

// Concat concatenates the values returned by the given iterators.
func Concat(its ...iter.Seq[rune]) iter.Seq[rune] {
	return func(yield func(rune) bool) {
		for _, it := range its {
			for r := range it {
				yield(r)
			}
		}
	}
}

// Except yields the values of `it` unless they are in `x`.
func Except(it, x iter.Seq[rune]) iter.Seq[rune] {
	m := make(map[rune]struct{})
	for r := range x {
		m[r] = struct{}{}
	}
	if len(m) == 0 {
		return it
	}
	return func(yield func(rune) bool) {
		for r := range it {
			if _, ok := m[r]; !ok {
				yield(r)
			}
		}
	}
}

type TestingT interface {
	Helper()
	FailNow()
	Errorf(string, ...any)
}

func Equal[T comparable](t TestingT, expected T, got T, format string, args ...any) bool {
	t.Helper()
	if expected == got {
		return true
	}
	t.Errorf(format, args...)
	t.Errorf("\tExpected: %#v", expected)
	t.Errorf("\tGot:      %#v", got)
	return false
}

func MustEqual[T comparable](t TestingT, expected T, got T, format string, args ...any) {
	t.Helper()
	if !Equal(t, expected, got, format, args...) {
		t.FailNow()
	}
}
