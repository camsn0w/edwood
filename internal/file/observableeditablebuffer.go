package main

import (
	"fmt"
	"io"
)

// The ObservableEditableBuffer is used by the main program
// to add, remove and check on the current observer(s) for a Text.
// Text in turn, implements BufferObserver for the various required callback functions in BufferObserver.
type ObservableEditableBuffer struct {
	currobserver BufferObserver
	observers    map[BufferObserver]struct{} // [private I think]
	f            *File
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

// AllObservers preforms tf(all observers...)
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
	f := NewFile(filename)
	f.b = b
	oeb := &ObservableEditableBuffer{
		currobserver: nil,
		observers:    nil,
		f:            f,
	}
	oeb.f.oeb = oeb
	return oeb
}

// MakeObservableEditableBufferTag is a constructor wrapper for NewTagFile() to abstract File from the main program.
func MakeObservableEditableBufferTag(b RuneArray) *ObservableEditableBuffer {
	f := NewTagFile()
	f.b = b
	oeb := &ObservableEditableBuffer{
		currobserver: nil,
		observers:    nil,
		f:            f,
	}
	oeb.f.oeb = oeb
	return oeb
}

// Clean is a forwarding function for file.Clean.
func (e *ObservableEditableBuffer) Clean() {
	e.f.Clean()
}

// FileSize is a forwarding function for file.Size.
func (e *ObservableEditableBuffer) FileSize() int {
	return e.f.Size()
}

// Mark is a forwarding function for file.Mark.
func (e *ObservableEditableBuffer) Mark(seq int) {
	e.f.Mark(seq)
}

// Reset is a forwarding function for file.Reset.
func (e *ObservableEditableBuffer) Reset() {
	e.f.Reset()
}

// HasUncommitedChanges is a forwarding function for file.HasUncommitedChanges.
func (e *ObservableEditableBuffer) HasUncommitedChanges() bool {
	return e.f.HasUncommitedChanges()
}

// IsDir is a forwarding function for file.IsDir.
func (e *ObservableEditableBuffer) IsDir() bool {
	return e.f.IsDir()
}

// SetDir is a forwarding function for file.SetDir.
func (e *ObservableEditableBuffer) SetDir(flag bool) {
	e.f.SetDir(flag)
}

// Nr is a forwarding function for file.Nr.
func (e *ObservableEditableBuffer) Nr() int {
	return e.f.Nr()
}

// ReadC is a forwarding function for file.ReadC.
func (e *ObservableEditableBuffer) ReadC(q int) rune {
	return e.f.ReadC(q)
}

// SaveableAndDirty is a forwarding function for file.SaveableAndDirty.
func (e *ObservableEditableBuffer) SaveableAndDirty() bool {
	return e.f.SaveableAndDirty()
}

// Load is a forwarding function for file.Load()
func (e *ObservableEditableBuffer) Load(q0 int, fd io.Reader, sethash bool) (n int, hasNulls bool, err error) {
	return e.f.Load(q0, fd, sethash)
}

// Dirty is a forwarding function for file.Dirty.
func (e *ObservableEditableBuffer) Dirty() bool {
	return e.f.Dirty()
}

// InsertAt is a forwarding function for file.InsertAt.
func (e *ObservableEditableBuffer) InsertAt(p0 int, s []rune) {
	e.AllObservers(func(i interface{}) {
		i.(BufferObserver).inserted(p0, s)
	})
	e.f.InsertAt(p0, s)
}

// SetName is a forwarding function for file.SetName.
func (e *ObservableEditableBuffer) SetName(name string) {
	e.f.SetName(name)
}

// Undo is a forwarding function for file.Undo.
func (e *ObservableEditableBuffer) Undo(isundo bool) (q0, q1 int, ok bool) {
	return e.f.Undo(isundo)
}

// DeleteAt is a forwarding function for file.DeleteAt.
func (e *ObservableEditableBuffer) DeleteAt(q0, q1 int) {
	e.AllObservers(func(i interface{}) {
		i.(BufferObserver).deleted(q0, q1)
	})
	e.f.DeleteAt(q0, q1)
}

// TreatAsClean is a forwarding function for file.TreatAsClean.
func (e *ObservableEditableBuffer) TreatAsClean() {
	e.f.TreatAsClean()
}

// Modded is a forwarding function for file.Modded.
func (e *ObservableEditableBuffer) Modded() {
	e.f.Modded()
}

// Name is a getter for file.details.Name.
func (e *ObservableEditableBuffer) Name() string {
	return e.f.details.Name
}
