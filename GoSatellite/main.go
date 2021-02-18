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
		if _, ok := web.WithTmpDB(flow.Value); ok {
			web.L("info", "exists:", ok)
			web.NextActionNotify("show-data", flow.Value, c)
		} else {
			web.L("info", "exists:", ok)
			urls := smarthtml.SmartExtractLinks(flow.Value).AsString(0)
			o := engine.FromArrays(urls)
			db := o.WithTmpDB(flow.Value)
			web.L("save to:", db.FileName)
			c.WriteJSON(web.FlowData{
				Id:    "",
				Tp:    "AddView",
				Value: o.ToHTML("show-links"),
			})
		}

	}))
}

func main() {
	flag.StringVar(&addr, "addr", ":8080", "listen address")
	flag.Parse()
	InitServers()
	web.StartServer(addr)
}
