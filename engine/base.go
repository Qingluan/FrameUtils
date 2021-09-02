package engine

import (
	"github.com/Qingluan/FrameUtils/utils"
	"github.com/fatih/color"
)

var (
	Red   = color.New(color.FgRed).SprintFunc()
	Blue  = color.New(color.FgBlue).SprintFunc()
	Green = color.New(color.FgGreen).SprintFunc()
)

type Base interface {
	header(keylength ...int) utils.Line
	Iter(header ...string) <-chan utils.Line
	Close() error
	Tp() string
}

type Obj interface {
	Search(key string) []utils.Line
	SearchTo(key string, linesChan chan []utils.Line)
	Work(sender chan string, reciver chan []utils.Line)
	Header(k ...int) utils.Line
	GetHeader(k string) utils.Line
	// DiffBy(other Obj, key ...string) []Line
	// GetRow(i int) []Line
	Iter(filterheader ...string) <-chan utils.Line
	Where(filter func(lineno int, line utils.Line, wordno int, word string) bool) (newObj *BaseObj)
	Join(other Obj, opt int, keys ...string) (newObj *BaseObj)
	ToHTML(tableID string, each ...func(row, col int, value string) string) string
	AsJson() []utils.Dict
	Marshal() ([]byte, []string, error)

	WithTmpDB(dbName string) *ObjDatabase
	ToMysql(sql *SqlConnectParm)
}
