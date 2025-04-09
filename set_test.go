package runes

import (
	"encoding/binary"
	"fmt"
	"iter"
	"slices"
	"testing"
	"unicode/utf8"
)

const (
	maxUint8  = 1<<8 - 1
	maxUint32 = 1<<32 - 1
	maxInt32  = 1<<31 - 1

	maxRune = '\U0010FFFF'
)

type containsFunc func(rune) bool

func (f containsFunc) Contains(r rune) bool {
	return f(r)
}

var (
	always = containsFunc(func(rune) bool { return true })
	never  = containsFunc(func(rune) bool { return false })
)

func TestUnion(t *testing.T) {
	t.Parallel()
	type unionSets = Union[Set]

	testCases := []struct {
		set      Set
		expected bool
	}{
		{(Union[Set])(nil), false},
		{unionSets{}, false},
		{unionSets{never}, false},
		{unionSets{always}, true},
		{unionSets{never, always, never}, true},
		{unionSets{never, never, never}, false},
	}

	for i, tc := range testCases {
		got := tc.set.Contains('a')
		equals(t, tc.expected, got, "index=%v", i)
	}
}

type setTestCase struct {
	set                   Set
	contains, notContains iter.Seq[rune]
}

type setTestCases []setTestCase

func (tcs setTestCases) run(t *testing.T) {
	t.Helper()
	for i, tc := range tcs {
		t.Run(fmt.Sprintf("index=%v", i), func(t *testing.T) {
			t.Parallel()
			if tc.contains != nil {
				var count int
				for r := range tc.contains {
					count++
					equals(t, true, tc.set.Contains(r), "rune=0x%x", r)
				}
				t.Logf("verified %v items are contained", count)
			}
			if tc.notContains != nil {
				var count int
				for r := range tc.notContains {
					count++
					equals(t, false, tc.set.Contains(r), "rune=0x%x", r)
				}
				t.Logf("verified %v items are not contained", count)
			}
		})
	}
}

func TestBinarySlice(t *testing.T) {
	t.Parallel()
	testSlices(t, makeDynamicBinarySlice)
}

func TestLinearSlice(t *testing.T) {
	t.Parallel()
	testSlices(t, makeDynamicLinearSlice)
}

var (
	makeDynamicLinearSlice = newLowestSliceSet(true)
	makeDynamicBinarySlice = newLowestSliceSet(false)
)

func newLowestSliceSet(linear bool) func([]rune) Set {
	// could make this much simpler, just playing around with generics tbh
	return func(rs []rune) Set {
		if len(rs) == 0 {
			return newSliceSet[uint8](true, nil)
		}
		last := uint32(rs[len(rs)-1])
		switch {
		case last < 1<<8:
			return newSliceSet[uint8](linear, lower[uint8](rs))
		case last < 1<<16:
			return newSliceSet[uint16](linear, lower[uint16](rs))
		default:
			return newSliceSet[rune](linear, rs)
		}
	}
}

func newSliceSet[T RuneT](linear bool, rs []T) Set {
	if linear {
		return LinearSlice[T]{rs}
	}
	return BinarySlice[T]{rs}
}

func lower[T uint8 | uint16](rs []rune) []T {
	res := make([]T, len(rs))
	for i := range rs {
		res[i] = T(rs[i])
	}
	return res
}

func testSlices(t *testing.T, f func([]rune) Set) {
	t.Helper()
	setTestCases{
		{
			set:         f(nil),
			notContains: seq(-1, utf8.MaxRune),
		},
		{
			set:         f(slices.Collect(seq(utf8.MaxRune-10, 10))),
			contains:    seq(utf8.MaxRune-10, 10),
			notContains: seq(-1, utf8.MaxRune-10),
		},
		{
			set:         f(slices.Collect(seq(0, 10))),
			contains:    seq(0, 10),
			notContains: seq(11, utf8.MaxRune),
		},
		{
			set:         f([]rune{1, utf8.MaxRune}),
			contains:    slices.Values([]rune{1, utf8.MaxRune}), // too fast as well, just 2 items, even on the slow path
			notContains: seq(2, utf8.MaxRune-2),
		},
	}.run(t)
}

func TestInterval(t *testing.T) {
	t.Parallel()
	setTestCases{
		{
			set:         Interval[uint8]{},
			contains:    seq(0, 1),
			notContains: seq(1, utf8.MaxRune),
		},
		{
			set:         Interval[rune]{utf8.MaxRune, utf8.MaxRune},
			contains:    seq(utf8.MaxRune, 1),
			notContains: seq(-1, utf8.MaxRune),
		},
		{
			set:         Interval[uint16]{maxUint16 - 10, maxUint16},
			contains:    seq(maxUint16-10, 10),
			notContains: iters(seq(-1, maxUint16-10), seq(maxUint16+1, utf8.MaxRune-maxUint16-1)),
		},
	}.run(t)
}

