package runes

import (
	"testing"
	"unicode/utf8"
)

func BenchmarkEncodeRuneText(b *testing.B) {
	b.Skip("deprecated")
	const runesToWrite = 1024 * 1024

	for _, c := range runeCases {
		buf := make([]byte, runesToWrite*c.n+3)

		b.Run(c.name, func(b *testing.B) {
			b.Run("implem=stdlib", func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					dst := buf
					for j := 0; j < runesToWrite; j++ {
						n := utf8.EncodeRune(dst, c.r)
						dst = dst[n:]
					}
				}
			})

			b.Run("implem=local", func(b *testing.B) {
				b.SkipNow()
				for i := 0; i < b.N; i++ {
					dst := buf
					for j := 0; j < runesToWrite; j++ {
						n := EncodeRune(dst, c.r)
						dst = dst[n:]
					}
				}
			})

			b.Run("implem=EncodeRune4", func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					dst := buf
					for j := 0; j < runesToWrite; j++ {
						n := EncodeRune4(dst, c.r)
						dst = dst[n:]
					}
				}
			})
		})
	}
}

func BenchmarkEncodeRune(b *testing.B) {
	b.Skip("deprecated")
	for _, c := range runeCases {
		b.Run(c.name, func(b *testing.B) {
			p := make([]byte, 4)

			b.Run("implem=stdlib", func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					utf8.EncodeRune(p, c.r)
				}
			})

			b.Run("implem=local", func(b *testing.B) {
				b.SkipNow()
				for i := 0; i < b.N; i++ {
					EncodeRune(p, c.r)
				}
			})

			b.Run("implem=EncodeRune4", func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					EncodeRune4(p, c.r)
				}
			})
		})
	}
}

func BenchmarkAppendRune(b *testing.B) {
	b.Skip("deprecated")
	for _, c := range runeCases {
		b.Run(c.name, func(b *testing.B) {
			var p [4]byte

			b.Run("implem=stdlib", func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					utf8.AppendRune(p[:0], c.r)
				}
			})

			b.Run("implem=local", func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					AppendRune(p[:0], c.r)
				}
			})
		})
	}
}

func BenchmarkUTF8Bytes(b *testing.B) {
	b.Skip("deprecated")
	p := new([4]byte)
	for _, c := range runeCases {
		b.Run(c.name, func(b *testing.B) {
			b.Run("implem=UTF8Bytes", func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					UTF8Bytes(p, c.r)
				}
			})
		})
	}
}
