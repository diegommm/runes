package compat

import (
	"testing"
	"unicode"
	"unicode/utf8"

	"github.com/diegommm/runes/util"
)

func TestRangeTablesCompat(t *testing.T) {
	t.Parallel()
	testCases := map[string]*unicode.RangeTable{
		"White_Space": unicode.White_Space,
		"Upper":       unicode.Upper,
		"Lower":       unicode.Lower,
		"Letter":      unicode.Letter,
		"Mark":        unicode.Mark,
		"Number":      unicode.Number,
		"Punct":       unicode.Punct,
		"Symbol":      unicode.Symbol,
	}

	for name, urt := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			rt := FromUnicode(urt)

			for i := range rt {
				util.Equal(t, true, rt[i].Lo >= 0, "index=%d; negative Lo", i)
				util.Equal(t, true, rt[i].Hi >= 0, "index=%d; negative Hi", i)
				util.Equal(t, true, rt[i].Stride > 0, "index=%d; non-positive Stride", i)
				util.Equal(t, true, rt[i].Lo <= rt[i].Hi, "index=%d; Lo>Hi", i)
				if i > 0 {
					util.Equal(t, true, rt[i-1].Hi < rt[i].Lo, "index=%d; overlapping", i)
				}
			}

			var r rune
			i := 0
			for r = range util.IterSeq(rt.Iter()) {
				if i == 0 {
					util.Equal(t, rt.Min(), r, "invalid Min()")
				}
				i++
				util.MustEqual(t, true, unicode.Is(urt, r), "0x%x", r)
			}
			util.Equal(t, rt.Max(), r, "invalid Max()")
			util.Equal(t, i, rt.Len(), "invalid Len()")

			inverse := util.IterExcept(util.Seq[rune]{0, utf8.MaxRune, 1}.Iter(), rt.Iter())
			for r := range util.IterSeq(inverse) {
				util.MustEqual(t, false, unicode.Is(urt, r), "0x%x", r)
			}

			ri := rt.Iter()
			r1, ok := ri()
			i = 0
			prev := rune(-1)
			for r2 := range util.IterSeq(rt.Iter()) {
				util.MustEqual(t, true, ok, "index=%d; Iter run out", i)
				util.MustEqual(t, r1, r2, "index=%d; Iter and All differ", i)
				util.MustEqual(t, true, prev < r2, "index=%d; not in ascending order", i)
				i++
				r1, ok = ri()
			}
			util.Equal(t, false, ok, "Iter() should have been exhausted")
			util.Equal(t, 0, r1, "Iter() should return zero when exhausted")
		})
	}
}
