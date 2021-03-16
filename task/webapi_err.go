package task

import "net/http"

func ErrLackKey(w http.ResponseWriter, lackaKey string) {
	jsonWrite(w, TData{
		"state": "fail",
		"log":   "lack 'tp' in data ",
	})
}

func WithOrErr(w http.ResponseWriter, data TData, howtoDo func(args ...interface{}) TData, keys ...string) {
	Args := []interface{}{}
	for _, key := range keys {
		if v, ok := data[key]; ok {
			Args = append(Args, v)
		} else {
			ErrLackKey(w, key)
			return
		}
	}
	output := howtoDo(Args...)
	jsonWrite(w, output)
}
