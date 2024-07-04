package runes

const (
	maxUint8  = 1<<8 - 1
	maxUint16 = 1<<16 - 1

	lsb5Mask = 1<<5 - 1

	maxRuneListLinearSearch  = 10 // TODO: calibrate
	maxRangeListLinearSearch = 10 // TODO: calibrate
)

// Set is a set of runes.
type Set interface {
	// Contains returns whether the given rune is part of the set.
	Contains(rune) bool
}

// Range represents an ordered list of runes in the range `[Min(), Max()]`.
type Range interface {
	Set
	// Type is a human-readable range type.
	Type() string
	// Pos returns the position of the given rune within the range, or -1 if
	// it's not found.
	Pos(rune) int
	// Nth returns the N-th ordinal rune in the range and true, if there is a
	// rune at that position. Otherwise, it returns zero and false.
	Nth(int) (rune, bool)
	// RuneLen returns the number of runes. This is a method optimized to be
	// very fast and O(1).
	RuneLen() int
	// Min returns the numerically smallest rune. This is a method optimized to
	// be very fast and O(1).
	Min() rune
	// Max returns the numerically biggest rune. This is a method optimized to
	// be very fast and O(1).
	Max() rune
}

// Iterator returns a list of non-repeated runes in sorted ascending order.
type Iterator interface {
	// NextRune returns the next rune and true, or zero and false if there is no
	// next rune.
	NextRune() (rune, bool)
}

// LenIterator is an alternative interface that can be implemented by an
// [Iterator] to inform of the number of runes that it will yield.
type LenIterator interface {
	Iterator
	// RuneLen returns the number of runes that are still to be returned.
	RuneLen() int
}

// MaxIterator is an alternative interface that can be implemented by an
// [Iterator] to inform the max rune it will yield.
type MaxIterator interface {
	Iterator
	Max() rune
}

// RuneReader is the standard library's [io.RuneReader] interface.
type RuneReader interface {
	ReadRune() (r rune, size int, err error)
}

// RuneReaderIterator returns an [Iterator] from the given [RuneReader] which
// stops at the first error returned by the [ReadRune] method. The runes read
// must be sorted in ascending order and non-repeating.
func RuneReaderIterator(rr RuneReader) Iterator {
	return runeReaderIterator{rr}
}

type runeReaderIterator struct {
	RuneReader
}

func (x runeReaderIterator) NextRune() (rune, bool) {
	if x.RuneReader == nil {
		return 0, false
	}
	r, _, err := x.RuneReader.ReadRune()
	if err != nil {
		x.RuneReader = nil
		return 0, false
	}
	return r, true
}

// RangeIterator returns an [Iterator] from the given [Range].
func RangeIterator(r Range) Iterator {
	return &rangeIterator{r: r}
}

type rangeIterator struct {
	r   Range
	pos int
}

func (x *rangeIterator) NextRune() (rune, bool) {
	if r, ok := x.r.Nth(x.pos); ok {
		x.pos++
		return r, true
	}
	return 0, false
}

func (x *rangeIterator) RuneLen() int { return x.r.RuneLen() - x.pos }
func (x *rangeIterator) Max() rune    { return x.r.Max() }

// RunesIterator returns an iterator from the given slice of runes, which must
// be in sorted ascending order and non-repeating.
func RunesIterator(rs []rune) Iterator {
	return &runesIterator{rs: rs}
}

type runesIterator struct {
	rs  []rune
	pos uint
}

func (x *runesIterator) NextRune() (rune, bool) {
	if int(x.pos) < len(x.rs) {
		return x.rs[x.pos], true
	}
	return 0, false
}

func (x *runesIterator) RuneLen() int { return len(x.rs) }
func (x *runesIterator) Max() rune {
	if len(x.rs) > 0 {
		return x.rs[len(x.rs)-1]
	}
	return 0
}

// RangeGoString pretty prints a Range for debugging.
func RangeGoString(r Range) string {
	if gs, _ := r.(interface{ GoString() string }); gs != nil {
		return gs.GoString()
	}
	return rangeGoString(newBuffer(), r, nullbufferWriter).String()
}

func writerOrGoStringToBuffer[T Range](b *buffer, v T) *buffer {
	if bw, ok := any(v).(bufferWriter); ok {
		b.write(bw)
	} else {
		b.str(RangeGoString(v))
	}
	return b
}

// rangeGoString is used by internal helper to provide more details about an
// implementation's structure. It generally follows roughly a JSON format, but
// it doen't need to.
func rangeGoString[T bufferWriter](b *buffer, r Range, propertiesWriter T) *buffer {
	return b.
		str(`{"type": "`).str(r.Type()).
		str(`", "len": `).int(r.RuneLen()).
		str(`, "min": `).rune(r.Min()).
		str(`, "max": `).rune(r.Max()).
		str(`, "properties": `).write(propertiesWriter).
		str(`}`)
}

