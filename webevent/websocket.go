package webevent

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"text/template"

	"github.com/asticode/go-astikit"
	"github.com/asticode/go-astilectron"
	"github.com/gorilla/websocket"
)

var (
	upgrader        = websocket.Upgrader{} // use default options
	RegistedAction  = map[string]func(data FlowData){}
	GlobalBroadcast = make(chan FlowData, 10)
	homeTemplate    *template.Template
	locker          = sync.Mutex{}
)

type FlowData struct {
	Tp     string `json:"tp"`
	Id     string `json:"id"`
	Value  string `json:"value"`
	BackId string `json:"backid"`
}

func SetOnServerListener(id string, callback func(data FlowData)) {
	locker.Lock()
	defer locker.Unlock()
	RegistedAction[id] = callback
}

func GlobalMessageListen(c *websocket.Conn) {
	for {
		globalMsg := <-GlobalBroadcast
		if err := c.WriteJSON(globalMsg); err != nil {
			log.Println("[err]:", err)
			break
		}
	}
}

func FlowSocket(w http.ResponseWriter, r *http.Request) {
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
		log.Printf("[action] : id:%s tp:%s value: %s", data.Id, data.Tp, data.Value)
		if callback, ok := RegistedAction[data.Id]; ok {
			go callback(*data)
		}

	}
}

func handleHome(w http.ResponseWriter, r *http.Request) {
	homeTemplate.Execute(w, "ws://"+r.Host+"/flow")
}

func BroadcastMsg(msg FlowData) {
	GlobalBroadcast <- msg
}

func StartServer(appname string, useApp bool, eles ...ElementAble) {
	homeTemplate = template.Must(template.New("").Parse(Render(appname, eles...)))
	http.HandleFunc("/flow", FlowSocket)
	http.HandleFunc("/", handleHome)
	if useApp {
		addr := "localhost:38080"
		go http.ListenAndServe(addr, nil)
		l := log.New(log.Writer(), log.Prefix(), log.Flags())
		a, err := astilectron.New(l, astilectron.Options{
			AppName:           "Test",
			BaseDirectoryPath: "example",
		})
		if err != nil {
			panic(err)
		}
		defer a.Close()
		if err = a.Start(); err != nil {
			l.Fatal(fmt.Errorf("main: starting astilectron failed: %w", err))
		}

		var w *astilectron.Window
		if w, err = a.NewWindow("http://"+addr, &astilectron.WindowOptions{
			Center: astikit.BoolPtr(true),
			Height: astikit.IntPtr(700),
			Width:  astikit.IntPtr(1200),
		}); err != nil {
			l.Fatal(fmt.Errorf("main: new window failed: %w", err))
		}

		// Create windows
		if err = w.Create(); err != nil {
			l.Fatal(fmt.Errorf("main: creating window failed: %w", err))
		}

		// Blocking pattern
		a.Wait()
	} else {
		http.ListenAndServe("localhost:38080", nil)
	}

}
