package webevent

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
)

var (
	JsActions   = []string{}
	JsListeners = []string{}
)

type Callback func(data FlowData)

func RegistJsAction(tp, howToDoJs string) {
	render := fmt.Sprintf(`actions["%s"] = function(data){
	%s
}`, tp, howToDoJs)
	JsActions = append(JsActions, render)
}

func SetOnClickAction(id string, callback func(data FlowData)) {
	tmp := `document.getElementById("%s").onclick = function(evt) {
		SendAction(this.id, "click", this.value);
		return false;
	};
	`
	if id != "" {
		JsListeners = append(JsListeners, fmt.Sprintf(tmp, id))
		SetOnServerListener(strings.TrimSpace(id), callback)
	}
}

func SetOnInputAction(id string, callback func(data FlowData)) {
	tmp := `document.getElementById("%s").onkeypress  = function(e) { 
		var keyCode = null; 
		if(e.which)
			keyCode = e.which;
		else if(e.keyCode)
			keyCode = e.keyCode;
		if(keyCode == 13) {
			SendAction(this.id, "input", this.value);
			return false; 
		}
		return true;
	}
	`
	if id != "" {
		JsListeners = append(JsListeners, fmt.Sprintf(tmp, id))
		SetOnServerListener(strings.TrimSpace(id), callback)
	}
}

func SetOnCustomAction(tp string, callback func(data FlowData), customeJs string) {
	RegistJsAction(tp, customeJs)
	id := "tmp-id"
	// RegistedAction[id] = callback
	JsListeners = append(JsListeners, customeJs)
	SetOnServerListener(id, callback)

}

func JsExecute(customejs string) {
	BroadcastMsg(FlowData{
		Id:    uuid.Must(uuid.NewRandom()).String(),
		Tp:    JsExecuteTp,
		Value: customejs,
	})
}

func GetById(id, tp, value string, callbcak Callback) {
	tmpid := uuid.Must(uuid.NewRandom())
	SetOnServerListener(tmpid.String(), callbcak)
	BroadcastMsg(FlowData{
		Id:     id,
		Tp:     tp,
		Value:  value,
		BackId: tmpid.String(),
	})
}

func GetAttrById(id string, name string, callback Callback) {
	GetById(id, GetAttr, name, callback)
}

func GetContentById(id string, callback Callback) {
	GetById(id, GetContent, "", callback)
}

func GetHtmlById(id string, callback Callback) {
	GetById(id, GetHtml, "", callback)
}

const (
	AddView     = "AddView"
	SetView     = "SetView"
	SetAttr     = "SetAttr"
	AddAction   = "AddAction"
	GetAttr     = "GetAttr"
	GetContent  = "GetContent"
	GetHtml     = "GetHtml"
	JsExecuteTp = "OnDo"
)
