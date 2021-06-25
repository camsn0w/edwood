package main

// BufferObserver separates observer functionality out of file
type BufferObserver interface {
	inserted(q0 int, r []rune) // Callback function to update the observer's texts
	deleted(q0, q1 int)        // Callback function to delete the observer's texts
}
