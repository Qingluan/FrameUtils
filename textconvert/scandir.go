package textconvert

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/Qingluan/FrameUtils/utils"
)

type Res struct {
	Path string
	Res  ElasticFileDocs
}
type ScanTask struct {
	Num         int
	FileHandle  map[string]func(name string) (ElasticFileDocs, error)
	OkChannel   chan Res
	Waiter      sync.WaitGroup
	dir         string
	WaitChannel chan string

	okHandle func(result Res)
}

func NewDirScan(dir string, num int) (scan *ScanTask) {
	scan = new(ScanTask)
	scan.Num = num
	scan.dir = dir
	scan.OkChannel = make(chan Res, 2048)
	scan.WaitChannel = make(chan string, 10)
	scan.FileHandle = make(map[string]func(name string) (ElasticFileDocs, error))
	return
}

func (scan *ScanTask) GetType(path string) string {
	fs := strings.Split(path, ".")
	if len(fs) > 0 {
		return fs[len(fs)-1]
	}
	return ""

}

func (scan *ScanTask) SetHandle(tp string, h func(path string) (ElasticFileDocs, error)) {
	scan.FileHandle[tp] = h
}
func (scan *ScanTask) SetOkHandle(h func(r Res)) {
	scan.okHandle = h
}

func (scan *ScanTask) Scan() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	var waiter sync.WaitGroup
	stopChan := make(chan int)
	// var mem runtime.MemStats
	go func(ch chan string, okch chan Res) {
		runningNum := 0
		all := 0
		for {
			path := <-ch
			if fu, ok := scan.FileHandle[scan.GetType(path)]; ok {
				if all%200 == 0 {
					// runtime.ReadMemStats(&mem)
					fmt.Printf("got :%d : %d  \n", all, runningNum)
					// fmt.Println(mem)
				}
				runningNum++
				all++
				waiter.Add(1)
				go func(p string, okch chan Res, function func(name string) (ElasticFileDocs, error), w *sync.WaitGroup) {
					defer w.Done()
					if out, err := function(p); err != nil {
						log.Println(path, ":", utils.Red(err))
					} else {
						okch <- Res{
							Path: p,
							Res:  out,
						}
					}

				}(path, okch, fu, &waiter)
			}
			if runningNum >= scan.Num {
				waiter.Wait()
				waiter = sync.WaitGroup{}
				runningNum = 0
			}
		}
	}(scan.WaitChannel, scan.OkChannel)

	go func(ochan chan Res, handle func(res Res)) {
		tick := time.NewTicker(3 * time.Second)
		for {
			select {
			case <-tick.C:
				// time.Sleep()
			case res := <-ochan:

				if handle != nil {
					handle(res)
				}
			case <-stopChan:
				break
			}
			// res := <-ochan
		}
	}(scan.OkChannel, scan.okHandle)
	filepath.Walk(scan.dir, func(f string, fs os.FileInfo, err error) error {
		if !fs.IsDir() && strings.Contains(f, ".") {
			scan.WaitChannel <- f
		}
		return nil
	})
	time.Sleep(2 * time.Second)
	for {
		if len(scan.OkChannel) > 0 && len(scan.WaitChannel) > 0 {
			continue
		} else {
			break
		}
	}
	stopChan <- 1

}
