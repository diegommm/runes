package runes

import (
	"fmt"
	"testing"
	"unsafe"

	"github.com/diegommm/runes/test"
)

func (x twoRange[R]) ByteLen() (int, bool) {
	s0, ok0 := EstimateByteLen(x[0])
	s1, ok1 := EstimateByteLen(x[1])
	return s0 + s1, ok0 && ok1
}

func (x bsRange[R]) ByteLen() (l int, exact bool) {
	l, exact = int(unsafe.Sizeof(x)), true
	for i := range x {
		lx, okx := EstimateByteLen(x[i])
		l += lx
		exact = exact && okx
	}
	return
}

func (x uniformRange5) ByteLen() (int, bool) {
	return int(unsafe.Sizeof(x)), true
}

func TestEmptyRange(t *testing.T) {
	t.Parallel()

	test.RangeInvariantTestCases{
		{"always empty", EmptyRange, nil},
	}.Run(t)
}

func TestOneValueRange(t *testing.T) {
	t.Parallel()

	run := func(name string, f func(r rune) Range, testCases []rune) {
		t.Run(name, func(t *testing.T) {
			for _, tc := range testCases {
				test.TestRangeInvariants(t, &test.RangeInvariantTestCase{
					Name:  fmt.Sprintf("%d", tc),
					Range: f(tc),
					Runes: []rune{tc},
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

	if _, ok := NewDynamicOneValueRange(0).(OneValueRange1); !ok {
		t.Fatalf("invalid type")
	}
	if _, ok := NewDynamicOneValueRange(maxUint8).(OneValueRange1); !ok {
		t.Fatalf("invalid type")
	}
	if _, ok := NewDynamicOneValueRange(maxUint8 + 1).(OneValueRange2); !ok {
		t.Fatalf("invalid type")
	}
	if _, ok := NewDynamicOneValueRange(maxUint16).(OneValueRange2); !ok {
		t.Fatalf("invalid type")
	}
	if _, ok := NewDynamicOneValueRange(maxUint16 + 1).(OneValueRange3); !ok {
		t.Fatalf("invalid type")
	}
}

func TestSimpleRange(t *testing.T) {
	t.Parallel()

	run := func(name string, f func(rune, rune) Range, testCases [][2]rune) {
		t.Run(name, func(t *testing.T) {
			for _, tc := range testCases {
				test.TestRangeInvariants(t, &test.RangeInvariantTestCase{
					Name:  fmt.Sprintf("[%d,%d]", tc[0], tc[1]),
					Range: f(tc[0], tc[1]),
					Runes: seq[rune](tc[0], tc[1]),
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
	testCases2 = append(testCases2, testCases1...)

	testCases34 := [][2]rune{
		{maxUint16 - 10, maxUint16 + 10},
		{1 << 20, 1<<20 + 10},
	}
	testCases34 = append(testCases34, testCases2...)

	run("OneValueRange1", func(from, to rune) Range {
		return Must(NewSimpleRange[OneValueRange1](from, to))
	}, testCases1)

	run("OneValueRange2", func(from, to rune) Range {
		return Must(NewSimpleRange[OneValueRange2](from, to))
	}, testCases2)

	run("OneValueRange3", func(from, to rune) Range {
		return Must(NewSimpleRange[OneValueRange3](from, to))
	}, testCases34)

	run("OneValueRange4", func(from, to rune) Range {
		return Must(NewSimpleRange[OneValueRange4](from, to))
	}, testCases34)
}

func TestNewDynamicSimpleRange(t *testing.T) {
	t.Parallel()

	_, ok := Must(NewDynamicSimpleRange(0, 0)).(SimpleRange[OneValueRange1])
	if !ok {
		t.Fatalf("invalid type")
	}
	_, ok = Must(NewDynamicSimpleRange(0, maxUint8)).(SimpleRange[OneValueRange1])
	if !ok {
		t.Fatalf("invalid type")
	}
	_, ok = Must(NewDynamicSimpleRange(0, maxUint8+1)).(SimpleRange[OneValueRange2])
	if !ok {
		t.Fatalf("invalid type")
	}
	_, ok = Must(NewDynamicSimpleRange(0, maxUint16)).(SimpleRange[OneValueRange2])
	if !ok {
		t.Fatalf("invalid type")
	}
	_, ok = Must(NewDynamicSimpleRange(0, maxUint16+1)).(SimpleRange[OneValueRange3])
	if !ok {
		t.Fatalf("invalid type")
	}
}

func TestUniformRange(t *testing.T) {
	t.Parallel()
	test.RangeInvariantTestCases{
		{"1 each 1", Must(NewUniformRange5(0, 1, 1)), []rune{0}},
		{"1 each 1", Must(NewUniformRange5(3, 1, 1)), []rune{3}},
		{"3 each 1", Must(NewUniformRange5(3, 3, 1)), []rune{3, 4, 5}},
		{"3 each 5", Must(NewUniformRange5(3, 3, 5)), []rune{3, 8, 13}},
		{"3 each 7", Must(NewUniformRange5(31, 3, 7)), []rune{31, 38, 45}},
	}.Run(t)
}
