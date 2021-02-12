package engine

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"log"
	"os"
	"strings"
	"sync"

	"github.com/google/uuid"
)

/*ObjHeader struct:
one block Obj Header , can include many keys data
	TP [2]byte version info
	bodystart addr can found start,
	crc , will genreate a uuid to split block, to check if startaddr is ok
	haskeys can search quick , to make sure if this key in this.
	nextaddr
*/
type ObjHeader struct {
	tp            [2]byte
	Bodystartaddr [8]byte
	Bodylen       [8]byte
	SpecialCRC    [16]byte
	HasKeys       [214]byte
	Nextaddr      [8]byte
}

/*ObjBody body of ObjFIle



|  crc [16] | objBody      ...................                       |   crc [16] objBody
|	...		| ..............................................         | ...................
|	...		|  len [8] | keyslen[8] |   keys ... |    body ......    |   crc [16] next body

*/
type ObjBody struct {
	Length    [8]byte
	keyLength [8]byte
	Body      []byte
}

var (
	GlobalLock = sync.RWMutex{}
)

func NewObj(keys ...string) (obj *ObjHeader) {
	obj = new(ObjHeader)
	u, _ := uuid.NewUUID()
	u.UnmarshalBinary(obj.SpecialCRC[:])
	if keys != nil {
		obj.PushKeys(keys...)
	}
	return
}

func (objheader *ObjHeader) WriteTo(client *ObjDatabase, data []byte, keys ...string) {
	datalen := len(data)
	objheader.SetDataLen(datalen)
	body := new(ObjBody)

	dataKeys, _ := json.Marshal(&keys)
	klen := len(dataKeys)

	body.SetDataLen(len(data) + klen)
	body.SetKeyLen(klen)

	body.Body = make([]byte, int(body.Len()))
	copy(body.Body, dataKeys)
	copy(body.Body, data)

	lastHeader := client.LastHeader()
	if lastHeader == nil {
		objheader.SetBodyAddr(256)
		client.writeTo([]*ObjHeader{objheader}, []*ObjBody{body})
	} else {

		allheaders := client.AllHeaders()
		objheader.SetBodyAddr(len(allheaders)*256 + 256)
		allBodys := client.AllBody()
		allheaders = append(allheaders, objheader)
		allBodys = append(allBodys, body)
		client.writeTo(allheaders, allBodys)
	}
}

func (objbody *ObjBody) SetDataLen(l int) error {
	err := binary.Read(bytes.NewBuffer(objbody.Length[:]), binary.BigEndian, &l)
	if err != nil {
		return err
	}
	return nil
}

func (objbody *ObjBody) SetKeyLen(l int) error {
	return binary.Read(bytes.NewBuffer(objbody.keyLength[:]), binary.BigEndian, &l)
}

func (objbody *ObjBody) KeyLen() int64 {
	return int64(binary.BigEndian.Uint32(objbody.keyLength[:]))
}

func (objbody *ObjBody) Len() int64 {
	return int64(binary.BigEndian.Uint32(objbody.Length[:]))
}

func (objheader *ObjHeader) SetDataLen(l int) error {
	// binary.BigEndian.PutUint16(objheader.Bodylen, l)
	err := binary.Read(bytes.NewBuffer(objheader.Bodylen[:]), binary.BigEndian, &l)
	if err != nil {
		return err
	}
	return nil
}

func (objheader *ObjHeader) SetBodyAddr(l int) error {
	// binary.BigEndian.PutUint16(objheader.Bodylen, l)
	err := binary.Read(bytes.NewBuffer(objheader.Bodystartaddr[:]), binary.BigEndian, &l)
	if err != nil {
		return err
	}
	return nil
}

func (objheader *ObjHeader) SetNextAddr(l int) error {
	// binary.BigEndian.PutUint16(objheader.Bodylen, l)
	err := binary.Read(bytes.NewBuffer(objheader.Nextaddr[:]), binary.BigEndian, &l)
	if err != nil {
		return err
	}
	return nil
}

