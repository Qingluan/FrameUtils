package main

import (
	"flag"
	"log"

	"github.com/Qingluan/FrameUtils/engine"
)

func main() {
	sql := ""
	thread := 5
	batchSize := 5000
	flag.StringVar(&sql, "c", "root:123456@tcp(127.0.0.1:3306)/sms?table=sms&time=1,2,3", "set sql connection string")
	flag.IntVar(&thread, "-t", 10, "set thread num")
	flag.IntVar(&batchSize, "-b", 5000, "set batch size")

	flag.Parse()
	args := flag.Args()
	if len(args) == 1 {
		obj, err := engine.OpenObj(args[0])
		if err != nil {
			log.Fatal(err)
		}
		sqlP := engine.ParseSqlConnectionStr(sql)
		sqlP.Batch = batchSize
		sqlP.Thread = thread
		obj.ToMysql(sqlP)
	}

}
