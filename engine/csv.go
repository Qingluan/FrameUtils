package engine

import (
	"encoding/csv"
	"io"
	"strings"
)

type Csv struct {
	raw string
	obj *csv.Reader
}

func (self *Csv) Iter() <-chan Line {
	ch := make(chan Line)
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
			ch <- Line(row)
			c++
		}
		close(ch)
	}()
	return ch
}

func (self *Csv) GetHead(k string) Line {
	return nil
}

func (self *Csv) Close() error {
	return nil
}
func (self *Csv) header(...int) (l Line) {
	return
}
func (s *Csv) Tp() string {
	return "csv"
}

// func (self *Csv) DiffBy(other Obj, key ...string) (l []Line) {
// 	if key == nil {
// 		return
// 	}
// 	if len(key) == 1 {

// 	} else {

// 	}
// 	return
// }
// func (self *Csv) GetColumn(key interface{}) (ls Line) {
// 	colint := -1
// 	switch key.(type) {
// 	case int:
// 		self.obj = csv.NewReader(strings.NewReader(self.raw))
// 		self.obj.LazyQuotes = true
// 		c := 0
// 		colint = key.(int)
// 		for {
// 			row, err := self.obj.Read()
// 			if err == io.EOF {
// 				break
// 			}
// 			if err != nil {
// 				continue
// 			}
// 			if colint != -1 {
// 				ls = append(ls, row[colint])

// 			}
// 			c++
// 		}
// 	case string:
// 		self.obj = csv.NewReader(strings.NewReader(self.raw))
// 		self.obj.LazyQuotes = true
// 		c := 0
// 		header := false
// 		for {
// 			row, err := self.obj.Read()
// 			if err == io.EOF {
// 				break
// 			}
// 			if err != nil {
// 				continue
// 			}
// 			if !header {
// 				for i, v := range row {
// 					if strings.Contains(v, key.(string)) {
// 						colint = i
// 						break
// 					}
// 				}
// 				header = true
// 				continue
// 			}
// 			if colint != -1 {
// 				ls = append(ls, row[colint])
// 			}
// 			c++
// 		}
// 	}
// 	return
// }

// func (self *Csv) GetRow(i int) (ls []Line) {
// 	self.obj = csv.NewReader(strings.NewReader(self.raw))
// 	self.obj.LazyQuotes = true
// 	c := 0
// 	for {
// 		row, err := self.obj.Read()
// 		if err == io.EOF {
// 			break
// 		}
// 		if err != nil {
// 			continue
// 		}
// 		if c == i {
// 			ls = append(ls, Line(row))
// 			return
// 		}
// 		c++
// 	}
// 	return
// }

// func (self *Csv) Search(k string) (lines []Line) {
// 	self.obj = csv.NewReader(strings.NewReader(self.raw))
// 	self.obj.LazyQuotes = true
// 	hit := 0

// 	if strings.Contains(k, "=") {
// 		ks := strings.SplitN(k, "=", 2)
// 		key, value := strings.TrimSpace(ks[0]), strings.TrimSpace(ks[1])

// 		header := false
// 		found := -1
// 		for {
// 			row, err := self.obj.Read()
// 			if err == io.EOF {
// 				break
// 			}
// 			if err != nil {
// 				continue
// 			}
// 			if !header {
// 				header = true
// 				for i, v := range row {
// 					if strings.Contains(v, key) {
// 						found = i
// 						break
// 					}
// 				}
// 				if found == -1 {
// 					return
// 				}
// 				continue
// 			}

// 			if strings.Contains(row[found], value) {
// 				lines = append(lines, Line(row))
// 			}
// 		}
// 		return
// 	}
// All:
// 	for {
// 		row, err := self.obj.Read()
// 		if err == io.EOF {
// 			break
// 		}
// 		if err != nil {
// 			// log.Println(err)
// 			continue
// 		}
// 	Match:
// 		for _, v := range row {
// 			if strings.Contains(v, k) {
// 				// fmt.Println(v)
// 				hit++
// 				if hit%2000 == 0 {
// 					fmt.Println("hit:", hit)
// 				}
// 				if hit > 10000 {
// 					break All
// 				}
// 				lines = append(lines, Line(row))
// 				break Match
// 			}
// 		}

// 	}
// 	return
// }

// func (self *Csv) SearchTo(key string, linesChan chan []Line) {
// 	lines := self.Search(key)
// 	linesChan <- lines
// }

// func (self *Csv) Work(sender chan string, reciver chan []Line) {
// 	for {
// 		op := <-sender
// 		if op == "[END]" {
// 			break
// 		}
// 		self.SearchTo(op, reciver)
// 	}
// }
