package engine

import (
	"bufio"
	"bytes"
	"io"
	"log"
	"strings"

	"github.com/sirupsen/logrus"
)

const (
	TP_MYSQL = 0
	TP_SQLSERVER = 1
	TP_SQLITE = 2
)

type SqlTxt struct {
	obj        *bufio.Scanner
	raw        string
	headers    map[string]Line
	datas      map[string][]Dict
	cacheLines map[string][]Line
	sqlType int
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
	} else if strings.HasPrefix(a, "N'") && strings.HasSuffix(a, "'"){
		return a[2: len(a) -1]
	}
	return a
}

func (self *SqlTxt) ParseSqlValue(v string) (tableName string, line Line) {

	fsss := strings.Fields(v)
	if len(fsss) < 2 {
		log.Fatal("Fatal:::", v)
	}
	tableName = Sqlname(fsss[2])
	fieldsPre := strings.SplitN(string(v), "(", 2)[1]
	if !strings.Contains(fieldsPre, ")") {
		log.Fatal("Err not found:", v)
		// return "", ""
	}
	for _, field := range strings.Split(fieldsPre[:strings.LastIndex(fieldsPre, ")")], ",") {
		if l := strings.TrimSpace(field); l != "" {
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

	// fmt.Println("Header:", v)
	defer func(){
		switch self.sqlType{
		case TP_SQLSERVER:
			self.sqlLineEnd = "GO"
		default:
			self.sqlLineEnd = ");"
	
		}
	}()
	tableName = Sqlname(strings.Fields(v)[2])
	if strings.Contains(tableName, "[dbo]"){
		self.sqlType = TP_SQLSERVER
	}
	fieldsPre := strings.SplitN(v, "(", 2)[1]
	// fmt.Println("Header Mid:", fieldsPre)

	for _, field := range strings.Split(fieldsPre[:strings.LastIndex(fieldsPre, ")")], ",") {
		// fmt.Println("f:", field)
		if l := strings.TrimSpace(field); l != "" {
			fieldName := Sqlname(strings.Fields(l)[0])
			line = append(line, fieldName)
		}
	}
	// fmt.Println("Header Name:", tableName)
	// fmt.Println("Header End:", line)
	self.headers[tableName] = line
	return

}

func (self *SqlTxt) Iter() <-chan Line {
	ch := make(chan Line)
	self.obj = bufio.NewScanner(strings.NewReader(self.raw))
	if self.headers == nil {
		self.headers = make(map[string]Line)
	}

	if len(self.cacheLines) == 0 {
		// fmt.Println("--- 0")
		all := 0
		self.cacheLines = make(map[string][]Line)
		self.obj.Split(func(data []byte, atEOF bool) (advance int, token []byte, err error) {
			if atEOF && len(data) == 0 {
				return 0, nil, nil
			}
			if atEOF {
				return len(data), data, io.EOF
			}
			if e := bytes.Index(data, []byte(self.sqlLineEnd)); e > 0 {
				if cs := bytes.Index(data[:e+2], []byte("CREATE TABLE")); cs > 0 {
					self.ParseSqlHeader(string(data[cs : e+2]))

					// return e + 1, data[cs : e+1], nil
				}

				if cs := bytes.Index(data[:e+2], []byte("INSERT")); cs > 0 {
					// fmt.Println(string(data[cs : e+1]))
					if all%10000 == 0 {
						logrus.Infof("count : %d/\n", all)
					}
					all++
					return e + 1, data[cs : e+2], nil
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
