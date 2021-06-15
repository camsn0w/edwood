package main

type BufferObserver interface {
	Inserted(q0 int, r []rune)
	Deleted(q0, q1 int)
}
