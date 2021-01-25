package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Qingluan/FrameUtils/webevent"
	"github.com/sirupsen/logrus"
)

var (
	Path    string
	proname string
	newpro  string
)

func main() {
	flag.StringVar(&Path, "conf", "", "config path")
	flag.StringVar(&proname, "app", "", "app name in config path")
	flag.StringVar(&newpro, "new", "", "create new app project in conf path  . if new is not \"\"")
	flag.Parse()
	if Path != "" {
		webevent.SetConfigPath(Path)
	}
	if newpro != "" {
		err := os.MkdirAll(newpro, os.ModePerm)
		if err != nil {
			logrus.Error(err)
			return
		}
		name := filepath.Base(newpro)
		if name == "" || name == "." {
			logrus.Error("new path is not correct!, must some path/${ProjectName}")
			return
		}
		htmlFile, err := os.Create(filepath.Join(newpro, "ui.html"))
		if err != nil {
			logrus.Error(err)
			return
		}

		defer htmlFile.Close()
		cssFile, err := os.Create(filepath.Join(newpro, "ui.css"))
		if err != nil {
			logrus.Error(err)
			return
		}
		defer cssFile.Close()
		htmlFile.WriteString(webevent.BaseHTMLStyle)
		cssFile.WriteString(webevent.BaseCssStyle)
		isLocal := false
		if Path == "" {
			Path = filepath.Join(newpro, "config.ini")
			isLocal = true
		}
		confFile, err := os.OpenFile(Path, os.O_APPEND|os.O_RDWR|os.O_CREATE, os.ModePerm)
		if err != nil {
			logrus.Error(err)
			return
		}
		defer confFile.Close()
		if isLocal {
			_, err = confFile.Write([]byte(fmt.Sprintf("\n[%s]\nbody = %s\ncss = %s\n", name, "ui.html", "ui.css")))

		} else {
			h, _ := filepath.Abs(filepath.Join(newpro, "ui.html"))
			c, _ := filepath.Abs(filepath.Join(newpro, "ui.css"))
			_, err = confFile.Write([]byte(fmt.Sprintf("\n[%s]\nbody = %s\ncss = %s\n", name, h, c)))

		}

		if err != nil {
			logrus.Error(err)
		}

		goFile, err := os.Create(filepath.Join(newpro, "main.go"))
		if err != nil {
			logrus.Error(err)
			return
		}
		defer goFile.Close()
		goFile.Write([]byte(fmt.Sprintf(webevent.BaseMainGoFile, name)))
	}
}
