package runes

import (
	"fmt"
	"slices"
	"unicode"
	"unsafe"
)

// RangeTableToRangeList converts a [unicode.RangeTable] to a [Range] using
// [NewRangeList] with item(s) of type returned by [NewUniformRange5].
func RangeTableToRangeList(rt *unicode.RangeTable) (Range, error) {
	rs := make([]uniformRange5, 0, len(rt.R16)+len(rt.R32))
	for _, t := range rt.R16 {
		minRune := rune(t.Lo)
		runeCount := uint16(Range16RuneCount(t))
		if t.Stride < 8 {
			r, err := NewUniformRange5(minRune, runeCount, byte(t.Stride))
			if err != nil {
				return nil, fmt.Errorf("single convert Range16 %#v: %w", t, err)
			}
			rs = append(rs, r)
		} else {
			for i := uint16(0); i < runeCount; i++ {
				r, err := NewUniformRange5(minRune, 1, 1)
				if err != nil {
					return nil, fmt.Errorf("multi convert Range16 [%d] %#v: %w",
						i, t, err)
				}
				rs = append(rs, r)
				minRune += rune(t.Stride)
			}
		}
	}

	for _, t := range rt.R32 {
		minRune := rune(t.Lo)
		runeCountInt := Range32RuneCount(t)
		// doesn't happen in practice that the rune count is >= MaxUint16
		for runeCountInt > 0 {
			runeCount := uint16(min(runeCountInt, maxUint16))
			runeCountInt -= int(runeCount)
			if t.Stride < 8 {
				r, err := NewUniformRange5(minRune, runeCount, byte(t.Stride))
				if err != nil {
					return nil, fmt.Errorf("single convert Range32 %#v: %w", t,
						err)
				}
				rs = append(rs, r)
			} else {
				for i := uint16(0); i < runeCount; i++ {
					r, err := NewUniformRange5(minRune, 1, 1)
					if err != nil {
						return nil, fmt.Errorf("multi convert Range32 [%d]"+
							" %#v: %w", i, t, err)
					}
					rs = append(rs, r)
					minRune += rune(t.Stride)
				}
			}
		}
	}

	return NewRangeList(rs...)
}

var (
	rangeTableSize = int(unsafe.Sizeof(unicode.RangeTable{}))
	range16Size    = int(unsafe.Sizeof(unicode.Range16{}))
	range32Size    = int(unsafe.Sizeof(unicode.Range32{}))
)

type RangeTableStats struct {
	Name                   string
	NumRanges              int
	MaxStride              int
	MaxRangeLen            int
	NumRangesWithStrideOne int
	ByteSize               int
	EstimatedNewByteSize   int
	Strategy               string
	Min, Max, Span         rune
	Count                  int
	Capacity               int
	Density                float32
	Distribution           []*RuneDistribution
}

type RuneDistribution struct {
	EstimatedNewByteSize     int
	Strategy                 string
	DiffLast, Min, Max, Span rune
	Count                    int
	Capacity                 int
	Density                  float32
}

func estimateNewByteSize(span rune, count int) (byteSize int, capacity int, name string) {
	switch {
	case int(span+1) == count && count <= maxUint16:
		return 5, 65535, "runeRange"

	case span < 32:
		return 4 + // min rune
				4, // bit mask
			32,
			"mask32" // = 8

	case span < 96:
		return 4 + // min rune
				12, // [3]uint32 bit mask
			96,
			"mask32_96" // = 16

	case span < 224:
		return 4 + // min rune
				28, // [7]uint32 bit mask
			224,
			"mask32_224" // = 32

	default:
		b := ceilDiv(int(span+1), 8)
		return 16 + // value of type string
				3 + // first three bytes of min rune fixed encoding
				b, // 1 byte for every 8 values
			b * 8,
			"maskString"
	}
}

