package web

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Qingluan/FrameUtils/engine"
	"github.com/gorilla/websocket"
)

func WithTmpDB(name string) (db *engine.ObjDatabase, exists bool) {
	tmp := os.TempDir()
	name = strings.ReplaceAll(name, ":", "-")

	name = strings.ReplaceAll(name, "/", "_")
	db = engine.NewObjClient(filepath.Join(tmp, name))
	exists = db.Exists()
	return
}

func dataFunc(flow FlowData, c *websocket.Conn) {
	d := make(map[string]interface{})
	if err := json.Unmarshal([]byte(flow.Value), &d); err != nil {
		c.WriteJSON(FlowData{Value: err.Error(), Tp: "Err"})
	}
	if db, ok := d["db"]; ok {
		dbh, ok := WithTmpDB(db.(string))
		if !ok {
			c.WriteJSON(FlowData{
				Tp:    "Err",
				Value: "no such db now:" + db.(string),
			})
			return
		}
		num := d["num"].(int)
		size := 100
		if siz, ok := d["size"]; ok {
			size = siz.(int)
		}
		html := dbh.Page(num, size).ToHTML("data")
		nextdataQueryD := map[string]interface{}{
			"db":   db.(string),
			"size": size,
			"num":  num + 1,
		}
		nextDataQuery, _ := json.Marshal(nextdataQueryD)
		html = fmt.Sprintf(`<button type="button" class="btn btn-primary" onclick="return SendAction(\"db\",\"\", '%s')"  >%s <span class="badge badge-light">%d</span></button>`, string(nextDataQuery), "Next Page", num) + html
		c.WriteJSON(FlowData{
			Tp:    "SetView",
			Value: html,
		})
	}
}
