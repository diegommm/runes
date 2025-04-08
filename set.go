package runes

const (
	maxUint8  = 1<<8 - 1
	maxUint16 = 1<<16 - 1

	lsb5Mask = 1<<5 - 1
)

// Set is a set of runes.
type Set interface {
	// Contains returns whether the given rune is part of the set.
	Contains(rune) bool
}

// Sets is a [Set] that will use a linear search over its inner [Set] elements.
// Using a type parameter is useful to create a compact set of the same type to
// further reduce the memory footprint.
type Sets[T Set] []T

func (x Sets[T]) Contains(r rune) bool {
	for i := range x {
		if x[i].Contains(r) {
			return true
		}
	}
	return false
}

// LinearSlice is a [Set] uses a linear search in its `Contains` method.
type LinearSlice[T interface{ uint16 | rune }] struct {
	Values []T // must be sorted asc
}

func (x LinearSlice[T]) Contains(r rune) bool {
	if len(x.Values) == 0 || r < rune(x.Values[0]) || rune(x.Values[len(x.Values)-1]) < r {
		return false
	}
	return x.containsSlow(r)
}

func (x LinearSlice[T]) containsSlow(r rune) bool {
	for i := range x.Values {
		if r == rune(x.Values[i]) {
			return true
		}
	}
	return false
}

// BinarySlice is a [Set] uses a binary search in its `Contains` method.
type BinarySlice[T interface{ uint16 | rune }] struct {
	Values []T // must be sorted asc
}

func (x BinarySlice[T]) Contains(r rune) bool {
	if len(x.Values) == 0 || r < rune(x.Values[0]) || rune(x.Values[len(x.Values)-1]) < r {
		return false
	}
	return x.containsSlow(r)
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

// Range is the set of runes in the inclusive set [From, To].
type Range[T interface{ uint16 | rune }] struct {
	From, To T // `From` must be less than or equal to `To`
}

func (x Range[T]) Contains(r rune) bool {
	return r >= rune(x.From) && r <= rune(x.To)
}

// Uniform is a [Set] that contains `Count` runes uniformly distributed `Stride`
// apart from each other, starting at `First`. For a `Stride` value of 1, a
// `Range` is slightly more compact and faster.
type Uniform[T interface{ uint16 | rune }] struct {
	First  T
	Stride uint16
	Count  uint16
}

func (x Uniform[T]) Contains(r rune) bool {
	u, s, c := uint32(r-rune(x.First)), uint32(x.Stride), uint32(x.Count)
	return s > 0 && // always true, but removes runtime.panicdivide
		u < s*c && u%s == 0
}

// NewBitmap creates a [Bitmap] from the given runes, which must be sorted in
// ascending order.
func NewBitmap(rs []rune) (Bitmap, error) {
	if len(rs) == 0 {
		return "", &errString{"NewBitmap: no runes given"}
	}
	if len(rs) > maxUint16 {
		return "", &errString{"NewBitmap: too many elements"}
	}

	// allocate for the whole string
	span := uint32(rs[len(rs)-1] + 2 - rs[0])
	bin := make([]byte, stringBitmapHeaderLen+ceilDiv(span, 8))

	// encode header
	encodeFixedRune((*[3]byte)(bin), rs[0])
	encodeUint16((*[2]byte)(bin[3:]), uint16(len(rs)))

	// build the bitmap
	bm := bin[stringBitmapHeaderLen:]
	for _, r := range rs {
		u := uint32(r - rs[0])
		bm[u>>3] |= 1 << (u & 7)
	}

	return Bitmap(bin), nil
}

// Bitmap is a general purpose [Set] that uses an internal bitmap for constant
// time search that is also quite fast, ideal for patchy sets of runes with a
// non-trivial distribution.
type Bitmap string

const stringBitmapHeaderLen = 0 +
	3 + // first rune encoded in 3 bytes
	2 // number of runes as a uint16

func (x Bitmap) Contains(r rune) bool {
	if len(x) <= stringBitmapHeaderLen {
		return false
	}
	u := uint32(r - decodeFixedRune(x[0], x[1], x[2]))
	i := stringBitmapHeaderLen + int(u>>3)
	// why having runtime.panicIndex is faster here???
	return i < len(x) && 1<<(u&7)&x[i] != 0
}

type errString struct {
	string
}

func (e *errString) Error() string {
	return e.string
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

// encodeFixedRune encodes a rune in 3 bytes with little-endian. The rune should
// be no longer than 21 bits (only an invalid rune would be), and the 3 msb of
// the last byte are unused.
func encodeFixedRune(b *[3]byte, r rune) {
	b[0] = byte(r)
	b[1] = byte(r >> 8)
	b[2] = byte(r>>16) & lsb5Mask // only use the 5 lsb
}

// decodeFixedRune decodes a rune encoded with encodeFixedRune.
func decodeFixedRune(b0, b1, b2 byte) rune {
	return rune(b0) |
		rune(b1)<<8 |
		rune(b2&lsb5Mask)<<16 // discard the 3 msb
}

// encodeUint16 encodes a uint16 in 2 bytes with little-endian.
func encodeUint16(b *[2]byte, c uint16) {
	b[0] = byte(c)
	b[1] = byte(c >> 8)
}

// decodeUint16 decodes a uint16 encoded with encodeUint16.
func decodeUint16(b0, b1 byte) uint16 {
	return uint16(b0) |
		uint16(b1)<<8
}
