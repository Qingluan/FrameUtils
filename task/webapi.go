package task

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"
	"strings"
)

type TData map[string]interface{}

func jsonWrite(w io.Writer, data TData) {
	buf, err := json.Marshal(&data)
	if err != nil {
		w.Write([]byte("{\"state\":\"fail\",\"log\":\"json unmarshal failed!\"}"))
	}
	w.Write(buf)
}

func getCmdTaskID(args []string) string {
	c := ""

	for _, arg := range args {
		c += fmt.Sprintf("%x", byte(arg[0]))
	}
	return c
}

func (config TaskConfig) TaskHandle(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Println(err)
			return
		}
		data := TData{}
		json.Unmarshal(body, &data)
		if op, ok := data["oper"]; ok {
			switch op {
			case "pull":
				if input, ok := data["input"]; ok {
					fs := strings.Split(input.(string), ",")
					DefaultTaskWaitChnnael <- append([]string{"cmd"}, fs...)
					jsonWrite(w, TData{
						"state": "ok",
						"id":    getCmdTaskID(fs),
					})
				}
			case "state":
				if id, ok := data["id"]; ok {
					d := config.LogPath()
					path := filepath.Join(d, id.(string)) + ".log"
					buf, err := ioutil.ReadFile(path)
					if err != nil {
						jsonWrite(w, TData{
							"state": "fail",
							"log":   err.Error(),
						})
					} else {
						jsonWrite(w, TData{
							"state": "ok",
							"log":   string(buf),
						})
					}
				}
			}
		} else {
			jsonWrite(w, TData{
				"state": "fail",
				"log":   "lack \"oper\" ",
			})
		}
	default:
		DefaultTaskWaitChnnael <- append([]string{"state"})
		log := TData{}
		json.Unmarshal([]byte(<-DefaultTaskOutputChannle), &log)
		jsonWrite(w, TData{
			"state": "ok",
			"log":   log,
		})
	}
}
