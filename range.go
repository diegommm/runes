package runes

import "github.com/diegommm/runes/iface"

const (
	maxUint8  = 1<<8 - 1
	maxUint16 = 1<<16 - 1
	maxInt32  = 1<<31 - 1

	lsb5Mask = 1<<5 - 1

	maxRuneListLinearSearch  = 10 // TODO: calibrate
	maxRangeListLinearSearch = 10 // TODO: calibrate
)

type (
	Set   = iface.Set
	Range = iface.Range
)

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

var nullbufferWriter = bufferWriterFunc(func(b *buffer) { b.str("null") })

// rangeGoString is used by internal helper to provide more details about an
// implementation's structure. It generally follows roughly a JSON format, but
// it doen't need to.
func rangeGoString[T bufferWriter](b *buffer, r Range, propertiesWriter T) {
	b.
		str(`{"type": "`).str(r.Type()).
		str(`", "len": `).int32(r.RuneLen()).
		str(`, "min": `).rune(r.Min()).
		str(`, "max": `).rune(r.Max()).
		str(`, "properties": `).write(propertiesWriter).
		str(`}`)
}

// SortRangeFunc is a function that can be used with the `slices` package to
// sort a slice of Ranges. To check if a set of Ranges overlap, first sort them
// using this function and then call [Overlap].
func SortRangeFunc[R Range](a, b R) int {
	am, bm := a.Min(), b.Min()
	switch {
	case am < bm:
		return -1
	case am > bm:
		return 1
	default:
		return 0
	}
}

// TODO: add "Sorted" and check for sorting in Overlap as well

// Overlap returns wheter the given ranges overlap, and the first pair of them
// that do, in sorted order. The ranges themselves are expected to be sorted in
// ascending order.
func Overlap[R Range](rs ...R) (i, j int, overlap bool) {
	oldMax := rune(-1)
	for i, r := range rs {
		if r.Min() <= oldMax {
			return i - 1, i, true
		}
		oldMax = r.Max()
	}

	return 0, 0, false
}

// EmptyRange is a [Range] that contains no runes.
var EmptyRange = emptyRange{}

type emptyRange struct{}

func (emptyRange) Contains(rune) bool { return false }
func (emptyRange) Type() string       { return "empty" }
func (emptyRange) Pos(rune) int32     { return -1 }
func (emptyRange) Nth(int32) rune     { return -1 }
func (emptyRange) RuneLen() int32     { return 0 }
func (emptyRange) Min() rune          { return -1 }
func (emptyRange) Max() rune          { return -1 }

// Valid types that can be used to encode a [Range] that contains a single rune,
// depending on how many bytes will be used to store it.
type (
	// OneValueRange1 is the most space efficient option, using a single byte,
	// and has no performance cost, but requires that the rune can be
	// represented with a single byte. This means that it has to be in the range
	// [0, 255] (which is where all Latin1 and ASCII runes are located). A value
	// of this type can be easily and safely converted to rune at no cost.
	OneValueRange1 = oneValueRange124[byte]
	// OneValueRange2 follows in space afficiency with 2 bytes for a rune and
	// no performance cost. The rune must be in the range [0, 65535]. A value of
	// this type can be easily and safely converted to rune at no cost.
	OneValueRange2 = oneValueRange124[uint16]
	// OneValueRange3 can represent any rune using 3 bytes. In comparison,
	// OneValueRange4 is ~35% faster, but OneValueRange3 saves 25% of space. The
	// difference, either in speed or in space, is negligible for most
	// applications, but can be critical in hot paths of parsers dealing with
	// high rune values. Note that the space efficiency can only be leveraged
	// when many values of the type are tightly packed, such as in correctly
	// typed arrays and slices. This benefit can be voided if used as a struct
	// element due to struct aligment, or completely unnoticed for low volumes
	// of elements of this type. It is also noteworthy that a value of this type
	// cannot be directly converted to a rune, but rather its accesor methods
	// should be used instead (like calling Min or Max) in order to decode the
	// value.
	OneValueRange3 = oneValueRange3
	// OneValueRange4 can represent any rune using 4 bytes. This is the most
	// conservative option and also performs better than OneValueRange3 (again,
	// in terms of fractions of nanoseconds). A value of this type can be easily
	// and safely converted to rune at no cost.
	OneValueRange4 = oneValueRange124[rune]

	OneValueRange interface {
		Range
		OneValueRange1 | OneValueRange2 | OneValueRange3 | OneValueRange4
	}
)

