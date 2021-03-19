package task

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/Qingluan/FrameUtils/utils"
	jupyter "github.com/Qingluan/jupyter/http"
)

const ()

func _to_end(path string, buf []byte) (err error) {
	outfile, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_RDWR, os.ModePerm)
	outfile.Write([]byte(fmt.Sprintf("\n====================== %s ======================\n", time.Now().Local().String())))
	outfile.Write(buf)
	defer outfile.Close()
	return
}

// 必须把TaskConfig 塞进TaskObj 里

func CmdCall(tconfig *TaskConfig, args []string, extensions ...string) (TaskObj, error) {

	var cmd *exec.Cmd
	var shellStr []string
	cmdObj := CmdObj{
		args:   args,
		config: tconfig,
	}
	if runtime.GOOS == "windows" {
		cmdObj.pre = []string{"cmd.exe", "/c"}
	} else {
		cmdObj.pre = []string{"bash", "-c"}
	}

	shellStr = append(cmdObj.pre, cmdObj.args...)
	outfile, err := os.OpenFile(cmdObj.String(), os.O_CREATE|os.O_APPEND|os.O_RDWR, os.ModePerm)
	outfile.Write([]byte(fmt.Sprintf("\n====================== %s ======================\n", time.Now().Local().String())))
	if err != nil {
		cmdObj.err = err
		return cmdObj, err
	}
	defer outfile.Close()

	// 设置config 中任务的状态
	tconfig.MakeSureTask(cmdObj.ID(), true)
	defer tconfig.MakeSureTask(cmdObj.ID(), false)

	cmd = exec.Command(shellStr[0], shellStr[1:]...)
	cmd.Stdout = outfile
	cmd.Stderr = outfile
	err = cmd.Run()

	if err != nil {
		cmdObj.err = err
		return cmdObj, nil
	}
	err = cmd.Wait()
	if err != nil {
		return nil, err
	}
	return cmdObj, nil
}

func HTTPCall(tconfig *TaskConfig, args []string, extensions ...string) (TaskObj, error) {
	sess := jupyter.NewSession()
	var res *jupyter.SmartResponse
	var err error
	obj := ObjHTTP{

		url:         strings.TrimSpace(args[0]),
		args:        args,
		afterHandle: extensions,
		config:      tconfig,
	}
	if len(args) > 2 {
		switch strings.TrimSpace(args[1]) {
		case "post":
			data := utils.BDict{}
			err = json.Unmarshal([]byte(args[2]), &data)
			if err != nil {
				return obj, err
			}
			res, err = sess.Post(obj.url, data)
		case "json":
			data := utils.BDict{}
			err = json.Unmarshal([]byte(args[2]), &data)
			if err != nil {
				return obj, err
			}
			res, err = sess.Json(obj.url, data)
		default:
			res, err = sess.Get(obj.url)
		}
		if err != nil {
			obj.err = err
		} else {

		}
	} else {
		res, err = sess.Get(strings.TrimSpace(args[0]))

	}

	if extensions != nil {
		es := map[string]string{}
		for i, e := range extensions {
			es[fmt.Sprintf("%d", i)] = e
		}
		o := res.CssExtract(es)
		buf, err := json.Marshal(o)
		if err != nil {
			obj.err = err
		} else {
			if err = _to_end(obj.String(), buf); err != nil {
				return obj, err
			}
		}
	} else {
		if err = _to_end(obj.String(), res.Html()); err != nil {
			return obj, err
		}
	}
	return obj, err

}
