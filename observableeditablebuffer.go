package main

import "fmt"

type ObservableEditableBuffer interface {
	AddText(observer BufferObserver)
	DelText(observer BufferObserver) error
	SetCurText(observer BufferObserver)
	GetCurText() interface{}
	GetTextSize() int
	AllText(tf func(i interface{}))
	HasMultipleTexts() bool
	InsertAt(p0 int, s []rune)
	InsertAtWithoutCommit(p0 int, s []rune)
	DeleteAt(p0, p1 int)
	Undo(isundo bool) (q0, q1 int, ok bool)
	New() *Editbuf
}

type Editbuf struct {
	curtext BufferObserver
	text    map[BufferObserver]struct{} // [private I think]
}

func (f *File) AddText(observer BufferObserver) {
	if f.buf.text == nil {
		f.buf.text = make(map[BufferObserver]struct{})
	}
	f.buf.text[observer] = struct{}{}
	f.buf.curtext = observer

}

func (f *File) DelText(observer BufferObserver) error {
	if _, exists := f.buf.text[observer]; exists {
		delete(f.buf.text, observer)
		if observer == f.buf.curtext {
			for k := range f.buf.text {
				f.buf.curtext = k
				break
			}
		}
		return nil
	}
	return fmt.Errorf("can't find text in File.DelText")
}

func (f *File) SetCurText(observer BufferObserver) {
	if f == nil {
		println("F is nil in SetCurText")

	}
	if observer == nil {
		println("Observer is nil in SetCurText")
	}
	if f.buf.curtext == nil {
		println("Curtext IS NIL")
	}

	f.buf.curtext = observer
}

func (f *File) GetCurText() interface{} {
	return f.buf.curtext
}

func (f *File) AllText(tf func(i interface{})) {
	for t := range f.buf.text {
		tf(t)
	}
}

func (f *File) GetTextSize() int {
	return len(f.buf.text)
}

func (f *File) HasMultipleTexts() bool {
	return len(f.buf.text) > 1
}

func (f *File) InsertAt(p0 int, s []rune) {
	f.treatasclean = false
	if p0 > f.b.nc() {
		panic("internal error: fileinsert")
	}
	if f.seq > 0 {
		f.Uninsert(&f.delta, p0, len(s))
	}
	f.b.Insert(p0, s)
	if len(s) != 0 {
		f.Modded()
	}
	f.AllText(func(i interface{}) {
		i.(BufferObserver).inserted(p0, s)
	})
}

func (f *File) InsertAtWithoutCommit(p0 int, s []rune) {
	f.treatasclean = false
	if p0 > f.b.nc()+len(f.cache) {
		panic("File.InsertAtWithoutCommit insertion off the end")
	}

	if len(f.cache) == 0 {
		f.cq0 = p0
	} else {
		if p0 != f.cq0+len(f.cache) {
			// TODO(rjk): actually print something useful here
			acmeerror("File.InsertAtWithoutCommit cq0", nil)
		}
	}
	f.cache = append(f.cache, s...)
}

func (f *File) DeleteAt(p0, p1 int) {
	f.treatasclean = false
	if !(p0 <= p1 && p0 <= f.b.nc() && p1 <= f.b.nc()) {
		acmeerror("internal error: DeleteAt", nil)
	}
	if len(f.cache) > 0 {
		acmeerror("internal error: DeleteAt", nil)
	}

	if f.seq > 0 {
		f.Undelete(&f.delta, p0, p1)
	}
	f.b.Delete(p0, p1)

	// Validate if this is right.
	if p1 > p0 {
		f.Modded()
	}
	f.AllText(func(i interface{}) {
		i.(BufferObserver).deleted(p0, p1)
	})
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
			f.AllText(func(i interface{}) {
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
			f.AllText(func(i interface{}) {
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
func (f *File) New() *Editbuf {
	return &Editbuf{
		curtext: nil,
		text:    nil,
	}
}
