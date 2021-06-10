package main

type BufferObserver interface {
	inserted(q0 int, r []rune)
	deleted(q0, q1 int)
}
