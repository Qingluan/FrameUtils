package task

import (
	"encoding/json"
	"fmt"
	"log"
	"path/filepath"
)

type ErrObj struct {
	Root string
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
func (erro ErrObj) Path() string {
	return filepath.Join(erro.Root, erro.ID()+".log")
}
func (erro ErrObj) Error() error {
	return erro.Err
}
func (erro ErrObj) ID() string {
	return erro.tp + "-" + NewID(erro.raw)

}
func (erro ErrObj) Args() []string {
	fmt.Println("Err try: args:", erro.raw)
	return []string{erro.tp, erro.raw}
}

func (erro ErrObj) LogToLocal() {
	// fmt.Println("\n\nTo ENd:", path)
	_to_end(erro.Path(), []byte(erro.Err.Error()))
}
