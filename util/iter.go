package util

import (
	"cmp"
	"iter"
)

// OrderedListIter is the generic form of runes.OrderedRunesIter.
type OrderedListIter[T cmp.Ordered] = func() (T, bool)

// OrderedList is the generic form of runes.OrderedRunesList.
type OrderedList[T cmp.Ordered] interface {
	Min() T
	Max() T
	Len() int
	Iter() OrderedListIter[T]
}

// SliceList is a [OrderedList] from a slice of items sorted in ascending order.
type SliceList[T cmp.Ordered] []T

func (x SliceList[T]) Min() (_ T) {
	if len(x) == 0 {
		return
	}
	return x[0]
}

func (x SliceList[T]) Max() (_ T) {
	if len(x) == 0 {
		return
	}
	return x[len(x)-1]
}

func (x SliceList[T]) Len() int {
	return len(x)
}

func (x SliceList[T]) Iter() OrderedListIter[T] {
	var i int
	return func() (_ T, _ bool) {
		if ret := i; ret < len(x) {
			i++
			return x[ret], true
		}
		return
	}
}

type Integer interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr
}

// Seq is an OrderedList of `Count` elements starting at `First`, `Stride` apart
// from each other.
type Seq[T Integer] struct {
	First  T
	Count  int // must be positive
	Stride T   // must be positive
}

func (s Seq[T]) Min() T {
	return s.First
}

func (s Seq[T]) Max() T {
	return s.First + T(s.Count-1)*s.Stride
}

func (s Seq[T]) Len() int {
	return s.Count
}

func (s Seq[T]) Iter() OrderedListIter[T] {
	i := s.Count
	v := s.First
	return func() (_ T, _ bool) {
		if i < 1 {
			return
		}
		i--
		ret := v
		v += s.Stride
		return ret, true
	}
}

// IterMergeFunc returns combines the items produced by the iterators
// successively returned by `f`. When `f` returns nil, then X is exhausted.
func IterMergeFunc[T cmp.Ordered](f func() OrderedListIter[T]) OrderedListIter[T] {
	if f == nil {
		return (SliceList[T])(nil).Iter()
	}
	it := f()
	return func() (_ T, _ bool) {
		for {
			if it == nil {
				return
			}
			if r, ok := it(); ok {
				return r, true
			}
			it = f()
		}
	}
}

func IterMerge[T cmp.Ordered](ss ...OrderedListIter[T]) OrderedListIter[T] {
	i := 0
	return IterMergeFunc[T](func() OrderedListIter[T] {
		if i >= len(ss) {
			return nil
		}
		i++
		return ss[i-1]
	})
}

// IterExcept returns a new iterator with all the values of `s` that are not in
// `x`.
func IterExcept[T cmp.Ordered](s, x OrderedListIter[T]) OrderedListIter[T] {
	if s == nil {
		return (SliceList[T])(nil).Iter()
	}
	m := map[T]struct{}{}
	for v := range IterSeq(x) {
		m[v] = struct{}{}
	}
	return func() (_ T, _ bool) {
		for v, ok := s(); ok; v, ok = s() {
			if _, ok2 := m[v]; !ok2 {
				return v, true
			}
		}
		return
	}
}

// Collect returns a slice of all the items of `s`.
func Collect[T cmp.Ordered](s OrderedListIter[T]) []T {
	var ret []T
	for v := range IterSeq(s) {
		ret = append(ret, v)
	}
	return ret
}

// IterSeq converts `s` to a standard library iterator.
func IterSeq[T cmp.Ordered](s OrderedListIter[T]) iter.Seq[T] {
	if s == nil {
		return func(func(T) bool) {}
	}
	return func(yield func(T) bool) {
		for v, ok := s(); ok; v, ok = s() {
			yield(v)
		}
	}
}
