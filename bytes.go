package runes

import "unsafe"

// EncodeRune writes into p (which must be large enough) the UTF-8 encoding of
// the rune. If the rune is out of range, it writes the encoding of
// [utf8.RuneError]. It returns the number of bytes written.
func EncodeRune(p []byte, r rune) int {
	switch u := uint32(r); {
	case u <= rune1Max:
		p[0] = byte(u)
		return 1
	case u <= rune2Max:
		_ = p[1] // eliminate bounds checks
		p[0] = t2 | byte(u>>6)
		p[1] = tx | byte(u)&maskx
		return 2
	case u > maxRune, u&surrogateMaskUint32 == surrogateMin: // utf8.RuneError
		_ = p[2] // eliminate bounds checks
		p[0] = 239
		p[1] = 191
		p[2] = 189
		return 3
	case u <= rune3Max:
		_ = p[2] // eliminate bounds checks
		p[0] = t3 | byte(u>>12)
		p[1] = tx | byte(u>>6)&maskx
		p[2] = tx | byte(u)&maskx
		return 3
	default:
		_ = p[3] // eliminate bounds checks
		p[0] = t4 | byte(u>>18)
		p[1] = tx | byte(u>>12)&maskx
		p[2] = tx | byte(u>>6)&maskx
		p[3] = tx | byte(u)&maskx
		return 4
	}
}

// EncodeRune4 writes into p (which must be have a minimum length of 4) the
// UTF-8 encoding of the rune. If the rune is out of range, it writes the
// encoding of [utf8.RuneError]. It returns the number of bytes written.
//
// NOTE: This is faster and more efficient than [EncodeRune], but it requires a
// minimum length of 4, regardless if they will be used.
func EncodeRune4(p []byte, r rune) int {
	_ = p[3]
	return UTF8Bytes((*[4]byte)(unsafe.Pointer(unsafe.SliceData(p))), r)
}

// UTF8Bytes writes into p the UTF-8 encoding of the rune. If the rune is out of
// range, it writes the encoding of [utf8.RuneError]. It returns the number of
// bytes written.
func UTF8Bytes(p *[4]byte, r rune) int {
	switch u := uint32(r); {
	case u <= rune1Max:
		p[0] = byte(r)
		return 1
	case u <= rune2Max:
		p[0] = t2 | byte(u>>6)
		p[1] = tx | byte(u)&maskx
		return 2
	case u&surrogateMaskUint32 == surrogateMin: // utf8.RuneError
		p[0] = 239
		p[1] = 191
		p[2] = 189
		return 3
	case u <= rune3Max:
		p[0] = t3 | byte(u>>12)
		p[1] = tx | byte(u>>6)&maskx
		p[2] = tx | byte(u)&maskx
		return 3
	case u <= maxRune:
		p[0] = t4 | byte(u>>18)
		p[1] = tx | byte(u>>12)&maskx
		p[2] = tx | byte(u>>6)&maskx
		p[3] = tx | byte(u)&maskx
		return 4
	}
	// utf8.RuneError
	p[0] = 239
	p[1] = 191
	p[2] = 189
	return 3
}

// AppendRune appends the UTF-8 encoding of r to the end of p and returns the
// extended buffer. If the rune is out of range, it appends the encoding of
// [utf8.RuneError].
func AppendRune(p []byte, r rune) []byte {
	u := uint32(r)
	if u <= rune1Max {
		return append(p, byte(u))
	}
	return appendRuneNonASCII(p, u)
}

func appendRuneNonASCII(p []byte, u uint32) []byte {
	switch {
	case u <= rune2Max:
		return append(p,
			t2|byte(u>>6),
			tx|byte(u)&maskx,
		)
	case u > maxRune, surrogateMin <= u && u <= surrogateMax:
		return append(p, 239, 191, 189) // utf8.RuneError
	case u <= rune3Max:
		return append(p,
			t3|byte(u>>12),
			tx|byte(u>>6)&maskx,
			tx|byte(u)&maskx,
		)
	default:
		return append(p,
			t4|byte(u>>18),
			tx|byte(u>>12)&maskx,
			tx|byte(u>>6)&maskx,
			tx|byte(u)&maskx,
		)
	}
}
