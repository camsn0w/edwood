package main

import "fmt"

type Editor struct {
	currobserver BufferObserver
	observers    map[BufferObserver]struct{} // [private I think]
}

func (e *Editor) AddObserver(observer BufferObserver) {
	if e.observers == nil {
		e.observers = make(map[BufferObserver]struct{})
	}
	e.observers[observer] = struct{}{}
	e.currobserver = observer

}

func (e *Editor) DelObserver(observer BufferObserver) error {
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

func (e *Editor) SetCurObserver(observer BufferObserver) {
	e.currobserver = observer
}

func (e *Editor) GetCurObserver() interface{} {
	return e.currobserver
}

func (e *Editor) AllObservers(tf func(i interface{})) {
	for t := range e.observers {
		tf(t)
	}
}

func (e *Editor) GetObserverSize() int {
	return len(e.observers)
}

func (e *Editor) HasMultipleObservers() bool {
	return len(e.observers) > 1
}
