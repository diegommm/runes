package runes

import (
	"testing"
	"unicode/utf8"
)

func TestValidRune(t *testing.T) {
	t.Parallel()

	runRuneTest(t, 8, func(t *testing.T, r rune) {
		if utf8.ValidRune(r) != ValidRune(r) {
			t.Fatalf("failed for rune %x", r)
		}
	})
}
