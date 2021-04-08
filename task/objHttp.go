package task

import (
	"path/filepath"

	"github.com/Qingluan/FrameUtils/utils"
)

type ObjHTTP struct {
	url    string
	args   []string
	raw    string
	err    error
	toGo   string
	kargs  utils.Dict
	config *TaskConfig
}

func (cmd ObjHTTP) ToGo() string {
	return cmd.toGo
}

func (cmd ObjHTTP) ID() string {
	return "http-" + NewID(cmd.raw)
}

func (cmd ObjHTTP) Args() []string {
	return cmd.args
}

func (cmd ObjHTTP) String() string {
	d := cmd.config.LogPath()
	name := cmd.ID()
	return filepath.Join(d, name) + ".log"
}

func (cmd ObjHTTP) Error() error {
	return cmd.err
}
