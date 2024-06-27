package runes

const (
	rune1Max = 1<<7 - 1  // 127
	rune2Max = 1<<11 - 1 // 2047
	rune3Max = 1<<16 - 1 // 65535

	tx    = 0b10000000 // 128
	t2    = 0b11000000 // 192
	t3    = 0b11100000 // 224
	t4    = 0b11110000 // 240
	maskx = 0b00111111 // 63

	surrogateMin        = 0xD800 // 55296
	surrogateMax        = 0xDFFF // 57343
	surrogateMaskUint32 = ^uint32(0x800 - 1)
	surrogateMaskRune   = ^rune(0x800 - 1)
)
