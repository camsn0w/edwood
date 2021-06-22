package main

import "fmt"

type Editor struct {
	currobserver BufferObserver
	observers    map[BufferObserver]struct{} // [private I think]
}

func (f *File) AddObserver(observer BufferObserver) {
	if f.editor.observers == nil {
		f.editor.observers = make(map[BufferObserver]struct{})
	}
	f.editor.observers[observer] = struct{}{}
	f.editor.currobserver = observer

}

func (f *File) DelObserver(observer BufferObserver) error {
	if _, exists := f.editor.observers[observer]; exists {
		delete(f.editor.observers, observer)
		if observer == f.editor.currobserver {
			for k := range f.editor.observers {
				f.editor.currobserver = k
				break
			}
		}
		return nil
	}
	return fmt.Errorf("can't find editor in File.DelObserver")
}

func (f *File) SetCurObserver(observer BufferObserver) {
	f.editor.currobserver = observer
}

func (f *File) GetCurObserver() interface{} {
	return f.editor.currobserver
}

func (f *File) AllObservers(tf func(i interface{})) {
	for t := range f.editor.observers {
		tf(t)
	}
}

func (f *File) GetObserverSize() int {
	return len(f.editor.observers)
}

func (f *File) HasMultipleObservers() bool {
	return len(f.editor.observers) > 1
}
