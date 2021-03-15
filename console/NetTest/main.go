package main

import (
	"flag"
	"io/ioutil"
	"log"
	"strings"

	jupyter "gitee.com/dark.H/Jupyter/http"
	"github.com/Qingluan/FrameUtils/utils"
	"github.com/fatih/color"
)

var (
	target = "localhost"
	tp     = "json"
)

func main() {
	flag.StringVar(&target, "u", "http://localhost:4099", "set target")

	flag.StringVar(&tp, "t", "json", "set target")
	flag.Parse()

	cmds := flag.Args()
	if tp == "json" {
		sess := jupyter.NewSession()
		data := utils.BDict{}
		data = data.FromCmd(strings.Join(cmds, " "))
		log.Println(cmds, "\n", color.New(color.FgBlue).Sprint(data))
		if res, err := sess.Json(target, data); err != nil {
			log.Fatal(color.New(color.FgRed).Sprint(err))
		} else {
			buf, _ := ioutil.ReadAll(res.Body)
			log.Println(color.New(color.FgGreen).Sprint(string(buf)))
		}
	}

	// flag.StringVar(&target,"t","localhost","set target")
}
