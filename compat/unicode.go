package compat

import (
	"iter"
	"unicode"
	"unsafe"

	"github.com/diegommm/runes"
)

type Table interface {
	Len() int                     // 0 if empty
	Min() rune                    // -1 if empty
	Max() rune                    // -1 if empty
	All() iter.Seq[rune]          // sorted asc
	RuneIter() runes.RuneIterator // sorted asc
}

func SizeofUnicodeRangeTable(rt *unicode.RangeTable) int {
	return int(unsafe.Sizeof(*rt)) +
		len(rt.R16)*int(unsafe.Sizeof(unicode.Range16{})) +
		len(rt.R32)*int(unsafe.Sizeof(unicode.Range32{}))
}

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
		return -1
	}
	return rts[0].Min()
}

func (rts RangeTables) Max() rune {
	if len(rts) == 0 {
		return -1
	}
	return rts[len(rts)-1].Max()
}

func (rts RangeTables) All() iter.Seq[rune] {
	return func(yield func(rune) bool) {
		for i := range rts {
			rts[i].All()(yield)
		}
	}
}

func (rts RangeTables) RuneIter() runes.RuneIterator {
	var i int
	return runes.RuneIters(func() runes.RuneIterator {
		if ret := i; ret < len(rts) {
			i++
			return rts[ret].RuneIter()
		}
		return nil
	})
}

type RangeTable struct {
	Lo, Hi, Stride rune
}

func (rt RangeTable) Len() int {
	return (int(rt.Hi) - int(rt.Lo) + 1) / int(rt.Stride)
}
func (rt RangeTable) Min() rune { return rt.Lo }
func (rt RangeTable) Max() rune { return rt.Hi }

func (rt RangeTable) All() iter.Seq[rune] {
	return func(yield func(rune) bool) {
		for i := rt.Lo; i <= rt.Hi; i += rt.Stride {
			yield(i)
		}
	}
}

func (rt RangeTable) RuneIter() runes.RuneIterator {
	i := rt.Lo
	return func() (rune, bool) {
		if ret := i; ret <= rt.Hi {
			i += rt.Stride
			return ret, true
		}
		return -1, false
	}
}
