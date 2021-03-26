package task

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"sync"
	"time"

	"github.com/Qingluan/FrameUtils/utils"
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
	call           map[string]func(config *TaskConfig, args []string, extension ...string) (TaskObj, error)
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

func (task *TaskPool) LogTo(ok TaskObj, after func(ok TaskObj, res interface{}, err error)) {
	proxy := task.config.Proxy
	if res, err := Upload(ok.ID(), ok.String(), task.config.LogServer, proxy); err != nil {
		after(ok, res, err)
	} else {
		after(ok, res, err)
	}

}

func (task *TaskPool) Patch(patchargs []string) {
	if len(patchargs) < 1 {
		return
	}
	op := patchargs[0]
	if call, ok := task.call[op]; ok {
		task.RunningChannel <- 1
		task.TaskCounter.Add(1)
		go func(args []string) {
			id := NewID(args)
			log.Println(utils.Yellow("Start:", id))
			defer func() {
				<-task.RunningChannel
				task.TaskCounter.Done()
				log.Println(utils.Green("Finish:", id))
			}()
			if obj, err := call(task.config, args); err != nil {
				log.Println(utils.UnderLine("Err:", id))
				task.ErrChannel <- ErrObj{err, args}
			} else if obj != nil {
				task.OkChannel <- obj
			} else {

			}
		}(patchargs[1:])
	}
}

func (task *TaskPool) StateCall(config *TaskConfig, args []string, extensions ...string) (ok TaskObj, err error) {
	logfiles, err := ioutil.ReadDir(task.config.LogPath())
	if err != nil {
		return nil, err
	}
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

func (task *TaskPool) StartTask(after func(ok TaskObj, res interface{}, err error)) {
	tick := time.NewTicker(15 * time.Second)
	task.SetRuntime("state", task.StateCall)
	task.SetRuntime("http", HTTPCall)
	for {
		select {
		case args := <-task.WaitChannel:
			task.Patch(args)
		case args := <-DefaultTaskWaitChnnael:
			if len(args) > 0 {
				if _, ok := task.call[args[0]]; ok {
					log.Println("task entry:", utils.Magenta(args[0]))
					task.Patch(args)
				} else {

					task.DelayRetryPass(args...)
				}
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

func (task *TaskPool) SetRuntime(name string, call func(config *TaskConfig, args []string, extension ...string) (TaskObj, error)) {
	if task.call == nil {
		task.call = make(map[string]func(config *TaskConfig, args []string, extension ...string) (TaskObj, error))
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
