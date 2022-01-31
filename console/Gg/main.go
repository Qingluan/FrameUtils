package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/Qingluan/FrameUtils/console/Gg/ui"
	"github.com/Qingluan/FrameUtils/servermanager"
	"github.com/Qingluan/FrameUtils/textconvert"
	"github.com/Qingluan/FrameUtils/textsearch"
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
	notemode := false
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
	flag.BoolVar(&PasswordMode, "pwd", false, "true to open my password")
	flag.BoolVar(&notemode, "note", false, "true to open my password")

	flag.StringVar(&catmode, "cat", "", "true cat files")
	flag.StringVar(&PROXY, "proxy", "", "set proxy")
	flag.Parse()
	args := flag.Args()
	// fmt.Println("res", args)
	if todoPro {
		ui.Main(root)
		return
	}

	if notemode {
		val := ui.MainNote()
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
		textsearch.Search(root, tps, search, matchAll, openvim)
		return
	}
	if len(args) > 0 {
		FindPath(root, args)
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
