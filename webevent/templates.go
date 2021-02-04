package webevent

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

var (
	home, _       = os.UserHomeDir()
	AppConfigPath = filepath.Join(home, "AppConfig.conf")
	HtmlBase      = `
<!DOCTYPE html>
<html>
  <head>
    <meta charset="utf-8">
	<style text="css">
	  %s
	</style>
	<style text="css">%s</style>
  </head>
  <body id="layout-body">
  %s
  %s
  </body>
</html>
	`
	WebSocketBase = `
var ws = new WebSocket("ws://" + window.location.host + "/flow");
var SendAction;
ws.onopen = function() {
	SendAction("hello","hello","hello world!");
}
var actions = {
	AddView: function(data){
		id = data.id
		value = data.value
		ele = document.getElementById(id)
		ele.innerHTML = ele.innerHTML + value 
	},
	SetView: function(data){
		id = data.id
		value = data.value
		ele = document.getElementById(id)
		ele.innerHTML = value 
	},
	SetAttr: function(data){
		id = data.id
		value = data.value
		ele = document.getElementById(id)
		kv = value.split("=")
		ele.setAttribute(kv[0],kv[1])  
	},
	AddAction: function(data){
		id = data.id
		value = data.value
		ele = document.getElementById("layout-body")
		var newScript = document.createElement('script');
		newScript.type = 'text/javascript';
		newScript.innerHTML = value;
		ele.appendChild(newScript);
	},
	GetAttr: function(data){
		id = data.id
		value = data.value
		ele = document.getElementById(id)
		kv = value.split("=")
		var v;
		if (value == "value"){
			v = ele.value;
		}else{
			v = ele.getAttribute(value);
		}
		if (data.hasOwnProperty("backid") == true){
			SendAction(data.backid, "GetAttr",v)
		}else{
			SendAction(id, "GetAttr",v)
		}
	},
	GetContent: function(data){
		id = data.id
		value = data.value
		ele = document.getElementById(id)
		v = ele.textContent
		if (data.hasOwnProperty("backid") == true){
			SendAction(data.backid, "GetContent",v)
		}else{
			SendAction(id, "GetContent",v)
		}
	},
	GetHtml: function(data){
		id = data.id
		value = data.value
		ele = document.getElementById(id)
		v = ele.innerHTML
		if (data.hasOwnProperty("backid") == true){
			SendAction(data.backid, "GetHtml",v)
		}else{
			SendAction(id, "GetHtml",v)
		}
	},
	OnDo:function(data) {
		console.log(data.value);
		eval(data.value);
	}
}

window.addEventListener("load", function(evt) {
	ws.onmessage = function(event) {
		var m = JSON.parse(event.data);
		console.debug("Received message", m.id, m.tp, m.value);
		callback = actions[m.tp];
		if (callback != null){
			callback(m);
		}
	}
	ws.onerror = function(event) {
		console.debug(event)
	}
	
})

SendAction = function(id, tp , value){
	ws.send(JSON.stringify({
		id:id,
		tp:tp,
		value:value
	}))
}
`
	BaseCssStyle = `
html{
	background: whitesmoke;
}
table.hit
{
	border-collapse: collapse;
	margin: 0 auto;
	text-align: left;
}
table.hit td, table th
{
	/* border: 1px solid #cad9ea; */
	border-radius:10px;
	margin: 5px;
	margin-left:3px;
	color: #666;
	height: 30px;
}
table.hit thead th
{
	background-color: #CCE8EB;
	/* // width: 100px; */
}
table.hit tr:nth-child(odd)
{
	background: #fff;
}
table.hit tr:nth-child(even)
{
	background: #F5FAFA;
}
input{
	position: absolute;
	width: 92%;
	border-radius: 7px;
	border-top-style: dashed;
	margin-bottom: 10px;
	bottom: 53px;
}
button {
	position: absolute;
	bottom: 20px;
	border-radius: 10px;
}
button#send{
	position: absolute;
	bottom: 20px;
}

button#close{
	left:60%;
}
button#clear{
	position: absolute;
	bottom: 20px;
	left: 30%;
	
	background: cadetblue;
	color: white;
}

td#left{
	border-radius: 20px;
	background: white;
	position: fixed;
	height: 97%;
	width: 16%;
	z-index:10000;
}
td#right{
	position:absolute;
	left:20%;
}
.dragele {
	position: absolute;
	z-index: 9;
	background-color: #f1f1f1;
	text-align: center;
	border: 1px solid #d3d3d3;
}

.dragele > .dragele-header {
	padding: 10px;
	cursor: move;
	z-index: 10;
	background-color: #2196F3;
	color: #fff;
}
	
`
	BaseHTMLStyle = `<table>
	<tr>
		<td valign="top" id="left" width="20%">
			<form>
				<button id="close">Close</button>
				<p><input id="input" name="Thing" type="text" value="geek" onkeypress="return onKeyPress(event)">
				<button id="send" onclick="return false">Send</button>
				<button id="clear" onclick="return false">Clear</button>
			</form>
		</td>
		<td valign="top" id="right" width="70%">
			<div id="output"></div>
		</td>
	</tr>
</table>
`
	BaseMainGoFile = `package main
import (
	"fmt"
	"os"

	"github.com/Qingluan/FrameUtils/webevent"
	we "github.com/Qingluan/FrameUtils/webevent"
)

func main() {
	if _, err := os.Stat("config.ini"); err == nil {
		webevent.SetConfigPath("config.ini")
	}
	we.SetOnClickAction("send", func(data we.FlowData) {
		we.GetAttrById("input", "value", func(recv we.FlowData) {
			fmt.Println("input:", recv.Value)
			we.BroadcastMsg(we.FlowData{
				Id:    "output",
				Tp:    we.AddView,
				Value: "<p > hello world</p>",
			})
			we.JsExecute(we.CSSSelect("#right").AppendChild("h2", "h2t", "Test Ht2", we.Dict{}).WithId("#input").ValueWith("").String())

		})
	})

	we.SetOnInputAction("input", func(data we.FlowData) {
		fmt.Println()
	})

	we.StartServer("%s", true)
}
`
)

func RenderJs() string {
	tmp := `<script>` + WebSocketBase + `
	%s
</script>`
	return fmt.Sprintf(tmp, strings.Join(JsActions, ";\n")+strings.Join(JsListeners, ";\n"))
}

func SetConfigPath(path string) {
	AppConfigPath = path
}

/* Render to template
...
<html>
	<head><title>${#title}</title></head>
	<div><p>
	${#idName}
	</p></div>
</html>
will replace "idName,title" with Elements, which id name matched.
*/
func Render(AppName string, eles ...ElementAble) string {
	myConfig := new(Config)
	myConfig.InitConfig(AppConfigPath)

	body := myConfig.Read(AppName, "body")
	css := myConfig.Read(AppName, "css")

	cssbuf, err := ioutil.ReadFile(css)
	if err != nil {
		log.Fatal("Render css:", err, "|", css)
	}
	bodybuf, err := ioutil.ReadFile(body)
	bodyStr := string(bodybuf)
	if err != nil {
		log.Fatal("Render body:", err, "|", body)
	}
	cssbufExtend := ""
	for _, css := range CSSCache {
		cssbufExtend += css.String()
	}
	for _, ele := range eles {
		k := "${#" + ele.GetID() + "}"
		bodyStr = strings.ReplaceAll(bodyStr, k, ele.String())
	}
	return fmt.Sprintf(HtmlBase, string(cssbuf), cssbufExtend, bodyStr, RenderJs())
}

const (
	BodyLayout = "layout-body"
)
