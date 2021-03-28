package task

import (
	"crypto/md5"
	"fmt"
	"log"
	"strings"

	"github.com/Qingluan/FrameUtils/utils"
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

func NewID(raw string) string {
	// args, _ := utils.DecodeToOptions(raw)
	c := strings.ReplaceAll(raw, " ", "")
	buf := md5.Sum([]byte(c))
	log.Println("create id by:", utils.Yellow(c))
	return fmt.Sprintf("%x", buf)
}
