package file

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"testing"
	"unicode/utf8"
)

func TestOverall(t *testing.T) {
	b := NewBufferNoNr(nil)
	b.checkPiecesCnt(t, 2)
	b.checkContent("#0", t, "")

	b.insertString(0, "")
	b.checkPiecesCnt(t, 2)
	b.checkContent("#1", t, "")

	b.insertString(0, "All work makes John a dull boy")
	b.checkPiecesCnt(t, 3)
	b.checkContent("#2", t, "All work makes John a dull boy")

	b.insertString(9, "and no playing ")
	b.checkPiecesCnt(t, 6)
	b.checkContent("#3", t, "All work and no playing makes John a dull boy")

	b.Commit()
	// Also check that multiple change commits don't create empty changes.
	b.Commit()
	b.deleteCreateOffsetTuple(20, 34)
	b.checkContent("#4", t, "All work and no play a dull boy")

	b.insertString(20, " makes Jack")
	b.checkContent("#5", t, "All work and no play makes Jack a dull boy")

	b.Undo()
	b.checkContent("#6", t, "All work and no play a dull boy")
	b.Undo()
	b.checkContent("#7", t, "All work and no playing makes John a dull boy")
	b.Undo()
	b.checkContent("#8", t, "All work makes John a dull boy")

	b.Redo()
	b.checkContent("#9", t, "All work and no playing makes John a dull boy")
	b.Redo()
	b.checkContent("#10", t, "All work and no play a dull boy")
	b.Redo()
	b.checkContent("#11", t, "All work and no play makes Jack a dull boy")
	b.Redo()
	b.checkContent("#12", t, "All work and no play makes Jack a dull boy")
}

func TestCacheInsertAndDelete(t *testing.T) {
	b := NewBufferNoNr([]byte("testing insertation"))
	b.checkPiecesCnt(t, 3)
	b.checkContent("#0", t, "testing insertation")

	b.cacheInsertString(8, "caching")
	b.checkPiecesCnt(t, 6)
	b.checkContent("#1", t, "testing cachinginsertation")

	b.cacheInsertString(15, " ")
	b.checkPiecesCnt(t, 6)
	b.checkContent("#2", t, "testing caching insertation")

	b.cacheDelete(12, 3)
	b.checkPiecesCnt(t, 6)
	b.checkContent("#3", t, "testing cach insertation")

	b.cacheInsertString(12, "ed")
	b.checkPiecesCnt(t, 6)
	b.checkContent("#4", t, "testing cached insertation")
}

func TestSimulateBackspace(t *testing.T) {
	b := NewBufferNoNr([]byte("apples and oranges"))
	for i := 5; i > 0; i-- {
		b.cacheDelete(i, 1)
	}
	b.checkContent("#0", t, "a and oranges")
	b.Undo()
	b.checkContent("#1", t, "apples and oranges")
}

func TestSimulateDeleteKey(t *testing.T) {
	b := NewBufferNoNr([]byte("apples and oranges"))
	for i := 0; i < 4; i++ {
		b.cacheDelete(7, 1)
	}
	b.checkContent("#0", t, "apples oranges")
	b.Undo()
	b.checkContent("#1", t, "apples and oranges")
}

func TestDelete(t *testing.T) {
	b := NewBufferNoNr([]byte("and what is a dream?"))
	b.insertString(9, "exactly ")
	b.checkContent("#0", t, "and what exactly is a dream?")

	b.delete(22, 2000)
	b.checkContent("#1", t, "and what exactly is a ")
	b.insertString(22, "joke?")
	b.checkContent("#2", t, "and what exactly is a joke?")

	cases := []struct {
		off, len int
		expected string
	}{
		{9, 8, "and what is a joke?"},
		{9, 13, "and what joke?"},
		{5, 6, "and wactly is a joke?"},
		{9, 14, "and what oke?"},
		{11, 3, "and what exly is a joke?"},
	}
	for _, c := range cases {
		b.delete(c.off, c.len)
		b.checkContent("#3", t, c.expected)
		b.Undo()
		b.checkContent("#4", t, "and what exactly is a joke?")
	}
}

func TestDeleteAtTheEndOfCachedPiece(t *testing.T) {
	b := NewBufferNoNr([]byte("Original data."))
	b.cacheInsertString(8, ",")
	b.cacheDelete(9, 1)
	b.checkContent("#0", t, "Original,data.")
	b.Undo()
	b.checkContent("#1", t, "Original data.")
}

func TestGroupChanges(t *testing.T) {
	b := NewBufferNoNr([]byte("group 1, group 2, group 3"))
	b.checkPiecesCnt(t, 3)
	// b.GroupChanges()

	b.cacheDelete(0, 6)
	b.checkContent("#0", t, "1, group 2, group 3")

	b.cacheDelete(3, 6)
	b.checkContent("#1", t, "1, 2, group 3")

	b.cacheDelete(6, 6)
	b.checkContent("#2", t, "1, 2, 3")

	b.Undo()
	b.checkContent("#3", t, "group 1, group 2, group 3")
	b.Undo()
	b.checkContent("#4", t, "group 1, group 2, group 3")

	b.Redo()
	b.checkContent("#5", t, "1, 2, 3")
	b.Redo()
	b.checkContent("#6", t, "1, 2, 3")
}

