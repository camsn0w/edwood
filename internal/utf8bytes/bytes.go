// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package utf8Bytes provides an efficient way to index bytes by rune rather than by byte.
// utf8Bytes is a modified version of utf8string found at "https://cs.opensource.google/go/x/exp/+/master:utf8string/"
package utf8bytes // import "golang.org/x/exp/utf8Bytes"

import (
	"errors"
	"unicode/utf8"

	"github.com/rjkroege/edwood/internal/undo"
)

// Bytes wraps a regular bytes with a small structure that provides more
// efficient indexing by code point index, as opposed to byte index.
// Scanning incrementally forwards or backwards is O(1) per index operation
// (although not as fast a range clause going forwards).  Random access is
// O(N) in the length of the string, but the overhead is less than always
// scanning from the beginning.
// If the string is ASCII, random access is O(1).
type Bytes struct {
	buf      *undo.Buffer
	numRunes int
	// If width > 0, the rune at runePos starts at bytePos and has the specified width.
	width        int
	bytePos      int
	runePos      int
	nonASCII     int // byte index of the first non-ASCII rune.
	mod          bool
	treatasclean bool
	isdir        bool
}

// NewBytes returns a new UTF-8 Bytes with the provided contents.
func NewBytes(contents []byte) *Bytes {
	return new(Bytes).Init(contents)
}

// Init initializes an existing Bytes to hold the provided contents.
// It returns a pointer to the initialized Bytes.
func (b *Bytes) Init(contents []byte) *Bytes {
	b.buf = undo.NewBuffer(contents)
	b.bytePos = 0
	b.runePos = 0
	for i := 0; i < len(contents); i++ {
		if contents[i] >= utf8.RuneSelf {
			// Not ASCII.
			b.numRunes = utf8.RuneCount(contents)
			_, b.width = utf8.DecodeRune(contents)
			b.nonASCII = i
			return b
		}
	}
	// ASCII is simple.  Also, the empty string is ASCII.
	b.numRunes = len(contents)
	b.width = 0
	b.nonASCII = len(contents)
	return b
}

// Bytes returns the contents of the Bytes.  This method also means the
// Bytes is directly printable by fmt.Print.
func (b *Bytes) Bytes() []byte {
	return b.buf.Bytes()
}

// Nr returns the number of runes (Unicode code points) in the Bytes.
func (b *Bytes) Nr() int {
	return b.numRunes
}

// IsASCII returns a boolean indicating whether the Bytes contains only ASCII bytes.
func (b *Bytes) IsASCII() bool {
	return b.width == 0
}

// Slice returns the string sliced at rune positions [i:j].
func (b *Bytes) Slice(i, j int) []byte {
	// ASCII is easy.  Let the compiler catch the indexing error if there is one.
	if j < b.nonASCII {
		return b.Bytes()[i:j]
	}
	if i < 0 || j > b.numRunes || i > j {
		panic(errSliceOutOfRange)
	}
	if i == j {
		return []byte("")
	}
	// For non-ASCII, after At(i), bytePos is always the position of the indexed character.
	var low, high int
	switch {
	case i < b.nonASCII:
		low = i
	case i == b.numRunes:
		low = b.Nb()
	default:
		b.At(i)
		low = b.bytePos
	}
	switch {
	case j == b.numRunes:
		high = b.Nb()
	default:
		b.At(j)
		high = b.bytePos
	}
	return b.Bytes()[low:high]
}

// At returns the rune with index i in the Bytes.  The sequence of runes is the same
// as iterating over the contents with a "for range" clause.
func (b *Bytes) At(i int) rune {
	// ASCII is easy.  Let the compiler catch the indexing error if there is one.
	if i < b.nonASCII {
		return rune(b.Bytes()[i])
	}

	// Now we do need to know the index is valid.
	if i < 0 || i >= b.numRunes {
		panic(errOutOfRange)
	}

	var r rune

	// Five easy common cases: within 1 spot of bytePos/runePos, or the beginning, or the end.
	// With these cases, all scans from beginning or end work in O(1) time per rune.
	switch {

	case i == b.runePos-1: // backing up one rune
		r, b.width = utf8.DecodeLastRune(b.Bytes()[0:b.bytePos])
		b.runePos = i
		b.bytePos -= b.width
		return r
	case i == b.runePos+1: // moving ahead one rune
		b.runePos = i
		b.bytePos += b.width
		fallthrough
	case i == b.runePos:
		r, b.width = utf8.DecodeRune(b.Bytes()[b.bytePos:])
		return r
	case i == 0: // start of string
		r, b.width = utf8.DecodeRune(b.Bytes())
		b.runePos = 0
		b.bytePos = 0
		return r

	case i == b.numRunes-1: // last rune in string
		r, b.width = utf8.DecodeLastRune(b.buf.Bytes())
		b.runePos = i
		b.bytePos = b.Nb() - b.width
		return r
	}

	// We need to do a linear scan.  There are three places to start from:
	// 1) The beginning
	// 2) bytePos/runePos.
	// 3) The end
	// Choose the closest in rune count, scanning backwards if necessary.
	forward := true
	if i < b.runePos {
		// Between beginning and pos.  Which is closer?
		// Since both i and runePos are guaranteed >= nonASCII, that'bytes the
		// lowest location we need to start from.
		if i < (b.runePos-b.nonASCII)/2 {
			// Scan forward from beginning
			b.bytePos, b.runePos = b.nonASCII, b.nonASCII
		} else {
			// Scan backwards from where we are
			forward = false
		}
	} else {
		// Between pos and end.  Which is closer?
		if i-b.runePos < (b.numRunes-b.runePos)/2 {
			// Scan forward from pos
		} else {
			// Scan backwards from end
			b.bytePos, b.runePos = b.Nb(), b.numRunes
			forward = false
		}
	}
	if forward {
		// TODO: Is it much faster to use a range loop for this scan?
		for {
			r, b.width = utf8.DecodeRune(b.Bytes()[b.bytePos:])
			if b.runePos == i {
				break
			}
			b.runePos++
			b.bytePos += b.width
		}
	} else {
		for {
			r, b.width = utf8.DecodeLastRune(b.Bytes()[0:b.bytePos])
			b.runePos--
			b.bytePos -= b.width
			if b.runePos == i {
				break
			}
		}
	}
	return r
}

// HasNull returns true if Bytes contains a null rune.
func (b *Bytes) HasNull() bool {
	for i := 0; i < b.numRunes; i++ {
		if b.At(i) == 0 {
			return true
		}
	}
	return false
}

// Read implements the io.Reader interface.
func (b *Bytes) Read(buf []byte) (n int, err error) {
	n = copy(buf, b.Bytes())
	return n, nil
}

// Nb returns the size of the buffer in bytes.
func (b *Bytes) Nb() int {
	return int(b.buf.Size())
}

// Clean is a forwarding function for undo.Clean.
func (b *Bytes) Clean() {
	b.mod = false
	b.treatasclean = false
	b.buf.Clean()
}

var errOutOfRange = errors.New("utf8Bytes: index out of range")
var errSliceOutOfRange = errors.New("utf8Bytes: slice index out of range")
