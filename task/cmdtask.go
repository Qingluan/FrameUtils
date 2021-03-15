package task

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

type CmdObj struct {
	shellS []string
	err    error
	config *TaskConfig
}

func (cmd CmdObj) ID() string {
	c := ""

	for _, arg := range cmd.shellS {
		c += fmt.Sprintf("%x", byte(arg[0]))
	}
	return c
}

func (cmd CmdObj) Args() []string {
	return cmd.shellS
}

func (cmd CmdObj) String() string {
	d := cmd.config.LogPath()
	name := cmd.ID()
	return filepath.Join(d, name) + ".log"
}

func (cmd CmdObj) Error() error {
	return cmd.err
}

func CmdCall(tconfig *TaskConfig, args []string) (TaskObj, error) {

	var cmd *exec.Cmd
	var shellStr []string
	if runtime.GOOS == "windows" {
		shellStr = append([]string{"cmd.exe", "/c"}, args...)
	} else {
		shellStr = append([]string{"bash", "-c"}, args...)
	}
	cmdObj := CmdObj{
		shellS: shellStr,
		config: tconfig,
	}

	outfile, err := os.OpenFile(cmdObj.String(), os.O_CREATE|os.O_APPEND|os.O_RDWR, os.ModePerm)
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
	err = cmd.Start()

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
