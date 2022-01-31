package textsearch

import (
	"bytes"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/Qingluan/FrameUtils/console/Gg/ui"
	"github.com/Qingluan/FrameUtils/textconvert"
	"github.com/Qingluan/FrameUtils/utils"
)

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

/* CheckFromObj

key : openobjstr@colmun

*/
func CheckFromObj(file string, querys []string, waiter *sync.WaitGroup, matchAll bool) (founds []string, found bool) {
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

	// founds = make(map[int]string)
	for _, key := range querys {
		index := -1

		if index = bytes.Index(buf, []byte(key)); index >= 0 {
			founds = append(founds, key)
			found = true
		}
	}

	return
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
