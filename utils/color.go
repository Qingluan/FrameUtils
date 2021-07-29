package utils

import "github.com/fatih/color"

var (
	Magenta = color.New(color.FgMagenta).SprintFunc()

	Yellow = color.New(color.FgYellow).SprintFunc()
	Green  = color.New(color.FgGreen).SprintFunc()
	Blue   = color.New(color.FgBlue).SprintFunc()
	BGreen = color.New(color.FgBlack, color.BgGreen).SprintFunc()

	BYellow = color.New(color.FgWhite, color.BgYellow).SprintFunc()

	BBlue     = color.New(color.FgWhite, color.BgBlue).SprintFunc()
	Red       = color.New(color.FgRed).SprintFunc()
	BRed      = color.New(color.FgWhite, color.BgRed).SprintFunc()
	Bold      = color.New(color.Bold).SprintFunc()
	UnderLine = color.New(color.Underline, color.Bold).SprintFunc()
)