// NewOneValueRange returns a [OneValueRange] containing a single rune. See the
// documentation on [OneValueRange] and its options to learn how to choose the
// right type parameter.
func NewOneValueRange[R OneValueRange](r rune) R {
	var ret R
	switch ptr := any(&ret).(type) {
	case *OneValueRange1:
		*ptr = OneValueRange1(([1]byte{byte(r)}))
	case *OneValueRange2:
		*ptr = OneValueRange2([1]uint16{uint16(r)})
	case *OneValueRange3:
		var runeBytes [3]byte
		encodeFixedRune(&runeBytes, r)
		*ptr = OneValueRange3(runeBytes)
	case *OneValueRange4:
		*ptr = OneValueRange4([1]rune{r})
	default:
		panic("NewOneValueRange: unexpected type")
	}
	return ret
}

// NewDynamicOneValueRange is the same as [NewOneValueRange], but the returned
// type parameter is the smallest needed to represent the given rune.
func NewDynamicOneValueRange(r rune) Range {
	switch u := uint32(r); {
	case u <= maxUint8:
		return NewOneValueRange[OneValueRange1](r)
	case u <= maxUint16:
		return NewOneValueRange[OneValueRange2](r)
	default:
		return NewOneValueRange[OneValueRange3](r)
	}
}

func oneValuePos(r1, r2 rune) int32 {
	if r1 == r2 {
		return 0
	}
	return -1
}

type oneValueRange124[T interface{ byte | uint16 | rune }] [1]T

func (x oneValueRange124[T]) Type() string {
	switch any(x[0]).(type) {
	case byte:
		return "single rune stored in 1 byte"
	case uint16:
		return "single rune stored in 2 bytes"
	case rune:
		return "single rune stored in 4 bytes"
	default:
		panic("unknown type")
	}
}

func (x oneValueRange124[T]) Pos(r rune) int32 {
	return oneValuePos(r, rune(x[0]))
}

func (x oneValueRange124[T]) Nth(i int32) rune {
	if i == 0 {
		return rune(x[0])
	}
	return -1
}

func (x oneValueRange124[T]) Contains(r rune) bool    { return r == rune(x[0]) }
func (x oneValueRange124[T]) RuneLen() int32          { return 1 }
func (x oneValueRange124[T]) Min() rune               { return rune(x[0]) }
func (x oneValueRange124[T]) Max() rune               { return rune(x[0]) }
func (x oneValueRange124[T]) writeToBuffer(b *buffer) { b.rune(x.Min()) }
func (x oneValueRange124[T]) GoString() string        { return bufferString(x) }

type oneValueRange3 [3]byte

func (x oneValueRange3) Contains(r rune) bool {
	return rune(decodeFixedRune(x[0], x[1], x[2])) == r
}

func (x oneValueRange3) Type() string {
	return "single rune stored in 3 bytes"
}

func (x oneValueRange3) Min() rune {
	return rune(decodeFixedRune(x[0], x[1], x[2]))
}

func (x oneValueRange3) Nth(i int32) rune {
	if i == 0 {
		return x.Min()
	}
	return -1
}

func (x oneValueRange3) Pos(r rune) int32        { return oneValuePos(r, x.Min()) }
func (x oneValueRange3) RuneLen() int32          { return 1 }
func (x oneValueRange3) Max() rune               { return x.Min() }
func (x oneValueRange3) writeToBuffer(b *buffer) { b.rune(x.Min()) }
func (x oneValueRange3) GoString() string        { return bufferString(x) }