func GetRangeTableStats(name string, r *unicode.RangeTable) *RangeTableStats {
	ret := &RangeTableStats{
		Name:      name,
		NumRanges: len(r.R16) + len(r.R32),
		ByteSize:  RangeTableByteSize(r),
	}

	for _, t := range r.R16 {
		s := int(t.Stride)
		if s == 1 {
			ret.NumRangesWithStrideOne++
		}
		if s > ret.MaxStride {
			ret.MaxStride = s
		}
		if l := int(t.Hi - t.Lo); ret.MaxRangeLen < l {
			ret.MaxRangeLen = l
		}
	}
	for _, t := range r.R32 {
		s := int(t.Stride)
		if s == 1 {
			ret.NumRangesWithStrideOne++
		}
		if s > ret.MaxStride {
			ret.MaxStride = s
		}
		if l := int(t.Hi - t.Lo); ret.MaxRangeLen < l {
			ret.MaxRangeLen = l
		}
	}

	r3 := NewRangeTableRuneIterator(r)
	ret.Min, ret.Span, ret.Count = startSpanCount(r3)
	ret.Max = ret.Min + ret.Span
	ret.Density = float32(ret.Count) / (float32(ret.Max - ret.Min + 1))
	ret.EstimatedNewByteSize, ret.Capacity, ret.Strategy = estimateNewByteSize(ret.Span, ret.Count)

	if ret.Count > 1 {
		r3.Restart()
		runes := make([]rune, 0, ret.Count)
		for {
			r, _, err := r3.ReadRune()
			if err != nil {
				break
			}
			runes = append(runes, r)
		}
		slices.Sort(runes)

		if dists, size, capacity := makeDistribution(runes, float32(1)/1.6); len(dists) > 1 {
			ret.Distribution = dists
			ret.Capacity = capacity
			ret.EstimatedNewByteSize = 24 + // size of []interface{...}
				size*16 + // size of each interface{...} in 64 bit arch
				size
			ret.Strategy = "ifaceSlice"
		}
	}

	return ret
}

func makeDistribution(runes []rune, minDensity float32) (dists []*RuneDistribution, size, capacity int) {
	dists = []*RuneDistribution{
		{
			Min:     runes[0],
			Max:     runes[0],
			Span:    0,
			Count:   1,
			Density: 1,
		},
	}
	dists[0].EstimatedNewByteSize, dists[0].Capacity, dists[0].Strategy = estimateNewByteSize(dists[0].Span, dists[0].Count)

	for i := 1; i < len(runes); i++ {
		diff := runes[i] - runes[i-1]
		curDist := dists[len(dists)-1]
		newDensity := float32(curDist.Count+1) / (float32(runes[i] - curDist.Min + 1))

		if newDensity < minDensity {
			newDist := &RuneDistribution{
				DiffLast: diff,
				Min:      runes[i],
				Max:      runes[i],
				Span:     0,
				Count:    1,
				Density:  1,
			}
			newDist.EstimatedNewByteSize, newDist.Capacity, newDist.Strategy = estimateNewByteSize(newDist.Span, newDist.Count)
			dists = append(dists, newDist)
		} else {
			curDist.Max = runes[i]
			curDist.Span = curDist.Max - curDist.Min
			curDist.Count++
			curDist.Density = newDensity
			curDist.EstimatedNewByteSize, curDist.Capacity, curDist.Strategy = estimateNewByteSize(curDist.Span, curDist.Count)
		}
	}

	const runeSliceByteSizePerValue = 4 // value of type rune
	runeSliceDist := &RuneDistribution{
		EstimatedNewByteSize: 24, // value of type []rune
		Strategy:             "runeSlice",
	}

	for i := len(dists) - 1; i >= 0; i-- {
		dist := dists[i]
		if runeSliceByteSizePerValue*dist.Count < dist.EstimatedNewByteSize {
			// remove the current dist
			copy(dists[i:], dists[i+1:])
			dists[len(dists)-1] = nil
			dists = dists[:len(dists)-1]

			// add it to the runeSliceDist
			if runeSliceDist.Count == 0 {
				runeSliceDist.Min = dist.Min
				runeSliceDist.Max = dist.Max
			} else {
				if runeSliceDist.Min > dist.Min {
					runeSliceDist.Min = dist.Min
				}
				if runeSliceDist.Max < dist.Max {
					runeSliceDist.Max = dist.Max
				}
			}
			runeSliceDist.Count += dist.Count
			runeSliceDist.EstimatedNewByteSize += runeSliceByteSizePerValue * dist.Count
		}
	}

	if runeSliceDist.Count > 0 {
		dists = append(dists, runeSliceDist)
		runeSliceDist.Capacity = runeSliceDist.Count
		runeSliceDist.Span = runeSliceDist.Max - runeSliceDist.Min
		runeSliceDist.Density = float32(runeSliceDist.Count+1) / (float32(runeSliceDist.Span + 1))
	}

	for _, dist := range dists {
		size += dist.EstimatedNewByteSize
		capacity += dist.Capacity
	}

	return dists, size, capacity
}

