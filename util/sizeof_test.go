//go:build sizeof

package util

import (
	"testing"
	"unicode"
)

func TestSizeof(t *testing.T) {
	s, ok := Sizeof(ContainsFunc(nil))
	MustEqual(t, 0, s, "expected zero size")
	MustEqual(t, true, ok, "expected true")
}

func TestSizeofSlice(t *testing.T) {
	s := SizeofSlice([]rune{1, 2, 3})
	MustEqual(t, true, s != 0, "expected non-zero size")

	type set interface {
		Contains(rune) bool
	}
	s = SizeofSlice[[]set, set]([]set{ContainsFunc(nil)})
	MustEqual(t, true, s != 0, "expected non-zero size")
}

func TestSizeofUnicodeRangeTable(t *testing.T) {
	s, ok := SizeofUnicodeRangeTable(unicode.White_Space)
	MustEqual(t, true, s != 0, "expected non-zero size")
	MustEqual(t, true, ok, "expected true")
}
