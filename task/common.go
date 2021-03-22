package task

import (
	"crypto/md5"
	"fmt"
	"strings"
)

const (
	DefaultTaskConfigString = `
[default]

taskNum = 100
listen = :4099
logserver = http://localhost:4099/task/v1/log
try = 3
schema = http
#logPath = 
#others = 
#proxy = 

	`
)

var (
	DefaultTaskWaitChnnael   = make(chan []string, 24)
	DefaultTaskOutputChannle = make(chan string, 36)
)

func NewID(args []string) string {
	c := strings.ReplaceAll(strings.Join(args, " "), " ", "")
	buf := md5.Sum([]byte(c))
	// log.Println("create id by:", color.New(color.FgYellow).Sprint(c))
	return fmt.Sprintf("%x", buf)
}
