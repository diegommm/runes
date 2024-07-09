package runes

import (
	"fmt"
	"testing"

	"github.com/diegommm/runes/test"
)

func TestEmptyRange(t *testing.T) {
	t.Parallel()

	test.RangeInvariantTestCases{
		{"always empty", EmptyRange, nil, nil},
	}.Run(t)
}

func TestOneValueRange(t *testing.T) {
	t.Parallel()

	run := func(name string, f func(r rune) Range, testCases []rune) {
		t.Run(name, func(t *testing.T) {
			for _, tc := range testCases {
				test.TestRangeInvariants(t, &test.RangeInvariantTestCase{
					Name:          fmt.Sprintf("%d", tc),
					Range:         f(tc),
					AllValidRunes: []rune{tc},
				})
			}
		})
	}

	testCases1 := []rune{0, maxUint8}
	testCases2 := append(testCases1, maxUint8+1, maxUint16)
	testCases34 := append(testCases2, maxUint16+1, 1<<20)

	run("OneValueRange1", func(r rune) Range {
		return NewOneValueRange[OneValueRange1](r)
	}, testCases1)

	run("OneValueRange2", func(r rune) Range {
		return NewOneValueRange[OneValueRange2](r)
	}, testCases2)

	run("OneValueRange3", func(r rune) Range {
		return NewOneValueRange[OneValueRange3](r)
	}, testCases34)

	run("OneValueRange4", func(r rune) Range {
		return NewOneValueRange[OneValueRange4](r)
	}, testCases34)
}

func TestNewDynamicOneValueRange(t *testing.T) {
	t.Parallel()

	shouldType[OneValueRange1](NewDynamicOneValueRange(0))(t)
	shouldType[OneValueRange1](NewDynamicOneValueRange(maxUint8))(t)
	shouldType[OneValueRange2](NewDynamicOneValueRange(maxUint8 + 1))(t)
	shouldType[OneValueRange2](NewDynamicOneValueRange(maxUint16))(t)
	shouldType[OneValueRange3](NewDynamicOneValueRange(maxUint16 + 1))(t)
}

func TestSimpleRange(t *testing.T) {
	t.Parallel()

	run := func(name string, f func(*testing.T, rune, rune) Range, testCases [][2]rune) {
		t.Run(name, func(t *testing.T) {
			for _, tc := range testCases {
				test.TestRangeInvariants(t, &test.RangeInvariantTestCase{
					Name:          fmt.Sprintf("[%d,%d]", tc[0], tc[1]),
					Range:         f(t, tc[0], tc[1]),
					AllValidRunes: seq(tc[0], tc[1]),
				})
			}
		})
	}

	testCases1 := [][2]rune{
		{0, 0},
		{0, 1},
		{0, 2},
		{maxUint8 - 2, maxUint8},
		{maxUint8 - 1, maxUint8},
		{maxUint8, maxUint8},
	}

	testCases2 := [][2]rune{
		{maxUint8 - 10, maxUint8 + 10},
		{maxUint16 - 1, maxUint16},
		{maxUint16, maxUint16},
	}
	testCases2 = append(testCases1, testCases2...)

	testCases34 := [][2]rune{
		{maxUint16 - 10, maxUint16 + 10},
		{1 << 20, 1<<20 + 10},
	}
	testCases34 = append(testCases2, testCases34...)

	run("OneValueRange1", func(t *testing.T, from, to rune) Range {
		return withoutErr(NewSimpleRange[OneValueRange1](from, to))(t)
	}, testCases1)

	run("OneValueRange2", func(t *testing.T, from, to rune) Range {
		return withoutErr(NewSimpleRange[OneValueRange2](from, to))(t)
	}, testCases2)

	run("OneValueRange3", func(t *testing.T, from, to rune) Range {
		return withoutErr(NewSimpleRange[OneValueRange3](from, to))(t)
	}, testCases34)

	run("OneValueRange4", func(t *testing.T, from, to rune) Range {
		return withoutErr(NewSimpleRange[OneValueRange4](from, to))(t)
	}, testCases34)

	shouldErr(NewSimpleRange[OneValueRange1](1, 0))(t)
}

