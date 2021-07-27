package engine

import (
	"bufio"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/junhsieh/goexamples/fieldbinding/fieldbinding"
	"github.com/tealeg/xlsx"
)

type MySql struct {
	obj         *sql.DB
	filtertable string
	nowtable    string
	level       string
	tables      []string
}

func ConnectMysql(name, pwd, host, db string) *MySql {
	m := &MySql{}
	m.Open(name, pwd, host, db)
	return m
}

func ConnectByHost(host string) *MySql {
	cli := &MySql{}
	var err error
	cli.obj, err = sql.Open("mysql", host)
	if err != nil {
		panic(err)
	}
	// See "Important settings" section.
	cli.obj.SetConnMaxLifetime(time.Minute * 3)
	cli.obj.SetMaxOpenConns(10)
	cli.obj.SetMaxIdleConns(10)
	return cli
}

func (cli *MySql) Open(name, pwd, host, db string) (err error) {
	cli.obj, err = sql.Open("mysql", name+":"+pwd+"@tcp("+host+")/"+db)
	if err != nil {
		panic(err)
	}
	// See "Important settings" section.
	cli.obj.SetConnMaxLifetime(time.Minute * 3)
	cli.obj.SetMaxOpenConns(10)
	cli.obj.SetMaxIdleConns(10)
	return
}

func (cli *MySql) WriteToExcel(name, sheet string, rows *sql.Rows) (err error) {
	var fArr []string
	fb := fieldbinding.NewFieldBinding()
	if fArr, err = rows.Columns(); err != nil {
		return
	}
	// fmt.Println("Match :", len(fArr))
	fb.PutFields(fArr)
	// outArr := []interface{}{}
	var sh *xlsx.Sheet
	var wb *xlsx.File
	var ok bool
	if _, err = os.Stat(name); err != nil {
		xfile := xlsx.NewFile()
		sh, err = xfile.AddSheet(sheet)
		firstrow := sh.AddRow()
		for i, name := range fArr {
			cell := firstrow.AddCell()

			cell.SetValue(name)
			sh.SetColWidth(i, i, float64(len(name)))
		}
		defer xfile.Save(name)
	} else {
		wb, err = xlsx.OpenFile(name)
		if err != nil {
			return
		}
		sh, ok = wb.Sheet[sheet]
		if !ok {
			sh, err = wb.AddSheet(sheet)
			firstrow := sh.AddRow()
			for i, name := range fArr {
				cell := firstrow.AddCell()
				cell.SetValue(name)
				sh.SetColWidth(i, i, float64(len(name)))
			}
		}

		defer wb.Save(name)
	}

	// maxline := sh.MaxRow
	i := 0
	for rows.Next() {
		if err := rows.Scan(fb.GetFieldPtrArr()...); err != nil {
			return err
		}
		i += 1
		row := sh.AddRow()
		s := ""
		for _, name := range fArr {
			value := fb.Get(name)
			switch value.(type) {
			case nil:
			default:
				cell := row.AddCell()
				cell.SetValue(value)
				s += fmt.Sprint(string(value.([]byte))) + " , "

			}
		}
		if strings.Contains(cli.level, "val") {
			fmt.Println("match: ", s)
		}
		// fmt.Printf("Row: %v, %v, %v, %s\n", fb.Get("IDOrder"), fb.Get("IsConfirm"), fb.Get("IDUser"), fb.Get("Created"))
		// outArr = append(outArr, fb.GetFieldArr())
	}
	return
}

func (cli *MySql) SetLogLevel(loglevel string) {
	cli.level = loglevel
}

func (cli *MySql) MatchFromFile(input, output, table, key string, matchFields ...string) (err error) {

	file, err := os.Open(input)
	if err != nil {
		log.Fatal(err)

	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	// optionally, resize scanner's capacity for lines over 64K, see next example
	tmps := `
	SELECT 
		%s
	FROM
		%s
	WHERE
		%s
	IN
		%s;`
	accounts := []string{}
	var dbExe *sql.Rows
	done := 0
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		if len(accounts) > 300 {
			fileds := "*"
			if matchFields != nil {
				fileds = strings.Join(matchFields, ",")
			}
			matchItems := "('" + strings.Join(accounts, "', '") + "')"

			queryTmp := fmt.Sprintf(tmps, fileds, table, key, matchItems)
			if strings.Contains(cli.level, "sql") {
				log.Println("sql:\n", queryTmp)

			}
			// places := []map[string]interface{}
			dbExe, err = cli.obj.Query(queryTmp)
			if err != nil {
				log.Println("Err :\n", queryTmp)
				return
			}
			err = cli.WriteToExcel(output, table, dbExe)

			done += len(accounts)
			accounts = []string{}
			if strings.Contains(cli.level, "state") {
				log.Println("Ready:", done)
			}

		} else {
			accounts = append(accounts, line)
		}
	}

	done += len(accounts)
	if len(accounts) > 0 {
		fileds := "*"
		if matchFields != nil {
			fileds = strings.Join(matchFields, ",")
		}
		matchItems := "('" + strings.Join(accounts, "', '") + "')"
		queryTmp := fmt.Sprintf(tmps, fileds, table, key, matchItems)
		// places := []map[string]interface{}
		dbExe, err = cli.obj.Query(queryTmp)
		if err != nil {
			log.Println("Err :\n", queryTmp)

			return
		}
		if strings.Contains(cli.level, "state") {
			log.Println("Ready:", done)
		}
		err = cli.WriteToExcel(output, table, dbExe)
	}
	return
}

func (cli *MySql) Close() (err error) {
	err = cli.obj.Close()
	return
}
