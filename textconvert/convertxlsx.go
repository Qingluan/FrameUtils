package textconvert

import (
	"strings"

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

func XlsxToStr(path string) (content string, err error) {

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
			tables = append(tables, strings.Join(line, " | "))
		}
		// tables = append(tables, strings.Join(tables, "\n"))
		tables = append(tables, table)
		tables = append(tables, "\n")
	}
	content = strings.Join(tables, "\n")
	// es.Path = path
	return

}
