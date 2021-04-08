package task

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

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

func (config *TaskConfig) DepatchTask(data TData) (reply TData, err error) {
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
