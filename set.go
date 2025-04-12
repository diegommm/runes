package runes

const MaxUint32 = 1<<32 - 1

// Set is a set of runes.
type Set interface {
	// Contains returns whether the given rune is part of the set.
	Contains(rune) bool
}

// MinMaxSet is a set with additional methods for internal optimizations.
type MinMaxSet interface {
	Set
	Min() uint32 // MaxUint32 if Set is empty
	Max() uint32 // MaxUint32 if Set is empty
}

// RuneT are the types with which runes of different width can be represented
// without losing information.
type RuneT interface {
	uint8 | uint16 | uint32 | rune
}

// Union is a [Set] that represents the union of its elements. The first element
// must have the smallest rune, and the last element the biggest.
type Union[T MinMaxSet] []T

func (x Union[T]) Contains(r rune) bool {
	if len(x) == 0 { // || uint32(r) > x[len(x)-1].Max() {
		return false
	}
	for i := range x {
		if x[i].Contains(r) {
			return true
		}
		//if uint32(r) < x[i].Min() {
		//return false
		//}
	}
	return false
}

func (x Union[T]) Min() uint32 {
	if len(x) == 0 {
		return MaxUint32
	}
	return x[0].Min()
}

func (x Union[T]) Max() uint32 {
	if len(x) == 0 {
		return MaxUint32
	}
	return x[len(x)-1].Max()
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

func (x LinearSlice[T]) Min() uint32 {
	if len(x) == 0 {
		return MaxUint32
	}
	return uint32(x[0])
}

func (x LinearSlice[T]) Max() uint32 {
	if len(x) == 0 {
		return MaxUint32
	}
	return uint32(x[len(x)-1])
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

func (x BinarySlice[T]) Min() uint32 {
	if len(x) == 0 {
		return MaxUint32
	}
	return uint32(x[0])
}

func (x BinarySlice[T]) Max() uint32 {
	if len(x) == 0 {
		return MaxUint32
	}
	return uint32(x[len(x)-1])
}

// Interval is the set of runes in the interval [From, To].
type Interval[T RuneT] struct {
	From, To T // `From` must be less than or equal to `To`
}

func (x Interval[T]) Contains(r rune) bool {
	return r >= rune(x.From) && r <= rune(x.To)
}

func (x Interval[T]) Min() uint32 {
	return uint32(x.From)
}

func (x Interval[T]) Max() uint32 {
	return uint32(x.To)
}

// Uniform is a [Set] that contains the runes uniformly distributed `Stride`
// apart from each other, starting at `First` and ending in `Hi`.
type Uniform[T RuneT] struct {
	Lo     T
	Hi     T // must be >= Lo
	Stride T // must be positive
}

func (x Uniform[T]) Contains(r rune) bool {
	v, stride := uint32(r-rune(x.Lo)), uint32(x.Stride)
	return v < uint32(x.Hi) && stride > 0 && v%stride == 0
}

func (x Uniform[T]) Min() uint32 {
	return uint32(x.Lo)
}

func (x Uniform[T]) Max() uint32 {
	return uint32(x.Hi)
}

// NewBitmap creates a [Bitmap] from the given runes, which must be sorted in
// ascending order.
func NewBitmap(rs []rune) Bitmap {
	if len(rs) == 0 {
		return ""
	}
	bm := make([]byte, bitmapLen(rs))
	writeBitmapHeader(bm, rs)
	writeBitmapBody(bm[bmHdrLen:], rs)

	return Bitmap(bm)
}

func bitmapLen(rs []rune) uint32 {
	return bmHdrLen + ceilDiv(uint32(rs[len(rs)-1]-rs[0]+1), 8)
}

func writeBitmapHeader(bm []byte, rs []rune) {
	hdr := (*[bmHdrLen]byte)(bm)
	hdr[0], hdr[1], hdr[2] = bmEncodeMinRune(rs[0])
	hdr[2] &= lsb5 // ensure an incorrect min rune does not break encoding
	// encode the highest 1 in the last byte corresponding to Max()
	posOfHighestBitInLastByte := byte(uint32(rs[len(rs)-1]-rs[0]) & 7)
	hdr[2] |= posOfHighestBitInLastByte << lsb5
}

func writeBitmapBody(bmBody []byte, rs []rune) {
	for _, r := range rs {
		u := uint32(r - rs[0])
		bmBody[u>>3] |= 1 << (u & 7)
	}
}

// Bitmap is a general purpose [Set] that uses an internal bitmap for fast and
// constant time search.
type Bitmap string

// bmHdrLen is the length of the header of a Bitmap that contains the first
// rune, encoded as 3 bytes in little-endian. Even the rune type being an int32,
// utf8.MaxRune is only 3 bytes long, so a valid rune will never take more than
// that. The following is a way of seeing the header:
//
//	AAAAAAAA AAAAAAAA BBBAAAAA
//
// Where the 21 A bits are the little-endian representation of the first rune
// (enough to represent utf8.MaxRune), and the 3 B bits are the position of the
// bit representing the max rune within the last byte of the bitmap.
const bmHdrLen = 3

const lsb5 = 0b00011111

func (x Bitmap) Contains(r rune) bool {
	if len(x) < bmHdrLen {
		return false
	}
	u := uint32(r - bmDecodeMinRune(x[0], x[1], x[2]&lsb5))
	i := bmHdrLen + u>>3
	return int(i) < len(x) && 1<<(u&7)&x[i] != 0
}

func (x Bitmap) Min() uint32 {
	if len(x) < bmHdrLen {
		return MaxUint32
	}
	return uint32(bmDecodeMinRune(x[0], x[1], x[2]&lsb5))
}

func (x Bitmap) Max() uint32 {
	switch {
	case len(x) > bmHdrLen:
		return uint32(bmDecodeMinRune(x[0], x[1], x[2]&lsb5)) +
			uint32(len(x)<<3) +
			uint32(x[2]>>lsb5)
	case len(x) == bmHdrLen:
		return uint32(bmDecodeMinRune(x[0], x[1], x[2]&lsb5))
	default:
		return MaxUint32
	}
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
func bmEncodeMinRune(r rune) (b0, b1, b2 byte) {
	return byte(r), byte(r >> 8), byte(r >> 16)
}

// bmDecodeMinRune decodes a rune encoded with bmEncodeMinRune.
func bmDecodeMinRune(b0, b1, b2 byte) rune {
	return rune(b0) |
		rune(b1)<<8 |
		rune(b2)<<16
}
