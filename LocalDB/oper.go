package LocalDB

import (
	"fmt"
	"github.com/Qingluan/FrameUtils/engine"
	"io"
	"log"
	"os"
	"path/filepath"
)

func (db *DBHandler) Query(filter engine.Dict, batchLeng ...int) chan []engine.Dict {
	fs := make(chan []engine.Dict)
	batchLen := 1000
	if batchLeng != nil {
		batchLen = batchLeng[0]
	}
	go func() {
		var index *Index

		if len(filter) > 0 {
			init := false

			for k := range filter {
				if v, ok := db.Header.Indexes[k]; ok {
					if !init {
						index = v
					} else {
						if newbias := index.And(v); len(newbias) > 0 {
							index.Include = newbias
						} else {
							return
						}
					}
				}
			}
			if index == nil {
				return
			}
			onebatch := []engine.Dict{}
			for _, b := range index.Include {
				if d, err := db.getByBias(b); err != nil {
					log.Fatal("broken db !!", err, b)
				} else if d.And(filter) == 1 {
					// fmt.Println("found!")
					onebatch = append(onebatch, d)
					if len(onebatch) >= batchLen {
						fs <- onebatch
						onebatch = []engine.Dict{}

					}
				}
			}

			if len(onebatch) > 0 {
				fs <- onebatch
			}
		} else {
			onebatch := []engine.Dict{}

			for _, b := range db.Header.ItemsBias {
				if d, err := db.getByBias(b); err != nil {
					log.Fatal("broken db !!", err)
				} else {
					onebatch = append(onebatch, d)
					if len(onebatch) >= batchLen {
						fs <- onebatch
						onebatch = []engine.Dict{}

					}
				}
			}
			if len(onebatch) > 0 {
				fs <- onebatch
			}
		}
		close(fs)
	}()
	return fs
}

// Filter batchLength default is 1.
func (db *DBHandler) Filter(filterKey engine.Dict, batchLengths ...int) (onebatch []engine.Dict) {

	batchLength := 1
	if batchLengths != nil {
		batchLength = batchLengths[0]
	}
	if len(db.datas) > 0 {
		for _, d := range db.datas {
			if d.And(filterKey) == 1 {
				onebatch = append(onebatch, d)
				if len(onebatch) >= batchLength {
					// fs <- onebatch
					// onebatch = []engine.Dict{}
					return
				}
			}
		}

	} else {
		for _, b := range db.Header.ItemsBias {
			if d, err := db.getByBias(b); err != nil {
				log.Fatal("broken db !!", err, b)
			} else if d.And(filterKey) == 1 {
				// fmt.Println("found!")
				onebatch = append(onebatch, d)
				if len(onebatch) >= batchLength {
					// fs <- onebatch
					// onebatch = []engine.Dict{}
					return
				}
			}
		}
	}

	return
}

// Find .
func (db *DBHandler) Find(filterKey engine.Dict) (one engine.Dict) {

	if ds := db.Filter(filterKey); len(ds) > 0 {
		return ds[0]
	}
	return nil
}

func (db *DBHandler) Remove() (err error) {
	os.Remove(db.DBPath)
	os.Remove(db.DBPath + ".header")
	return
}

func (db *DBHandler) Delete(filter engine.Dict) *DBHandler {
	return db
}

func (db *DBHandler) Insert(newdict []engine.Dict, keys ...string) *DBHandler {
	if keys != nil {
		db.Cursor.indexKeys = keys[0]
	}
	for _, d := range newdict {
		db.Cursor.cache = append(db.Cursor.cache, d)
		if len(db.Cursor.cache) > 10000 {
			db.saveToStorage(db.Cursor.cache)
			db.Cursor.cache = []engine.Dict{}
		}
	}

	return db
}

func (db *DBHandler) Update(filter engine.Dict, newpart engine.Dict) *DBHandler {

	type TmpBd struct {
		bias Bias
		no   int
		d    engine.Dict
	}
	ts := []TmpBd{}
	for k := range filter {
		if v, ok := db.Header.Indexes[k]; ok {
			for no, b := range v.Include {
				// if b[0] <
				if d, err := db.getByBias(b); err != nil {
					log.Fatal("broken update db !!", err)
				} else if d.And(filter) == 1 {
					ts = append(ts, TmpBd{b, no, d})
				}
			}

		}
	}

	for _, bd := range ts {
		db.changeByBias(bd.bias, bd.d, newpart)
	}
	return db
}

/*Commit commit update or save
 */
func (db *DBHandler) Commit() (err error) {
	changed := false
	if len(db.Cursor.cache) > 0 {
		changed = true
		db.saveToStorage(db.Cursor.cache)

	}

	db.Close()
	if db.Cursor.change != nil {
		changed = true
		if tmpN, err := db.WithTmp(func(fs *os.File) error {
			return db.changeWithFile(fs)
		}); err != nil {
			return err
		} else {
			// fmt.Println("write -->", tmpN)
			os.Rename(tmpN, db.DBPath)
		}
	}

	defer func() {
		if changed {
			db.Header = ReBuildIndexes(db.DBPath)
			db.Header.Save(db.DBPath + ".header")
		}
		db.fb, err = os.OpenFile(db.DBPath, os.O_CREATE|os.O_RDWR, os.ModePerm)
	}()
	return
}

func (db *DBHandler) WithTmp(handle func(fb *os.File) error) (tmpName string, err error) {
	tmpdir := filepath.Join(os.TempDir(), filepath.Base(db.DBPath))
	fs, err := os.OpenFile(tmpdir, os.O_CREATE|os.O_RDWR|os.O_TRUNC, os.ModePerm)
	if err != nil {
		return "", err
	}
	defer fs.Close()
	return tmpdir, handle(fs)
}

func (db *DBHandler) changeWithFile(fs *os.File) error {
	oldFb, err := os.OpenFile(db.DBPath, os.O_RDONLY, os.ModePerm)
	oldStat, _ := oldFb.Stat()
	oldlength := oldStat.Size()
	if err != nil {
		return err
	}
	defer oldFb.Close()
	start := int64(0)
	nowChange := db.Cursor.change.First()
	for nowChange != nil {
		fmt.Println("Change -->", start, nowChange.Newbias, string(nowChange.Buf))
		// if len(nowChange.ChangeKeys) != 0 && nowChange.Newbias != nil {
		// 	// for k, no := range nowChange.ChangeKeys {
		// 	// 	db.Header.Indexes[k][no] = nowChange.Newbias
		// 	// }
		// }
		start = nowChange.propagate(fs, oldFb, start)
		nowChange = nowChange.Next
	}
	lastChange := db.Cursor.change.Last()
	if lastChange.Oldbias[1] < oldlength {
		oldFb.Seek(lastChange.Oldbias[1], io.SeekStart)
		io.Copy(fs, oldFb)
	}
	return nil
}

func (self *DBHandler) UseCache() {
	for _, i := range self.Header.ItemsBias {
		if d, err := self.getByBias(i); err != nil {

		} else {
			self.datas = append(self.datas, d)
		}
	}
}
