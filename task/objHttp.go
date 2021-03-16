package task

import "path/filepath"

type ObjHTTP struct {
	url         string
	args        []string
	err         error
	afterHandle []string
	config      *TaskConfig
}

func (cmd ObjHTTP) ID() string {
	return "http-" + NewID(cmd.args)
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
