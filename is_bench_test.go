package runes

import (
	"fmt"
	"slices"
	"strings"
	"testing"
	"unicode"
)

func BenchmarkIsInSet(b *testing.B) {
	isInSetBenchmarkInputs := []struct {
		name  string
		input string
	}{
		{name: "zero", input: ""},
		{name: "ASCII", input: "A"},
		{name: "Non-ASCII", input: "Ñ"},
		{name: "ASCII", input: "AB"},
		{name: "Non-ASCII", input: "Ñ世"},
		{name: "ASCII", input: makeContiguousRunesString(' ', 64)},
		{name: "Non-ASCII", input: makeContiguousRunesString('Ñ', 64)},
		{name: "ASCII", input: makeContiguousRunesString(' ', 65)},
		{name: "Non-ASCII", input: makeContiguousRunesString('Ñ', 65)},
		{name: "ASCII", input: makeContiguousRunesString(0, 127)},
		{name: "Non-ASCII", input: makeContiguousRunesString('Ñ', 127)},
		{name: "ASCII+", input: makeContiguousRunesString(0, 128)},
		{name: "Other", input: makeContiguousRunesString(1024, 256)},
		{name: "Other", input: makeContiguousRunesString(1024, 512)},
		{name: "Other", input: makeContiguousRunesString(1024, 1025)},
		{
			name:  "Full Unicode span, few items",
			input: string([]rune{0, 1, 2, '\U0010FFFF'}),
		},
		{
			name:  "Half Unicode span, few items",
			input: string([]rune{0, 1, 2, '\U0010FFFF' / 2}),
		},
	}

	for name, implem := range isInSetImplems {
		b.Run("implem="+name, func(b *testing.B) {
			for _, bi := range isInSetBenchmarkInputs {
				start, span, count := startSpanCount(strings.NewReader(bi.input))
				if implem.cond != nil && !implem.cond(start, span, count) {
					continue
				}
				isInSetFunc := implem.isInSetMaker(bi.input)

				b.Run(fmt.Sprint("count=", count), func(b *testing.B) {
					b.Run(fmt.Sprint("span=", span), func(b *testing.B) {
						b.Run(fmt.Sprint("start=", start), func(b *testing.B) {

							b.Run("case=init", func(b *testing.B) {
								b.ReportAllocs()
								for i := 0; i < b.N; i++ {
									if f := implem.isInSetMaker(bi.input); f == nil {
										b.FailNow()
									}
								}
							})

							for _, rc := range runeCases {
								b.Run("case="+rc.name, func(b *testing.B) {
									b.ReportAllocs()
									for i := 0; i < b.N; i++ {
										isInSetFunc(rc.r)
									}
								})
							}

						})
					})
				})
			}
		})
	}
}

func BenchmarkStdlib(b *testing.B) {
	generateUnicodeIsFuncsMap()

	baseTestRunes := make([]rune, len(runeCases))
	for i, rc := range runeCases {
		baseTestRunes[i] = rc.r
	}

	for name, bench := range unicodeIsFuncsMap {
		if unicode.IsLower(rune(name[0])) {
			continue // skip dumb funcs
		}

		b.Run(name, func(b *testing.B) {
			f := IsInRunesSet(bench.runes)
			testRunes := append(baseTestRunes, bench.runes[0],
				bench.runes[len(bench.runes)-1])
			slices.Sort(testRunes)
			slices.Compact(testRunes)

			for _, r := range testRunes {
				b.Run(fmt.Sprintf("rune=%v", r), func(b *testing.B) {
					b.Run("implem=stdlib", func(b *testing.B) {
						for i := 0; i < b.N; i++ {
							bench.f(r)
						}
					})
					b.Run("implem=local", func(b *testing.B) {
						for i := 0; i < b.N; i++ {
							f(r)
						}
					})
				})
			}
		})
	}
}
