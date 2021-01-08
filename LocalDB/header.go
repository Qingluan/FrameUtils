package LocalDB

import (
	"bufio"
	"bytes"
	"encoding/json"
	"github.com/Qingluan/FrameUtils/engine"
	"io"
	"io/ioutil"
	"os"

	"github.com/cheggaaa/pb/v3"
)

func LoadHeader(name string) (header *DBHeader, err error) {
	buf, err := ioutil.ReadFile(name)
	header = new(DBHeader)
	if err != nil {
		header.Indexes = make(map[string]*Index)
		return header, nil
	}
	err = json.Unmarshal(buf, header)
	return
}

func (header *DBHeader) LastBias() Bias {
	if len(header.ItemsBias) == 0 {
		return Bias{0, 0}
	} else {
		return header.ItemsBias[len(header.ItemsBias)-1]
	}
}

func (header *DBHeader) createNewByBuf(buf string) Bias {
	bias := header.LastBias()
	if bias[1] != 0 {
		return Bias{bias[1], bias[1] + int64(len(buf))}
	}
	return Bias{0, int64(len(buf))}
}

func (header *DBHeader) Save(name string) error {
	// buf, err := json.MarshalIndent(header, "", "    ")

	buf, err := json.Marshal(header)
	// for k, v := range header.Indexes {
	// 	fmt.Println(k, v)
	// }
	// fmt.Println(string(buf))
	if err != nil {
		return err
	}
	return ioutil.WriteFile(name, buf, os.ModePerm)
}

func (index *Index) And(other *Index) (r []Bias) {
	for _, bias := range index.Include {
		for _, ob := range other.Include {
			if bias[0] == ob[0] {
				r = append(r, ob)
			}
		}
	}
	return
}

func (bias Bias) Length() int64 {
	return bias[1] - bias[0]
}

func ReBuildIndexes(dbpath string, registedKeys ...string) (head *DBHeader) {
	// fmt.Println("--- rebuild ----")
	head = new(DBHeader)
	fb, err := os.Open(dbpath)
	stat, _ := fb.Stat()
	if err != nil {
		panic(err)
	}
	defer fb.Close()
	var bar *pb.ProgressBar
	if stat.Size() > 100000 {
		bar = pb.Start64(stat.Size())
	}
	scanner := bufio.NewScanner(fb)
	scanner.Split(func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		if atEOF && len(data) == 0 {
			return 0, nil, nil
		}
		if atEOF {
			return len(data), data, io.EOF
		}

		if e := bytes.Index(data, []byte("}\n")); e > 0 {

			if es := bytes.Index(data[:e+2], []byte{'{'}); es >= 0 {
				// fmt.Println("F::", string(data[es:e+2]))

				return e + 2, data[es : e+2], nil
			}
			return e + 2, nil, nil
		}

		return 0, nil, nil
	})
	head.Indexes = make(map[string]*Index)

	start := int64(0)

	for scanner.Scan() {

		oned := scanner.Text()
		length := int64(len(oned))
		end := start + length
		bias := Bias{start, end}
		d := engine.AsDict(oned)
		// fmt.Println("len:", length, oned)
		if bar != nil {
			bar.Add64(length)
		}

		for k := range d {
			for _, k2 := range registedKeys {
				if k2 == k {
					if index, ok := head.Indexes[k]; ok {
						index.Add(bias)
					} else {
						head.Indexes[k] = &Index{
							Include: []Bias{bias},
							Name:    k,
						}
					}
				}

			}

		}
		start = end
		head.ItemsBias = append(head.ItemsBias, bias)
		// fmt.Println("found:", oned)
	}
	if bar != nil {
		bar.Finish()
	}
	return
}
