package textsearch

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/Qingluan/FrameUtils/engine"
	"github.com/Qingluan/FrameUtils/utils"
)

func Search(root, tps, search string, matchAll bool, openVim bool) {
	fileTpes := make(map[string]int)
	for _, f := range strings.Split(tps, ",") {
		fileTpes[f] = 1
	}
	startAt := time.Now()
	chans := make(chan string, 100)
	waiter := sync.WaitGroup{}
	openFiles := []string{}

	querys := []string{}
	mode := "key"

	/// deal query
	if strings.Contains(search, "@") {
		fs := strings.SplitN(search, "@", 2)
		objfilePath := strings.TrimSpace(fs[0])
		colmunName := strings.TrimSpace(fs[1])

		obj, err := engine.OpenObj(objfilePath)
		if err == nil {
			mode = "file"
		}
		searchValues, err := obj.SelectAllByNames(colmunName)
		if err != nil {
			log.Fatal("extract values from query obj err :", err)
		}
		for l := range searchValues {
			for _, k := range l {
				querys = append(querys, k)
			}
		}
	}

	if mode == "file" {

		if len(querys) > 0 {
			fmt.Println("search key:", len(querys))
			go func(task chan string, waiter *sync.WaitGroup) {
				// log.Println("Run")
				wait := time.NewTicker(2 * time.Second)
				for {

					select {
					case t := <-task:
						waiter.Add(1)
						go func(waiter *sync.WaitGroup) {
							// startAt := time.Now()

							defer waiter.Done()
							// defer fmt.Println("Used:", time.Since(startAt))

							if founds, ok := CheckFromObj(t, querys, waiter, matchAll); ok {
								// fmt.Print("\n", utils.Green(t), " +", utils.Yellow(ix+1), "\n", Hit(res, search))
								fmt.Println(t, ",", strings.Join(founds, "|"))
							}
						}(waiter)

					case <-wait.C:
						time.Sleep(10 * time.Microsecond)
					}
				}
			}(chans, &waiter)

		} else {
			fmt.Println("no key query !")
			return
		}

	} else if mode == "key" {
		go func(task chan string, waiter *sync.WaitGroup) {
			// log.Println("Run")
			wait := time.NewTicker(2 * time.Second)
			for {

				select {
				case t := <-task:
					waiter.Add(1)
					go func(waiter *sync.WaitGroup) {
						// startAt := time.Now()

						defer waiter.Done()
						// defer fmt.Println("Used:", time.Since(startAt))

						if founds, ok := Check(t, search, waiter, matchAll); ok {
							for ix, res := range founds {
								if openVim {
									openFiles = append(openFiles, fmt.Sprintf("%s +%d %s", t, ix+1, Hit(res, search)))
								} else {
									fmt.Print("\n", utils.Green(t), " +", utils.Yellow(ix+1), "\n", Hit(res, search))
								}
							}
						}
					}(waiter)

				case <-wait.C:
					time.Sleep(10 * time.Microsecond)
				}
			}
		}(chans, &waiter)

	}

	no := 0
	filter := 0
	filepath.Walk(root, func(path string, state os.FileInfo, inerr error) (err error) {
		_t := strings.Split(path, ".")
		t := _t[len(_t)-1]
		if state.IsDir() {
			return nil
		}

		filter += 1
		if tps != "" {
			// fmt.Println(tps, t)
			if _, ok := fileTpes[t]; ok {
				no += 1

				waiter.Add(1)

				chans <- path
			}
		} else {
			no += 1
			// fmt.Printf("\rsearch file>> %d ", no)
			waiter.Add(1)

			chans <- path
		}

		if filter%10000 == 0 {
			fmt.Printf("\rsearch file>> %d  Grep:%d", filter, no)
		}
		return nil
	})
	waiter.Wait()
	time.Sleep(200 * time.Millisecond)
	fmt.Println("\nSearched ", no, "files", time.Since(startAt))
	OpenVim(openFiles...)
}
