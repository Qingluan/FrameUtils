package utils

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
)

type Line []string
type BDict map[string]string
type Dict map[string]interface{}

func (line Line) Filter(each func(int, string) bool) (int, bool) {
	for i, v := range line {
		if each(i, v) {
			return i, true
		}
	}
	return -1, false
}

func (line *Line) Push(is ...interface{}) {
	for _, i := range is {
		*line = append(*line, fmt.Sprintf("%v", i))
	}
}

func (line Line) FromKey(key Line) (d Dict) {
	d = make(Dict)
	klen := len(key)
	line.Filter(func(i int, v string) bool {
		if i < klen {
			d[key[i]] = v
		}
		return false
	})
	return
}

func (line Line) Or(other Line) (newl Line) {
	newl = append(newl, line...)
	for _, v := range other {
		if i, _ := newl.Filter(func(i int, s string) bool {
			if s != v {
				return false
			}
			return true
		}); i < 0 {
			newl = append(newl, v)
		}
	}
	return
}

func (line Line) And(other Line) (newl Line) {
	for _, v := range other {
		if _, ok := line.Filter(func(i int, s string) bool {
			if s == v {
				return true
			}
			return false
		}); ok {
			newl = append(newl, v)
		}
	}
	return
}

func (line Line) Xor(other Line) (newl Line) {
	// newl = append(newl, line...)
	for _, v := range other {
		if i, _ := line.Filter(func(i int, s string) bool {
			if s != v {
				return false
			}
			return true
		}); i < 0 {
			newl = append(newl, v)
		}
	}
	return
}

func (line Line) Diff(other Line) (newl Line) {
	// newl = append(newl, line...)
	for _, v := range line {
		if i, _ := other.Filter(func(i int, s string) bool {
			if s != v {
				return false
			}
			return true
		}); i < 0 {
			newl = append(newl, v)
		}
	}
	return
}

func (line Line) Contain(word string) bool {
	for _, v := range line {
		if v == word {
			return true
		}
	}
	return false
}

func (line Line) Contains(oline Line) bool {
	for _, w := range line {
		if oline.Contain(w) {
			return true
		}
	}
	return false
}

func (line Line) Collect(filter func(i int, s string) bool) (newl Line) {
	for i, v := range line {
		if filter(i, v) {
			newl = append(newl, v)
		}
	}
	return
}

func (dict Dict) Keys() (l Line) {
	for k := range dict {
		l = append(l, k)
	}
	sort.Slice(l, func(i int, j int) bool {
		if strings.Compare(l[i], l[j]) < 0 {
			return true
		}
		return false
	})
	return l
}

func (dict Dict) String() string {
	b, err := json.Marshal(dict)
	if err != nil {
		log.Fatal("can not be as Dict from :", dict)
	}
	return string(b) + "\n"
}

func (dict Dict) Format() string {
	b, err := json.MarshalIndent(dict, "", "  ")
	if err != nil {
		log.Fatal("can not be as Dict from :", dict)
	}
	return string(b) + "\n"
}

func (dict Dict) And(other Dict) float32 {
	sc := len(other)
	scb := float32(sc)
	for k, v := range other {
		if v2, ok := dict[k]; !ok || v2 != v {
			sc--
		}
	}
	return float32(sc) / scb
}

func ContainInt(ss []int, s int) bool {
	for _, v := range ss {
		if v == s {
			return true
		}
	}
	return false
}

func AsDict(s string) (d Dict) {
	d = make(Dict)
	json.Unmarshal([]byte(s), &d)
	return
}

func SaveDict(ds []Dict, name string) error {
	data, err := json.Marshal(ds)
	if err != nil {
		return err
	}
	f, err := os.OpenFile(name, os.O_CREATE|os.O_RDWR, os.ModePerm)
	if err != nil {
		return err
	}
	defer f.Close()
	f.Write(data)
	return nil
}

func SplitByIgnoreQuote(raw string, by string, quotes ...string) (out []string) {
	quoted := false
	key := ""
	c := ' '
	if quotes != nil {
		out = strings.FieldsFunc(raw, func(r rune) (ifsplit bool) {
			if key != "" && strings.HasPrefix(by, key+string(r)) {
				key += string(r)
			} else if by[0] == byte(r) {
				key += string(r)
			} else {
				key = ""
			}
			if !quoted && r == rune(quotes[0][0]) {
				c = r
				quoted = !quoted
			} else if quoted && r == rune(quotes[0][1]) {
				quoted = !quoted
			}

			return !quoted && key == by
		})
	} else {
		out = strings.FieldsFunc(raw, func(r rune) (ifsplit bool) {
			if key != "" && strings.HasPrefix(by, key+string(r)) {
				key += string(r)
			} else if by[0] == byte(r) {
				key += string(r)
			} else {
				key = ""
			}
			if !quoted && (r == '"' || r == '\'') {
				c = r
				quoted = !quoted
			} else if quoted && r == c {
				quoted = !quoted
			}

			return !quoted && key == by
		})
	}

	return
}