func (objHeader *ObjHeader) HasKey(key string) bool {
	return bytes.Contains(objHeader.HasKeys[:], []byte(key))
}

func (objHeader *ObjHeader) StartAddr() int64 {
	return int64(binary.BigEndian.Uint32(objHeader.Bodystartaddr[:]))
}

func (objHeader *ObjHeader) NextAddr() int64 {
	return int64(binary.BigEndian.Uint32(objHeader.Nextaddr[:]))
}

func (objHeader *ObjHeader) BodyLen() int64 {
	return int64(binary.BigEndian.Uint32(objHeader.Bodylen[:]))
}

func (objHeader *ObjHeader) PushKeys(keys ...string) {
	keyBytes := strings.TrimSpace(string(objHeader.HasKeys[:]))
	for _, k := range keys {
		if len(keyBytes) > 213 {
			break
		}
		if len(keyBytes) != 0 {
			keyBytes += "," + k
		} else {
			keyBytes += k
		}
	}
	a := [214]byte{}
	objHeader.HasKeys = a
	copy(objHeader.HasKeys[:], []byte(keyBytes))
}

type ObjDatabase struct {
	FileName      string
	BlockLen      int
	BodyStartAddr int
	fb            *os.File
}

func (o *ObjHeader) Bytes() []byte {
	e := make([]byte, 256)
	copy(e, o.tp[:])
	copy(e, o.Bodylen[:])
	copy(e, o.Bodystartaddr[:])
	copy(e, o.SpecialCRC[:])
	copy(e, o.HasKeys[:])
	copy(e, o.Nextaddr[:])
	return e
}

func (o *ObjBody) Bytes() []byte {
	e := make([]byte, int(o.Len()))
	copy(e, o.Length[:])
	copy(e, o.keyLength[:])
	copy(e, o.Body)
	return e
}

func (o *ObjBody) ToObj() (base *BaseObj) {
	k := o.KeyLen()
	// l := o.Len()
	keysBuf := o.Body[:k]
	realobjsBUf := o.Body[k:]
	ds := []Dict{}
	keys := Line{}
	json.Unmarshal(realobjsBUf, &ds)

	json.Unmarshal(keysBuf, &keys)

	return
}

func (odb *ObjDatabase) IterHeaders() (<-chan *ObjHeader, error) {
	fp, err := os.Open(odb.FileName)
	odb.fb = fp
	headers := make(chan *ObjHeader)
	// GlobalLock.Lock()
	// defer GlobalLock.Unlock()
	now := 0
	go func() {
		ifend := false
		var onheader *ObjHeader
		for {
			onheader, ifend, now, err = odb.readHeader(now)
			if err != nil {
				log.Fatal(err)
				break
			}
			if ifend {
				break
			}
			if onheader != nil {
				headers <- onheader
			}
		}
		close(headers)
	}()
	currentPosition, err := fp.Seek(0, 1)
	if err != nil {
		return headers, err
	}
	odb.BodyStartAddr = int(currentPosition)
	return headers, err
}

func (odb *ObjDatabase) readHeader(now int) (header *ObjHeader, end bool, newnow int, err error) {
	buf := make([]byte, 256)
	odb.fb.Seek(int64(now), os.SEEK_SET)
	GlobalLock.Lock()
	defer GlobalLock.Unlock()
	n, err := odb.fb.Read(buf)
	if err != nil {
		return
	} else if n != 256 {
		end = true
		odb.fb.Seek(int64(-n), os.SEEK_CUR)
		// err = fmt.Errorf("Err with read header: only read , %d; need read 256, file is broken", n)
		return
	}
	header = new(ObjHeader)
	if string(buf[0:2]) != "hi" {
		end = true
		odb.fb.Seek(int64(-n), os.SEEK_CUR)

		return
	}
	copy(header.tp[:], buf[0:2])
	copy(header.Bodystartaddr[:], buf[2:2+8])
	copy(header.Bodylen[:], buf[2+8:2+8+8])
	copy(header.Bodylen[:], buf[2+8:2+8+8])
	copy(header.SpecialCRC[:], buf[2+8+8:2+8+8+16])
	copy(header.HasKeys[:], buf[2+8+8+16:2+8+8+16+214])
	copy(header.Nextaddr[:], buf[2+8+8+16+214:2+8+8+16+214+8])
	newnow = now + 256
	return
}

