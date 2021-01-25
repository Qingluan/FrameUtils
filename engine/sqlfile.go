package engine

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"strings"
)

const (
	TP_MYSQL     = 0
	TP_SQLSERVER = 1
	TP_SQLITE    = 2
)

type SqlTxt struct {
	obj          *bufio.Scanner
	raw          string
	headers      map[string]Line
	datas        map[string][]Dict
	cacheLines   map[string][]Line
	sqlType      int
	sqlLineEnd   string
	nowheader    string
	filterheader string
}

func Sqlname(a string) string {
	if len(a) < 2 {
		return a
	}
	if strings.HasPrefix(a, "`") && strings.HasSuffix(a, "`") {
		return a[1 : len(a)-1]
	} else if strings.HasPrefix(a, "'") && strings.HasSuffix(a, "'") {
		return a[1 : len(a)-1]
		// sqlserver name : example
	} else if strings.HasPrefix(a, "N'") && strings.HasSuffix(a, "'") {
		if len(a) < 3 {
			fmt.Println("bug a:", a)
		}
		return a[2 : len(a)-1]
	}
	return a
}

func (self *SqlTxt) ParseSqlValue(v string) (tableName string, line Line) {

	fsss := strings.Fields(v)
	if len(fsss) < 2 {
		log.Fatal("Fatal:::", v)
	}
	tableName = Sqlname(fsss[2])
	fieldsPre := ""
	switch self.sqlType {
	case TP_SQLSERVER:
		// fieldsPre = strings.SplitN(string(v), "VALUES", 2)[1]
		tmps := strings.SplitN(string(v), "(", 3)
		if len(tmps) < 3 {
			log.Fatal("Err:", tmps, "|", v)
		}
		fieldsPre = tmps[2]
	default:
		fieldsPre = strings.SplitN(string(v), "(", 2)[1]
	}

	if !strings.Contains(fieldsPre, ")") {
		log.Fatal("Err not found:", v)
		// return "", ""
	}
	for _, field := range splitByIgnoreQuote(fieldsPre[:strings.LastIndex(fieldsPre, ")")], ",") {
		if l := strings.TrimSpace(field); l != "" {
			if l == "N'" {
				fmt.Println("Bug line:", v)
			}
			line = append(line, Sqlname(l))
		}
	}

	// fmt.Println("H:", tableName, "V:", line)
	return

}
func (self *SqlTxt) GetHead(k string) Line {
	header := self.headers[k]
	if header != nil {
		// d := values.FromKey(header)
		return header
	}
	return nil
}

func (self *SqlTxt) ParseSqlHeader(v string) (tableName string, line Line) {

	tableName = Sqlname(strings.Fields(v)[2])
	if strings.Contains(tableName, "[dbo]") {
		self.sqlType = TP_SQLSERVER
	}
	fieldsPre := strings.SplitN(v, "(", 2)[1]
	// fmt.Println("Header Mid:", fieldsPre)
	for _, field := range splitByIgnoreQuote(fieldsPre[:strings.LastIndex(fieldsPre, ")")], ",", "()") {
		// fmt.Println("f:", field)
		if l := strings.TrimSpace(field); l != "" {
			fieldName := Sqlname(strings.Fields(l)[0])
			line = append(line, fieldName)
		}
	}
	// log.Fatal(fieldsPre, "\ntest:", line)
	self.nowheader = tableName
	// fmt.Println("Header Name:", tableName)
	// fmt.Println("Header End:", line)
	self.headers[tableName] = line
	return

}

func (self *SqlTxt) switchSqlTp(data []byte) {
	if bytes.Contains(data, []byte("CREATE TABLE [dbo]")) {
		self.sqlLineEnd = "GO"
		self.sqlType = TP_SQLSERVER
	}
}

func (self *SqlTxt) Iter(header ...string) <-chan Line {
	ch := make(chan Line)
	self.obj = bufio.NewScanner(strings.NewReader(self.raw))
	if self.headers == nil {
		self.headers = make(map[string]Line)
	}

	if header != nil {
		self.filterheader = header[0]
	}
	if len(self.cacheLines) == 0 {
		// fmt.Println("--- 0")
		// all := 0
		self.cacheLines = make(map[string][]Line)
		self.obj.Split(func(data []byte, atEOF bool) (advance int, token []byte, err error) {
			if atEOF && len(data) == 0 {
				return 0, nil, nil
			}
			if atEOF {
				return len(data), data, io.EOF
			}
			self.switchSqlTp(data)

			if e := bytes.Index(data, []byte(self.sqlLineEnd)); e > 0 {

				if cs := bytes.Index(data[:e+2], []byte("CREATE TABLE")); cs > 0 {
					self.ParseSqlHeader(string(data[cs : e+2]))

					// return e + 1, data[cs : e+1], nil
				}

				if cs := bytes.Index(data[:e+2], []byte("INSERT")); cs > 0 {
					// fmt.Println(string(data[cs : e+1]))
					buf := data[cs : e+2]
					if !bytes.Contains(buf, []byte("(")) || !bytes.Contains(buf, []byte(")")) {
						// jump like sqlserver : INSERT [dbo].[xxx] ON
						return e + 2, nil, nil
					}
					// if all%100000 == 0 {
					// 	logrus.Infof("count : %d/\n", all)
					// }
					// all++
					return e + 1, buf, nil
				}

				return e + 2, nil, nil

				// return 0, nil, nil
			}

			return 0, nil, nil
		})

		go func() {
			c := 0
			for self.obj.Scan() {
				line := strings.TrimSpace(self.obj.Text())
				// fmt.Println(line)
				tbName, l := self.ParseSqlValue(line)
				if self.filterheader != "" && tbName != self.filterheader {
					continue
				}
				iterLine := append(Line{tbName}, l...)
				if as, ok := self.cacheLines[tbName]; ok {
					as = append(as, iterLine)
					self.cacheLines[tbName] = as
				} else {
					self.cacheLines[tbName] = []Line{iterLine}
				}

				ch <- iterLine
				c++
			}
			close(ch)
		}()

	} else {
		// fmt.Println("--- 1")

		go func() {
			for _, vs := range self.cacheLines {
				for _, l := range vs {
					ch <- l
				}
			}
			close(ch)
		}()

	}
	return ch

}

func (self *SqlTxt) Close() error {
	// return self.obj.Close()
	return nil
}

func (self *SqlTxt) header(k ...int) (l Line) {
	return
}

func (self *SqlTxt) ToJson() (ds []Dict) {
	for line := range self.Iter() {
		// fmt.Println("tb:", line)
		tb, values := line[0], line[1:]
		// fmt.Println("tb:", tb)
		if header, ok := self.headers[tb]; ok {
			// fmt.Println(tb, header)
			d := values.FromKey(header)
			ds = append(ds, d)
		}
	}
	return
}

func (s *SqlTxt) Tp() string {
	return "sqltxt"
}
