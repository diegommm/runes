package util

import (
	"fmt"
	"slices"
	"testing"
)

func TestSliceList(t *testing.T) {
	testCases := []struct {
		s        SliceList[rune]
		expected []rune
	}{
		{nil, nil},
		{SliceList[rune]{}, nil},
		{SliceList[rune]{1}, []rune{1}},
		{SliceList[rune]{1, 2, 3}, []rune{1, 2, 3}},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("index=%v", i), func(t *testing.T) {
			TestOrderedList[rune](t, tc.s, tc.expected...)
		})
	}
}

func TestSeq(t *testing.T) {
	testCases := []struct {
		s        Seq[rune]
		expected []rune
	}{
		{Seq[rune]{1, 1, 1}, []rune{1}},
		{Seq[rune]{10, 10, 10}, []rune{10, 20, 30, 40, 50, 60, 70, 80, 90, 100}},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("index=%v", i), func(t *testing.T) {
			TestOrderedList[rune](t, tc.s, tc.expected...)
		})
	}
}

func TestIterMerge(t *testing.T) {
	testCases := []struct {
		s        []OrderedListIter[rune]
		expected []rune
	}{
		{nil, nil},
		{[]OrderedListIter[rune]{}, nil},
		{
			[]OrderedListIter[rune]{
				SliceList[rune]{1}.Iter(),
			},
			[]rune{1},
		},
		{
			[]OrderedListIter[rune]{
				SliceList[rune]{1, 2, 3, 4, 5, 6}.Iter(),
			},
			[]rune{1, 2, 3, 4, 5, 6},
		},
		{
			[]OrderedListIter[rune]{
				SliceList[rune]{1, 2, 3}.Iter(),
				SliceList[rune]{4, 5, 6}.Iter(),
			},
			[]rune{1, 2, 3, 4, 5, 6},
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("index=%v", i), func(t *testing.T) {
			TestOrderedListIter[rune](t, IterMerge[rune](tc.s...), tc.expected...)
		})
	}

	TestOrderedListIter[rune](t, IterMergeFunc[rune](nil))
}

func TestIterExcept(t *testing.T) {
	testCases := []struct {
		s        OrderedListIter[rune]
		x        OrderedListIter[rune]
		expected []rune
	}{
		{nil, nil, nil},
		{
			SliceList[rune]{}.Iter(),
			nil,
			nil,
		},
		{
			SliceList[rune]{1, 2, 3}.Iter(),
			nil,
			[]rune{1, 2, 3},
		},
		{
			SliceList[rune]{1, 2, 3}.Iter(),
			SliceList[rune]{}.Iter(),
			[]rune{1, 2, 3},
		},
		{
			SliceList[rune]{1, 2, 3}.Iter(),
			SliceList[rune]{1}.Iter(),
			[]rune{2, 3},
		},
		{
			SliceList[rune]{1, 2, 3}.Iter(),
			SliceList[rune]{3}.Iter(),
			[]rune{1, 2},
		},
		{
			SliceList[rune]{1, 2, 3, 4, 5, 6}.Iter(),
			SliceList[rune]{1, 6}.Iter(),
			[]rune{2, 3, 4, 5},
		},
		{
			SliceList[rune]{1, 2, 3, 4, 5, 6}.Iter(),
			SliceList[rune]{2, 3}.Iter(),
			[]rune{1, 4, 5, 6},
		},
		{
			SliceList[rune]{1, 2, 3, 4, 5, 6}.Iter(),
			SliceList[rune]{1, 4, 5, 6}.Iter(),
			[]rune{2, 3},
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("index=%v", i), func(t *testing.T) {
			TestOrderedListIter[rune](t, IterExcept[rune](tc.s, tc.x), tc.expected...)
		})
	}
}

func TestCollect(t *testing.T) {
	testCases := []struct {
		s        OrderedListIter[rune]
		expected []rune
	}{
		{},
		{SliceList[rune]{}.Iter(), nil},
		{SliceList[rune]{1}.Iter(), []rune{1}},
		{SliceList[rune]{1, 2, 3}.Iter(), []rune{1, 2, 3}},
	}

	for i, tc := range testCases {
		Equal(t, true, slices.Equal(tc.expected, Collect(tc.s)), "index=%v", i)
	}
}

func TestIterSeq(t *testing.T) {
	testCases := []struct {
		s        OrderedListIter[rune]
		expected []rune
	}{
		{},
		{SliceList[rune]{}.Iter(), nil},
		{SliceList[rune]{1}.Iter(), []rune{1}},
		{SliceList[rune]{1, 2, 3}.Iter(), []rune{1, 2, 3}},
	}

	for i, tc := range testCases {
		Equal(t, true, slices.Equal(tc.expected, slices.Collect(IterSeq[rune](tc.s))), "index=%v", i)
	}
}
