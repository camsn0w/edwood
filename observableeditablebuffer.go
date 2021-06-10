package main

type ObservableEditableBuffer interface {
	AddText(t *Text) *File
	DelText(t *Text) error
	AllText(tf func(t *Text))
	HasMultipleTexts() bool
	InsertAt(p0 int, s []rune)
	InsertAtWithoutCommit(p0 int, s []rune)
	DeleteAt(p0, p1 int)
	Undo(isundo bool) (q0, q1 int, ok bool)
}
