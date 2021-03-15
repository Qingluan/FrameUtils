package engine

import (
	"encoding/csv"
	"io"
	"strings"

	"github.com/Qingluan/FrameUtils/utils"
)

type Csv struct {
	raw       string
	obj       *csv.Reader
	Header    utils.Line
	tableName string
}

func (self *Csv) Iter(header ...string) <-chan utils.Line {
	ch := make(chan utils.Line)
	self.obj = csv.NewReader(strings.NewReader(self.raw))
	self.obj.LazyQuotes = true

	go func() {
		c := 0
		for {
			row, err := self.obj.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				continue
			}
			if c == 0 {
				self.Header = utils.Line(row)
			}
			ch <- append(utils.Line{self.tableName}, utils.Line(row)...)
			c++
		}
		close(ch)
	}()
	return ch
}

func (self *Csv) GetHead(k string) utils.Line {
	return self.Header
}

func (self *Csv) Close() error {
	return nil
}
func (self *Csv) header(...int) (l utils.Line) {
	return self.Header
}
func (s *Csv) Tp() string {
	return "csv"
}
