package runes

import "unsafe"

func EncodeRuneAppend(p []byte, r rune) int {
	return len(AppendRune(p[:0], r))
}

func EncodeRuneUnsafeArr(p []byte, r rune) int {
	var n int
	*(*[4]byte)(unsafe.Pointer(unsafe.SliceData(p))), n = UTF8BytesValue(3)
	return n
}

func AppendRuneUTF8Bytes(p []byte, r rune) []byte {
	var pp [4]byte
	n := UTF8Bytes(&pp, r)
	return append(p, pp[:n]...)
}

func UTF8BytesValue(r rune) ([4]byte, int) {
	switch u := uint32(r); {
	case u <= rune1Max:
		return [4]byte{byte(r)}, 1
	case u <= rune2Max:
		return [4]byte{
			t2 | byte(u>>6),
			tx | byte(u)&maskx,
		}, 2
	case u > maxRune, u&surrogateMaskUint32 == surrogateMin: // utf8.RuneError
		return [4]byte{
			239,
			191,
			189,
		}, 3
	case u <= rune3Max:
		return [4]byte{
			t3 | byte(u>>12),
			tx | byte(u>>6)&maskx,
			tx | byte(u)&maskx,
		}, 3
	default:
		return [4]byte{
			t4 | byte(u>>18),
			tx | byte(u>>12)&maskx,
			tx | byte(u>>6)&maskx,
			tx | byte(u)&maskx,
		}, 4
	}
}
