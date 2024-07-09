package runes

import "fmt"

var nullbufferWriter = bufferWriterFunc(func(b *buffer) { b.str("null") })

func writerOrGoStringToBuffer[R Range](b *buffer, r R) *buffer {
	if bw, ok := any(r).(bufferWriter); ok {
		b.write(bw)
	} else if gs, ok := any(r).(interface{ GoString() string }); ok {
		b.str(gs.GoString())
	} else {
		rangeGoString(b, r, nullbufferWriter)
	}
	return b
}

// rangeGoString is used by internal helper to provide more details about an
// implementation's structure. It generally follows roughly a JSON format, but
// it doen't need to.
func rangeGoString[T bufferWriter](b *buffer, r Range, propertiesWriter T) {
	var rangeType string
	if typ, ok := r.(interface{ Type() string }); ok {
		rangeType = typ.Type()
	} else {
		rangeType = fmt.Sprintf("%T", r)
	}

	b.
		str(`{"type": "`).str(rangeType).
		str(`", "len": `).int32(r.RuneLen()).
		str(`, "min": `).rune(r.Min()).
		str(`, "max": `).rune(r.Max()).
		str(`, "properties": `).write(propertiesWriter).
		str(`}`)
}

func (emptyRange) Type() string { return "empty" }
func (x emptyRange) writeToBuffer(b *buffer) {
	rangeGoString(b, x, nullbufferWriter)
}
func (x emptyRange) GoString() string { return bufferString(x) }

func (x oneValueRange124[T]) Type() string {
	switch any(x[0]).(type) {
	case byte:
		return "single rune stored in 1 byte"
	case uint16:
		return "single rune stored in 2 bytes"
	}
	return "single rune stored in 4 bytes"
}
func (x oneValueRange124[T]) writeToBuffer(b *buffer) { b.rune(x.Min()) }
func (x oneValueRange124[T]) GoString() string        { return bufferString(x) }

func (x oneValueRange3) Type() string {
	return "single rune stored in 3 bytes"
}
func (x oneValueRange3) writeToBuffer(b *buffer) { b.rune(x.Min()) }
func (x oneValueRange3) GoString() string        { return bufferString(x) }

func (x SimpleRange[R]) Type() string {
	return "simple range of from-to rune values"
}
func (x SimpleRange[R]) writeToBuffer(b *buffer) {
	rangeGoString(b, x, nullbufferWriter)
}
func (x SimpleRange[R]) GoString() string { return bufferString(x) }

func (x RuneListRangeLinear[R]) Type() string {
	return "list of runes, using linear search"
}
func (x RuneListRangeLinear[R]) writeToBuffer(b *buffer) {
	b.str(`{"runes:"[`)
	for i := range x {
		if i > 0 {
			b.str(`,`)
		}
		b.int(int(x[i].Min()))
	}
	b.str(`]}`)
}
func (x RuneListRangeLinear[R]) GoString() string { return bufferString(x) }

func (x RuneListRangeBinary[R]) Type() string {
	return "list of runes, using binary search"
}
func (x RuneListRangeBinary[R]) writeToBuffer(b *buffer) {
	b.str(`{"runes:"[`)
	for i := range x {
		if i > 0 {
			b.str(`,`)
		}
		b.int(int(x[i].Min()))
	}
	b.str(`]}`)
}
func (x RuneListRangeBinary[R]) GoString() string { return bufferString(x) }

func (x exceptionRange[M, X]) Type() string {
	return "range excepting interior subset"
}
func (x exceptionRange[M, X]) writeToBuffer(b *buffer) {
	rangeGoString(b, x, bufferWriterFunc(func(b *buffer) {
		b.str(`{"matching": `).
			write(toBufferWriter(writerOrGoStringToBuffer, x.m)).
			str(`,"excepting":`).
			write(toBufferWriter(writerOrGoStringToBuffer, x.x)).
			str(`}`)
	}))
}
func (x exceptionRange[M, X]) GoString() string { return bufferString(x) }

func (x twoRange[R]) Type() string {
	return "two ranges combined"
}
func (x twoRange[R]) writeToBuffer(b *buffer) {
	rangeGoString(b, x, bufferWriterFunc(func(b *buffer) {
		b.str(`{"len": 2, "items": [`).
			write(toBufferWriter(writerOrGoStringToBuffer, x[0])).
			str(`,`).
			write(toBufferWriter(writerOrGoStringToBuffer, x[1])).
			str(`]}`)
	}))
}
func (x twoRange[R]) GoString() string { return bufferString(x) }

func (x bsRange[R]) Type() string {
	return "several ranges combined, using binary search"
}
func (x bsRange[R]) writeToBuffer(b *buffer) {
	rangeGoString(b, x, bufferWriterFunc(func(b *buffer) {
		b.str(`{"len": `).int(len(x)).str(`, "items": [`)
		for i, r := range x {
			if i > 0 {
				b.byte(',')
			}
			b.write(toBufferWriter(writerOrGoStringToBuffer, r))
		}
		b.str(`]}`)
	}))
}
func (x bsRange[R]) GoString() string { return bufferString(x) }

func (x uniformRange5) Type() string {
	return "range of uniformly distributed runes, stored in 5 bytes"
}
func (x uniformRange5) writeToBuffer(b *buffer) {
	rangeGoString(b, x, bufferWriterFunc(func(b *buffer) {
		b.str(`{"stride": `).byte(decode3MSB(x[2]) + 2).
			str(`, "raw_bytes": `).write(toBufferWriter(intsToBuffer, x[:])).
			str(`}`)
	}))
}
func (x uniformRange5) GoString() string { return bufferString(x) }

func (x uniformRange68[T]) Type() string {
	switch any(x.min).(type) {
	case uint16:
		return "range of uniformly distributed runes, stored in 6 bytes"
	case rune:
		return "range of uniformly distributed runes, stored in 8 bytes"
	default:
		panic("unknown type")
	}
}
func (x uniformRange68[T]) writeToBuffer(b *buffer) {
	rangeGoString(b, x, bufferWriterFunc(func(b *buffer) {
		b.str(`{"stride": `).uint64(uint64(x.stride)).str(`}`)
	}))
}
func (x uniformRange68[T]) GoString() string { return bufferString(x) }

func (x stringBitmap) Type() string {
	return "string bitmap"
}
func (x stringBitmap) writeToBuffer(b *buffer) {
	bmLen := max(0, len(x)-stringBitmapHeaderLen)
	rangeGoString(b, x, bufferWriterFunc(func(b *buffer) {
		b.str(`{"header_len":`).int(stringBitmapHeaderLen).
			str(`, "bitmap_len":`).int(bmLen).
			str(`}`)
	}))
}
func (x stringBitmap) GoString() string { return bufferString(x) }
