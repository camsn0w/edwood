package main

import (
	"github.com/rjkroege/edwood/internal/file"
	"testing"
)

// Let's make sure our test fixture has the right form.
func TestBufferDelete(t *testing.T) {
	tab := []struct {
		q0, q1   int
		tb       file.RuneArray
		expected string
	}{
		{0, 5, file.RuneArray([]rune("0123456789")), "56789"},
		{0, 0, file.RuneArray([]rune("0123456789")), "0123456789"},
		{0, 10, file.RuneArray([]rune("0123456789")), ""},
		{1, 5, file.RuneArray([]rune("0123456789")), "056789"},
		{8, 10, file.RuneArray([]rune("0123456789")), "01234567"},
	}
	for _, test := range tab {
		tb := test.tb
		tb.Delete(test.q0, test.q1)
		if string(tb) != test.expected {
			t.Errorf("Delete Failed.  Expected %v, got %v", test.expected, string(tb))
		}
	}
}

func TestBufferInsert(t *testing.T) {
	tab := []struct {
		q0       int
		tb       file.RuneArray
		insert   string
		expected string
	}{
		{5, file.RuneArray([]rune("01234")), "56789", "0123456789"},
		{0, file.RuneArray([]rune("56789")), "01234", "0123456789"},
		{1, file.RuneArray([]rune("06789")), "12345", "0123456789"},
		{5, file.RuneArray([]rune("01234")), "56789", "0123456789"},
	}
	for _, test := range tab {
		tb := test.tb
		tb.Insert(test.q0, []rune(test.insert))
		if string(tb) != test.expected {
			t.Errorf("Insert Failed.  Expected %v, got %v", test.expected, string(tb))
		}
	}
}

func TestBufferIndexRune(t *testing.T) {
	tt := []struct {
		b file.RuneArray
		r rune
		n int
	}{
		{file.RuneArray(nil), '0', -1},
		{file.RuneArray([]rune("01234")), '0', 0},
		{file.RuneArray([]rune("01234")), '3', 3},
		{file.RuneArray([]rune("αβγ")), 'α', 0},
		{file.RuneArray([]rune("αβγ")), 'γ', 2},
	}
	for _, tc := range tt {
		n := tc.b.IndexRune(tc.r)
		if n != tc.n {
			t.Errorf("IndexRune(%v) for buffer %v returned %v; expected %v",
				tc.r, tc.b, n, tc.n)
		}
	}
}

func TestBufferEqual(t *testing.T) {
	tt := []struct {
		a, b file.RuneArray
		ok   bool
	}{
		{file.RuneArray(nil), file.RuneArray(nil), true},
		{file.RuneArray(nil), file.RuneArray([]rune{}), true},
		{file.RuneArray([]rune{}), file.RuneArray(nil), true},
		{file.RuneArray([]rune("01234")), file.RuneArray([]rune("01234")), true},
		{file.RuneArray([]rune("01234")), file.RuneArray([]rune("01x34")), false},
		{file.RuneArray([]rune("αβγ")), file.RuneArray([]rune("αβγ")), true},
		{file.RuneArray([]rune("αβγ")), file.RuneArray([]rune("αλγ")), false},
	}
	for _, tc := range tt {
		ok := tc.a.Equal(tc.b)
		if ok != tc.ok {
			t.Errorf("Equal(%v) for buffer %v returned %v; expected %v",
				tc.b, tc.a, ok, tc.ok)
		}
	}
}
