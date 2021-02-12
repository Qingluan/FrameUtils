package main

import (
	"flag"

	"github.com/Qingluan/FrameUtils/engine"
	"github.com/Qingluan/FrameUtils/smarthtml"
	"github.com/Qingluan/FrameUtils/web"
	"github.com/gorilla/websocket"
)

var (
	addr = ""
)

func InitServers() {

	web.AddPageWithNoBody("/add", true).AddBody(web.InputWidget{
		ID: "test-link",
		TP: "add-url",
	}.OnEvent(web.EVENT_ENTER).OnWebsocket(func(flow web.FlowData, c *websocket.Conn) {
		web.L("/add", "Callback", flow)
		urls := smarthtml.SmartExtractLinks(flow.Value).AsString(0)
		o := engine.FromArrays(urls)
		c.WriteJSON(web.FlowData{
			Id:    "",
			Tp:    "AddView",
			Value: o.ToHTML("show-links"),
		})
	}))
}

func main() {
	flag.StringVar(&addr, "addr", ":8080", "listen address")
	flag.Parse()
	InitServers()
	web.StartServer(addr)
}
