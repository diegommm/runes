package runes

import "github.com/diegommm/runes/iface"

type Iterator = iface.Iterator

func ExpandIterator(i Iterator) []rune {
	ret := make([]rune, 0, i.RuneLen())
	for {
		r, ok := i.NextRune()
		if !ok {
			break
		}
		ret = append(ret, r)
	}
	return ret
}

// RunesIterator returns an iterator from the given slice of runes, which must
// be in sorted ascending order and non-repeating.
func RunesIterator(rs []rune) Iterator {
	if len(rs) > maxInt32 {
		panic("RunesIterator: too many runes")
	}
	return &runesIterator{rs: rs}
}

type runesIterator struct {
	rs  []rune
	pos uint32
}

func (x *runesIterator) NextRune() (rune, bool) {
	if pos := int(x.pos); pos < len(x.rs) {
		x.pos++
		return x.rs[pos], true
	}
	return 0, false
}

func (x *runesIterator) Max() rune {
	if len(x.rs) > 0 {
		return x.rs[len(x.rs)-1]
	}
	return -1
}

func (x *runesIterator) RuneLen() int32 { return int32(len(x.rs)) }
func (x *runesIterator) Restart()       { x.pos = 0 }

// RuneReader is the standard library's [io.RuneReader] interface.
type RuneReader interface {
	ReadRune() (r rune, size int, err error)
}

// ReadAllRunes returns all the runes read from the given [RuneReader]. It stops
// at the first error.
func ReadAllRunes(rr RuneReader) []rune {
	var l int
	if rrl, ok := rr.(interface{ Len() int }); ok {
		l = rrl.Len() // we will have at most this amount of runes
	}
	rs := make([]rune, 0, l)
	for {
		r, _, err := rr.ReadRune()
		if err != nil {
			break
		}
		rs = append(rs, r)
	}

	return rs
}

// RangeIterator returns an [Iterator] from the given [Range].
func RangeIterator(r Range) Iterator {
	return &rangeIterator{r: r}
}

type rangeIterator struct {
	r   Range
	pos int32
}

func (x *rangeIterator) NextRune() (rune, bool) {
	if r := x.r.Nth(x.pos); r >= 0 {
		x.pos++
		return r, true
	}
	return 0, false
}

func (x *rangeIterator) RuneLen() int32 { return x.r.RuneLen() - x.pos }
func (x *rangeIterator) Max() rune      { return x.r.Max() }
func (x *rangeIterator) Restart()       { x.pos = 0 }

// Merge merges multiple [Range] values into a single iterator.
func Merge(rs ...Range) Iterator {
	is := make([]Iterator, len(rs))
	for i := range rs {
		is[i] = RangeIterator(rs[i])
	}
	return MergeIterators(is...)
}

// MergeIterators merges multiple iterators in a consistent manner.
func MergeIterators(is ...Iterator) Iterator {
	var l int32
	for i := range is {
		l += is[i].RuneLen()
	}
	rs := make([]rune, 0, l) // we will have at most `l` runes

	// initialize list of iterators runes
	isRunes := make([]rune, 0, len(is))
	for i := len(is) - 1; i >= 0; i-- {
		r, ok := is[i].NextRune()
		if !ok {
			removeAtIndex(&is, i)
			continue
		}
		isRunes = append(isRunes, r)
	}

	for len(isRunes) > 0 {
		// find the smallest rune and append it to our list
		var minIdx int
		minRune := isRunes[minIdx]
		for i, r := range isRunes {
			if r < minRune {
				minRune, minIdx = r, i
			}
		}
		rs = append(rs, minRune)

		// prune the smallest rune
		for i := len(isRunes) - 1; i >= minIdx; i-- {
			if isRunes[i] == minRune {
				r, ok := is[i].NextRune()
				if !ok {
					removeAtIndex(&is, i)
					removeAtIndex(&isRunes, i)
					continue
				}
				isRunes[i] = r
			}
		}
	}

	return RunesIterator(rs)
}