func TestUniform(t *testing.T) {
	t.Parallel()
	setTestCases{
		{
			set:         Uniform[uint8, uint8, uint8]{},
			notContains: seq(-1, utf8.MaxRune),
		},
		{
			set:         Uniform[uint8, uint8, uint8]{maxUint8, maxUint8, maxUint8},
			contains:    seqs(maxUint8, maxUint8, maxUint8),
			notContains: except(seq(-1, utf8.MaxRune), seqs(maxUint8, maxUint8, maxUint8)),
		},
		{
			set:         Uniform[uint8, uint8, uint8]{3, 5, 7},
			contains:    slices.Values([]rune{3, 10, 17, 24, 31}),
			notContains: except(seq(-1, utf8.MaxRune), slices.Values([]rune{3, 10, 17, 24, 31})),
		},
	}.run(t)
}

func TestBitmap(t *testing.T) {
	t.Parallel()
	must(t, equals(t, true, utf8.MaxRune < 1<<24-1,
		"expected max rune to be representable in 3 bytes"))
	someRunes := []rune{1, 3, 99, 410}
	setTestCases{
		{
			set:         NewBitmap(nil),
			notContains: seq(-1, utf8.MaxRune),
		},
		{
			set:         NewBitmap([]rune{}),
			notContains: seq(-1, utf8.MaxRune),
		},
		{
			set:         NewBitmap(someRunes),
			contains:    slices.Values(someRunes),
			notContains: except(seq(-1, utf8.MaxRune), slices.Values(someRunes)),
		},
		{
			set:         NewBitmap([]rune{utf8.MaxRune}),
			contains:    slices.Values([]rune{utf8.MaxRune}),
			notContains: seq(-1, utf8.MaxRune-1),
		},
	}.run(t)
}

func TestCeilDiv(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		dividend, divisor, expected uint32
	}{
		{0, 1, 0},
		{2, 1, 2},
		{1023, 1024, 1},
		{1024, 1024, 1},
		{1025, 1024, 2},
		{maxUint32, maxUint32, 1},
	}

	for i, tc := range testCases {
		got := ceilDiv(tc.dividend, tc.divisor)
		equals(t, tc.expected, got, "index=%v; ceilDiv(%d, %d)", i, tc.dividend, tc.divisor)
	}
}

func TestU32Mid(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		a, b, expected uint32
	}{
		{0, 0, 0},
		{1, 0, 0},
		{0, 1, 0},
		{1, 1, 1},
		{1000, 2000, 1500},
		{1001, 2000, 1500},
		{1000, 2001, 1500},
		{0, maxUint32, 1<<31 - 1},
	}

	for i, tc := range testCases {
		got := u32Mid(tc.a, tc.b)
		equals(t, tc.expected, got, "index=%v; u32Mid(%d, %d)", i, tc.a, tc.b)
	}
}

func TestBitmapMinRuneEncoding(t *testing.T) {
	t.Parallel()
	testCases := []rune{'a', 'ñ', maxRune, '世', '界', 1, maxUint16, 1<<24 - 1}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("index=%d,value=0x%x", i, tc), func(t *testing.T) {
			var b [4]byte
			binary.LittleEndian.PutUint32(b[:], uint32(tc))
			equals(t, 0, b[bmHdrLen], "invalid test case: %v bytes max rune", bmHdrLen)

			expectedEnc := *(*[bmHdrLen]byte)(b[:])
			var gotEnc [bmHdrLen]byte
			bmEncodeMinRune(&gotEnc, tc)
			equals(t, expectedEnc, gotEnc, "bmEncodeMinRune of 0x%x", tc)

			gotRune := bmDecodeMinRune(expectedEnc[0], expectedEnc[1], expectedEnc[2])
			equals(t, tc, gotRune, "bmDecodeMinRune of %#v", expectedEnc)
		})
	}
}

func seq(from, count rune) iter.Seq[rune] {
	return seqs(from, count, 1)
}

func seqs(from, count, stride rune) iter.Seq[rune] {
	if count < 0 || stride < 1 {
		panic("count < 0 || stride < 1")
	}
	to := from + count*stride
	return func(yield func(rune) bool) {
		for i := from; i < to; i += stride {
			if !yield(i) {
				return
			}
		}
	}
}

func TestSeqs(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		from, count, stride rune
		expected            []rune
	}{
		{1, 0, 1, []rune{}},
		{1, 1, 1, []rune{1}},
		{3, 5, 7, []rune{3, 10, 17, 24, 31}},
	}

	for i, tc := range testCases {
		got := slices.Collect(seqs(tc.from, tc.count, tc.stride))
		equals(t, true, slices.Equal(tc.expected, got), "index=%v", i)
	}
}

func iters[T any](its ...iter.Seq[T]) iter.Seq[T] {
	return func(yield func(T) bool) {
		for _, it := range its {
			for v := range it {
				if !yield(v) {
					return
				}
			}
		}
	}
}

func except(s, x iter.Seq[rune]) iter.Seq[rune] {
	m := map[rune]struct{}{}
	for v := range x {
		m[v] = struct{}{}
	}
	if len(m) == 0 {
		return s
	}
	return func(yield func(rune) bool) {
		for v := range s {
			if _, ok := m[v]; !ok && !yield(v) {
				return
			}
		}
	}
}

func equals[T comparable](t *testing.T, expected T, got T, format string, args ...any) bool {
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

func must(t *testing.T, v bool) {
	t.Helper()
	if !v {
		t.FailNow()
	}
}
