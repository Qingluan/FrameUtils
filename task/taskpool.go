package task

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/Qingluan/FrameUtils/utils"
	"github.com/Qingluan/FrameUtils/web"
)

type TaskPool struct {
	config         *TaskConfig
	ErrCounter     map[string]int
	ThrowCounter   map[string]int
	OkChannel      chan TaskObj
	ErrChannel     chan ErrObj
	WaitChannel    chan []string
	RunningChannel chan interface{}
	TaskCounter    sync.WaitGroup
	call           map[string]func(config *TaskConfig, args []string, kargs utils.Dict) (TaskObj, error)
	callinok       func(ok TaskObj)
	callinerr      func(erro ErrObj)
}

func NewTaskPool(config *TaskConfig) *TaskPool {
	return &TaskPool{
		config:         config,
		ErrCounter:     make(map[string]int),
		OkChannel:      make(chan TaskObj, config.TaskNum),
		ErrChannel:     make(chan ErrObj, config.TaskNum),
		WaitChannel:    make(chan []string, config.TaskNum),
		RunningChannel: make(chan interface{}, config.TaskNum),
		ThrowCounter:   make(map[string]int),
	}
}

/** StartTask : async get depatched task
这里从管道获取任务
所有的任务都以 input 的形式过来， 通过utils.来解析
如果有 logTo="" 的可选项，则完成后log打印和上传结果日志到这里否则在本地
*/
func (task *TaskPool) StartTask(after func(ok TaskObj, res interface{}, err error)) {
	tick := time.NewTicker(15 * time.Second)
	tickAutoSave := time.NewTicker(30 * time.Second)
	task.SetRuntime("state", task.StateCall)
	task.SetRuntime("http", HTTPCall)
	task.SetRuntime("cmd", CmdCall)
	task.SetRuntime("tcp", TCPCall)
	task.SetRuntime("config", ConfigCall)

	// 這裏 增加 websocket 插件
	task.config.Websocket = web.NewWebSocket("/task/v1/websocket")
	task.config.Websocket.Regist("hello", func(data map[string]interface{}) (id, tp, value string) {
		return "hello", "hello", "connected"
	})
	task.config.DeployedLoadStateFromLocal()
	for {
		select {
		case args := <-task.WaitChannel:
			if len(args) == 1 {
				task.Patch(args[0], "")
			} else if len(args) == 2 {
				task.Patch(args[0], args[1])
			} else {
				log.Println(utils.BRed("err args from waiChannel:", args))
			}
		case args := <-DefaultTaskWaitChnnael:

			if len(args) == 1 {
				if args[0] == "stop" || args[0] == "restart" {
					task.config.SaveState()
					os.Exit(0)
					break
				}
				task.Patch(args[0], "")
			} else if len(args) == 2 {
				if _, ok := task.call[args[0]]; ok {
					// log.Println("task entry:", utils.Magenta(args[0]))
					task.Patch(args[0], args[1])
				} else {

					task.DelayRetryPass(args...)
				}
			}
		case okObj := <-task.OkChannel:
			if task.callinok != nil {
				go task.callinok(okObj)
			}
			// 任务完成改变状态
			task.config.DeployedSwitchState(okObj.ID(), "Finished")
			// task.config.DeployedSaveLogState(okObj.ID())

			task.LogTo(okObj, "Finished", after)

		case errObj := <-task.ErrChannel:
			if task.callinerr != nil {
				go task.callinerr(errObj)
			}
			if task.ErrCount(errObj) {
				errObj.LogToLocal()
				task.WaitChannel <- errObj.Args()
			} else {
				// 任务失败改变状态
				task.config.DeployedSwitchState(errObj.ID(), "Failed")
				// task.config.DeployedSaveLogState(errObj.ID())
				delete(task.ErrCounter, errObj.ID())
				task.LogTo(errObj, "Failed", after)

			}
		case broadMsg := <-task.config.Websocket.MsgChanel:

			go task.config.Websocket.Broadcast(broadMsg)
		case <-tick.C:
			task.clearErrCounter()

		case <-tickAutoSave.C:
			task.config.DeployedSaveStateToLocal()
		default:
			if len(task.RunningChannel) >= task.config.TaskNum {
				task.TaskCounter.Wait()
			} else {
				time.Sleep(100 * time.Microsecond)
			}
		}

	}
}

// 发送结果到日志服务器
func (task *TaskPool) LogTo(ok TaskObj, state string, after func(ok TaskObj, res interface{}, err error)) {

	// 如果日志服务器地址和本地地址重合则不在发送日志
	if strings.Contains(task.config.LogServer, task.config.MyIP()+":"+task.config.MyPort()) {
		log.Println(utils.Red("Skip Log : ", task.config.LogServer))
		return
	}
	proxy := task.config.Proxy
	logTo := ok.ToGo()
	if logTo == "" {
		logTo = task.config.LogServer
	}
	logToUrl := task.config.UrlApiLog(logTo)
	// log.Println(ok, " LogTo:", logToUrl)
	if res, err := Upload(ok.ID(), ok.Path(), state, logToUrl, proxy); err != nil {
		after(ok, res, err)
	} else {
		after(ok, res, err)
	}

}

