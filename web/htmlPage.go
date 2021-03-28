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

	AllPages = map[string]*Page{}
	HomePage = Page{
		Route:  "/",
		Method: METHOD_GET,
		Body:   "",
		Handler: func(w http.ResponseWriter, r *http.Request) {

		},
		JsEvents: map[string]JSEvent{},
	}
)

// Page : web server page
type Page struct {
	Route       string
	Method      int
	Body        string
	BodyWidgets []Widget
	Style       map[string]string
	Handler     func(w http.ResponseWriter, r *http.Request)
	JsEvents    map[string]JSEvent
	ExtendJS    []Js
}

// PreparedRoute : regist a func to map
func PreparedRoute(route string, fun func(w http.ResponseWriter, r *http.Request)) {
	preparedFuncs[route] = fun
}

func (page *Page) AddBody(body interface{}) *Page {

	switch body.(type) {
	case string:
		page.Body = body.(string)
	case Widget:
		page.BodyWidgets = append(page.BodyWidgets, body.(Widget))

		// L(page.Route, "debug", body)
		page.Body = body.(Widget).String()
	default:
		L("Warrning", "add body", "body is not Widget/string", body)
		return page
	}
	return page
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
func (page *Page) OnEvent(IDOrWidget interface{}, jsBody Js, toggleMethod int, callback ...func(flowData FlowData, c *websocket.Conn)) *Page {
	method := toggleMethod
	// if toggleMethod != nil {
	// 	method = toggleMethod[0]
	// }
	ID := ""
	switch IDOrWidget.(type) {
	case string:
		ID = IDOrWidget.(string)
	case Widget:
		ID = IDOrWidget.(Widget).GetID()
	default:
		L(page.Route, "Err On Init", "on Event: widget can not be ID or not string.")
		return page
	}
	L(page.Route, "Event", "regist : "+ID, method)
	page.JsEvents[ID] = JSEvent{
		Body:   jsBody,
		Toogle: method,
	}
	if callback != nil {
		RegistWebSocketCallback(ID, callback[0])
	}
	return page
}

func (page *Page) AddStyle(selector, cssbody string) *Page {
	page.Style[selector] = cssbody
	return page
}

func (page *Page) OnWebsocket(id string, call func(flow FlowData, c *websocket.Conn)) *Page {
	RegistWebSocketCallback(id, call)
	return page
}

func (page *Page) OnPost(call func(w http.ResponseWriter, r *http.Request)) *Page {
	page.Handler = call
	return page
}

// RenderPage : get base bootstrap html
func (page *Page) RenderPage(jsArea ...string) *Page {
	defer L(page.Route, "Registed!")
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
	for _, widget := range page.BodyWidgets {
		selector, method, jsevent, callback := widget.Event()
		// L(page.Route, "debug", selector, method, jsevent)
		if callback != nil {
			RegistWebSocketCallback(selector, callback)
		}
		if !strings.HasPrefix(selector, "#") {
			selector = "#" + selector
		}
		js += "\n" + Query(selector).Call(method, jsevent.String()).String()

	}
	if jsArea != nil {
		js += "\n" + jsArea[0]

		// log.Println(js)
	}
	for _, j := range page.ExtendJS {
		js += "\n" + j.String()
	}
	http.HandleFunc(page.Route, func(w http.ResponseWriter, r *http.Request) {

		switch r.Method {
		case "GET":
			if page.Method&METHOD_GET == METHOD_GET {
				AllStyle := ""
				for name, css := range page.Style {
					AllStyle += fmt.Sprintf("%s{\n%s\n};", name, css)
				}
				h := fmt.Sprintf(`<!DOCTYPE html>
			<html>
				<head>
					<meta charset="utf-8">
					<style type="text/css" >%s</style>
					<style type="text/css" custome="true" >%s</style>
					<style type="text/css" custome="true" cssfor="toast" >%s</style>
				</head>
				<body class="h-100" style="        position: absolute; width:100%%;">
			`, AllStyle, BootstrapCSS, ToastCSS) + page.Body + fmt.Sprintf(`
					<script jsName="jqueryv3.3.1">%s</script>
					<script jsName="bootstrapPop">%s</script>
					<script jsName="bootstrap">%s</script>
					<script jsName="toast">%s</script>
					<script jsname="base-functions">%s</script>
					<script jsname="uploadJS">%s</script>
				</body>
			</html>`, Jquery, BootstrapPopJS, BootstrapJS, ToastJS, baseFunctionJS+extend, js)
				fmt.Fprint(w, h)
			}
		case "POST":
			if page.Handler != nil {
				page.Handler(w, r)
			}
		}

	})
	return page
}

func StartServer(listenAddr string) {
	http.HandleFunc("/api", ApiSocketHandle)
	links := map[string]string{"/Home": "/"}
	for name := range AllPages {
		links[name] = name
	}

	search := &SearchWidget{
		InputWidget: InputWidget{
			ID: "SearchInput",
			TP: "search",
		},
	}
	search.Margin("", 10)
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
	HomePage.OnWebsocket("db", dataFunc)
	HomePage.RenderPage()

	for _, page := range AllPages {
		page.RenderPage()
	}
	log.Fatal(http.ListenAndServe(listenAddr, nil))
}

func SetRouteStyle(route, cssselector, cssbody string) {
	AllPages[route].AddStyle(cssselector, cssbody)
}

func AddPageWithNoBody(route string, UseGetPage bool) *Page {
	page := &Page{
		Route: route,
	}
	if UseGetPage {
		page.Method |= METHOD_GET
	}
	AllPages[route] = page
	return page
}
