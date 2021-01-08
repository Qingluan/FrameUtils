package LocalDB

import (
	"github.com/Qingluan/FrameUtils/engine"
	"os"
)

type Bias [2]int64
type Index struct {
	Name    string `json:"key"`
	Include []Bias `json:"include"`
}

// type Ranger struct {
// 	Include []int  `json:"include"`
// 	Range   string `json:"range"`
// }

type DBHeader struct {
	Indexes        map[string]*Index `json:"indexes"`
	ItemsBias      []Bias            `json: "items_bias"`
	tmpChangedBias []Bias
}

type DBCursor struct {
	Now       int
	change    *ChangePoint
	cache     []engine.Dict
	indexKeys string
}

type DBHandler struct {
	Header *DBHeader
	DBPath string
	fb     *os.File
	Cursor *DBCursor
	datas  []engine.Dict
}
type DBHeaderErr error
type DBHeaderLoadErr error
type DBNotFoundErr error
