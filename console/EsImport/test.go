package main

import (
	"flag"
	"log"

	"github.com/Qingluan/FrameUtils/textconvert"
)

var (
	Host     = "localhost"
	User     = ""
	Pass     = ""
	Dir      = ""
	ScanTask = 0
)

func main() {
	flag.StringVar(&Host, "H", "http://localhost:9200", "set es host")
	flag.StringVar(&User, "u", "", "set es user")
	flag.StringVar(&Pass, "p", "", "set es pass")
	flag.StringVar(&Dir, "d", ".", "set dir")
	flag.IntVar(&ScanTask, "tasknum", 20, "set task num")
	flag.Parse()
	scan := textconvert.NewDirScan(Dir, ScanTask)
	scan.SetHandle("doc", textconvert.DocxToEs)
	scan.SetHandle("docx", textconvert.DocxToEs)
	scan.SetHandle("xlsx", textconvert.XlsxToEs)
	scan.SetHandle("pdf", textconvert.PdfToEs)
	scan.SetHandle("txt", textconvert.NormalToEs)
	scan.SetHandle("csv", textconvert.NormalToEs)

	es, err := textconvert.NewEsCli(User, Pass, Host)
	if err != nil {
		log.Println(err)
	}

	scan.SetOkHandle(func(res textconvert.Res) {
		// fmt.Println(utils.Yellow(res.Path), utils.Green(len(res.Res.SomeStr)))
		// fmt.Println(res.Res.SomeStr)
		es.BatchingThenImport("test1", res.Res, 1000)
	})
	scan.Scan()
	es.Wait("test1")
	// text, err := textconvert.DocxToEs(os.Args[1])
	// fmt.Println(text.Json(), err)
	// fmt.Println(utils.Green(text.SomeStr))
}
