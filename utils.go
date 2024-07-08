package runes

func Must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}

type sInt interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64
}

type uInt interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr
}

type xInt interface {
	sInt | uInt
}

func ceilDiv[T xInt](dividend, divisor T) T {
	return (dividend + divisor - 1) / divisor
}

func u32Half(a, b uint32) uint32 {
	return uint32((uint64(a) + uint64(b)) >> 1)
}

// leadingOnePos returns the position of the most-signigicant bit that is set in
// a byte.
func leadingOnePos(b byte) byte {
	var r byte
	if b > 1<<4-1 {
		r += 4
		b >>= 4
	}
	if b > 1<<2-1 {
		r += 2
		b >>= 2
	}
	if b > 1<<1-1 {
		r++
		b >>= 1
	}
	return r + b
}

// ones returns the number of bits set in a byte.
func ones(b byte) byte {
	return 0 +
		((b >> 7) & 1) + ((b >> 6) & 1) + ((b >> 5) & 1) + ((b >> 4) & 1) +
		((b >> 3) & 1) + ((b >> 2) & 1) + ((b >> 1) & 1) + (b & 1)
}

// nthOnePos returns the position of the n-th one set in the given byte, or zero
// if the bit is not found.
func nthOnePos(b, n byte) byte {
	for i := byte(0); i < 8; i++ {
		if b&(1<<i) != 0 {
			n--
			if n == 0 {
				return i + 1
			}
		}
	}
	return 0
}

func removeSliceItem[S ~[]E, E any](s *S, i int) {
	if i < 0 || s == nil || i >= len(*s) {
		return
	}
	var zero E
	copy((*s)[i:], (*s)[i+1:])
	(*s)[len(*s)-1] = zero
	*s = (*s)[:len(*s)-1]
}

// encoding utilities

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

// equalsFixedRune determines if a a rune is equal to another one previously
// encoded with encodeFixedRune.
func compareWhileEncoding(r rune, b0, b1, b2 byte) bool {
	// compare least significant bytes first
	return byte(r) == b0 &&
		byte(r>>8) == b1 &&
		byte(r>>16)&lsb5Mask == b2 // only use the 5 lsb
}

func compareWhileDecoding(r rune, b0, b1, b2 byte) bool {
	// compare least significant bytes first
	return r&maxUint8 == rune(b0) &&
		r&maxUint8<<8 == rune(b1) &&
		r&lsb5Mask<<16 == rune(b2) // only use the 5 lsb
}

// encode3MSB uses the 3 LSB of the given byte and returns a value with them as
// the 3 MSB.
func encode3MSB(b byte) byte {
	return (b & 7) << 5
}

// decode3MSB decodes a value encoded with encode3MSB.
func decode3MSB(b byte) byte {
	return b >> 5
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
