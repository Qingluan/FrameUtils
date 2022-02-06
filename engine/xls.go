package engine

import (
	"github.com/Qingluan/FrameUtils/utils"
	"github.com/shakinm/xlsReader/xls"
)

type Xls struct {
	obj         *xls.Workbook
	filtertable string
	nowtable    string
	name        string
	tables      []string
}

func addLine(shet *xls.Sheet, ix int) utils.Line {
	l := utils.Line{}
	rw, _ := shet.GetRow(ix)
	for _, cell := range rw.GetCols() {
		l = append(l, cell.GetString())
	}
	return l
}

func (self *Xls) GetHead(k string) utils.Line {
	tn := self.obj.GetNumberSheets()
	for i := 0; i < tn; i++ {
		ss, _ := self.obj.GetSheet(i)
		s := ss.GetName()
		if k == s {
			return addLine(ss, 0)
		}
	}
	return nil
}

func (self *Xls) Iter(filtertable ...string) <-chan utils.Line {
	ch := make(chan utils.Line)
	if filtertable != nil {
		self.filtertable = filtertable[0]
	}
	notFilter := false
	if len(self.tables) > 0 {
		notFilter = true
	}
	go func() {
		for _, table := range self.obj.GetSheets() {

			self.nowtable = table.GetName()
			if notFilter {
				self.tables = append(self.tables, table.GetName())
			}
			if self.filtertable != "" && self.filtertable != table.GetName() {
				continue
			}
			for ix := range table.GetRows() {
				l := addLine(&table, ix)
				ch <- append(utils.Line{table.GetName()}, l...)
			}
		}
		close(ch)
	}()
	return ch
}

func (self *Xls) Close() error {
	return nil
}

func (self *Xls) header(k ...int) (l utils.Line) {
	if k != nil {
		return self.GetHead(self.tables[k[0]])
	}
	return
}
func (s *Xls) Tp() string {
	return "xls"
}

func (s *Xls) Tables() (l []string) {

	for _, s := range s.obj.GetSheets() {
		l = append(l, s.GetName())
	}
	return
}
