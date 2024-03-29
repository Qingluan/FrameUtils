package engine

import (
	"bufio"
	"io"
	"log"
	"os"
	"os/exec"

	"github.com/Qingluan/FrameUtils/utils"
)

type BaseConnection struct {
	Host         string
	tableName    string
	filterName   string
	connecitonTp string
	tables       map[string]utils.Line
}

func (self *BaseConnection) Close() error {
	return nil
}
func (self *BaseConnection) header(...int) (l utils.Line) {
	return
}
func (self *BaseConnection) Tp() string {
	return self.connecitonTp
}

func (self *BaseConnection) Query(queryCmd string) (reader io.ReadCloser, err error) {
	cmd := exec.Command(queryCmd)
	cmd.Env = os.Environ()
	reader, err = cmd.StdoutPipe()
	return
}

func (self *BaseConnection) Iter(tableFilter ...string) <-chan utils.Line {
	ch := make(chan utils.Line)
	reader, err := self.AllLine(tableFilter...)
	if err != nil {
		log.Fatal("All line err:", err)
	}
	go func() {
		// c := 0

		lines := bufio.NewReader(reader)
		for {
			one, err := lines.ReadString(byte('\n'))
			if err == io.EOF {
				break
			}
			// include tableName as first column
			ch <- self.ParseValue(one)
		}
		close(ch)
	}()
	return ch
}
func (self *BaseConnection) ToJson(tables ...string) (ds []utils.Dict) {
	for line := range self.Iter(tables...) {
		header := self.GetHead(line[0])
		ds = append(ds, line[1:].FromKey(header))
	}
	return
}

func (self *BaseConnection) GetHead(k string) utils.Line {
	return self.tables[k]
}

func (self *BaseConnection) AllHeader() (err error) {
	return
}

func (self *BaseConnection) ParseValue(value string) (line utils.Line) {
	return
}

func (self *BaseConnection) AllLine(tablefilter ...string) (reader io.ReadCloser, err error) {
	return
}
