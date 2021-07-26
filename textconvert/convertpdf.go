package textconvert

import (
	"fmt"
	"log"

	"github.com/Qingluan/FrameUtils/textconvert/pdf"
)

func readPdf(path string) (string, error) {
	f, r, err := pdf.Open(path)
	defer func() {
		_ = f.Close()
	}()
	if err != nil {
		log.Println("Pdf open err:", path, err)
		return "", err
	}
	totalPage := r.NumPage()
	msg := ""
	for pageIndex := 1; pageIndex <= totalPage; pageIndex++ {
		p := r.Page(pageIndex)
		if p.V.IsNull() {
			continue
		}

		rows, err2 := p.GetTextByRow()
		if err2 != nil {
			// log.Println("Pdf get text err:", err2)
			return "", err2
		}
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

func PDFToStr(path string) (content string, err error) {
	content, err = readPdf(path)
	return
}
func PdfToEs(path string) (es ElasticFileDocs, err error) {
	es.SomeStr, err = readPdf(path)
	es.Path = path
	return
}
