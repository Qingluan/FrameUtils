package task

import (
	"crypto/md5"
	"encoding/json"
	"log"
	"strings"
)

type ErrObj struct {
	Err  error
	args []string
}

func (erro ErrObj) String() string {
	buf, err := json.Marshal(map[string]string{
		"Tp":   "Err",
		"Data": erro.Err.Error(),
		"Args": strings.Join(erro.Args(), "|"),
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
	return erro.args
}
