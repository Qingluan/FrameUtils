package LocalDB

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/Qingluan/FrameUtils/utils"
)

var (
	FileLock = sync.Mutex{}
)

func NewDBHandler(name string) (db *DBHandler, err error) {
	db = new(DBHandler)
	db.fb, err = os.OpenFile(name, os.O_CREATE|os.O_RDWR, os.ModePerm)
	db.DBPath = name
	db.Header, err = LoadHeader(name + ".header")
	db.Cursor = new(DBCursor)
	if len(db.Header.ItemsBias) == 0 {
		db.Header = ReBuildIndexes(name)
	}
	return
}

func (db *DBHandler) saveToStorage(items []utils.Dict) *DBHandler {
	FileLock.Lock()
	defer FileLock.Unlock()
	mainkey := db.Cursor.indexKeys
	defer func() {
		db.Cursor.cache = []utils.Dict{}
	}()

	for _, v := range items {
		buf := v.String()
		newbias := db.Header.createNewByBuf(buf)
		// fmt.Println(len(buf), newbias)
		if mainkey != "" {
			if _, ok := v[mainkey]; ok {
				if index, ok := db.Header.Indexes[mainkey]; ok {

					// fmt.Println("insert index:", mainkey)
					index.Include = append(index.Include, newbias)
					db.Header.Indexes[mainkey] = index
				} else {
					// fmt.Println("create index:", mainkey)
					db.Header.Indexes[mainkey] = &Index{
						Include: []Bias{newbias},
						Name:    mainkey,
					}
				}
			}
		}
		db.Header.ItemsBias = append(db.Header.ItemsBias, newbias)
		_, err := db.fb.WriteAt([]byte(buf), newbias[0])
		if err != nil {
			panic(err)
		}
		// for k := range v {
		//
		// }
	}
	return db
}

func (db *DBHandler) deleteByBias(bias Bias) {
	db.Cursor.change.Delete(bias)
}

func (db *DBHandler) Close() error {
	if err := db.Header.Save(db.DBPath + ".header"); err != nil {
		return err
	}
	return db.fb.Close()
}

func (db *DBHandler) getByBias(bias Bias) (d utils.Dict, err error) {
	FileLock.Lock()
	defer FileLock.Unlock()
	buf := make([]byte, bias[1]-bias[0])
	var n int
	n, err = db.fb.ReadAt(buf, bias[0])
	if err != nil {
		return nil, DBNotFoundErr(fmt.Errorf("not found by bias %s", err.Error()))
	}
	if n < int(bias[1]-bias[0]) {
		return nil, DBNotFoundErr(fmt.Errorf("read item by bias error !: %s", bias))
	}
	d = make(utils.Dict)
	err = json.Unmarshal(buf, &d)
	if err != nil {
		log.Fatal("err json:", string(buf), bias)
	}
	return
}

func (db *DBHandler) changeBiasbyBias(b Bias, offset int, changeAllAfterHit bool) {
	FileLock.Lock()
	defer FileLock.Unlock()
	start := -1
	L := len(db.Header.ItemsBias)
	for i, k := range db.Header.ItemsBias {
		if b[0] == k[0] {
			start = i
			break
		}
	}
	if start > 0 {
		init := true
		for ; start < L; start++ {
			if init {
				init = false
				db.Header.ItemsBias[start][1] += int64(offset)
				if !changeAllAfterHit {
					break
				}
			} else {
				db.Header.ItemsBias[start][0] += int64(offset)
				db.Header.ItemsBias[start][1] += int64(offset)
			}
		}
	}
}

// func (db *DBHandler) WithTruncateMode(handle func(fb *os.File)) {
// 	db.fb.Close()

// 	var err error
// 	// oldSeek := db.fb.Seek
// 	db.fb, err = os.OpenFile(db.DBPath, os.O_WRONLY|os.O_TRUNC, os.ModePerm)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	handle(db.fb)
// 	defer func() {
// 		db.fb.Close()
// 		db.fb, err = os.OpenFile(db.DBPath, os.O_CREATE|os.O_RDWR, os.ModePerm)
// 	}()

// }

func (db *DBHandler) changeByBias(bias Bias, old, newpart utils.Dict) {
	for k, v := range newpart {
		old[k] = v
	}
	buf := old.String()
	if db.Cursor.change == nil {
		db.Cursor.change = &ChangePoint{
			Oldbias: bias,
			Newbias: Bias{bias[0], int64(len(buf)) + bias[0]},
			Buf:     []byte(buf),
			// ChangeKeys:   changeKeys,
			// ItemPosition: itemNo,
		}
	} else {
		db.Cursor.change.Change(bias, buf)
	}
	// if offset > 0 {
	// 	db.changeBiasbyBias(bias, offset, true)
	// 	db.WithTruncateMode(func(fb *os.File) {
	// 		db.fb.WriteAt([]byte(buf[:bias[1]-bias[0]]), bias[0])
	// 	})
	// 	db.fb.WriteAt([]byte(buf[bias[1]-bias[0]:]), bias[1])
	// 	// need to implement
	// 	// remove key then add key
	// } else if offset < 0 {
	// 	db.changeBiasbyBias(bias, len(buf)-int(bias[1]-bias[0]), false)
	// 	db.WithTruncateMode(func(fb *os.File) {
	// 		db.fb.WriteAt([]byte(buf), bias[0])
	// 	})
	// } else {
	// 	db.WithTruncateMode(func(fb *os.File) {
	// 		db.fb.WriteAt([]byte(buf), bias[0])
	// 	})
	// }
}
