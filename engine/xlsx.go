package engine

import (
	"fmt"

	"github.com/Qingluan/FrameUtils/utils"
	"github.com/thedatashed/xlsxreader"
	"github.com/xuri/excelize/v2"
)

var (
	RAW_IX = []string{
		"A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L", "M", "N", "O", "P", "Q", "R", "S", "T", "U", "V", "W", "X", "Y", "Z",
	}
)

type Xlsx struct {
	obj         *xlsxreader.XlsxFileCloser
	filtertable string
	nowtable    string
	name        string
	tables      []string
}

func AddLine(r []xlsxreader.Cell) (l utils.Line) {
	for _, v := range r {
		l = append(l, v.Value)
	}
	return
}
func (self *Xlsx) GetHead(k string) utils.Line {
	for _, s := range self.obj.Sheets {
		if k == s {
			for row := range self.obj.ReadRows(k) {
				return AddLine(row.Cells)
			}
		}
	}
	return nil
}

func (self *Xlsx) Iter(filtertable ...string) <-chan utils.Line {
	ch := make(chan utils.Line)
	if filtertable != nil {
		self.filtertable = filtertable[0]
	}
	notFilter := false
	if len(self.tables) > 0 {
		notFilter = true
	}
	go func() {
		for _, table := range self.obj.Sheets {

			self.nowtable = table
			if notFilter {
				self.tables = append(self.tables, table)
			}
			if self.filtertable != "" && self.filtertable != table {
				continue
			}
			for row := range self.obj.ReadRows(table) {
				l := AddLine(row.Cells)
				ch <- append(utils.Line{table}, l...)
			}
		}
		close(ch)
	}()
	return ch
}

func (self *Xlsx) Close() error {
	return self.obj.Close()
}

func (self *Xlsx) header(k ...int) (l utils.Line) {
	if k != nil {
		return self.GetHead(self.tables[k[0]])
	}
	return
}
func (s *Xlsx) Tp() string {
	return "xlsx"
}

func (s *Xlsx) Tables() []string {
	return s.obj.Sheets
}

func GetLocation(row, column int) string {
	turn := column / len(RAW_IX)
	left := column % len(RAW_IX)
	if turn == 0 {
		return fmt.Sprintf("%s%d", RAW_IX[left], row+1)
	} else {
		return fmt.Sprintf("%s%s%d", RAW_IX[turn], RAW_IX[left], row+1)
	}

}

type LocationValue struct {
	row    int
	column int
	values utils.BDict
}

func (obj *Xlsx) InsertInto(maches utils.Dict, values utils.BDict) (num int64, err error) {
	for _, tableName := range obj.Tables() {
		headers := obj.GetHead(tableName)
		matchesAll := true
		matchesIx := map[string]int{}
		for k := range maches {
			if !headers.Contain(k) {
				matchesAll = false
				break
			} else {
				ix := headers.Index(k)
				if ix == -1 {
					matchesAll = false
					break
				}
				matchesIx[k] = ix
			}
		}

		if !matchesAll {
			continue
		}
		foundRowIx := []LocationValue{}
		no := 0
		for row := range obj.obj.ReadRows(tableName) {
			found := true
			for k, ix := range matchesIx {
				if row.Cells[ix].Value != fmt.Sprint(maches[k]) {
					found = false
					break
				}
			}
			if !found {
				continue
			} else {
				foundRowIx = append(foundRowIx, LocationValue{
					row:    no,
					column: len(headers),
					values: values,
				})
			}
			// l := AddLine(row.Cells)
			no += 1
		}
		if len(foundRowIx) > 0 {
			f, werr := excelize.OpenFile(obj.name)
			if werr != nil {
				fmt.Println(err)
				return
			}
			defer func() {
				// Close the spreadsheet.
				if err := f.Close(); err != nil {
					fmt.Println(err)
				}
			}()
			for _, maches := range foundRowIx {

				for k, v := range maches.values {
					column := headers.Index(k)
					if column == -1 {
						column = len(headers)
						xlsxix := GetLocation(0, column)
						f.SetCellValue(tableName, xlsxix, k)
						headers = append(headers, k)
					}

					xlsxix := GetLocation(maches.row, column)
					f.SetCellValue(tableName, xlsxix, v)
				}
			}

		}

	}
	return
}
