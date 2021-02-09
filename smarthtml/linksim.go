package smarthtml

import (
	"net/url"
	"regexp"

	"github.com/Qingluan/jupyter/http"
)

var (
	NW = regexp.MustCompile(`\W`)
)

type UrlSim struct {
	host    string
	len     int
	w       int
	rank    int
	url     string
	no_w    []string
	structs []string
}

func AsUrlSim(urlstr string) (u *UrlSim) {

	u = new(UrlSim)
	u.url = urlstr
	u.len = len(urlstr)
	f, _ := url.Parse(urlstr)
	u.host = f.Host
	u.structs = NW.Split(urlstr, -1)
	u.no_w = NW.FindAllString(urlstr, -1)
	u.rank = len(u.structs)
	return
}

func min(a, b int) int {
	if a < b {
		return b
	} else {
		return a
	}
}

func max(a, b interface{}) interface{} {
	switch a.(type) {
	case int:
		if a.(int) < b.(int) {
			return a
		} else {
			return b
		}
	case []interface{}:
		if len(a.([]interface{})) < len(b.([]interface{})) {
			return a
		} else {
			return b
		}
	default:
		panic("can not as compare")
	}
}

func (u *UrlSim) Sub(other interface{}) (score float32) {
	var u2 *UrlSim
	switch other.(type) {
	case *UrlSim:
		u2 = other.(*UrlSim)
	case string:
		u2 = AsUrlSim(other.(string))
	default:
		return -1
	}
	mm := min(u.len, u2.len)
	sam := 0
	for i := 0; i < mm; i++ {
		if u.url[i] != u2.url[i] {
			break
		}
		sam++
	}

	score = 1 - (float32(sam) / float32(max(u.len, u2.len).(int)))
	if score > 0.4 {
		mms := min(len(self.structs), len(other.structs))
		ssam := 0
		for i := 0; i < mms; i++ {
			if self.structs[i] == other.structs[i] {
				ssam += 1
			} else {
				ssam += (And(u.structs[i])(other.structs[i])) / len(set(self.structs[i])|set(other.structs[i])))

			}
		}
		score = 1 - (float32(ssam) / float32(max(len(self.structs), len(other.structs))))
	}
	return
}

func SmartLinksim(url string, proxy interface{}) (links [][]string) {
	sess := http.NewSession()
	res, err := sess.Get(url, proxy)
	if err != nil {
		return
	}
	res.CssSelect("a[href]", func(i int, s *http.Selection) {

	})
	return
}