func TestNewDynamicSimpleRange(t *testing.T) {
	t.Parallel()
	var rr Range

	rr = withoutErr(NewDynamicSimpleRange(0, 0))(t)
	shouldType[SimpleRange[OneValueRange1]](rr)(t)

	rr = withoutErr(NewDynamicSimpleRange(0, maxUint8))(t)
	shouldType[SimpleRange[OneValueRange1]](rr)(t)

	rr = withoutErr(NewDynamicSimpleRange(0, maxUint8+1))(t)
	shouldType[SimpleRange[OneValueRange2]](rr)(t)

	rr = withoutErr(NewDynamicSimpleRange(0, maxUint16))(t)
	shouldType[SimpleRange[OneValueRange2]](rr)(t)

	rr = withoutErr(NewDynamicSimpleRange(0, maxUint16+1))(t)
	shouldType[SimpleRange[OneValueRange3]](rr)(t)
}

func TestRuneListRange(t *testing.T) {
	t.Parallel()

	run := func(t *testing.T, name string, f func(Iterator) Range, testCases [][]rune) {
		t.Run(name, func(t *testing.T) {
			for i, tc := range testCases {
				test.TestRangeInvariants(t, &test.RangeInvariantTestCase{
					Name:          fmt.Sprintf("[%d] len=%d", i, len(tc)),
					Range:         f(RunesIterator(tc)),
					AllValidRunes: tc,
				})
			}
		})
	}

	testCases1 := [][]rune{
		{},
		{0},
		{0, 1},
		{10, 20, 30, 40, 50, 60, 70},
		{0, maxUint8 - 1, maxUint8},
		{maxUint8 - 1, maxUint8},
		{maxUint8},
	}
	testCases2 := [][]rune{
		{0, maxUint8 - 1, maxUint8, maxUint8 + 1, maxUint16 - 1, maxUint16},
		{maxUint16 - 1, maxUint16},
		{maxUint16},
	}
	testCases2 = append(testCases1, testCases2...)
	testCases34 := [][]rune{
		{0, maxUint8 - 1, maxUint8, maxUint8 + 1, maxUint16 - 1, maxUint16,
			maxUint16 + 1, 1 << 20},
		{1 << 20},
	}
	testCases34 = append(testCases2, testCases34...)

	t.Run("RuneListRangeLinear", func(t *testing.T) {
		t.Parallel()

		run(t, "OneValueRange1", func(i Iterator) Range {
			return NewRuneListRange[RuneListRangeLinear[OneValueRange1]](i)
		}, testCases1)

		run(t, "OneValueRange2", func(i Iterator) Range {
			return NewRuneListRange[RuneListRangeLinear[OneValueRange2]](i)
		}, testCases2)

		run(t, "OneValueRange3", func(i Iterator) Range {
			return NewRuneListRange[RuneListRangeLinear[OneValueRange3]](i)
		}, testCases34)

		run(t, "OneValueRange4", func(i Iterator) Range {
			return NewRuneListRange[RuneListRangeLinear[OneValueRange4]](i)
		}, testCases34)
	})

	t.Run("RuneListRangeBinary", func(t *testing.T) {
		t.Parallel()

		run(t, "OneValueRange1", func(i Iterator) Range {
			return NewRuneListRange[RuneListRangeBinary[OneValueRange1]](i)
		}, testCases1)

		run(t, "OneValueRange2", func(i Iterator) Range {
			return NewRuneListRange[RuneListRangeBinary[OneValueRange2]](i)
		}, testCases2)

		run(t, "OneValueRange3", func(i Iterator) Range {
			return NewRuneListRange[RuneListRangeBinary[OneValueRange3]](i)
		}, testCases34)

		run(t, "OneValueRange4", func(i Iterator) Range {
			return NewRuneListRange[RuneListRangeBinary[OneValueRange4]](i)
		}, testCases34)
	})
}

func TestNewDynamicRuneListRange(t *testing.T) {
	t.Parallel()
	var rr Range

	// dumb cases
	rr = NewDynamicRuneListRange(RunesIterator(nil))
	shouldType[emptyRange](rr)(t)

	rr = NewDynamicRuneListRange(RunesIterator([]rune{0}))
	shouldType[OneValueRange1](rr)(t)

	// linear search cases
	rr = NewDynamicRuneListRange(RunesIterator(seq(0, 1)))
	shouldType[RuneListRangeLinear[OneValueRange1]](rr)(t)

	rr = NewDynamicRuneListRange(RunesIterator(seq(0, 1, maxUint16)))
	shouldType[RuneListRangeLinear[OneValueRange2]](rr)(t)

	rr = NewDynamicRuneListRange(RunesIterator(seq(0, 1, 1<<20)))
	shouldType[RuneListRangeLinear[OneValueRange3]](rr)(t)

	// binary search cases
	rr = NewDynamicRuneListRange(RunesIterator(seq(0, maxRuneListLinearSearch+1)))
	shouldType[RuneListRangeBinary[OneValueRange1]](rr)(t)

	rr = NewDynamicRuneListRange(RunesIterator(seq(0, maxRuneListLinearSearch, maxUint16)))
	shouldType[RuneListRangeBinary[OneValueRange2]](rr)(t)

	rr = NewDynamicRuneListRange(RunesIterator(seq(0, maxRuneListLinearSearch, 1<<20)))
	shouldType[RuneListRangeBinary[OneValueRange3]](rr)(t)
}

