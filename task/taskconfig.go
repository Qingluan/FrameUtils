package task

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/Qingluan/FrameUtils/utils"
	"github.com/Qingluan/FrameUtils/web"
	"github.com/fatih/color"
)

type TaskConfig struct {
	TaskNum    int      `json:"taskNum" config:"taskNum"`
	Listen     string   `json:"listen" config:"listen"`
	LogServer  string   `json:"logserver" config:"logserver"`
	Others     []string `json:"others" config:"others"`
	Proxy      string   `json:"proxy" config:"proxy"`
	ReTry      int      `json:"try" config:"try"`
	LogPathStr string   `json:"logPath" config:"logPath"`
	Schema     string   `json:"schema" config:"schema"`
	SSLCert    string   `json:"sslcert" config:"sslcert"`
	SSLKey     string   `json:"sslkey" config:"sslkey"`
	Timeout    int      `json:"timeout" config:"timeout"`
	// state      map[string]string
	depatch map[string]string
	state   map[string]TaskState
	procs   map[string]string
	// 用来记录当前任务分配的服务器序号 Others[n] , n = (n + 1) % (len(Others))
	taskDipatchCursor int
	// 开启服务时初始的登陆密码，每次随机
	RandomLoginSession string
	// websocket 控制
	Websocket *web.Websocket
	lock      sync.RWMutex
}

func NewTaskConfig(fileName string) (t *TaskConfig) {
	// defer
	t = new(TaskConfig)
	err := utils.Unmarshal(fileName, t)
	t.state = make(map[string]TaskState)
	t.depatch = make(map[string]string)
	t.procs = make(map[string]string)
	t.LoadState()
	log.Println("Listen:", t.Listen)
	if err != nil {
		log.Fatal(err)
	}
	return

}

func NewTaskConfigOrDefault(fileName string) (t *TaskConfig) {

	t = new(TaskConfig)
	if _, err := os.Stat(fileName); err != nil {
		t = NewTaskConfigDefault("http://localhost:4099/task/v1/log")
		t.state = make(map[string]TaskState)
		t.procs = make(map[string]string)
		t.depatch = make(map[string]string)
		if t.Timeout == 0 {
			t.Timeout = 10
		}
		t.LoadState()

	} else {
		err := utils.Unmarshal(fileName, t)
		t.state = make(map[string]TaskState)
		t.procs = make(map[string]string)
		t.depatch = make(map[string]string)
		if t.Timeout == 0 {
			t.Timeout = 10
		}
		t.LoadState()
		if t.Schema == "" {

			t.Schema = "http"
		}
		if err != nil {
			log.Fatal(err)
		}
	}

	return

}

func (tconfig *TaskConfig) StartTaskWebServer() {
	err := tconfig.BuildWebInitialization()
	if err != nil {
		log.Fatal(err)
	}
	tconfig.PatchWebAPI()
	if tconfig.SSLCert != "" {
		log.Fatal("HTTPS Server:", http.ListenAndServeTLS(tconfig.Listen, tconfig.SSLCert, tconfig.SSLKey, nil))
	} else {
		log.Fatal("HTTP Server:", http.ListenAndServe(tconfig.Listen, nil))

	}
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
	if tconfig.LogPathStr == "" {
		w := filepath.Join(os.TempDir(), "my-task")
		if _, err := os.Stat(w); err != nil {
			os.MkdirAll(w, os.ModePerm)
		}
		return w
	} else {
		return tconfig.LogPathStr
	}
}

func (tconfig *TaskConfig) MakeSureTask(id string, runOrStop bool) {
	tconfig.lock.Lock()
	defer tconfig.lock.Unlock()
	if runOrStop {
		log.Println("+", color.New(color.FgGreen).Sprint(id))
		if tconfig.state == nil {
			log.Println("not init")
		}
		tconfig.state[id] = TaskState{
			ID:             id,
			State:          "Running",
			DeployedServer: tconfig.MyIP() + ":" + tconfig.MyPort(),
		}
	} else {
		delete(tconfig.state, id)
	}
}

func (tconfig *TaskConfig) PatchWebAPI() {
	http.HandleFunc("/task/v1/api", tconfig.TaskHandle)
	http.HandleFunc("/task/v1/log", tconfig.uploadFile)
	http.HandleFunc("/task/v1/", tconfig.SimeplUI)
	http.HandleFunc("/task/v1/taskfile", tconfig.DealWithUploadFile)

}

func NewTaskConfigDefault(logServer string) (t *TaskConfig) {
	defer t.LoadState()
	t = &TaskConfig{
		TaskNum:   100,
		LogServer: logServer,
		Others:    []string{},
		Proxy:     "",
		ReTry:     2,
		Listen:    "0.0.0.0:4099",
		Schema:    "http",
		state:     make(map[string]TaskState),
	}
	return
}

/*DefaultTaskConfigJson :
TaskNum:   100,
LogServer: "http://localhost:8084/log",
Others:    []string{},
Proxy:     "",
ReTry:     2,
*/
func DefaultTaskConfigJson() string {
	t := &TaskConfig{
		TaskNum:   100,
		LogServer: "http://localhost:8084/task/v1/log",
		Others:    []string{},
		Proxy:     "",
		ReTry:     2,
		Listen:    ":4099",
		Schema:    "http",
		state:     make(map[string]TaskState),
	}
	b, _ := json.MarshalIndent(t, "", "    ")
	return string(b)
}

func DefaultTaskConfigIni() string {
	return DefaultTaskConfigString
}
func (taskConfig *TaskConfig) MyIP() (ip string) {
	return utils.GetLocalIP()
}

func (taskConfig *TaskConfig) MyPort() (port string) {
	return strings.Split(taskConfig.Listen, ":")[1]
}
