package engine

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"

	"github.com/Qingluan/FrameUtils/utils"
	"github.com/google/uuid"
)

const (
	BodyHeaderLen = 34
)

/*ObjHeader struct:
one block Obj Header , can include many keys data
	TP [2]byte version info
	bodystart addr can found start,
	crc , will genreate a uuid to split block, to check if startaddr is ok
	Info can search quick , to make sure if this key in this.
	nextaddr
*/
type ObjHeader struct {
	Tp            [2]byte
	Bodystartaddr [8]byte
	Bodylen       [8]byte
	Crc           [16]byte
	Info          [214]byte
	Nextaddr      [8]byte
}

/*ObjBody body of ObjFIle

header : 16+ 8 + 18 = 32

|  crc [16] | objBody      ...................                       |   crc [16] objBody
|	...		| ..............................................         | ...................
|	...		|  len [8] | keyslen[8] |   keys ... |    body ......    |   crc [16] next body

*/
type ObjBody struct {
	Tp        [2]byte
	Crc       [16]byte
	Length    [8]byte
	keyLength [8]byte
	Body      []byte
}

type ObjDatabase struct {
	FileName      string
	BlockLen      int
	BodyStartAddr int
	UseInfo       string
	fb            *os.File
}

var (
	GlobalLock = sync.RWMutex{}
)

func NewObj(keys ...string) (obj *ObjHeader) {
	obj = new(ObjHeader)
	u, _ := uuid.NewRandom()
	fmt.Println("new uuid crc:", u)
	b, _ := u.MarshalBinary()
	// u.UnmarshalBinary(obj.SpecialCRC[:])
	copy(obj.Tp[:], []byte("hi"))
	copy(obj.Crc[:], b)
	if keys != nil {
		obj.PushKeys(keys...)
	}
	return
}

func (objheader *ObjHeader) String() string {
	return fmt.Sprintf("(CRC:%x, addr:%d, len:%d) ", objheader.Crc, objheader.StartAddr(), objheader.BodyLen())
}

func (objheader *ObjHeader) Update(client *ObjDatabase, info string, data []byte, keys ...string) {
	dataKeys, _ := json.Marshal(&keys)
	// fmt.Println("write to : /keys:", string(dataKeys))
	klen := len(dataKeys)
	datalen := len(data)

	objheader.SetDataLen(int64(datalen + klen + BodyHeaderLen))
	body := new(ObjBody)
	ic := int64(datalen - int(objheader.BodyLen()))
	body.SetDataLen(int64(datalen + klen))
	body.SetKeyLen(int64(klen))

	body.Body = make([]byte, int(body.Len()))
	copy(body.Tp[:], []byte("ha"))
	copy(body.Crc[:], objheader.Crc[:])
	copy(body.Body[:klen], dataKeys)
	copy(body.Body[klen:], data)

	lastHeader := client.LastHeader()
	oldCRC := objheader.UUID()
	if lastHeader == nil {

		objheader.SetBodyAddr(256)
		client.writeTo([]*ObjHeader{objheader}, []*ObjBody{body})
		fmt.Println("init :", objheader)
	} else {
		// client.AllHeadersSync()
		allheaders := client.AllHeaders()
		count := len(allheaders)
		// nowStart := 0
		hit := false
		allheaders = ObjHeaders(allheaders).With(func(h *ObjHeader) {
			if h.GetInfo() == info {
				hit = true
			}
			if hit {
				h.SetBodyAddr(h.BodyLen() + ic)
			}

		})

		lastHeader = allheaders[count-1]

		fmt.Println("update :", objheader)
		allBodys := client.AllBody()
		hitI := -1
		for i, b := range allBodys {
			if b.UUID() == oldCRC {
				hitI = i
			}
		}
		allBodys[hitI] = body
		// allheaders = append(allheaders, objheader)
		// allBodys = append(allBodys, body)
		client.writeTo(allheaders, allBodys)

	}
}

