package engine

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/thedatashed/xlsxreader"
)

func OpenObj(file string) (Obj, error) {
	if strings.HasSuffix(file, ".xlsx") {
		xl, err := xlsxreader.OpenFile(file)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		return &BaseObj{&Xlsx{xl}}, nil
	} else if strings.HasSuffix(file, ".csv") {
		buf, err := ioutil.ReadFile(file)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		return &BaseObj{
			&Csv{
				raw: string(buf),
			},
		}, nil
	} else if strings.HasSuffix(file, ".txt") {
		buf, err := ioutil.ReadFile(file)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		return &BaseObj{
			&Txt{
				raw: string(buf),
			},
		}, nil
	} else if strings.HasSuffix(file, ".json") {
		buf, err := ioutil.ReadFile(file)
		v := []Dict{}
		err = json.Unmarshal(buf, &v)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		if len(v) > 0 {
			return &BaseObj{
				&JsonObj{
					Header: v[0].Keys(),
					Datas:  v,
				},
			}, nil
		} else {
			return nil, fmt.Errorf("%s is empty", file)
		}

	} else if strings.HasSuffix(file, ".sql") {
		buf, err := ioutil.ReadFile(file)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		return &BaseObj{
			&SqlTxt{
				raw: string(buf),
			},
		}, nil
	}
	return nil, nil
}

func (self *BaseObj) SaveJson(fileName string) {
	fp, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE, os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}
	defer fp.Close()
	ds := self.AsJson()
	data, err := json.Marshal(ds)
	if err != nil {
		log.Fatal(err)
	}
	_, err = fp.Write(data)
	if err != nil {
		log.Fatal(err)
	}
}
