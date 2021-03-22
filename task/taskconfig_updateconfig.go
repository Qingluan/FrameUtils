package task

import (
	"encoding/json"
	"log"

	"github.com/Qingluan/FrameUtils/utils"
)

func (config *TaskConfig) Update(v interface{}) (reply TData) {
	msg := ""
	data := TData{}
	var err error
	switch v.(type) {
	case map[string]interface{}:
		data = v.(TData)
	case string:
		err = json.Unmarshal([]byte(v.(string)), &data)
	case []byte:
		err = json.Unmarshal(v.([]byte), &data)
	default:
		reply = TData{
			"state": "fail",
			"log":   "not support type",
		}
		return
	}
	if err != nil {
		reply = TData{
			"state": "fail",
			"log":   err.Error(),
		}
	}
	if v, ok := data["others"]; ok {
		config.Others = v.([]string)
		log.Println(utils.Blue("update others="), utils.Green(v))
		msg += "other"
	}

	if v, ok := data["proxy"]; ok {
		config.Proxy = v.(string)
		log.Println(utils.Blue("update proxy="), utils.Green(v))
		msg += "|proxy"
	}
	reply = TData{
		"state": "ok",
		"log":   msg,
	}
	return

}