func TestExceptionRange(t *testing.T) {
	t.Parallel()

	one10 := withoutErr(NewDynamicSimpleRange(1, 10))(t)
	zero100 := withoutErr(NewDynamicSimpleRange(0, 100))(t)

	var xx Range
	xx = withoutErr(NewDynamicSimpleRange(1, 9))(t)
	xx = withoutErr(ExceptionRange(xx, NewOneValueRange[OneValueRange4](6)))(t)
	xx = withoutErr(ExceptionRange(xx, NewOneValueRange[OneValueRange4](2)))(t)
	xx = withoutErr(ExceptionRange(xx, NewOneValueRange[OneValueRange4](8)))(t)
	xx = withoutErr(ExceptionRange(xx, NewOneValueRange[OneValueRange4](4)))(t)
	xxEls := []rune{1, 3, 5, 7, 9}

	test.RangeInvariantTestCases{
		{"[0,100] except 1",
			withoutErr(ExceptionRange(zero100, NewDynamicOneValueRange(1)))(t),
			seq(2, 100, 0), nil},
		{"[0,100] except [1,10]",
			withoutErr(ExceptionRange(zero100, one10))(t),
			seq(11, 100, 0), nil},
		{"[1,9] except {2,4,6,8}", xx, xxEls, nil},
	}.Run(t)

	shouldErr(ExceptionRange(one10, EmptyRange))(t)
	shouldErr(ExceptionRange(zero100, NewDynamicOneValueRange(0)))(t)
	shouldErr(ExceptionRange(one10, zero100))(t)
}

func TestUniformRange(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		minRune   rune
		runeCount uint16
		stride    uint16
		expected  []rune
		invalid   []rune
	}{
		{"1 each 2", 0, 2, 2, []rune{0, 2}, []rune{1, 4}},
		{"1 each 2", 3, 2, 2, []rune{3, 5}, []rune{1, 7}},
		{"3 each 2", 3, 3, 2, []rune{3, 5, 7}, []rune{1, 9}},
		{"3 each 5", 3, 3, 5, []rune{3, 8, 13}, []rune{1, 14, 15, 16, 17, 18, 19}},
		{"3 each 7", 31, 3, 7, []rune{31, 38, 45}, []rune{25, 26, 27, 28, 29, 47, 48, 49, 50, 51, 52, 53}},
	}

	run := func(name string, f func(rune, uint16, uint16) (Range, error)) {
		t.Run("implem="+name, func(t *testing.T) {
			for _, tc := range testCases {
				test.TestRangeInvariants(t, &test.RangeInvariantTestCase{
					tc.name,
					withoutErr(f(tc.minRune, tc.runeCount, tc.stride))(t),
					tc.expected,
					tc.invalid,
				})
			}

			shouldErr(f(0, 1, 2))(t)
			shouldErr(f(0, 2, 1))(t)
		})
	}

	run("NewUniformRange5", func(minRune rune, runeCount, stride uint16) (Range, error) {
		return NewUniformRange5(minRune, runeCount, byte(stride))
	})

	run("NewUniformRange6", func(minRune rune, runeCount, stride uint16) (Range, error) {
		return NewUniformRange68[uint16](uint16(minRune), runeCount, stride)
	})

	run("NewUniformRange8", func(minRune rune, runeCount, stride uint16) (Range, error) {
		return NewUniformRange68[rune](minRune, runeCount, stride)
	})
}

func TestBitmapRange(t *testing.T) {
	t.Parallel()
	testCases := [][]rune{
		{},
		{0},
		{0, 5, 6},
		{1, 6, 7, 9},
		{0, 1, 5, 7, 90, 213, 256, 257, 258, 259, 512},
		{maxUint16, maxUint16 + 10, maxUint16 + 34, maxUint16 + 98},
		seq(100, 200),
	}

	for i, tc := range testCases {
		test.TestRangeInvariants(t, &test.RangeInvariantTestCase{
			fmt.Sprintf("%d", i),
			withoutErr(NewBitmapRange(RunesIterator(tc)))(t),
			tc,
			nil,
		})
	}
}
