package task

import (
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

func (config *TaskConfig) UrlApi(urlOrIp string) (api string) {
	if !strings.HasPrefix(urlOrIp, "http") {
		api = "http://" + urlOrIp
	} else {
		api = urlOrIp
	}
	if strings.Count(api, "/") < 3 {
		api += "/task/v1/api"
	}
	return
}

func (config *TaskConfig) UpdateRequest(url string, data TData) bool {
	if reply, err := config.ForwardCustom(url, "config", data); err != nil {
		log.Println("update fail:", utils.Red(err))
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
	if servers := utils.SplitByIgnoreQuote(allservers, ","); utils.ArrayContains(servers, config.MyIP()) {
		var syncCounter sync.WaitGroup
		iC := len(servers) - 1
		for i, server := range servers {
			if server != config.MyIP() {
				syncCounter.Add(1)
				go func(i, iC int, datai TData) {
					defer syncCounter.Done()
					delete(datai, "others")
					if config.UpdateRequest(server, datai) {
						if !utils.ArrayContains(config.Others, server) {
							config.Others = append(config.Others, server)
						}
					}
					log.Println("Wait update all servers:", utils.Yellow("left ", i, "/", iC), " waiting...")
				}(i, iC, data)
			}
		}
		syncCounter.Wait()
	} else {
		info += "no myip include in 'others'"
		// return
	}
	return
}

func (config *TaskConfig) UpdateMyConfig(data TData) (info string) {
	ifsync := false
	allserver := ""
	if v, ok := try2str(data["others"]); ok {
		ifsync = true
		allserver = v
	}

	if v, ok := try2str(data["proxy"]); ok {
		config.Proxy = v
	}
	if v, ok := try2int(data["try"]); ok {
		config.ReTry = v
	}
	if v, ok := try2int(data["taskNum"]); ok {
		config.TaskNum = v
	}
	if ifsync {
		config.SyncAllConfig(allserver, data)
	}
	return
}
