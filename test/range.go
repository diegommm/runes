package test

import (
	"fmt"
	"testing"

	"github.com/diegommm/runes/iface"
)

type (
	Set      = iface.Set
	Range    = iface.Range
	Iterator = iface.Iterator
)

var RuneCases = []struct {
	Rune rune
	Name string
}{
	{-1, "invalid-negative"},
	{97, "valid-1-byte, 1 byte"},
	{209, "valid-2-bytes"},
	{26412, "valid-3-bytes"},
	{55296, "invalid-surrogate"},
	{128170, "valid-4-bytes"},
	{33554432, "invalid-too-long"},
}

type RangeInvariantTestCase struct {
	Name  string
	Range Range
	Runes []rune
}

type RangeInvariantTestCases []*RangeInvariantTestCase

func (tcs RangeInvariantTestCases) Run(t *testing.T) {
	for _, tc := range tcs {
		TestRangeInvariants(t, tc)
	}
}

func TestRangeInvariants(t *testing.T, tc *RangeInvariantTestCase) {
	t.Helper()

	t.Run("case="+tc.Name, func(t *testing.T) {
		t.Parallel()
		t.Cleanup(func() {
			if t.Failed() {
				t.Logf("Range: %#v", tc.Range)
			}
		})

		Len, Min, Max := tc.Range.RuneLen(), tc.Range.Min(), tc.Range.Max()

		// basic Len/Min/Max behaviour
		if len(tc.Runes) != int(Len) {
			t.Fatalf("expected len %d, actual %d", len(tc.Runes), Len)
		}
		if Len > 0 {
			if xMin, xMax := tc.Runes[0], tc.Runes[Len-1]; xMin != Min ||
				xMax != Max {
				t.Fatalf("expected Min/Max %d/%d, actual %d/%d", xMin, xMax,
					Min, Max)
			}
		}
		if Len < 0 || Len == 0 && (Min != -1 || Max != -1) ||
			Len == 1 && (Min != Max) {
			t.Fatalf("Len=%v Min=%v Max=%v", Len, Min, Max)
		}

		// Nth behaviour
		if v := tc.Range.Nth(-1); v != -1 {
			t.Fatalf("Nth(-1): expected -1, got %v", v)
		}
		if v := tc.Range.Nth(Len); v != -1 {
			t.Fatalf("Nth(Len): expected -1, got %v", v)
		}
		if v := tc.Range.Nth(0); (v >= 0) != (Len > 0) || Min != v {
			t.Fatalf("Nth(0)=%v Len=%v Min=%v", v, Len, Min)
		}
		if v := tc.Range.Nth(Len - 1); (v >= 0) != (Len > 0) || Max != v {
			t.Fatalf("Nth(Len-1)=%v Len=%v Max=%v", v, Len, Max)
		}

		// Nth behaviour
		if v := tc.Range.Pos(Min - 1); v != -1 {
			t.Fatalf("Pos(Min-1): expected -1, got %v", v)
		}
		if v := tc.Range.Pos(Max + 1); v != -1 {
			t.Fatalf("Pos(Max+1): expected -1, got %v", v)
		}
		if v := tc.Range.Pos(Min); (Len > 0) != (v == 0) {
			t.Fatalf("Pos(Min)=%v Len=%v Min=%v", v, Len, Min)
		}
		if v := tc.Range.Pos(Max); (Len > 0) != (v >= 0 && v == Len-1) {
			t.Fatalf("Pos(Max)=%v Len=%v Max=%v", v, Len, Max)
		}

		runes := make([]rune, Len)
		m := make(map[rune]struct{}, Len)
		for i := range runes {
			runes[i] = tc.Range.Nth(int32(i))
			if runes[i] != tc.Runes[i] {
				t.Fatalf("at index: %d, expected: %v, actual: %v", i,
					tc.Runes[i], runes[i])
			}
			if pos := tc.Range.Pos(runes[i]); int(pos) != i {
				t.Fatalf("expected position: %d, actual: %d", i, pos)
			}
			if i > 0 && runes[i] <= runes[i-1] {
				t.Fatalf("rune %v at index %v should be greater than the "+
					"previous rune %v", runes[i], i, runes[i-1])
			}
			m[runes[i]] = struct{}{}
		}

		for i, rr := range runes {
			if !tc.Range.Contains(rr) {
				t.Fatalf("should contain rune %v with index %d", rr, i)
			}
		}

		for _, c := range RuneCases {
			t.Run(fmt.Sprintf("rune=%v", c.Rune), func(t *testing.T) {
				if _, ok := m[c.Rune]; !ok && tc.Range.Contains(c.Rune) {
					t.Fatalf("should not contain rune %v", c.Rune)
				}
			})
		}
	})
}