/** Patch :
### 部署函数
	这里正式从 task.call中通过callTp选取函数来执行 目前有  Httpcall / cmdcall / 后续还支持tcpcall 等...

> 如果 Others 存在,而 kargs 里没有特别指明 Local=trueOrAnyThing 则使用远端分布式部署任务
#### 支持调用类型

 * http
 * cmd
 * config
 * tcp

*/
func (task *TaskPool) Patch(callTp, raw string) {
	if callTp == "" {
		return
	}
	op := callTp
	if call, ok := task.call[op]; ok {
		task.RunningChannel <- 1
		task.TaskCounter.Add(1)

		id := NewID(raw)
		if op != "state" {
			// log.Println("|", raw, "|")
			task.config.DeploySaveState(op+"-"+id, "localhost", raw)
		}
		go func(raw string, waiter *sync.WaitGroup, taskConfigCopy *TaskConfig, run chan interface{}, okChan chan TaskObj, errChan chan ErrObj) {
			id := NewID(raw)

			args, kargs := utils.DecodeToOptions(raw)

			logTo := ""
			if e, ok := kargs["logTo"]; ok {
				logTo = e.(string)
			}
			defer func() {
				<-run
				waiter.Done()
				// log.Println(utils.Green("Finish:", id))
			}()
			// 对于几个特殊的call 函数特别调用，比如configCall 不会进入 okchannel
			if obj, err := call(taskConfigCopy, args, kargs); err != nil {
				log.Println(utils.UnderLine("Err:", id))
				errChan <- ErrObj{taskConfigCopy.LogPath(), err, op, raw, logTo}
			} else if obj != nil {
				// 从ID 获取类型
				if strings.HasPrefix(obj.ID(), "config-") {
					log.Println(utils.UnderLine(obj.ID()))
				} else {
					okChan <- obj
				}
			} else {

			}
		}(raw, &task.TaskCounter, task.config.Copy(), task.RunningChannel, task.OkChannel, task.ErrChannel)
	}
}

func (task *TaskPool) StateCall(config *TaskConfig, args []string, kargs utils.Dict) (ok TaskObj, err error) {
	logfiles, err := ioutil.ReadDir(task.config.LogPath())
	if err != nil {
		return nil, err
	}
	// args, kargs := utils.DecodeToOptions(raw)
	fs := []string{}
	for _, f := range logfiles {
		fs = append(fs, f.Name())
	}
	buf, _ := json.Marshal(task.config)
	buf2, _ := json.Marshal(fs)
	buf3, _ := json.Marshal(task.config.state)
	d := TData{
		"running": fmt.Sprintf("%d", len(task.RunningChannel)),
		"wait":    fmt.Sprintf("%d", len(task.WaitChannel)),
		"config":  string(buf),
		"logs":    string(buf2),
		"task":    string(buf3),
		"others":  strings.Join(task.config.Others, ","),
		"errnum":  fmt.Sprintf("%d", len(task.ErrChannel)),
		"lognum":  fmt.Sprintf("%d", len(fs)),
	}
	out, _ := json.Marshal(d)
	DefaultTaskOutputChannle <- string(out)
	return nil, nil
}

func (task *TaskPool) DelayRetryPass(args ...string) {
	if v, ok := task.ThrowCounter[args[0]]; ok {
		if v > 10 {

			log.Println("task give up:", utils.Magenta(args[0]), v)
			return
		}

		task.ThrowCounter[args[0]] = v + 1

		log.Println("task pass:", utils.Magenta(args[0]), v)
	} else {
		task.ThrowCounter[args[0]] = 1

		log.Println("task pass:", utils.Magenta(args[0], v))
	}
	time.Sleep(5 * time.Second)
	DefaultTaskWaitChnnael <- args
}

func (task *TaskPool) TaskSystemState() (wait int, running int, errCount int) {
	return len(task.WaitChannel), len(task.RunningChannel), len(task.ErrCounter)
}

func (task *TaskPool) ErrCount(errObj ErrObj) bool {
	defer log.Println(utils.Red("[retry]:"), task.ErrCounter[errObj.ID()], "/", task.config.ReTry, utils.Yellow(errObj.ID(), " : "))
	if c, ok := task.ErrCounter[errObj.ID()]; ok {
		if c+1 < task.config.ReTry {
			task.ErrCounter[errObj.ID()] = c + 1
			task.config.DeployedSwitchState(errObj.ID(), fmt.Sprintf("retry:%d/%d", c+1, task.config.ReTry))
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
		if v >= task.config.ReTry {
			clear = append(clear, k)
		}
	}
	for _, k := range clear {
		delete(task.ErrCounter, k)
	}
}

func (task *TaskPool) SetRuntime(name string, call func(config *TaskConfig, args []string, kargs utils.Dict) (TaskObj, error)) {
	if task.call == nil {
		task.call = make(map[string]func(config *TaskConfig, args []string, kargs utils.Dict) (TaskObj, error))
	}
	task.call[name] = call
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
