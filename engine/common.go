package engine

type EmptyBaseClass struct {
	Arrays []Line
}

func (empty EmptyBaseClass) Iter(header ...string) <-chan Line {
	ch := make(chan Line)
	go func() {
		for _, l := range empty.Arrays {
			ch <- append(Line{""}, l...)
		}
		close(ch)
	}()
	return ch
}

func (empty EmptyBaseClass) Close() error {
	empty.Arrays = []Line{}
	return nil
}

func (empty EmptyBaseClass) Tp() string {
	return "Empty"
}

func (empty EmptyBaseClass) header(keylength ...int) (l Line) {
	if len(empty.Arrays) > 0 {
		if keylength != nil {
			l.Push(empty.Arrays[0][0])
		} else {
			for _, f := range empty.Arrays[0] {
				l.Push(f)
			}
		}
	}
	return
}
