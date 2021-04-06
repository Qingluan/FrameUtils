package textconvert

import (
	"bufio"
	"io"
	"io/ioutil"
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
	info, _ := fp.Stat()
	if info.Size() > 1024*1024*10 {
		reader := bufio.NewReader(fp)
		msg := ""
		for {
			l, err := reader.ReadString(byte('\n'))
			if err == io.EOF || err != nil {
				break
			}
			msg += l + "\n"
		}

		es.SomeStr, err = TOUTF8(msg)
		es.Path = path

	} else {
		buf, err := ioutil.ReadAll(fp)
		if err != nil {
			return es, err
		}
		es.SomeStr, err = TOUTF8(string(buf))
	}
	// if buf, err := ioutil.ReadFile(path); err != nil {
	// 	return es, err
	// } else {
	// 	es.SomeStr, err = TOUTF8(string(buf))
	// 	es.Path = path
	// }
	return
}