func (objheader *ObjHeader) Write(client *ObjDatabase, data []byte, keys ...string) {
	dataKeys, _ := json.Marshal(&keys)
	// fmt.Println("write to : /keys:", string(dataKeys))
	klen := len(dataKeys)
	datalen := len(data)

	objheader.SetDataLen(int64(datalen + klen + BodyHeaderLen))
	body := new(ObjBody)

	body.SetDataLen(int64(datalen + klen))
	body.SetKeyLen(int64(klen))

	body.Body = make([]byte, int(body.Len()))
	copy(body.Tp[:], []byte("ha"))
	copy(body.Crc[:], objheader.Crc[:])
	copy(body.Body[:klen], dataKeys)
	copy(body.Body[klen:], data)

	lastHeader := client.LastHeader()

	if lastHeader == nil {

		objheader.SetBodyAddr(256)
		client.writeTo([]*ObjHeader{objheader}, []*ObjBody{body})
		fmt.Println("init :", objheader)
	} else {
		// client.AllHeadersSync()
		allheaders := client.AllHeaders()
		count := len(allheaders)
		allheaders = ObjHeaders(allheaders).With(func(h *ObjHeader) {
			h.SetBodyAddr(h.StartAddr() + 256)
		})

		lastHeader = allheaders[count-1]
		objheader.SetBodyAddr(lastHeader.StartAddr() + lastHeader.BodyLen())

		fmt.Println("add :", objheader)
		allBodys := client.AllBody()
		allheaders = append(allheaders, objheader)
		allBodys = append(allBodys, body)
		err := client.writeTo(allheaders, allBodys)
		fmt.Println("debug", "err", err)
	}
}

func (objhead *ObjHeader) SetInfo(info string) {
	copy(objhead.Info[:], []byte(info))
}

func (objhead *ObjHeader) GetInfo() string {
	return strings.TrimSpace(string(objhead.Info[:]))
}

func (objbody *ObjBody) SetDataLen(l int64) {
	binary.BigEndian.PutUint32(objbody.Length[:], uint32(l))
}

func (objbody *ObjBody) SetKeyLen(l int64) {
	// return binary.Read(bytes.NewBuffer(objbody.keyLength[:]), binary.BigEndian, &l)
	binary.BigEndian.PutUint32(objbody.keyLength[:], uint32(l))
}

func (objbody *ObjBody) KeyLen() int64 {
	return int64(binary.BigEndian.Uint32(objbody.keyLength[:]))
}

func (objbody *ObjBody) Len() int64 {
	return int64(binary.BigEndian.Uint32(objbody.Length[:]))
}

func (objheader *ObjHeader) SetDataLen(l int64) error {
	// binary.BigEndian.PutUint16(objheader.Bodylen, l)
	// err := binary.Read(bytes.NewBuffer([:]), binary.BigEndian, &l)
	binary.BigEndian.PutUint32(objheader.Bodylen[:], uint32(l))
	// if err != nil {
	// 	return err
	// }
	return nil
}

func (objheader *ObjHeader) SetBodyAddr(l int64) {
	// binary.BigEndian.PutUint16(objheader.Bodylen, l)
	binary.BigEndian.PutUint32(objheader.Bodystartaddr[:], uint32(l))
	// err := binary.Read(bytes.NewBuffer(objheader.Bodystartaddr[:]), binary.BigEndian, &l)
}

func (objheader *ObjHeader) SetNextAddr(l int64) {
	binary.BigEndian.PutUint32(objheader.Nextaddr[:], uint32(l))
}

func (objHeader *ObjHeader) HasKey(key string) bool {
	return bytes.Contains(objHeader.Info[:], []byte(key))
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
	keyBytes := strings.TrimSpace(string(objHeader.Info[:]))
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
	objHeader.Info = a
	copy(objHeader.Info[:], []byte(keyBytes))
}

func (o *ObjHeader) UUID() string {
	uu, err := uuid.NewUUID()
	if err != nil {
		log.Fatal("uuid broken :", err)
	}
	uu.UnmarshalBinary(o.Crc[:])
	return uu.String()
}

func (o *ObjHeader) Bytes() []byte {
	var buf bytes.Buffer
	binary.Write(&buf, binary.BigEndian, o)
	return buf.Bytes()
}

func (o *ObjHeader) FromBytes(data []byte) {
	buf := bytes.NewBuffer(data)
	binary.Read(buf, binary.BigEndian, o)
}

func (o *ObjBody) Bytes() []byte {
	var buf bytes.Buffer
	// enc := gob.NewEncoder(&buf)
	// enc.Encode(o)
	binary.Write(&buf, binary.BigEndian, &o.Tp)
	binary.Write(&buf, binary.BigEndian, &o.Crc)
	binary.Write(&buf, binary.BigEndian, &o.Length)

	binary.Write(&buf, binary.BigEndian, &o.keyLength)

	binary.Write(&buf, binary.BigEndian, &o.Body)
	// buf.Write(o.Crc[:])
	// buf.Write(o.Length[:])
	// buf.Write(o.keyLength[:])
	// buf.Write(o.Body)
	return buf.Bytes()
}

