package textconvert

import (
	"bufio"
	"io"
	"os"
	"strings"

	"github.com/thedatashed/xlsxreader"
)

func XlsxToEs(path string) (es ElasticFileDocs, err error) {

	xl, err := xlsxreader.OpenFile(path)
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

func NormalToEs(path string) (es ElasticFileDocs, err error) {

	fp, err := os.Open(path)
	if err != nil {
		return es, err
	}
	defer fp.Close()
	reader := bufio.NewReader(fp)
	for {
		l, err := reader.ReadString(byte('\n'))
		if err == io.EOF {
			break
		}
		msg, err := TOUTF8(l)
		es.SomeStr += msg + "\n"

	}
	es.Path = path
	// if buf, err := ioutil.ReadFile(path); err != nil {
	// 	return es, err
	// } else {
	// 	es.SomeStr, err = TOUTF8(string(buf))
	// 	es.Path = path
	// }
	return
}
