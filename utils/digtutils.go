package utils

import (
	"encoding/json"
	"fmt"
	"strings"
	// "github.com/Qingluan/FrameUtils/utils"
)

func (bdict BDict) FromCmd(cmd string) BDict {
	// var fs []string
	fs := SplitByIgnoreQuote(cmd, ",")
	for _, f := range fs {
		fmt.Println(f)
		if strings.Contains(f, "=") {
			fs2 := strings.SplitN(f, "=", 2)
			name := strings.TrimSpace(fs2[0])
			value := strings.TrimSpace(fs2[1])
			value = strings.TrimLeft(value, "\"")
			value = strings.TrimRight(value, "\"")
			bdict[name] = value
		}
	}
	return bdict
}

func (bdict BDict) String() string {
	b, _ := json.Marshal(bdict)
	return string(b)
}
