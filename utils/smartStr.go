package utils

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
)

/*Any Usage
${file:file_name}; open this file as data
${i:23++} ; {i=23; ;i++}{ ... use i ....}
${e:regex}; regex from read data, as data,
*/
type Any struct {
	raw string
	acs []*Ac
}

// Dict map[string]interface{}
// type Dict map[string]interface{}

var (
	smartMatched   = regexp.MustCompile(`\$\{(\S+?)\}`)
	cachedSessions = make(map[string]string)
)

/*Ac :
t: 0:format 1:file 2:int 3:regex
*/
type Ac struct {
	key    string
	t      int
	datai  int
	datas  string
	matchr *regexp.Regexp
}

// ArrayContains : if array str contains key , will return true
func ArrayContains(arrs []string, key string) bool {
	for _, v := range arrs {
		if v == key {
			return true
		}
	}
	return false
}

// Str panic err from NewSmartString
func Str(raw string) (any *Any) {
	any, err := NewSmartString(raw)
	if err != nil {
		panic(err)
	}
	return
}

func IsNumber(f string) (int, bool) {
	for _, c := range f {
		if rune('0') <= c && rune('9') >= c {
			continue
		} else {
			return -1, false
		}
	}
	ii, _ := strconv.Atoi(f)
	return ii, true
}

/*NewSmartString : generate a smart string obj
can use follow:
	${key}
	${i:1++} : from 1 ++ like for(i=1; ; i++)
	${e:hello\w} : regext matched
	${file:some_path}: will direction replace this from file!
*/
func NewSmartString(raw string) (any *Any, err error) {
	any = &Any{
		raw: raw,
	}
	ms := smartMatched.FindAllStringSubmatch(raw, -1)
	for _, ss := range ms {
		if strings.HasPrefix(ss[1], "file:") {
			fp, err := os.Open(strings.TrimLeft(ss[1], "file:"))
			if err != nil {
				return nil, err
			}
			defer fp.Close()
			buf, err := ioutil.ReadAll(fp)
			if err != nil {
				return nil, err
			}
			any.raw = strings.Replace(any.raw, ss[0], string(buf), 1)
		} else if strings.HasPrefix(ss[1], "i:") {
			mayInt := strings.TrimLeft(ss[1], "i:")
			i, err := strconv.Atoi(mayInt)
			if err != nil {
				return nil, err
			}
			any.acs = append(any.acs, &Ac{
				key:   ss[0],
				datai: i,
				t:     2,
			})

		} else if strings.HasPrefix(ss[1], "e:") {
			regexStr := strings.TrimLeft(ss[1], "e:")
			r, err := regexp.Compile(regexStr)
			if err != nil {
				return nil, err
			}
			any.acs = append(any.acs, &Ac{
				key:    ss[0],
				matchr: r,
				t:      3,
			})
		} else {
			any.acs = append(any.acs, &Ac{
				key: ss[0],
			})
		}
	}
	return
}

// Try to generate string , from resp buf / or directly
func (any *Any) Try(resp ...[]byte) string {
	out := any.raw
	for _, ac := range any.acs {
		if ac.t == 3 && ac.matchr != nil {
			out = strings.Replace(out, ac.key, ac.matchr.FindString(string(resp[0])), 1)
		} else if ac.t == 2 {
			out = strings.Replace(out, ac.key, fmt.Sprintf("%d", ac.datai), 1)
			ac.datai++
		} else if ac.t == 0 {
			out = strings.Replace(out, ac.key, ac.datas, 1)

		}
	}
	return out
}

func (any *Any) String() string {
	out := any.raw
	for _, ac := range any.acs {
		if ac.t == 2 {
			out = strings.Replace(out, ac.key, fmt.Sprintf("%d", ac.datai), 1)
			ac.datai++
		} else if ac.t == 0 {
			out = strings.Replace(out, ac.key, ac.datas, 1)
		}
	}

	return out
}
func (any *Any) UnEscape() string {
	out := any.String()

	if strings.Contains(out, "\\n") {
		out = strings.ReplaceAll(out, "\\n", "\n")
	}

	if strings.Contains(out, "\\r") {
		out = strings.ReplaceAll(out, "\\r", "\r")
	}

	if strings.Contains(out, "\\t") {
		out = strings.ReplaceAll(out, "\\t", "\t")
	}

	if strings.HasPrefix(out, "\"") && strings.HasSuffix(out, "\"") {
		out = out[1 : len(out)-1]
	}
	return out
}

// Format will tmp render smartstring's format key
func (any *Any) Format(d Dict) *Any {
	for k, v := range d {
		for _, ac := range any.acs {
			if ac.t == 0 && strings.HasPrefix(ac.key[2:], k+"}") {
				ac.datas = fmt.Sprint(v)
			}
		}
	}
	return any
}

/*SplitBy method
can use some build-in regex pattern
SMART_RE_IP
SMART_RE_URL
SMART_RE_JSON
*/
func (any *Any) SplitBy(reStr interface{}) (a []string) {
	var r *regexp.Regexp
	var err error
	switch reStr.(type) {
	case string:
		r, err = regexp.Compile(reStr.(string))
		if err != nil {
			a = strings.Split(any.String(), reStr.(string))
		} else {
			a = r.Split(any.String(), -1)
		}
	case *regexp.Regexp:
		r = reStr.(*regexp.Regexp)
		a = r.Split(any.String(), -1)
	}

	return
}

// func (any *Any) Check(keys map[string]interface{}) error{
// 	for _,ac := range any.acs{
// 		if ac.key
// 	}
// }

/*Collect method
can use some build-in regex pattern
SMART_RE_IP
SMART_RE_URL
SMART_RE_JSON
*/
func (any *Any) Collect(reStr interface{}) (a []string) {
	var r *regexp.Regexp
	var err error
	switch reStr.(type) {
	case string:
		r, err = regexp.Compile(reStr.(string))
	case *regexp.Regexp:
		r = reStr.(*regexp.Regexp)
	case func(string) []string:
		return reStr.(func(string) []string)(any.String())
	}
	if err != nil {
		log.Println(err)
		return
	}
	for _, i := range r.FindAllStringSubmatch(any.String(), -1) {
		a = append(a, i[0])
	}
	return
}

// Contains like strings.Contains
func (any *Any) Contains(reStr string) (b bool) {
	if strings.Contains(any.String(), reStr) {
		return true
	}
	return
}

//JSONFormat for format json string
func JSONFormat(v string) string {
	f := make(map[string]interface{})
	json.Unmarshal([]byte(v), &f)
	ff, _ := json.MarshalIndent(&f, "", "  ")
	return string(ff)
}
