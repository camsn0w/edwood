package file

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"unicode/utf8"

	"github.com/rjkroege/edwood/internal/sam"
	"github.com/rjkroege/edwood/internal/undo"
	"github.com/rjkroege/edwood/internal/util"
)

// The ObservableEditableBuffer is used by the main program
// to add, remove and check on the current observer(s) for a Text.
// Text in turn, implements BufferObserver for the various required callback functions in BufferObserver.
type ObservableEditableBuffer struct {
	currobserver BufferObserver
	observers    map[BufferObserver]struct{} // [private I think]
	Elog         sam.Elog
	// TODO(rjk): Remove this when I've inserted undo.RuneArray.
	// At present, InsertAt and DeleteAt have an implicit Commit operation
	// associated with them. In an undo.RuneArray context, these two ops
	// don't have an implicit Commit. We set editclean in the Edit cmd
	// implementation code to let multiple Inserts be grouped together?
	// Figure out how this inter-operates with seq.
	EditClean bool
	details   *DiskDetails
	isscratch bool // Used to track if this File should warn on unsaved deletion. [private]
	rbi       *Bytes
	undo      *undo.Buffer
}

// Set is a forwarding function for file_hash.Set
func (e *ObservableEditableBuffer) Set(hash []byte) {
	e.details.Hash.Set(hash)
}

func (e *ObservableEditableBuffer) SetInfo(info os.FileInfo) {
	e.details.Info = info
}

// AddObserver adds e as an observer for edits to this File.
func (e *ObservableEditableBuffer) AddObserver(observer BufferObserver) {
	if e.observers == nil {
		e.observers = make(map[BufferObserver]struct{})
	}
	e.observers[observer] = struct{}{}
	e.currobserver = observer
}

// DelObserver removes e as an observer for edits to this File.
func (e *ObservableEditableBuffer) DelObserver(observer BufferObserver) error {
	if _, exists := e.observers[observer]; exists {
		delete(e.observers, observer)
		if observer == e.currobserver {
			for k := range e.observers {
				e.currobserver = k
				break
			}
		}
		return nil
	}
	return fmt.Errorf("can't find editor in File.DelObserver")
}

// SetCurObserver sets the current observer.
func (e *ObservableEditableBuffer) SetCurObserver(observer BufferObserver) {
	e.currobserver = observer
}

// GetCurObserver gets the current observer and returns it as a interface type.
func (e *ObservableEditableBuffer) GetCurObserver() interface{} {
	return e.currobserver
}

// AllObservers preforms tf(all observers...).
func (e *ObservableEditableBuffer) AllObservers(tf func(i interface{})) {
	for t := range e.observers {
		tf(t)
	}
}

// GetObserverSize will return the size of the observer map.
func (e *ObservableEditableBuffer) GetObserverSize() int {
	return len(e.observers)
}

// HasMultipleObservers returns true if their are multiple observers to the File.
func (e *ObservableEditableBuffer) HasMultipleObservers() bool {
	return len(e.observers) > 1
}

// MakeObservableEditableBuffer is a constructor wrapper for NewFile() to abstract File from the main program.
func MakeObservableEditableBuffer(filename string, b RuneArray) *ObservableEditableBuffer {
	data := []byte(b.String())

	oeb := &ObservableEditableBuffer{
		currobserver: nil,
		observers:    nil,
		details:      &DiskDetails{Name: filename, Hash: Hash{}},
		Elog:         sam.MakeElog(),
		EditClean:    true,
		undo:         undo.NewBuffer(data, len(b)),
		rbi:          NewBytes(data),
	}
	oeb.rbi.oeb = oeb
	return oeb
}

// MakeObservableEditableBufferTag is a constructor wrapper for NewTagFile() to abstract File from the main program.
func MakeObservableEditableBufferTag(b RuneArray) *ObservableEditableBuffer {
	data := []byte(b.String())

	oeb := &ObservableEditableBuffer{
		currobserver: nil,
		observers:    nil,
		Elog:         sam.MakeElog(),
		details:      &DiskDetails{Hash: Hash{}},
		EditClean:    true,
		undo:         undo.NewBuffer(data, len(b)),
		rbi:          NewBytes(data),
	}
	oeb.rbi.oeb = oeb
	return oeb
}

