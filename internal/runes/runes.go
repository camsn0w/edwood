// Package runes implements functions for the manipulation of rune slices.
package runes

import (
	"strings"
	"unicode/utf8"
)

// HasPrefix tests whether the rune slice s begins with prefix.
func HasPrefix(s, prefix []rune) bool {
	if len(prefix) > len(s) {
		return false
	}
	for i, r := range prefix {
		if s[i] != r {
			return false
		}
	}
	return true
}

// Index returns the index of the first instance of sep in s, or -1 if sep is not present in s.
func Index(s, sep []rune) int {
	n := len(sep)
	switch {
	case n > len(s):
		return -1
	case n == 0:
		return 0
	}
	for i := range s[:len(s)-n+1] {
		if HasPrefix(s[i:], sep) {
			return i
		}
	}
	return -1
}

// IndexRune returns the index of the first occurrence in s of the given rune r.
// It returns -1 if rune is not present in s.
func IndexRune(s []rune, r rune) int {
	for i, c := range s {
		if c == r {
			return i
		}
	}
	return -1
}

// ContainsRune reports whether the rune is contained in the runes slice s.
func ContainsRune(s []rune, r rune) bool {
	return IndexRune(s, r) >= 0
}

// Equal returns a boolean reporting whether a and b
// are the same length and contain the same runes.
func Equal(a, b []rune) bool {
	if len(a) != len(b) {
		return false
	}
	for i, r := range a {
		if r != b[i] {
			return false
		}
	}
	return true
}

// TrimLeft returns a subslice of s by slicing off all leading
// UTF-8-encoded code points contained in cutset.
func TrimLeft(s []rune, cutset string) []rune {
	switch {
	case len(s) == 0:
		return nil
	case len(cutset) == 0:
		return s
	}
	inCutset := func(r rune) bool {
		for _, c := range cutset {
			if c == r {
				return true
			}
		}
		return false
	}
	for i, r := range s {
		if !inCutset(r) {
			return s[i:]
		}
	}
	return nil
}

// Cvttorunes decodes runes r from p. It's guaranteed that first n
// bytes of p will be interpreted without worrying about partial runes.
// This may mean reading up to UTFMax-1 more bytes than n; the caller
// must ensure p is large enough. Partial runes and invalid encodings
// are converted to RuneError. Nb (always >= n) is the number of bytes
// interpreted.
//
// If any U+0000 rune is present in r, they are elided and nulls is set
// to true.
func Cvttorunes(p []byte, n int) (r []rune, nb int, nulls bool) {
	for nb < n {
		var w int
		var ru rune
		if p[nb] < utf8.RuneSelf {
			w = 1
			ru = rune(p[nb])
		} else {
			ru, w = utf8.DecodeRune(p[nb:])
		}
		if ru != 0 {
			r = append(r, ru)
		} else {
			nulls = true
		}
		nb += w
	}
	return
}

// TODO(flux) The "correct" answer here is return unicode.IsNumber(c) || unicode.IsLetter(c)
func Isalnum(c rune) bool {
	// Hard to get absolutely right.  Use what we know about ASCII
	// and assume anything above the Latin control characters is
	// potentially an alphanumeric.
	if c <= ' ' {
		return false
	}
	if 0x7F <= c && c <= 0xA0 {
		return false
	}
	if strings.ContainsRune("!\"#$%&'()*+,-./:;<=>?@[\\]^`{|}~", c) {
		return false
	}
	return true
}

func Bytetorune(s []byte) []rune {
	r, _, _ := Cvttorunes(s, len(s))
	return r
}

const QuoteChar = '\''

func NeedsQuote(s string) bool {
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == QuoteChar || c <= ' ' { // quote, blanks, or control characters
			return true
		}
	}
	return false
}

// Quote adds single quotes to s in the style of rc(1) if they are needed.
// The behaviour should be identical to Plan 9's quote(3).
func Quote(s string) string {
	if s == "" {
		return "''"
	}
	if !NeedsQuote(s) {
		return s
	}
	var b strings.Builder
	b.Grow(10 + len(s)) // Enough room for few quotes
	b.WriteByte(QuoteChar)
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == QuoteChar {
			b.WriteByte(QuoteChar)
		}
		b.WriteByte(c)
	}
	b.WriteByte(QuoteChar)
	return b.String()
}

func Skipbl(r []rune) []rune {
	return TrimLeft(r, " \t\n")
}
