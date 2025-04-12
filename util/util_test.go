package util

import (
	"fmt"
	"iter"
	"slices"
	"testing"
	"unicode"
	"unicode/utf8"
)

func TestFormatSizeEstimation(t *testing.T) {
	Equal(t, "unknown", FormatSizeEstimation(0, false), "should be unknown")
	Equal(t, "10", FormatSizeEstimation(10, true), "should be known")
}

func TestContainsFunc(t *testing.T) {
	f := ContainsFunc(func(rune) bool {
		return true
	})
	Equal(t, true, f.Contains(-1), "unexpected set")
}

func TestRangeTableIter(t *testing.T) {
	testCases := []*unicode.RangeTable{
		unicode.White_Space,
		unicode.Upper,
		unicode.Lower,
		unicode.Letter,
		unicode.Mark,
		unicode.Number,
		unicode.Punct,
		unicode.Symbol,
	}

	for i, rt := range testCases {
		t.Run(fmt.Sprintf("index=%v", i), func(t *testing.T) {
			t.Parallel()
			next, stop := iter.Pull(RangeTableIter(rt))
			defer stop()
			v, ok := next()
			MustEqual(t, true, ok, "empty range table")
			for i := rune(0); i <= utf8.MaxRune; i++ {
				MustEqual(t, unicode.Is(rt, i), i == v, "0x%x", i)
				if i == v {
					v, _ = next()
				}
			}
			v, ok = next()
			MustEqual(t, false, ok, "unexpected extra rune 0x%x", v)
		})
	}
}

func TestSeq(t *testing.T) {
	testCases := []struct {
		lo, hi, stride rune
		panics         bool
		yields         []rune
	}{
		{1, 1, 0, true, nil},
		{1, 1, -1, true, nil},
		{1, 2, 3, true, nil},
		{2, 2, 1, false, []rune{2}},
		{2, 2, 500, false, []rune{2}},
		{2, 8, 1, false, []rune{2, 3, 4, 5, 6, 7, 8}},
		{2, 8, 2, false, []rune{2, 4, 6, 8}},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("index=%v", i), func(t *testing.T) {
			defer func() {
				Equal(t, tc.panics, recover() != nil, "panic expectation unmet")
			}()
			got := slices.Collect(Seq(tc.lo, tc.hi, tc.stride))
			Equal(t, true, slices.Equal(tc.yields, got), "want: %#v; got: %#v", tc.yields, got)
		})
	}
}

func TestConcat(t *testing.T) {
	testCases := []struct {
		its    []iter.Seq[rune]
		yields []rune
	}{
		{
			its:    []iter.Seq[rune]{},
			yields: []rune{},
		},
		{
			its: []iter.Seq[rune]{
				slices.Values([]rune{}),
			},
			yields: []rune{},
		},
		{
			its: []iter.Seq[rune]{
				slices.Values([]rune{1}),
			},
			yields: []rune{1},
		},
		{
			its: []iter.Seq[rune]{
				slices.Values([]rune{}),
				slices.Values([]rune{1, 2, 3}),
				slices.Values([]rune{4, 5}),
				slices.Values([]rune{}),
			},
			yields: []rune{1, 2, 3, 4, 5},
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("index=%v", i), func(t *testing.T) {
			got := slices.Collect(Concat(tc.its...))
			Equal(t, true, slices.Equal(tc.yields, got), "want: %#v; got: %#v", tc.yields, got)
		})
	}
}

func TestExcept(t *testing.T) {
	testCases := []struct {
		it, x  iter.Seq[rune]
		yields []rune
	}{
		{
			it:     slices.Values([]rune{}),
			x:      slices.Values([]rune{}),
			yields: []rune{},
		},
		{
			it:     slices.Values([]rune{}),
			x:      slices.Values([]rune{1, 2, 3}),
			yields: []rune{},
		},
		{
			it:     slices.Values([]rune{1, 2, 3}),
			x:      slices.Values([]rune{}),
			yields: []rune{1, 2, 3},
		},
		{
			it:     slices.Values([]rune{1, 2, 3, 4, 5, 6, 7}),
			x:      slices.Values([]rune{-1, 3, 4, 5, 99}),
			yields: []rune{1, 2, 6, 7},
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("index=%v", i), func(t *testing.T) {
			got := slices.Collect(Except(tc.it, tc.x))
			Equal(t, true, slices.Equal(tc.yields, got), "want: %#v; got: %#v", tc.yields, got)
		})
	}
}

func TestEqual(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		tt := new(tt)
		Equal(tt, true, true, "should not fail")
		if tt.failed || tt.calledFailNow {
			t.Fatalf("unexpected failure: %#v", tt)
		}
	})
	t.Run("fail", func(t *testing.T) {
		tt := new(tt)
		Equal(tt, true, false, "should fail")
		if !tt.failed || tt.calledFailNow {
			t.Fatalf("expected failure but not FailNow: %#v", tt)
		}
	})
}

func TestMustEqual(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		tt := new(tt)
		MustEqual(tt, true, true, "should not fail")
		if tt.failed || tt.calledFailNow {
			t.Fatalf("unexpected failure: %#v", tt)
		}
	})
	t.Run("fail", func(t *testing.T) {
		tt := new(tt)
		MustEqual(tt, true, false, "should fail")
		if !tt.failed || !tt.calledFailNow {
			t.Fatalf("expected failure: %#v", tt)
		}
	})
}

type tt struct {
	failed        bool
	calledFailNow bool
}

func (t *tt) Helper() {}

func (t *tt) FailNow() {
	t.failed = true
	t.calledFailNow = true
}

func (t *tt) Errorf(string, ...any) {
	t.failed = true
}
