package engine

import (
	"encoding/json"
	"fmt"
	"strings"
)

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

/*
<table class="table">
  <thead class="thead-dark">
    <tr>
      <th scope="col">#</th>
      <th scope="col">First</th>
      <th scope="col">Last</th>
      <th scope="col">Handle</th>
    </tr>
  </thead>
  <tbody>
    <tr>
      <th scope="row">1</th>
      <td>Mark</td>
      <td>Otto</td>
      <td>@mdo</td>
    </tr>
    <tr>
      <th scope="row">2</th>
      <td>Jacob</td>
      <td>Thornton</td>
      <td>@fat</td>
    </tr>
    <tr>
      <th scope="row">3</th>
      <td>Larry</td>
      <td>the Bird</td>
      <td>@twitter</td>
    </tr>
  </tbody>
</table>
*/
func (self *BaseObj) ToHTML(tableID ...string) string {
	ID := "default-table"
	// usePage := false
	if tableID != nil {
		ID = tableID[0]

		// if len(tableID) > 1 && tableID[1] == "#page" {
		// 	usePage = true
		// }
	}
	headers := self.Header()
	pre := fmt.Sprintf(`<table  class="table" id="%s" ><thead class="thead-dark">`, ID)
	hs := []string{}
	hasHeader := false
	pre += "<tr>%s</tr></thead><tbody>"
	if len(headers) > 0 {
		hasHeader = true
		for _, i := range headers {
			hs = append(hs, fmt.Sprintf("<th scope=\"col\">%s</th>", i))
		}
		pre = fmt.Sprintf(pre, strings.Join(hs, "\n"))
	}

	row := -1
	for line := range self.Iter() {
		row++

		items := []string{}
		col := -1
		for i, li := range line[1:] {
			key := ""
			col++
			if i < len(headers) {
				key = headers[i]

				// fmt.Println("Key:", headers[i], key)
			}
			items = append(items, fmt.Sprintf("<td data-row=\"%d\" data-col=\"%d\" data=\"%s\" key=\"%s\" >%s</td>", row, col, li, key, li))
		}
		if hasHeader {
			hasHeader = false
			continue

		}
		pre += fmt.Sprintf("\n\t<tr data-row=\"%d\" onclick=\"click_tr(this);\" >%s</tr>", row, strings.Join(items, ""))
	}
	return pre + "\n    </tbody></table>"

}

func (self *BaseObj) Bytes() (body []byte, err error) {
	js := self.AsJson()
	body, err = json.Marshal(&js)
	return
}

func (self *BaseObj) Marshal() (body []byte, keys []string, err error) {
	body, err = self.Bytes()
	keys = self.header()
	return
}
