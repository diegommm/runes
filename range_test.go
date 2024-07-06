package runes

import (
	"slices"
	"testing"
	"unsafe"
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

func testRangeInvariants(t *testing.T, r Range, expectedRunes []rune) {
	t.Helper()

	Len, Min, Max := r.RuneLen(), r.Min(), r.Max()

	if Len < 0 || Min < 0 || Max < 0 ||
		Len == 0 && (Min != 0 || Max != 0) {
		t.Fatalf("unexpected invariants: len=%v min=%v max=%v", Len, Min, Max)
	}

	if v := r.Nth(0); (v >= 0) != (Len > 0) || Min != v {
		t.Fatalf("unexpected invariants: len=%v min=%v first=%v", Len, Min, v)
	}

	if v := r.Nth(Len - 1); (v >= 0) != (Len > 0) || Max != v {
		t.Fatalf("unexpected invariants: len=%v max=%v last=%v", Len, Max, v)
	}

	runes := make([]rune, Len)
	m := make(map[rune]struct{}, Len)
	for i := range runes {
		runes[i] = r.Nth(int32(i))
		if runes[i] < 0 {
			t.Fatalf("index %d not found with len=%v", i, Len)
		}
		if i > 0 && runes[i] <= runes[i-1] {
			t.Fatalf("rune %v at index %v should be greater than the "+
				"previous rune %v", runes[i], i, runes[i-1])
		}
		m[runes[i]] = struct{}{}
	}

	if !slices.Equal(expectedRunes, runes) {
		t.Fatalf("mismatched runes:\n\texpected: %v\n\t  actual: %v",
			expectedRunes, runes)
	}

	for i, rr := range runes {
		if !r.Contains(rr) {
			t.Fatalf("should contain rune %v with index %d", rr, i)
		}
	}

	runRuneTest(t, 0, func(t *testing.T, rr rune) {
		if _, ok := m[rr]; !ok && r.Contains(rr) {
			t.Fatalf("should not contain rune %v", rr)
		}
	})
}

func TestUniformRange(t *testing.T) {
	t.Parallel()
	const minRune = 3

	t.Run("1 value", func(t *testing.T) {
		t.Parallel()

		r, err := NewUniformRange5(minRune, 1, 1)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		testRangeInvariants(t, r, []rune{minRune})
	})

	t.Run("3 contiguous values", func(t *testing.T) {
		t.Parallel()

		r, err := NewUniformRange5(minRune, 3, 1)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		testRangeInvariants(t, r, []rune{minRune, minRune + 1, minRune + 2})
	})

	t.Run("3 values spaced by 4", func(t *testing.T) {
		t.Parallel()

		r, err := NewUniformRange5(minRune, 3, 4)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		testRangeInvariants(t, r, []rune{minRune, minRune + 5, minRune + 10})
	})
}
