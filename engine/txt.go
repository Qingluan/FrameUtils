package engine

import (
	"bufio"
	"strings"

	"github.com/Qingluan/FrameUtils/utils"
)

type Txt struct {
	obj           *bufio.Scanner
	raw           string
	tableName     string
	fileterheader string
	nowheader     string
}

func (self *Txt) Iter(header ...string) <-chan utils.Line {
	ch := make(chan utils.Line)
	if header != nil {
		self.fileterheader = header[0]
	}
	self.obj = bufio.NewScanner(strings.NewReader(self.raw))

	go func() {
		c := 0
		for self.obj.Scan() {
			line := strings.TrimSpace(self.obj.Text())
			l := utils.Line(strings.Fields(line))
			ch <- append(utils.Line{self.tableName}, l...)
			c++
		}
		close(ch)
	}()
	return ch
}

func (self *Txt) GetHead(k string) utils.Line {
	return utils.Line{self.tableName}
}

func (self *Txt) Close() error {
	if self.obj != nil {
		self.obj = nil
	}
	return nil
}
func (self *Txt) header(...int) (l utils.Line) {
	return
}

func (s *Txt) Tp() string {
	return "txt"
}
