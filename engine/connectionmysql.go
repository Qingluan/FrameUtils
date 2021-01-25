package engine

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"strings"
)

type MysqlConnection struct {
	BaseConnection
}

func (self *MysqlConnection) AllLine(tables ...string) (reader io.ReadCloser, err error) {
	msg := ""
	for _, table := range tables {
		msg += fmt.Sprintf("select (select \"%s\") as tableName,* from %s ;", table, table)
	}
	reader, err = self.Query(fmt.Sprintf("mysql %s -e '%s'; ", self.Host, msg))
	return
}

func (self *MysqlConnection) AllHeader() {
	msg := "show tables"
	reader, err := self.Query(fmt.Sprintf("mysql %s -e '%s'; ", self.Host, msg))
	if err != nil {
		log.Fatal("all header err:", err)
	}
	lines := bufio.NewReader(reader)
	lines.ReadString(byte('\n'))
	lines.ReadString(byte('\n'))
	lines.ReadString(byte('\n'))
	for {
		tableInfo, err := lines.ReadString(byte('\n'))
		if err == io.EOF {
			break
		}
		if strings.HasPrefix(tableInfo, "+--") {
			break
		}
		table := strings.TrimSpace(tableInfo[1 : len(tableInfo)-1])
		self.tables[table] = Line{}
	}
	for tableName := range self.tables {
		reader, err := self.Query(fmt.Sprintf("mysql %s -e '%s'; ", self.Host, "desc "+tableName+";"))
		if err != nil {
			log.Fatal("read desc "+tableName+" header err:", err)
		}
		headLines := bufio.NewReader(reader)
		headLines.ReadString(byte('\n'))
		headLines.ReadString(byte('\n'))
		headLines.ReadString(byte('\n'))
		fields := Line{}
		for {
			line, err := headLines.ReadString(byte('\n'))
			if err != io.EOF {
				break
			}
			if strings.HasPrefix(line, "+--") {
				break
			}
			field := strings.SplitN(strings.TrimSpace(line[1:]), "|", 2)[0]
			fields = append(fields, field)
		}
		self.tables[tableName] = fields
	}
}

func (self *MysqlConnection) ParseValue(value string) (line Line) {

	line = splitByIgnoreQuote(value[1:len(value)-1], "|")
	return
}
