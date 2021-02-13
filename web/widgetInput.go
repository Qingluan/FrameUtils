package web

import (
	"fmt"

	"github.com/Qingluan/FrameUtils/HTML"
	"github.com/gorilla/websocket"
)

type Widget interface {
	String() string
	API() string
	GetID() string
	Event() (string, string, Js, func(flowData FlowData, c *websocket.Conn))
	Margin(postion string, num interface{})
}

type InputWidgets interface {
	Widget
	Val() Js
}

type InputWidget struct {
	ID      string
	TP      string
	Style   string
	Eventjs JSEvent
}

func (input InputWidget) Event() (id string, method string, js Js, callback func(flowData FlowData, c *websocket.Conn)) {
	js = Js(input.Eventjs.String())
	method = input.Eventjs.Method()
	id = input.ID
	if input.Eventjs.Callback != nil {
		callback = input.Eventjs.Callback
	}
	return
}

func (input InputWidget) Val() Js {
	return Query("#" + input.ID).Val()
}

func (input InputWidget) API() string {
	return input.Val().WithVar("data", func(val Js) Js {
		return Js(fmt.Sprintf("SendAction(\"%s\", \"%s\", data)", input.ID, input.TP))
	}).String()
}

type SearchWidget struct {
	InputWidget
}

func (search InputWidget) GetID() string {
	return search.ID
}
func (search InputWidget) OnEvent(toggle int, howToDoVal ...func(id string, val Js) Js) InputWidget {
	self := &search
	self.Eventjs.Toogle = toggle
	if howToDoVal != nil {
		self.Eventjs.Body = self.Val().WithVar("data", func(val Js) Js {
			return howToDoVal[0](self.ID, val)
		})
	} else {
		self.Eventjs.Body = self.Val().WithVar("data", func(val Js) Js {
			return Js(fmt.Sprintf("SendAction(\"%s\", \"%s\", data)", search.ID, search.TP))
		})
	}
	return *self
}
func (search InputWidget) OnWebsocket(callback func(flowData FlowData, c *websocket.Conn)) InputWidget {
	(&search).Eventjs.Callback = callback
	return search
}
func (self InputWidget) Margin(postion string, nums interface{}) {
	num := ""
	search := &self
	switch nums.(type) {
	case int:
		num = fmt.Sprintf("%d%%", nums.(int))
	case string:
		// num, _ = strconv.Atoi(nums.(string))
		num = nums.(string)
	}
	switch postion {
	case "left":
		search.Style += fmt.Sprintf("margin-left: %s;", num)
	case "right":
		search.Style += fmt.Sprintf("margin-right: %s;", num)

	case "bottom":
		search.Style += fmt.Sprintf("margin-bottom: %s;", num)

	case "top":
		search.Style += fmt.Sprintf("margin-top: %s;", num)
	default:
		search.Style += fmt.Sprintf("margin: %s;", num)
	}
}
func (search InputWidget) String() string {
	in := Field{
		Tag:   "div",
		Class: "input-group input-group-lg",
		ID:    search.ID + "-div",
		Subs: []Field{
			{
				Tag:   "span",
				Class: "input-group-text",
				Text:  "Search >>",
			},
			{
				ID:    search.ID,
				Style: search.Style,
			},
		},
	}
	// in.Input.ID = search.ID
	// in.Label.Text = "Search >>>"
	// in.Input.Style = search.Style
	return HTML.MarshalHTML(in, "\n")
}
