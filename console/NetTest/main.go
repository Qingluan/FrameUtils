package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"strings"

	"github.com/Qingluan/FrameUtils/utils"
	jupyter "github.com/Qingluan/jupyter/http"
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
	output := func(res *jupyter.SmartResponse, err error) {
		if err != nil {
			log.Fatal(color.New(color.FgRed).Sprint(err))
		} else {
			buf, _ := ioutil.ReadAll(res.Body)
			fmt.Println(color.New(color.FgGreen).Sprint(string(buf)))
		}
	}
	cmds := flag.Args()
	if tp == "json" {
		sess := jupyter.NewSession()
		data := utils.BDict{}
		data = data.FromCmd(strings.Join(cmds, " "))
		// log.Println(cmds, "\n", color.New(color.FgBlue).Sprint(data))
		output(sess.Json(target, data))
	} else if tp == "get" || len(cmds) == 0 {
		sess := jupyter.NewSession()
		output(sess.Get(target))
	} else if tp == "upload" && len(cmds) > 1 {
		sess := jupyter.NewSession()
		data := utils.BDict{}
		data = data.FromCmd(strings.Join(cmds[2:], " "))
		fmt.Println(data)
		output(sess.Upload(target, strings.TrimSpace(cmds[0]), strings.TrimSpace(cmds[1]), data, true))
	}

	// flag.StringVar(&target,"t","localhost","set target")
}
