package WebRender

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/Qingluan/FrameUtils/HTML"
)

var (
	reType = regexp.MustCompile(`${(\w+)}`)
)

type Field struct {
	Tag         string `html:"tag=input"`
	Name        string `html:"name"`
	ID          string `html:"id"`
	Value       string `html:"value"`
	Type        string `html:"type=text"`
	Class       string `html:"class=form-controll"`
	Placeholder string `html:"placeholder"`
	Readonly    bool   `html:"readonly"`
	Text        string `html:"text"`
	Subs        []Field
}

func (feild Field) Set(value string) {
	feild.Value = value
}

func (feild Field) Get() string {
	return feild.Value
}

type Input interface {
	String() string
	Set(value string)
	Get() string
}

type BootrapInput struct {
	Class       string `html:"class=form-group"`
	tag         string `html:"tag=div"`
	Name        string `json:"name"`
	ID          string `json:"id"`
	Placeholder string `json:"placeholder"`
	Value       string `json:"value"`
	Type        string `json:"type"`
	Describle   string `json:"desc"`
	Label       struct {
		tag  string `html:"tag=label"`
		text string `html:"text"`
	}
	Input Field
}

type Form map[string]Input

/*Parse from str to input
Name / type[value]

example:
user  ==>  <input type=text name="user">
user / zhangsan  ==>  <input type=text name="user" value="zhangsan" >
id / ${number} ==> <input type=num name="id">
id / 1234 ==> <input type=num name="id" value="1234" >

ident / id card / passport  ==> <select class="form-control" name=ident ><option>id card</option><option>passport </option></select>

*/
func Parse(raw string) (boot BootrapInput) {
	n := strings.Count(raw, "/")
	name := strings.TrimSpace(raw)
	types := "text"
	defaults := ""
	if n == 0 {

	} else if n == 1 {
		fs := strings.Split(raw, "/")
		name = strings.TrimSpace(fs[0])
		if strings.Contains(fs[1], "${") && strings.Contains(fs[1], "}") {
			types = reType.FindString(fs[1])
			fs[1] = reType.ReplaceAllString(fs[1], "")
		}
		if strings.TrimSpace(fs[1]) != "" {
			defaults = strings.TrimSpace(fs[1])
		}
	} else {
		fs := strings.Split(raw, "/")
		for _, f := range fs[1:] {
			boot.Input.Subs = append(boot.Input.Subs, Field{
				Text: strings.TrimSpace(f),
				Tag:  "option",
			})
		}
		types = ""
		boot.tag = "select"
	}
	boot.Name = name
	boot.Type = types
	boot.Value = defaults

	return
}

func (boot BootrapInput) String() string {
	if boot.ID != "" {
		boot.Input.ID = "Input" + boot.ID
	}
	boot.Input.Name = boot.Name
	if boot.Value != "" {
		boot.Input.Value = boot.Value
	}

	if boot.Type != "" {
		boot.Input.Type = boot.Type
	}

	if boot.Class != "" {
		boot.Input.Class = boot.Class
	}
	boot.Label.text = boot.Describle
	boot.Input.Placeholder = boot.Placeholder
	base := HTML.MarshalHTML(boot)
	return base
}

func (boot BootrapInput) Set(value string) {
	boot.Input.Set(value)
}

func (boot BootrapInput) Get() string {
	return boot.Input.Get()
}

func (form Form) Set(name, value string) {
	form[name].Set(value)
}

func (form Form) RecvFromReq(r *http.Request) {
	if r != nil {
		return
	}
	for name := range form {
		v := r.Form.Get(name)
		form.Set(name, v)
	}
	return
}

func (form Form) Render(jsFunction ...string) string {
	j := ""
	if jsFunction != nil {
		j = fmt.Sprintf("onclick=\"return %s\"", jsFunction[0])
	}
	subs := []string{}
	for _, v := range form {
		subs = append(subs, v.String())
	}
	sub := strings.Join(subs, "\n")
	return fmt.Sprintf(`<form >
	%s
	<button type="submit" class="btn btn-primary" %s >Ok</button>
</form>`, sub, j)
}
