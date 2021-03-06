package engine

import (
	"strings"

	"github.com/Qingluan/FrameUtils/utils"
)

type BaseObj struct {
	Base
}

func (self *BaseObj) String() string {
	head := self.header()
	msg := strings.Join(head, " -|- ")
	n := 0
	for i := range self.Iter() {
		if n > 100 {
			break
		}
		msg += "\n" + strings.Join(i, "  |  ")
		n++
	}
	// fmt.Println("F", msg)
	return msg
}

func (self *BaseObj) Search(k string) (lines []utils.Line) {
	if strings.Contains(k, "=") {
		ks := strings.SplitN(k, "=", 2)
		key, value := strings.TrimSpace(ks[0]), strings.TrimSpace(ks[1])

		header := false
		found := -1
		for line := range self.Iter() {
			if !header {
				header = true
				if found, _ = line.Filter(func(i int, s string) bool {
					return strings.Contains(s, key)
				}); found == -1 {
					break
				}
				continue
			}
			if strings.Contains(line[found], value) {
				lines = append(lines, line)
			}

		}
	} else {
		for line := range self.Iter() {

			if _, f := line.Filter(func(i int, s string) bool {
				return strings.Contains(s, k)
			}); f {
				lines = append(lines, line)
			}
		}
	}
	return
}

func (self *BaseObj) GetHeader(k string) (header utils.Line) {
	if self.Tp() == "json" {
		return (self.Base.(*JsonObj)).GetHead(k)
	} else if self.Tp() == "sqltxt" {
		return (self.Base.(*SqlTxt)).GetHead(k)
	} else if self.Tp() == "xlsx" {
		(self.Base.(*Xlsx)).GetHead(k)
	} else {
		return
	}
	return
}

func (self *BaseObj) Header(ks ...int) (header utils.Line) {
	if l := self.header(ks...); len(l) > 0 {
		if len(l) > 0 {
			return l
		}
	}
	for line := range self.Iter() {
		return line[1:]
	}
	return
}

func (self *BaseObj) SearchTo(key string, linesChan chan []utils.Line) {
	lines := self.Search(key)
	linesChan <- lines
}
func (self *BaseObj) Close() error {
	return self.Close()
}

func (self *BaseObj) Work(sender chan string, reciver chan []utils.Line) {
	for {
		op := <-sender
		if op == "[END]" {
			break
		}
		self.SearchTo(op, reciver)
	}
}

func (self *BaseObj) AsJson() (ds []utils.Dict) {
	if self.Tp() == "json" {
		return (self.Base.(*JsonObj)).Datas
	} else if self.Tp() == "sqltxt" {
		return (self.Base.(*SqlTxt)).ToJson()
	} else {
		header := self.Header()
		for line := range self.Iter() {
			if len(header) == 0 {
				header = line
				continue
			}
			ds = append(ds, line[1:].FromKey(header))
		}
	}

	return
}

func (self *BaseObj) Where(filter func(lineno int, line utils.Line, wordno int, word string) bool) (newObj *BaseObj) {
	header := self.Header()
	// fmt.Println(header)
	jsonObj := &JsonObj{
		Header: header,
	}
	c := 0
	for line := range self.Iter() {
		if l := line.Collect(func(i int, word string) bool {
			return filter(c, line, i, word)
		}); len(l) > 0 {
			jsonObj.Datas = append(jsonObj.Datas, l.FromKey(header))
		}
	}
	return
}

const (
	LEFTJOIN  = 0
	RIGHTJOIN = 1
)
