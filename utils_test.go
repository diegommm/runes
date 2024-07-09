package runes

import (
	"math/bits"
	"slices"
	"testing"
)

const maxUint32 = 1<<32 - 1

func withoutErr[T any](v T, err error) func(*testing.T) T {
	return func(t *testing.T) T {
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		return v
	}
}

func shouldErr[T any](_ T, err error) func(*testing.T) {
	return func(t *testing.T) {
		t.Helper()
		if err == nil {
			t.Fatalf("expected error")
		}
	}
}

func shouldType[T any](v any) func(*testing.T) {
	return func(t *testing.T) {
		t.Helper()
		x, ok := v.(T)
		if !ok {
			t.Fatalf("expected type %T, got %T", x, v)
		}
	}
}

// seq generates an inclusive sequence of runes starting at `start` and ending
// in `end`. If given, the runes in `extra` are later added and the final slice
// is returned sorted in ascending order.
func seq(from, to rune, extra ...rune) []rune {
	if from > to {
		panic("from cannot be greater than to")
	}
	s := make([]rune, 0, to+1-from+rune(len(extra)))
	for i := from; i <= to; i++ {
		s = append(s, i)
	}
	s = append(s, extra...)
	slices.Sort(s)
	return s
}

func TestCeilDiv(t *testing.T) {
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

	for _, tc := range testCases {
		actual := ceilDiv(tc.dividend, tc.divisor)
		if tc.expected != actual {
			t.Errorf("ceilDiv(%d, %d)=%d, expected=%d", tc.dividend, tc.divisor,
				actual, tc.expected)
		}
	}
}

func TestU32Mid(t *testing.T) {
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

	for _, tc := range testCases {
		actual := u32Mid(tc.a, tc.b)
		if tc.expected != actual {
			t.Errorf("u32Mid(%d, %d)=%d, expected=%d", tc.a, tc.b, actual,
				tc.expected)
		}
	}
}

func TestOnes(t *testing.T) {
	for i := 0; i < 256; i++ {
		b := byte(i)
		if int(ones(b)) != bits.OnesCount8(b) {
			t.Fatalf("failed for %v", b)
		}
	}
}

func TestMSBPos(t *testing.T) {
	for i := 0; i < 256; i++ {
		b := byte(i)
		expected := 8 - bits.LeadingZeros8(b)
		actual := int(leadingOnePos(b))
		if expected != actual {
			t.Fatalf("byte: %v, expected: %v, actual: %v", b, expected, actual)
		}
	}
}

func TestNthOnePos(t *testing.T) {
	// map from `b` to the result for each value of `n`, adding 1 to the array
	// index
	testCases := map[byte][8]byte{
		0b00000000: {},

		0b00000001: {1},
		0b10000000: {8},
		0b00001000: {4},

		0b10000001: {1, 8},
		0b00000011: {1, 2},
		0b00101000: {4, 6},

		0b10101010: {2, 4, 6, 8},
		0b01010101: {1, 3, 5, 7},

		0b11111111: {1, 2, 3, 4, 5, 6, 7, 8},
	}

	for b, res := range testCases {
		if got := nthOnePos(b, 0); got != 0 {
			t.Errorf("b=0b%08b, n=%d, got=%d, exp=0", b, 0, got)
		}
		if got := nthOnePos(b, 9); got != 0 {
			t.Errorf("b=0b%08b, n=%d, got=%d, exp=0", b, 9, got)
		}
		for n := byte(1); n <= 8; n++ {
			exp := res[n-1]
			got := nthOnePos(b, n)
			if exp != got {
				t.Errorf("b=0b%08b, n=%d, got=%d, exp=%d", b, n, got, exp)
			}
		}
	}
}

func TestRemoveAtIndex(t *testing.T) {
	testCases := []struct {
		input    []rune
		idx      int
		expected []rune
	}{
		{},
		{seq(1, 10), 0, seq(2, 10)},
		{seq(1, 10), 1, seq(3, 10, 1)},
		{seq(1, 10), 8, seq(1, 8, 10)},
		{seq(1, 10), 9, seq(1, 9)},
	}

	for i, tc := range testCases {
		removeAtIndex(&tc.input, tc.idx)
		if !slices.Equal(tc.expected, tc.input) {
			t.Fatalf("[%d]\nexpected: %v\n  actual: %v", i, tc.expected,
				tc.input)
		}
	}
}
