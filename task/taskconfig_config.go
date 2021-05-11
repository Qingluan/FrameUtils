package task

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"

	"github.com/Qingluan/FrameUtils/utils"
)

func try2int(v interface{}) (int, bool) {

	switch v.(type) {
	case float64:
		return int(v.(float64)), true
	case string:
		if f, e := strconv.Atoi(v.(string)); e == nil {
			return f, true
		}
	}
	return -1, false
}

func try2str(v interface{}) (string, bool) {

	switch v.(type) {
	case string:
		return v.(string), true
	}
	return "", false
}

func try2array(v interface{}) (e []string, b bool) {
	switch v.(type) {
	case []interface{}:
		for _, v := range v.([]interface{}) {
			e = append(e, v.(string))
		}
		return e, true
	}
	return []string{}, false
}

func (config *TaskConfig) Copy() (copyConfig *TaskConfig) {
	copyConfig = new(TaskConfig)
	copyConfig.state = make(map[string]TaskState)
	copyConfig.depatch = make(map[string]string)
	copyConfig.procs = make(map[string]string)
	copyConfig.TaskNum = config.TaskNum
	copyConfig.Listen = config.Listen
	copyConfig.LogServer = config.LogServer
	copyConfig.Others = config.Others
	copyConfig.Proxy = config.Proxy
	copyConfig.ReTry = config.ReTry
	copyConfig.LogPathStr = config.LogPathStr
	copyConfig.Schema = config.Schema
	copyConfig.Websocket = config.Websocket
	for k, v := range config.state {
		copyConfig.state[k] = v
	}

	for k, v := range config.depatch {
		copyConfig.depatch[k] = v
	}

	for k, v := range config.procs {
		copyConfig.procs[k] = v
	}
	return
}

func (config *TaskConfig) UrlApiLog(urlOrIp string) (api string) {
	if !strings.HasPrefix(urlOrIp, "http") {
		api = "https://" + urlOrIp
	} else {
		api = urlOrIp
	}
	if !strings.Contains(urlOrIp, ":") {
		urlOrIp += ":4099"
	}
	if strings.Count(api, "/") < 3 {
		api += "/task/v1/log"
	}
	return
}

func (config *TaskConfig) UrlApi(urlOrIp string) (api string) {
	if !strings.HasPrefix(urlOrIp, "https") {
		api = "https://" + urlOrIp
	} else {
		api = urlOrIp
	}
	if !strings.Contains(urlOrIp, ":") {
		api += ":4099"
	}
	if strings.Count(api, "/") < 3 {
		api += "/task/v1/api"
	}
	return
}

func (config *TaskConfig) UpdateRequest(url string, data TData) bool {
	if reply, err := config.ForwardCustom(url, "config", data); err != nil {
		log.Println("update fail:", url, ":", utils.Red(err))
	} else {
		if reply["state"] != "ok" {
			return false
		} else {
			return true
		}
	}
	return false
}

func (config *TaskConfig) SyncAllConfig(allservers string, data TData) (info string) {
	if servers := utils.SplitByIgnoreQuote(allservers, ","); len(servers) > 0 {
		var syncCounter sync.WaitGroup
		iC := len(servers) - 1

		delete(data, "others")
		data["logTo"] = config.MyIP() + ":" + config.MyPort()

		for i, s := range servers {
			// if server != config.MyIP() {
			syncCounter.Add(1)
			log.Println("Sync Data:", utils.Yellow(data))
			go func(i, iC int, server string, datai TData, w *sync.WaitGroup) {
				defer w.Done()

				if config.UpdateRequest(config.UrlApi(server), datai) {
					if !utils.ArrayContains(config.Others, server) {
						log.Println("+ Controller:", utils.Green(server), utils.Yellow(" left ", i, "/", iC))
						info += fmt.Sprintf("+ Controller:%s\n", server)
						config.Others = append(config.Others, server)
					} else {
						log.Println(utils.Red("Fail :", server), utils.Yellow(" left ", i, "/", iC))
					}
				} else {
					log.Println(utils.Red("Fail :", server), utils.Yellow(" left ", i, "/", iC))
				}
			}(i, iC, s, data, &syncCounter)
			// }
		}
		syncCounter.Wait()
	} else {
		info += "no ip in 'others'"
		// return
	}
	log.Println("Server:", config.Others)
	return
}

func (config *TaskConfig) UpdateMyConfig(data TData) (info string) {
	ifsync := false
	allserver := ""

	info = fmt.Sprintf("config : %d", len(data))
	if v, ok := try2str(data["others"]); ok {
		log.Println("Found Other:", utils.Green(v))
		ifsync = true
		allserver = v
	} else {
		if v, ok := try2array(data["others"]); ok {
			ifsync = true
			allserver = strings.Join(v, ",")
		}
	}

	if v, ok := try2str(data["proxy"]); ok {
		info += "\nProxy:" + v
		log.Println("Found Proxy:", utils.Green(v))
		config.Proxy = v
	}
	if v, ok := try2int(data["try"]); ok {
		config.ReTry = v
	}
	if v, ok := try2int(data["taskNum"]); ok {
		config.TaskNum = v
	}
	if v, ok := try2str(data["logTo"]); ok {
		config.LogServer = v
		log.Println("Setting LogTo : ", data["logTo"])
		info += "\nlogTo:" + v
	}
	if ifsync {
		info = config.SyncAllConfig(allserver, data)
	}
	return
}
