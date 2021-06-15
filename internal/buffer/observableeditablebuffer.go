package main

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
}

type Editbuf struct {
	curtext BufferObserver
	text    map[BufferObserver]struct{} // [private I think]
}
