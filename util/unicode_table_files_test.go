package util

import (
	"bufio"
	"io"
	"os"
	"strconv"
	"strings"
	"testing"
	"unicode"
)

const (
	writeTableFilesEnvVar = "WRITE_TABLE_FILES"
	tableFilesDir         = "range_tables.out" // ends in ".out" to git-ignore it
)

func TestWriteTableFiles(t *testing.T) {
	v, err := strconv.ParseBool(os.Getenv(writeTableFilesEnvVar))
	if err != nil {
		t.Skipf("NOTE: if you want to dump \"unicode\"'s package tables to "+
			"one file per table in this directory, with one decimal rune per "+
			"line, ordered asc, then set the environment variable %q to a "+
			"truthy value (i.e. something strconv.ParseBool deems to be true)",
			writeTableFilesEnvVar)
	}
	if !v {
		t.Skip()
	}
	err = os.Mkdir(tableFilesDir, 0o755)
	if err != nil && !strings.Contains(err.Error(), "file exists") {
		t.Fatalf("create directory %q for tables: %v", tableFilesDir, err)
	}

	for name, rt := range tablesToWrite {
		writeTableFile(t, name, rt)
	}
}

func writeTableFile(t *testing.T, name string, rt *unicode.RangeTable) {
	const mode = os.O_CREATE | os.O_TRUNC | os.O_WRONLY
	filename := tableFilesDir + "/" + strings.ToLower(name)
	f, err := os.OpenFile(filename, mode, 0o644)
	MustEqual(t, true, err == nil, "open file %q for write: %v", filename, err)
	defer func() {
		err := f.Close()
		MustEqual(t, true, err == nil, "close file %q: %v", filename, err)
	}()
	w := newErrRuneWriter(f)
	for r := range RangeTableIter(rt) {
		w.WriteString(strconv.Itoa(int(r)) + "\n")
	}
	MustEqual(t, true, w.err == nil, "write to file %q: %v", filename, err)
	err = w.Flush()
	MustEqual(t, true, err == nil, "flush changes to file %q: %v", filename, err)
}

func newErrRuneWriter(w io.Writer) *errRuneWriter {
	return &errRuneWriter{Writer: bufio.NewWriter(w)}
}

type errRuneWriter struct {
	*bufio.Writer
	err error
}

func (w *errRuneWriter) WriteString(s string) (size int) {
	if w.err != nil {
		return 0
	}
	size, w.err = w.Writer.WriteString(s)
	return
}