// RangeTableByteSize returns the runtime size in bytes needed used by the given
// [unicode.RangeTable].
func RangeTableByteSize(r *unicode.RangeTable) int {
	return rangeTableSize + len(r.R16)*range16Size + len(r.R32)*range32Size
}

// Range16RuneCount returns the number of runes that the given [unicode.Range16]
// represents.
func Range16RuneCount(r unicode.Range16) int {
	return int(ceilDiv(r.Hi-r.Lo+1, r.Stride))
}

// Range32RuneCount returns the number of runes that the given [unicode.Range32]
// represents.
func Range32RuneCount(r unicode.Range32) int {
	return int(ceilDiv(r.Hi-r.Lo+1, r.Stride))
}

// NewRangeTableRuneIterator returns a [RuneIterator] from the given list of
// [unicode.RangeTable].
func NewRangeTableRuneIterator(rts ...*unicode.RangeTable) RuneIterator {
	var l int
	for _, rt := range rts {
		l += len(rt.R16) + len(rt.R32)
	}
	ris := make([]rangeTableSubIterator, 0, l)
	for _, rt := range rts {
		for _, t := range rt.R16 {
			ris = append(ris, range16Iterator(t))
		}
		for _, t := range rt.R32 {
			ris = append(ris, range32Iterator(t))
		}
	}

	return &rangeTablesIterator{
		ris: ris,
	}
}

type rangeTablesIterator struct {
	ris []rangeTableSubIterator
	pos int
}

func (i *rangeTablesIterator) ReadRune() (r rune, size int, err error) {
	for {
		if i.pos >= len(i.ris) {
			return readRuneEOF()
		}
		r, size, err = i.ris[i.pos].ReadRune()
		if err == nil {
			return r, size, err
		}
		i.ris[i.pos].Restart()
		i.pos++
	}
}

func (i *rangeTablesIterator) Restart() {
	if i.pos < len(i.ris) {
		i.ris[i.pos].Restart()
	}
	i.pos = 0
}

type rangeTableSubIterator struct {
	rune, min, max, stride rune
}

func range16Iterator(t unicode.Range16) rangeTableSubIterator {
	return rangeTableSubIterator{
		rune:   rune(t.Lo),
		min:    rune(t.Lo),
		max:    rune(t.Hi),
		stride: rune(t.Stride),
	}
}

func range32Iterator(t unicode.Range32) rangeTableSubIterator {
	return rangeTableSubIterator{
		rune:   rune(t.Lo),
		min:    rune(t.Lo),
		max:    rune(t.Hi),
		stride: rune(t.Stride),
	}
}

func (r *rangeTableSubIterator) ReadRune() (rune, int, error) {
	if r.rune > r.max {
		return readRuneEOF()
	}
	rr := r.rune
	r.rune += r.stride
	return readRuneReturn(rr)
}

func (r *rangeTableSubIterator) Restart() {
	r.rune = r.min
}

func NewRange16[T, S, U integer](min T, count S, stride U) unicode.Range16 {
	return unicode.Range16{
		Lo:     uint16(min),
		Hi:     uint16(min) + uint16(count)*uint16(stride),
		Stride: uint16(stride),
	}
}

// Range16Contains mimics the internal behaviour of the unicode package.
func Range16Contains(r16 unicode.Range16) func(rune) bool {
	return func(rr rune) bool {
		r := uint16(rr)
		range_ := &r16
		if r < range_.Lo {
			return false
		}
		if r <= range_.Hi {
			return range_.Stride == 1 || (r-range_.Lo)%range_.Stride == 0
		}
		return false
	}
}

func NewRange32[T, S, U integer](min T, count S, stride U) unicode.Range32 {
	return unicode.Range32{
		Lo:     uint32(min),
		Hi:     uint32(min) + uint32(count)*uint32(stride),
		Stride: uint32(stride),
	}
}

// Range32Contains mimics the internal behaviour of the unicode package.
func Range32Contains(r32 unicode.Range32) func(rune) bool {
	return func(rr rune) bool {
		r := uint32(rr)
		range_ := &r32
		if r < range_.Lo {
			return false
		}
		if r <= range_.Hi {
			return range_.Stride == 1 || (r-range_.Lo)%range_.Stride == 0
		}
		return false
	}
}

