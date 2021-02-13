package engine

import (
	"log"
	"os"
	"path/filepath"
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
	clientFile := filepath.Join(tmpFile, dbName)
	client := NewObjClient(clientFile)
	body, key, err := self.Marshal()
	if err != nil {
		log.Fatal("create db err:", err)
	}
	return client.CreateBlock(self.Tp(), body, key...)
}
