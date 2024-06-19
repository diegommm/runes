package runes

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
	"unicode"
)

func makeContiguousRunesString(start, count rune) string {
	var b strings.Builder
	b.Grow(int(count)) // could be bigger, nvm
	for i := rune(0); i < count; i++ {
		b.WriteRune(start + i)
	}
	return b.String()
}

// isMaskSlice32Wrapper is a wrapper for fair comparison.
func isMaskSlice32Wrapper(s []rune) func(rune) bool {
	minRune, span := minAndSpan(s)
	return isMaskSlice32(s, minRune, span)
}

// isMaskSlice64Wrapper is a wrapper for fair comparison.
func isMaskSlice64Wrapper(s []rune) func(rune) bool {
	minRune, span := minAndSpan(s)
	return isMaskSlice64(s, minRune, span)
}

// isMask64Wrapper is a wrapper for fair comparison.
func isMask64Wrapper(s []rune) func(rune) bool {
	minRune, span := minAndSpan(s)
	if span >= 64 {
		return nil
	}

	return isMask64(s, minRune)
}

func exclude(orig, exclude map[rune]struct{}) {
	for r := range exclude {
		delete(orig, r)
	}
}

func BenchmarkIsInSetFunc(b *testing.B) {
	// control which benchmarks to run

	const (
		executeStdlib          = false
		executeSmallOnes       = false
		executeIsInString      = false
		executeIsInTable       = false
		executeIsInSlice       = false
		executeIsInMap         = false
		executeIsInMask64      = false
		executeIsInMaskSlice32 = true
		executeIsInMaskSlice64 = true
		executeIsInSparseSet   = false
		executeIsInStrategy    = false
	)

	const (
		benchInit    = false
		benchRuntime = true
	)

	// skipFunc will cause a benchmark to be skipped if it returns true
	skipFunc := func(length int, minRune, span rune) bool {
		return false
	}

	// define benchmarks

	benchCases := []struct {
		name  string
		input string
	}{
		{name: "zero", input: ""},
		{name: "ASCII", input: "A"},
		{name: "Non-ASCII", input: "Ñ"},
		{name: "ASCII", input: "AB"},
		{name: "Non-ASCII", input: "Ñ世"},
		{name: "ASCII", input: makeContiguousRunesString(' ', 24)},
		{name: "ASCII", input: makeContiguousRunesString(' ', 25)},
		{name: "ASCII", input: makeContiguousRunesString(' ', 26)},
		{name: "ASCII", input: makeContiguousRunesString(' ', 64)},
		{name: "Non-ASCII", input: makeContiguousRunesString('Ñ', 64)},
		{name: "ASCII", input: makeContiguousRunesString(' ', 65)},
		{name: "Non-ASCII", input: makeContiguousRunesString('Ñ', 65)},
		{name: "ASCII", input: makeContiguousRunesString(0, 127)},
		{name: "Non-ASCII", input: makeContiguousRunesString('Ñ', 127)},
		{name: "ASCII+", input: makeContiguousRunesString(0, 128)},
		{name: "Non-ASCII", input: makeContiguousRunesString('Ñ', 128)},
		{name: "Other", input: makeContiguousRunesString(1024, 256)},
		{name: "Other", input: makeContiguousRunesString(1024, 512)},
		{name: "Other", input: makeContiguousRunesString(1024, 1025)},
		{
			name:  "Full Unicode span, few items",
			input: string([]rune{0, 1, 2, '\U0010FFFF'}),
		},
		{
			name:  "Half Unicode span, few items",
			input: string([]rune{0, 1, 2, '\U0010FFFF' / 2}),
		},
		{
			name:  "a lot",
			input: makeContiguousRunesString('\U0010FFFF'/2, 13_000),
		},
		{
			name:  "everything",
			input: makeContiguousRunesString(0, '\U0010FFFF'),
		},
	}

	// stdlib rune-based implementations

	var (
		stdlibBytesContainsRune = func(r rune) bool {
			return bytes.ContainsRune(testDataBytes, r)
		}
		stdlibStringsContainsRune = func(r rune) bool {
			return strings.ContainsRune(rawTestData, r)
		}
	)

	// init benchmarking helpers (overhead negligible)

	benchInitFunc := func(b *testing.B, cond bool, name string, f func() func(rune) bool) {
		if !cond {
			return
		}
		b.Run(name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				if f() == nil {
					b.FailNow()
				}
			}
		})
	}
	benchInitString := func(b *testing.B, cond bool, name string, s string, f func(string) func(rune) bool) {
		benchInitFunc(b, cond, name, func() func(rune) bool {
			return f(s)
		})
	}
	benchInitMap := func(b *testing.B, cond bool, name string, s map[rune]struct{}, f func(map[rune]struct{}) func(rune) bool) {
		benchInitFunc(b, cond, name, func() func(rune) bool {
			return f(s)
		})
	}
	benchInitSlice := func(b *testing.B, cond bool, name string, s []rune, f func([]rune) func(rune) bool) {
		benchInitFunc(b, cond, name, func() func(rune) bool {
			return f(s)
		})
	}

	// runtime benchmarking helpers (overhead negligible)

	benchRuntimeFunc := func(b *testing.B, cond bool, name string, f func(rune) bool) {
		if !cond {
			return
		}
		b.Run(name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				for _, r := range testDataSlice {
					f(r)
				}
			}
		})
	}

	// run benchmarks for each alternative

	for _, bc := range benchCases {
		m := stringToRuneMap(bc.input)
		exclude(m, testDataMap) // ensure negative match
		sl := runeMapToSlice(m)
		str := string(sl)

		minRune, span := minAndSpan(sl)
		if skipFunc(len(sl), minRune, span) {
			continue
		}

		// implementations in this package

		var (
			isStringFunc      = isString(str)
			isTableFunc       = isTable(sl)
			isSliceFunc       = isSlice(sl)
			isMapFunc         = isMap(m)
			isMask64Func      = isMask64Wrapper(sl)
			isMaskSlice32Func = isMaskSlice32Wrapper(sl)
			isMaskSlice64Func = isMaskSlice64Wrapper(sl)
			isSparseSetFunc   = isSparseSet(sl)
			isStrategyFunc    = isStrategy(sl)
		)

		b.Run(fmt.Sprintf("[len=%d min=%v span=%v init] %s", len(sl), minRune, span, bc.name), func(b *testing.B) {
			if !benchInit {
				b.SkipNow()
			}

			benchInitString(b, executeIsInString, "isString", str, isString)
			benchInitSlice(b, executeIsInTable, "isTable", sl, isTable)
			benchInitSlice(b, executeIsInSlice, "isSlice", sl, isSlice)
			benchInitMap(b, executeIsInMap, "isMap", m, isMap)
			benchInitSlice(b, executeIsInMask64 && isMask64Func != nil, "isMask64", sl, isMask64Wrapper)
			benchInitSlice(b, executeIsInMaskSlice32, "isMaskSlice32", sl, isMaskSlice32Wrapper)
			benchInitSlice(b, executeIsInMaskSlice64, "isMaskSlice64", sl, isMaskSlice64Wrapper)
			benchInitSlice(b, executeIsInSparseSet, "isSparseSet", sl, isSparseSet)
			benchInitSlice(b, executeIsInStrategy, "isStrategy", sl, isStrategy)
		})

		b.Run(fmt.Sprintf("[len=%d min=%v span=%v runtime] %s", len(sl), minRune, span, bc.name), func(b *testing.B) {
			if !benchRuntime {
				b.SkipNow()
			}

			if executeStdlib {
				b.Run("stdlib", func(b *testing.B) {
					if len(sl) > 1 {
						// doesn't make sense to compare for len(sl)<2 since
						// *.ContainsAny will short-circuit

						b.Run("bytes.ContainsAny", func(b *testing.B) {
							for i := 0; i < b.N; i++ {
								if bytes.ContainsAny(testDataBytes, str) {
									b.FailNow()
								}
							}
						})

						b.Run("strings.ContainsAny", func(b *testing.B) {
							for i := 0; i < b.N; i++ {
								if strings.ContainsAny(rawTestData, str) {
									b.FailNow()
								}
							}
						})
					}

					benchRuntimeFunc(b, true, "bytes.ContainsRune", stdlibBytesContainsRune)
					benchRuntimeFunc(b, true, "strings.ContainsRune", stdlibStringsContainsRune)
				})
			}

			switch len(sl) {
			case 0:
				benchRuntimeFunc(b, executeSmallOnes, "zero", isNeverIn)
			case 1:
				benchRuntimeFunc(b, executeSmallOnes, "one", func(r rune) bool {
					return r == sl[0]
				})
			case 2:
				benchRuntimeFunc(b, executeSmallOnes, "two with or", func(r rune) bool {
					return r == sl[0] || r == sl[1]
				})
				benchRuntimeFunc(b, executeSmallOnes, "two with switch", func(r rune) bool {
					switch r {
					case sl[0], sl[1]:
						return true
					}
					return false
				})
			case 3:
				benchRuntimeFunc(b, executeSmallOnes, "three", func(r rune) bool {
					return r == sl[0] || r == sl[1] || r == sl[2]
				})
			}

			benchRuntimeFunc(b, executeIsInString, "isString", isStringFunc)
			benchRuntimeFunc(b, executeIsInTable, "isTable", isTableFunc)
			benchRuntimeFunc(b, executeIsInSlice, "isSlice", isSliceFunc)
			benchRuntimeFunc(b, executeIsInMap, "isMap", isMapFunc)
			benchRuntimeFunc(b, executeIsInMask64 && isMask64Func != nil, "isMask64", isMask64Func)
			benchRuntimeFunc(b, executeIsInMaskSlice32, "isMaskSlice32", isMaskSlice32Func)
			benchRuntimeFunc(b, executeIsInMaskSlice64, "isMaskSlice64", isMaskSlice64Func)
			benchRuntimeFunc(b, executeIsInSparseSet, "isSparseSet", isSparseSetFunc)
			benchRuntimeFunc(b, executeIsInStrategy, "isStrategy", isStrategyFunc)
		})
	}
}