var RangeTables = map[string]*unicode.RangeTable{
	"ASCII_Hex_Digit":                    unicode.ASCII_Hex_Digit,
	"Bidi_Control":                       unicode.Bidi_Control,
	"Dash":                               unicode.Dash,
	"Deprecated":                         unicode.Deprecated,
	"Diacritic":                          unicode.Diacritic,
	"Extender":                           unicode.Extender,
	"Hex_Digit":                          unicode.Hex_Digit,
	"Hyphen":                             unicode.Hyphen,
	"IDS_Binary_Operator":                unicode.IDS_Binary_Operator,
	"IDS_Trinary_Operator":               unicode.IDS_Trinary_Operator,
	"Ideographic":                        unicode.Ideographic,
	"Join_Control":                       unicode.Join_Control,
	"Logical_Order_Exception":            unicode.Logical_Order_Exception,
	"Noncharacter_Code_Point":            unicode.Noncharacter_Code_Point,
	"Other_Alphabetic":                   unicode.Other_Alphabetic,
	"Other_Default_Ignorable_Code_Point": unicode.Other_Default_Ignorable_Code_Point,
	"Other_Grapheme_Extend":              unicode.Other_Grapheme_Extend,
	"Other_ID_Continue":                  unicode.Other_ID_Continue,
	"Other_ID_Start":                     unicode.Other_ID_Start,
	"Other_Lowercase":                    unicode.Other_Lowercase,
	"Other_Math":                         unicode.Other_Math,
	"Other_Uppercase":                    unicode.Other_Uppercase,
	"Pattern_Syntax":                     unicode.Pattern_Syntax,
	"Pattern_White_Space":                unicode.Pattern_White_Space,
	"Prepended_Concatenation_Mark":       unicode.Prepended_Concatenation_Mark,
	"Quotation_Mark":                     unicode.Quotation_Mark,
	"Radical":                            unicode.Radical,
	"Regional_Indicator":                 unicode.Regional_Indicator,
	"Sentence_Terminal":                  unicode.Sentence_Terminal,
	"Soft_Dotted":                        unicode.Soft_Dotted,
	"Terminal_Punctuation":               unicode.Terminal_Punctuation,
	"Unified_Ideograph":                  unicode.Unified_Ideograph,
	"Variation_Selector":                 unicode.Variation_Selector,
	"White_Space":                        unicode.White_Space,
	"Adlam":                              unicode.Adlam,
	"Ahom":                               unicode.Ahom,
	"Anatolian_Hieroglyphs":              unicode.Anatolian_Hieroglyphs,
	"Arabic":                             unicode.Arabic,
	"Armenian":                           unicode.Armenian,
	"Avestan":                            unicode.Avestan,
	"Balinese":                           unicode.Balinese,
	"Bamum":                              unicode.Bamum,
	"Bassa_Vah":                          unicode.Bassa_Vah,
	"Batak":                              unicode.Batak,
	"Bengali":                            unicode.Bengali,
	"Bhaiksuki":                          unicode.Bhaiksuki,
	"Bopomofo":                           unicode.Bopomofo,
	"Brahmi":                             unicode.Brahmi,
	"Braille":                            unicode.Braille,
	"Buginese":                           unicode.Buginese,
	"Buhid":                              unicode.Buhid,
	"Canadian_Aboriginal":                unicode.Canadian_Aboriginal,
	"Carian":                             unicode.Carian,
	"Caucasian_Albanian":                 unicode.Caucasian_Albanian,
	"Chakma":                             unicode.Chakma,
	"Cham":                               unicode.Cham,
	"Cherokee":                           unicode.Cherokee,
	"Chorasmian":                         unicode.Chorasmian,
	"Common":                             unicode.Common,
	"Coptic":                             unicode.Coptic,
	"Cuneiform":                          unicode.Cuneiform,
	"Cypriot":                            unicode.Cypriot,
	"Cypro_Minoan":                       unicode.Cypro_Minoan,
	"Cyrillic":                           unicode.Cyrillic,
	"Deseret":                            unicode.Deseret,
	"Devanagari":                         unicode.Devanagari,
	"Dives_Akuru":                        unicode.Dives_Akuru,
	"Dogra":                              unicode.Dogra,
	"Duployan":                           unicode.Duployan,
	"Egyptian_Hieroglyphs":               unicode.Egyptian_Hieroglyphs,
	"Elbasan":                            unicode.Elbasan,
	"Elymaic":                            unicode.Elymaic,
	"Ethiopic":                           unicode.Ethiopic,
	"Georgian":                           unicode.Georgian,
	"Glagolitic":                         unicode.Glagolitic,
	"Gothic":                             unicode.Gothic,
	"Grantha":                            unicode.Grantha,
	"Greek":                              unicode.Greek,
	"Gujarati":                           unicode.Gujarati,
	"Gunjala_Gondi":                      unicode.Gunjala_Gondi,
	"Gurmukhi":                           unicode.Gurmukhi,
	"Han":                                unicode.Han,
	"Hangul":                             unicode.Hangul,
	"Hanifi_Rohingya":                    unicode.Hanifi_Rohingya,
	"Hanunoo":                            unicode.Hanunoo,
	"Hatran":                             unicode.Hatran,
	"Hebrew":                             unicode.Hebrew,
	"Hiragana":                           unicode.Hiragana,
	"Imperial_Aramaic":                   unicode.Imperial_Aramaic,
	"Inherited":                          unicode.Inherited,
	"Inscriptional_Pahlavi":              unicode.Inscriptional_Pahlavi,
	"Inscriptional_Parthian":             unicode.Inscriptional_Parthian,
	"Javanese":                           unicode.Javanese,
	"Kaithi":                             unicode.Kaithi,
	"Kannada":                            unicode.Kannada,
	"Katakana":                           unicode.Katakana,
	"Kawi":                               unicode.Kawi,
	"Kayah_Li":                           unicode.Kayah_Li,
	"Kharoshthi":                         unicode.Kharoshthi,
	"Khitan_Small_Script":                unicode.Khitan_Small_Script,
	"Khmer":                              unicode.Khmer,
	"Khojki":                             unicode.Khojki,
	"Khudawadi":                          unicode.Khudawadi,
	"Lao":                                unicode.Lao,
	"Latin":                              unicode.Latin,
	"Lepcha":                             unicode.Lepcha,
	"Limbu":                              unicode.Limbu,
	"Linear_A":                           unicode.Linear_A,
	"Linear_B":                           unicode.Linear_B,
	"Lisu":                               unicode.Lisu,
	"Lycian":                             unicode.Lycian,
	"Lydian":                             unicode.Lydian,
	"Mahajani":                           unicode.Mahajani,
	"Makasar":                            unicode.Makasar,
	"Malayalam":                          unicode.Malayalam,
	"Mandaic":                            unicode.Mandaic,
	"Manichaean":                         unicode.Manichaean,
	"Marchen":                            unicode.Marchen,
	"Masaram_Gondi":                      unicode.Masaram_Gondi,
	"Medefaidrin":                        unicode.Medefaidrin,
	"Meetei_Mayek":                       unicode.Meetei_Mayek,
	"Mende_Kikakui":                      unicode.Mende_Kikakui,
	"Meroitic_Cursive":                   unicode.Meroitic_Cursive,
	"Meroitic_Hieroglyphs":               unicode.Meroitic_Hieroglyphs,
	"Miao":                               unicode.Miao,
	"Modi":                               unicode.Modi,
	"Mongolian":                          unicode.Mongolian,
	"Mro":                                unicode.Mro,
	"Multani":                            unicode.Multani,
	"Myanmar":                            unicode.Myanmar,
	"Nabataean":                          unicode.Nabataean,
	"Nag_Mundari":                        unicode.Nag_Mundari,
	"Nandinagari":                        unicode.Nandinagari,
	"New_Tai_Lue":                        unicode.New_Tai_Lue,
	"Newa":                               unicode.Newa,
	"Nko":                                unicode.Nko,
	"Nushu":                              unicode.Nushu,
	"Nyiakeng_Puachue_Hmong":             unicode.Nyiakeng_Puachue_Hmong,
	"Ogham":                              unicode.Ogham,
	"Ol_Chiki":                           unicode.Ol_Chiki,
	"Old_Hungarian":                      unicode.Old_Hungarian,
	"Old_Italic":                         unicode.Old_Italic,
	"Old_North_Arabian":                  unicode.Old_North_Arabian,
	"Old_Permic":                         unicode.Old_Permic,
	"Old_Persian":                        unicode.Old_Persian,
	"Old_Sogdian":                        unicode.Old_Sogdian,
	"Old_South_Arabian":                  unicode.Old_South_Arabian,
	"Old_Turkic":                         unicode.Old_Turkic,
	"Old_Uyghur":                         unicode.Old_Uyghur,
	"Oriya":                              unicode.Oriya,
	"Osage":                              unicode.Osage,
	"Osmanya":                            unicode.Osmanya,
	"Pahawh_Hmong":                       unicode.Pahawh_Hmong,
	"Palmyrene":                          unicode.Palmyrene,
	"Pau_Cin_Hau":                        unicode.Pau_Cin_Hau,
	"Phags_Pa":                           unicode.Phags_Pa,
	"Phoenician":                         unicode.Phoenician,
	"Psalter_Pahlavi":                    unicode.Psalter_Pahlavi,
	"Rejang":                             unicode.Rejang,
	"Runic":                              unicode.Runic,
	"Samaritan":                          unicode.Samaritan,
	"Saurashtra":                         unicode.Saurashtra,
	"Sharada":                            unicode.Sharada,
	"Shavian":                            unicode.Shavian,
	"Siddham":                            unicode.Siddham,
	"SignWriting":                        unicode.SignWriting,
	"Sinhala":                            unicode.Sinhala,
	"Sogdian":                            unicode.Sogdian,
	"Sora_Sompeng":                       unicode.Sora_Sompeng,
	"Soyombo":                            unicode.Soyombo,
	"Sundanese":                          unicode.Sundanese,
	"Syloti_Nagri":                       unicode.Syloti_Nagri,
	"Syriac":                             unicode.Syriac,
	"Tagalog":                            unicode.Tagalog,
	"Tagbanwa":                           unicode.Tagbanwa,
	"Tai_Le":                             unicode.Tai_Le,
	"Tai_Tham":                           unicode.Tai_Tham,
	"Tai_Viet":                           unicode.Tai_Viet,
	"Takri":                              unicode.Takri,
	"Tamil":                              unicode.Tamil,
	"Tangsa":                             unicode.Tangsa,
	"Tangut":                             unicode.Tangut,
	"Telugu":                             unicode.Telugu,
	"Thaana":                             unicode.Thaana,
	"Thai":                               unicode.Thai,
	"Tibetan":                            unicode.Tibetan,
	"Tifinagh":                           unicode.Tifinagh,
	"Tirhuta":                            unicode.Tirhuta,
	"Toto":                               unicode.Toto,
	"Ugaritic":                           unicode.Ugaritic,
	"Vai":                                unicode.Vai,
	"Vithkuqi":                           unicode.Vithkuqi,
	"Wancho":                             unicode.Wancho,
	"Warang_Citi":                        unicode.Warang_Citi,
	"Yezidi":                             unicode.Yezidi,
	"Yi":                                 unicode.Yi,
	"Zanabazar_Square":                   unicode.Zanabazar_Square,
	"Cc":                                 unicode.Cc,
	"Cf":                                 unicode.Cf,
	"Co":                                 unicode.Co,
	"Cs":                                 unicode.Cs,
	"Digit":                              unicode.Digit,
	"Letter":                             unicode.Letter,
	"Lm":                                 unicode.Lm,
	"Lo":                                 unicode.Lo,
	"Lower":                              unicode.Lower,
	"Mark":                               unicode.Mark,
	"Mc":                                 unicode.Mc,
	"Me":                                 unicode.Me,
	"Mn":                                 unicode.Mn,
	"Nl":                                 unicode.Nl,
	"No":                                 unicode.No,
	"Number":                             unicode.Number,
	"Other":                              unicode.Other,
	"Pc":                                 unicode.Pc,
	"Pd":                                 unicode.Pd,
	"Pe":                                 unicode.Pe,
	"Pf":                                 unicode.Pf,
	"Pi":                                 unicode.Pi,
	"Po":                                 unicode.Po,
	"Ps":                                 unicode.Ps,
	"Punct":                              unicode.Punct,
	"Sc":                                 unicode.Sc,
	"Sk":                                 unicode.Sk,
	"Sm":                                 unicode.Sm,
	"So":                                 unicode.So,
	"Space":                              unicode.Space,
	"Symbol":                             unicode.Symbol,
	"Title":                              unicode.Title,
	"Upper":                              unicode.Upper,
	"Zl":                                 unicode.Zl,
	"Zp":                                 unicode.Zp,
	"Zs":                                 unicode.Zs,
}
