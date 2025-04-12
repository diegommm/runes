//go:build sizeof

package util

import (
	"reflect"
	"unicode"
	"unsafe"
)

// Sizeof returns an estimation of the size in bytes of the given `Set` and
// true. This is the counterpart of the function in a file with the opposite
// build tag.
func Sizeof(x interface{ Contains(rune) bool }) (uintptr, bool) {
	return x.(Sizerof).Sizeof(), true
}

// Sizerof must be satisfied by all `Set` types when using the build constraint
// in this file.
type Sizerof interface {
	Sizeof() uintptr
}

func SizeofSlice[S ~[]E, E any](s S) uintptr {
	size := unsafe.Sizeof(s)
	switch reflect.TypeOf(new(E)).Elem().Kind() {
	case reflect.Array, reflect.Chan, reflect.Func, reflect.Interface,
		reflect.Map, reflect.Pointer, reflect.Slice:
		size := uintptr(len(s)) * unsafe.Sizeof(*new(E))
		for _, elem := range s {
			so, ok := any(elem).(Sizerof)
			if ok {
				size += so.Sizeof()
			}
		}
	default:
		if _, ok := any(*new(E)).(Sizerof); ok {
			for _, elem := range s {
				size += any(elem).(Sizerof).Sizeof()
			}
		} else {
			size += uintptr(len(s)) * unsafe.Sizeof(*new(E))
		}
	}
	return size
}

func (f ContainsFunc) Sizeof() uintptr {
	return 0 // cannot estimate
}

// SizeofUnicodeRangeTable returns an estimation of the size in bytes of the
// given *unicode.RangeTable and true. This is the counterpart of the function
// in a file with the opposite build tag.
func SizeofUnicodeRangeTable(rt *unicode.RangeTable) (uintptr, bool) {
	return unsafe.Sizeof(*rt) +
			uintptr(len(rt.R16))*unsafe.Sizeof(unicode.Range16{}) +
			uintptr(len(rt.R32))*unsafe.Sizeof(unicode.Range32{}),
		true
}
