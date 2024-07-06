package runes

import (
	"fmt"
	"slices"
	"testing"
	"unicode"

	"github.com/diegommm/runes/test"
)

func benchmarkContains(b *testing.B, implem string, f func(rune) bool, extraRunes ...rune) {
	testRunes := make([]rune, 0, len(test.RuneCases)+len(extraRunes))
	testRunes = append(testRunes, extraRunes...)
	for _, rc := range test.RuneCases {
		testRunes = append(testRunes, rc.Rune)
	}
	slices.Sort(testRunes)
	slices.Compact(testRunes)

	for _, r := range testRunes {
		b.Run(fmt.Sprintf("rune=%v", r), func(b *testing.B) {
			b.Run("implem="+implem, func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					f(r)
				}
			})
		})
	}
}

func BenchmarkUniformRange_Contains(b *testing.B) {
	benchCases := []struct {
		min    rune
		count  uint16
		stride byte
		name   string
	}{
		{7, 1, 1, "1 value of 1 byte"},
		{256, 1, 1, "1 value of 2 bytes"},
		{65536, 1, 1, "1 value of 3 bytes"},
		{7, 6, 1, "6 contiguous values"},
		{7, 3, 5, "3 values spaced every 5"},
	}

	for _, bc := range benchCases {
		b.Run("case="+bc.name, func(b *testing.B) {
			r5, err5 := NewUniformRange5(bc.min, bc.count, bc.stride)
			r8, err8 := NewUniformRange68(bc.min, bc.count, uint16(bc.stride))
			if err5 != nil || err8 != nil {
				b.Fatalf("error creating uniform ranges: err5: %v, err8: %v",
					err5, err8)
			}
			//r16 := Range16Contains(NewRange16(bc.min, bc.count, bc.stride))
			//r32 := Range32Contains(NewRange32(bc.min, bc.count, bc.stride))

			maxRune := r8.Max()

			if bc.min >= 0 && bc.min < maxUint16 {
				r6, err6 := NewUniformRange68(uint16(bc.min), bc.count, uint16(bc.stride))
				if err6 != nil {
					b.Fatalf("error creating uniform ranges: err6: %v", err6)
				}
				benchmarkContains(b, "uniformRange6", r6.Contains, bc.min, maxRune)
			}

			if bc.count == 1 {
				switch u := uint32(bc.min); {
				case u <= maxUint8:
					v1 := NewOneValueRange[OneValueRange1](bc.min).Contains
					benchmarkContains(b, "oneValueRange1", v1, bc.min, maxRune)
				case u <= maxUint16:
					v1 := NewOneValueRange[OneValueRange2](bc.min).Contains
					benchmarkContains(b, "oneValueRange2", v1, bc.min, maxRune)
				default:
					v1 := NewOneValueRange[OneValueRange3](bc.min).Contains
					benchmarkContains(b, "oneValueRange3", v1, bc.min, maxRune)
					v1 = NewOneValueRange[OneValueRange4](bc.min).Contains
					benchmarkContains(b, "oneValueRange4", v1, bc.min, maxRune)
				}
			}
			benchmarkContains(b, "uniformRange5", r5.Contains, bc.min, maxRune)
			benchmarkContains(b, "uniformRange8", r8.Contains, bc.min, maxRune)
			//benchmarkContains(b, "unicode.Range16", r16, bc.min, maxRune)
			//benchmarkContains(b, "unicode.Range32", r32, bc.min, maxRune)
		})
	}
}

func BenchmarkFixedRuneEncoding(b *testing.B) {
	b.Run("case=encodeFixedRune", func(b *testing.B) {
		for _, rc := range test.RuneCases {
			r := rc.Rune
			b.Run(fmt.Sprintf("rune=%v", r), func(b *testing.B) {
				var runeBytes [3]byte
				for i := 0; i < b.N; i++ {
					encodeFixedRune(&runeBytes, r)
				}
			})
		}
	})

	b.Run("case=encodeFixedRune", func(b *testing.B) {
		for _, rc := range test.RuneCases {
			var rb [3]byte
			encodeFixedRune(&rb, rc.Rune)
			b.Run(fmt.Sprintf("rune=%v", rc.Rune), func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					decodeFixedRune(rb[0], rb[1], rb[2])
				}
			})
		}
	})

	b.Run("case=comparison", func(b *testing.B) {
		b.Run("implem=decode-compare", func(b *testing.B) {
			for _, rc1 := range test.RuneCases {
				r1 := rc1.Rune
				if uint32(r1) > unicode.MaxRune {
					continue
				}
				var r1b [3]byte
				encodeFixedRune(&r1b, r1)
				b.Run(fmt.Sprintf("rune1=%v", r1), func(b *testing.B) {
					for _, rc2 := range test.RuneCases {
						r2 := rc2.Rune
						b.Run(fmt.Sprintf("rune2=%v", r2), func(b *testing.B) {
							for i := 0; i < b.N; i++ {
								if (r1 == r2) != (decodeFixedRune(r1b[0], r1b[1], r1b[2]) == r2) {
									b.Fatalf("invalid comparison")
								}
							}
						})
					}
				})
			}
		})

		b.Run("implem=encode-compare", func(b *testing.B) {
			for _, rc1 := range test.RuneCases {
				r1 := rc1.Rune
				if uint32(r1) > unicode.MaxRune {
					continue
				}
				var r1b [3]byte
				encodeFixedRune(&r1b, r1)
				b.Run(fmt.Sprintf("rune1=%v", r1), func(b *testing.B) {
					for _, rc2 := range test.RuneCases {
						r2 := rc2.Rune
						b.Run(fmt.Sprintf("rune2=%v", r2), func(b *testing.B) {
							for i := 0; i < b.N; i++ {
								var r2b [3]byte
								encodeFixedRune(&r2b, r2)
								if (r1 == r2) != (r1b == r2b) {
									b.Fatalf("invalid comparison")
								}
							}
						})
					}
				})
			}
		})

		b.Run("implem=compareWhileEncoding", func(b *testing.B) {
			for _, rc1 := range test.RuneCases {
				r1 := rc1.Rune
				if uint32(r1) > unicode.MaxRune {
					continue
				}
				var r1b [3]byte
				encodeFixedRune(&r1b, r1)
				b.Run(fmt.Sprintf("rune1=%v", r1), func(b *testing.B) {
					for _, rc2 := range test.RuneCases {
						r2 := rc2.Rune
						b.Run(fmt.Sprintf("rune2=%v", r2), func(b *testing.B) {
							for i := 0; i < b.N; i++ {
								if (r1 == r2) != compareWhileEncoding(r2, r1b[0], r1b[1], r1b[2]) {
									b.Fatalf("invalid comparison")
								}
							}
						})
					}
				})
			}
		})

		b.Run("implem=compareWhileDecoding", func(b *testing.B) {
			for _, rc1 := range test.RuneCases {
				r1 := rc1.Rune
				if uint32(r1) > unicode.MaxRune {
					continue
				}
				var r1b [3]byte
				encodeFixedRune(&r1b, r1)
				b.Run(fmt.Sprintf("rune1=%v", r1), func(b *testing.B) {
					for _, rc2 := range test.RuneCases {
						r2 := rc2.Rune
						b.Run(fmt.Sprintf("rune2=%v", r2), func(b *testing.B) {
							for i := 0; i < b.N; i++ {
								if (r1 == r2) != compareWhileDecoding(r2, r1b[0], r1b[1], r1b[2]) {
									b.Fatalf("invalid comparison")
								}
							}
						})
					}
				})
			}
		})
	})
}