var tablesToWrite = map[string]*unicode.RangeTable{
	"Cc":     unicode.Cc,
	"Cf":     unicode.Cf,
	"Co":     unicode.Co,
	"Cs":     unicode.Cs,
	"Digit":  unicode.Digit,
	"Nd":     unicode.Nd,
	"Letter": unicode.Letter,
	"L":      unicode.L,
	"Lm":     unicode.Lm,
	"Lo":     unicode.Lo,
	"Lower":  unicode.Lower,
	"Ll":     unicode.Ll,
	"Mark":   unicode.Mark,
	"M":      unicode.M,
	"Mc":     unicode.Mc,
	"Me":     unicode.Me,
	"Mn":     unicode.Mn,
	"Nl":     unicode.Nl,
	"No":     unicode.No,
	"Number": unicode.Number,
	"N":      unicode.N,
	"Other":  unicode.Other,
	"C":      unicode.C,
	"Pc":     unicode.Pc,
	"Pd":     unicode.Pd,
	"Pe":     unicode.Pe,
	"Pf":     unicode.Pf,
	"Pi":     unicode.Pi,
	"Po":     unicode.Po,
	"Ps":     unicode.Ps,
	"Punct":  unicode.Punct,
	"P":      unicode.P,
	"Sc":     unicode.Sc,
	"Sk":     unicode.Sk,
	"Sm":     unicode.Sm,
	"So":     unicode.So,
	"Space":  unicode.Space,
	"Z":      unicode.Z,
	"Symbol": unicode.Symbol,
	"S":      unicode.S,
	"Title":  unicode.Title,
	"Lt":     unicode.Lt,
	"Upper":  unicode.Upper,
	"Lu":     unicode.Lu,
	"Zl":     unicode.Zl,
	"Zp":     unicode.Zp,
	"Zs":     unicode.Zs,

	"Adlam":                  unicode.Adlam,
	"Ahom":                   unicode.Ahom,
	"Anatolian_Hieroglyphs":  unicode.Anatolian_Hieroglyphs,
	"Arabic":                 unicode.Arabic,
	"Armenian":               unicode.Armenian,
	"Avestan":                unicode.Avestan,
	"Balinese":               unicode.Balinese,
	"Bamum":                  unicode.Bamum,
	"Bassa_Vah":              unicode.Bassa_Vah,
	"Batak":                  unicode.Batak,
	"Bengali":                unicode.Bengali,
	"Bhaiksuki":              unicode.Bhaiksuki,
	"Bopomofo":               unicode.Bopomofo,
	"Brahmi":                 unicode.Brahmi,
	"Braille":                unicode.Braille,
	"Buginese":               unicode.Buginese,
	"Buhid":                  unicode.Buhid,
	"Canadian_Aboriginal":    unicode.Canadian_Aboriginal,
	"Carian":                 unicode.Carian,
	"Caucasian_Albanian":     unicode.Caucasian_Albanian,
	"Chakma":                 unicode.Chakma,
	"Cham":                   unicode.Cham,
	"Cherokee":               unicode.Cherokee,
	"Chorasmian":             unicode.Chorasmian,
	"Common":                 unicode.Common,
	"Coptic":                 unicode.Coptic,
	"Cuneiform":              unicode.Cuneiform,
	"Cypriot":                unicode.Cypriot,
	"Cypro_Minoan":           unicode.Cypro_Minoan,
	"Cyrillic":               unicode.Cyrillic,
	"Deseret":                unicode.Deseret,
	"Devanagari":             unicode.Devanagari,
	"Dives_Akuru":            unicode.Dives_Akuru,
	"Dogra":                  unicode.Dogra,
	"Duployan":               unicode.Duployan,
	"Egyptian_Hieroglyphs":   unicode.Egyptian_Hieroglyphs,
	"Elbasan":                unicode.Elbasan,
	"Elymaic":                unicode.Elymaic,
	"Ethiopic":               unicode.Ethiopic,
	"Georgian":               unicode.Georgian,
	"Glagolitic":             unicode.Glagolitic,
	"Gothic":                 unicode.Gothic,
	"Grantha":                unicode.Grantha,
	"Greek":                  unicode.Greek,
	"Gujarati":               unicode.Gujarati,
	"Gunjala_Gondi":          unicode.Gunjala_Gondi,
	"Gurmukhi":               unicode.Gurmukhi,
	"Han":                    unicode.Han,
	"Hangul":                 unicode.Hangul,
	"Hanifi_Rohingya":        unicode.Hanifi_Rohingya,
	"Hanunoo":                unicode.Hanunoo,
	"Hatran":                 unicode.Hatran,
	"Hebrew":                 unicode.Hebrew,
	"Hiragana":               unicode.Hiragana,
	"Imperial_Aramaic":       unicode.Imperial_Aramaic,
	"Inherited":              unicode.Inherited,
	"Inscriptional_Pahlavi":  unicode.Inscriptional_Pahlavi,
	"Inscriptional_Parthian": unicode.Inscriptional_Parthian,
	"Javanese":               unicode.Javanese,
	"Kaithi":                 unicode.Kaithi,
	"Kannada":                unicode.Kannada,
	"Katakana":               unicode.Katakana,
	"Kawi":                   unicode.Kawi,
	"Kayah_Li":               unicode.Kayah_Li,
	"Kharoshthi":             unicode.Kharoshthi,
	"Khitan_Small_Script":    unicode.Khitan_Small_Script,
	"Khmer":                  unicode.Khmer,
	"Khojki":                 unicode.Khojki,
	"Khudawadi":              unicode.Khudawadi,
	"Lao":                    unicode.Lao,
	"Latin":                  unicode.Latin,
	"Lepcha":                 unicode.Lepcha,
	"Limbu":                  unicode.Limbu,
	"Linear_A":               unicode.Linear_A,
	"Linear_B":               unicode.Linear_B,
	"Lisu":                   unicode.Lisu,
	"Lycian":                 unicode.Lycian,
	"Lydian":                 unicode.Lydian,
	"Mahajani":               unicode.Mahajani,
	"Makasar":                unicode.Makasar,
	"Malayalam":              unicode.Malayalam,
	"Mandaic":                unicode.Mandaic,
	"Manichaean":             unicode.Manichaean,
	"Marchen":                unicode.Marchen,
	"Masaram_Gondi":          unicode.Masaram_Gondi,
	"Medefaidrin":            unicode.Medefaidrin,
	"Meetei_Mayek":           unicode.Meetei_Mayek,
	"Mende_Kikakui":          unicode.Mende_Kikakui,
	"Meroitic_Cursive":       unicode.Meroitic_Cursive,
	"Meroitic_Hieroglyphs":   unicode.Meroitic_Hieroglyphs,
	"Miao":                   unicode.Miao,
	"Modi":                   unicode.Modi,
	"Mongolian":              unicode.Mongolian,
	"Mro":                    unicode.Mro,
	"Multani":                unicode.Multani,
	"Myanmar":                unicode.Myanmar,
	"Nabataean":              unicode.Nabataean,
	"Nag_Mundari":            unicode.Nag_Mundari,
	"Nandinagari":            unicode.Nandinagari,
	"New_Tai_Lue":            unicode.New_Tai_Lue,
	"Newa":                   unicode.Newa,
	"Nko":                    unicode.Nko,
	"Nushu":                  unicode.Nushu,
	"Nyiakeng_Puachue_Hmong": unicode.Nyiakeng_Puachue_Hmong,
	"Ogham":                  unicode.Ogham,
	"Ol_Chiki":               unicode.Ol_Chiki,
	"Old_Hungarian":          unicode.Old_Hungarian,
	"Old_Italic":             unicode.Old_Italic,
	"Old_North_Arabian":      unicode.Old_North_Arabian,
	"Old_Permic":             unicode.Old_Permic,
	"Old_Persian":            unicode.Old_Persian,
	"Old_Sogdian":            unicode.Old_Sogdian,
	"Old_South_Arabian":      unicode.Old_South_Arabian,
	"Old_Turkic":             unicode.Old_Turkic,
	"Old_Uyghur":             unicode.Old_Uyghur,
	"Oriya":                  unicode.Oriya,
	"Osage":                  unicode.Osage,
	"Osmanya":                unicode.Osmanya,
	"Pahawh_Hmong":           unicode.Pahawh_Hmong,
	"Palmyrene":              unicode.Palmyrene,
	"Pau_Cin_Hau":            unicode.Pau_Cin_Hau,
	"Phags_Pa":               unicode.Phags_Pa,
	"Phoenician":             unicode.Phoenician,
	"Psalter_Pahlavi":        unicode.Psalter_Pahlavi,
	"Rejang":                 unicode.Rejang,
	"Runic":                  unicode.Runic,
	"Samaritan":              unicode.Samaritan,
	"Saurashtra":             unicode.Saurashtra,
	"Sharada":                unicode.Sharada,
	"Shavian":                unicode.Shavian,
	"Siddham":                unicode.Siddham,
	"SignWriting":            unicode.SignWriting,
	"Sinhala":                unicode.Sinhala,
	"Sogdian":                unicode.Sogdian,
	"Sora_Sompeng":           unicode.Sora_Sompeng,
	"Soyombo":                unicode.Soyombo,
	"Sundanese":              unicode.Sundanese,
	"Syloti_Nagri":           unicode.Syloti_Nagri,
	"Syriac":                 unicode.Syriac,
	"Tagalog":                unicode.Tagalog,
	"Tagbanwa":               unicode.Tagbanwa,
	"Tai_Le":                 unicode.Tai_Le,
	"Tai_Tham":               unicode.Tai_Tham,
	"Tai_Viet":               unicode.Tai_Viet,
	"Takri":                  unicode.Takri,
	"Tamil":                  unicode.Tamil,
	"Tangsa":                 unicode.Tangsa,
	"Tangut":                 unicode.Tangut,
	"Telugu":                 unicode.Telugu,
	"Thaana":                 unicode.Thaana,
	"Thai":                   unicode.Thai,
	"Tibetan":                unicode.Tibetan,
	"Tifinagh":               unicode.Tifinagh,
	"Tirhuta":                unicode.Tirhuta,
	"Toto":                   unicode.Toto,
	"Ugaritic":               unicode.Ugaritic,
	"Vai":                    unicode.Vai,
	"Vithkuqi":               unicode.Vithkuqi,
	"Wancho":                 unicode.Wancho,
	"Warang_Citi":            unicode.Warang_Citi,
	"Yezidi":                 unicode.Yezidi,
	"Yi":                     unicode.Yi,
	"Zanabazar_Square":       unicode.Zanabazar_Square,

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
	"STerm":                              unicode.STerm,
	"Sentence_Terminal":                  unicode.Sentence_Terminal,
	"Soft_Dotted":                        unicode.Soft_Dotted,
	"Terminal_Punctuation":               unicode.Terminal_Punctuation,
	"Unified_Ideograph":                  unicode.Unified_Ideograph,
	"Variation_Selector":                 unicode.Variation_Selector,
	"White_Space":                        unicode.White_Space,
}
