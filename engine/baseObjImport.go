package engine

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
	"github.com/schollz/progressbar/v3"
)

var (
	GREEN  = color.New(color.FgGreen, color.Bold).SprintFunc()
	YELLOW = color.New(color.FgYellow).SprintFunc()
)

func ParseTime(timestamp string) string {
	i, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		panic(err)
	}
	tm := time.Unix(i, 0)
	return tm.Format("2006-01-02 15:04:05")

}

func (p *SqlConnectParm) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s)/%s", p.User, p.Pwd, p.Host, p.DB)
}

func ParseTimeAsColumns(p string, line []string) []string {
	if p == "" {
		return line
	}
	// nLines := []string{}
	for _, pi := range strings.Split(p, ",") {
		if i, er := strconv.Atoi(pi); er == nil && len(line) > i {
			line[i] = ParseTime(line[i])
		}
	}
	return line

}

/*SqlConnectParm
如果要解析的列需要进行时间戳解析，设置
	TimeParseCol  如 ： "1,2" #解析第二列和第三列
*/
type SqlConnectParm struct {
	User         string
	Pwd          string
	Host         string
	DB           string
	Table        string
	PrePare      string
	Batch        int
	Thread       int
	TimeParseCol string
}

/*ParseSqlConnectionStr

user:pwd@tcp(addr)/db?table
*/
func ParseSqlConnectionStr(s string) (sql *SqlConnectParm) {
	sql = new(SqlConnectParm)
	user := "root"
	pwd := ""
	db := ""
	table := ""
	addr := "127.0.0.1:3306"
	args := ""
	if !strings.Contains(s, "?") {
		log.Fatal("must include a table like :user:pwd@tcp(addr)/db?table=tableName")
	}

	if strings.Contains(s, "@") {
		fs := strings.SplitN(s, "@", 2)
		if strings.Contains(fs[0], ":") {
			fs2 := strings.SplitN(fs[0], ":", 2)
			user = strings.TrimSpace(fs2[0])
			pwd = strings.TrimSpace(fs2[1])
		} else {
			pwd = fs[0]
		}
		if strings.Contains(fs[1], "tcp(") {
			_t := strings.SplitN(strings.SplitN(fs[1], "tcp(", 2)[1], ")", 2)
			addr = _t[0]
			db = strings.SplitN(_t[1], "/", 2)[1]
			if strings.Contains(db, "?") {
				_t2 := strings.SplitN(db, "?", 2)
				db = _t2[0]
				args = _t2[1]
			}
		} else {
			addr = strings.SplitN(fs[1], "/", 2)[0]
			db = strings.SplitN(fs[1], "/", 2)[1]
			if strings.Contains(db, "?") {
				_t2 := strings.SplitN(db, "?", 2)
				db = _t2[0]
				args = _t2[1]
			}
		}

	} else {
		if strings.Contains(s, "tcp(") {
			_t := strings.SplitN(strings.SplitN(s, "tcp(", 2)[1], ")", 2)
			addr = _t[0]
			db = strings.SplitN(_t[1], "/", 2)[1]
			if strings.Contains(db, "?") {
				_t2 := strings.SplitN(db, "?", 2)
				db = _t2[0]
				table = _t2[1]
			}
		}
	}
	for _, v := range strings.Split(args, "&") {
		if strings.Contains(v, "=") {
			kv := strings.SplitN(v, "=", 2)
			if kv[0] == "table" {
				table = strings.TrimSpace(kv[1])
			} else if kv[0] == "time" {
				sql.TimeParseCol = strings.TrimSpace(kv[1])
			}
		}
	}
	sql.User = user
	sql.Pwd = pwd
	sql.Host = addr
	sql.DB = db
	sql.Table = table
	if table == "" {
		log.Fatal("must include a table like :user:pwd@tcp(addr)/db?table ")
	}
	return

}

