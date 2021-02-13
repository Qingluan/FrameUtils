package engine

import (
	"log"
	"strings"
)

func (client *ObjDatabase) QueryBlock(info string) *ObjBody {
	for header := range client.IterHeaders() {
		if strings.Contains(header.GetInfo(), info) {
			b, err := client.readBody(header)
			if err != nil {
				log.Fatal("Query header success but body can not found , broken file !!!!")
			}
			return b
		}
	}
	return nil
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
	return client.QueryBlock(client.UseInfo).ToObj().Page(num, size)
}
