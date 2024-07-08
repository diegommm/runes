package runes

import "testing"

func TestBuffer_String(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		expected, actual string
	}{
		{"asd", newBuffer().raw([]byte{'a', 's', 'd'}).Err().Error()},
		{"asd", newBuffer().str("asd").String()},
		{"a", newBuffer().byte('a').String()},
		{"asd", newBuffer().str("asd").String()},
		{"0", newBuffer().int32(0).String()},
		{"10", newBuffer().rune(10).String()},
		{"-10", newBuffer().int(-10).String()},
		{"18446744073709551615", newBuffer().uint64(1<<64 - 1).String()},
		{"9223372036854775807", newBuffer().int64(1<<63 - 1).String()},
		{"-9223372036854775808", newBuffer().int64(-(1 << 63)).String()},
		{"[1,2,3]", newBuffer().write(toBufferWriter(intsToBuffer,
			[]int{1, 2, 3})).String()},
		{`["x","y"]`, bufferString(toBufferWriterv(rawsToBuffer,
			`"x"`, `"y"`))},
	}

	for i, tc := range testCases {
		if tc.expected != tc.actual {
			t.Fatalf("[%d] expected: %q, actual: %q, bytes: %v", i, tc.expected,
				tc.actual, []byte(tc.actual))
		}
	}
}
