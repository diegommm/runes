//go:build !sizeof

package util

import "unicode"

// Sizeof returns zero and false. This is the counterpart of the function in a
// file with the opposite build tag.
func Sizeof(x interface{ Contains(rune) bool }) (uintptr, bool) {
	return 0, false
}

// SizeofUnicodeRangeTable returns zero and false.  This is the counterpart of
// the function in a file with the opposite build tag.
func SizeofUnicodeRangeTable(rt *unicode.RangeTable) (uintptr, bool) {
	return 0, false
}
