package textconvert

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
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
	scan.WaitChannel = make(chan string, 100)
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

	go func(ch chan string, okch chan Res) {
		runningNum := 0
		var waiter sync.WaitGroup
		for {
			path := <-ch
			if fu, ok := scan.FileHandle[scan.GetType(path)]; ok {
				fmt.Printf("got : %s                                      \r", filepath.Base(path))
				runningNum++
				waiter.Add(1)
				go func(p string, okch chan Res) {
					defer waiter.Done()
					if out, err := fu(p); err != nil {
						log.Println(path, ":", err)
					} else {
						okch <- Res{
							Path: p,
							Res:  out,
						}
					}

				}(path, okch)
			}
			if runningNum >= scan.Num {
				waiter.Wait()
				waiter = sync.WaitGroup{}
				runningNum = 0
			}
		}
	}(scan.WaitChannel, scan.OkChannel)

	go func(ochan chan Res, handle func(res Res)) {
		for {
			res := <-ochan
			if handle != nil {
				handle(res)
			}
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

}
