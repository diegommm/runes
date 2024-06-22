package runes

import (
	"testing"
	"unicode/utf8"
)

func BenchmarkValidRune(b *testing.B) {
	for _, c := range runeCases {
		b.Run(c.name, func(b *testing.B) {
			b.Run("stdlib", func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					utf8.ValidRune(c.r)
				}
			})
			b.Run("local", func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					ValidRune(c.r)
				}
			})
		})
	}
}
