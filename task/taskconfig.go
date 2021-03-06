package task

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"github.com/Qingluan/FrameUtils/utils"
	"github.com/fatih/color"
)

type TaskConfig struct {
	TaskNum   int      `json:"taskNum" config:"taskNum"`
	Listen    string   `json:"listen" config:"listen"`
	LogServer string   `json:"logserver" config:"logserver"`
	Others    []string `json:"others" config:"others"`
	Proxy     string   `json:"proxy" config:"proxy"`
	ReTry     int      `json:"try" config:"try"`
	logPath   string   `json:"logPath" config:"logPath"`
	state     map[string]string
	lock      sync.RWMutex
}

func NewTaskConfig(fileName string) (t *TaskConfig) {
	t = new(TaskConfig)
	err := utils.Unmarshal(fileName, t)
	t.state = make(map[string]string)

	if err != nil {
		log.Fatal(err)
	}
	return

}

func NewTaskConfigOrDefault(fileName string) (t *TaskConfig) {
	t = new(TaskConfig)
	if _, err := os.Stat(fileName); err != nil {
		return NewTaskConfigDefault("http://localhost:4099/task/v1/log")
	} else {
		err := utils.Unmarshal(fileName, t)
		t.state = make(map[string]string)

		if err != nil {
			log.Fatal(err)
		}
	}

	return

}

func (tconfig *TaskConfig) StartTaskWebServer() {
	tconfig.PatchWebAPI()
	log.Fatal(http.ListenAndServe(tconfig.Listen, nil))
}

func (tconfig *TaskConfig) Get(name string) interface{} {
	switch name {
	case "num":
		return tconfig.TaskNum
	case "proxy":
		return tconfig.Proxy
	case "others":
		return tconfig.Others
	case "logserver":
		return tconfig.LogServer
	case "retry":
		return tconfig.ReTry
	default:
		return ""
	}
}

func (tconfig *TaskConfig) LogPath() string {
	if tconfig.logPath == "" {
		w := filepath.Join(os.TempDir(), "my-task")
		if _, err := os.Stat(w); err != nil {
			os.MkdirAll(w, os.ModePerm)
		}
		return w
	} else {
		return tconfig.logPath
	}
}

func (tconfig *TaskConfig) MakeSureTask(id string, runOrStop bool) {
	tconfig.lock.Lock()
	defer tconfig.lock.Unlock()
	if runOrStop {
		log.Println("+", color.New(color.FgGreen).Sprint(id))
		tconfig.state[id] = "running"
	} else {
		delete(tconfig.state, id)
	}
}

func (tconfig TaskConfig) PatchWebAPI() {
	http.HandleFunc("/task/v1/api", tconfig.TaskHandle)
	http.HandleFunc("/task/v1/log", tconfig.uploadFile)

}

func NewTaskConfigDefault(logServer string) *TaskConfig {
	return &TaskConfig{
		TaskNum:   100,
		LogServer: logServer,
		Others:    []string{},
		Proxy:     "",
		ReTry:     3,
		Listen:    "0.0.0.0:4099",
		state:     make(map[string]string),
	}
}

/*DefaultTaskConfigJson :
TaskNum:   100,
LogServer: "http://localhost:8084/log",
Others:    []string{},
Proxy:     "",
ReTry:     3,
*/
func DefaultTaskConfigJson() string {
	t := &TaskConfig{
		TaskNum:   100,
		LogServer: "http://localhost:8084/task/v1/log",
		Others:    []string{},
		Proxy:     "",
		ReTry:     3,
		Listen:    ":4099",
		state:     make(map[string]string),
	}
	b, _ := json.MarshalIndent(t, "", "    ")
	return string(b)
}

func DefaultTaskConfigIni() string {
	return DefaultTaskConfigString
}
