package task

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/Qingluan/FrameUtils/utils"
)

func (config *TaskConfig) SaveState() {
	Home, _ := os.Hostname()
	state := filepath.Join(Home, ".config", "task-map.json")
	data, _ := json.Marshal(config.depatch)
	ioutil.WriteFile(state, data, os.ModePerm)
}
func (config *TaskConfig) LoadState() {
	Home, _ := os.Hostname()
	state := filepath.Join(Home, ".config", "task-map.json")

	if data, err := ioutil.ReadFile(state); err == nil {
		json.Unmarshal(data, &config.depatch)
	}
}

func (config *TaskConfig) depatchTask(data TData) (reply TData, err error) {
	reply = make(TData)
	if _, ok := data["input"]; !ok {
		reply["state"] = "fail"
		reply["log"] = "lack 'input'"
		return
	}
	if _, ok := data["tp"]; !ok {
		reply["state"] = "fail"
		reply["log"] = "lack 'tp'"
		return
	}
	server := utils.RandomChoice(config.Others)
	data["input"] = data["input"].(string) + fmt.Sprintf(" , logTo=\"%s\"", config.MyIP())
	reply, err = config.ForwardCustom(config.UrlApi(server), "pull", data)
	return
}

func (config *TaskConfig) DepatchTask(line string) (reply TData, err error) {

	args := strings.SplitN(line, ",", 2)
	if len(args) != 2 {
		err = fmt.Errorf("args is not valid!!:%s", line)
		return
	}
	t := TData{
		"oper":  "push",
		"input": strings.TrimSpace(args[1]),
		"tp":    strings.TrimSpace(args[0]),
	}

	if len(config.Others) == 0 {
		patch(line)
		reply = make(TData)
		reply["log"] = "run in local:" + line
		reply["state"] = "wait"
		return
	}
	return config.depatchTask(t)
}

func patch(line string) {
	args := strings.SplitN(line, ",", 2)
	if len(args) == 2 {
		DefaultTaskWaitChnnael <- []string{strings.TrimSpace(args[0]), strings.TrimSpace(args[1])}
	} else if len(args) == 1 {
		DefaultTaskWaitChnnael <- []string{strings.TrimSpace(args[0])}
	}
}

func (config *TaskConfig) DealWithUploadFile(w http.ResponseWriter, h *http.Request) {
	if h.Method == "POST" {
		f, _, err := h.FormFile("uploadFile")
		if err != nil {
			jsonWriteErr(w, err)
			return
		}
		buffer := bufio.NewScanner(f)
		buffer.Split(bufio.ScanLines)
		runOk := 0
		waitTaskLines := []string{}
		for buffer.Scan() {
			line := buffer.Text()
			lineStr := strings.TrimSpace(line)
			// fmt.Println(lineStr, "|")
			if strings.HasPrefix(lineStr, "http") {
				fmt.Println(utils.Green("[http]", lineStr))
				waitTaskLines = append(waitTaskLines, "http,"+lineStr)
			} else if strings.HasPrefix(lineStr, "tcp://") {
				fmt.Println(utils.Blue("[tcp]", lineStr))
			} else if strings.HasPrefix(lineStr, "cmd,") {
				fmt.Println(utils.Yellow("[cmd]", lineStr))
				waitTaskLines = append(waitTaskLines, lineStr)
			} else if strings.HasPrefix(lineStr, "config,") {
				fmt.Println(utils.Yellow("[config]", lineStr))
				patch(lineStr)
			} else {
				fmt.Println("[ignore]", lineStr)
				runOk -= 1
			}
		}
		config.CheckAlive(config.Others...)
		for _, waitTask := range waitTaskLines {
			config.DepatchTask(waitTask)
		}
		jsonWrite(w, TData{
			"log":   fmt.Sprintf("%d", runOk),
			"state": "ok",
		})
	}
}
