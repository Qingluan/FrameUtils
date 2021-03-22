package task

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strconv"

	"github.com/Qingluan/FrameUtils/utils"
	jupyter "github.com/Qingluan/jupyter/http"
)

func (taskConfig *TaskConfig) ProtocolRound(data TData) (reply TData, ido bool, err error) {
	if state := taskConfig.GetMyState(); len(state) > 0 {
		runningNum, _ := strconv.Atoi(state["running"].(string))
		if configstr, ok := state["config"]; ok {
			config := TData{}
			if err := json.Unmarshal([]byte(configstr.(string)), &config); err != nil {
				return nil, true, err
			}
			taskNum, _ := strconv.Atoi(config["taskNum"].(string))
			if runningNum < taskNum {
				ido = true
				return
			}
		}

	}
	reply, err = taskConfig.Forward(data)
	return
}

func (taskconfig *TaskConfig) Forward(data TData) (reply TData, err error) {
	otherServer := taskconfig.NextOthers()
	data["oper"] = "forward"
	return taskconfig.SendToOtherServer(otherServer, data)
}

func (taskconfig *TaskConfig) NextOthers() string {
	taskconfig.taskDipatchCursor = (taskconfig.taskDipatchCursor + 1) % len(taskconfig.Others)
	return taskconfig.Others[taskconfig.taskDipatchCursor]
}

func (taskconfig *TaskConfig) SendToOtherServer(ip string, data TData) (reply TData, err error) {
	var res *jupyter.SmartResponse
	sess := jupyter.NewSession()
	sess.SetSocks5Proxy(taskconfig.Proxy)
	sendData := utils.BDict{}
	for k, v := range data {
		switch v.(type) {
		case string:

			sendData[k] = v.(string)
		default:
			sendData[k] = fmt.Sprintf("%v", v)
		}
	}

	if res, err = sess.Json(fmt.Sprintf("%s://%s/task/v1/api", taskconfig.Schema, ip), sendData); err != nil {
		return reply, err
	} else {
		data, _ := ioutil.ReadAll(res.Body)
		err = json.Unmarshal(data, reply)
	}
	return
}
