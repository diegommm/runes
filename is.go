package runes

import "unicode/utf8"

func IsStringFunc(s string) func(rune) bool {
	return isStrategy(runeMapToSlice(stringToRuneMap(s)))
}

func IsBytesFunc(s []byte) func(rune) bool {
	return isStrategy(runeMapToSlice(bytesToRuneMap(s)))
}

func IsRunesFunc(s []rune) func(rune) bool {
	return isStrategy(dedupRuneSlice(s))
}

// isStrategy combines all the strategies to have the best of each.
func isStrategy(s []rune) func(rune) bool {
	if len(s) == 0 {
		return isNeverIn
	}

	minRune, span := minAndSpan(s)
	if span < 64 {
		return isMask64(s, minRune)
	}

	// this could use some help here. If we can prove that we have less than the
	// practical thresold of ~25 runes to check, and they are /very/ sparse,
	// then it may be a good job for `isSlice`. `isMap` works well at runtime,
	// but is able to make several allocations for a much higher total than just
	// the max 136KiB of `isMaskSlice64` for comparatively much fewer elements.
	// I think it was around >12k runes that would cause `isMap` to allocate
	// more than 136KiB, so it is may still be a good choice memory-wise for
	// very sparse data with more than ~25 runes, and less than those ~12k. It
	// does, however, perform much worse than `isMaskSlice64`, at least ~7 times
	// slower at runtime. And it does grow worse as more items are added, so it
	// can't beat the low and constant time of `isMaskSlice64`

	return isMaskSlice64(s, minRune, span)
}

// isNeverIn is here to do its job. Why create a closure when I can use the
// same?
func isNeverIn(rune) bool {
	return false
}

// isMaskSlice64 is the same principle of isMask64, but is able to handle any
// set of runes to check by allocating a []uint64. The first bit of the first
// item in the slice is the smallest rune to check. Also runs in constant time,
// but it allocates a single big contiguous bitmask, so if you feed it only two
// runes to check, one being the smallest valid rune and the other being
// utf8.MaxRune, then it will allocate 136KiB only to check them both.
func isMaskSlice64(s []rune, minRune, span rune) func(rune) bool {
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

// isMask64 works great when the runes to check for don't have more than 63
// runes among any combination of them, whatever is their value. Constant time.
func isMask64(s []rune, minRune rune) func(rune) bool {
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
