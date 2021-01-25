package engine

type JsonObj struct {
	Header  Line
	KeyMode int
	Datas   []Dict
}

func (self *JsonObj) GetHead(k string) Line {
	return nil
}

func (self *JsonObj) Iter() <-chan Line {
	ch := make(chan Line)
	go func() {
		c := 0
		for _, d := range self.Datas {
			l := Line{}
			for _, v := range self.Header {

				if vv, ok := d[v]; ok {
					l = append(l, vv.(string))
				} else {
					l = append(l, "")
				}
			}
			ch <- l
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
func (self *JsonObj) header(KeySearchLengths ...int) (l Line) {
	if len(self.Header) != 0 {
		return self.Header
	} else {
		tmp := make(Dict)
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