// Clean is a forwarding function for undo.Clean.
func (e *ObservableEditableBuffer) Clean() {
	e.undo.Clean()
}

// Size is a forwarding function for file.Size.
func (e *ObservableEditableBuffer) Size() int {
	return e.rbi.RuneCount()
}

// Mark is a forwarding function for file.Mark.
func (e *ObservableEditableBuffer) Mark(seq int) {
	e.undo.Commit()
}

// Reset removes all Undo records for this File.
func (e *ObservableEditableBuffer) Reset() {
	e.undo = undo.NewBuffer(e.Bytes(), e.Nr())
}

// HasUncommitedChanges is a forwarding function for undo.HasUncommitedChanges
func (e *ObservableEditableBuffer) HasUncommitedChanges() bool {
	// TODO(sn0w): This will need to be changed along with Commit as the idea for Commit is changing.
	return e.undo.HasUncommitedChanges()
}

// HasRedoableChanges is a forwarding function for file.HasRedoableChanges.
func (e *ObservableEditableBuffer) HasRedoableChanges() bool {
	return e.undo.HasRedoableChanges()
}

// HasUndoableChanges is a forwarding function for file.HasUndoableChanges
func (e ObservableEditableBuffer) HasUndoableChanges() bool {
	return e.undo.HasUndoableChanges()
}

// IsDir is a forwarding function for DiskDetails.IsDir.
func (e *ObservableEditableBuffer) IsDir() bool {
	return e.details.IsDir()
}

// SetDir is a forwarding function for DiskDetails.SetDir.
func (e *ObservableEditableBuffer) SetDir(flag bool) {
	e.details.SetDir(flag)
}

// Nr is a forwarding function for bytes.RuneCount.
func (e *ObservableEditableBuffer) Nr() int {
	return e.rbi.RuneCount()
}

// ReadC is a forwarding function for bytes.At.
func (e *ObservableEditableBuffer) ReadC(q int) rune {
	return e.rbi.At(q)
}

// SaveableAndDirty returns true if the File's contents differ from the
// backing diskfile File.name, and the diskfile is plausibly writable
// (not a directory or scratch file).
//
// When this is true, the tag's button should
// be drawn in the modified state if appropriate to the window type
// and Edit commands should treat the file as modified.
//
// TODO(rjk): figure out how this overlaps with hash. (hash would appear
// to be used to determine the "if the contents differ")
//
// Latest thought: there are two separate issues: are we at a point marked
// as clean and is this File writable to a backing. They are combined in this
// this method.
func (e *ObservableEditableBuffer) SaveableAndDirty() bool {
	return e.details.Name != "" && (e.undo.Mod() || e.undo.Dirty() || len(e.undo.GetCache()) > 0) && !e.IsDirOrScratch()
}

// Load is a forwarding function for file.Load.
func (e *ObservableEditableBuffer) Load(q0 int, fd io.Reader, sethash bool) (n int, hasNulls bool, err error) {
	d, err := ioutil.ReadAll(fd)
	if err != nil {
		err = errors.New("read error in file.Load")
	}
	if sethash {
		e.SetHash(CalcHash(d))
	}

	runes, _, hasNulls := util.Cvttorunes(d, len(d))

	// Would appear to require a commit operation.
	// NB: Runs the observers.
	e.InsertAt(q0, runes)
	e.inserted(q0, runes)
	e.Reset()

	e.undo.Modded()
	return len(d), hasNulls, err
}

// Dirty is a forwarding function for file.Dirty.
func (e *ObservableEditableBuffer) Dirty() bool {
	return e.undo.Dirty()
}

// InsertAt is a forwarding function for file.InsertAt.
func (e *ObservableEditableBuffer) InsertAt(p0 int, s []rune) {
	e.InsertAtWithoutCommit(p0, s)
}

