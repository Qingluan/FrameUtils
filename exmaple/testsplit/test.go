package main

import (
	"fmt"
	"os"
	"strings"
)

func splitBy(raw string, by string) (out []string) {
	quoted := false
	key := ""
	c := ' '
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
		}
		if r == c {
			quoted = !quoted
		}

		return !quoted && key == by
	})
	return
}

func main() {
	fmt.Println(splitBy(os.Args[2], os.Args[1]))
}