func TestSaving(t *testing.T) {
	b := NewBufferNoNr(nil)

	b.checkModified(t, 1, false)
	b.insertString(0, "stars can frighten")
	b.checkModified(t, 2, true)

	b.Clean()
	b.checkModified(t, 3, false)

	b.Undo()
	b.checkModified(t, 4, true)
	b.Redo()
	b.checkModified(t, 5, false)

	b.insertString(0, "Neptun, Titan, ")
	b.checkModified(t, 6, true)
	b.Undo()
	b.checkModified(t, 7, false)

	b.Redo()
	b.checkModified(t, 8, true)

	b.Clean()
	b.checkModified(t, 9, false)

	b = NewBufferNoNr([]byte("my book is closed"))
	b.checkModified(t, 10, false)

	b.insertString(17, ", I read no more")
	b.checkModified(t, 11, true)
	b.Undo()
	b.checkModified(t, 12, false)

	b.Redo()
	b.Clean()
	b.checkModified(t, 13, false)

	b.Undo()
	b.Clean()
	b.checkModified(t, 14, false)
}

func TestReader(t *testing.T) {
	b := NewBufferNoNr(nil)
	b.insertString(0, "So many")
	b.insertString(7, " books,")
	b.insertString(14, " so little")
	b.insertString(24, " time.")
	b.checkContent("#0", t, "So many books, so little time.")

	cases := []struct {
		off, len int
		expected string
		err      error
	}{
		{0, 7, "So many", nil},
		{1, 11, "o many book", nil},
		{8, 4, "book", nil},
		{15, 20, "so little time.", io.EOF},
	}

	for _, c := range cases {
		data := make([]byte, c.len)
		n, err := b.ReadAt(data, int64(c.off))
		if err != c.err {
			t.Errorf("expected error %v, got %v", c.err, err)
		}
		if n != len(c.expected) {
			t.Errorf("n should be %d, got %d", len(c.expected), n)
		}
		if !bytes.Equal(data[:n], []byte(c.expected)) {
			t.Errorf("got '%s', want '%s'", data[:n], c.expected)
		}
	}
}

func TestBufferSize(t *testing.T) {
	b := NewBufferNoNr(nil)
	tests := []struct {
		action func()
		want   int64
	}{
		0: {func() {}, 0},
		1: {func() { b.insertString(0, " Like") }, 5},
		2: {func() { b.insertString(0, " Colour") }, 12},
		3: {func() { b.insertString(7, " You") }, 16},
		4: {func() { b.delete(5, 1) }, 15},
		5: {func() { b.insertString(0, "Pink is the") }, 26},
		6: {func() { b.Undo() }, 15},
		7: {func() { b.Redo() }, 26},
	}

	for i, tt := range tests {
		tt.action()
		if got := b.Size(); got != tt.want {
			t.Fatalf("%d: got %d, want %d", i, got, tt.want)
		}
	}
}

func TestUndoRedoReturnedOffsets(t *testing.T) {
	b := NewBufferNoNr(nil)
	insert := func(off, len int) {
		b.insertString(off, strings.Repeat(".", len))
	}
	insert(0, 7)
	insert(7, 5)
	insert(12, 9)
	b.delete(8, 8)
	insert(3, 19)
	b.delete(0, 20)

	undo, redo := (*Buffer).Undo, (*Buffer).Redo
	tests := []struct {
		op      func(*Buffer) (int64, int64)
		wantOff int64
		wantN   int64
	}{
		0:  {redo, -1, 0},
		1:  {undo, 0, 20},
		2:  {undo, 3, -19},
		3:  {undo, 8, 8},
		4:  {undo, 12, -9},
		5:  {undo, 7, -5},
		6:  {undo, 0, -7},
		7:  {undo, -1, 0},
		8:  {redo, 0, 7},
		9:  {redo, 7, 5},
		10: {redo, 12, 9},
		11: {redo, 8, -8},
		12: {redo, 3, 19},
		13: {redo, 0, -20},
		14: {redo, -1, 0},
	}

	for i, tt := range tests {
		off, n := tt.op(b)
		if off != tt.wantOff {
			t.Errorf("%d: got offset %d, want %d", i, off, tt.wantOff)
		}
		if n != tt.wantN {
			t.Errorf("%d: got n %d, want %d", i, n, tt.wantN)
		}
	}
}

