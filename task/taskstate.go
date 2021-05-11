package task

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/Qingluan/FrameUtils/asset"
	"github.com/Qingluan/FrameUtils/utils"
)

type TaskState struct {
	DeployedServer string     `json:"deployed_server"`
	Args           []string   `json:"args"`
	Kargs          utils.Dict `json:"kargs"`
	ID             string     `json:"id"`
	State          string     `json:"state"`
	PID            string     `json:"pid"`
	LogSize        string     `json:"log_size"`
	LogLast        string     `json:"log_last"`
}

func NewTaskState(id, deployed_server string) *TaskState {
	return &TaskState{
		ID:             id,
		DeployedServer: deployed_server,
	}
}

func String2TaskState(js string) (*TaskState, error) {
	t := new(TaskState)
	if err := json.Unmarshal([]byte(js), t); err != nil {
		return nil, err
	}
	return t, nil
}

func (taskState *TaskState) String() string {
	bf, _ := json.Marshal(taskState)
	return string(bf)
}

/* DeploytSwitchState
改变部署任务的状态
*/
func (config *TaskConfig) DeployedSwitchState(id string, state string) {
	config.lock.Lock()
	defer config.lock.Unlock()

	// fmt.Println("Switching....")
	// defer fmt.Println("Switched")
	if e, ok := config.state[id]; ok {
		e.State = state
		path := filepath.Join(config.LogPath(), e.ID) + ".log"
		if state, err := os.Stat(path); err == nil {
			e.LogSize = fmt.Sprintf("%fMB", float64(state.Size())/float64(1024*1024))
			e.LogLast = state.ModTime().Local().String()
		}
		config.state[id] = e

		config.state[id] = e
		log.Println(utils.Magenta("[DeploySwitch] : ", id), " IN :", utils.Yellow(e.DeployedServer), " => ", utils.Green(state))
		config.Websocket.MsgChanel <- TData{
			"id":    id,
			"tp":    "updatelog",
			"value": e,
		}
	} else {

		log.Println(" Switch Not found this task in TaskState:", utils.Red(id))
	}
}

/*
保存部署任务的大部分状态
*/
func (config *TaskConfig) DeploySaveState(id string, useServer string, input string) {
	config.lock.Lock()
	defer config.lock.Unlock()
	args, kargs := utils.DecodeToOptions(input)
	log.Println(utils.Magenta("[Deploy] : ", id), " IN :", utils.Yellow(useServer))

	if s, ok := config.state[id]; ok {
		s.Args = args
		s.Kargs = kargs
		s.DeployedServer = useServer
		config.state[strings.TrimSpace(id)] = s
	} else {
		config.state[id] = TaskState{
			ID:             strings.TrimSpace(id),
			State:          "Depatching",
			Args:           args,
			Kargs:          kargs,
			DeployedServer: useServer,
		}
	}

}

func (config *TaskConfig) DeployedTaskGet(id string) (TaskState, bool) {
	if e, ok := config.state[id]; ok {
		return e, ok
	} else {
		return TaskState{}, false
	}
}

/*
查找部署任务信息
*/
func (config *TaskConfig) DeployStateFind(key string) (states []TaskState) {
	// config.lock.Lock()

	for id, v := range config.state {
		if key == "" {
			states = append(states, v)
			continue
		}
		if strings.Contains(v.String(), key) {
			states = append(states, v)
			log.Println("Searching task found:", id)
		}
	}
	// fmt.Println(states)
	return
}

/* DeploytSwitchState
更新部署任务的日志文件状态
*/
// func (config *TaskConfig) DeployedSaveLogState(id string) {
// 	config.lock.Lock()
// 	defer config.lock.Unlock()

// 	if e, ok := config.state[id]; ok {
// 		// 獲取任務日志文件信息
// 		path := filepath.Join(config.LogPath(), id) + ".log"
// 		if state, err := os.Stat(path); err == nil {
// 			e.LogSize = fmt.Sprintf("%fMB", float64(state.Size())/float64(1024*1024))
// 			e.LogLast = state.ModTime().Local().String()
// 		}
// 		config.state[id] = e
// 	} else {
// 		log.Println("Not found this task in TaskState:", utils.Red(id))

// 	}
// }

/*
保存所有的state 到 ${LogPath}/STATE.json
*/
func (config *TaskConfig) DeployedSaveStateToLocal() {
	localFile := filepath.Join(config.LogPath(), "STATE.json")
	// 保存舊時的版本
	if _, err := os.Stat(localFile); err == nil {
		os.Rename(localFile, localFile+".bak")
	}
	file, err := os.Create(localFile)
	if err != nil {
		log.Println("Save Fail:", utils.Red(err))
		return
	}
	defer file.Close()
	c := 0
	for _, v := range config.state {
		stateStr := v.String()
		file.WriteString(stateStr + "\n")
		c++
	}
	fmt.Print("[AUTO SAVE]: ", c, "\r")

}

/*
從存在的目錄獲取原有部署信息
*/
func (config *TaskConfig) DeployedLoadStateFromLocal() {
	config.lock.Lock()
	defer config.lock.Unlock()

	if _, err := os.Stat(filepath.Join(config.LogPath(), "STATE.json")); err != nil {
		// 如果沒有找到 STATE.json 則簡單讀取目錄下的log
		if fs, err := ioutil.ReadDir(config.LogPath()); err == nil {
			// paths := []string{}
			for _, f := range fs {
				if strings.HasSuffix(f.Name(), ".log") {
					id := strings.SplitN(f.Name(), ".", 2)[0]
					task := TaskState{
						ID:      id,
						LogLast: f.ModTime().Local().String(),
						LogSize: fmt.Sprintf("%fMB", float64(f.Size())/float64(1024*1024)),
						State:   "Finished",
					}
					config.state[id] = task
				}

			}
		}
	} else {
		// 找到則讀取 STATE.json
		buf, err := ioutil.ReadFile(filepath.Join(config.LogPath(), "STATE.json"))
		if err != nil {
			log.Println(utils.Red("[RESTORE FAILED BY STATE JSON~~~]"))
		}

		for _, line := range strings.Split(string(buf), "\n") {
			if strings.TrimSpace(line) != "" {
				if state, err := String2TaskState(line); err == nil {
					config.state[state.ID] = *state
				}
			}
		}
	}

}

func (state TaskState) HTML() (string, error) {
	cardtemp, _ := asset.Asset("Res/services/TaskService/web/card.html")
	temp, _ := template.New("card").Parse(string(cardtemp))
	buffer := bytes.NewBuffer([]byte{})
	err := temp.Execute(buffer, state)
	if err != nil {
		return "", err
	}

	return buffer.String(), nil
}
