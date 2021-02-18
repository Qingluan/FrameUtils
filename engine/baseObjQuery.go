package engine

import (
	"log"
	"os"
	"path/filepath"
	"strings"
)

func (self *BaseObj) Page(num int, size int) (page Obj) {
	start := num * size
	end := num*size + size
	n := 0
	lines := []Line{}
	for line := range self.Iter() {
		if n >= start && n < end {
			lines = append(lines, line)
		}
		n++
	}
	return FromLines(lines, self.Header())
}

func (self *BaseObj) WithTmpDB(dbName string) *ObjDatabase {
	tmpFile := os.TempDir()
	dbName = strings.ReplaceAll(dbName, ":", "-")

	dbName = strings.ReplaceAll(dbName, "/", "_")
	clientFile := filepath.Join(tmpFile, dbName)

	client := NewObjClient(clientFile)
	if _, err := os.Stat(clientFile); err == nil {
		body, key, err := self.Marshal()
		if err != nil {
			log.Fatal("create db err:", err)
		}
		return client.UpdateBlock(self.Tp(), body, key...)
	} else {
		body, key, err := self.Marshal()
		if err != nil {
			log.Fatal("create db err:", err)
		}
		return client.CreateBlock(self.Tp(), body, key...)
	}
}
