package runes

import (
	"fmt"
	"slices"
	"testing"
	"unicode"
	"unicode/utf8"

	"github.com/diegommm/runes/util"
)

func BenchmarkIsSpace(b *testing.B) {
	b.Skip("not ready :)")

	rt := unicode.White_Space
	rts := slices.Collect(util.RangeTableIter(rt))
	testRunes := append(rts, -1, 0, utf8.MaxRune)
	testRunes = testRunes[:1] // TODO

	// alternative implementation
	custom := isSpaceCustom()

	// size estimations
	rtSize := util.FormatSizeEstimation(util.SizeofUnicodeRangeTable(rt))
	b.Log("estimated size in bytes of standard library implem: ", rtSize)
	bmuSize := util.FormatSizeEstimation(util.Sizeof(custom))
	b.Log("estimated size in bytes of Custom implem: ", bmuSize)

	for _, r := range testRunes {
		b.Run(fmt.Sprintf("0x%x", r), func(b *testing.B) {
			b.Run("implem=Custom", func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					custom.Contains(r)
				}
			})
			b.Run("implem=stdlib", func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					unicode.IsSpace(r)
				}
			})
		})
	}
}

func isSpaceCustom() Bitmap {
	// TODO
	return NewBitmap([]rune{
		9,
		10,
		11,
		12,
		13,
		32,
		133,
		160,
		5760,
		8192,
		8193,
		8194,
		8195,
		8196,
		8197,
		8198,
		8199,
		8200,
		8201,
		8202,
		8232,
		8233,
		8239,
		8287,
		12288,
	})
}
