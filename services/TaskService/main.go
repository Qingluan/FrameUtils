package main

import (
	// "fmt"

	"flag"
	"log"
	"os"
	"time"

	// "math/rand"
	// "time"

	"github.com/Qingluan/FrameUtils/task"
	"github.com/Qingluan/FrameUtils/utils"
	jupyter "github.com/Qingluan/jupyter/http"
	"github.com/fatih/color"
)

func main() {
	// config := task.NewTaskConfigDefault("http://localhost:8084")
	configPath := ""
	daemon := false
	stop := false
	restart := false
	flag.StringVar(&configPath, "c", "conf.ini", "set config ini path")
	flag.BoolVar(&daemon, "d", false, "true to deamon mode")
	flag.BoolVar(&stop, "stop", false, "to stop service")
	flag.BoolVar(&restart, "restart", false, "to restart service")
	flag.Parse()

	if _, err := os.Stat(configPath); err != nil {
		log.Fatal(utils.BRed(err))
		return
	}

	if daemon {
		utils.Deamon("-d")
		return
	}
	config := task.NewTaskConfig(configPath)

	if stop {
		s := jupyter.NewSession()
		log.Println(utils.Yellow(config.UrlApi(config.Listen)))
		if res, err := s.Json(config.UrlApi(config.Listen), map[string]string{
			"oper": "stop",
		}); err == nil {
			log.Println(utils.Green(res.Json()))
		}
		return
	} else if restart {

		s := jupyter.NewSession()
		log.Println(utils.Yellow(config.UrlApi(config.Listen)))

		if res, err := s.Json(config.UrlApi(config.Listen), map[string]string{
			"oper": "restart",
		}); err == nil {
			log.Println(utils.Green(res.Json()))
		}
		time.Sleep(5 * time.Second)
		as := []string{}
		for _, a := range os.Args {
			if a == "-restart" {
				continue
			} else {
				as = append(as, a)
			}
		}

		as = append(as, "-d")
		os.Args = as
		utils.Deamon("-d")
		return
	}
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
