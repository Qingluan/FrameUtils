package engine

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"strings"

	"github.com/Qingluan/FrameUtils/utils"
)

type SqliteConnection struct {
	BaseConnection
}

func (self *SqliteConnection) AllLine(tables ...string) (reader io.ReadCloser, err error) {
	msg := ""
	for _, table := range tables {
		msg += fmt.Sprintf("select * from %s ;", table)
	}
	reader, err = self.Query(fmt.Sprintf("sqlite3 %s \"%s\"; ", self.Host, msg))
	return
}

func (self *SqliteConnection) ParseValue(value string) (line utils.Line) {
	line = utils.SplitByIgnoreQuote(value, "|")
	return
}

func (self *SqliteConnection) AllHeader() (err error) {
	getheaderstr := ".schema "
	reader, err := self.Query(fmt.Sprintf("sqlite3 %s \"%s\"; ", self.Host, getheaderstr))

	if err != nil {
		log.Fatal("get header err:", err)
	}
	allBuf, err := ioutil.ReadAll(reader)
	if err != nil {
		log.Fatal("read header error")
	}

	for _, headerstr := range utils.SplitByIgnoreQuote(string(allBuf), ");") {
		var line utils.Line

		tableName := Sqlname(strings.Fields(headerstr)[2])
		fieldsPre := strings.SplitN(headerstr, "(", 2)[1]
		// fmt.Println("Header Mid:", fieldsPre)

		for _, field := range utils.SplitByIgnoreQuote(fieldsPre[:strings.LastIndex(fieldsPre, ")")], ",", "()") {
			// fmt.Println("f:", field)
			if l := strings.TrimSpace(field); l != "" {
				fieldName := Sqlname(strings.Fields(l)[0])
				line = append(line, fieldName)
			}

		}
		self.tables[tableName] = line
	}
	return
}
