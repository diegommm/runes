package test

import (
	"fmt"
	"slices"
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
	{0, "zero"},
	{97, "valid-1-byte, 1 byte"},
	{209, "valid-2-bytes"},
	{26412, "valid-3-bytes"},
	{55296, "invalid-surrogate"},
	{128170, "valid-4-bytes"},
	{33554432, "invalid-too-long"},
}

type RangeInvariantTestCase struct {
	Name              string
	Range             Range
	AllValidRunes     []rune
	ExtraInvalidRunes []rune
}

type RangeInvariantTestCases []*RangeInvariantTestCase

func (tcs RangeInvariantTestCases) Run(t *testing.T) {
	for _, tc := range tcs {
		TestRangeInvariants(t, tc)
	}
}

func TestRangeInvariants(t *testing.T, tc *RangeInvariantTestCase) {
	t.Helper()

	if !slices.IsSorted(tc.AllValidRunes) {
		t.Fatalf("runes in test cases expected to be sorted in ascending order"+
			", got: %v", tc.AllValidRunes)
	}

	t.Run("case="+tc.Name, func(t *testing.T) {
		t.Parallel()
		t.Cleanup(func() {
			if t.Failed() {
				t.Logf("Range: %#v", tc.Range)
			}
		})

		Len, Min, Max := tc.Range.RuneLen(), tc.Range.Min(), tc.Range.Max()

		t.Run("basic Len/Min/Max behaviour", func(t *testing.T) {
			if l := len(tc.AllValidRunes); l != int(Len) {
				t.Fatalf("expected len %d, actual %d", l, Len)
			}
			if Len > 0 {
				xMin, xMax := tc.AllValidRunes[0], tc.AllValidRunes[Len-1]
				if xMin != Min || xMax != Max {
					t.Fatalf("expected Min/Max %d/%d, actual %d/%d", xMin, xMax,
						Min, Max)
				}
			}
			if Len < 0 || Len == 0 && (Min != -1 || Max != -1) ||
				Len == 1 && (Min != Max) {
				t.Fatalf("Len=%v Min=%v Max=%v", Len, Min, Max)
			}
		})

		t.Run("boundaries of Contains", func(t *testing.T) {
			if tc.Range.Contains(Min - 1) {
				t.Fatalf("Contains(Min-1)=true Min=%v", Min)
			}
			if tc.Range.Contains(Max + 1) {
				t.Fatalf("Contains(Max+1)=true Max=%v", Max)
			}
		})

		t.Run("common cases of Nth", func(t *testing.T) {
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
		})

		t.Run("common cases of Pos", func(t *testing.T) {
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
		})

		t.Run("all valid runes in the range by index", func(t *testing.T) {
			for i := range tc.AllValidRunes {
				expected := tc.AllValidRunes[i]
				actual := tc.Range.Nth(int32(i))
				if actual != expected {
					t.Fatalf("Nth(%d)=%v, expected: %v", i, actual, expected)
				}
				if !tc.Range.Contains(actual) {
					t.Fatalf("Contains(%v)=false", actual)
				}
				if pos := tc.Range.Pos(actual); int(pos) != i {
					t.Fatalf("Pos(%v)=%v, expected %d", actual, pos, i)
				}
			}
		})

		t.Run("interstitial invalid runes", func(t *testing.T) {
			for i := 1; i < len(tc.AllValidRunes); i++ {
				for r := tc.AllValidRunes[i-1] + 1; r < tc.AllValidRunes[i]; r++ {
					if tc.Range.Contains(r) {
						t.Fatalf("Contains(%v)=true", r)
					}
					if pos := tc.Range.Pos(r); pos != -1 {
						t.Fatalf("Pos(%v)=%v (should be -1)", r, pos)
					}
				}
			}
		})

		t.Run("extra invalid runes", func(t *testing.T) {
			for _, r := range tc.ExtraInvalidRunes {
				if tc.Range.Contains(r) {
					t.Fatalf("Contains(%v)=true", r)
				}
				if pos := tc.Range.Pos(r); pos != -1 {
					t.Fatalf("Pos(%v)=%v (should be -1)", r, pos)
				}
			}
		})

		t.Run("predefined rune cases", func(t *testing.T) {
			m := make(map[rune]struct{}, Len)
			for _, r := range tc.AllValidRunes {
				m[r] = struct{}{}
			}
			for _, c := range RuneCases {
				t.Run(fmt.Sprintf("rune=%v", c.Rune), func(t *testing.T) {
					_, ok := m[c.Rune]
					if ok != tc.Range.Contains(c.Rune) {
						t.Fatalf("Contains(%v)=%v", c.Rune, ok)
					}
					if pos := tc.Range.Pos(c.Rune); ok != (pos != -1) {
						t.Fatalf("Pos(%v)=%v", c.Rune, pos)
					}
				})
			}
		})
	})
}
