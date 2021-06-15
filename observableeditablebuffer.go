package main

import (
	"io"
	"os"
)

type ObservableEditableBuffer interface {
	AddText(observer BufferObserver)
	DelText(observer BufferObserver) error
	SetCurText(observer BufferObserver)
	GetCurText() interface{}
	GetTextSize() int
	AllText(tf func(i interface{}))
	HasMultipleTexts() bool
	InsertAt(p0 int, s []rune)
	InsertAtWithoutCommit(p0 int, s []rune)
	DeleteAt(p0, p1 int)
	Undo(isundo bool) (q0, q1 int, ok bool)
	New() *Editbuf
	Size() int
	IsDir() bool
	HasUncommitedChanges() bool
	SetDir(b bool)
	Load(q0 int, fd io.Reader, sethash bool) (n int, hasNulls bool, err error)
	Name() string
	ReadC(q int) rune
	Mark(s int)
	Commit()
	Reset()
	ResetRunes()
	Info() os.FileInfo
	SetInfo(info os.FileInfo)
	SetName(name string)
	Read(q0 int, r []rune) (int, error)
	View(q0 int, q1 int) []rune
	Clean()
	HasUndoableChanges() bool
	HasRedoableChanges() bool
	SaveableAndDirty() bool
	Modded()
	IsDirOrScratch() bool
	TreatAsDirty() bool
	TreatAsClean()
	Dirty() bool
}

type Editbuf struct {
	curtext BufferObserver
	text    map[BufferObserver]struct{} // [private I think]
}
