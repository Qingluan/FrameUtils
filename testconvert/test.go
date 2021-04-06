package main

import (
	"log"
	"os"

	"github.com/Qingluan/FrameUtils/textconvert"
)

func main() {
	scan := textconvert.NewDirScan(os.Args[1], 20)
	scan.SetHandle("doc", textconvert.DocxToEs)
	scan.SetHandle("docx", textconvert.DocxToEs)
	scan.SetHandle("xlsx", textconvert.XlsxToEs)
	scan.SetHandle("pdf", textconvert.PdfToEs)
	scan.SetHandle("txt", textconvert.NormalToEs)
	scan.SetHandle("csv", textconvert.NormalToEs)

	es, err := textconvert.NewEsCli("", "", os.Args[2])
	if err != nil {
		log.Println(err)
	}

	scan.SetOkHandle(func(res textconvert.Res) {
		// fmt.Println(utils.Yellow(res.Path), utils.Green(len(res.Res.SomeStr)))
		// fmt.Println(res.Res.SomeStr)
		es.BatchingThenImport("test1", res.Res, 100)
	})
	scan.Scan()
	es.Wait("test1")
	// text, err := textconvert.DocxToEs(os.Args[1])
	// fmt.Println(text.Json(), err)
	// fmt.Println(utils.Green(text.SomeStr))
}
