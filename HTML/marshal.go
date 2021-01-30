package HTML

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
)

var (
	extractValue = regexp.MustCompile(`((?:=\s*)[\w]+)`)
	tagTemp      = "<%s ></%s>"
)

func IsNull(value reflect.Value) bool {
	v := value.Interface()
	switch v.(type) {
	case string:
		return v.(string) == ""
	case float32:
		return value.IsZero()
	case float64:
		return value.IsZero()
	case int:
		return value.IsZero()
	default:
		return value.IsNil()
	}
}

func MarshalHTML(field interface{}) string {
	fieldrec := reflect.ValueOf(field)
	tagStr := tagTemp
	values := []string{}
	text := ""
	subs := []string{}
	for i := 0; i < fieldrec.NumField(); i++ {
		property := fieldrec.Field(i)

		tag := fieldrec.Type().Field(i).Tag.Get("html")
		tagname := tag
		defaultvalue := fmt.Sprintf("%v", property.Interface())
		if strings.Contains(tagname, "=") {
			tagnamefs := strings.SplitN(tagname, "=", 2)
			tagname = tagnamefs[0]
			if IsNull(property) {
				defaultvalue = strings.ReplaceAll(tagnamefs[1], "_", "-")
			}
			// defaultvalue = strings.ReplaceAll(extractValue.FindString(tagname), "_", "-")
		} else if property.Kind().String() == "slice" {
			s := reflect.ValueOf(property.Interface())
			for i := 0; i < s.Len(); i++ {
				ele := s.Index(i)
				subs = append(subs, MarshalHTML(ele.Interface()))
			}
			continue
		} else if property.Kind().String() == "struct" {
			subs = append(subs, MarshalHTML(property.Interface()))
		}

		// fmt.Println("test: ", tag, "|", tagname, defaultvalue)
		if tagname == "tag" {
			tagStr = fmt.Sprintf(tagTemp, defaultvalue, defaultvalue)
		} else if tagname == "text" {
			text = defaultvalue
		} else if property.Kind() == reflect.Bool {
			if property.Interface().(bool) {
				values = append(values, fmt.Sprintf("%s", tagname))
			}
		} else {
			values = append(values, fmt.Sprintf("%s=\"%s\"", tagname, defaultvalue))
		}
	}
	fs := strings.SplitN(tagStr, "><", 2)
	return fs[0] + strings.Join(values, " ") + ">" + text + "\n" + strings.Join(subs, "\n") + "<" + fs[1]
}
