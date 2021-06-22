package main

type BufferObserver interface { //Interface to separate text functionality out of file
	inserted(q0 int, r []rune) //Callback function to update text observers
	deleted(q0, q1 int)        //Callback function to delete text observers
}