// SetName sets the name of the backing for this file.
// Some backings that opt them out of typically being persisted.
// Resetting a file name to a new value does not have any effect.
func (e *ObservableEditableBuffer) SetName(name string) {
	if e.Name() == name {
		return
	}

	e.Setnameandisscratch(name)
}

// Undo is a forwarding function for file.Undo.
func (e *ObservableEditableBuffer) Undo(isundo bool) (q0, q1 int, ok bool) {
	e.undo.MarkUnclean()

	var info undo.ChangeInfo
	if isundo {
		info = e.undo.Undo()
		e.rbi.numRunes += info.Nr
		if info.Off == -1 {
			return
		}
	} else {
		info = e.undo.Redo()
		e.rbi.numRunes -= info.Nr
		if info.Off == -1 {
			return
		}
	}
	if info == (undo.ChangeInfo{}) {
		return 0, 0, false
	}
	if info.NonAscii == -1 {
		if info.Off+int64(info.Nr) < int64(e.rbi.nonASCII) {
			e.rbi.nonASCII += int(info.Off) + info.Nr
		}
	} else {
		if info.Off < int64(e.rbi.nonASCII) {
			e.rbi.nonASCII = info.NonAscii + int(info.Off)
			e.rbi.width = info.Width
		}
	}

	return
}

// DeleteAt is a forwarding function for file.DeleteAt.
func (e *ObservableEditableBuffer) DeleteAt(q0, q1 int) {
	e.rbi.At(q0)
	b0 := e.rbi.bytePos
	e.rbi.At(q1)
	b1 := e.rbi.bytePos
	e.undo.Delete(int64(b0), int64(b1-b0))
	n := q1 - q0
	e.rbi.numRunes -= n

	if q1 < e.rbi.nonASCII {
		e.rbi.nonASCII -= n
	} else {
		nonASCII, width := e.undo.FindNewNonAscii()
		if nonASCII != -1 {
			e.rbi.nonASCII = nonASCII
			e.rbi.width = width
		}
	}

	e.deleted(b0, b1)
}

// TreatAsClean is a forwarding function for undo.TreatAsClean.
func (e *ObservableEditableBuffer) TreatAsClean() {
	e.undo.TreatAsClean()
}

// Modded is a forwarding function for undo.Modded.
func (e *ObservableEditableBuffer) Modded() {
	e.undo.Modded()
}

// Name is a getter for file.details.Name.
func (e *ObservableEditableBuffer) Name() string {
	return e.details.Name
}

// Info is a Getter for e.details.Info
func (e *ObservableEditableBuffer) Info() os.FileInfo {
	return e.details.Info
}

// UpdateInfo is a forwarding function for file.UpdateInfo
func (e *ObservableEditableBuffer) UpdateInfo(filename string, d os.FileInfo) error {
	return e.details.UpdateInfo(filename, d)
}

// Hash is a getter for DiskDetails.Hash
func (e *ObservableEditableBuffer) Hash() Hash {
	return e.details.Hash
}

// SetHash is a setter for DiskDetails.Hash
func (e *ObservableEditableBuffer) SetHash(hash Hash) {
	e.details.Hash = hash
}

// Seq is a getter for file.details.Seq.
func (e *ObservableEditableBuffer) Seq() int {
	// TODO(sn0w): Remove this.
	return 0
}

// RedoSeq is a getter for file.details.RedoSeq.
func (e *ObservableEditableBuffer) RedoSeq() int {
	// TODO(sn0w): Remove this.
	return 0
}

// inserted is a forwarding function for text.inserted.
func (e *ObservableEditableBuffer) inserted(q0 int, r []rune) {
	for observer := range e.observers {
		observer.Inserted(q0, r)
	}
}

// deleted is a forwarding function for text.deleted.
func (e *ObservableEditableBuffer) deleted(q0 int, q1 int) {
	for observer := range e.observers {
		observer.Deleted(q0, q1)
	}
}

