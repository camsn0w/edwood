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

func (f *File) Undo(isundo bool) (q0, q1 int, ok bool) {
	var (
		stop           int
		delta, epsilon *[]*Undo
	)
	if isundo {
		// undo; reverse delta onto epsilon, seq decreases
		delta = &f.delta
		epsilon = &f.epsilon
		stop = f.seq
	} else {
		// redo; reverse epsilon onto delta, seq increases
		delta = &f.epsilon
		epsilon = &f.delta
		stop = 0 // don't know yet
	}

	for len(*delta) > 0 {
		u := (*delta)[len(*delta)-1]
		if isundo {
			if u.seq < stop {
				f.seq = u.seq
				return
			}
		} else {
			if stop == 0 {
				stop = u.seq
			}
			if u.seq > stop {
				return
			}
		}
		switch u.t {
		default:
			panic(fmt.Sprintf("undo: 0x%x\n", u.t))
		case Delete:
			f.seq = u.seq
			f.Undelete(epsilon, u.p0, u.p0+u.n)
			f.mod = u.mod
			f.treatasclean = false
			f.b.Delete(u.p0, u.p0+u.n)
			f.AllObservers(func(i interface{}) {
				i.(BufferObserver).deleted(u.p0, u.p0+u.n)

			})
			q0 = u.p0
			q1 = u.p0
			ok = true
		case Insert:
			f.seq = u.seq
			f.Uninsert(epsilon, u.p0, u.n)
			f.mod = u.mod
			f.treatasclean = false
			f.b.Insert(u.p0, u.buf)
			f.AllObservers(func(i interface{}) {
				i.(BufferObserver).inserted(u.p0, u.buf)
			})
			q0 = u.p0
			q1 = u.p0 + u.n
			ok = true
		case Filename:
			// TODO(rjk): If I have a zerox, does undo a filename change update?
			f.seq = u.seq
			f.UnsetName(epsilon)
			f.mod = u.mod
			f.treatasclean = false
			newfname := string(u.buf)
			f.setnameandisscratch(newfname)
		}
		*delta = (*delta)[0 : len(*delta)-1]
	}
	// TODO(rjk): Why do we do this?
	if isundo {
		f.seq = 0
	}
	return q0, q1, ok
}
