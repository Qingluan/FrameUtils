package engine

import "github.com/Qingluan/FrameUtils/utils"

type EmptyBaseClass struct {
	Arrays []utils.Line
	Key    utils.Line
}

func (empty EmptyBaseClass) Iter(header ...string) <-chan utils.Line {
	ch := make(chan utils.Line)
	go func() {
		for _, l := range empty.Arrays {
			ch <- append(utils.Line{""}, l...)
		}
		close(ch)
	}()
	return ch
}

func (empty EmptyBaseClass) Close() error {
	empty.Arrays = []utils.Line{}
	return nil
}

func (empty EmptyBaseClass) Tp() string {
	return "Empty"
}

func (empty EmptyBaseClass) header(keylength ...int) (l utils.Line) {
	if len(empty.Key) > 0 {
		return empty.Key
	} else if len(empty.Arrays) > 0 {
		return empty.Arrays[0]
	}
	return
}

func (empty EmptyBaseClass) InsertInto(maches utils.Dict, values ...interface{}) (num int64, err error) {
	return
}
