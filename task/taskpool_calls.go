package task

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/Qingluan/FrameUtils/utils"
	jupyter "github.com/Qingluan/jupyter/http"
	"github.com/Qingluan/merkur"
	"golang.org/x/net/proxy"
)

const ()

func _to_end(path string, buf []byte) (err error) {
	outfile, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_RDWR, os.ModePerm)
	if err != nil {
		log.Println("_to_end:", err)
		return
	}
	outfile.Write([]byte(fmt.Sprintf("\n====================== %s ======================\n", time.Now().Local().String())))
	outfile.Write(buf)
	defer outfile.Close()
	return
}

// 必须把TaskConfig 塞进TaskObj 里

func CmdCall(tconfig *TaskConfig, args []string, kargs utils.Dict) (TaskObj, error) {

	var cmd *exec.Cmd
	var shellStr []string

	// args := utils.SplitByIgnoreQuote(raw, ",")
	// args,kargs := utils.DecodeToOptions(raw)
	cmdObj := CmdObj{
		raw:    utils.EncodeToRaw(args, kargs),
		args:   args,
		config: tconfig,
	}
	if e, ok := kargs["logTo"]; ok {
		cmdObj.toGo = e.(string)
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
	// err = cmd.Wait()
	// if err != nil {
	// 	return nil, err
	// }
	return cmdObj, nil
}

func HTTPCall(tconfig *TaskConfig, args []string, kargs utils.Dict) (TaskObj, error) {
	sess := jupyter.NewSession()
	var res *jupyter.SmartResponse
	var err error
	// args, kargs := utils.DecodeToOptions(raw)
	// fmt.Println("Raw:", raw, "\nargs:", args, "\nkargs:", kargs)
	if !strings.HasPrefix(args[0], "http") {
		args[0] = "http://" + args[0]
	}
	obj := ObjHTTP{
		raw:    utils.EncodeToRaw(args, kargs),
		url:    strings.TrimSpace(args[0]),
		args:   args,
		kargs:  kargs,
		config: tconfig,
	}
	if logTo, ok := kargs["logTo"]; ok {
		obj.toGo = logTo.(string)
		delete(kargs, "logTo")
	}
	var proxy proxy.Dialer
	if v, ok := kargs["proxy"]; ok {
		proxy = merkur.NewProxyDialer(v)
		log.Println("This Task Use Proxy:", utils.Magenta(v))
		delete(kargs, "proxy")
		sess.SetProxyDialer(proxy)
	} else {
		proxy = nil
	}

	// 设置config 中任务的状态
	tconfig.MakeSureTask(obj.ID(), true)
	defer tconfig.MakeSureTask(obj.ID(), false)
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
		if res, err = sess.Get(obj.url); err != nil {
			obj.err = err
		}

	}
	if err != nil || obj.err != nil {
		// obj.err = err
		return obj, err
	}
	if len(kargs) > 0 {
		es := map[string]string{}
		// i := 0
		for k, v := range kargs {
			es[k] = v.(string)
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

func ConfigCall(tconfig *TaskConfig, args []string, kargs utils.Dict) (TaskObj, error) {
	raw := utils.EncodeToRaw(args, kargs)
	if tp, ok := kargs["type"]; ok {
		switch tp.(string) {
		case "server":
			for _, server := range args {
				server = strings.TrimSpace(server)
				if !utils.ArrayContains(tconfig.Others, server) {
					tconfig.Others = append(tconfig.Others, server)
				}
			}
		case "proxy":
			if len(args) > 0 {
				proxyStr := strings.TrimSpace(args[0])
				tconfig.Proxy = proxyStr
			}
		}
	}
	return NewBaseObj(tconfig, raw, "", "config"), nil
}
