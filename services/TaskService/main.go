package main

import (
	// "fmt"

	"log"

	// "math/rand"
	// "time"

	"github.com/Qingluan/FrameUtils/task"
	"github.com/Qingluan/FrameUtils/utils"
	"github.com/fatih/color"
)

func main() {
	// config := task.NewTaskConfigDefault("http://localhost:8084")
	config := task.NewTaskConfig("conf.ini")
	taskPool := task.NewTaskPool(config)
	taskPool.SetRuntime("cmd", task.CmdCall)

	taskPool.SetOkCall(func(o task.TaskObj) {
		// log.Println(o.ID(), o.String())
	})

	taskPool.SetErrCall(func(e task.ErrObj) {
		// log.Println(e.ID(), e.String())
	})

	go taskPool.StartTask(func(ok task.TaskObj, res interface{}, err error) {
		log.Println("after:", ok.Args(), "|", ok.String(), "                                                        ")
		// buf, err := ioutil.ReadFile(ok.String())
		if err != nil {
			color.New(color.FgBlue).Println(err)
		} else {
			// color.New(color.FgBlue).Println(string(buf))
		}
		log.Println(utils.UnderLine(res))
	})

	config.StartTaskWebServer()
	// fmt.Println(taskPool)
}
