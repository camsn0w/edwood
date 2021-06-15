package main

import "io"

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
}

type Editbuf struct {
	curtext BufferObserver
	text    map[BufferObserver]struct{} // [private I think]
}
