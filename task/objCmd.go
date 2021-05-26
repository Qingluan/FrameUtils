package task

import (
	"path/filepath"
)

type BaseObj struct {
	pre    []string
	args   []string
	raw    string
	idpre  string
	err    error
	toGo   string
	config *TaskConfig
}

func NewBaseObj(conf *TaskConfig, raw, logto, idpre string) BaseObj {
	return BaseObj{
		toGo:   logto,
		raw:    raw,
		config: conf,
		idpre:  idpre,
	}
}

func (cmd BaseObj) ID() string {
	return cmd.idpre + "-" + NewID(cmd.raw)
}
func (cmd BaseObj) ToGo() string {
	return cmd.toGo
}

func (cmd BaseObj) Args() []string {
	return append(cmd.pre, cmd.args...)
}

func (cmd BaseObj) String() string {
	d := cmd.config.LogPath()
	name := cmd.ID()
	return filepath.Join(d, name) + ".log"
}

func (cmd BaseObj) Path() string {
	d := cmd.config.LogPath()
	name := cmd.ID()
	return filepath.Join(d, name) + ".log"

}

func (cmd BaseObj) Error() error {
	return cmd.err
}

type CmdObj struct {
	pre    []string
	args   []string
	raw    string
	err    error
	toGo   string
	config *TaskConfig
}

func (cmd CmdObj) ID() string {
	return "cmd-" + NewID(cmd.raw)
}
func (cmd CmdObj) ToGo() string {
	return cmd.toGo
}

func (cmd CmdObj) Args() []string {
	return append(cmd.pre, cmd.args...)
}

func (cmd CmdObj) Path() string {
	d := cmd.config.LogPath()
	name := cmd.ID()
	return filepath.Join(d, name) + ".log"

}

func (cmd CmdObj) String() string {
	d := cmd.config.LogPath()
	name := cmd.ID()
	return filepath.Join(d, name) + ".log"
}

func (cmd CmdObj) Error() error {
	return cmd.err
}
