//go:build !sizeof

package util

import "testing"

func TestSizeof(t *testing.T) {
	s, ok := Sizeof(nil)
	MustEqual(t, 0, s, "expected zero size")
	MustEqual(t, false, ok, "expected false")
}

func TestSizeofUnicodeRangeTable(t *testing.T) {
	s, ok := SizeofUnicodeRangeTable(nil)
	MustEqual(t, 0, s, "expected zero size")
	MustEqual(t, false, ok, "expected false")
}
