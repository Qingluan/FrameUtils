package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"strings"

	"github.com/Qingluan/FrameUtils/servermanager"
	"github.com/Qingluan/FrameUtils/tui"
	"github.com/Qingluan/FrameUtils/utils"
	jupyter "github.com/Qingluan/jupyter/http"
	"github.com/fatih/color"
)

var (
	target    = "localhost"
	tp        = "json"
	seemyip   = false
	PassFile  = ""
	deploy    = ""
	deployRun = ""
	proxy     = ""
	ssh       = false
)

func main() {
	flag.StringVar(&target, "u", "http://localhost:4099", "set target")

	flag.StringVar(&tp, "t", "json", "set target")
	flag.BoolVar(&seemyip, "ip", false, "see my ip")
	flag.StringVar(&PassFile, "pass", "", "set pass file")
	flag.StringVar(&deploy, "D", "", "set deploy file path.")
	flag.StringVar(&proxy, "proxy", "", "set proxy.")
	flag.StringVar(&deployRun, "Dcmd", "", "set deploy upload then run shell.")
	flag.BoolVar(&ssh, "ssh", false, "true to ssh shell.")

	flag.Parse()

	if seemyip {
		fmt.Printf("my ip: %s =v=\n", utils.Green(utils.GetLocalIP()))
		return
	}
	if ssh {
		bytePassword := tui.GetPass("API/ ssh auth( ip=xxxx , pass= xxxx)")
		// args := flag.Args()
		// files := append([]string{deploy}, args...)
		if strings.Contains(bytePassword, "=") {
			kargs := utils.BDict{}
			kargs = kargs.FromCmd(bytePassword)
			if ip, ok := kargs["ip"]; ok {
				if passwd, ok := kargs["pass"]; ok {
					vps := servermanager.Vps{
						IP:    ip,
						PWD:   passwd,
						USER:  "root",
						Proxy: proxy,
					}
					// vps.Upload(PassFile, true)
					fmt.Println(vps.Shell())
				}
			}
		} else {
			manager := servermanager.NewVultr(bytePassword)
			if err := manager.Update(); err == nil {
				ee := []tui.CanString{}
				for _, w := range manager.GetServers() {
					ee = append(ee, w)
				}
				if oneVps, ok := tui.SelectOne("select one:", ee); ok {
					vps := oneVps.(servermanager.Vps)
					vps.Proxy = proxy
					fmt.Println(vps.Shell())
				}
			} else {
				log.Fatal(utils.Red(err))
			}
		}

		return
	}
	if deploy != "" {
		bytePassword := tui.GetPass("API/ ssh auth( ip=xxxx , pass= xxxx)")
		args := flag.Args()
		files := append([]string{deploy}, args...)
		if strings.Contains(bytePassword, "=") {
			kargs := utils.BDict{}
			kargs = kargs.FromCmd(bytePassword)
			if ip, ok := kargs["ip"]; ok {
				if passwd, ok := kargs["pass"]; ok {
					vps := servermanager.Vps{
						IP:    ip,
						PWD:   passwd,
						USER:  "root",
						Proxy: proxy,
					}
					// vps.Upload(PassFile, true)
					fmt.Println(vps.Deploy(files, deployRun))
				}
			}
		} else {
			manager := servermanager.NewVultr(bytePassword)
			if err := manager.Update(); err == nil {
				ee := []tui.CanString{}
				for _, w := range manager.GetServers() {
					ee = append(ee, w)
				}
				if oneVps, ok := tui.SelectOne("select one:", ee); ok {
					vps := oneVps.(servermanager.Vps)
					fmt.Println(vps, "|", vps.PWD)
					vps.Proxy = proxy
					fmt.Println(vps.Deploy(files, deployRun))
				}
			} else {
				log.Fatal(utils.Red(err))
			}
		}

		return

	}

	if PassFile != "" {
		bytePassword := tui.GetPass("API/ ssh auth( ip=xxxx , pass= xxxx)")
		if strings.Contains(bytePassword, "=") {
			kargs := utils.BDict{}
			kargs = kargs.FromCmd(bytePassword)
			if ip, ok := kargs["ip"]; ok {
				if passwd, ok := kargs["pass"]; ok {
					vps := servermanager.Vps{
						IP:   ip,
						PWD:  passwd,
						USER: "root",
					}
					vps.Upload(PassFile, true)
				}
			}
		} else {
			manager := servermanager.NewVultr(bytePassword)
			if err := manager.Update(); err == nil {
				ee := []tui.CanString{}
				for _, w := range manager.GetServers() {
					ee = append(ee, w)
				}
				if oneVps, ok := tui.SelectOne("select one:", ee); ok {
					// fmt.Println(, "|", vps.PWD)

					oneVps.(servermanager.Vps).Upload(PassFile, true)
				}
			} else {
				log.Fatal(utils.Red(err))
			}
		}

		return

	}

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
