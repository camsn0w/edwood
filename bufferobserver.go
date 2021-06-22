package main

type BufferObserver interface { //Interface to separate observers functionality out of file
	inserted(q0 int, r []rune) //Callback function to update observers observers
	deleted(q0, q1 int)        //Callback function to delete observers observers
}
