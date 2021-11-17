package main

import (
	"log"
	"strings"

	"github.com/Qingluan/FrameUtils/engine"
	"gopkg.in/ini.v1"
)

func main() {
	cfg, err := ini.Load("config.ini")
	if err != nil {
		log.Fatal("load err:", err)
	}
	connectStr := cfg.Section("DEFAULT").Key("host").Value()
	fs := strings.SplitN(connectStr, "+", 2)
	host := fs[0]
	table := fs[1]
	input := cfg.Section("DEFAULT").Key("input").Value()
	key := cfg.Section("DEFAULT").Key("key").Value()
	slqENgin := engine.ConnectByHost(host)
	err = slqENgin.MatchFromFile(input, "output.xlsx", table, key)
	log.Fatal(err)
}
