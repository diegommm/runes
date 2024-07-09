package runes

import "testing"

const defaultBufferSize = 64

// errString is an `error` based on a string.
type errString struct {
	string
}

// Error implements the `error` interface.
func (e *errString) Error() string { return e.string }

// bufferWriter can be implemented to write parts of a string within a possibly
// longer string.
type bufferWriter interface {
	writeToBuffer(*buffer)
}

// bufferWriterFunc is an adapter for bufferWriter.
type bufferWriterFunc func(*buffer)

// writeToBuffer implements bufferWriter interface.
func (f bufferWriterFunc) writeToBuffer(b *buffer) { f(b) }

// toBufferWriter converts to a bufferWriter a variadic func, allowing it to be
// passed to a buffer's `write` method.
func toBufferWriter[T any](f func(*buffer, T) *buffer, v T) bufferWriterFunc {
	return func(b *buffer) {
		f(b, v)
	}
}

// toBufferWriterv is a variadic alternative of toBufferWriter for funcs that
// receive a slice type. This is only syntax sugar, since you can still pass a
// single slice value using toBufferWriter.
func toBufferWriterv[T any](f func(*buffer, []T) *buffer, vs ...T) bufferWriterFunc {
	return func(b *buffer) {
		f(b, vs)
	}
}

// bufferString returns a string from the given bufferWriter.
func bufferString(w bufferWriter) string {
	b := newBufferWithLen(0)
	b.write(w)
	return b.String()
}

// buffer is a buffer of bytes intended to create a string. It is not safe for
// concurrent use and its internals should not be directly accessed in other
// files.
type buffer struct {
	buf []byte
}

// newBufferLen returns a new buffer with an initial length of `n`.
func newBufferWithLen(n int) *buffer {
	return &buffer{
		buf: make([]byte, 0, n),
	}
}

// newBuffer returns a new buffer with an initial length of defaultBufferSize.
func newBuffer() *buffer {
	return newBufferWithLen(defaultBufferSize)
}

// Err calls String and returns an error with that description.
func (b *buffer) Err() error { return &errString{b.String()} }

// String converts the given buffer into a string and returns it.
func (b *buffer) String() string { return string(b.buf) }

func (b *buffer) raw(v []byte) *buffer    { return rawToBuffer(b, v) }
func (b *buffer) str(v string) *buffer    { return rawToBuffer(b, v) }
func (b *buffer) int(v int) *buffer       { return writeIntToBuffer(b, v) }
func (b *buffer) int32(v int32) *buffer   { return writeIntToBuffer(b, v) }
func (b *buffer) uint64(v uint64) *buffer { return writeIntToBuffer(b, v) }
func (b *buffer) int64(v int64) *buffer   { return writeIntToBuffer(b, v) }
func (b *buffer) rune(v rune) *buffer     { return writeIntToBuffer(b, v) }

// write passes the buffer to a given bufferWriter and then returns the same
// buffer. This methods helps chaining.
func (b *buffer) write(w bufferWriter) *buffer {
	w.writeToBuffer(b)
	return b
}

// byte writes a raw byte to a buffer.
func (b *buffer) byte(v byte) *buffer {
	b.buf = append(b.buf, v)
	return b
}

// intsToBuffer writes a list of raw bytes to a buffer. Contents are enclosed in
// square brackets and separated by a comma.
func rawsToBuffer[T interface{ string | []byte }](b *buffer, vs []T) *buffer {
	b.byte('[')
	for i, v := range vs {
		if i > 0 {
			b.byte(',')
		}
		rawToBuffer(b, v)
	}
	b.byte(']')
	return b
}

// intsToBuffer writes a list of integers to a buffer. Contents are enclosed in
// square brackets and separated by a comma.
func intsToBuffer[T xInt](b *buffer, vs []T) *buffer {
	b.byte('[')
	for i, v := range vs {
		if i > 0 {
			b.byte(',')
		}
		writeIntToBuffer(b, v)
	}
	b.byte(']')
	return b
}

// rawToBuffer writes raw bytes to a buffer.
func rawToBuffer[T interface{ string | []byte }](b *buffer, v T) *buffer {
	b.buf = append(b.buf, v...)
	return b
}

// writeIntToBuffer writes any integer type in base 10 to a buffer.
func writeIntToBuffer[T xInt](b *buffer, v T) *buffer {
	if v == 0 {
		b.byte('0')
		return b
	}

	u := uint64(v)
	if v < 0 {
		u = uint64(-v)
	}

	var buf [20]byte // max string len of any int including sign in base 10
	i := len(buf) - 1
	for ; i >= 0 && u > 0; i-- {
		buf[i] = '0' + byte(u%10)
		u /= 10
	}

	if v < 0 {
		buf[i] = '-'
		i--
	}

	b.raw(buf[i+1:])

	return b
}

func TestBuffer_String(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		expected, actual string
	}{
		{"asd", newBuffer().raw([]byte{'a', 's', 'd'}).Err().Error()},
		{"asd", newBuffer().str("asd").String()},
		{"a", newBuffer().byte('a').String()},
		{"asd", newBuffer().str("asd").String()},
		{"0", newBuffer().int32(0).String()},
		{"10", newBuffer().rune(10).String()},
		{"-10", newBuffer().int(-10).String()},
		{"18446744073709551615", newBuffer().uint64(1<<64 - 1).String()},
		{"9223372036854775807", newBuffer().int64(1<<63 - 1).String()},
		{"-9223372036854775808", newBuffer().int64(-(1 << 63)).String()},
		{"[1,2,3]", newBuffer().write(toBufferWriter(intsToBuffer,
			[]int{1, 2, 3})).String()},
		{`["x","y"]`, bufferString(toBufferWriterv(rawsToBuffer,
			`"x"`, `"y"`))},
	}

	for i, tc := range testCases {
		if tc.expected != tc.actual {
			t.Fatalf("[%d] expected: %q, actual: %q, bytes: %v", i, tc.expected,
				tc.actual, []byte(tc.actual))
		}
	}
}
