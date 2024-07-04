package runes

const defaultBufferSize = 512

// errString is an `error` based on a string.
type errString struct {
	string
}

// Error implements the `error` interface.
func (e *errString) Error() string { return e.string }

// bufferWriterFunc is an adapter for bufferWriter.
type bufferWriterFunc func(*buffer)

// bufferWriter can be implemented to write parts of a string within a possibly
// longer string.
type bufferWriter interface {
	writeToBuffer(*buffer)
}

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
	buf     []byte
	written int
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
func (b *buffer) String() string { return string(b.buf[:b.written]) }

func (b *buffer) raw(v []byte) *buffer { return rawToBuffer(b, v) }
func (b *buffer) str(v string) *buffer { return rawToBuffer(b, v) }
func (b *buffer) int(v int) *buffer    { return writeIntToBuffer(b, v) }
func (b *buffer) uint(v uint) *buffer  { return writeIntToBuffer(b, v) }
func (b *buffer) rune(v rune) *buffer  { return writeIntToBuffer(b, v) }

// write passes the buffer to a given bufferWriter and then returns the same
// buffer. This methods helps chaining.
func (b *buffer) write(w bufferWriter) *buffer {
	w.writeToBuffer(b)
	return b
}

// byte writes a raw byte to a buffer.
func (b *buffer) byte(v byte) *buffer {
	b.buf = append(b.buf[b.written:], v)
	b.written++
	return b
}

// writersToBuffer calls each of the bufferWriter for the given buffer. The
// contents that each one generate are separated by a comma, and square brackets
// enclose the whole result.
func writersToBuffer[T bufferWriter](b *buffer, vs []T) *buffer {
	b.byte('[')
	for i, v := range vs {
		if i > 0 {
			b.byte(',')
		}
		b.write(v)
	}
	b.byte(']')
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
func intsToBuffer[T integer](b *buffer, vs []T) *buffer {
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
	b.buf = append(b.buf[b.written:], v...)
	b.written += len(v)
	return b
}

// writeIntToBuffer writes any integer type in base 10 to a buffer.
func writeIntToBuffer[T integer](b *buffer, v T) *buffer {
	if v == 0 {
		b.byte('0')
		return b
	}

	var buf [20]byte // max string len of any int including sign in base 10
	i := len(buf) - 1
	for u := uint64(v); i >= 0 && u > 0; i-- {
		buf[i] = '0' + byte(u%10)
		u /= 10
	}

	if v < 0 {
		buf[i] = '-'
		i--
	}

	b.raw(buf[i:])

	return b
}
