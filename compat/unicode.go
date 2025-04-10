package compat

import (
	"unicode"
	"unsafe"

	"github.com/diegommm/runes"
	"github.com/diegommm/runes/util"
)

// SizeofUnicodeRangeTable estimates the amount of bytes required to store a
// *unicode.RangeTable in memory.
func SizeofUnicodeRangeTable(rt *unicode.RangeTable) int {
	return int(unsafe.Sizeof(*rt)) +
		len(rt.R16)*int(unsafe.Sizeof(unicode.Range16{})) +
		len(rt.R32)*int(unsafe.Sizeof(unicode.Range32{}))
}

// FromUnicode converts a *unicode.RangeTable to a [RangeTables].
func FromUnicode(rt *unicode.RangeTable) RangeTables {
	ret := make(RangeTables, 0, len(rt.R16)+len(rt.R32))
	for _, rt := range rt.R16 {
		ret = append(ret, RangeTable{
			Lo:     rune(rt.Lo),
			Hi:     rune(rt.Hi),
			Stride: rune(rt.Stride),
		})
	}
	for _, rt := range rt.R32 {
		ret = append(ret, RangeTable{
			Lo:     rune(rt.Lo),
			Hi:     rune(rt.Hi),
			Stride: rune(rt.Stride),
		})
	}
	return ret
}

// RangeTables is an OrderedRunesList built from a *unicode.RangeTable.
type RangeTables []RangeTable

func (rts RangeTables) Len() int {
	var l int
	for i := range rts {
		l += rts[i].Len()
	}
	return l
}

func (rts RangeTables) Min() rune {
	if len(rts) == 0 {
		return 0
	}
	return rts[0].Min()
}

func (rts RangeTables) Max() rune {
	if len(rts) == 0 {
		return 0
	}
	return rts[len(rts)-1].Max()
}

func (rts RangeTables) Iter() runes.OrderedRunesIter {
	var i int
	return util.IterMergeFunc(func() runes.OrderedRunesIter {
		if ret := i; ret < len(rts) {
			i++
			return rts[ret].Iter()
		}
		return nil
	})
}

// RangeTable is an OrderedRunesList built from either a unicode.Range16 or
// unicode.Range32.
type RangeTable struct {
	Lo, Hi, Stride rune
}

func (rt RangeTable) Len() int {
	return 1 + (int(rt.Hi)-int(rt.Lo))/int(rt.Stride)
}

func (rt RangeTable) Min() rune {
	return rt.Lo
}

func (rt RangeTable) Max() rune {
	return rt.Hi
}

func (rt RangeTable) Iter() runes.OrderedRunesIter {
	i := rt.Lo
	return func() (rune, bool) {
		if ret := i; ret <= rt.Hi {
			i += rt.Stride
			return ret, true
		}
		return -1, false
	}
}
