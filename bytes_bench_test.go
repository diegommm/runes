package runes

import (
	"testing"
	"unicode/utf8"
)

func BenchmarkEncodeRune(b *testing.B) {
	for _, c := range runeCases {
		b.Run(c.name, func(b *testing.B) {
			p := make([]byte, 4)

			b.Run("stdlib", func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					utf8.EncodeRune(p, c.r)
				}
			})

			b.Run("local", func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					EncodeRune(p, c.r)
				}
			})

			b.Run("EncodeRune4", func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					EncodeRune4(p, c.r)
				}
			})
		})
	}
}

func BenchmarkAppendRune(b *testing.B) {
	for _, c := range runeCases {
		b.Run(c.name, func(b *testing.B) {
			var p [4]byte

			b.Run("stdlib", func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					utf8.AppendRune(p[:0], c.r)
				}
			})

			b.Run("local", func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					AppendRune(p[:0], c.r)
				}
			})
		})
	}
}

func BenchmarkUTF8Bytes(b *testing.B) {
	p := new([4]byte)
	for _, c := range runeCases {
		b.Run(c.name, func(b *testing.B) {
			b.Run("UTF8Bytes", func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					UTF8Bytes(p, c.r)
				}
			})
		})
	}
}
