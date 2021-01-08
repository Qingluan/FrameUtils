package webevent

import (
	"fmt"
	"log"
	"regexp"
	"strings"
)

var (
	TAG_AND_ATTR = regexp.MustCompile(`<([a-zA-Z]{1,5})(.+?(?:id|class)=\s?['"]?.+?['"]\s?[\s]*)>`)
	ATTR_EXTRACT = regexp.MustCompile(`(\w+)\s*=\s*["']?(.+?)["'\s]`)
	TAG_FIND_TMP = `<(%s)(.+?(?:%s)=\s?['"]%s['"]\s*.+?)>`
	CLASS_CSS    = regexp.MustCompile(`(\.[\w\-]+?)`)
	ID_CSS       = regexp.MustCompile(`(\#[\w\-]+?)`)

	CSSCache = make(map[string]*Css)
)

type FrontXml string
type Dict map[string]string
type Element struct {
	Tag      string
	Id       string
	Attrs    Dict
	Text     string
	Children []ElementAble
}

func (ele Element) String() string {
	attrs := ""
	for k, v := range ele.Attrs {
		attrs += fmt.Sprintf(" %s=\"%s\"", k, v)
	}
	pre := `<` + ele.Tag + attrs + ">" + ele.Text
	for _, e := range ele.Children {
		pre += e.String()
	}
	return pre + "</" + ele.Tag + ">"
}

func (ele Element) Front() FrontXml {
	return FrontXml(ele.String())
}

func (ele Element) GetID() string {
	return ele.Id
}

func (ele Element) Content(c string) ElementAble {
	ele.Text = c
	return ele
}

type Css struct {
	name  string
	attrs map[string]string
}

func (css *Css) Set(name string, val string) {
	css.attrs[name] = val
}

func (css *Css) String() string {
	pre := fmt.Sprintf(`%s \n{`, css.name)
	for k, v := range css.attrs {
		pre += fmt.Sprintf("%s : %s ;\n", k, v)
	}
	return pre + "\n}\n"
}

func WithCss(selectcss string, name string, val ...string) string {
	if val != nil {
		if e, ok := CSSCache[selectcss]; ok {
			e.Set(name, val[0])
		} else {
			c := &Css{
				name:  selectcss,
				attrs: make(map[string]string),
			}
			c.Set(name, val[0])
			CSSCache[selectcss] = c
		}
		return val[0]
	} else {
		if c, ok := CSSCache[selectcss]; ok {
			return c.attrs[name]
		} else {
			return ""
		}
	}
}

func (self FrontXml) Tag() string {
	return TAG_AND_ATTR.FindStringSubmatch(string(self))[1]
}

func (self FrontXml) Attr(name string) string {

	f := TAG_AND_ATTR.FindStringSubmatch(string(self))[0]
	r, err := regexp.Compile(fmt.Sprintf(`(%s)\s*=\s*["']?(.+?)["'\s]`, name))
	if err != nil {
		log.Fatal(err)
	}
	return r.FindStringSubmatch(f)[2]
}

func (self FrontXml) CssSelect(css string) (a []FrontXml) {
	tag := "[a-zA-Z]{1,5}"
	name := "id|class"
	value := ".+?"

	var findre *regexp.Regexp
	if strings.HasPrefix(css, "#") {
		name = "id"
		value = strings.TrimLeft(css, "#")
	} else if strings.HasPrefix(css, ".") {
		name = "class"
		value = strings.TrimLeft(css, ".")
	} else {
		if strings.Contains(css, ".") {
			tag = strings.SplitN(css, ".", 2)[0]
			name = "class"
			value = strings.SplitN(css, ".", 2)[1]
		} else if strings.Contains(css, "#") {

			tag = strings.SplitN(css, "#", 2)[0]
			name = "id"
			value = strings.SplitN(css, "#", 2)[1]
		} else {
			tag = strings.TrimSpace(css)
		}
	}
	fmt.Println(fmt.Sprintf(TAG_FIND_TMP, tag, name, value))
	findre = regexp.MustCompile(fmt.Sprintf(TAG_FIND_TMP, tag, name, value))
	for _, e := range findre.FindAllStringSubmatch(string(self), -1) {
		a = append(a, FrontXml(e[0]))
	}
	return
}

func (self FrontXml) AddClickAction(callback func(data FlowData)) {
	tmp := `document.getElementById("%s").onclick = function(evt) {
		SendAction(this.id, "click", this.value);
		return false;
	};
	`
	if id := self.Attr("id"); id != "" {
		JsListeners = append(JsListeners, fmt.Sprintf(tmp, id))
		SetOnServerListener(strings.TrimSpace(id), callback)
	}
}

func (self FrontXml) AddInputAction(callback func(data FlowData)) {
	tmp := `document.getElementById("%s").onkeypress  = function(e) { 
		var keyCode = null; 
		if(e.which)
			keyCode = e.which;
		else if(e.keyCode)
			keyCode = e.keyCode;
		if(keyCode == 13) {
			SendAction(this.id, "input", this.value);
			return false; 
		}
		return true;
	}
	`
	if self.Tag() != "input" {
		return
	}
	if id := self.Attr("id"); id != "" {
		JsListeners = append(JsListeners, fmt.Sprintf(tmp, id))
		SetOnServerListener(strings.TrimSpace(id), callback)
	}
}

type ElementAble interface {
	String() string
	GetID() string
	Content(string) ElementAble
}

func Ele(name string, id string, childrem ...ElementAble) ElementAble {
	return Element{
		Tag:      name,
		Id:       id,
		Children: childrem,
	}
}
