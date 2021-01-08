package webevent

import (
	"fmt"
	"strings"
)

type SelectObj string

func CSSSelect(css string) SelectObj {
	if strings.HasPrefix(css, "#") {
		value := strings.TrimLeft(css, "#")
		return SelectObj(fmt.Sprintf("document.getElementById(\"%s\")", value))
	} else if strings.HasPrefix(css, ".") {
		value := strings.TrimLeft(css, ".")
		return SelectObj(fmt.Sprintf("document.getElementsByClass(\"%s\")[0]", value))
	}
	return ""
}

func (s SelectObj) GetAttr(name string) SelectObj {
	return SelectObj(string(s) + fmt.Sprintf(".getAttribute(\"%s\")", name))
}

func (s SelectObj) SetAttr(name, val string) SelectObj {
	return SelectObj(string(s) + fmt.Sprintf(".setAttribute(\"%s\",\"%s\");", name, val))
}

func (s SelectObj) GetText() SelectObj {
	return SelectObj(string(s) + ".textContent;")
}

func (s SelectObj) GetHTML() SelectObj {
	return SelectObj(string(s) + ".innerHTML;")
}

func (s SelectObj) String() string {
	return string(s)
}

func (s SelectObj) AppendChild(tag, id, content string, attrs Dict) SelectObj {
	tmp := `
	newele = document.createElement('` + tag + `');
	newele.id = "` + id + `";
	newele.textContent = "` + content + `"
	`
	for k, v := range attrs {
		tmp += "\n" + fmt.Sprintf("newele.setAttribute(\"%s\",\"%s\");", k, v)
	}
	return SelectObj(tmp + string(s) + ".appendChild(newele)")
}

func (s SelectObj) RemoveChild(id string) SelectObj {
	ele := "oldele = " + CSSSelect("#"+id).String() + ";"
	return SelectObj(ele + string(s) + ".removeChild(oldele);")
}

func (s SelectObj) Then(customjs string) SelectObj {
	return SelectObj(s.String() + ";" + customjs)
}

func (s SelectObj) WithId(id string) SelectObj {
	return SelectObj(s.String() + ";" + CSSSelect(id).String())
}

func (s SelectObj) ValueWith(v string) SelectObj {
	return SelectObj(s.String() + fmt.Sprintf(".value = \"%s\"", v))

}
