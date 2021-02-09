package main

import (
	"flag"

	"github.com/Qingluan/FrameUtils/web"
	"github.com/gorilla/websocket"
)

var (
	addr = ""
)

func InitServers() {

	web.AddPageWithNoBody("/add").AddBody((&web.SearchWidget{ID: "check-url"})).OnWebsocket("check-url", func(flow web.FlowData, c *websocket.Conn) {

	})
}

func main() {
	flag.StringVar(&addr, "addr", ":8080", "listen address")
	flag.Parse()
	InitServers()
	web.StartServer(addr)
}
