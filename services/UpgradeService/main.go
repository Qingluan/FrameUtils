package main

import (
	"flag"

	"github.com/Qingluan/FrameUtils/upgradeservice"
	"github.com/Qingluan/FrameUtils/utils"
)

func main() {
	address := ""
	daemon := false
	flag.StringVar(&address, "s", "127.0.0.1:13080", "set upgrade server address")

	flag.BoolVar(&daemon, "D", false, "set deamon mode")
	flag.Parse()

	if daemon {
		utils.Deamon("-D")
	} else {
		upgradeservice.StartUpgradeService(address)
	}

}
