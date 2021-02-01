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

func MarshalHTML(fieldptr interface{}, identy ...string) string {
	res := reflect.ValueOf(fieldptr)
	var fieldrec reflect.Value
	// isPoint := false
	if res.Kind() == reflect.Ptr {
		// field = fieldptr

		// isPoint = true
		fieldrec = res.Elem()
		if !fieldrec.IsValid() {

			// fmt.Println("Val:", fieldrec)

			return ""
		}

		// fmt.Println("Ptr Num:", fieldrec)

	} else {
		// field = fieldptr
		fieldrec = res
	}

	// if isPoint {
	// 	fieldrec = res.Elem()
	// }
	tagStr := tagTemp
	values := []string{}
	text := ""
	subs := []string{}
	// found := false
	// if fieldrec.IsZero() {
	// 	return ""
	// }
	for i := 0; i < fieldrec.NumField(); i++ {
		property := fieldrec.Field(i)
		// if isPoint {
		// 	property = property.Elem()
		// }

		tag := fieldrec.Type().Field(i).Tag.Get("html")
		tagname := tag

		defaultvalue := ""
		if property.IsValid() {

			// fmt.Println("Tgus:", property.Interface())
			// if isPoint {

			// defaultvalue = fmt.Sprintf("%v", property.Addr().Interface())
			// } else {
			defaultvalue = fmt.Sprintf("%v", property.Interface())
			// }
		} else {
		}
		if strings.Contains(tagname, "=") {
			tagnamefs := strings.SplitN(tagname, "=", 2)
			tagname = tagnamefs[0]
			if defaultvalue == "" {
				defaultvalue = strings.ReplaceAll(tagnamefs[1], "_", "-")
			}
			// defaultvalue = strings.ReplaceAll(extractValue.FindString(tagname), "_", "-")
		} else if property.Kind().String() == "slice" {
			s := reflect.ValueOf(property.Interface())
			for i := 0; i < s.Len(); i++ {
				ele := s.Index(i)
				subs = append(subs, MarshalHTML(ele.Interface(), identy[0]+identy[0]))
			}
			continue
		} else if property.Kind().String() == "struct" {
			subs = append(subs, MarshalHTML(property.Interface(), identy[0]+identy[0]))
		} else if property.Kind() == reflect.Ptr {
			subs = append(subs, MarshalHTML(property.Interface(), identy[0]+identy[0]))
		}

		// fmt.Println("test: ", tag, "|", tagname, defaultvalue)
		if tagname == "tag" {
			// if strings.TrimSpace(defaultvalue) != "" {
			// found = true
			tagStr = fmt.Sprintf(tagTemp, defaultvalue, defaultvalue)
			// }
		} else if tagname == "text" {
			text = defaultvalue
		} else if property.Kind() == reflect.Bool {
			if property.Interface().(bool) {
				values = append(values, fmt.Sprintf("%s", tagname))
			}
		} else {
			if tagname != "" && strings.TrimSpace(defaultvalue) != "" {
				values = append(values, fmt.Sprintf("%s=\"%s\"", tagname, defaultvalue))
			}
		}
	}
	// if !found {
	// 	return ""
	// }
	fs := strings.SplitN(tagStr, "><", 2)
	if identy != nil {
		return fs[0] + strings.Join(values, " ") + ">" + "\n" + identy[0] + text + "\n" + identy[0] + strings.Join(subs, "\n"+identy[0]) + "<" + fs[1]
	} else {
		return fs[0] + strings.Join(values, " ") + ">" + "\n" + text + "\n" + strings.Join(subs, "\n") + "<" + fs[1]

	}
}
