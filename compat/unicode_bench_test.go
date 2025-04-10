package compat

import (
	"fmt"
	"testing"
	"unicode"
	"unicode/utf8"

	"github.com/diegommm/runes"
	"github.com/diegommm/runes/util"
)

func BenchmarkIsSpace(b *testing.B) {
	rt := FromUnicode(unicode.White_Space)
	rs := util.Collect(rt.Iter())
	ls := runes.LinearSlice[rune](rs)
	bs := runes.BinarySlice[rune](rs)
	bm := runes.NewBitmap(rt)

	testRunes := append(rs, -1, 0, utf8.MaxRune)

	for _, r := range testRunes {
		b.Run(fmt.Sprintf("0x%x", r), func(b *testing.B) {
			b.Run("implem=stdlib", func(b *testing.B) {
				b.Skip("no comparison yet")
				for i := 0; i < b.N; i++ {
					unicode.IsSpace(r)
				}
			})
			b.Run("implem=LinearSearch", func(b *testing.B) {
				b.Skip("slower than stdlib")
				for i := 0; i < b.N; i++ {
					ls.Contains(r)
				}
			})
			b.Run("implem=BinarySearch", func(b *testing.B) {
				b.Skip("slower than stdlib")
				for i := 0; i < b.N; i++ {
					bs.Contains(r)
				}
			})
			b.Run("implem=Bitmap", func(b *testing.B) {
				b.Skip("too much memory")
				for i := 0; i < b.N; i++ {
					bm.Contains(r)
				}
			})
		})
	}
}
