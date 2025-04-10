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
// Its elements must be sorted in ascending order.
type LinearSlice[T RuneT] []T

func (x LinearSlice[T]) Contains(r rune) bool {
	return len(x) > 0 &&
		r >= rune(x[0]) &&
		r <= rune(x[len(x)-1]) &&
		x.containsSlow(r)
}

func (x LinearSlice[T]) containsSlow(r rune) bool {
	for i := range x {
		if r == rune(x[i]) {
			return true
		}
	}
	return false
}

// BinarySlice is a [Set] that uses a binary search in its `Contains` method.
// Its elements must be sorted in ascending order.
type BinarySlice[T RuneT] []T

func (x BinarySlice[T]) Contains(r rune) bool {
	return len(x) > 0 &&
		r >= rune(x[0]) &&
		r <= rune(x[len(x)-1]) &&
		x.containsSlow(r)
}

func (x BinarySlice[T]) containsSlow(r rune) bool {
	i, j := uint32(0), uint32(len(x)-1)
	for h := u32Mid(i, j); i <= j && int(h) < len(x); h = u32Mid(i, j) {
		switch v := rune(x[h]); {
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
	Count  C // must be non-negative
	Stride S // must be non-negative
}

func (x Uniform[F, S, C]) Contains(r rune) bool {
	u, s, c := uint32(r-rune(x.First)), uint32(x.Stride), uint32(x.Count)
	return s > 0 && u < s*c && u%s == 0
}

// OrderedRunesIter is a function that iterates over a list of runes in
// ascending order, returning a rune and true while there are runes available.
// When there are no more runes, it returns zero value and false indefinitely.
type OrderedRunesIter = func() (rune, bool)

// OrderedRunesList is a list of ordered runes.
type OrderedRunesList interface {
	Min() rune // zero value if list is empty, immutable
	Max() rune // zero value if list is empty, immutable
	Len() int  // number of items in the list, immutable
	Iter() OrderedRunesIter
}

// NewBitmap creates a [Bitmap] from the given runes, which must be sorted in
// ascending order.
func NewBitmap(ri OrderedRunesList) Bitmap {
	if ri == nil || ri.Len() == 0 {
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
	it := ri.Iter()
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
