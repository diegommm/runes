package runes

import (
	"fmt"
	"io"
	"unicode/utf8"
)

// RuneIterator iterates over runes and provides a method to restart the
// iterator. The runes returned by the iterator must be ordered in ascending
// order.
type RuneIterator interface {
	io.RuneReader
	Restart()
}

// ResetRuneIterator allows restarting an iterator using a different source.
type ResetRuneIterator[T any] interface {
	RuneIterator
	Reset(T)
}

// NewRangeIterator returns a [ResetRuneIterator] for [Range].
func NewRangeIterator(r Range) ResetRuneIterator[Range] {
	ri := new(rangeIterator)
	ri.Reset(r)
	return ri
}

type rangeIterator struct {
	r        Range
	pos, len int
}

func (r3 *rangeIterator) ReadRune() (r rune, size int, err error) {
	if r3.pos >= r3.len {
		return readRuneEOF()
	}
	r, ok := r3.r.Nth(r3.pos)
	if !ok {
		return 0, 0, fmt.Errorf("rune at position %d unexpectedly not found",
			r3.pos)
	}
	r3.pos++
	return readRuneReturn(r)
}

func (r3 *rangeIterator) Restart() {
	r3.pos = 0
}

func (r3 *rangeIterator) Reset(r Range) {
	r3.pos = 0
	r3.len = r.Len()
}

// NewRuneSliceRuneIterator returns a [ResetRuneIterator] for slices of runes.
func NewRuneSliceRuneIterator(runes []rune) ResetRuneIterator[[]rune] {
	return &runeReadRestarter{
		runes: runes,
	}
}

type runeReadRestarter struct {
	runes []rune
	pos   int
}

func (r3 *runeReadRestarter) ReadRune() (r rune, size int, err error) {
	if r3.pos >= len(r3.runes) {
		return readRuneEOF()
	}
	r = r3.runes[r3.pos]
	r3.pos++
	return readRuneReturn(r)
}

func (r3 *runeReadRestarter) Restart() {
	r3.pos = 0
}

func (r3 *runeReadRestarter) Reset(runes []rune) {
	r3.runes = runes
	r3.pos = 0
}

// RuneReadSeekerToRestarter returns a [RuneIterator] that restarts by
// seeking to the start, discarding the error if any is found. This can be used
// to wrap a [bytes.Reader] or [strings.Reader].
func RuneReadSeekerToRestarter(rrs interface {
	io.RuneReader
	io.Seeker
}) RuneIterator {
	return runeReadSeekerWrapper{
		runeReadSeeker: rrs,
	}
}

type runeReadSeeker interface {
	io.RuneReader
	io.Seeker
}

type runeReadSeekerWrapper struct {
	runeReadSeeker
}

func (r runeReadSeekerWrapper) Restart() {
	r.runeReadSeeker.Seek(0, io.SeekStart)
}

func readRuneReturn(r rune) (rune, int, error) {
	return r, utf8.RuneLen(r), nil
}

func readRuneEOF() (rune, int, error) {
	return 0, 0, io.EOF
}
