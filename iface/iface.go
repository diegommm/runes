package iface

// Set is a set of runes.
type Set interface {
	// Contains returns whether the given rune is part of the set.
	Contains(rune) bool
}

// Range represents an ordered list of runes in the range `[Min(), Max()]`.
// Positions start at zero. A Range is a building block for [Set]s, but need not
// be.
type Range interface {
	Set
	// RuneLen returns the number of runes in the range. This is a method
	// optimized to be very fast.
	RuneLen() int32
	// Min returns the numerically smallest rune in the range, or -1 if the
	// range is empty. This is a method optimized to be very fast.
	Min() rune
	// Max returns the numerically biggest rune in the range, or -1 if the range
	// is empty. This is a method optimized to be very fast.
	Max() rune
	// Pos returns the position of the given rune within the range, or -1 if
	// it's not found. This is best effort.
	Pos(rune) int32
	// Nth returns the N-th ordinal rune in the range, or -1 if the index is out
	// of bound. This is best effort.
	Nth(int32) rune
}

// Iterator returns a list of non-repeated runes in sorted ascending order.
type Iterator interface {
	// NextRune returns the next rune and true, or zero and false if there is no
	// next rune.
	NextRune() (rune, bool)
	// RuneLen returns the number of runes that haven't been returned yet.
	RuneLen() int32
	// Max returns the last rune that will be returned by the iterator, without
	// consuming it, or -1 if there are no more runes to return.
	Max() rune
	// Restart restarts the iterator.
	//
	// TODO: review if needed
	Restart()
}
