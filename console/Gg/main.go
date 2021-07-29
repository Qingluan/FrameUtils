package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/Qingluan/FrameUtils/console/Gg/ui"
	"github.com/Qingluan/FrameUtils/servermanager"
	"github.com/Qingluan/FrameUtils/textconvert"
	"github.com/Qingluan/FrameUtils/tui"
	"github.com/Qingluan/FrameUtils/utils"
)

func main() {
	root := "."
	search := ""
	tps := ""
	todoPro := false
	matchAll := false
	openvim := false
	PasswordMode := false
	catmode := ""
	PROXY := ""
	// findPath := ""

	flag.StringVar(&root, "r", ".", "root dir")
	flag.StringVar(&search, "s", "", "search str")
	flag.StringVar(&tps, "t", "", "set search types ")
	flag.BoolVar(&matchAll, "v", false, "true to show every match")
	flag.BoolVar(&todoPro, "todo", false, "true to start todo program")
	flag.BoolVar(&openvim, "vim", false, "true to open with editer")
	flag.StringVar(&catmode, "cat", "", "true cat files")
	flag.BoolVar(&PasswordMode, "pwd", false, "true to open my password")
	flag.StringVar(&PROXY, "proxy", "", "set proxy")
	flag.Parse()
	args := flag.Args()
	// fmt.Println("res", args)
	if todoPro {
		ui.Main(root)
		return
	}
	if PasswordMode {
		pn := ui.Load()
		val := pn.CHoose()
		fmt.Println("Choose : ", val)
		if strings.HasPrefix(val, "ssh://") {
			vps := servermanager.Parse(val)
			vps.Proxy = PROXY
			vps.Shell()
		} else if strings.HasPrefix(val, "vul://") {
			manager := servermanager.NewVultr(val[6:])
			if err := manager.Update(); err == nil {
				ee := []tui.CanString{}
				for _, w := range manager.GetServers() {
					ee = append(ee, w)
				}
				if oneVps, ok := tui.SelectOne("select one:", ee); ok {
					vps := oneVps.(servermanager.Vps)
					vps.Proxy = PROXY
					fmt.Println(vps.Shell())
				}
			} else {
				log.Fatal(utils.Red(err))
			}
		}
		return
	}
	if catmode != "" {
		Cat(catmode)
		return
	}
	if search != "" {
		Search(root, tps, search, matchAll, openvim)
		return
	}
	if len(args) > 0 {
		FindPath(root, args)
	}
}

func OpenVim(files ...string) {

	if len(files) > 2 {
		model := ui.NewModel(files...)
		model.Do = func(chooeds []string) {
			for _, file := range chooeds {
				fs := strings.Fields(file)
				cmd := exec.Command("vim", fs[0], fs[1])
				cmd.Stdin = os.Stdin
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				cmd.Run()
			}
		}
		ui.StartModel(model)
	} else {
		for _, file := range files {
			fs := strings.Fields(file)
			cmd := exec.Command("vim", fs[0], fs[1])
			cmd.Stdin = os.Stdin
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			cmd.Run()
		}
	}

}

func FindPath(root string, args []string) {
	F := ""
	fmt.Fprintln(os.Stderr, "Find in ", args)
	filepath.Walk(root, func(path string, state os.FileInfo, err error) (oerr error) {
		found := true
		raw := path
		span := "/"
		// file := filepath.Base(path)
		if runtime.GOOS == "windows" {
			span = "\\"
		}
		lat := -1
		if state.IsDir() {
			for _, arg := range args {
				if c := strings.Index(path, span+arg); c >= 0 && c > lat {

					lat = c
				} else if c := strings.Index(path, arg); c == 0 {
					lat = c
				} else {

					found = false
					break
				}
			}

		} else {
			found = false
		}
		if found {
			F = raw
			return fmt.Errorf("found %s", "sf")
		}
		return
	})
	if F != "" {
		fmt.Print(F)
	}
}

func Cat(file string) {
	var buf []byte
	var err error
	// defer waiter.Done()
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
	fmt.Println(string(buf))
}

func Search(root, tps, search string, matchAll bool, openVim bool) {
	fileTpes := make(map[string]int)
	for _, f := range strings.Split(tps, ",") {
		fileTpes[f] = 1
	}
	startAt := time.Now()
	chans := make(chan string, 100)
	waiter := sync.WaitGroup{}
	openFiles := []string{}
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

func Check(file, key string, waiter *sync.WaitGroup, matchAll bool) (founds map[int]string, found bool) {
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
	founds = make(map[int]string)
	base := 0
	for {
		index := -1
		out := ""
		bufleng := len(buf)
		linenum := 0
		if index = bytes.Index(buf, []byte(key)); index >= 0 {
			var hist []byte
			leng := 0

			if index > 100 {
				hist = buf[index-100 : index+len(key)]
			} else {
				hist = buf[:index+len(key)]
			}
			if index+len(key)+100 > bufleng {
				leng = bufleng
			} else {
				leng = index + len(key) + 100
			}
			tail := buf[index+len(key) : leng]
			tails := bytes.Split(tail, []byte("\n"))[0]
			bufs := bytes.Split(hist, []byte("\n"))
			out = string(bufs[len(bufs)-1]) + string(tails)

			found = true
			linenum += bytes.Count(buf[:index], []byte("\n"))
			founds[linenum] = out
			if !matchAll {
				break
			}

			buf = buf[leng:]
			base += leng
		} else {
			break
		}
	}

	return
}

func Hit(raw, hit string) string {
	ix := strings.Index(raw, hit)
	pre := raw[:ix] + utils.Red(hit)
	tail := raw[ix+len(hit):]
	for {
		if ix = strings.Index(tail, hit); ix >= 0 {
			pre += tail[:ix] + utils.Red(hit)
			tail = tail[ix+len(hit):]
		} else {
			break
		}
	}
	return pre + tail
}
