package engine

import (
	"fmt"
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
		lineNo := 0
		// fmt.Println("search:", self.Base.Tp())
		for line := range self.Iter() {
			lineNo += 1
			// fmt.Println("f:", line)
			if _, f := line.Filter(func(i int, s string) bool {
				return strings.Contains(s, k)
			}); f {
				// fmt.Println("found :", k, line)
				lines = append(lines, utils.Line{fmt.Sprintf("%s +%3d   :%s", line[0], lineNo, strings.Join(line[1:], " "))})
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
		return (self.Base.(*Xlsx)).GetHead(k)
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
	// fmt.Println("searching:")

	for {
		op := <-sender
		// fmt.Println("searching2:")
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

func (obj *BaseObj) Select(header string, columnIndex ...int) <-chan utils.Line {
	ch := make(chan utils.Line)
	go func() {
		defer close(ch)
		for line := range obj.Iter(header) {
			newline := utils.Line{}
			for _, ix := range columnIndex {
				newline = append(newline, line[ix+1])
			}
			ch <- newline
		}
	}()
	return ch
}

func (obj *BaseObj) Tables() []string {
	if obj.Tp() == "xlsx" {
		return (obj.Base.(*Xlsx)).obj.Sheets
	} else if obj.Tp() == "csv" {
		return []string{(obj.Base.(*Csv)).tableName}
	} else if obj.Tp() == "mbox" {
		return []string{(obj.Base.(*Mbox)).tableName}
	} else if obj.Tp() == "docx" {
		return []string{(obj.Base.(*Docx)).tableName}
	} else if obj.Tp() == "mbox" {
		return []string{(obj.Base.(*Mbox)).tableName}
	}
	return []string{}
}

func (obj *BaseObj) SelectByNames(header string, column_names ...string) (output <-chan utils.Line, err error) {
	// for _, name := range obj.Tables() {
	headerNames := obj.GetHeader(header)
	ixs := []int{}
	for _, h := range column_names {
		ix := headerNames.Index(h)

		if ix > -1 {
			ixs = append(ixs, ix)
		} else {
			err = fmt.Errorf("no such name in header:" + h)
			break
		}
	}
	if err != nil {
		return
	}
	if len(ixs) > 0 {
		return obj.Select(header, ixs...), nil
	}
	return
}

func (obj *BaseObj) SelectAllByNames(column_names ...string) (<-chan utils.Line, error) {
	ch := make(chan utils.Line)
	var err error
	go func() {
		defer close(ch)

		foundAll := false

		for _, name := range obj.Tables() {

			found := true
			headerNames := obj.GetHeader(name)

			// fmt.Println("table:", name, headerNames)
			ixs := []int{}
			for _, h := range column_names {
				ix := headerNames.Index(h)
				if ix > -1 {
					ixs = append(ixs, ix)
				} else {

					// fmt.Println("Select:", h, ix)
					// err = fmt.Errorf("no such name in header:")
					found = false
					break
				}
			}

			// fmt.Println("Select:", name, ixs)
			if len(ixs) > 0 && found {
				// fmt.Println("Select:", name, ixs)
				for l := range obj.Select(name, ixs...) {
					ch <- l
				}
			}
			if found {
				foundAll = true
			}
		}
		if !foundAll {
			err = fmt.Errorf("every header can not include :'%s'", strings.Join(column_names, ","))
		}

	}()
	return ch, err
}

func (obj *BaseObj) InsertInto(maches utils.Dict, values utils.BDict) (num int64, err error) {
	if obj.Tp() == "xlsx" {
		return (obj.Base.(*Xlsx)).InsertInto(maches, values)
	}
	return
	// return obj.Base.InsertInto(maches, values...)
	// return
}

const (
	LEFTJOIN  = 0
	RIGHTJOIN = 1
)
