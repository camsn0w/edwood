package file

// BufferObserver separates observer functionality out of file.
// A BufferObserver is something that can be kept track
// of through the ObservableEditableBuffer.
type BufferObserver interface {

	// Inserted is a callback function which updates the observer's texts.
	Inserted(q0 int, r []rune)

	// Deleted is a callback function which deletes the observer's texts.
	Deleted(q0, q1 int)
}
