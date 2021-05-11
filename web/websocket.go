package web

import (
	"crypto/md5"
	"fmt"
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/Qingluan/FrameUtils/utils"
	"github.com/gorilla/websocket"
)

var (
	upgrader               = websocket.Upgrader{} // use default options
	RegistedAction         = map[string]func(data FlowData, c *websocket.Conn){}
	GlobalBroadcast        = make(chan FlowData, 10)
	homeTemplate           *template.Template
	locker                 = sync.Mutex{}
	RegistedWebSocketFuncs = map[string]Js{}
	GlobalChannels         = map[string]*websocket.Conn{}
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

// 用於websocket
type Websocket struct {
	lock           sync.Mutex
	GlobalChannels map[string]*websocket.Conn
	RegistedAction map[string]func(data map[string]interface{}) (id, tp, value string)
	MsgChanel      chan map[string]interface{}
}

// 創建websocket 並監聽
func NewWebSocket(uri string) *Websocket {
	sock := &Websocket{
		lock:           sync.Mutex{},
		GlobalChannels: map[string]*websocket.Conn{},
		RegistedAction: map[string]func(data map[string]interface{}) (id, tp, value string){},
		MsgChanel:      make(chan map[string]interface{}, 10),
	}
	http.HandleFunc(uri, sock.WebSocketAPI)
	return sock
}

func newid(raw string) string {
	// args, _ := utils.DecodeToOptions(raw)
	c := strings.ReplaceAll(raw, " ", "")
	buf := md5.Sum([]byte(c))
	// log.Println("create id by:", utils.Yellow(c))
	return fmt.Sprintf("%x", buf)
}

func (sock *Websocket) WebSocketAPI(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	seed := rand.New(rand.NewSource(time.Now().Unix())).Int()
	id := newid(time.Now().String() + fmt.Sprintf("%d", seed))
	sock.lock.Lock()
	sock.GlobalChannels[id] = c
	sock.lock.Unlock()
	defer c.Close()
	defer func(id string) {
		log.Println("Clear channel")
		sock.lock.Lock()
		delete(GlobalChannels, id)
		sock.lock.Unlock()
	}(id)
	// go GlobalMessageListen(c, id)

	for {
		tdata := make(map[string]interface{})
		err := c.ReadJSON(&tdata)
		if err != nil {
			log.Println("read error:", err)
			break
		}
		if callback, ok := sock.RegistedAction[(tdata)["tp"].(string)]; ok {
			go func() {
				id, tp, val := callback(tdata)
				if tp != "" {
					c.WriteJSON(map[string]interface{}{
						"id":    id,
						"tp":    tp,
						"value": val,
					})
				}

			}()
		}
	}

}

//廣播消息給websocket 傳遞到前端
func (sock *Websocket) Broadcast(msg map[string]interface{}) {
	// log.Println("broadCast:", msg["tp"].(string))
	for id, c := range sock.GlobalChannels {
		log.Println("id:", id, msg["tp"].(string))
		if err := c.WriteJSON(msg); err != nil && err.Error() != "websocket: close sent" {
			log.Println("[websocket err]:", utils.Red(err))
		}
	}
	// GlobalBroadcast <- msg
}

// 注冊一個回調函數
func (sock *Websocket) Regist(name string, action func(data map[string]interface{}) (id, tp, value string)) {
	sock.lock.Lock()
	if action != nil {
		sock.RegistedAction[name] = action
	}
	defer sock.lock.Unlock()
}
