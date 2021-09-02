package task

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
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

func TCPCall(config *TaskConfig, args []string, kargs utils.Dict) (TaskObj, error) {
	var obj BaseObj
	if e, ok := kargs["logTo"]; ok {
		obj = NewBaseObj(config, utils.EncodeToRaw(args, kargs), e.(string), "tcp")
	} else {
		obj = NewBaseObj(config, utils.EncodeToRaw(args, kargs), config.LogServer, "tcp")
	}
	if len(args) < 2 {
		return nil, fmt.Errorf("%s", "TCP call must  [ip:port] , [payload] ")
	}
	target := args[0]
	if strings.HasPrefix(target, "tcp://") {
		target = strings.SplitN(target, "://", 2)[1]
	}
	var dialer proxy.Dialer
	if v, ok := kargs["proxy"]; ok {
		dialer = merkur.NewProxyDialer(v)
		log.Println("This Task Use Proxy:", utils.Magenta(v))
		delete(kargs, "proxy")
	} else {
		if config.Proxy != "" && !IsLocalDomain(target) {
			dialer = merkur.NewProxyDialer(config.Proxy)
			log.Println("This Task Use Proxy:", utils.Magenta(config.Proxy))
		}
	}
	if dialer == nil {
		dialer = &net.Dialer{}
	}
	c, err := dialer.Dial("tcp", target)
	if err != nil {
		return nil, err
	}
	defer c.Close()

	// log.Println("Timeout:", config.Timeout, kargs)
	/*
		发送数据 并 将结果存储在本地文件
		如果时第一个pay 则只读不写
	*/
	writeAndSaveResult := func(obj TaskObj, input string, buffer *bytes.Buffer, no int) (reply []byte, n int, err error) {
		c.SetWriteDeadline(time.Now().Add(time.Duration(config.Timeout) * time.Second))

		reply = make([]byte, 4096)
		if no != 0 {
			// str := buffer.String()
			// fmt.Println(no, "==>", str)
			// fmt.Println(no,"wr")
			buf := make([]byte, 4096)
			io.CopyBuffer(c, buffer, buf)
			// c.Write([]byte(str))
		}
		c.SetReadDeadline(time.Now().Add(time.Duration(config.Timeout) * time.Second))
		if n, err = c.Read(reply); err != nil {
			// fmt.Println(no, "<==", err, reply[:n])
			return
		} else {
			// fmt.Println(no, "<==", string(reply[:n]))
			// _to_end(obj.Path(), []byte())
			_to_end(obj.Path(), []byte("<-------\n"+string(reply[:n])))
		}
		return
	}

	lastReply := make([]byte, 4096)
	lastReplyLen := 0
	for no, pay := range args[1:] {

		if strings.Contains(pay, ":") {
			if no == 0 {
				lastReply, lastReplyLen, err = writeAndSaveResult(obj, pay, nil, no)
				if err != nil {
					return nil, err
				}
			}
			fs := utils.SplitByIgnoreQuote(pay, ":")
			cond := strings.TrimSpace(fs[0])
			pay = fs[1]

			if !bytes.Contains(lastReply[:lastReplyLen], []byte(cond)) {
				log.Println("[tcp]:", "not include:", utils.Yellow(cond))
				_to_end(obj.Path(), []byte("not include:"+cond))
				break
			}
		}

		buffer := payloadSmartEncode(pay, kargs)

		lastReply, lastReplyLen, err = writeAndSaveResult(obj, pay, buffer, no+1)
		if err != nil {
			return nil, err
		}

	}
	return obj, err

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

/*
	http 抓取網頁用

*/
func HTTPCall(tconfig *TaskConfig, args []string, kargs utils.Dict) (TaskObj, error) {
	sess := jupyter.NewSession()
	var res *jupyter.SmartResponse
	var err error

	// 這個地方一定要保持原樣否則會讓原來的 NewID 失效
	targetUrl := args[0]

	if !strings.HasPrefix(targetUrl, "http") {
		targetUrl = "http://" + targetUrl
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
		if tconfig.Proxy != "" && !IsLocalDomain(obj.url) {
			proxy = merkur.NewProxyDialer(tconfig.Proxy)
			log.Println("This Task Use Proxy:", utils.Magenta(tconfig.Proxy))
			sess.SetProxyDialer(proxy)
		} else {
			proxy = nil
		}
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
			res, err = sess.Post(targetUrl, data)
		case "json":
			data := utils.Dict{}
			err = json.Unmarshal([]byte(args[2]), &data)
			if err != nil {
				return obj, err
			}
			res, err = sess.Json(targetUrl, data)
		default:
			res, err = sess.Get(targetUrl)
		}
		if err != nil {
			obj.err = err
		} else {

		}
	} else {
		if res, err = sess.Get(targetUrl); err != nil {
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
			if err = _to_end(obj.Path(), buf); err != nil {
				return obj, err
			}
		}
	} else {
		if err = _to_end(obj.Path(), res.Html()); err != nil {
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

/*
	1. 尝试进行BASE64解密

	2. 判断字符串中是否存在 ${} 等特殊符号，存在则试图从 kargs 填充
*/
func payloadSmartEncode(payload string, kargs map[string]interface{}) *bytes.Buffer {
	payload = strings.TrimSpace(payload)
	if !strings.Contains(payload, " ") && len(payload)%4 == 0 {
		if payloadbuf, err := base64.StdEncoding.DecodeString(payload); err != nil {
			payload = string(payloadbuf)
		}
	}
	var buffer *bytes.Buffer
	if len(kargs) > 0 {
		str := utils.Str(payload).Format(kargs).UnEscape()
		// log.Println("Render : ", utils.BGreen(str))
		buffer = bytes.NewBuffer([]byte(str))
	} else {
		// log.Println("Render : ", utils.BGreen(payload))
		buffer = bytes.NewBuffer([]byte(payload))
	}
	return buffer
}
