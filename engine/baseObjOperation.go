package engine

/*Join usage
Join two Frame by keys, if not set keys use first key

*/
func (self *BaseObj) Join(other Obj, opt int, keys ...string) (newObj *BaseObj) {
	headerL := self.Header()
	headerR := other.Header()
	var mergeHeader Line
	var diffHeader Line
	lft := false
	useLast := false
	if opt&OPT_JOINT_NO_SAME > 0 {
		useLast = true
	}
	if opt&OPT_RIGHTJOIN > 0 {

		mergeHeader = headerR.Or(headerL)
		diffHeader = headerL.Diff(headerR)
	} else {
		lft = true
		mergeHeader = headerL.Or(headerR)
		diffHeader = headerR.Diff(headerL)
	}
	// sameHeader := headerL.And(headerR)
	// fmt.Println("d:", diffHeader)
	jsonObj := &JsonObj{
		// Header: mergeHeader,
	}

	if keys == nil {
		keys = []string{mergeHeader[0]}
	}
	ltmp := []Dict{}
	rtmp := []Dict{}

	choosed := []int{}
	haveUsed := map[int]bool{}
	if lft {

		for liner := range other.Iter() {
			rd := liner.FromKey(headerR)
			var matchd Dict
			choosedNo := 0
			for linel := range self.Iter() {
				if _, ok := haveUsed[choosedNo]; ok && useLast {
					continue
				}
				ld := linel.FromKey(headerL)
				found := false
				for _, key := range keys {
					if v, ok := rd[key]; ok {
						if v2, ok2 := ld[key]; ok2 && v == v2 {
							found = true
							break
						}
					}
				}
				if found {
					matchd = ld
					choosed = append(choosed, choosedNo)
					haveUsed[choosedNo] = true
					break
				}
				choosedNo++
			}
			if matchd != nil {

				// fmt.Println("match:", matchd)
				// fmt.Println("match r:", rd)
				for _, k := range diffHeader {
					if v, ok := rd[k]; ok {
						matchd[k] = v
					}
				}

				// fmt.Println("match:", matchd)
				// fmt.Println()
				ltmp = append(ltmp, matchd)
			} else {
				// for k := range headerL{
				// 	rd[k] = ld[]
				// }
				// fmt.Println("no match", rd)

				rtmp = append(rtmp, rd)
			}
		}
		c := 0
		for l := range self.Iter() {

			if !ContainInt(choosed, c) {
				kkk := l.FromKey(headerL)
				// fmt.Println("no match l:", kkk)
				ltmp = append(ltmp, kkk)
			}
			c++
		}

	} else {
		for linel := range self.Iter() {
			ld := linel.FromKey(headerL)
			var matchd Dict
			choosedNo := 0
			for liner := range other.Iter() {
				rd := liner.FromKey(headerR)
				found := false
				for _, key := range keys {
					if v, ok := rd[key]; ok {
						if v2, ok2 := ld[key]; ok2 && v == v2 {
							found = true
							break
						}
					}
				}
				if found {
					matchd = rd
					choosed = append(choosed, choosedNo)
					haveUsed[choosedNo] = true
					break
				}
				choosedNo++
			}
			if matchd != nil {
				for _, k := range diffHeader {
					if v, ok := ld[k]; ok {
						matchd[k] = v
					}
				}
				rtmp = append(rtmp, matchd)
			} else {
				ltmp = append(ltmp, ld)
			}
		}

		c := 0
		for l := range other.Iter() {

			if !ContainInt(choosed, c) {
				rtmp = append(rtmp, l.FromKey(headerR))
			}
			c++
		}

	}
	jsonObj.Datas = append(ltmp, rtmp...)
	// fmt.Println("All:", jsonObj.Datas)
	newObj = &BaseObj{
		jsonObj,
	}
	return
}

func (self *BaseObj) Match(line Line, keys ...string) bool {
	for linel := range self.Iter() {
		if keys != nil {

			d := linel.FromKey(self.Header())
			checklines := Line{}
			for _, k := range keys {
				if vvv, ok := d[k]; ok {
					checklines = append(checklines, vvv.(string))
				}
			}
			if line.Contains(checklines) {
				return true
			}
		} else {
			if linel.Contains(line) {
				return true
			}
		}
	}
	return false

}
