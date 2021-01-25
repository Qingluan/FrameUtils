package engine

import (
	"bufio"
	"strings"
)

type Txt struct {
	obj           *bufio.Scanner
	raw           string
	tableName     string
	fileterheader string
	nowheader     string
}

func (self *Txt) Iter(header ...string) <-chan Line {
	ch := make(chan Line)
	if header != nil {
		self.fileterheader = header[0]
	}
	self.obj = bufio.NewScanner(strings.NewReader(self.raw))

	go func() {
		c := 0
		for self.obj.Scan() {
			line := strings.TrimSpace(self.obj.Text())
			l := Line(strings.Fields(line))
			ch <- append(Line{self.tableName}, l...)
			c++
		}
		close(ch)
	}()
	return ch
}

func (self *Txt) GetHead(k string) Line {
	return Line{self.tableName}
}

func (self *Txt) Close() error {
	if self.obj != nil {
		self.obj = nil
	}
	return nil
}
func (self *Txt) header(...int) (l Line) {
	return
}

func (s *Txt) Tp() string {
	return "txt"
}

// func (self *Txt) Search(k string) (lines []Line) {
// 	hit := 0
// 	self.obj = bufio.NewScanner(strings.NewReader(self.raw))
// 	for self.obj.Scan() {
// 		line := strings.TrimSpace(self.obj.Text())
// 		if strings.Contains(line, k) {

// 			row := strings.Fields(line)
// 			hit++
// 			if hit%2000 == 0 {
// 				fmt.Println("hit:", hit)
// 			}
// 			if hit > 10000 {
// 				break
// 			}
// 			lines = append(lines, Line(row))
// 		}
// 	}
// 	return
// }

// func (self *Txt) DiffBy(other Obj, key ...string) (l []Line) {
// 	if key == nil {
// 		return
// 	}
// 	if len(key) == 1 {

// 	} else {

// 	}
// 	return
// }

// func (self *Txt) GetRow(i int) (ls []Line) {
// 	self.obj = bufio.NewScanner(strings.NewReader(self.raw))
// 	c := 0
// 	for self.obj.Scan() {
// 		line := strings.TrimSpace(self.obj.Text())
// 		if c == i {
// 			ls = append(ls, Line(strings.Fields(line)))
// 			return
// 		}
// 		c++
// 	}
// 	return
// }

// func (self *Txt) SearchTo(key string, linesChan chan []Line) {
// 	lines := self.Search(key)
// 	linesChan <- lines
// }
// func (self *Txt) Close() error {
// 	return nil
// }

// func (self *Txt) Work(sender chan string, reciver chan []Line) {
// 	for {
// 		op := <-sender
// 		if op == "[END]" {
// 			break
// 		}
// 		self.SearchTo(op, reciver)
// 	}
// }