// format test data in different ways

var (
	testDataMap   = stringToRuneMap(rawTestData)
	testDataSlice = []rune(rawTestData)
	testDataBytes = []byte(rawTestData)
)

// rawTestData will have all negative matches to test worst case scenario.
const rawTestData = `
（聆聽ˈɛːˈɛ或聆聽ˈːə，，结构化查询语言）是一种特定目的程式语言，用于管理关系数据库管理系统（），或在关系流数据管理系统（）中进行流处理。
基于关系代数和元组关系演算，包括一个数据定义语言和数据操纵语言。的范围包括数据插入、查询、更新和删除，数据库模式创建和修改，以及数据访问控制。尽管经常被描述为，而且很大程度上是一种声明式编程（），但是其也含有过程式编程的元素。
是对埃德加·科德的关系模型的第一个商业化语言实现，这一模型在其年的一篇具有影响力的���文《一个对于大型共享型数据库的关系模型》中被描述。尽管并非完全按照科德的关系模型设计，但其依然成为最为广泛运用的数据库语言。
在年成为美国国家标准学会（）的一项标准，在年成为国际标准化组织（）标准。此后，这一标准经过了一系列的增订，加入了大量新特性。虽然有这一标准的存在，但大部分的代码在不同的数据库系统中并不具�	完全的跨平台性。
歷史
在年代初，由研究院下属愛曼登研究中心的埃德加·科德發表將資料組成表格的應用原則（）。年，同一實驗室的唐纳德·钱柏林和雷蒙德·博伊斯参考了科德的模型后，在研制关系数据库管理系统中，开发出了一套規範語言（，结构化英语查询语言），並在年月的《研究与开发杂志》上公布��版本的（叫）。年改名為。
年，甲骨文公司（当时名为关系式软件公司）首先提供商用的，公司在和数据库系统中也实现了。
年月，美国采用作为关系数据库管理系统的标准语言（），后为国际标准化组织（）采纳为国际标准。
年，美国采纳在报告中定义的关系数据库管理系统的标准语言，称为，���标准替代版本。该标准为下列组织所采纳：
国际标准化组织，为报告《》
美国联邦政府，发布在《》
目前，所有主要的关系数据库管理系统支持某些形式的，大部分数据库至少遵守标准。
标准在交叉连接（）和内部连接之上，新增加了外部连接，��支持在子句中写连接表达式。支持集合的并运算、交运算。支持表达式。支持约束。创建临时表。支持。支持事务隔离。
语法
主条目：语法
图表显示了语言元素组成的一个语句
语言分成了几种要素，包括：
子句，是语句和查询的组成成分。（在某些情况下，这些都是可选的。）
表达式，可以产生任何标量值，或由列和行组成的数据库表
谓词，给需要��估的三值逻辑（）（）或布尔真值指定条件，并限制语句和查询的效果，或改变程序流程。
查询，基于特定条件检索数据。这是的一个重要组成部分。
语句，可以持久地影响纲要和数据，也可以控制数据库事务、程序流程、连接、会话或诊断。
语句也包括分號（）语句终结符。尽管并不是每个平台都必需，但它是作为语法的标准部分定义的。
无意义的空��在语句和查询中一般会被忽略，更容易格式化代码便于阅读。
语言特点
是高级的非過程化編程語言，它允许用户在高层数据结构上工作。它不要求用户指定对数据的存放方法，也不需要用户了解其具体的数据存放方式。而它的界面，能使具有底层结构完全不同的数据库系统和不同数据库之间，使用相同的作为数据的输入与管理。它以记录项目〔〕的合集（）〔项集，〕作为操纵对象，所有语句接受项集作为输入，回送出的项集作为输出，这种项集特性允许一条语句的输出作为另一条语句的输入，所以语句可以嵌套，这使它拥有极大的灵活性和强大的功能。在多数情况下，在其他編程語言中需要用一大段程序才可实践的一个单独事件，而其在上只需要一个语句就可以被表达出来。这也意味着用可以写出非常复杂的语句，在不特�考慮效能下。
同時也是数据库文件格式的扩展名。
`

func BenchmarkUnicodeIs(b *testing.B) {
	generateUnicodeIsFuncsMap()

	// Run with:
	//	time go test -count=10 -benchmem -timeout=30m -bench=BenchmarkUnicodeIs -run=-

	benchFunc := func(b *testing.B, name string, s []rune, f func(rune) bool) {
		b.Run(name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				for _, r := range s {
					f(r)
				}
			}
		})
	}

	for name, bench := range unicodeIsFuncsMap {
		if unicode.IsLower(rune(name[0])) {
			continue // skip dumb funcs
		}

		b.Run(name, func(b *testing.B) {
			f := IsRunesFunc(bench.runes)
			b.Run("matching only", func(b *testing.B) {
				benchFunc(b, "stdlib", bench.runes, bench.f)
				benchFunc(b, "IsRunesFunc", bench.runes, f)
			})
			b.Run("all valid UTF8", func(b *testing.B) {
				benchFunc(b, "stdlib", validUTF8, bench.f)
				benchFunc(b, "IsRunesFunc", validUTF8, f)
			})
		})
	}
}