func (obj *BaseObj) ToMysql(sql *SqlConnectParm) {
	header := []string{}
	i := 0
	HEADER := "INSERT INTO " + sql.Table
	H := "(`"
	log.Println("use sql config:", GREEN(sql.DSN()), " table:", YELLOW(sql.Table), GREEN("insert batch size:", sql.Batch))

	allnum := 0
	log.Print(GREEN("Scan all line...\r"))
	for range obj.Iter() {
		allnum += 1
	}
	bar := Showbar(allnum, GREEN(sql.DSN()))
	bactch := [][]interface{}{}
	continueCounter := sync.WaitGroup{}
	// waitChan := make(chan int, 3)
	// doingChan := make(chan int, thread)
	running := 0
	bnum := 0
	// now := time.Now()
	for line := range obj.Iter() {
		i += 1
		if i == 1 {
			preH := ""
			for _, item := range line[1:] {
				header = append(header, item)
				H += strings.TrimSpace(item) + "`,`"
				preH += "?,"
			}
			H = H[:len(H)-2] + ") "
			sql.PrePare = HEADER + H + " VALUES(" + preH[:len(preH)-1] + ")"

			log.Println("use sql prepare:", GREEN(sql.PrePare))
			continue
		}

		/// dealed herer , include parse time
		values := []interface{}{}
		if sql.TimeParseCol != "" {
			for _, v := range ParseTimeAsColumns(sql.TimeParseCol, line[1:]) {
				values = append(values, v)
			}
		} else {
			for _, v := range line[1:] {
				values = append(values, v)
			}
		}
		///
		// t := HEADER + H + " VALUES(" + "'" + strings.Join(values, "','") + "'" + ");"
		// fmt.Println(GREEN(t))

		bactch = append(bactch, values)
		if len(bactch) > 0 && len(bactch)%sql.Batch == 0 {

			// for {
			// LA:
			// 	select {
			// 	case doingChan <- 0:
			// 		// continueCounter.Add(1)
			// 		// log.Println("Async run")
			// 		running += 1
			// 		go Upload(sql, bactch, doingChan)
			// 		bactch = [][]interface{}{}
			// 		break LA
			// 	default:
			// 		fmt.Printf("%s\r", "wait ....")
			// 		time.Sleep(1 * time.Second)
			// 	}

			// }
			// log.Println("start insert :", bnum*len(bactch), "-", (bnum+1)*len(bactch), time.Now().Format("2006-01-02 15:04:05"))
			if running == sql.Thread {
				continueCounter.Wait()
				running = 0
				// now = time.Now()
			}
			continueCounter.Add(1)
			go Upload(sql, bactch, &continueCounter, bar)
			running += 1
			bnum += 1
			bactch = [][]interface{}{}

		}
	}

	if len(bactch) > 0 {
		// doingChan <- 0
		continueCounter.Add(1)
		Upload(sql, bactch, &continueCounter, bar)
	}
	fmt.Println("wait stop!!!")
}

func Upload(sqlPar *SqlConnectParm, batch [][]interface{}, wait *sync.WaitGroup, bar *progressbar.ProgressBar) int64 {
	now := time.Now()
	var fi int64 = int64(len(batch))
	defer func() {
		// <-doing
		wait.Done()
	}()

	db, err := sql.Open("mysql", sqlPar.DSN())
	if err != nil {
		log.Printf("Error %s when opening DB\n", err)
		return -1
	}
	defer db.Close()

	ctx, cancelfunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelfunc()

	insert, err := db.PrepareContext(ctx, sqlPar.PrePare)
	if err != nil {
		log.Println(YELLOW(err.Error()))
	}
	begin, err := db.Begin()
	if err != nil {
		log.Println(YELLOW(err.Error()))
	}
	for _, line := range batch {
		_, err := begin.Stmt(insert).Exec(line...)
		if err != nil {
			log.Println(YELLOW(err.Error()))
		}
	}
	err = begin.Commit()
	if err != nil {
		log.Println(YELLOW(err.Error()))
	}
	// db.
	// log.Println("connect success:", GREEN(sqlPar.DSN()))

	// db.Exec(strings.Join(batch, ";"))
	// bar.Describe(fmt.Sprint("Upload :", len(batch), GREEN("used :", time.Now().Sub(now)), YELLOW("insert :", fi), " "))
	desc := GREEN(fmt.Sprintf("batch:%d thread:%d|use:%v", sqlPar.Batch, sqlPar.Thread, time.Now().Sub(now)))
	bar.Describe(desc)
	bar.Add(sqlPar.Batch)

	return fi
}

func Showbar(num int, pre string) *progressbar.ProgressBar {
	bar := progressbar.NewOptions(num,
		// progressbar.OptionSetWriter(os.NewAnsiStdout()),
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionShowBytes(true),
		progressbar.OptionSetWidth(15),
		progressbar.OptionSetDescription(pre),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "[green]=[reset]",
			SaucerHead:    "[green]>[reset]",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}))
	return bar
}