func TestPieceNr(t *testing.T) {
	b := NewBufferNoNr(nil)
	manderianBytes := []byte("痛苦本身可能是很多痛苦, 但主要的原因是痛苦, 但我给它时间陷入这种痛苦, 以至于有些巨大的痛苦")
	eng1 := []byte("Lorem ipsum in Mandarin")
	eng2 := []byte("This is the")
	eng3 := []byte("In the midst")

	b.insertCreateOffsetTuple(0, manderianBytes)
	b.checkContent("TestPieceNr: First insert", t, string(manderianBytes))

	b.insertCreateOffsetTuple(b.Nr(), eng1)
	b.checkContent("TestPieceNr: Second insert", t, string(manderianBytes)+string(eng1))

	b.insertCreateOffsetTuple(0, eng2)
	buffAfterInserts := string(eng2) + string(manderianBytes) + string(eng1)
	b.checkContent("TestPieceNr: third insert", t, buffAfterInserts)

	b.deleteCreateOffsetTuple(13, 10)
	buffAfterDelete := []rune(buffAfterInserts)
	buffAfterDelete = append(buffAfterDelete[:13], buffAfterDelete[41:]...)
	b.checkContent("TestPieceNr: after 1 delete", t, string(buffAfterDelete))

	b.insertCreateOffsetTuple(8, eng3)
	buffAfterDelete = append(buffAfterDelete[:8], append([]rune(string(eng3)), buffAfterDelete[8:]...)...)
	b.checkContent("TestPieceNr: after everything", t, string(buffAfterDelete))

	undo, redo := (*Buffer).Undo, (*Buffer).Redo
	tests := []struct {
		op func(*Buffer) (int64, int64)
	}{
		0:  {redo},
		1:  {undo},
		2:  {undo},
		3:  {undo},
		4:  {undo},
		5:  {undo},
		6:  {undo},
		7:  {undo},
		8:  {redo},
		9:  {redo},
		10: {redo},
		11: {redo},
		12: {redo},
		13: {redo},
		14: {redo},
	}

	for i, tt := range tests {
		tt.op(b)
		nr := b.Nr()
		wantNr := countRunes(b)
		if nr != wantNr {
			t.Errorf("%d: got n %d, want %d", i, nr, wantNr)
		}
	}
}

func (b *Buffer) checkPiecesCnt(t *testing.T, expected int) {
	if b.piecesCnt != expected {
		t.Errorf("got %d pieces, want %d", b.piecesCnt, expected)
	}
}

func (b *Buffer) checkContent(name string, t *testing.T, expected string) {
	c := b.allContent()
	if c != expected {
		t.Errorf("%s: got '%s', want '%s'", name, c, expected)
	}
}

func (t *Buffer) insertString(off int, data string) {
	t.Commit()
	t.cacheInsertString(off, data)
}

func (t *Buffer) cacheInsertString(off int, data string) {
	err := t.insertCreateOffsetTuple(int64(off), []byte(data))
	if err != nil {
		panic(err)
	}
}

func (t *Buffer) delete(off, length int) {
	t.Commit()
	t.cacheDelete(off, length)
}

func (t *Buffer) cacheDelete(off, length int) {
	t.deleteCreateOffsetTuple(int64(off), int64(length))
}

func (t *Buffer) printPieces() {
	for p := t.begin; p != nil; p = p.next {
		prev, next := 0, 0
		if p.prev != nil {
			prev = p.prev.id
		}
		if p.next != nil {
			next = p.next.id
		}
		fmt.Printf("%d, p:%d, n:%d = %s\n", p.id, prev, next, string(p.data))
	}
	fmt.Println()
}

func TestRuneTuple(t *testing.T) {
	tt := []struct {
		name  string
		buf   string
		nr    int
		roff  int
		bwant int
	}{
		{
			name:  "one buf, start",
			buf:   "foo",
			nr:    len("foo"),
			roff:  0,
			bwant: 0,
		},
		{
			name:  "one buf, middle",
			buf:   "foo",
			nr:    len("foo"),
			roff:  1,
			bwant: 1,
		},
		{
			name:  "one buf, end",
			buf:   "foo",
			nr:    len("foo"),
			roff:  2,
			bwant: 2,
		},
	}
	for _, tv := range tt {
		t.Run(tv.name, func(t *testing.T) {
			b := NewBuffer([]byte(tv.buf), tv.nr)
			gt := b.RuneTuple(int64(tv.roff))
			if got, want := gt.b, tv.bwant; got != int64(want) {
				t.Errorf("%s got %d != want %d", "byte", got, want)
			}
			if got, want := gt.r, tv.roff; got != int64(want) {
				t.Errorf("%s got %d != want %d", "rune", got, want)
			}
		})
	}
}

func (b *Buffer) checkModified(t *testing.T, id int, expected bool) {
	if b.Dirty() != expected {
		if expected {
			t.Errorf("#%d should be modified", id)
		} else {
			t.Errorf("#%d should not be modified", id)
		}
	}
}

func (t *Buffer) allContent() string {
	var data []byte
	p := t.begin.next
	for p != t.end {
		data = append(data, p.data...)
		p = p.next

	}
	return string(data)
}

func countRunes(b *Buffer) int64 {
	return int64(utf8.RuneCount(b.Bytes()))
}

func NewBufferNoNr(content []byte) *Buffer {
	return NewBuffer(content, utf8.RuneCount(content))
}

func (b *Buffer) insertCreateOffsetTuple(off int64, content []byte) error {
	return b.Insert(b.RuneTuple(off), content)
}

func (b *Buffer) deleteCreateOffsetTuple(off, length int64) error {
	return b.Delete(b.RuneTuple(off), b.RuneTuple(off+length))
}
