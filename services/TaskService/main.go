package main

import (
	// "fmt"

	"flag"
	"fmt"
	"log"
	"os"
	"time"

	// "math/rand"
	// "time"

	"github.com/Qingluan/FrameUtils/asset"
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
	gen := false
	flag.StringVar(&configPath, "c", "conf.ini", "set config ini path")
	flag.BoolVar(&daemon, "d", false, "true to deamon mode")
	flag.BoolVar(&stop, "stop", false, "to stop service")
	flag.BoolVar(&restart, "restart", false, "to restart service")
	flag.BoolVar(&gen, "G", false, "generate a conf.ini template")

	flag.Parse()

	if _, err := os.Stat(configPath); err != nil {
		configPath, err = asset.AssetAsFile("Res/services/TaskService/conf.ini")

		if err != nil {
			log.Fatal(utils.BRed(err))
		}
	}
	if gen {
		fmt.Println(`
[default]

taskNum = 100
listen = :4099
logserver = https://localhost:4099/task/v1/log
try = 3
#logPath = 
#others = 
#proxy = 
sslcert = "server.crt"
sslkey = "server.key"		
		`)
		os.Exit(0)
	}

	if daemon {
		utils.Deamon("-d")
		return
	}
	config := task.NewTaskConfig(configPath)
	if config.SSLCert != "" {
		if _, err := os.Stat(config.SSLCert); err != nil {
			k, _ := asset.AssetAsFile("Res/services/TaskService/server.key")
			c, _ := asset.AssetAsFile("Res/services/TaskService/server.crt")
			config.SSLCert = c
			config.SSLKey = k

		}
	}
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
	taskPool.SetOkCall(func(o task.TaskObj) {
		// log.Println(o.ID(), o.String())
	})

	taskPool.SetErrCall(func(e task.ErrObj) {
		// log.Println(e.ID(), e.String())
	})

	go taskPool.StartTask(func(ok task.TaskObj, res interface{}, err error) {
		// log.Println("after:", ok.Args(), "|", ok.String(), "                                                        ")
		// buf, err := ioutil.ReadFile(ok.String())
		if err != nil {
			color.New(color.FgBlue).Println(err)
		} else {
			// color.New(color.FgBlue).Println(string(buf))
		}
		log.Println(utils.UnderLine(res))
	})

	logo := `
    ___________              __       _________               __                  
    \__    ___/____    _____|  | __  /   _____/__.__. _______/  |_  ____   _____  
      |    |  \__  \  /  ___/  |/ /  \_____  <   |  |/  ___/\   __\/ __ \ /     \ 
      |    |   / __ \_\___ \|    <   /        \___  |\___ \  |  | \  ___/|  Y Y  \
      |____|  (____  /____  >__|_ \ /_______  / ____/____  > |__|  \___  >__|_|  /
                   \/     \/     \/         \/\/         \/            \/      \/ 

    `
	fmt.Println(utils.Green(logo))
	fmt.Println(utils.Yellow("Listen:", config.Listen))
	config.StartTaskWebServer()
	// fmt.Println(taskPool)
}
