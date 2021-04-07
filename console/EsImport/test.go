package main

import (
	"flag"
	"log"

	"github.com/Qingluan/FrameUtils/textconvert"
)

var (
	Host      = "localhost"
	User      = ""
	Pass      = ""
	Dir       = ""
	ScanTask  = 0
	BatchSize = 0
	indexName = ""
)

func main() {
	flag.StringVar(&Host, "H", "http://localhost:9200", "set es host")
	flag.StringVar(&User, "u", "", "set es user")
	flag.StringVar(&Pass, "p", "", "set es pass")
	flag.StringVar(&Dir, "d", ".", "set dir")
	flag.StringVar(&indexName, "i", "Name1", "set name")
	flag.IntVar(&ScanTask, "tasknum", 100, "set task num")
	flag.IntVar(&BatchSize, "batch", 2000, "set batch size")
	flag.Parse()
	scan := textconvert.NewDirScan(Dir, ScanTask)
	// scan.SetHandle("doc", textconvert.DocxToEs)
	scan.SetHandle("docx", textconvert.DocxToEs)
	scan.SetHandle("xlsx", textconvert.XlsxToEs)
	// scan.SetHandle("pdf", textconvert.PdfToEs)
	scan.SetHandle("txt", textconvert.NormalToEs)
	scan.SetHandle("csv", textconvert.NormalToEs)

	es, err := textconvert.NewEsCli(User, Pass, Host)
	if err != nil {
		log.Println("err:", err)
	}

	scan.SetOkHandle(func(res textconvert.Res) {
		// fmt.Println(utils.Yellow(res.Path), utils.Green(len(res.Res.SomeStr)))
		// fmt.Println(res.Res.SomeStr)
		es.BatchingThenImport(indexName, res.Res, 2000)
	})
	scan.Scan()
	es.Wait(indexName)
	// text, err := textconvert.DocxToEs(os.Args[1])
	// fmt.Println(text.Json(), err)
	// fmt.Println(utils.Green(text.SomeStr))
}