var nullbufferWriter = bufferWriterFunc(func(b *buffer) { b.str("null") })

// SortRangeFunc is a function that can be used with the `slices` package to
// sort Ranges. To check if a set of Ranges overlap, first sort them using this
// function and then call [Overlap].
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

func (emptyRange) Contains(rune) bool   { return false }
func (emptyRange) Type() string         { return "empty" }
func (emptyRange) Pos(rune) int         { return -1 }
func (emptyRange) Nth(int) (rune, bool) { return 0, false }
func (emptyRange) RuneLen() int         { return 0 }
func (emptyRange) Min() rune            { return 0 }
func (emptyRange) Max() rune            { return 0 }

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

func oneValuePos(r1, r2 rune) int {
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

func (x oneValueRange124[T]) Pos(r rune) int {
	return oneValuePos(r, rune(x[0]))
}

func (x oneValueRange124[T]) Nth(i int) (rune, bool) {
	if i == 0 {
		return rune(x[0]), true
	}
	return 0, false
}

func (x oneValueRange124[T]) Contains(r rune) bool    { return r == rune(x[0]) }
func (x oneValueRange124[T]) RuneLen() int            { return 1 }
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

func (x oneValueRange3) Nth(i int) (rune, bool) {
	if i == 0 {
		return x.Min(), true
	}
	return 0, false
}

func (x oneValueRange3) Pos(r rune) int          { return oneValuePos(r, x.Min()) }
func (x oneValueRange3) RuneLen() int            { return 1 }
func (x oneValueRange3) Max() rune               { return x.Min() }
func (x oneValueRange3) writeToBuffer(b *buffer) { b.rune(x.Min()) }
func (x oneValueRange3) GoString() string        { return bufferString(x) }

func rangeSliceNth[S ~[]R, R Range](x S, i int) (rune, bool) {
	if i >= 0 && i < len(x) {
		return x[i].Min(), true
	}
	return 0, false
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

// NewSimpleRange returns an inclusive range of all the runes starting at `from`
// and ending at `to`.
func NewSimpleRange[R OneValueRange](from, to rune) SimpleRange[R] {
	return SimpleRange[R]{
		NewOneValueRange[R](from),
		NewOneValueRange[R](to),
	}
}

// NewDynamicSimpleRange is like [NewSimpleRange], but dynamically chooses the
// most storage-efficient alternative for the given values.
func NewDynamicSimpleRange(from, to rune) Range {
	switch u := uint32(to); {
	case u <= maxUint8:
		return NewSimpleRange[OneValueRange1](from, to)
	case u <= maxUint16:
		return NewSimpleRange[OneValueRange2](from, to)
	default:
		return NewSimpleRange[OneValueRange2](from, to)
	}
}

type SimpleRange[R OneValueRange] [2]R

func (x SimpleRange[R]) Contains(r rune) bool {
	return r >= x[0].Min() && r <= x[1].Max()
}

func (x SimpleRange[R]) Type() string {
	return "simple range of from-to rune values"
}

func (x SimpleRange[R]) Pos(r rune) int {
	if m := x[0].Min(); r >= m && r <= x[1].Max() {
		return int(r - m)
	}
	return -1
}

func (x SimpleRange[R]) Nth(i int) (rune, bool) {
	if r, m := rune(i), x[0].Min(); r >= 0 && r < x[1].Max()-m {
		return r + m, true
	}
	return -1, false
}

func (x SimpleRange[R]) RuneLen() int { return int(x[1].Max() - x[0].Min()) }
func (x SimpleRange[R]) Min() rune    { return x[0].Min() }
func (x SimpleRange[R]) Max() rune    { return x[1].Max() }

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

// NewRuneListRange return a [RuneListRange] of the specified type from its
// arguments, which are expected to be sorted in ascending order.
func NewRuneListRange[L RuneListRange[R], R OneValueRange](i Iterator) L {
	var l int
	if li, ok := i.(LenIterator); ok {
		l = li.RuneLen()
	}
	r := make(L, 0, l)
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
	l := maxRuneListLinearSearch + 1
	if il, ok := i.(LenIterator); ok {
		l = il.RuneLen()
		switch l {
		case 0:
			return EmptyRange
		case 1:
			if r, ok := i.NextRune(); ok {
				return NewDynamicOneValueRange(r)
			}
			return EmptyRange
		}
	}

	maxRune := rune(-1) // default to max length when converted to uint32
	if mi, ok := i.(MaxIterator); ok {
		maxRune = mi.Max()
	}

	switch u := uint32(maxRune); {
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

func (x RuneListRangeLinear[R]) Pos(r rune) int {
	if len(x) == 0 || r < x[0].Min() || x[len(x)-1].Min() < r {
		return -1
	}
	return x.posSlow(r)
}

func (x RuneListRangeLinear[R]) posSlow(r rune) int {
	for i := range x {
		if r == x[i].Min() {
			return i
		}
	}
	return -1
}

func (x RuneListRangeLinear[R]) Contains(r rune) bool   { return x.Pos(r) >= 0 }
func (x RuneListRangeLinear[R]) Nth(i int) (rune, bool) { return rangeSliceNth(x, i) }
func (x RuneListRangeLinear[R]) RuneLen() int           { return len(x) }
func (x RuneListRangeLinear[R]) Min() rune              { return rangeSliceMin(x) }
func (x RuneListRangeLinear[R]) Max() rune              { return rangeSliceMax(x) }

func (x RuneListRangeBinary[R]) Type() string {
	return "list of runes, using binary search"
}

func (x RuneListRangeBinary[R]) Pos(r rune) int {
	if len(x) == 0 || r < x[0].Min() || x[len(x)-1].Min() < r {
		return -1
	}
	return x.posSlow(r)
}

func (x RuneListRangeBinary[R]) posSlow(r rune) int {
	i, j := uint(0), uint(len(x)-1)
	for h := (i + j) >> 1; i <= j && int(h) < len(x); h = (i + j) >> 1 {
		switch v := x[h].Min(); {
		case r < v:
			j = h - 1
		case v < r:
			i = h + 1
		default:
			return int(h)
		}
	}
	return -1
}

func (x RuneListRangeBinary[R]) Contains(r rune) bool   { return x.Pos(r) >= 0 }
func (x RuneListRangeBinary[R]) Nth(i int) (rune, bool) { return rangeSliceNth(x, i) }
func (x RuneListRangeBinary[R]) RuneLen() int           { return len(x) }
func (x RuneListRangeBinary[R]) Min() rune              { return rangeSliceMin(x) }
func (x RuneListRangeBinary[R]) Max() rune              { return rangeSliceMax(x) }

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

func (x twoRange[R]) Pos(r rune) int {
	// inlined method: Min and Max are generally very fast, so we can rapidly
	// discard most of the rune-space by checking boundaries
	if r < x[0].Min() || r > x[1].Max() {
		return -1
	}
	return x.posSlow(r)
}

func (x twoRange[R]) posSlow(r rune) int {
	pos := -1
	if rx := x.rxPos(r); rx < 2 {
		pos = x[rx].Pos(r)
		if pos >= 0 && rx == 1 {
			pos += x[0].RuneLen()
		}
	}
	return pos
}

func (x twoRange[R]) rxPos(r rune) uint {
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

func (x twoRange[R]) Nth(i int) (rune, bool) {
	l0, l1 := x[0].RuneLen(), x[1].RuneLen()
	switch {
	case i < 0:
		return 0, false
	case i < l0:
		return x[0].Nth(i)
	case i < l0+l1:
		return x[1].Nth(i - l0)
	default:
		return 0, false
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

func (x twoRange[R]) RuneLen() int     { return x[0].RuneLen() + x[1].RuneLen() }
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
func (x bsRange[R]) Pos(r rune) int {
	pos := -1
	if rx := x.rxPos(r); int(rx) < len(x) {
		pos = x[rx].Pos(r)
		if pos >= 0 {
			for i := uint(0); i < rx; i++ {
				pos += x[i].RuneLen()
			}
		}
	}
	return pos
}

func (x bsRange[R]) rxPos(r rune) uint {
	// inlined method: Min and Max are generally very fast, so we can rapidly
	// discard most of the rune-space by checking boundaries
	if r < x[0].Min() || r > x[len(x)-1].Max() {
		return uint(len(x))
	}
	return x.rxPosSlow(r)
}

func (x bsRange[R]) rxPosSlow(r rune) uint {
	i, j := uint(0), uint(len(x)-1)
	for h := (i + j) >> 1; i < j && int(h) < len(x); h = (i + j) >> 1 {
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
	return uint(len(x))
}

// Nth is not optimized. lsRange/bsRange are best effort for anything other than
// Contains.
func (x bsRange[R]) Nth(i int) (rune, bool) {
	if i < 0 || len(x) < 1 {
		return 0, false
	}
	for _, rr := range x {
		l := rr.RuneLen()
		if i < l {
			return rr.Nth(i)
		}
		i -= l
	}
	return 0, false
}

func (x bsRange[R]) RuneLen() int {
	var total int
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
	runeCountBytes := encodeUint16(runeCount)

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

func (x uniformRange5) Pos(r rune) int {
	u := uint32(r - x.Min())
	s := uint32(decode3MSB(x[2]) + 1)
	if s > 0 && u%s == 0 && u < s*uint32(decodeUint16(x[3], x[4])) {
		return int(u / s)
	}
	return -1
}

func (x uniformRange5) Nth(i int) (rune, bool) {
	count := decodeUint16(x[3], x[4])
	if i < 0 || i >= int(count) {
		return 0, false
	}
	return x.Min() + rune(i)*rune(decode3MSB(x[2])+1), true
}

func (x uniformRange5) Min() rune {
	return decodeFixedRune(x[0], x[1], x[2])
}

func (x uniformRange5) Max() rune {
	return decodeFixedRune(x[0], x[1], x[2]) + // Min
		rune(decodeUint16(x[3], x[4])-1)*rune(decode3MSB(x[2])+1)
}

func (x uniformRange5) RuneLen() int {
	return int(decodeUint16(x[3], x[4]))
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
		return uniformRange68[T]{}, &errString{"NewUniformRange8: runeCount cannot be zero"}
	}
	if stride == 0 {
		return uniformRange68[T]{}, &errString{"NewUniformRange8: stride cannot be zero"}
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

func (x uniformRange68[T]) Pos(r rune) int {
	u, s := uint32(r-rune(x.min)), uint32(x.stride)
	if s > 0 && // always true, but removes runtime.panicdivide
		u%s == 0 && u < s*uint32(x.count) {
		return int(u / s)
	}
	return -1
}

func (x uniformRange68[T]) Nth(i int) (rune, bool) {
	if i < 0 || i >= int(x.count) {
		return 0, false
	}
	return rune(x.min) + rune(i)*rune(x.stride), true
}

func (x uniformRange68[T]) Max() rune {
	return rune(x.min) + rune(x.count-1)*rune(x.stride)
}

func (x uniformRange68[T]) writeToBuffer(b *buffer) {
	rangeGoString(b, x, bufferWriterFunc(func(b *buffer) {
		b.str(`{"stride": `).uint(uint(x.stride)).str(`}`)
	}))
}

func (x uniformRange68[T]) RuneLen() int     { return int(x.count) }
func (x uniformRange68[T]) Min() rune        { return rune(x.min) }
func (x uniformRange68[T]) GoString() string { return bufferString(x) }

// encoding utilities

// encodeFixedRune encodes a rune in 3 bytes with little-endian. The rune should
// be no longer than 21 bits (only an invalid rune would be), and the 3 msb of
// the last byte are unused.
func encodeFixedRune(b *[3]byte, r rune) {
	b[0] = byte(r)
	b[1] = byte(r >> 8)
	b[2] = byte(r>>16) & lsb5Mask // only use the 5 lsb
}

// decodeFixedRune decodes a rune encoded with encodeFixedRune.
func decodeFixedRune(b0, b1, b2 byte) rune {
	return rune(b0) |
		rune(b1)<<8 |
		rune(b2&lsb5Mask)<<16 // discard the 3 msb
}

// equalsFixedRune determines if a a rune is equal to another one previously
// encoded with encodeFixedRune.
func compareWhileEncoding(r rune, b0, b1, b2 byte) bool {
	// compare least significant bytes first
	return byte(r) == b0 &&
		byte(r>>8) == b1 &&
		byte(r>>16)&lsb5Mask == b2 // only use the 5 lsb
}

func compareWhileDecoding(r rune, b0, b1, b2 byte) bool {
	// compare least significant bytes first
	return r&maxUint8 == rune(b0) &&
		r&maxUint8<<8 == rune(b1) &&
		r&lsb5Mask<<16 == rune(b2) // only use the 5 lsb
}

// encode3MSB uses the 3 LSB of the given byte and returns a value with them as
// the 3 MSB.
func encode3MSB(b byte) byte {
	return (b >> 5) << 5
}

// decode3MSB decodes a value encoded with encode3MSB.
func decode3MSB(b byte) byte {
	return b >> 5
}

// encodeUint16 encodes a uint16 in 2 bytes with little-endian.
func encodeUint16(c uint16) []byte {
	return []byte{
		byte(c),
		byte(c >> 8),
	}
}

// decodeUint16 decodes a uint16 encoded with encodeUint16.
func decodeUint16(b0, b1 byte) uint16 {
	return uint16(b0) |
		uint16(b1)<<8
}
