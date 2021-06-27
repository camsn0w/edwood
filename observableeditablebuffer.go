package main

import "fmt"

// observableeditablebuffer has a file that and is a
// type through which the main program will add, remove and check
// on the current observer(s) for a Text
type observableeditablebuffer struct {
	currobserver BufferObserver
	observers    map[BufferObserver]struct{} // [private I think]
}

// AddObserver adds e as an observer for edits to this File.
func (e *observableeditablebuffer) AddObserver(observer BufferObserver) {
	if e.observers == nil {
		e.observers = make(map[BufferObserver]struct{})
	}
	e.observers[observer] = struct{}{}
	e.currobserver = observer

}

// DelObserver removes e as an observer for edits to this File.
func (e *observableeditablebuffer) DelObserver(observer BufferObserver) error {
	if _, exists := e.observers[observer]; exists {
		delete(e.observers, observer)
		if observer == e.currobserver {
			for k := range e.observers {
				e.currobserver = k
				break
			}
		}
		return nil
	}
	return fmt.Errorf("can't find editor in File.DelObserver")
}

// SetCurObserver sets the current observer
func (e *observableeditablebuffer) SetCurObserver(observer BufferObserver) {
	e.currobserver = observer
}

// GetCurObserver gets the current observer and returns a interface{}
func (e *observableeditablebuffer) GetCurObserver() interface{} {
	return e.currobserver
}

// AllObservers preforms tf(all observers...)
func (e *observableeditablebuffer) AllObservers(tf func(i interface{})) {
	for t := range e.observers {
		tf(t)
	}
}

// GetObserverSize will return the size of the observer map
func (e *observableeditablebuffer) GetObserverSize() int {
	return len(e.observers)
}

// HasMultipleObservers returns true if their are multiple observers to the File
func (e *observableeditablebuffer) HasMultipleObservers() bool {
	return len(e.observers) > 1
}

// insertOnAll inserts at q0 for all observers in the observer map
func (e *observableeditablebuffer) insertOnAll(q0 int, r []rune) {
	e.AllObservers(func(i interface{}) {
		i.(BufferObserver).inserted(q0, r)
	})
}

// delete on all deletes q0 to q1 on all of the observer in the observer map
func (e *observableeditablebuffer) deleteOnAll(q0 int, q1 int) {
	e.AllObservers(func(i interface{}) {
		i.(BufferObserver).deleted(q0, q1)
	})
}
