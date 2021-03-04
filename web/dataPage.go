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
	fmt.Println(flow)
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
		num := int(d["num"].(float64))
		size := 100
		if _, ok := d["size"]; ok {
			size = int(d["size"].(float64))
		}
		html := dbh.Page(num, size).ToHTML("data", func(r, c int, v string) string {
			if engine.SmartExtractURL.Match([]byte(v)) {
				// fmt.Println("ok:", "match")
				return fmt.Sprintf("<a href=\"%s\" >%s</a>", v, v)
			}
			// fmt.Println("no:", "match", v)
			return v
		})
		// nextdataQueryD := map[string]interface{}{
		// 	"db":   db.(string),
		// 	"size": size,
		// 	"num":  num + 1,
		// }
		// nextDataQuery, _ := json.Marshal(nextdataQueryD)
		// html = fmt.Sprintf(`<button type="button" class="btn btn-primary" onclick="return SendAction(\"db\",\"\", '%s')"  >%s <span class="badge badge-light">%d</span></button>`, string(nextDataQuery), "Next Page", num) + html
		c.WriteJSON(FlowData{
			Tp:    "SetView",
			Value: html,
		})
	}
}
