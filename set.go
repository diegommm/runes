package runes

const maxUint16 = 1<<16 - 1

// Set is a set of runes.
type Set interface {
	// Contains returns whether the given rune is part of the set.
	Contains(rune) bool
}

// RuneT are the types with which runes of different width can be represented
// without losing information.
type RuneT interface {
	uint8 | uint16 | rune
}

// Union is a [Set] that represents the union of its elements. Using a type
// parameter is useful to create a compact set of the same type to further
// reduce the memory footprint.
type Union[T Set] []T

func (x Union[T]) Contains(r rune) bool {
	for i := range x {
		if x[i].Contains(r) {
			return true
		}
	}
	return false
}

// LinearSlice is a [Set] that uses a linear search in its `Contains` method.
type LinearSlice[T RuneT] struct {
	Values []T // must be sorted asc
}

func (x LinearSlice[T]) Contains(r rune) bool {
	return len(x.Values) > 0 &&
		r >= rune(x.Values[0]) &&
		r <= rune(x.Values[len(x.Values)-1]) &&
		x.containsSlow(r)
}

func (x LinearSlice[T]) containsSlow(r rune) bool {
	for i := range x.Values {
		if r == rune(x.Values[i]) {
			return true
		}
	}
	return false
}

// BinarySlice is a [Set] that uses a binary search in its `Contains` method.
type BinarySlice[T RuneT] struct {
	Values []T // must be sorted asc
}

func (x BinarySlice[T]) Contains(r rune) bool {
	return len(x.Values) > 0 &&
		r >= rune(x.Values[0]) &&
		r <= rune(x.Values[len(x.Values)-1]) &&
		x.containsSlow(r)
}

func (x BinarySlice[T]) containsSlow(r rune) bool {
	i, j := uint32(0), uint32(len(x.Values)-1)
	for h := u32Mid(i, j); i <= j && int(h) < len(x.Values); h = u32Mid(i, j) {
		switch v := rune(x.Values[h]); {
		case r < v:
			j = h - 1
		case v < r:
			i = h + 1
		default:
			return true
		}
	}
	return false
}

// Interval is the set of runes in the interval [From, To].
type Interval[T RuneT] struct {
	From, To T // `From` must be less than or equal to `To`
}

func (x Interval[T]) Contains(r rune) bool {
	return r >= rune(x.From) && r <= rune(x.To)
}

// Uniform is a [Set] that contains `Count` runes uniformly distributed `Stride`
// apart from each other, starting at `First`. For a `Stride` value of 1, a
// `Interval` is slightly more compact and faster.
type Uniform[F, S, C RuneT] struct {
	First  F
	Count  C // must be positive
	Stride S // must be positive
}

func (x Uniform[F, S, C]) Contains(r rune) bool {
	u, s, c := uint32(r-rune(x.First)), uint32(x.Stride), uint32(x.Count)
	return s > 0 && u < s*c && u%s == 0
}

// RuneList allows efficient iteration over an otherwise unknown set of runes.
type RuneList interface {
	Min() rune // -1 if empty, immutable
	Max() rune // -1 if empty, immutable

	// RuneIter returns an iterotor that returns true as long as there are
	// items, and starts returning (0, false) afterward. See `compat`
	// sub-package for adapters.
	RuneIter() RuneIterator
}

// SliceRuneList is a [RuneList] backed by a []rune.
type SliceRuneList []rune

func (x SliceRuneList) Min() rune {
	if len(x) == 0 {
		return -1
	}
	return x[0]
}

func (x SliceRuneList) Max() rune {
	if len(x) == 0 {
		return -1
	}
	return x[len(x)-1]
}

func (x SliceRuneList) RuneIter() RuneIterator {
	var i int
	return func() (rune, bool) {
		if ret := i; ret < len(x) {
			i++
			return x[ret], true
		}
		return 0, false
	}
}

type RuneIterator = func() (rune, bool)

// RuneIters returns a rune iterator X that combines the runes produced by the
// rune iterators successively returned by `f`. When `f` returns nil, then X is
// done.
func RuneIters(f func() RuneIterator) RuneIterator {
	it := f()
	return func() (rune, bool) {
		for {
			if it == nil {
				return 0, false
			}
			if r, ok := it(); ok {
				return r, true
			}
			it = f()
		}
	}
}

// NewBitmap creates a [Bitmap] from the given runes, which must be sorted in
// ascending order.
func NewBitmap(ri RuneList) Bitmap {
	if ri == nil || ri.Min() < 0 {
		return ""
	}
	mn, mx := ri.Min(), ri.Max()

	// allocate for the whole string
	span := uint32(mx + 2 - mn)
	bin := make([]byte, bmHdrLen+ceilDiv(span, 8))

	// encode header
	bmEncodeMinRune((*[bmHdrLen]byte)(bin), mn)

	// build the bitmap
	bm := bin[bmHdrLen:]
	it := ri.RuneIter()
	for r, ok := it(); ok; r, ok = it() {
		u := uint32(r - mn)
		bm[u>>3] |= 1 << (u & 7)
	}

	return Bitmap(bin)
}

// Bitmap is a general purpose [Set] that uses an internal bitmap for fast and
// constant time search.
type Bitmap string

// bmHdrLen is the length of the header of a Bitmap that contains the first
// rune, encoded as 3 bytes in little-endian. Even the rune type being an int32,
// utf8.MaxRune is only 3 bytes long, so a valid rune will never take more than
// that.
const bmHdrLen = 3

func (x Bitmap) Contains(r rune) bool {
	if len(x) <= bmHdrLen {
		return false
	}
	u := uint32(r - bmDecodeMinRune(x[0], x[1], x[2]))
	i := bmHdrLen + int(u>>3)
	// why having runtime.panicIndex is faster here???
	return i < len(x) && 1<<(u&7)&x[i] != 0
}

// ceilDiv performs the integer division of two uint32, rounding to the next
// (bigger) integer.
func ceilDiv(dividend, divisor uint32) uint32 {
	return uint32((uint64(dividend) + uint64(divisor) - 1) / uint64(divisor))
}

// u32Mid returns the number in the middle of two uint32.
func u32Mid(a, b uint32) uint32 {
	return uint32((uint64(a) + uint64(b)) >> 1)
}

// bmEncodeMinRune encodes a rune in bmHdrLen bytes with little-endian.
func bmEncodeMinRune(b *[bmHdrLen]byte, r rune) {
	b[0] = byte(r)
	b[1] = byte(r >> 8)
	b[2] = byte(r >> 16)
}

// bmDecodeMinRune decodes a rune encoded with bmEncodeMinRune.
func bmDecodeMinRune(b0, b1, b2 byte) rune {
	return rune(b0) |
		rune(b1)<<8 |
		rune(b2)<<16
}
