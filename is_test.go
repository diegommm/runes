package runes

import (
	"sync"
	"testing"
	"unicode"
	"unicode/utf8"
)

type isFuncTestCase struct {
	f     func(rune) bool
	runes []rune
}

type isFuncTested struct {
	name string
	f    func(rune) bool
}

var (
	unicodeIsFuncsDataOnce sync.Once
	validUTF8              []rune
	invalidUTF8            []rune
	unicodeIsFuncsMap      = map[string]*isFuncTestCase{
		// Keep the map key of dumb funcs starting with lowercase
		"tiny": {f: func(r rune) bool {
			return r < 32
		}},
		"byte": {f: func(r rune) bool {
			return r < utf8.RuneSelf
		}},

		// make sure to set a big timeout
		"limits32": {f: func(r rune) bool {
			m := r % 32
			return m == 0 || m == 31
		}},

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
		"ValidRune": {f: utf8.ValidRune},
	}
)

func generateUnicodeIsFuncsMap() {
	// brute, very brute indeed, but too lazy to do the fine work of getting
	// every rune analyzing the code
	unicodeIsFuncsDataOnce.Do(func() {
		validUTF8 = make([]rune, 0, utf8.MaxRune)
		for r := rune(0); r < utf8.MaxRune; r++ {
			if utf8.ValidRune(r) {
				validUTF8 = append(validUTF8, r)
			} else {
				invalidUTF8 = append(validUTF8, r)
			}
		}

		var wg sync.WaitGroup
		defer wg.Wait()
		for _, x := range unicodeIsFuncsMap {
			if len(x.runes) != 0 {
				panic("use generateUnicodeIsFuncsMap instead")
			}
			wg.Add(1)
			go func() {
				defer wg.Done()
				for _, r := range validUTF8 {
					if x.f(r) {
						x.runes = append(x.runes, r)
					}
				}
			}()
		}

		unicodeIsFuncsMap["nothing"] = &isFuncTestCase{f: isNeverIn}
	})
}

func TestIsFuncsUnicode(t *testing.T) {
	t.Parallel()

	generateUnicodeIsFuncsMap()

	buildFuncTestCases := func(s []rune) []isFuncTested {
		duped := append(s, s...)
		m := runeSliceToMap(s)
		str := string(s)
		b := []byte(str)
		minRune, span := minAndSpan(s)

		ret := []isFuncTested{
			{"IsStringFunc", IsStringFunc(str)},
			{"IsBytesFunc", IsBytesFunc(b)},
			{"IsRunesFunc", IsRunesFunc(duped)},
			{"isStrategy", isStrategy(s)},
			{"isMaskSlice64", isMaskSlice64(s, minRune, span)},
			{"isMaskSlice32", isMaskSlice32(s, minRune, span)},
			{"isMap", isMap(m)},
			{"isSlice", isSlice(s)},
			{"isSparseSet", isSparseSet(s)},
			{"isTable", isTable(s)},
			{"isString", isString(str)},
		}

		if span < 64 {
			ret = append(ret, isFuncTested{"isMask64", isMask64(s, minRune)})
		}

		if span < 32 {
			ret = append(ret, isFuncTested{"isMask32", isMask32(s, minRune)})
		}

		return ret
	}

	for name, isFunc := range unicodeIsFuncsMap {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			for _, tc := range buildFuncTestCases(isFunc.runes) {
				t.Run(tc.name, func(t *testing.T) {
					t.Parallel()
					for _, r := range validUTF8 {
						if tc.f(r) != isFunc.f(r) {
							t.Fatalf("failed for rune %v", r)
						}
					}
				})
			}
		})
	}
}