func (o *ObjBody) FromBytes(data []byte) {

	// fmt.Println(o)
	// buffer := bytes.NewBuffer(data)
	// fmt.Println(data)
	// binary.Read(buffer, binary.BigEndian, o.Crc)
	// fmt.Println(data)
	copy(o.Tp[:], data[:2])
	// fmt.Println("crc:", data[2:18])
	copy(o.Crc[:], data[2:18])

	// fmt.Println("len:", data[18:26])
	copy(o.Length[:], data[18:26])

	// fmt.Println("klen:", data[26:34])
	copy(o.keyLength[:], data[26:34])

	// fmt.Println("from bytes:", o.UUID(), "addr:", "len:", o.Len(), "key len:", o.KeyLen(), "bodyheader:", BodyHeaderLen)
	o.Body = make([]byte, int(o.Len()))
	copy(o.Body, data[BodyHeaderLen:int(o.Len()+BodyHeaderLen)])
}

func (o *ObjBody) UUID() string {
	uu, err := uuid.NewUUID()
	if err != nil {
		log.Fatal("uuid broken :", err)
	}
	uu.UnmarshalBinary(o.Crc[:])
	return uu.String()
}

func (o *ObjBody) ToObj() (base *BaseObj) {
	k := o.KeyLen()
	// l := o.Len()
	keysBuf := o.Body[:k]
	realobjsBUf := o.Body[k:]
	ds := []utils.Dict{}
	keys := utils.Line{}
	json.Unmarshal(realobjsBUf, &ds)

	json.Unmarshal(keysBuf, &keys)
	return &BaseObj{
		&JsonObj{
			Header:    keys,
			Datas:     ds,
			tableName: "unknow",
		},
	}
}

func (odb *ObjDatabase) Open() (err error) {
	var fp *os.File
	if _, err := os.Stat(odb.FileName); err != nil {
		fp, err = os.Create(odb.FileName)
	} else {
		fp, err = os.Open(odb.FileName)
	}
	odb.fb = fp
	return
}

func (odb *ObjDatabase) IterHeaders() <-chan *ObjHeader {
	var err error

	headers := make(chan *ObjHeader)
	// GlobalLock.Lock()
	// defer GlobalLock.Unlock()
	now := 0
	go func() {
		ifend := false
		var onheader *ObjHeader
		for {
			onheader, ifend, now, err = odb.readHeader(now)
			// fmt.Println("now:", now, onheader)
			if err != nil {
				if err.Error() != "EOF" {
					log.Fatal(err)
				} else {
					err = nil
				}
				break
			}
			if ifend {
				odb.BodyStartAddr = now

				break
			}
			if onheader != nil && !ifend {
				headers <- onheader
			}
		}
		close(headers)
	}()
	// currentPosition, err := fp.Seek(0, 1)
	if err != nil {
		log.Fatal("Err iter headers:", err)
		return headers
	}
	return headers
}

func (odb *ObjDatabase) readHeader(now int) (header *ObjHeader, end bool, newnow int, err error) {

	odb.Open()
	defer odb.Close()
	GlobalLock.Lock()
	defer GlobalLock.Unlock()
	buf := make([]byte, 256)
	ret, err := odb.fb.Seek(int64(now), os.SEEK_SET)
	if err != nil {
		log.Fatal("seek ret err:", ret, err)
		return nil, true, -1, err
	}
	n, err := odb.fb.Read(buf)
	if err != nil {
		fmt.Println("readHeader err:", err)
		end = true
		return
	} else if n != 256 {
		end = true
		odb.fb.Seek(int64(-n), os.SEEK_CUR)
		fmt.Println("readHeader not 256:", buf)
		return
	}
	// fmt.Println(now)
	header = new(ObjHeader)
	header.FromBytes(buf)
	if string(header.Tp[:]) != "hi" {
		end = true
		odb.fb.Seek(int64(-n), os.SEEK_CUR)
		// fmt.Println("readHeader:", buf[:2])
		return
	}
	newnow = now + 256
	return
}

