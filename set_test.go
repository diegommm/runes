package runes

import (
	"encoding/binary"
	"fmt"
	"slices"
	"testing"
	"unicode/utf8"

	"github.com/diegommm/runes/util"
)

const (
	maxUint8  = 1<<8 - 1
	maxUint32 = 1<<32 - 1
	maxInt32  = 1<<31 - 1

	maxRune = '\U0010FFFF'
)

var (
	_ util.OrderedListIter[rune] = OrderedRunesIter(nil)
	_ util.OrderedList[rune]     = OrderedRunesList(nil)
	_ OrderedRunesIter           = util.OrderedListIter[rune](nil)
	_ OrderedRunesList           = util.OrderedList[rune](nil)
)

func runes(rs ...rune) OrderedRunesIter {
	return util.SliceList[rune](rs).Iter()
}

func seq(from rune, count int, stride ...rune) OrderedRunesIter {
	s := rune(1)
	if len(stride) == 1 {
		s = stride[0]
	}
	return util.Seq[rune]{from, count, s}.Iter()
}

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
		util.Equal(t, tc.expected, got, "index=%v", i)
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
			return newSliceSet[uint8](linear, reduceWidth[uint8](rs))
		case last < 1<<16:
			return newSliceSet[uint16](linear, reduceWidth[uint16](rs))
		default:
			return newSliceSet[rune](linear, rs)
		}
	}
}

func newSliceSet[T RuneT](linear bool, rs []T) Set {
	if linear {
		return LinearSlice[T](rs)
	}
	return BinarySlice[T](rs)
}

func reduceWidth[T uint8 | uint16](rs []rune) []T {
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
			set:         f(util.Collect(seq(utf8.MaxRune-10, 10))),
			contains:    seq(utf8.MaxRune-10, 10),
			notContains: seq(-1, utf8.MaxRune-10),
		},
		{
			set:         f(util.Collect(seq(0, 10))),
			contains:    seq(0, 10),
			notContains: seq(11, utf8.MaxRune),
		},
		{
			set:         f([]rune{1, utf8.MaxRune}),
			contains:    runes(1, utf8.MaxRune),
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
			notContains: util.IterMerge(seq(-1, maxUint16-10), seq(maxUint16+1, utf8.MaxRune-maxUint16-1)),
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
			contains:    seq(maxUint8, maxUint8, maxUint8),
			notContains: util.IterExcept(seq(-1, utf8.MaxRune), seq(maxUint8, maxUint8, maxUint8)),
		},
		{
			set:         Uniform[uint8, uint8, uint8]{3, 5, 7},
			contains:    runes(3, 10, 17, 24, 31),
			notContains: util.IterExcept(seq(-1, utf8.MaxRune), runes(3, 10, 17, 24, 31)),
		},
	}.run(t)
}

func TestBitmap(t *testing.T) {
	t.Parallel()
	util.MustEqual(t, true, utf8.MaxRune < 1<<24-1,
		"expected max rune to be representable in 3 bytes")
	someRunes := []rune{1, 3, 99, 410}
	setTestCases{
		{
			set:         NewBitmap(util.SliceList[rune](nil)),
			notContains: seq(-1, utf8.MaxRune),
		},
		{
			set:         NewBitmap(util.SliceList[rune]([]rune{})),
			notContains: seq(-1, utf8.MaxRune),
		},
		{
			set:         NewBitmap(util.SliceList[rune](someRunes)),
			contains:    runes(someRunes...),
			notContains: util.IterExcept(seq(-1, utf8.MaxRune), runes(someRunes...)),
		},
		{
			set:         NewBitmap(util.SliceList[rune]([]rune{utf8.MaxRune})),
			contains:    runes(utf8.MaxRune),
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
		util.Equal(t, tc.expected, got, "index=%v; ceilDiv(%d, %d)", i, tc.dividend, tc.divisor)
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
		util.Equal(t, tc.expected, got, "index=%v; u32Mid(%d, %d)", i, tc.a, tc.b)
	}
}

func TestBitmapMinRuneEncoding(t *testing.T) {
	t.Parallel()
	testCases := []rune{'a', 'ñ', maxRune, '世', '界', 1, maxUint16, 1<<24 - 1}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("index=%d,value=0x%x", i, tc), func(t *testing.T) {
			var b [4]byte
			binary.LittleEndian.PutUint32(b[:], uint32(tc))
			util.Equal(t, 0, b[bmHdrLen], "invalid test case: %v bytes max rune", bmHdrLen)

			expectedEnc := *(*[bmHdrLen]byte)(b[:])
			var gotEnc [bmHdrLen]byte
			bmEncodeMinRune(&gotEnc, tc)
			util.Equal(t, expectedEnc, gotEnc, "bmEncodeMinRune of 0x%x", tc)

			gotRune := bmDecodeMinRune(expectedEnc[0], expectedEnc[1], expectedEnc[2])
			util.Equal(t, tc, gotRune, "bmDecodeMinRune of %#v", expectedEnc)
		})
	}
}

func TestSeqs(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		from     rune
		count    int
		stride   rune
		expected []rune
	}{
		{1, 0, 1, []rune{}},
		{1, 1, 1, []rune{1}},
		{3, 5, 7, []rune{3, 10, 17, 24, 31}},
	}

	for i, tc := range testCases {
		got := util.Collect(seq(tc.from, tc.count, tc.stride))
		util.Equal(t, true, slices.Equal(tc.expected, got), "index=%v", i)
	}
}

type setTestCase struct {
	set                   Set
	contains, notContains OrderedRunesIter
}

type setTestCases []setTestCase

func (tcs setTestCases) run(t *testing.T) {
	t.Helper()
	for i, tc := range tcs {
		t.Run(fmt.Sprintf("index=%v", i), func(t *testing.T) {
			t.Parallel()
			if tc.contains != nil {
				var count int
				for r := range util.IterSeq(tc.contains) {
					count++
					util.Equal(t, true, tc.set.Contains(r), "rune=0x%x", r)
				}
				t.Logf("verified %v items are contained", count)
			}
			if tc.notContains != nil {
				var count int
				for r := range util.IterSeq(tc.notContains) {
					count++
					util.Equal(t, false, tc.set.Contains(r), "rune=0x%x", r)
				}
				t.Logf("verified %v items are not contained", count)
			}
		})
	}
}
