package runes

type integer interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr
}

func ceilDiv[T integer](dividend, divisor T) T {
	return (dividend + divisor - 1) / divisor
}
