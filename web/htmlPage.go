package web

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/websocket"
)

var (
	preparedFuncs = map[string]func(w http.ResponseWriter, r *http.Request){
		"/api": ApiSocketHandle,
	}

	METHOD_POST = 0x01
	METHOD_GET  = 0x10

	ServerPagesStack = map[string]Page{}
	HomePage         = Page{
		Route:  "/",
		Method: METHOD_GET,
		Body:   "",
		Handler: func(w http.ResponseWriter, r *http.Request) {

		},
		JsEvents: map[string]JSEvent{},
	}
)

type Page struct {
	Route    string
	Method   int
	Body     string
	Handler  func(w http.ResponseWriter, r *http.Request)
	JsEvents map[string]JSEvent
	ExtendJS []Js
}

// PreparedRoute : regist a func to map
func PreparedRoute(route string, fun func(w http.ResponseWriter, r *http.Request)) {
	preparedFuncs[route] = fun
}

/*OnEvent :
@ID :
	cssselector
@jsBody :

@toggleMethod :
		EVENT_CLICK = 0
		EVENT_ENTER = 1
		EVENT_FOCUS = 2

*/
func (page Page) OnEvent(IDOrWidget Widget, jsBody Js, toggleMethod int, callback ...func(flowData FlowData, c *websocket.Conn)) {
	method := toggleMethod
	// if toggleMethod != nil {
	// 	method = toggleMethod[0]
	// }

	L(page.Route, "Event", "regist : "+IDOrWidget.GetID(), method)
	page.JsEvents[IDOrWidget.GetID()] = JSEvent{
		Body:   jsBody,
		Toogle: method,
	}
	if callback != nil {
		RegistWebSocketCallback(IDOrWidget.GetID(), callback[0])
	}
}

// RenderPage : get base bootstrap html
func (page Page) RenderPage(jsArea ...string) {
	js := ""
	extend := ""
	for _, v := range RegistedWebSocketFuncs {
		extend += v.String() + "\n"
	}

	for selector, jsevent := range page.JsEvents {
		if !strings.HasPrefix(selector, "#") {
			selector = "#" + selector
		}
		js += "\n" + Query(selector).Call(jsevent.Method(), jsevent.String()).String()
	}
	if jsArea != nil {
		js += "\n" + jsArea[0]

		// log.Println(js)
	}
	for _, j := range page.ExtendJS {
		js += "\n" + j.String()
	}
	http.HandleFunc(page.Route, func(w http.ResponseWriter, r *http.Request) {
		if page.Handler != nil {
			page.Handler(w, r)
		}
		if page.Method&METHOD_GET == METHOD_GET {
			h := fmt.Sprintf(`<!DOCTYPE html>
		<html>
			<head>
				<meta charset="utf-8">
				<style type="text/css" >%s</style>
			</head>
			<body class="h-100" style="        position: absolute; width:100%%;">
		`, BootstrapCSS) + page.Body + fmt.Sprintf(`
				<script >%s</script>
				<script >%s</script>
				<script >%s</script>
				<script jsname="base-functions">%s</script>
				<script >%s</script>
			</body>
		</html>`, Jquery, BootstrapPopJS, BootstrapJS, baseFunctionJS+extend, js)
			fmt.Fprint(w, h)
		}

	})
}

func StartServer(listenAddr string) {
	http.HandleFunc("/api", ApiSocketHandle)
	links := map[string]string{}
	for name := range ServerPagesStack {
		links[name] = name
	}

	search := SearchWidget{
		ID: "SearchInput",
		TP: "search",
	}
	search.Margin("", 30)
	container := ColsContainer{
		Links:   links,
		Content: search.String(),
	}
	if HomePage.Body == "" {
		HomePage.Body = container.String()

		HomePage.OnEvent(search, func() Js {
			return Js(search.API())
		}(), EVENT_ENTER, func(flowData FlowData, c *websocket.Conn) {
			L(HomePage.Route, "Callback", flowData)
			c.WriteJSON(flowData)
		})
	}
	HomePage.RenderPage()
	for _, page := range ServerPagesStack {
		page.RenderPage()
	}
	log.Fatal(http.ListenAndServe(listenAddr, nil))
}
