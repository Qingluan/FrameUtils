package web

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
)

// L : route (green) , action (red) , contents .... (white)
func L(route, action string, info ...interface{}) {
	g := color.New(color.BgCyan, color.FgYellow).SprintFunc()
	r := color.New(color.BgRed, color.FgWhite).SprintFunc()
	content := color.New(color.BgWhite, color.FgBlack).SprintFunc()
	msg := ""
	for _, i := range info {
		msg += strings.ReplaceAll(fmt.Sprintf("%v", i), "\n", "||")
	}
	fmt.Println(g(route), r(action), content(msg))
}
