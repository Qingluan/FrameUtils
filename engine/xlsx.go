package engine

import (
	"github.com/thedatashed/xlsxreader"
)

type Xlsx struct {
	obj         *xlsxreader.XlsxFileCloser
	filtertable string
	nowtable    string
}

func AddLine(r []xlsxreader.Cell) (l Line) {
	for _, v := range r {
		l = append(l, v.Value)
	}
	return
}
func (self *Xlsx) GetHead(k string) Line {
	for _, s := range self.obj.Sheets {
		if k == s {
			for row := range self.obj.ReadRows(k) {
				return AddLine(row.Cells)
			}
		}
	}
	return nil
}

func (self *Xlsx) Iter(filtertable ...string) <-chan Line {
	ch := make(chan Line)
	if filtertable != nil {
		self.filtertable = filtertable[0]
	}
	go func() {
		for _, table := range self.obj.Sheets {
			self.nowtable = table
			if self.filtertable != "" && self.filtertable != table {
				continue
			}
			for row := range self.obj.ReadRows(table) {
				l := AddLine(row.Cells)
				ch <- append(Line{table}, l...)
			}
		}
		close(ch)
	}()
	return ch
}

func (self *Xlsx) Close() error {
	return self.obj.Close()
}

func (self *Xlsx) header(k ...int) (l Line) {
	return
}
func (s *Xlsx) Tp() string {
	return "xlsx"
}
