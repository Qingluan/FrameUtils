package web

import "fmt"

type Widget interface {
	String() string
	API() string
	GetID() string
	Margin(postion string, num int)
}

type InputWidgets interface {
	Widget
	Val() Js
}

type InputWidget struct {
	ID    string
	TP    string
	Style string
}

func (input SearchWidget) Val() Js {
	return Query("#" + input.ID).Val()
}

func (input SearchWidget) API() string {
	return input.Val().WithVar("data", func(val Js) Js {
		return Js(fmt.Sprintf("SendAction(\"%s\", \"%s\", data)", input.ID, input.TP))
	}).String()
}

type SearchWidget struct {
	ID string
	TP string

	Style string
}

func (search SearchWidget) GetID() string {
	return search.ID
}

func (search SearchWidget) Margin(postion string, num int) {
	switch postion {
	case "left":
		search.Style += fmt.Sprintf("margin-left: %d%%;", num)
	case "right":
		search.Style += fmt.Sprintf("margin-right: %d%%;", num)

	case "bottom":
		search.Style += fmt.Sprintf("margin-bottom: %d%%;", num)

	case "top":
		search.Style += fmt.Sprintf("margin-top: %d%%;", num)
	default:
		search.Style += fmt.Sprintf("margin: %d%%;", num)
	}
}
func (search SearchWidget) String() string {
	_, in := parse("Search")
	in.Input.ID = search.ID
	in.Label.Text = "Search >>>"
	in.Input.Style = search.Style
	return in.String()
}
