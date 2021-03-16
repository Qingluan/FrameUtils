package task

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
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

func (config *TaskConfig) TaskHandle(w http.ResponseWriter, r *http.Request) {
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
				WithOrErr(w, data, func(args ...interface{}) TData {
					input := args[0].(string)
					tp := args[1].(string)
					objType := strings.TrimSpace(tp)
					fs := strings.Split(input, ",")
					DefaultTaskWaitChnnael <- append([]string{objType}, fs...)
					return TData{
						"state": "ok",
						"id":    objType + "-" + NewID(fs),
					}
				}, "input", "tp")

			case "state":
				WithOrErr(w, data, func(args ...interface{}) TData {
					id := args[0].(string)
					d := config.LogPath()
					path := filepath.Join(d, id) + ".log"
					buf, err := ioutil.ReadFile(path)
					if err != nil {
						return TData{
							"state": "fail",
							"log":   err.Error(),
						}
					} else {
						return TData{
							"state": "ok",
							"log":   string(buf),
						}
					}
				}, "id")
			case "clear":
				WithOrErr(w, data, func(args ...interface{}) TData {
					id := args[0].(string)
					if fs, err := ioutil.ReadDir(config.LogPath()); err != nil {
						return TData{"state": "fail", "log": err.Error()}
					} else {
						res := []string{}
						msg := ""
						r := config.LogPath()
						for _, f := range fs {
							if strings.Contains(f.Name(), id) {
								if err := os.Remove(filepath.Join(r, f.Name())); err != nil {
									msg += "\n" + f.Name() + " : " + err.Error()
								} else {
									res = append(res, f.Name())
								}
							}
						}
						return TData{"state": "ok", "log": TData{"success": res, "err": msg}}
					}
				}, "id")
			case "ls":
				WithOrErr(w, data, func(args ...interface{}) TData {
					if fs, err := ioutil.ReadDir(config.LogPath()); err != nil {
						return TData{"state": "fail", "log": err.Error()}
					} else {
						paths := []string{}
						for _, f := range fs {
							paths = append(paths, f.Name())
						}
						return TData{"state": "ok", "log": paths}
					}
				})

			}
		} else {
			jsonWrite(w, TData{
				"state": "fail",
				"log":   "lack \"oper\" ",
			})
		}
	default:
		DefaultTaskWaitChnnael <- []string{"state"}
		log := TData{}
		json.Unmarshal([]byte(<-DefaultTaskOutputChannle), &log)
		jsonWrite(w, TData{
			"state": "ok",
			"log":   log,
		})
	}
}
