package task

import (
	"crypto/md5"
	"encoding/json"
	"io/ioutil"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/Qingluan/jupyter/http"
)

type TaskConfig struct {
	TaskNum   int      `json:"taskNum"`
	LogServer string   `json:"logserver"`
	Others    []string `json:"others"`
	Proxy     string   `json:"proxy"`
	ReTry     int      `json:"try"`
}

func (tconfig TaskConfig) Get(name string) interface{} {
	switch name {
	case "num":
		return tconfig.TaskNum
	case "proxy":
		return tconfig.Proxy
	case "others":
		return tconfig.Others
	case "logserver":
		return tconfig.LogServer
	default:
		return ""
	}
}

type TaskObj interface {
	Args() []string
	ID() string
	String() string
	Error() error
}

type ErrObj struct {
	Err  error
	args []string
}

func (erro ErrObj) String() string {
	buf, err := json.Marshal(map[string]string{
		"Tp":   "Err",
		"Data": erro.Err.Error(),
		"Args": strings.Join(erro.Args(), "|"),
	})
	if err != nil {
		log.Fatal(err)
	}
	return string(buf)
}
func (erro ErrObj) Error() error {
	return erro.Err
}
func (erro ErrObj) ID() string {
	md := md5.New()
	// md.Write()

	ks := strings.Join(erro.Args(), "|")
	raw := []byte(ks)
	b := md.Sum(raw)
	return string(b)
}
func (erro ErrObj) Args() []string {
	return erro.args
}

type TaskPool struct {
	config         *TaskConfig
	ErrCounter     map[string]int
	OkChannel      chan TaskObj
	ErrChannel     chan ErrObj
	WaitChannel    chan []string
	RunningChannel chan interface{}

	TaskCounter sync.WaitGroup
	call        func(args []string) (TaskObj, error)
	callinok    func(ok TaskObj)
	callinerr   func(erro ErrObj)
}

func (task *TaskPool) LogTo(ok TaskObj, after func(ok TaskObj, res interface{}, err error)) {
	sess := http.NewSession()
	proxy := task.config.Proxy
	if proxy != "" {
		if res, err := sess.Post(task.config.LogServer, map[string]string{
			"data":    ok.String(),
			"session": ok.ID(),
		}, proxy); err != nil {
			after(ok, res, err)
		} else {
			after(ok, res, err)
		}
	} else {
		if res, err := sess.Post(task.config.LogServer, map[string]string{
			"data":    ok.String(),
			"session": ok.ID(),
		}); err != nil {
			after(ok, res, err)
		} else {
			after(ok, res, err)
		}
	}
}

func (task *TaskPool) Build(after func(ok TaskObj, res interface{}, err error)) {
	tick := time.NewTicker(15 * time.Second)
	for {
		select {
		case args := <-task.WaitChannel:
			if task.call != nil {
				task.RunningChannel <- 1
				task.TaskCounter.Add(1)
				go func(args []string) {
					defer func() {
						<-task.RunningChannel
						task.TaskCounter.Done()
					}()
					if obj, err := task.call(args); err != nil {
						task.ErrChannel <- ErrObj{err, args}
					} else {

						task.OkChannel <- obj

					}
				}(args)
			}
		case okObj := <-task.OkChannel:
			if task.callinok != nil {
				go task.callinok(okObj)
			}
			task.LogTo(okObj, after)
		case errObj := <-task.ErrChannel:
			if task.callinerr != nil {
				go task.callinerr(errObj)
			}
			if task.ErrCount(errObj) {
				task.WaitChannel <- errObj.Args()
			} else {
				task.LogTo(errObj, after)
			}
		case <-tick.C:
			task.clearErrCounter()
		default:
			if len(task.RunningChannel) >= task.config.TaskNum {
				task.TaskCounter.Wait()
			} else {
				time.Sleep(100 * time.Microsecond)
			}
		}

	}
}

func (task *TaskPool) State() (wait int, running int, errCount int) {
	return len(task.WaitChannel), len(task.RunningChannel), len(task.ErrCounter)
}

func (task *TaskPool) ErrCount(errObj ErrObj) bool {

	if c, ok := task.ErrCounter[errObj.ID()]; ok {
		if c+1 < task.config.ReTry {
			task.ErrCounter[errObj.ID()] = c + 1
		} else {
			// delete(task.ErrCounter, errObj.ID())
			return false
		}
	} else {
		task.ErrCounter[errObj.ID()] = 1
	}
	return true
}

func (task *TaskPool) clearErrCounter() {
	clear := []string{}
	for k, v := range task.ErrCounter {
		if v+1 >= task.config.ReTry {
			clear = append(clear, k)
		}
	}
	for _, k := range clear {
		delete(task.ErrCounter, k)
	}
}

func (task *TaskPool) SetRuntime(call func(args []string) (TaskObj, error)) {
	task.call = call
}

func (task *TaskPool) SetOkCall(call func(o TaskObj)) {
	task.callinok = call
}

func (task *TaskPool) SetErrCall(call func(o ErrObj)) {
	task.callinerr = call
}

func (task *TaskPool) Push(args []string) {
	// for {
	// 	select
	// }
	task.WaitChannel <- args
}

func NewTaskPool(config *TaskConfig) *TaskPool {
	return &TaskPool{
		config:         config,
		ErrCounter:     make(map[string]int),
		OkChannel:      make(chan TaskObj, config.TaskNum),
		ErrChannel:     make(chan ErrObj, config.TaskNum),
		WaitChannel:    make(chan []string, config.TaskNum),
		RunningChannel: make(chan interface{}, config.TaskNum),
	}
}

func NewTaskConfig(fileName string) (t *TaskConfig) {
	buf, err := ioutil.ReadFile(fileName)
	if err != nil {
		log.Fatal("Read TaskConfig Err:", err)
	}
	t = new(TaskConfig)
	json.Unmarshal(buf, t)
	return
}

func NewTaskConfigDefault(logServer string) *TaskConfig {
	return &TaskConfig{
		TaskNum:   100,
		LogServer: logServer,
		Others:    []string{},
		Proxy:     "",
		ReTry:     3,
	}
}

/*DefaultTaskConfig :
TaskNum:   100,
LogServer: "http://localhost:8084/log",
Others:    []string{},
Proxy:     "",
ReTry:     3,
*/
func DefaultTaskConfig() string {
	t := &TaskConfig{
		TaskNum:   100,
		LogServer: "http://localhost:8084/log",
		Others:    []string{},
		Proxy:     "",
		ReTry:     3,
	}
	b, _ := json.MarshalIndent(t, "", "    ")
	return string(b)
}
