package runes

import (
	"strings"
	"sync"
	"testing"
	"unicode"
	"unicode/utf8"
)

var isInSetImplems = map[string]*struct {
	isInSetMaker func(string) func(rune) bool
	cond         func(start, span rune, count int) bool // optional, true if nil
}{
	"isStrategy": {
		isInSetMaker: func(s string) func(rune) bool {
			return isStrategy(strings.NewReader(s))
		},
	},
	"isSlice": {
		isInSetMaker: func(s string) func(rune) bool {
			return isSlice([]rune(s))
		},
	},
	"isMap": {
		isInSetMaker: func(s string) func(rune) bool {
			return isMap(stringToRuneMap(s))
		},
	},
	"isMask64": {
		isInSetMaker: func(s string) func(rune) bool {
			rr := strings.NewReader(s)
			minRune, _, _ := startSpanCount(rr)
			rewindRuneReader(rr)
			return isMask64(rr, minRune)
		},
		cond: func(start, span rune, count int) bool {
			return span < 64
		},
	},
	"isMask32": {
		isInSetMaker: func(s string) func(rune) bool {
			rr := strings.NewReader(s)
			minRune, _, _ := startSpanCount(rr)
			rewindRuneReader(rr)
			return isMask32(rr, minRune)
		},
		cond: func(start, span rune, count int) bool {
			return span < 32
		},
	},
	"isMaskSlice32": {
		isInSetMaker: func(s string) func(rune) bool {
			rr := strings.NewReader(s)
			minRune, span, _ := startSpanCount(rr)
			rewindRuneReader(rr)
			return isMaskSlice32(rr, minRune, span)
		},
	},
	"isMaskSlice64": {
		isInSetMaker: func(s string) func(rune) bool {
			rr := strings.NewReader(s)
			minRune, span, _ := startSpanCount(rr)
			rewindRuneReader(rr)
			return isMaskSlice64(rr, minRune, span)
		},
	},
	"isSparseSet": {
		isInSetMaker: func(s string) func(rune) bool {
			return isSparseSet([]rune(s))
		},
	},
}

type isFuncTestCase struct {
	f     func(rune) bool
	runes []rune
}

var (
	unicodeIsFuncsDataOnce sync.Once
	unicodeIsFuncsMap      = map[string]*isFuncTestCase{
		// Keep the map key of non-stdlib funcs starting with lowercase
		"nothing": {
			f:     isNeverIn,
			runes: []rune{},
		},
		"tiny": {
			f: func(r rune) bool {
				return r >= 0 && r < 32
			},
			runes: validRuneSlice(0, 32),
		},
		"ascii": {
			f: func(r rune) bool {
				return r >= 0 && r < utf8.RuneSelf
			},
			runes: validRuneSlice(0, utf8.RuneSelf),
		},

		// Keep the map key of stdlib funcs starting with uppercase
		"IsControl": {f: unicode.IsControl},
		"IsDigit":   {f: unicode.IsDigit},
		"IsGraphic": {f: unicode.IsGraphic},
		"IsLetter":  {f: unicode.IsLetter},
		"IsLower":   {f: unicode.IsLower},
		"IsMark":    {f: unicode.IsMark},
		"IsNumber":  {f: unicode.IsNumber},
		"IsPrint":   {f: unicode.IsPrint},
		"IsPunct":   {f: unicode.IsPunct},
		"IsSpace":   {f: unicode.IsSpace},
		"IsSymbol":  {f: unicode.IsSymbol},
		"IsTitle":   {f: unicode.IsTitle},
		"IsUpper":   {f: unicode.IsUpper},
	}
)

func generateUnicodeIsFuncsMap() {
	// brute, very brute indeed, but too lazy to do the fine work of getting
	// every rune analyzing the code
	unicodeIsFuncsDataOnce.Do(func() {
		var wg sync.WaitGroup
		defer wg.Wait()

		for _, x := range unicodeIsFuncsMap {
			if x.runes != nil {
				continue
			}
			wg.Add(1)
			go func() {
				defer wg.Done()
				for i := newValidRuneIterator(); i.Next(); {
					if x.f(i.Rune) {
						x.runes = append(x.runes, i.Rune)
					}
				}
			}()
		}
	})
}

func TestIsFuncsUnicode(t *testing.T) {
	t.Parallel()
	generateUnicodeIsFuncsMap()

	for name, isFunc := range unicodeIsFuncsMap {
		t.Run("case="+name, func(t *testing.T) {
			t.Parallel()

			minRune, span, count := startSpanCount(newRuneReadSeeker(isFunc.runes))
			str := string(isFunc.runes)

			for name, implem := range isInSetImplems {
				if implem.cond != nil && !implem.cond(minRune, span, count) {
					continue
				}

				t.Run("implem="+name, func(t *testing.T) {
					t.Parallel()

					isInSetFunc := implem.isInSetMaker(str)
					for _, r := range isFunc.runes {
						if !isInSetFunc(r) {
							t.Fatalf("unexpected false for rune %v", r)
						}
					}

					for _, c := range runeCases {
						if isInSetFunc(c.r) != isFunc.f(c.r) {
							t.Fatalf("differs from cannonical in rune %v", c.r)
						}
					}
				})
			}
		})
	}
}
