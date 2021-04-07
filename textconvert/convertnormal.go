package textconvert

import (
	"bufio"
	"io"
	"io/ioutil"
	"os"
)

func NormalToEs(path string) (es ElasticFileDocs, err error) {

	fp, err := os.Open(path)
	if err != nil {
		return es, err
	}
	defer fp.Close()
	info, _ := fp.Stat()
	if info.Size() > 1024*1024*30 {
		reader := bufio.NewReader(fp)
		msg := ""
		for {
			l, _, err := reader.ReadLine()
			if err == io.EOF || err != nil {
				break
			}
			msg += string(l) + "\n"
		}

		es.SomeStr, err = TOUTF8(msg)
		es.Path = path

	} else {
		buf, err := ioutil.ReadAll(fp)
		if err != nil {
			return es, err
		}
		es.SomeStr, err = TOUTF8(string(buf))
	}
	// if buf, err := ioutil.ReadFile(path); err != nil {
	// 	return es, err
	// } else {
	// 	es.SomeStr, err = TOUTF8(string(buf))
	// 	es.Path = path
	// }
	return
}
