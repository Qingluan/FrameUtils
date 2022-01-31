package engine

import (
	"log"
	"strings"

	"github.com/Qingluan/FrameUtils/textconvert"
	"github.com/Qingluan/FrameUtils/utils"
)

type Docx struct {
	raw       string
	obj       string
	Header    utils.Line
	tableName string
}

func (self *Docx) Iter(header ...string) <-chan utils.Line {
	ch := make(chan utils.Line)
	var err error
	self.obj, err = textconvert.ToStr(self.tableName)

	if err != nil {
		log.Fatal("read docx err:", err)

	}
	self.raw = self.obj
	// self.obj.LazyQuotes = true

	go func() {

		for c, l := range strings.Split(self.obj, "\n") {

			if c == 0 {
				self.Header = utils.Line{l}
			}
			ch <- append(utils.Line{self.tableName}, utils.Line{l}...)
			c++
		}
		close(ch)
	}()
	return ch
}

func (self *Docx) GetHead(k string) utils.Line {
	return self.Header
}

func (self *Docx) Close() error {
	return nil
}
func (self *Docx) header(...int) (l utils.Line) {
	return self.Header
}
func (s *Docx) Tp() string {
	return "docx"
}

func (obj *Docx) InsertInto(maches utils.Dict, values ...interface{}) (num int64, err error) {
	return
}
