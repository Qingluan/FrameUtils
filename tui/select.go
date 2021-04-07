package tui

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/manifoldco/promptui"
)

type CanString interface {
	String() string
}

func SelectOne(label string, selects []CanString) (CanString, bool) {
	prompt := promptui.Select{
		Label:        label,
		Items:        selects,
		HideSelected: false,
		Size:         10,
		Searcher: func(s string, ix int) bool {
			return strings.Contains(selects[ix].String(), s)
		},
	}
	i, _, err := prompt.Run()
	if err != nil {
		return nil, false
	}
	return selects[i], true
}

func GetPass(label string) string {
	fmt.Printf("%s:", label)
	buffer := bufio.NewReader(os.Stdin)
	line, _, _ := buffer.ReadLine()
	return strings.TrimSpace(string(line))
}
