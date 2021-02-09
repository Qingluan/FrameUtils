package web

import (
	"fmt"

	"github.com/Qingluan/FrameUtils/HTML"
)

type Widget interface {
	String() string
	API() string
	GetID() string
	Margin(postion string, num interface{})
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

func (input *SearchWidget) Val() Js {
	return Query("#" + input.ID).Val()
}

func (input *SearchWidget) API() string {
	return input.Val().WithVar("data", func(val Js) Js {
		return Js(fmt.Sprintf("SendAction(\"%s\", \"%s\", data)", input.ID, input.TP))
	}).String()
}

type SearchWidget struct {
	ID    string `html:"id"`
	TP    string
	Style string `html:"style"`
}

func (search *SearchWidget) GetID() string {
	return search.ID
}

func (search *SearchWidget) Margin(postion string, nums interface{}) {
	num := ""
	switch nums.(type) {
	case int:
		num = fmt.Sprintf("%d%%", nums.(int))
	case string:
		// num, _ = strconv.Atoi(nums.(string))
		num = nums.(string)
	}
	switch postion {
	case "left":
		search.Style += fmt.Sprintf("margin-left: %s;", num)
	case "right":
		search.Style += fmt.Sprintf("margin-right: %s;", num)

	case "bottom":
		search.Style += fmt.Sprintf("margin-bottom: %s;", num)

	case "top":
		search.Style += fmt.Sprintf("margin-top: %s;", num)
	default:
		search.Style += fmt.Sprintf("margin: %s;", num)
	}
}
func (search *SearchWidget) String() string {
	in := Field{
		Tag:   "div",
		Class: "input-group input-group-lg",
		ID:    search.ID + "-div",
		Subs: []Field{
			{
				Tag:   "span",
				Class: "input-group-text",
				Text:  "Search >>",
			},
			{
				ID:    search.ID,
				Style: search.Style,
			},
		},
	}
	// in.Input.ID = search.ID
	// in.Label.Text = "Search >>>"
	// in.Input.Style = search.Style
	return HTML.MarshalHTML(in, "\n")
}
