package web

import (
	"html/template"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var (
	upgrader               = websocket.Upgrader{} // use default options
	RegistedAction         = map[string]func(data FlowData, c *websocket.Conn){}
	GlobalBroadcast        = make(chan FlowData, 10)
	homeTemplate           *template.Template
	locker                 = sync.Mutex{}
	RegistedWebSocketFuncs = map[string]Js{}
)

type FlowData struct {
	Tp     string `json:"tp"`
	Id     string `json:"id"`
	Value  string `json:"value"`
	BackId string `json:"backid"`
}

func BroadcastMsg(msg FlowData) {
	GlobalBroadcast <- msg
}

// RegistWebSocketCallback : regist a callback in server
func RegistWebSocketCallback(id string, callback func(data FlowData, c *websocket.Conn)) {
	locker.Lock()
	defer locker.Unlock()
	RegistedAction[id] = callback
}

// GlobalMessageListen : global message listen
func GlobalMessageListen(c *websocket.Conn) {
	for {
		globalMsg := <-GlobalBroadcast
		if err := c.WriteJSON(globalMsg); err != nil {
			log.Println("[err]:", err)
			break
		}
	}
}

func ApiSocketHandle(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()
	go GlobalMessageListen(c)
	for {
		data := new(FlowData)
		err := c.ReadJSON(data)
		if err != nil {
			log.Println("read error:", err)
			break
		}
		// log.Printf("[action] : id:%s tp:%s value: %s", data.Id, data.Tp, data.Value)
		if callback, ok := RegistedAction[data.Id]; ok {
			go callback(*data, c)
		}

	}
}