func (odb *ObjDatabase) readBody(header *ObjHeader) (body *ObjBody, err error) {
	odb.fb.Seek(header.StartAddr(), os.SEEK_SET)
	crc := make([]byte, 16)
	GlobalLock.Lock()
	defer GlobalLock.Unlock()
	odb.fb.ReadAt(crc, 16)
	if bytes.Compare(crc, header.SpecialCRC[:]) != 0 {
		log.Fatal("Crc start failed...")
		return
	}
	body = new(ObjBody)
	buf := make([]byte, int(header.BodyLen()))
	n, err := odb.fb.Read(buf)
	if err != nil {
		return
	}
	if n != int(header.BodyLen()) {
		log.Fatal("read body broken ...")
	}
	crc = make([]byte, 16)
	odb.fb.ReadAt(crc, 16)
	if bytes.Compare(crc, header.SpecialCRC[:]) != 0 {
		log.Fatal("Crc end failed...")
		return
	}
	if len(buf) < 8 {
		log.Fatal("Body broken , too small!!!")
	}

	body.SetDataLen(int(header.BodyLen()))
	copy(body.keyLength[:], buf[:8])
	copy(body.Body, buf[8:])

	return
}

func (odb *ObjDatabase) Close() error {
	if odb.fb != nil {
		return odb.fb.Close()
	}
	return nil
}

func (odb *ObjDatabase) IterBody(filterFunc ...func(body *ObjBody) bool) (<-chan *ObjBody, error) {
	bodys := make(chan *ObjBody)
	GlobalLock.Lock()
	defer GlobalLock.Unlock()
	go func() error {
		headers, err := odb.IterHeaders()
		if err != nil {
			log.Fatal(err)
		}
		for header := range headers {
			if onbody, err := odb.readBody(header); err != nil {
				log.Fatal(err)
			} else {
				if filterFunc != nil {
					if filterFunc[0](onbody) {
						bodys <- onbody
					}
				} else {
					bodys <- onbody
				}

			}
		}
		close(bodys)
		return nil
	}()
	currentPosition, err := odb.fb.Seek(0, 1)
	if err != nil {
		return bodys, err
	}
	odb.BodyStartAddr = int(currentPosition)
	return bodys, err
}

func (odb *ObjDatabase) LastHeader() (header *ObjHeader) {
	headers, err := odb.IterHeaders()
	if err != nil {
		log.Fatal("Err with iterheaders:", err)
		return nil
	}
	for h := range headers {
		header = h
	}
	return
}

func (odb *ObjDatabase) AllHeaders() (hs []*ObjHeader) {
	hee, err := odb.IterHeaders()
	if err != nil {
		log.Fatal(err)
		return
	}
	for h := range hee {
		hs = append(hs, h)
	}
	return
}

func (odb *ObjDatabase) AllBody() (hs []*ObjBody) {
	hee, err := odb.IterBody()
	if err != nil {
		log.Fatal(err)
		return
	}
	for h := range hee {
		hs = append(hs, h)
	}
	return
}

func (odb *ObjDatabase) writeTo(headers []*ObjHeader, bodys []*ObjBody) {
	bak, err := os.Create(odb.FileName + ".bak")
	if err != nil {
		log.Fatal("Bakup err:", err)
	}
	defer bak.Close()
	crcs := [][]byte{}
	for _, h := range headers {
		bak.Write(h.Bytes())
		crcs = append(crcs, h.SpecialCRC[:])
	}

	for i, b := range bodys {
		bak.Write(crcs[i])
		bak.Write(b.Bytes())
		// bak.Write(crcs)
	}

}
