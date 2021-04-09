package task

import (
	"log"
	"sync"

	"github.com/Qingluan/FrameUtils/utils"
	jupyter "github.com/Qingluan/jupyter/http"
)

func (config *TaskConfig) CheckAlive(servers ...string) {
	defer log.Println(utils.BGreen("Alive: ", len(config.Others)))
	waiter := make(map[string]bool)
	for _, s := range servers {
		waiter[s] = true
	}
	for _, s := range config.Others {
		waiter[s] = true
	}
	var lockWait sync.WaitGroup
	alives := make(chan string, len(waiter))

	for k := range waiter {
		lockWait.Add(1)
		go func(w *sync.WaitGroup, s string, alive chan string) {
			defer w.Done()
			sess := jupyter.NewSession()
			if res, err := sess.Json(config.UrlApi(s), map[string]string{
				"oper": "test",
			}); err == nil {
				if _, ok := res.Json()["state"]; ok {
					alive <- s
				}
			}
		}(&lockWait, k, alives)
	}
	lockWait.Wait()
	aliveOthers := []string{}
	for {
		if len(alives) == 0 {
			break
		}
		e := <-alives
		aliveOthers = append(aliveOthers, e)
	}
	config.Others = aliveOthers
}
