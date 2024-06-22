package runes

import (
	"fmt"
	"testing"
)

const maxUint32 = 1<<32 - 1

func runRuneTest(t *testing.T, parallelism uint8, f func(*testing.T, rune)) {
	t.Helper()

	if testing.Short() {
		runShortRuneTest(t, f)
		return
	}

	partLen := 1 + maxUint32/int64(parallelism)
	for i := byte(0); i < parallelism; i++ {
		first := int64(i) * partLen
		last := min(first+partLen-1, maxUint32)

		t.Run(fmt.Sprintf("[%v,%v]", rune(first), rune(last)), func(t *testing.T) {
			t.Parallel()
			for j := first; j <= last; j++ {
				f(t, rune(j))
			}
		})
	}
}

func runShortRuneTest(t *testing.T, f func(*testing.T, rune)) {
	t.Helper()

	for _, c := range runeCases {
		t.Run(fmt.Sprintf("rune %v", c.r), func(t *testing.T) {
			f(t, c.r)
		})
	}
}

var runeCases = []struct {
	r    rune
	name string
}{
	{-1, "invalid, negative"},
	{10, "ASCII"},
	{209, "2 bytes"},
	{surrogateMin, "invalid, surrogate"},
	{65535, "3 bytes"},
	{maxRune, "4 bytes"},
	{maxRune + 11, "invalid, too long"},
}