// NewSimpleRange returns an inclusive range of all the runes starting at `from`
// and ending at `to`.
func NewSimpleRange[R OneValueRange](from, to rune) (SimpleRange[R], error) {
	if from < 0 || from > to {
		return SimpleRange[R]{}, &errString{"invalid range"}
	}
	return SimpleRange[R]{
		NewOneValueRange[R](from),
		NewOneValueRange[R](to),
	}, nil
}

// NewDynamicSimpleRange is like [NewSimpleRange], but dynamically chooses the
// most storage-efficient alternative for the given values.
func NewDynamicSimpleRange(from, to rune) (Range, error) {
	switch u := uint32(to); {
	case u <= maxUint8:
		return NewSimpleRange[OneValueRange1](from, to)
	case u <= maxUint16:
		return NewSimpleRange[OneValueRange2](from, to)
	default:
		return NewSimpleRange[OneValueRange3](from, to)
	}
}

// SimpleRange is the inclusive range of runes contained in both of its ends.
type SimpleRange[R OneValueRange] [2]R

func (x SimpleRange[R]) Contains(r rune) bool {
	return r >= x[0].Min() && r <= x[1].Max()
}

func (x SimpleRange[R]) Type() string {
	return "simple range of from-to rune values"
}

func (x SimpleRange[R]) Pos(r rune) int32 {
	if m := x[0].Min(); r >= m && r <= x[1].Max() {
		return int32(r - m)
	}
	return -1
}

func (x SimpleRange[R]) Nth(i int32) rune {
	if r, m := rune(i), x[0].Min(); r >= 0 && r <= x[1].Max()-m {
		return r + m
	}
	return -1
}

func (x SimpleRange[R]) RuneLen() int32 { return int32(x[1].Max() + 1 - x[0].Min()) }
func (x SimpleRange[R]) Min() rune      { return x[0].Min() }
func (x SimpleRange[R]) Max() rune      { return x[1].Max() }

// Options for simple lists of runes to create a [Range].
type (
	// RuneListRangeLinear needs its items sorted in ascending order.
	RuneListRangeLinear[R OneValueRange] []R
	// RuneListRangeBinary needs its items sorted in ascending order.
	RuneListRangeBinary[R OneValueRange] []R

	// RuneListRange is a [Range] based on a list of runes that are sorted in
	// ascending order, and whenever a search for a specific rune is done then
	// the search algorithm used will be either linear or binary.
	RuneListRange[R OneValueRange] interface {
		Range
		RuneListRangeLinear[R] | RuneListRangeBinary[R]
	}
)

func rangeSliceNth[S ~[]R, R Range](x S, i int32) rune {
	if i >= 0 && int(i) < len(x) {
		return x[i].Min()
	}
	return -1
}

func rangeSliceMin[S ~[]R, R Range](x S) rune {
	if len(x) > 0 {
		return x[0].Min()
	}
	return 0
}

func rangeSliceMax[S ~[]R, R Range](x S) rune {
	if len(x) > 0 {
		return x[len(x)-1].Max()
	}
	return 0
}

// NewRuneListRange return a [RuneListRange] of the specified type from its
// arguments, which are expected to be sorted in ascending order.
func NewRuneListRange[L RuneListRange[R], R OneValueRange](i Iterator) L {
	r := make(L, 0, i.RuneLen())
	for {
		rr, ok := i.NextRune()
		if !ok {
			break
		}
		r = append(r, NewOneValueRange[R](rr))
	}
	return r
}

