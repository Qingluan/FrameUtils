package web

var (
// Base = map[string]JSEvent{}
)

const (
	EVENT_CLICK = 0
	EVENT_ENTER = 1
	EVENT_FOCUS = 2
)

type JSEvent struct {
	Body   Js
	Toogle int
}

func (jsevent JSEvent) Method() string {
	switch jsevent.Toogle {
	case EVENT_FOCUS:
		return "focus"
	case EVENT_CLICK:
		return "click"
	case EVENT_ENTER:
		/*
			function(event){
			　　　　if(event.keyCode == 13){
			　　　　　　alert('你按下了Enter');
			　　　　}
			});
		*/
		return "keyup"
	default:
		return ""
	}
}

func (jsevent JSEvent) String() string {
	switch jsevent.Toogle {
	case EVENT_FOCUS:
		return WithFunc(func() Js {
			return jsevent.Body
		}, "").String()
	case EVENT_CLICK:
		return WithFunc(func() Js {
			return jsevent.Body
		}, "").String()
	case EVENT_ENTER:
		/*
			function(event){
			　　　　if(event.keyCode == 13){
			　　　　　　alert('你按下了Enter');
			　　　　}
			});
		*/
		return WithFunc(func() Js {
			return If(Val("event").Attr("keyCode").Eq(13), func() Js {
				return jsevent.Body
			})
		}, "", "event").String()
	default:
		return ""
	}
}
