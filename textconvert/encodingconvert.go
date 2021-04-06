package textconvert

import (
	"github.com/timakin/gonvert"
)

func TOUTF8(raw string) (utf8Str string, err error) {
	converter := gonvert.New(raw, gonvert.UTF8)
	utf8Str, err = converter.Convert()
	// This will print out the utf-8 encoded string: "月日は百代の過客にして、行かふ年も又旅人也。"
	return
}
