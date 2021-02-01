package web

import (
	"encoding/json"
	"fmt"
	"strings"
)

var (
	// PostionBeforeEnd : position option
	PostionBeforeEnd = "beforeend"
	// PostionAfterBegin : position option
	PostionAfterBegin = "afterbegin"
)

// Js : js wraper
type Js string

func (js Js) String() string {
	return string(js)
}

// CCheck : compile check if Js has string name..
func (js Js) CCheck(codes ...string) bool {
	e := false
	for _, code := range codes {
		e = strings.Contains(string(js), code)
		if e {
			return e
		}
	}
	return e
}

// Pop : back to last
func (js Js) Pop() Js {
	if strings.HasPrefix(string(js), "var ") && js.CCheck("\n") {
		fs := strings.Split(strings.TrimSpace(string(js)), "\n")
		return Js(strings.Join(fs[1:len(fs)-1], "\n"))
	}
	fs := strings.Split(string(js), ".")
	return Js(strings.Join(fs[1:len(fs)-1], "."))

}

// CEndswithFunction : check if js is endswith(funcName)
func (js Js) CEndswithFunction(funcName string) bool {
	// e := false
	fs := strings.Split(string(js), ".")
	namefs := strings.Split(fs[len(fs)-1], "(")
	name := namefs[len(namefs)-1]
	return strings.TrimSpace(name) == funcName
}

/*If :


if $condition is Js  and  ">","==",">=","<=" in  $condition :

===>
	if (${condition} ){
		$runJs
	}[ else { . $elseRun.. }]

else if $condition is Js but not "><==":
	if (${condition != null} ){
		$runJs
	}[ else { . $elseRun.. }]

else:
	if ("${condition}"){
		$runJs
	}[ else { . $elseRun.. }]


*/
func If(condition interface{}, runJs func() Js, elseRun ...func() Js) Js {
	pre := ""

	switch condition.(type) {
	case Js:
		if condition.(Js).CCheck("\n") {
			panic("\"" + string(condition.(Js)) + "\"" + " can not be as if ( $js ){ ...}")
		}
		if condition.(Js).CCheck("==", " > ", " < ", ">=", "<=") || condition.(Js).CEndswithFunction("hasAttribute") {

			pre += fmt.Sprintf(`if (%s){
%s
}`, condition.(Js), runJs().Intendence())
		} else {
			pre += fmt.Sprintf(`if (%s != null){
%s
}`, condition.(Js), runJs().Intendence())
		}
	default:
		pre += fmt.Sprintf(`if ( %s ){
%s
}`, condition.(string), runJs().Intendence())
	}
	if elseRun != nil {
		pre += fmt.Sprintf(`else{
%s
}`, elseRun[0]().Intendence())
	}
	return Js(pre + "\n")
}

// Call : ${js}.Name(args.....)
func (js Js) Call(name string, args ...string) Js {
	return Js(string(js) + "." + name + "(" + strings.Join(args, ",") + ")")
}

// AsArgToCall Name(${js}, others....)
func (js Js) AsArgToCall(funName string, otherArgs ...string) Js {
	args := append([]string{string(js)}, otherArgs...)
	return Js(funName + "(" + strings.Join(args, ",") + ")")
}

// Q $(${js}) !!! not $("${js}")
func (js Js) Q() Js {
	return Js("$(" + js + ")")
}

// End $js + ";"
func (js Js) End() Js {
	return js + Js(";")
}

// NewLine ${js} + ";\n"
func (js Js) NewLine() Js {
	return js + Js(";\n")
}

// AsVarThen var name = ${js};\n name
func (js Js) AsVarThen(name string) Js {
	return js.AsVar(name) + Js(name)
}

/*Get :
if  i is int:
	${js}[i]
else if i is string:
	${js}["name"]
*/
func (js Js) Get(i interface{}) Js {
	switch i.(type) {
	case string:
		return Js(string(js) + fmt.Sprintf("[\"%s\"]", i))
	default:
		return Js(string(js) + fmt.Sprintf("[%v]", i))
	}

}

/*HasAttr :
used in Element or class
$(js).hasAttribute(name)

*/
func (js Js) HasAttr(name string) Js {
	return js + Js(fmt.Sprintf(".hasAttribute(\"%s\")", name))
}

/*HasKey :
	used in dict
$(js).hasOwnProperty(name)

*/
func (js Js) HasKey(key string) Js {
	return js + Js(fmt.Sprintf(".hasOwnProperty(\"%s\")", key))
}

/*Eq :

if target is Js:
	(${js}) == $target
else:
	${js} == "$target"
*/
func (js Js) Eq(target interface{}) Js {
	switch target.(type) {
	case Js:
		return Js("(") + js + Js(") == ") + target.(Js)
	default:
		return Js("(") + js + Js(") == \""+target.(string)+"\"")
	}
}

/*MoreThan :

if target is Js:
	(${js}) >= $target
else:
	${js} >= "$target"
*/
func (js Js) MoreThan(target interface{}) Js {
	switch target.(type) {
	case Js:
		return Js("(") + js + Js(") >= ") + target.(Js)
	default:
		return Js("(") + js + Js(") >= \""+target.(string)+"\"")
	}
}

