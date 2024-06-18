package runes

// alternatives tried that didn't make it

// IsInString is the slowest of all.
func IsInString(s string) func(rune) bool {
	return func(r rune) bool {
		for _, rr := range s {
			if r == rr {
				return true
			}
		}
		return false
	}
}

// IsInTable is far less memory efficient than IsInMask.
func IsInTable(s []rune) func(rune) bool {
	minRune, span := minAndSpan(s)
	table := make([]bool, span+1)
	for _, r := range s {
		r -= minRune
		table[r] = true
	}

	return func(r rune) bool {
		r -= minRune
		return r >= 0 && r <= span && table[r]
	}
}

func IsInSparseSet(s []rune) func(rune) bool {
	const breakPoint = 25 // determined with benchmarks
	if len(s) > breakPoint {
		return isInMap(runeSliceToMap(s))
	}
	return isInSlice(dedupRuneSlice(s))
}

func isInSlice(s []rune) func(rune) bool {
	return func(r rune) bool {
		for _, rr := range s {
			if r == rr {
				return true
			}
		}
		return false
	}
}

func isInMap(s map[rune]struct{}) func(rune) bool {
	return func(r rune) bool {
		_, ok := s[r]
		return ok
	}
}

// isInMaskSlice32 performs slightly slower than the 64 bit version. It is
// possible that this is because of the arch on which the benchmarks were run.
// They were run with:
//
//	go test -bench=. -count=10 -benchmem -timeout=20m
//
// Then, manually split the output between the 32 and 64 bit, and mangled the
// names so that benchstat would show diffs.
func isInMaskSlice32(s []rune, minRune, span rune) func(rune) bool {
	t := make([]uint32, 1+span/32)
	for _, r := range s {
		r -= minRune
		t[r/32] |= 1 << (r % 32)
	}

	return func(r rune) bool {
		r -= minRune
		m := r % 32
		i := int(r / 32)
		// checking `r` or `i` is the same, but only `i` does BCE
		return i >= 0 && i < len(t) && // eliminate runtime.panicIndex
			m >= 0 && m < 32 && // eliminate runtime.panicshift
			1<<(m)&t[i] != 0
	}
}
