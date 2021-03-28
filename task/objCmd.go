package task

import (
	"path/filepath"
)

type CmdObj struct {
	pre    []string
	args   []string
	raw    string
	err    error
	config *TaskConfig
}

func (cmd CmdObj) ID() string {
	return "cmd-" + NewID(cmd.raw)
}

func (cmd CmdObj) Args() []string {
	return append(cmd.pre, cmd.args...)
}

func (cmd CmdObj) String() string {
	d := cmd.config.LogPath()
	name := cmd.ID()
	return filepath.Join(d, name) + ".log"
}

func (cmd CmdObj) Error() error {
	return cmd.err
}
