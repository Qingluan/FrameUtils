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
	DefaultTaskWaitChnnael   = make(chan []string, 100)
	DefaultTaskOutputChannle = make(chan string, 100)
)

func NewID(raw string) string {
	// args, _ := utils.DecodeToOptions(raw)
	c := strings.ReplaceAll(raw, " ", "")
	buf := md5.Sum([]byte(c))
	// log.Println("create id by:", utils.Yellow(c))
	return fmt.Sprintf("%x", buf)
}

func IsLocalDomain(ip string) bool {
	if strings.Contains(ip, "127.0.0.1") {
		return true
	}
	if strings.Contains(ip, "localhost") {
		return true
	}
	if strings.Contains(ip, "[::1]") {
		return true
	}
	return false
}
