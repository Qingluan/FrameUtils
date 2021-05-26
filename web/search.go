package web

import (
	"bytes"
	"html/template"
	"log"

	"github.com/Qingluan/FrameUtils/asset"
)

type SearchUI struct {
	ID      string
	Action  string
	Default string
	BtnName string
}

var (
	WEBROOT = "Res/services/TaskService/web/"
)

func NewSearchUI(id, action, btnName string) *SearchUI {
	return &SearchUI{
		ID:      id,
		Action:  action,
		BtnName: btnName,
	}
}

func (search *SearchUI) String() string {
	d, _ := asset.Asset(WEBROOT + "search.html")
	t, err := template.New("Search").Parse(string(d))
	// t.
	if err != nil {
		log.Fatal("SearchUI Err ")
	}
	buffer := bytes.NewBufferString("")
	t.Execute(buffer, search)
	return buffer.String()
}
