package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/Qingluan/FrameUtils/textconvert"
	"github.com/Qingluan/FrameUtils/utils"
)

func main() {
	root := "."
	search := ""
	tps := ""
	flag.StringVar(&root, "r", ".", "root dir")
	flag.StringVar(&search, "s", "", "search str")
	flag.StringVar(&tps, "t", "", "set search types ")
	flag.Parse()
	startAt := time.Now()
	chans := make(chan string, 100)
	waiter := sync.WaitGroup{}
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

					if res, ix, ok := Check(t, search, waiter); ok {
						fmt.Print("\n", utils.Green(t), ":", utils.Yellow(ix), "\n", Hit(res, search))
					}
				}(waiter)

			case <-wait.C:
				time.Sleep(10 * time.Microsecond)
			}
		}
	}(chans, &waiter)

	no := 0
	filepath.Walk(root, func(path string, state os.FileInfo, inerr error) (err error) {
		_t := strings.Split(path, ".")
		if state.IsDir() {
			return nil
		}
		if tps != "" && strings.Contains(tps, _t[len(_t)-1]) {
			no += 1
			// fmt.Printf("\rsearch file>> %d ", no)
			waiter.Add(1)

			chans <- path
		} else {
			no += 1
			// fmt.Printf("\rsearch file>> %d ", no)
			waiter.Add(1)

			chans <- path
		}
		return nil
	})
	waiter.Wait()
	fmt.Println("Searched ", no, "files", time.Since(startAt))

}

func Check(file, key string, waiter *sync.WaitGroup) (out string, index int, found bool) {
	var buf []byte
	var err error
	defer waiter.Done()
	_t := strings.Split(file, ".")
	switch _t[len(_t)-1] {
	case "docx":
		if s, err := textconvert.ToStr(file); err != nil {
			return
		} else {
			buf = []byte(s)
		}

	case "xlsx":
		if s, err := textconvert.XlsxToStr(file); err != nil {
			return
		} else {
			buf = []byte(s)
		}
	case "pdf":
		if s, err := textconvert.PDFToStr(file); err != nil {
			return
		} else {
			buf = []byte(s)
		}
	default:
		buf, err = ioutil.ReadFile(file)

	}
	if err != nil {
		return
	}

	if index = bytes.Index(buf, []byte(key)); index > 0 {
		var hist []byte
		if index > 100 {
			hist = buf[index-100 : index+len(key)]
		} else {
			hist = buf[:index+len(key)]
		}

		bufs := bytes.Split(hist, []byte("\n"))
		out = string(bufs[len(bufs)-1])
		found = true
	}
	return
}

func Hit(raw, hit string) string {
	ix := strings.Index(raw, hit)
	return raw[:ix] + utils.Red(hit) + raw[ix+len(hit):]
}
