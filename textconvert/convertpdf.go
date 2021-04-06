package textconvert

import (
	"fmt"

	"github.com/dcu/pdf"
)

func readPdf(path string) (string, error) {
	f, r, err := pdf.Open(path)
	defer func() {
		_ = f.Close()
	}()
	if err != nil {
		return "", err
	}
	totalPage := r.NumPage()
	msg := ""
	for pageIndex := 1; pageIndex <= totalPage; pageIndex++ {
		p := r.Page(pageIndex)
		if p.V.IsNull() {
			continue
		}

		rows, _ := p.GetTextByRow()
		for _, row := range rows {
			// println(">>>> row: ", row.Position)
			line := ""
			for _, word := range row.Content {
				line += fmt.Sprint(word.S)

			}
			msg += line + "\n"
		}
	}

	return msg, nil
}

func PdfToEs(path string) (es ElasticFileDocs, err error) {
	es.SomeStr, err = readPdf(path)
	es.Path = path
	return
}