/*LessThan :

if target is Js:
	(${js}) <= $target
else:
	${js} <= "$target"
*/
func (js Js) LessThan(target interface{}) Js {
	switch target.(type) {
	case Js:
		return Js("(") + js + Js(") <= ") + target.(Js)
	default:
		return Js("(") + js + Js(") <= \""+target.(string)+"\"")
	}
}

/*More :

if target is Js:
	(${js}) > $target
else:
	${js} > "$target"
*/
func (js Js) More(target interface{}) Js {
	switch target.(type) {
	case Js:
		return Js("(") + js + Js(") > ") + target.(Js)
	default:
		return Js("(") + js + Js(") > \""+target.(string)+"\"")
	}
}

/*Less :

if target is Js:
	(${js}) < $target
else:
	${js} < "$target"
*/
func (js Js) Less(target interface{}) Js {
	switch target.(type) {
	case Js:
		return Js("(") + js + Js(") < ") + target.(Js)
	default:
		return Js("(") + js + Js(") < \""+target.(string)+"\"")
	}
}

//A : get attr with : ${js}.name
func (js Js) A(name string) Js {
	return Js(string(js) + "." + name)
}

// AsVar : var name = ${js};\n
func (js Js) AsVar(varName string) Js {
	return Js("var " + strings.ReplaceAll(varName, " ", "_") + " = " + string(js) + ";\n")
}

// WithVar : var name = ${js};\n
func (js Js) WithVar(varName string, then func(val Js) Js) Js {
	return Js("var "+strings.ReplaceAll(varName, " ", "_")+" = "+string(js)+";\n") + then(Js(strings.ReplaceAll(varName, " ", "_")))
}

// Val : ${js}.val()
func (js Js) Val() Js {
	return js + Js(".val()")
}

// Intendence : all line : \n => \n\t
func (js Js) Intendence() Js {
	return Js("    " + strings.ReplaceAll(string(js), "\n", "\n    "))
}

// SetAttrValue : ${js}.setAttribute("name", "value") ; if value is string  else ${js}.setAttribute("name", value)
func (js Js) SetAttrValue(name, value interface{}) Js {
	switch value.(type) {
	case Js:
		// pre := "_tmpvalue = " + value.(Js).NewLine()
		if value.(Js).CCheck("\n") {
			panic("include \n in setattr() ")
		}
		return js + Js(fmt.Sprintf(`.setAttribute("%s", %s)`, name, value))
	default:
		return js + Js(fmt.Sprintf(`.setAttribute("%s","%s")`, name, value))

	}
}

/*AsFunction :
name = function(args...){
	${js}
}
*/
func (js Js) AsFunction(name string, args ...string) Js {
	if name == "" {
		return Js(fmt.Sprintf(`function(%s){
%s
}`, strings.Join(args, ","), js.Intendence()))

	}
	return Js(fmt.Sprintf(`var %s = function(%s){
%s
}`, name, strings.Join(args, ","), js.End().Intendence())).NewLine()

}

/*WithFunc :
name = function(args...){
	${js}
}
*/
func WithFunc(body func() Js, name string, args ...string) Js {
	if name == "" {
		return Js(fmt.Sprintf(`function(%s){
%s
}`, strings.Join(args, ","), body().Intendence()))

	}
	return Js(fmt.Sprintf(`var %s = function(%s){
%s
}`, name, strings.Join(args, ","), body().End().Intendence())).NewLine()

}

// AppendHTML : ${js}.insertAdjacentHTML("${position}", "${html}")
func (js Js) AppendHTML(position string, html string) Js {
	return Js(string(js) + fmt.Sprintf(".insertAdjacentHTML(\"%s\",\"%s\"))", position, html))
}

// AppendNode : ${js}.appendChild(${node})
func (js Js) AppendNode(nodeName interface{}) Js {
	switch nodeName.(type) {
	case Js:
		return Js(string(js) + ".appendChild(" + string(nodeName.(Js)) + ")")
	default:
		return Js(string(js) + ".appendChild(" + fmt.Sprintf("%v", nodeName) + ")")
	}
}

/*Post :
# this will be compile twice .
	by:
		${ ${inner} } => ${ innerstring } => "endstring"

$.post("${action}", json.dumps(args), function(result){
	${ ${callback( result )} }
	console.log(result);
})
*/
func Post(action string, args interface{}, callback ...func(j Js) Js) Js {
	var data []byte
	switch args.(type) {
	case map[string]string:
		data, _ = json.Marshal(args)
	case string:
		data = []byte(args.(string))
	}
	postArgs := string(data)
	jsfunction := ""
	if callback != nil {
		jsfunction = callback[0](Js("result")).Intendence().String()
	}
	base := fmt.Sprintf(`$.post("%s", %s, function(result){
%s
    console.log(result);
});
`, action, postArgs, jsfunction)
	return Js(base)
}

/*Query :
$("${selector}")
*/
func Query(selector interface{}) Js {
	switch selector.(type) {
	case Js:
		return Js(fmt.Sprintf("$(%s)", selector.(Js).String()))
	default:
		return Js(fmt.Sprintf("$(\"%s\")", selector))

	}
}
