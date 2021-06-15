package main

import "io"

type RuneBuffer interface {
	Insert(q0 int, r []rune)
	Delete(q0, q1 int)
	Read(q0 int, r []rune) (int, error)
	Reader(q0, q1 int) io.Reader
	ReadC(q int) rune
	String() string
	Reset()
	nc() int
	Nbyte() int
	View(q0, q1 int) []rune
	IndexRune(r rune) int
	Equal(s RuneArray) bool
}
