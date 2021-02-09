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
	// WebsocketOnMessageQueue :Websocket Out message
	WebsocketOnMessageQueue = make(chan string, 30)
	// WebsocketOnMessage : how to deal mssage from websocket
	WebsocketOnMessage = func(id, tp, conetnt string) string { return "NotImplemment" }
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

// CodeLen :
func (js Js) CodeLen() int {
	return len(string(js))
}

//Last return last attr / name
func (js Js) Last() Js {
	last := ""
	if js.CCheck("\n") {
		fs := strings.Split(string(js), "\n")
		last = fs[len(fs)-1]
	}
	if strings.Contains(last, ".") {
		fs := strings.Split(last, ".")
		return Js(fs[len(fs)-1])
	}
	return Js(last)
}

// Lastline : get last line
func (js Js) Lastline() Js {
	last := ""
	if js.CCheck("\n") {
		fs := strings.Split(string(js), "\n")
		last = fs[len(fs)-1]
	} else {
		return js
	}
	return Js(last)
}

// Split : split by key then get ix
func (js Js) Split(key string, i int) Js {
	return Js(strings.Split(string(js), key)[i])
}

// Pop : back to last
func (js Js) Pop() Js {
	if strings.HasPrefix(string(js), "var ") && js.CCheck("\n") {
		fs := strings.Split(strings.TrimSpace(string(js)), "\n")
		return Js(strings.Join(fs[1:len(fs)-1], "\n"))
	} else if strings.Contains(string(js), "\n") && strings.HasSuffix(string(js), ";") {
		fs := strings.Split(strings.TrimSpace(string(js)[:js.CodeLen()-1]), "\n")
		return Js(strings.Join(fs[:len(fs)-1], "\n"))
	}
	fs := strings.Split(string(js), ".")
	return Js(strings.Join(fs[1:len(fs)-1], "."))

}

// CEndswithFunction : check if js is endswith(funcName)
func (js Js) CEndswithFunction(funcName string) bool {
	// e := false
	fs := strings.Split(string(js), ".")
	namefs := strings.Split(fs[len(fs)-1], "(")
	name := namefs[0]
	// fmt.Println("End Function:", name)
	return strings.TrimSpace(name) == funcName
}

