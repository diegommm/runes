//go:build ignore

package main

import (
	"cmp"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"slices"
	"strconv"
	"unicode"

	"github.com/diegommm/runes"
)

func main() {
	if true {
		stats := make([]*runes.RangeTableStats, 0, len(runes.RangeTables))
		for name, rt := range runes.RangeTables {
			stats = append(stats, runes.GetRangeTableStats(name, rt))
		}
		slices.SortFunc(stats, func(a, b *runes.RangeTableStats) int {
			return cmp.Compare(a.Name, b.Name)
		})

		if err := printJSON(os.Stdout, stats); err != nil {
			panic(err)
		}
	}
	if false {
		if err := printRangeTableCSV(os.Stdout, runes.RangeTables); err != nil {
			panic(err)
		}
	}
}

func printBytesDiff(w io.Writer, stats []*runes.RangeTableStats) error {
	var totalOld, totalNew int
	for _, stat := range stats {
		totalOld += stat.ByteSize
		totalNew += stat.EstimatedNewByteSize
	}
	_, err := fmt.Fprintf(w, "Total bytes old: %v\n"+
		"Total bytes new: %v\n"+
		"Diff: %v\n", totalOld, totalNew, totalNew-totalOld)
	return err
}

func printJSON(w io.Writer, stats []*runes.RangeTableStats) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "\t")
	return enc.Encode(stats)
}

func printCSV(w io.Writer, stats []*runes.RangeTableStats) error {
	csvw := csv.NewWriter(w)

	record := []string{"name", "byte_size", "new_byte_size", "span", "density"}
	csvw.Write(record)
	for _, stat := range stats {
		record[0] = stat.Name
		record[1] = strconv.Itoa(stat.ByteSize)
		record[2] = strconv.Itoa(stat.EstimatedNewByteSize)
		record[3] = strconv.Itoa(int(stat.Span))
		record[4] = strconv.FormatFloat(float64(stat.Density), 'f', -1, 64)
		csvw.Write(record)
	}

	csvw.Flush()
	return csvw.Error()
}

func printRangeTableCSV(w io.Writer, m map[string]*unicode.RangeTable) error {
	csvw := csv.NewWriter(w)

	record := []string{"name", "width", "pos", "lo", "hi", "stride"}
	csvw.Write(record)
	for name, rt := range m {
		for i, t := range rt.R16 {
			record[0] = name
			record[2] = "16"
			record[1] = strconv.Itoa(i)
			record[3] = strconv.Itoa(int(t.Lo))
			record[4] = strconv.Itoa(int(t.Hi))
			record[5] = strconv.Itoa(int(t.Stride))
			csvw.Write(record)
		}
		for i, t := range rt.R32 {
			record[0] = name
			record[2] = "32"
			record[1] = strconv.Itoa(i)
			record[3] = strconv.Itoa(int(t.Lo))
			record[4] = strconv.Itoa(int(t.Hi))
			record[5] = strconv.Itoa(int(t.Stride))
			csvw.Write(record)
		}
	}

	csvw.Flush()
	return csvw.Error()
}