func (odb *ObjDatabase) readBody(header *ObjHeader) (body *ObjBody, err error) {

	odb.Open()
	defer odb.Close()
	GlobalLock.Lock()
	defer GlobalLock.Unlock()

	// fmt.Println("To:", header.StartAddr(), "Header len:", header.BodyLen())
	ret, err := odb.fb.Seek(header.StartAddr(), os.SEEK_SET)
	if err != nil {
		log.Println("seek ret:", ret)
		return nil, err
	}
	// ee, _ := odb.fb.Seek(0, 1)
	data := make([]byte, int(header.BodyLen()))
	if n, err := odb.fb.Read(data); err != nil {
		return nil, err
	} else if n != int(header.BodyLen()) {
		return nil, fmt.Errorf("broken header or body!!: %d/%d", n, header.BodyLen())
	}

	// fmt.Println("read body / now:", data[:16])
	body = new(ObjBody)
	body.FromBytes(data)
	// fmt.Println("readBody / addr:", header.StartAddr(), "crc:", body.UUID(), "len:", body.Len())
	if bytes.Compare(body.Crc[:], header.Crc[:]) != 0 {
		log.Fatal("Crc start failed...", body.Crc, header.Crc)
		return
	}
	// fmt.Println("Body:", string(body.Body))
	return
}

func (odb *ObjDatabase) Close() error {
	if odb.fb != nil {
		return odb.fb.Close()
	}
	return nil
}

func (odb *ObjDatabase) IterBody(filterFunc ...func(body *ObjBody) bool) <-chan *ObjBody {
	bodys := make(chan *ObjBody)

	GlobalLock.Lock()
	defer GlobalLock.Unlock()
	ifempty := true
	go func() error {
		headers := odb.IterHeaders()

		for header := range headers {
			ifempty = false
			// fmt.Println("success / header:", header)
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
	if ifempty {
		return bodys
	}
	currentPosition, err := odb.fb.Seek(0, 1)
	if err != nil {
		log.Fatal("Err Iter Body:", err)
		return bodys
	}
	odb.BodyStartAddr = int(currentPosition)
	return bodys
}

func (odb *ObjDatabase) LastHeader() (header *ObjHeader) {
	headers := odb.IterHeaders()

	for h := range headers {
		header = h
	}
	return
}

func (odb *ObjDatabase) AllHeaders(dome ...func(header *ObjHeader)) (hs []*ObjHeader) {
	hee := odb.IterHeaders()

	for h := range hee {
		if dome != nil {
			dome[0](h)
		}
		hs = append(hs, h)
	}
	return
}
func (odb *ObjDatabase) Count() int {
	n := 0
	for range odb.IterHeaders() {
		n++
	}
	return n
}

type ObjHeaders []*ObjHeader

func (headers ObjHeaders) With(cal func(head *ObjHeader)) ObjHeaders {
	// count := len(headers)
	for _, h := range headers {
		cal(h)
	}
	return headers
}

func (odb *ObjDatabase) AllBody() (hs []*ObjBody) {
	hee := odb.IterBody()

	for h := range hee {
		hs = append(hs, h)
	}
	return
}

func (odb *ObjDatabase) writeTo(headers []*ObjHeader, bodys []*ObjBody) error {

	odb.Close()
	if odb.fb != nil {
		odb.Close()
		defer func() {
			odb.fb, _ = os.Open(odb.FileName)
		}()
	}
	bak, err := os.Create(odb.FileName + ".bak")
	if err != nil {
		return err
	}
	if err != nil {
		log.Fatal("Bakup err:", err)
	}
	defer bak.Close()
	crcs := [][]byte{}
	for _, h := range headers {
		bak.Write(h.Bytes())
		crcs = append(crcs, h.Crc[:])
	}

	for _, b := range bodys {
		// fmt.Println(b)
		// bak.Write(crcs[i])
		bak.Write(b.Bytes())
		// bak.Write(crcs)
	}
	err = os.Remove(odb.FileName)
	if err != nil {
		fmt.Println("Err:", err)
		return err
	}

	bak.Close()
	err = os.Rename(odb.FileName+".bak", odb.FileName)
	if err != nil {

		fmt.Println("Err:", err)
		return err
	}
	return nil
}

func NewObjClient(fileName string) *ObjDatabase {
	c := new(ObjDatabase)
	c.FileName = fileName
	return c
}

func (client *ObjDatabase) UpdateBlock(info string, data []byte, keys ...string) *ObjDatabase {
	head, _ := client.QueryBlock(info)
	if head == nil {
		head := NewObj()
		head.SetInfo(info)
		head.Write(client, data, keys...)
	} else {
		head.Update(client, info, data, keys...)

	}
	// head.SetInfo(info)
	return client
}

func (client *ObjDatabase) CreateBlock(info string, data []byte, keys ...string) *ObjDatabase {
	head := NewObj()
	head.SetInfo(info)
	head.Write(client, data, keys...)
	return client
}

func (client *ObjDatabase) Exists() bool {
	if _, err := os.Stat(client.FileName); err == nil {
		return true
	} else {
		fmt.Println(Green(client.FileName))
		return false
	}
}