//RemoveLastLine : remove last line
func (js Js) RemoveLastLine() Js {
	if js.CCheck("\n") {
		fs := strings.Split(strings.TrimSpace(string(js)), "\n")
		return Js(strings.Join(fs[:len(fs)-1], "\n"))
	}
	return js
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

/*Val :
获得字符串的变量， 其实等同于  Js(valStr)
*/
func Val(valStr string) Js {
	return Js(valStr)
}

// End $js + ";"
func (js Js) End() Js {
	return js + Js(";")
}

// JsLog : console.log(val)
func JsLog(args ...string) Js {
	return Val("console").Call("log", args...).End()
}

// NewLine ${js} + ";\n"
func (js Js) NewLine(then ...interface{}) Js {
	if then == nil {
		return js + Js(";\n")
	}
	switch then[0].(type) {
	case Js:
		return js + Js(";\n") + then[0].(Js)
	default:
		return js + Js(";\n") + Js(fmt.Sprint(then[0]))
	}
}

// ForVal Js()
func (js Js) ForVal(valName string) Js {
	return js.NewLine() + Val(valName)
}

// func (js Js) Then(other Js) Js{
// 	return js + other
// }

// Text : .textContent
func (js Js) Text() Js {
	return js + Js(".textContent")
}

/*MapSet doc :
@isLoop: if true:
	example:
"""
	some = []
	for (i =0; i < items.length; i ++){
		tmp = {}
		for (ki = 0 ; ki < keys.length; ki++){
			tmp[keys[ki]] = ...some func...(items[i] ,keys[ki])
		}

		some.push(tmp)
	}
"""
	if false:
"""
	tmp = {}
	for (ki = 0 ; ki < keys.length; ki++){
		tmp[keys[ki]] = ...some func...(items[i] ,keys[ki])
	}

"""
*/
func (js Js) MapSet(valName string, itemsName Js, isLoop bool, mapFunc func(one Js, key string) Js, keys ...string) Js {
	if isLoop {
		return js.NewLine().WithArray(valName, func(val Js) Js {
			return itemsName.Each(func(one Js) Js {
				return Js("").WithDict("__tmp_dict", func(tmp_dict Js) Js {
					for _, key := range keys {
						tmp_dict = tmp_dict.SetVal(key, mapFunc(one, key))
					}
					return tmp_dict
				}).RemoveLastLine().ForVal(valName).Call("push", "__tmp_dict")

			})
		})
	}
	return js.NewLine().WithDict(valName, func(val Js) Js {

		return itemsName.Each(func(one Js) Js {
			for _, key := range keys {
				val = val.SetVal(key, mapFunc(one, key))
			}
			return val.Pop()
		})
	})
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
	case int:
		return Js("(") + js + Js(") == ") + Js(fmt.Sprintf("%d", target.(int)))
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

//Attr : get attr with : ${js}.name
func (js Js) Attr(name string) Js {
	return Js(string(js) + "." + name)
}

//GetAttr :
func (js Js) GetAttr(name string) Js {
	return js + Js(fmt.Sprintf(".getAttribute(\"%s\")", name))
}

// AsVar : var name = ${js};\n
func (js Js) AsVar(varName string) Js {
	return Js("var " + strings.ReplaceAll(varName, " ", "_") + " = " + string(js) + ";\n")
}

// WithVar : var name = ${js};\n
func (js Js) WithVar(varName string, then func(val Js) Js) Js {
	return Js("var "+strings.ReplaceAll(varName, " ", "_")+" = "+string(js)+";\n") + then(Js(strings.ReplaceAll(varName, " ", "_")))
}

// WithDict : var name = {};\n
// do some for this dict
func (js Js) WithDict(varName string, then func(val Js) Js) Js {
	return js + Js("var "+strings.ReplaceAll(varName, " ", "_")) + Js(" = {};\n") + then(Js(strings.ReplaceAll(varName, " ", "_")))
}

//WithArray :
func (js Js) WithArray(varName string, then func(arrval Js) Js) Js {
	return js + Js("var "+strings.ReplaceAll(varName, " ", "_")) + Js(" = [];\n") + then(Js(strings.ReplaceAll(varName, " ", "_")))
}

// Val : ${js}.val()
func (js Js) Val() Js {
	return js + Js(".val()")
}

// Intendence : all line : \n => \n\t
func (js Js) Intendence() Js {
	return Js("    " + strings.ReplaceAll(string(js), "\n", "\n    "))
}

/*SetVal :
${js}["name"] = "${value}";
${js}

*/
func (js Js) SetVal(name, value interface{}) Js {
	switch value.(type) {
	case Js:
		// pre := "_tmpvalue = " + value.(Js).NewLine()
		// if value.(Js).CCheck("\n") {
		// 	panic("include \n in setattr() ")
		// }
		// log.Println("V:", js)
		return js + Js(fmt.Sprintf("[\"%s\"] = %s;\n", name, value)) + js.Lastline().Split("[", 0)
	default:
		return js + Js(fmt.Sprintf("[\"%s\"] =\"%s\";\n", name, value)) + js.Lastline().Split("[", 0)

	}
}

// SetAttr : ${js}.setAttribute("name", "value") ; if value is string  else ${js}.setAttribute("name", value)
func (js Js) SetAttr(name, value interface{}) Js {
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

// Children : get child node by ix
func (js Js) Children(ix ...int) Js {
	if ix == nil {
		return js + Js(".children")
	}
	return js + Js(fmt.Sprintf(".children[%d]", ix))
}

// Lengnth : get loop's count
func (js Js) Lengnth() Js {
	return js + Js(".length")
}

// Each : loop if ${self} can loop
func (js Js) Each(eachOne func(Js) Js) Js {
	pre := fmt.Sprintf(`
_tmp_loop = %s;
for(i=0; i< %s ; i++){
%s
}`, js, Js("_tmp_loop").Lengnth(), eachOne(Js("_tmp_loop[i]")).End().Intendence())
	return Js(pre)
}

/*Post :
	@action : set url
	@args :  map[string]string/ string / Js ; upload data
	@isJson: bool ; if true use $.ajax()
			$.ajax({
				url: '%s',
				type: 'post',
				dataType: 'json',
				contentType: 'application/json',
				success: function (result) {
				%s
				},
				data: JSON.stringify(%s)
			});
	@callback :
		func(result Js) Js{
			handle ... callback result
		}

# this will be compile twice .
	by:
		${ ${inner} } => ${ innerstring } => "endstring"

$.post("${action}", json.dumps(args), function(result){
	${ ${callback( result )} }
	console.log(result);
})
*/
func Post(action string, args interface{}, isJSON bool, callback ...func(j Js) Js) Js {
	var data []byte
	switch args.(type) {
	case map[string]string:
		data, _ = json.Marshal(args)
	case string:
		data = []byte(args.(string))

	case Js:
		data = []byte(args.(Js).String())
	}
	postArgs := string(data)
	jsfunction := ""
	if callback != nil {
		jsfunction = callback[0](Js("result")).Intendence().String()
	}
	if isJSON {
		base := fmt.Sprintf(`$.ajax({
	url: '%s',
	type: 'post',
	dataType: 'json',
	contentType: 'application/json',
	success: function (result) {
	%s
	},
	data: JSON.stringify(%s)
});

`, action, jsfunction, postArgs)
		return Js(base)

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

// RegistWebSocketFunc : define how to deal onMessage in server and front
func RegistWebSocketFunc(name string, onBrowser func(comJs Js) Js) Js {
	base := fmt.Sprintf(`
actions["%s"] = function(data){
%s	
}
`, name, onBrowser("data").Intendence())
	RegistedWebSocketFuncs[name] = Js(base)
	return Js(base)
}
