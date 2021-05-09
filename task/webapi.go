package task

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/Qingluan/FrameUtils/utils"
)

var (
	StopSingal    = []string{"stop"}
	ReStartSingal = []string{"restart"}
)

type TData map[string]interface{}

func jsonWrite(w io.Writer, data TData) {
	buf, err := json.Marshal(&data)
	if err != nil {
		w.Write([]byte("{\"state\":\"fail\",\"log\":\"json unmarshal failed!\"}"))
	}
	w.Write(buf)
}

func jsonWriteErr(w io.Writer, err error) {
	jsonWrite(w, TData{
		"log":   err.Error(),
		"state": "fail",
	})
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
		if configV, ok := data["config"]; ok {
			config.Update(configV)
		}
		if op, ok := data["oper"]; ok {
			log.Println(utils.Green("[", op, "]"))
			switch op {
			case "stop":
				jsonWrite(w, TData{
					"state": "ok",
					"log":   "exit ...",
				})
				DefaultTaskWaitChnnael <- StopSingal
			case "restart":
				jsonWrite(w, TData{
					"state": "ok",
					"log":   "exit ...",
				})
				DefaultTaskWaitChnnael <- ReStartSingal
			case "test":
				jsonWrite(w, TData{
					"state": "ok",
					"log":   "alive me",
				})
			case "push":
				// 如果返回ok说明在当前taskServer处理，false 转发给了target
				if reply, ok, err := config.ProtocolRound(data); ok {
					if _, ok := data["encode"]; ok {
						WithOrErr(w, data, func(args ...interface{}) TData {
							data := args[0].(string)
							encode := args[1].(string)
							if encode == "base64" {
								if buf, err := base64.StdEncoding.DecodeString(data); err != nil {
									return TData{
										"state": "fail",
										"log":   err.Error(),
									}
								} else {
									msg := ""
									for _, line := range utils.SplitByIgnoreQuote(string(buf), "\n") {
										fs := strings.SplitN(line, ",", 2)
										objType := strings.TrimSpace(fs[0])
										input := strings.TrimSpace(fs[1])
										DefaultTaskWaitChnnael <- append([]string{objType}, input)
										msg += fmt.Sprintf("%s-%s\n", objType, NewID(input))
									}
									return TData{
										"state": "ok",
										"id":    msg,
									}
								}
							} else {
								return TData{
									"state": "fail",
									"log":   "now only supported encode: base64",
								}
							}
							// objType := strings.TrimSpace(tp)
							// fmt.Println("r:", args[0], "tp:", tp)
							// fs := utils.SplitByIgnoreQuote(input, ",")

						}, "input", "encode")
					} else {
						WithOrErr(w, data, func(args ...interface{}) TData {
							input := args[0].(string)
							tp := args[1].(string)
							objType := strings.TrimSpace(tp)

							// fmt.Println("r:", args[0], "tp:", tp)
							// fs := utils.SplitByIgnoreQuote(input, ",")
							DefaultTaskWaitChnnael <- append([]string{objType}, input)
							return TData{
								"state": "ok",
								"id":    objType + "-" + NewID(input),
							}
						}, "input", "tp")
					}

				} else if err != nil {
					jsonWrite(w, TData{
						"state": "fail",
						"log":   err.Error(),
					})
				} else {
					jsonWrite(w, reply)
				}

			// 转发的处理和push一样只是返回不同
			case "forward":
				if reply, ok, err := config.ProtocolRound(data); ok {
					WithOrErr(w, data, func(args ...interface{}) TData {
						input := args[0].(string)
						tp := args[1].(string)
						objType := strings.TrimSpace(tp)
						// fs := strings.Split(input, ",")
						DefaultTaskWaitChnnael <- append([]string{objType}, input)
						return TData{
							"state": "ok",
							"id":    objType + "-" + NewID(input),
							"ip":    config.MyIP(),
						}
					}, "input", "tp")
				} else if err != nil {
					jsonWrite(w, TData{
						"state": "fail",
						"log":   err.Error(),
					})
				} else {
					jsonWrite(w, reply)
				}
			case "config":
				log.Println(utils.Green("try to config ....:", data))
				if info := config.UpdateMyConfig(data); info != "" {
					jsonWrite(w, TData{
						"state": "ok",
						"log":   info,
					})
				} else {
					jsonWrite(w, TData{
						"state": "fail",
						"log":   "no reply",
					})
				}

			case "pull":
				WithOrErr(w, data, func(args ...interface{}) TData {
					id := args[0].(string)
					d := config.LogPath()
					path := filepath.Join(d, id)
					if !strings.HasSuffix(path, ".log") {
						path += ".log"
					}
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
				"log":   "lack \"oper\" [task : ls/push/pull/clear/config/forward | sys: test/stop/restart] ",
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

func (taskconfig *TaskConfig) GetMyState() TData {
	DefaultTaskWaitChnnael <- []string{"state"}
	log := TData{}
	json.Unmarshal([]byte(<-DefaultTaskOutputChannle), &log)
	return log
}
