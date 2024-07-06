package runes

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

// msbPos returns the position of the most-signigicant bit that is set in a
// byte.
func msbPos(b byte) byte {
	var r byte
	if b > 15 {
		r += 4
		b >>= 4
	}
	if b > 3 {
		r += 2
		b >>= 2
	}
	return r + b>>1
}

// ones returns the number of bits set in a byte.
func ones(b byte) byte {
	return 0 +
		((b >> 7) & 1) + ((b >> 6) & 1) + ((b >> 5) & 1) + ((b >> 4) & 1) +
		((b >> 3) & 1) + ((b >> 2) & 1) + ((b >> 1) & 1) + (b & 1)
}

// nthBitPos returns the position of the n-th bit set in the given byte, or zero
// if the bit is not found.
func nthBitPos(b, n byte) byte {
	for i := byte(0); i < 8; i++ {
		if b&(1<<i) != 0 {
			n--
			if n == 0 {
				return i
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
