package engine

type EmptyBaseClass struct {
	arrays []Line
}

func (empty EmptyBaseClass) Iter(header ...string) <-chan Line {
	ch := make(chan Line)
	go func() {
		for _, l := range empty.arrays {
			line := Line{}
			for _, i := range l {
				line.Push(i)
			}
			ch <- line
		}
		close(ch)
	}()
	return ch
}

func (empty EmptyBaseClass) Close() error {
	empty.arrays = []Line{}
	return nil
}

func (empty EmptyBaseClass) header(keylength ...int) (l Line) {
	if len(empty.arrays) > 0 {
		l = empty.arr
	}
}
