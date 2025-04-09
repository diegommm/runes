package compat

import (
	"fmt"
	"testing"
	"unicode"
)

func TestRageTable(t *testing.T) {
	testCases := []struct {
		rt            *unicode.RangeTable
		expectedCount int
	}{
		{unicode.White_Space, 25},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("index=%v", i), func(t *testing.T) {
			rt := FromUnicode(tc.rt)
			var count int
			for r := range rt.All() {
				count++
				if !unicode.Is(tc.rt, r) {
					t.Errorf("Expected 0x%x to be space", r)
				}
			}
			if count != tc.expectedCount {
				t.Errorf("Expected count %v, got %v", tc.expectedCount, count)
			}
		})
	}
}
