package engine

import (
	"log"
	"strings"
)

func (client *ObjDatabase) QueryBlock(info string) (*ObjHeader, *ObjBody) {
	for header := range client.IterHeaders() {
		if strings.Contains(header.GetInfo(), info) {
			b, err := client.readBody(header)
			if err != nil {
				log.Fatal("Query header success but body can not found , broken file !!!!")
			}
			return header, b
		}
	}
	return nil, nil
}

func (client *ObjDatabase) Page(num, size int) Obj {
	if client.UseInfo == "" {
		var body *ObjBody
		var err error
		for header := range client.IterHeaders() {
			body, err = client.readBody(header)
			if err != nil {
				log.Fatal("read body err but read header ok , file broken!!")
			}
			client.UseInfo = header.GetInfo()
			break
		}
		return body.ToObj().Page(num, size)
	}
	_, b := client.QueryBlock(client.UseInfo)
	return b.ToObj().Page(num, size)
}