// Commit is a forwarding function for file.Commit.
func (e *ObservableEditableBuffer) Commit() {
	// TODO(sn0w): This will need to be updated to be the global commit for changes.
}

// InsertAtWithoutCommit is a forwarding function for file.InsertAtWithoutCommit.
func (e *ObservableEditableBuffer) InsertAtWithoutCommit(p0 int, s []rune) {
	origpos := 0
	if p0 < e.rbi.numRunes {
		e.rbi.At(p0)
		origpos = e.rbi.bytePos
	} else {
		if e.rbi.numRunes > 0 {
			c := e.rbi.At(e.rbi.numRunes - 1)
			origpos = e.rbi.bytePos + utf8.RuneLen(c)
		}
	}

	e.rbi.numRunes += len(s)

	ob := make([]byte, 0, utf8.UTFMax*len(s))
	tb := 0
	for _, r := range s {
		b := make([]byte, utf8.UTFMax)
		nb := utf8.EncodeRune(b, r)
		e.rbi.bytePos += nb
		e.rbi.runePos++
		if nb > 1 && e.rbi.nonASCII > e.rbi.bytePos {
			e.rbi.nonASCII = e.rbi.bytePos
		}
		if nb > 1 {
			e.rbi.width = nb
		}
		ob = append(ob, b[:nb]...)
		tb += nb
	}
	e.undo.Insert(int64(origpos), ob[:tb], len(s))
	e.inserted(origpos, s)
}

// IsDirOrScratch returns true if the File has a synthetic backing of
// a directory listing or has a name pattern that excludes it from
// being saved under typical circumstances.
func (e *ObservableEditableBuffer) IsDirOrScratch() bool {
	return e.isscratch || e.IsDir()
}

// TreatAsDirty is a forwarding function for file.TreatAsDirty.
func (e *ObservableEditableBuffer) TreatAsDirty() bool {
	return e.undo.TreatAsDirty()
}

// Read is a forwarding function for rune_array.Read.
func (e *ObservableEditableBuffer) Read(q0 int, r []rune) (int, error) {
	if cap(r) > e.rbi.RuneCount() {
		r = r[:e.rbi.RuneCount()]

	}
	data := e.View(q0, len(r))
	copy(r, data)
	return len(r), nil
}

// View is a forwarding function for rune_array.View.
func (e *ObservableEditableBuffer) View(q0 int, q1 int) []rune {
	return []rune(string(e.rbi.Slice(q0, q1)))
}

// String is a forwarding function for rune_array.String.
func (e *ObservableEditableBuffer) String() string {
	return string(e.Bytes())
}

// ResetBuffer is a forwarding function for rune_array.Reset.
func (e *ObservableEditableBuffer) ResetBuffer() {
	e.undo = undo.NewBuffer(e.Bytes(), e.Nr())
}

// Reader is a forwarding function for rune_array.Reader.
func (e *ObservableEditableBuffer) Reader(q0 int, q1 int) io.Reader {
	return e.rbi.Reader(q0, q1)
}

// IndexRune is a forwarding function for rune_array.IndexRune.
func (e *ObservableEditableBuffer) IndexRune(r rune) int {
	return e.rbi.IndexRune(r)
}

// Nbyte is a forwarding function for rune_array.Nbyte.
func (e *ObservableEditableBuffer) Nbyte() int {
	return int(e.undo.Size())
}

// Setnameandisscratch updates the oeb.details.name and isscratch bit
// at the same time.
func (e *ObservableEditableBuffer) Setnameandisscratch(name string) {
	e.details.Name = name
	if strings.HasSuffix(name, slashguide) || strings.HasSuffix(name, plusErrors) {
		e.isscratch = true
	} else {
		e.isscratch = false
	}
}

// GetCache is a Getter for file.cache for use in tests.
func (e *ObservableEditableBuffer) GetCache() []byte {
	return e.undo.GetCache()
}

// Bytes returns the contents of the Undo.buffer as a byte slice.
func (e *ObservableEditableBuffer) Bytes() []byte {
	return e.undo.Bytes()
}
