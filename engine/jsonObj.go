package engine

import "github.com/Qingluan/FrameUtils/utils"

type JsonObj struct {
	Header    utils.Line
	KeyMode   int
	Datas     []utils.Dict
	tableName string
}

func (self *JsonObj) GetHead(k string) utils.Line {
	return utils.Line{self.tableName}
}

func (self *JsonObj) Iter(filterobj ...string) <-chan utils.Line {
	ch := make(chan utils.Line)
	go func() {
		c := 0
		for _, d := range self.Datas {
			l := utils.Line{}
			for _, v := range self.Header {

				if vv, ok := d[v]; ok {
					l = append(l, vv.(string))
				} else {
					l = append(l, "")
				}
			}
			ch <- append(utils.Line{self.tableName}, l...)
			c++
		}
		close(ch)
	}()
	return ch
}
func (s *JsonObj) Tp() string {
	return "json"
}

func (self *JsonObj) Close() error {
	return nil
}
func (self *JsonObj) header(KeySearchLengths ...int) (l utils.Line) {
	if len(self.Header) != 0 {
		return self.Header
	} else {
		tmp := make(utils.Dict)
		KeySearchLength := 0
		if KeySearchLengths != nil {
			KeySearchLength = KeySearchLengths[0]
		}
		if KeySearchLength == 0 {
			for _, v := range self.Datas {
				for _, kk := range v.Keys() {
					tmp[kk] = 1
				}
			}
		} else {
			ll := KeySearchLength
			if ll > len(self.Datas) {
				ll = len(self.Datas)
			}
			for _, v := range self.Datas[:ll] {
				for _, kk := range v.Keys() {
					tmp[kk] = 1
				}
			}
		}

		self.Header = tmp.Keys()
		l = self.Header
	}
	return
}
