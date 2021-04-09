package utils

import (
	"fmt"
	"os"
	"os/exec"
	"time"
)

func Deamon(deamonFlag string) {
	args := []string{}
	for _, a := range os.Args[1:] {
		if a == deamonFlag {
			continue
		}
		args = append(args, a)
	}
	cmd := exec.Command(os.Args[0], args...)
	cmd.Stdin = nil
	cmd.Stdout = nil
	cmd.Stderr = nil
	// cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}

	cmd.Start()

	time.Sleep(2 * time.Second)
	fmt.Printf("%s [PID] %d running...\n", os.Args[0], cmd.Process.Pid)
	os.Exit(0)
}
