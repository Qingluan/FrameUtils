package upgradeservice

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"time"

	jupyter "github.com/Qingluan/jupyter/http"
)

var (
	checkFunc func() bool
	// startCmd  string
	beforedo func()
)

func KillOtherThenRun(pid int, startCmd string) {
	var cmd *exec.Cmd
	var restartCmd *exec.Cmd
	tmpDir := os.TempDir()
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd.exe", "/c", fmt.Sprintf("\"taskkill.exe /f /pid %d \"", pid))

		restartCmd = exec.Command("cmd.exe", "/c", startCmd)
	} else {
		cmd = exec.Command("/bin/kill", "-9", fmt.Sprintf("%d", pid))
		restartCmd = exec.Command("/bin/bash", "-c", startCmd)
	}
	if cmd != nil {
		cmd.Output()
		if buf, err := restartCmd.Output(); err != nil {
			fp, err := os.OpenFile(filepath.Join(tmpDir, "restart-service.log"), os.O_APPEND|os.O_CREATE|os.O_RDWR, os.ModePerm)
			if err != nil {
				ioutil.WriteFile(filepath.Join(tmpDir, "restart-service.log"), []byte(err.Error()), os.ModePerm)
			}
			defer fp.Close()
			fp.Write(buf)
		}
	}
}

func UpgradeServer(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		buf, _ := ioutil.ReadAll(r.Body)
		J := make(map[string]string)
		err := json.Unmarshal(buf, &J)
		if err == nil {
			if pidS, ok := J["pid"]; ok {
				if pid, err := strconv.Atoi(pidS); err == nil {
					if startCmd, ok := J["cmd"]; ok {
						KillOtherThenRun(pid, startCmd)
						w.Write([]byte("UPgrading....."))
					}
				}
			}
		}
	}
}

func StartUpgradeClient(upgradeServiceURL, cmd string, waitSec int, check func() bool, before func()) {
	tic := time.NewTicker(time.Duration(waitSec) * time.Second)
	for {
		select {
		case <-tic.C:
			if check() {
				before()
				sess := jupyter.NewSession()
				if res, err := sess.Json(upgradeServiceURL, map[string]string{
					"pid": fmt.Sprintf("%d", os.Getpid()),
					"cmd": cmd,
				}); err == nil {
					buf, _ := ioutil.ReadAll(res.Body)
					fmt.Println(string(buf))
				}
			}
		default:
			time.Sleep(5 * time.Second)
		}
	}
}

func StartUpgradeService(address string) {
	http.HandleFunc("/upgrade", UpgradeServer)
	log.Fatal(http.ListenAndServe(address, nil))
}