// NewRuneListRange returns the most storage-efficient possible representation
// needed to represent as a single [Range] all of its arguments, which are
// expected to be sorted in ascending order. The search method, either linear or
// binary, is also dynamically determined. The arguments are treated as
// scattered runes that cannot be better represented with any of the more
// sophisticated techniques that use either bitmaps, regularity of distribution,
// etc.
func NewDynamicRuneListRange(i Iterator) Range {
	l := i.RuneLen()
	switch l {
	case 0:
		return EmptyRange
	case 1:
		if r, ok := i.NextRune(); ok {
			return NewDynamicOneValueRange(r)
		}
		return EmptyRange
	}

	switch u := uint32(i.Max()); {
	case u <= maxUint8:
		if l > maxRuneListLinearSearch {
			return NewRuneListRange[RuneListRangeBinary[OneValueRange1]](i)
		}
		return NewRuneListRange[RuneListRangeLinear[OneValueRange1]](i)

	case u <= maxUint16:
		if l > maxRuneListLinearSearch {
			return NewRuneListRange[RuneListRangeBinary[OneValueRange2]](i)
		}
		return NewRuneListRange[RuneListRangeLinear[OneValueRange2]](i)

	default:
		if l > maxRuneListLinearSearch {
			return NewRuneListRange[RuneListRangeBinary[OneValueRange2]](i)
		}
		return NewRuneListRange[RuneListRangeLinear[OneValueRange2]](i)
	}
}

func (x RuneListRangeLinear[R]) Type() string {
	return "list of runes, using linear search"
}

func (x RuneListRangeLinear[R]) Pos(r rune) int32 {
	if len(x) == 0 || r < x[0].Min() || x[len(x)-1].Min() < r {
		return -1
	}
	return x.posSlow(r)
}

func (x RuneListRangeLinear[R]) posSlow(r rune) int32 {
	for i := range x {
		if r == x[i].Min() {
			return int32(i)
		}
	}
	return -1
}

func (x RuneListRangeLinear[R]) Contains(r rune) bool { return x.Pos(r) >= 0 }
func (x RuneListRangeLinear[R]) Nth(i int32) rune     { return rangeSliceNth(x, i) }
func (x RuneListRangeLinear[R]) RuneLen() int32       { return int32(len(x)) }
func (x RuneListRangeLinear[R]) Min() rune            { return rangeSliceMin(x) }
func (x RuneListRangeLinear[R]) Max() rune            { return rangeSliceMax(x) }

func (x RuneListRangeBinary[R]) Type() string {
	return "list of runes, using binary search"
}

func (x RuneListRangeBinary[R]) Pos(r rune) int32 {
	if len(x) == 0 || r < x[0].Min() || x[len(x)-1].Min() < r {
		return -1
	}
	return x.posSlow(r)
}

func (x RuneListRangeBinary[R]) posSlow(r rune) int32 {
	i, j := uint32(0), uint32(len(x)-1)
	for h := u32Half(i, j); i <= j && int(h) < len(x); h = u32Half(i, j) {
		switch v := x[h].Min(); {
		case r < v:
			j = h - 1
		case v < r:
			i = h + 1
		default:
			return int32(h)
		}
	}
	return -1
}

func (x RuneListRangeBinary[R]) Contains(r rune) bool { return x.Pos(r) >= 0 }
func (x RuneListRangeBinary[R]) Nth(i int32) rune     { return rangeSliceNth(x, i) }
func (x RuneListRangeBinary[R]) RuneLen() int32       { return int32(len(x)) }
func (x RuneListRangeBinary[R]) Min() rune            { return rangeSliceMin(x) }
func (x RuneListRangeBinary[R]) Max() rune            { return rangeSliceMax(x) }

// ExceptionRange returns a [Range] that is exactly the same as `m`, except for
// the items in `x`, whose items must be a subset of `m` and not match x.Min()
// nor x.Max().
func ExceptionRange[M, X Range](match M, except X) Range {
	return exceptionRange[M, X]{match, except}
}

type exceptionRange[M, X Range] struct {
	m M
	x X
}

func (x exceptionRange[M, X]) Contains(r rune) bool {
	return !x.x.Contains(r) && x.m.Contains(r)
}

func (x exceptionRange[M, X]) Type() string {
	return "range excepting internal subset, first checking exception"
}

