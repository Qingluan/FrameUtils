package textconvert

import (
	"strings"

	"github.com/shakinm/xlsReader/xls"
	"github.com/thedatashed/xlsxreader"
)

func XlsxToEs(path string) (es ElasticFileDocs, err error) {

	xl, err := xlsxreader.OpenFile(path)
	if err != nil {
		return es, err
	}
	defer xl.Close()
	tables := []string{}
	for _, table := range xl.Sheets {
		// tableMsg := []string{}

		for row := range xl.ReadRows(table) {
			line := []string{}
			for _, v := range row.Cells {
				line = append(line, v.Value)
			}
			tables = append(tables, strings.Join(line, " | "))
		}
		// tables = append(tables, strings.Join(tables, "\n"))
		tables = append(tables, table)
		tables = append(tables, "\n")
	}
	es.SomeStr = strings.Join(tables, "\n")
	es.Path = path
	return

}

func XlsToStr(path string, sep ...string) (content string, err error) {
	seps := " | "
	if sep != nil {
		seps = sep[0]
	}
	xl, err := xls.OpenFile(path)
	// xl, err := xls.OpenReader()
	if err != nil {
		return "", err
	}
	// defer xl.Close()
	tables := []string{}
	for _, table := range xl.GetSheets() {
		// tableMsg := []string{}

		for rowid := 0; rowid < table.GetNumberRows(); rowid++ {
			line := []string{}
			row, _ := table.GetRow(rowid)

			for _, v := range row.GetCols() {
				line = append(line, v.GetString())
			}
			tables = append(tables, strings.Join(line, seps))
		}
		// tables = append(tables, strings.Join(tables, "\n"))
		tables = append(tables, table.GetName())
		tables = append(tables, "\n")
	}
	content = strings.Join(tables, "\n")
	// es.Path = path
	return

}

func XlsxToStr(path string, sep ...string) (content string, err error) {
	seps := " | "
	if sep != nil {
		seps = sep[0]
	}
	xl, err := xlsxreader.OpenFile(path)
	if err != nil {
		return "", err
	}
	defer xl.Close()
	tables := []string{}
	for _, table := range xl.Sheets {
		// tableMsg := []string{}

		for row := range xl.ReadRows(table) {
			line := []string{}
			for _, v := range row.Cells {
				line = append(line, v.Value)
			}
			tables = append(tables, strings.Join(line, seps))
		}
		// tables = append(tables, strings.Join(tables, "\n"))
		tables = append(tables, table)
		tables = append(tables, "\n")
	}
	content = strings.Join(tables, "\n")
	// es.Path = path
	return

}

// func InsertInto(targetDataframe string, matchKey map[string]string, values ...string) {
// 	if strings.Contains(targetDataframe, ".") {
// 		if strings.HasSuffix(targetDataframe, ".xlsx") {
// 			xlsx := engine.OpenObj()
// 		}
// 	}
// }
