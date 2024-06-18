package runes

import "unicode/utf8"

func IsStringFunc(s string) func(rune) bool {
	return isInStrategy(runeMapToSlice(stringToRuneMap(s)))
}

func IsBytesFunc(s []byte) func(rune) bool {
	return isInStrategy(runeMapToSlice(bytesToRuneMap(s)))
}

func IsRunesFunc(s []rune) func(rune) bool {
	return isInStrategy(dedupRuneSlice(s))
}

func isInStrategy(s []rune) func(rune) bool {
	if len(s) == 0 {
		return isNeverIn
	}

	minRune, span := minAndSpan(s)
	if span < 64 {
		return isInMask64(s, minRune)
	}

	return isInMaskSlice64(s, minRune, span)
}

func isNeverIn(rune) bool {
	return false
}

func isInMaskSlice64(s []rune, minRune, span rune) func(rune) bool {
	t := make([]uint64, 1+span/64)
	for _, r := range s {
		r -= minRune
		t[r/64] |= 1 << (r % 64)
	}

	return func(r rune) bool {
		r -= minRune
		m := r % 64
		i := int(r / 64)
		// checking `r` or `i` is the same, but only `i` does BCE
		return i >= 0 && i < len(t) && // eliminate runtime.panicIndex
			m >= 0 && m < 64 && // eliminate runtime.panicshift
			1<<(m)&t[i] != 0
	}
}

func isInMask64(s []rune, minRune rune) func(rune) bool {
	var mask uint64
	for _, r := range s {
		mask |= 1 << (r - minRune)
	}

	return func(r rune) bool {
		r -= minRune
		return r >= 0 && r < 64 && 1<<r&mask != 0
	}
}

func minAndSpan(s []rune) (minRune, span rune) {
	for i, r := range s {
		if i == 0 || r < minRune {
			minRune = r
		}
		if r > span {
			span = r
		}
	}
	return minRune, span - minRune
}

func stringToRuneMap(s string) map[rune]struct{} {
	m := make(map[rune]struct{}, len(s))
	for _, r := range s {
		if utf8.ValidRune(r) {
			m[r] = struct{}{}
		}
	}
	return m
}

func bytesToRuneMap(s []byte) map[rune]struct{} {
	m := make(map[rune]struct{}, len(s))
	for len(s) > 0 {
		r, size := utf8.DecodeRune(s)
		if utf8.ValidRune(r) {
			m[r] = struct{}{}
		}
		s = s[size:]
	}
	return m
}

func dedupRuneSlice(s []rune) []rune {
	if m := runeSliceToMap(s); len(m) != len(s) {
		return runeMapToSlice(m)
	}
	return s
}

func runeSliceToMap(s []rune) map[rune]struct{} {
	m := make(map[rune]struct{}, len(s))
	for _, r := range s {
		if utf8.ValidRune(r) {
			m[r] = struct{}{}
		}
	}
	return m
}

func runeMapToSlice(m map[rune]struct{}) []rune {
	s := make([]rune, 0, len(m))
	for r := range m {
		if utf8.ValidRune(r) {
			s = append(s, r)
		}
	}
	return s
}
