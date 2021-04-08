package task

import (
	"crypto/md5"
	"encoding/json"
	"log"
	"path/filepath"
	"strings"
)

type ErrObj struct {
	Err  error
	tp   string
	raw  string
	toGo string
}

func (erro ErrObj) ToGo() string {
	return erro.toGo
}

func (erro ErrObj) String() string {
	buf, err := json.Marshal(map[string]string{
		"Tp":   "Err",
		"Data": erro.Err.Error(),
		"Args": erro.raw,
	})
	if err != nil {
		log.Fatal(err)
	}
	return string(buf)
}
func (erro ErrObj) Error() error {
	return erro.Err
}
func (erro ErrObj) ID() string {
	md := md5.New()
	// md.Write()

	ks := strings.Join(erro.Args(), "|")
	raw := []byte(ks)
	b := md.Sum(raw)
	return string(b)
}
func (erro ErrObj) Args() []string {
	return []string{erro.tp, erro.raw}
}

func (erro ErrObj) LogToLocal(root string) {
	id := NewID(erro.raw)
	path := filepath.Join(root, "err-"+id) + ".log"
	_to_end(path, []byte(erro.Err.Error()))
}
