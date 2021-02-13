package engine

type EmptyBaseClass struct {
	Arrays []Line
	Key    Line
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
	if len(empty.Key) > 0 {
		return empty.Key
	} else if len(empty.Arrays) > 0 {
		return empty.Arrays[0]
	}
	return
}
