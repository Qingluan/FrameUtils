package task

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"strconv"
	"strings"

	"github.com/Qingluan/FrameUtils/utils"
	jupyter "github.com/Qingluan/jupyter/http"
	"github.com/Qingluan/merkur"
)

func (taskConfig *TaskConfig) ProtocolRound(data TData) (reply TData, ido bool, err error) {
	if state := taskConfig.GetMyState(); len(state) > 0 {
		runningNum, _ := strconv.Atoi(state["running"].(string))
		if configstr, ok := state["config"]; ok {
			config := TData{}
			if err := json.Unmarshal([]byte(configstr.(string)), &config); err != nil {
				return nil, true, err
			}
			switch config["taskNum"].(type) {
			case float64:
				// taskNum, _ := strconv.Atoi(.(string))
				if runningNum < int(config["taskNum"].(float64)) {
					ido = true
					return
				}
			case string:
				taskNum, _ := strconv.Atoi(config["taskNum"].(string))
				if runningNum < taskNum {
					ido = true
					return
				}
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

func (taskconfig *TaskConfig) ForwardCustom(url string, oper string, data TData) (reply TData, err error) {
	data["oper"] = oper
	return taskconfig.SendToOtherServer(url, data)
}

func (taskconfig *TaskConfig) NextOthers() string {
	taskconfig.taskDipatchCursor = (taskconfig.taskDipatchCursor + 1) % len(taskconfig.Others)
	return taskconfig.Others[taskconfig.taskDipatchCursor]
}

func (taskconfig *TaskConfig) SendToOtherServer(ip string, data TData) (reply TData, err error) {
	var res *jupyter.SmartResponse
	sess := jupyter.NewSession()
	if taskconfig.Proxy != "" && !IsLocalDomain(ip) {
		if pdialer := merkur.NewProxyDialer(taskconfig.Proxy); pdialer != nil {
			log.Println(utils.Red("set proxy:", taskconfig.Proxy))

			sess.SetProxyDialer(pdialer)
		} else {
			log.Println(utils.Red("set proxy:", taskconfig.Proxy), " failed!! use default direct connect!")
		}
	}
	sendData := utils.Dict{}
	for k, v := range data {
		switch v.(type) {
		case string:

			sendData[k] = v.(string)
		default:
			sendData[k] = fmt.Sprintf("%v", v)
		}
	}
	api := taskconfig.UrlApi(ip)
	if res, err = sess.Json(api, sendData); err != nil {
		if strings.Contains(err.Error(), "server gave HTTP response to HTTPS client") {
			log.Println("[DEBUG] may url is http :", utils.Yellow(api))
			api = "http://" + strings.SplitN(api, "://", 2)[1]
			res, err = sess.Json(api, sendData)
			data, _ := ioutil.ReadAll(res.Body)
			err = json.Unmarshal(data, &reply)
			if err != nil {
				log.Println("SendToOtherServer UnJson err:", err, string(data))
			}
		} else {
			return
		}
		return reply, err
	} else {
		data, _ := ioutil.ReadAll(res.Body)
		err = json.Unmarshal(data, &reply)
		if err != nil {
			log.Println("SendToOtherServer UnJson err:", err, string(data))
		}
	}
	return
}
