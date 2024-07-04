package runes

import (
	"io"
	"slices"
	"unicode/utf8"
)

// alternatives tried that didn't make it

// isSparseSet combines the best of isSlice and isMap in my local machine. This
// is just a toy.
func isSparseSet(s []rune) func(rune) bool {
	const breakPoint = 25 // determined with benchmarks
	if len(s) > breakPoint {
		return isMap(runeSliceToMap(s))
	}
	return isSlice(s)
}

// isSlice nice and easy, performs well for a handful of runes, but not better
// than the mask ones.
func isSlice(s []rune) func(rune) bool {
	s = append(s[:0:0], s...) // clone
	slices.Sort(s)
	return func(r rune) bool {
		_, found := slices.BinarySearch(s, r)
		return found
	}
}

// isMap is obviously better than isSlice as the number of runes increase past a
// certain amount, but its memory cost is very high.
func isMap(s map[rune]struct{}) func(rune) bool {
	return func(r rune) bool {
		_, ok := s[r]
		return ok
	}
}

// isMask32 may fare better in 32 bit systems than its 64 bit sibling.
func isMask32(rr io.RuneReader, minRune rune) func(rune) bool {
	var mask uint32
	for {
		r, _, err := rr.ReadRune()
		if err != nil {
			break
		}
		mask |= 1 << (r - minRune)
	}

	return func(r rune) bool {
		r -= minRune
		return r >= 0 && r < 32 && 1<<r&mask != 0
	}
}

// isMaskSlice32 performs slightly slower than the 64 bit version. It is
// possible that this is because of the arch on which the benchmarks were run.
// They were run with:
//
//	go test -count=10 -benchmem -timeout=20m -bench=BenchmarkIsInSetFunc -run=-
//
// Then, manually split the output between the 32 and 64 bit, and mangled the
// names so that benchstat would show diffs. See the file:
//
//	is-in-mask-slice-32-vs-64.txt
//
// If the 32 bit version works better in 32 bit systems, a strategy of checking
// the system architecture could be made, but I don't have a (real) 32 bit
// system at hand to test it now. Also, may not be the same for all
// architectures, but the operations are nothing fancy, so it probably should
// hold.
func isMaskSlice32(rr io.RuneReader, minRune, span rune) func(rune) bool {
	t := make([]uint32, ceilDiv(span+1, 32))
	for {
		r, _, err := rr.ReadRune()
		if err != nil {
			break
		}
		r -= minRune
		t[r/32] |= 1 << (r % 32)
	}

	return func(r rune) bool {
		u := uint32(r - minRune)
		i := int(u / 32)
		return i < len(t) && 1<<(u%32)&t[i] != 0
	}
}

// helper funcs

func runeSliceToMap(s []rune) map[rune]struct{} {
	m := make(map[rune]struct{}, len(s))
	for _, r := range s {
		if utf8.ValidRune(r) {
			m[r] = struct{}{}
		}
	}
	return m
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