func (x exceptionRange[M, X]) Pos(r rune) int32 {
	mPos := x.m.Pos(r)
	if mPos < 0 {
		return mPos
	}
	xStartPos := x.m.Nth(x.x.Min())
	if mPos < xStartPos {
		return mPos
	}
	xLen := x.x.RuneLen()
	if mPos < xStartPos+xLen {
		return -1
	}
	return mPos - xLen
}

func (x exceptionRange[M, X]) Nth(i int32) rune {
	mLen, xLen := x.m.RuneLen(), x.x.RuneLen()
	if i < 0 || i >= mLen-xLen {
		return -1
	}
	if xStartPos := x.m.Nth(x.x.Min()); i >= xStartPos {
		return x.m.Nth(i - xStartPos)
	}
	return x.m.Nth(i)
}

func (x exceptionRange[M, X]) RuneLen() int32 {
	return x.m.RuneLen() - x.x.RuneLen()
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

func (x exceptionRange[M, X]) Min() rune        { return x.m.Min() }
func (x exceptionRange[M, X]) Max() rune        { return x.m.Max() }
func (x exceptionRange[M, X]) GoString() string { return bufferString(x) }

// NewRangeList returns a new [Range] from the given list, which must be
// sorted in increasing order and non-overlapping. Use [Range] as type parameter
// for most flexibility, and use a non-interface type for most compactness.
func NewRangeList[R Range](rs ...R) (Range, error) {
	if i, j, overlap := Overlap(rs...); overlap {
		return nil, newBuffer().
			str("overlapping ranges [").int(i).str("]: ").
			write(bsRange[R](rs[i:j])).
			Err()
	}

	// we do not and we should not do any further level of optimization, like
	// checking whether we can convert to a one value Range. That should be done
	// at a different layer, we only care for optimizing the list here
	switch len(rs) {
	case 0:
		return EmptyRange, nil
	case 1:
		return rs[0], nil
	case 2:
		return twoRange[R]{rs[0], rs[1]}, nil
	default:
		return bsRange[R](rs), nil
	}
}

// twoRange saves 7 words and is faster than bsRange. Its elements must be
// sorted and non-overlapping.
type twoRange[R Range] [2]R

func (x twoRange[R]) Contains(r rune) bool {
	rx := x.rxPos(r)
	return rx < 2 && x[rx].Contains(r)
}

func (x twoRange[R]) Type() string {
	return "two ranges combined"
}

func (x twoRange[R]) Pos(r rune) int32 {
	// inlined method: Min and Max are generally very fast, so we can rapidly
	// discard most of the rune-space by checking boundaries
	if r < x[0].Min() || r > x[1].Max() {
		return -1
	}
	return x.posSlow(r)
}

func (x twoRange[R]) posSlow(r rune) int32 {
	pos := int32(-1)
	if rx := x.rxPos(r); rx < 2 {
		pos = x[rx].Pos(r)
		if pos >= 0 && rx == 1 {
			pos += x[0].RuneLen()
		}
	}
	return pos
}

func (x twoRange[R]) rxPos(r rune) uint32 {
	// inlined method: Min and Max are generally very fast, so we can rapidly
	// discard most of the rune-space by checking boundaries
	if r < x[0].Min() || r > x[1].Max() {
		return 2
	}
	if r > x[0].Max() {
		return 1
	}
	return 0
}

func (x twoRange[R]) Nth(i int32) rune {
	l0, l1 := x[0].RuneLen(), x[1].RuneLen()
	switch {
	case i < 0:
		return -1
	case i < l0:
		return x[0].Nth(i)
	case i < l0+l1:
		return x[1].Nth(i - l0)
	default:
		return -1
	}
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

func (x twoRange[R]) RuneLen() int32   { return x[0].RuneLen() + x[1].RuneLen() }
func (x twoRange[R]) Min() rune        { return x[0].Min() }
func (x twoRange[R]) Max() rune        { return x[1].Max() }
func (x twoRange[R]) GoString() string { return bufferString(x) }

// lsRange needs its elements to be sorted and non-overlapping.
type lsRange[R Range] []R

// bsRange needs its elements to be sorted and non-overlapping.
type bsRange[R Range] []R

func (x bsRange[R]) Contains(r rune) bool {
	rx := x.rxPos(r)
	return int(rx) < len(x) && x[rx].Contains(r)
}

func (x bsRange[R]) Type() string {
	return "several ranges combined, using binary search"
}

// Pos is not optimized. lsRange/bsRange are best effort for anything other than
// Contains.
func (x bsRange[R]) Pos(r rune) int32 {
	pos := int32(-1)
	if rx := x.rxPos(r); int(rx) < len(x) {
		pos = x[rx].Pos(r)
		if pos >= 0 {
			for i := uint32(0); i < rx; i++ {
				pos += x[i].RuneLen()
			}
		}
	}
	return pos
}

func (x bsRange[R]) rxPos(r rune) uint32 {
	// inlined method: Min and Max are generally very fast, so we can rapidly
	// discard most of the rune-space by checking boundaries
	if r < x[0].Min() || r > x[len(x)-1].Max() {
		return uint32(len(x))
	}
	return x.rxPosSlow(r)
}

func (x bsRange[R]) rxPosSlow(r rune) uint32 {
	i, j := uint32(0), uint32(len(x)-1)
	for h := u32Half(i, j); i < j && int(h) < len(x); h = u32Half(i, j) {
		xh := x[h]
		switch {
		case r < xh.Min():
			j = h - 1
		case xh.Max() < r:
			i = h + 1
		default:
			return h
		}
	}
	if i == j {
		return i // save one call to x[h].Min()
	}
	return uint32(len(x))
}

func (x bsRange[R]) Nth(i int32) rune {
	if i < 0 || len(x) < 1 {
		return -1
	}
	for _, rr := range x {
		l := rr.RuneLen()
		if i < l {
			return rr.Nth(i)
		}
		i -= l
	}
	return -1
}

func (x bsRange[R]) RuneLen() int32 {
	var total int32
	for _, rr := range x {
		total += rr.RuneLen()
	}
	return total
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
func (x bsRange[R]) Min() rune        { return rangeSliceMin(x) }
func (x bsRange[R]) Max() rune        { return rangeSliceMax(x) }

// NewUniformRange5 returs a [Range] that contains `runeCount` runes, starting
// with `minRune`. Consecutive runes in the returned Range are uniformly spaced
// every `stride` rune values, which must be in the range [1, 8]. One in
// `stride` means the range contains all runes starting at `minRune` until
// `minRune`+`runeCount`-1; two means every other rune starting at `minRune` for
// `runeCount`, etc.
// The range is highly coompressed and is fixed at 5 bytes with near-constant
// performance, regardless of the parameters.
func NewUniformRange5(minRune rune, runeCount uint16, stride byte) (uniformRange5, error) {
	if runeCount == 0 {
		return uniformRange5{}, &errString{"NewUniformRange5: runeCount cannot be zero"}
	}
	if stride == 0 {
		return uniformRange5{}, &errString{"NewUniformRange5: stride cannot be zero"}
	}
	if stride > 8 {
		return uniformRange5{}, &errString{"NewUniformRange5: value too long for stride"}
	}
	if runeCount == 1 {
		// stride > 1 here is acceptable but unnecessary. We set it to zero for
		// consistency
		stride = 1
	}
	var minRuneBytes [3]byte
	encodeFixedRune(&minRuneBytes, minRune)
	var runeCountBytes [2]byte
	encodeUint16(&runeCountBytes, runeCount)

	return uniformRange5{
		minRuneBytes[0],
		minRuneBytes[1],
		minRuneBytes[2] | encode3MSB(stride-1),
		runeCountBytes[0],
		runeCountBytes[1],
	}, nil
}

type uniformRange5 [5]byte

func (x uniformRange5) Contains(r rune) bool {
	u := uint32(r - x.Min())
	s := uint32(decode3MSB(x[2]) + 1)
	c := uint32(decodeUint16(x[3], x[4]))
	if s < 2 {
		return u < c
	}
	return u < s*c && u%s == 0
}

func (x uniformRange5) Type() string {
	return "range of uniformly distributed runes, stored in 5 bytes"
}

func (x uniformRange5) Pos(r rune) int32 {
	u := uint32(r - x.Min())
	s := uint32(decode3MSB(x[2]) + 1)
	if s > 0 && u%s == 0 && u < s*uint32(decodeUint16(x[3], x[4])) {
		return int32(u / s)
	}
	return -1
}

func (x uniformRange5) Nth(i int32) rune {
	count := decodeUint16(x[3], x[4])
	if i < 0 || i >= int32(count) {
		return -1
	}
	return x.Min() + rune(i)*rune(decode3MSB(x[2])+1)
}

func (x uniformRange5) Min() rune {
	return decodeFixedRune(x[0], x[1], x[2])
}

func (x uniformRange5) Max() rune {
	return decodeFixedRune(x[0], x[1], x[2]) + // Min
		rune(decodeUint16(x[3], x[4])-1)*rune(decode3MSB(x[2])+1)
}

func (x uniformRange5) RuneLen() int32 {
	return int32(decodeUint16(x[3], x[4]))
}

func (x uniformRange5) writeToBuffer(b *buffer) {
	rangeGoString(b, x, bufferWriterFunc(func(b *buffer) {
		b.str(`{"stride": `).byte(decode3MSB(x[2]) + 1).
			str(`, "raw_bytes": `).write(toBufferWriter(intsToBuffer, x[:])).
			str(`}`)
	}))
}

func (x uniformRange5) GoString() string { return bufferString(x) }

// NewUniformRange68 returs a [Range] that contains `runeCount` runes, starting
// with `minRune`. Consecutive runes in the returned Range are uniformly spaced
// every `stride` rune values, with a minimum of one. One in `stride` means the
// range contains all runes starting at `minRune` until `minRune`+`runeCount`-1;
// two means every other rune starting at `minRune` for `runeCount`, etc. The
// returned value is fixed at 6 or 8 bytes bytes depending if the type parameter
// is uint16 or rune, respectively. It has near-constant performance, regardless
// of the parameters, and performs around ~3 times better than NewUniformRange5.
func NewUniformRange68[T interface{ uint16 | rune }](minRune T, runeCount, stride uint16) (uniformRange68[T], error) {
	if runeCount == 0 {
		return uniformRange68[T]{},
			&errString{"NewUniformRange8: runeCount cannot be zero"}
	}
	if stride == 0 {
		return uniformRange68[T]{},
			&errString{"NewUniformRange8: stride cannot be zero"}
	}
	if runeCount == 1 {
		// stride > 1 here is acceptable but unnecessary. We set it to zero for
		// consistency
		stride = 1
	}

	return uniformRange68[T]{
		min:    minRune,
		count:  runeCount,
		stride: stride,
	}, nil
}

type uniformRange68[T interface{ uint16 | rune }] struct {
	min    T
	stride uint16
	count  uint16
}

func (x uniformRange68[T]) Contains(r rune) bool {
	u, s, c := uint32(r-rune(x.min)), uint32(x.stride), uint32(x.count)
	return s > 0 && // always true, but removes runtime.panicdivide
		u < s*c && u%s == 0
}

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

func (x uniformRange68[T]) Pos(r rune) int32 {
	u, s := uint32(r-rune(x.min)), uint32(x.stride)
	if s > 0 && // always true, but removes runtime.panicdivide
		u%s == 0 && u < s*uint32(x.count) {
		return int32(u / s)
	}
	return -1
}

func (x uniformRange68[T]) Nth(i int32) rune {
	if i < 0 || i >= int32(x.count) {
		return -1
	}
	return rune(x.min) + rune(i)*rune(x.stride)
}

func (x uniformRange68[T]) Max() rune {
	return rune(x.min) + rune(x.count-1)*rune(x.stride)
}

func (x uniformRange68[T]) writeToBuffer(b *buffer) {
	rangeGoString(b, x, bufferWriterFunc(func(b *buffer) {
		b.str(`{"stride": `).uint64(uint64(x.stride)).str(`}`)
	}))
}

func (x uniformRange68[T]) RuneLen() int32   { return int32(x.count) }
func (x uniformRange68[T]) Min() rune        { return rune(x.min) }
func (x uniformRange68[T]) GoString() string { return bufferString(x) }

// NewStringBitmapRange ...
func NewStringBitmapRange(i Iterator) (StringBitmap, error) {
	minRune, ok := i.NextRune() // get minRune
	if !ok {
		return "", nil
	}
	numEls := uint64(i.Max() + 1 - minRune)
	if numEls < 0 || numEls > maxUint16 {
		return "", newBuffer().uint64(numEls).
			str(" exceeds maximum number of elements: ").
			uint64(maxUint16).
			Err()
	}

	// allocate for the whole string
	bin := make([]byte, stringBitmapHeaderLen+ceilDiv(numEls, 8))

	// encode header
	encodeFixedRune((*[3]byte)(bin), minRune)
	encodeUint16((*[2]byte)(bin[3:]), uint16(numEls))

	// build the bitmap
	r := minRune
	bm := bin[stringBitmapHeaderLen:]
	for {
		u := uint32(r + 1 - minRune)
		bm[u>>3] |= 1 << (u & 7)

		if r, ok = i.NextRune(); !ok {
			break
		}
	}

	return StringBitmap(bin), nil
}

type StringBitmap string

const stringBitmapHeaderLen = 0 +
	3 + // min rune
	2 // number of elements

func (x StringBitmap) Contains(r rune) bool {
	if len(x) <= stringBitmapHeaderLen {
		return false
	}
	u := uint32(r + 1 - decodeFixedRune(x[0], x[1], x[2]))
	i := stringBitmapHeaderLen + int(u>>3)
	// why having runtime.panicIndex is faster here???
	return i < len(x) && 1<<(u&7)&x[i] != 0
}

func (x StringBitmap) Type() string {
	return "string bitmap"
}

func (x StringBitmap) Pos(r rune) int32 {
	if len(x) <= stringBitmapHeaderLen {
		return -1
	}
	u := uint32(r + 1 - decodeFixedRune(x[0], x[1], x[2]))
	i := stringBitmapHeaderLen + int(u>>3)
	if i < len(x) && 1<<(u&7)&x[i] != 0 {
		var pos int32
		bm := x[stringBitmapHeaderLen : i-1]
		for i := range bm {
			pos += int32(ones(bm[i]))
		}
		mask := byte(1<<(1+u&7) - 1)

		return pos + int32(ones(x[i]&mask))
	}
	return -1
}

func (x StringBitmap) Nth(i int32) rune {
	if i < 0 || i >= x.RuneLen() {
		return -1
	}
	r := decodeFixedRune(x[0], x[1], x[2])
	bm := x[stringBitmapHeaderLen:]
	for j := range bm {
		n := int32(ones(bm[j]))
		if n >= i {
			r += rune(j<<3) + rune(nthBitPos(bm[j], byte(i)))
			break
		}
		i -= n
	}
	return r
}

func (x StringBitmap) RuneLen() int32 {
	if len(x) <= stringBitmapHeaderLen {
		return 0
	}
	return int32(decodeUint16(x[3], x[4]))
}

func (x StringBitmap) Min() rune {
	if len(x) <= stringBitmapHeaderLen {
		return -1
	}
	return decodeFixedRune(x[0], x[1], x[2])
}

func (x StringBitmap) Max() rune {
	if len(x) <= stringBitmapHeaderLen {
		return -1
	}
	return decodeFixedRune(x[0], x[1], x[2]) +
		rune((len(x)-stringBitmapHeaderLen)<<3) +
		rune(msbPos(x[len(x)-1]))
}
