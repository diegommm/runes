package runes

import (
	"testing"
	"unicode/utf8"
)

func TestEncodeRune(t *testing.T) {
	t.Parallel()

	runRuneTest(t, 8, func(t *testing.T, r rune) {
		var expb, gotb, got4b [4]byte
		expn := utf8.EncodeRune(expb[:], r)
		gotn := EncodeRune(gotb[:], r)
		got4n := EncodeRune4(got4b[:], r)

		if expn != gotn {
			t.Fatalf("[EncodeRune] unexpected length %v for rune %x, expected %v", gotn, r, expn)
		}

		if expb != gotb {
			t.Fatalf("[EncodeRune] unexpected bytes %v for rune %x, expected %v", gotb, r, expb)
		}

		if expn != got4n {
			t.Fatalf("[EncodeRune4] unexpected length %v for rune %x, expected %v", got4n, r, expn)
		}

		if expb != got4b {
			t.Fatalf("[EncodeRune4] unexpected bytes %v for rune %x, expected %v", got4b, r, expb)
		}
	})
}

func TestAppendRune(t *testing.T) {
	t.Parallel()

	runRuneTest(t, 64, func(t *testing.T, r rune) {
		expb := utf8.AppendRune(nil, r)
		gotb := AppendRune(nil, r)

		if len(expb) != len(gotb) {
			t.Fatalf("unexpected bytes %v for rune %x, expected %v", gotb, r, expb)
		}
		for i := range expb {
			if expb[i] != gotb[i] {
				t.Fatalf("unexpected bytes %v for rune %x, expected %v", gotb, r, expb)
			}
		}
	})
}

func TestUTF8Bytes(t *testing.T) {
	t.Parallel()

	runRuneTest(t, 8, func(t *testing.T, r rune) {
		var expb [4]byte
		expn := utf8.EncodeRune(expb[:], r)
		gotb, gotn := UTF8BytesValue(r)

		if expn != gotn {
			t.Fatalf("unexpected length %v for rune %x, expected %v", gotn, r, expn)
		}

		if expb != gotb {
			t.Fatalf("unexpected bytes %v for rune %x, expected %v", gotb, r, expb)
		}
	})
}
