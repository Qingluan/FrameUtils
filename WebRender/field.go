package WebRender

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/Qingluan/FrameUtils/HTML"
)

var (
	reType       = regexp.MustCompile(`\$\{(\w+)\}`)
	JSOptionPost = "method=\"POST\""
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
	Tag         string `html:"tag=div"`
	Name        string `json:"name"`
	ID          string `json:"id"`
	Placeholder string `json:"placeholder"`
	Value       string `json:"value"`
	Type        string `json:"type"`
	Describle   string `json:"desc"`
	Label       *struct {
		Tag  string `html:"tag=label"`
		Text string `html:"text"`
	}
	Input *struct {
		Tag         string `html:"tag=input"`
		Name        string `html:"name"`
		ID          string `html:"id"`
		Value       string `html:"value"`
		Type        string `html:"type=text"`
		Class       string `html:"class=form-controll"`
		Placeholder string `html:"placeholder"`
		Readonly    bool   `html:"readonly"`
		Text        string `html:"text"`
	}
	Subs []Field
}

type Form map[string]Input

/*ParseMultilines from str to input
Name / type[value]

${type} : email / text/ number / password /checkbox
if "/" more than one will become radios/ select:
	2 < x < 5:
		radios
	5 <= x :
		selects

example:
user  ==>  <input type=text name="user">
user / zhangsan  ==>  <input type=text name="user" value="zhangsan" >
id / ${number} ==> <input type=num name="id">
id / 1234 ==> <input type=num name="id" value="1234" >

ident / id card / passport  ==> <select class="form-control" name=ident ><option>id card</option><option>passport </option></select>

*/
func ParseMultilines(raws []string) (form Form) {
	form = make(Form)
	for _, l := range raws {
		if strings.TrimSpace(l) != "" {
			name, bootstrapinput := parse(strings.TrimSpace(l))
			form[name] = bootstrapinput
		}
	}
	return
}

func parse(raw string) (name string, boot BootrapInput) {
	n := strings.Count(raw, "/")
	name = strings.TrimSpace(raw)
	types := "email"
	defaults := ""
	boot.Input = new(struct {
		Tag         string `html:"tag=input"`
		Name        string `html:"name"`
		ID          string `html:"id"`
		Value       string `html:"value"`
		Type        string `html:"type=text"`
		Class       string `html:"class=form-controll"`
		Placeholder string `html:"placeholder"`
		Readonly    bool   `html:"readonly"`
		Text        string `html:"text"`
	})
	boot.Label = new(struct {
		Tag  string `html:"tag=label"`
		Text string `html:"text"`
	})
	if n == 0 {

	} else if n == 1 {
		fs := strings.Split(raw, "/")
		name = strings.TrimSpace(fs[0])
		if strings.Contains(fs[1], "${") && strings.Contains(fs[1], "}") {
			types = reType.FindString(fs[1])
			types = types[2 : len(types)-1]
			fs[1] = reType.ReplaceAllString(fs[1], "")
		}
		if strings.TrimSpace(fs[1]) != "" {
			defaults = strings.TrimSpace(fs[1])
		}
	} else if 1 < n && n <= 4 {
		fs := strings.Split(raw, "/")
		name = strings.TrimSpace(fs[0])
		types = " "
		boot.Tag = "div"
		boot.Class = "form-check-group"
		boot.Input = nil
		boot.Label = nil
		for _, f := range fs[1:] {
			boot.Subs = append(boot.Subs, Field{
				Class: "form-check",
				Type:  " ",
				Tag:   "div",
				Subs: []Field{
					{
						Tag:   "input",
						Class: "form-check-input",
						Type:  "radio",
						Name:  name,
						Value: strings.TrimSpace(f),
					},
					{
						Tag:   "label",
						Class: "form-check-label",
						Text:  strings.TrimSpace(f),
					},
				},
			})
		}

	} else {
		fs := strings.Split(raw, "/")
		name = strings.TrimSpace(fs[0])
		types = ""
		types = "select"

		for _, f := range fs[1:] {
			boot.Subs = append(boot.Subs, Field{
				Text:  strings.TrimSpace(f),
				Tag:   "option",
				Class: " ",
				Type:  " ",
			})
		}
		types = ""
		boot.Tag = "select"
	}
	boot.Name = name
	boot.Type = types
	boot.Value = defaults
	boot.Describle = name
	// log.Println(boot)
	return
}

func (boot BootrapInput) String() string {

	if boot.Label != nil {
		boot.Label.Text = boot.Describle
	}
	if boot.Input != nil {
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
		boot.Input.Placeholder = boot.Placeholder

	}
	return HTML.MarshalHTML(boot, "\t")

}

func (boot BootrapInput) Set(value string) {
	boot.Input.Value = value
}

func (boot BootrapInput) Get() string {
	return boot.Input.Value
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

func (form Form) Render(jsOptions ...string) string {
	jcall := ""
	action := ""
	if jsOptions != nil {
		for _, j := range jsOptions {
			if strings.Contains(j, "()") {
				jcall = fmt.Sprintf("onclick=\"return %s\"", j)
			} else if strings.HasPrefix(j, "/") {
				action = fmt.Sprintf(" action=\"%s\"", j)
			} else {
				action += " " + j
			}

		}

	}
	subs := []string{}
	for _, v := range form {
		subs = append(subs, v.String())
	}
	sub := strings.Join(subs, "\n")
	return fmt.Sprintf(`<form %s >
	%s
	<button type="submit" class="btn btn-primary" %s >Ok</button>
</form>`, action, sub, jcall)
}
