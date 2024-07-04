package runes

import (
	"fmt"
	"runtime"
	"strings"
	"testing"
	"unicode/utf8"
	"unsafe"
)

const (
	surrogateMin = 0xD800 // 55296
	surrogateMax = 0xDFFF // 57343
)

type WithByteLen interface {
	ByteLen() (l int, exact bool)
}

func EstimateByteLen[T any](t T) (_ int, exact bool) {
	if v, _ := any(t).(WithByteLen); v != nil {
		return v.ByteLen()
	}
	return int(unsafe.Sizeof(t)), false
}

var runeCases = []struct {
	r    rune
	name string
	n    int // number of bytes of what it should encode to
}{
	{-1, "invalid-negative", 3},
	{97, "valid-1-byte, 1 byte", 1},
	{209, "valid-2-bytes", 2},
	{26412, "valid-3-bytes", 3},
	{55296, "invalid-surrogate", 3},
	{128169, "valid-4-bytes", 4},
	{33554432, "invalid-too-long", 3},
}

func runRuneTest(t *testing.T, parallelism uint, f func(*testing.T, rune)) {
	t.Helper()

	if parallelism == 0 {
		parallelism = uint(runtime.GOMAXPROCS(0))
	}

	if testing.Short() {
		for _, c := range runeCases {
			t.Run(fmt.Sprintf("rune=%v", c.r), func(t *testing.T) {
				f(t, c.r)
			})
		}
		return
	}

	partLen := 1 + (1<<32-1)/int64(parallelism)
	for i := uint(0); i < parallelism; i++ {
		first := int64(i) * partLen
		last := min(first+partLen-1, (1<<32 - 1))

		t.Run(fmt.Sprintf("[%v,%v]", rune(first), rune(last)), func(t *testing.T) {
			t.Parallel()
			for j := first; j <= last; j++ {
				f(t, rune(j))
			}
		})
	}
}

type runeIterator struct {
	Rune    rune
	last    rune
	started bool
}

func newRuneIterator(first, last rune) *runeIterator {
	return &runeIterator{
		Rune: first,
		last: last,
	}
}

func (i *runeIterator) Next() bool {
	if i.Rune >= i.last {
		return false
	}
	if !i.started {
		i.started = true
		return true
	}
	i.Rune++
	return true
}

type validRuneIterator struct {
	*runeIterator
}

func newValidRuneIterator() validRuneIterator {
	return validRuneIterator{
		runeIterator: newRuneIterator(0, utf8.MaxRune),
	}
}

func (i validRuneIterator) Next() bool {
	if i.Rune == surrogateMin-1 {
		i.Rune = surrogateMax + 1
		return true
	}
	return i.runeIterator.Next()
}

type invalidRuneIterator struct {
	*runeIterator
}

func newInvalidRuneIterator() invalidRuneIterator {
	return invalidRuneIterator{
		runeIterator: newRuneIterator(-(1 << 31), 1<<31-1),
	}
}

func (i invalidRuneIterator) Next() bool {
	switch i.Rune {
	case -1:
		i.Rune = surrogateMin
		return true
	case surrogateMax:
		i.Rune = utf8.MaxRune + 1
		return true
	default:
		return i.runeIterator.Next()
	}
}

func validRuneSlice(start, count rune) []rune {
	if count < 1 {
		panic("count must be positive")
	}
	end := start + count
	if !utf8.ValidRune(start) || !utf8.ValidRune(end) {
		panic("invalid position")
	}
	l := count
	if start < surrogateMin && end >= surrogateMin {
		l -= surrogateMax - surrogateMin
	}
	ret := make([]rune, 0, l)
	i := newValidRuneIterator()
	i.Rune = start
	for i.Next() && i.Rune < end {
		ret = append(ret, i.Rune)
	}
	return ret
}

func makeContiguousRunesString(start, count rune) string {
	var b strings.Builder
	b.Grow(int(count)) // could be bigger, nvm
	for i := rune(0); i < count; i++ {
		b.WriteRune(start + i)
	}
	return b.String()
}
