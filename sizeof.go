//go:build sizeof

package runes

import (
	"unsafe"

	"github.com/diegommm/runes/util"
)

func (x Union[T]) Sizeof() uintptr {
	return util.SizeofSlice(x)
}

func (x LinearSlice[T]) Sizeof() uintptr {
	return util.SizeofSlice(x)
}

func (x BinarySlice[T]) Sizeof() uintptr {
	return util.SizeofSlice(x)
}

func (x Interval[T]) Sizeof() uintptr {
	return unsafe.Sizeof(x)
}

func (x Uniform[T]) Sizeof() uintptr {
	return unsafe.Sizeof(x)
}

func (x Bitmap) Sizeof() uintptr {
	return unsafe.Sizeof(x) + uintptr(len(x))
}
