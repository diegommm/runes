package runes

import (
	"bytes"
	"io"
	"slices"
	"strings"
	"unicode/utf8"
)

// IsInStringSet returns a function that checks if a rune if found in the
// provided string.
func IsInStringSet(s string) func(rune) bool {
	return isStrategy(RuneReadSeekerToRestarter(strings.NewReader(s)))
}

// IsInBytesSet returns a function that checks if a rune if found in the slice
// of bytes, which is interpreted as a UTF-8 string.
func IsInBytesSet(s []byte) func(rune) bool {
	return isStrategy(RuneReadSeekerToRestarter(bytes.NewReader(s)))
}

// IsInRunesSet returns a function that checks if a rune if found in the
// provided slice of runes.
func IsInRunesSet(s []rune) func(rune) bool {
	return isStrategy(NewRuneSliceRuneIterator(s))
}

// isStrategy combines all the strategies to have the best of each.
func isStrategy(r3 RuneIterator) func(rune) bool {
	minRune, span, count := startSpanCount(r3)
	if count == 0 {
		return isNeverIn
	}

	r3.Restart()
	if span < 64 {
		return isMask64(r3, minRune)
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

	return isMaskString(r3, minRune, span)
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
func isMaskSlice64(rr io.RuneReader, minRune, span rune) func(rune) bool {
	t := make([]uint64, ceilDiv(span+1, 64))
	for {
		r, _, err := rr.ReadRune()
		if err != nil {
			break
		}
		u := uint32(r - minRune)
		t[u/64] |= 1 << (u % 64)
	}

	return func(r rune) bool {
		u := uint32(r - minRune)
		i := int(u / 64)
		return i < len(t) && 1<<(u%64)&t[i] != 0
	}
}

func isMaskString(rr io.RuneReader, minRune, span rune) func(rune) bool {
	bin := make([]byte, ceilDiv(span+1, 8))
	for {
		r, _, err := rr.ReadRune()
		if err != nil {
			break
		}
		u := uint32(r - minRune)
		bin[u/8] |= 1 << (u % 8)
	}

	var minRuneBytes [3]byte
	encodeFixedRune(&minRuneBytes, minRune)

	var b strings.Builder
	b.Grow(len(bin) + 3)
	b.Write(minRuneBytes[:])
	b.Write(bin)
	bin = nil
	b.Reset()
	t := b.String()

	return func(r rune) bool {
		if len(t) < 4 { // BCE
			return false
		}
		u := uint32(r - decodeFixedRune(t[0], t[1], t[2]))
		i := 3 + int(u/8)
		// why having runtime.panicIndex is faster here???
		return i < len(t) && 1<<(u%8)&t[i] != 0
	}
}

// isMask64 works great when the runes to check for don't have more than 63
// runes among any combination of them, whatever is their value. Constant time.
func isMask64(rr io.RuneReader, minRune rune) func(rune) bool {
	var mask uint64
	for {
		r, _, err := rr.ReadRune()
		if err != nil {
			break
		}
		u := uint32(r - minRune)
		mask |= 1 << u
	}

	return func(r rune) bool {
		u := uint32(r - minRune)
		return u < 64 && 1<<u&mask != 0
	}
}

func startSpanCount(rr io.RuneReader) (minRune, span rune, count int) {
	for ; ; count++ {
		r, _, err := rr.ReadRune()
		if err != nil {
			break
		}
		if count == 0 || r < minRune {
			minRune = r
		}
		if r > span {
			span = r
		}
	}
	return minRune, span - minRune, count
}

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
