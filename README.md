## FrameUtils for golang utils


### web

### JS Generate
> golang 生成 js代码

    测试

```golang
package main

import (
	"fmt"

	W "github.com/Qingluan/FrameUtils/web"
)

func main() {

	fmt.Println("//----- test2 js ------- create  click handle function ")
	fmt.Println(W.WithFunc(func() W.Js {
		return W.Query(W.Js("ele")).Val().WithVar("txt", func(val W.Js) W.Js {
			return W.Post("/post", "{name: "+val.String()+"}", func(res W.Js) W.Js {

				return W.If(res.Has("data"), func() W.Js {
					return W.Js("this").SetAttrValue("data", res.Get("data"))
				}, func() W.Js {
					return W.Js("console").Call("log", "this")
				})
			})

		})
	}, "click_ele", "ele").String())

}

```
> 运行结果

```js
//----- test2 js ------- create  click handle function
var click_ele = function(ele){
    var txt = $(ele).val();
    $.post("/post", {name: txt}, function(result){
        if (result.hasAttribute("data") != null){
            this.setAttribute("data", result["data"])
        }else{
            console.log(this)
        }

        console.log(result);
    });
    ;
};
```

#### 复杂点的例子， for , if , function, 参数设置

```golang
// for 

// Each 函数 等同于
// for ( i= 0; i <  items.length ; i ++){
//     ..... deal one .....
// }
// 
W.Query("p").Each(func(one Js) Js {
	return .... deal one ....
})

// If

// If 函数 等同于
// if( condition ){
//     ..... deal one .....
// }
// 

//   W.Query("title").Text().Eq("Hello")  ===> $("title")[0].textContent == "Hello"
// support comparation functions :  .Eq / MoreThan / MoreLess / Than / Less / HasKey / HasAttr
// example:
W.If( W.Query("title").Get(0).Text().Eq("Hello") ,func(one Js) Js {
	return .... deal one ....
}, func(){
	W.JsLog("no title")
})


// Set val
// var val = lastJs;
// ... do val .... 
lastJs.WithVar(val, func(val Js) Js{
	return .... handle val .....
})

// Set Dict
// var val = {};
lastJs.WithDict(val, func(val Js) Js{
	return .... handle val .....
})

// Set Array
// var val = [];
lastJs.WithArray(val, func(val Js) Js{
	return .... handle val .....
})


// multi set value like 
/*
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
lastJs.MapSet("some", W.Val("items"), isloop=true, func(one W.Js, key string) W.Js {
		return one.GetAttr(key)
	}, keys...)


// Post data


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
W.Post("/post", "{upload_data : data}",true)
/*
$.ajax({
	url: '/post',
	type: 'post',
	dataType: 'json',
	contentType: 'application/json',
	success: function (result) {
	%s
	},
	data: JSON.stringify({upload_Data: data})
});

*/

W.Post("/post", map[string]string{"upload_data" : "hello"},true , ...jscallback...)
/*
$.ajax({
	url: '/post',
	type: 'post',
	dataType: 'json',
	contentType: 'application/json',
	success: function (result) {
		... js callback...
	},
	data: JSON.stringify({"upload_Data": "hello"})
});

*/



W.Post("/post", "{upload_data : data}",false,...jscallback...)
// $.post("/post", {upload_data:data}, function(result){
//  	...jscallback...
//	console.log(result);
// })


W.Post("/post", map[string]string{"upload_data" : "Hello"},false,...jscallback...)
// $.post("/post", {"upload_data":"data"}, function(result){
// 		...jscallback...
//  	console.log(result);
// })




// exmaple :
jsClick = W.WithFunc(func() W.Js {
	return W.Js("").MapSet("upload_data", W.Val("ele").Children(), true, func(one W.Js, key string) W.Js {
		return one.GetAttr(key)
	}, "name", "data").NewLine(W.Post("/post", W.Val("upload_data"), func(res W.Js) W.Js {
		return W.JsLog(res.String())
	}))
}, "click_tr", "ele")
fmt.Println(jsClick)

```

==> 
```js
var click_tr = function(ele){
    ;
    var upload_data = [];

    _tmp_loop = ele.children;
    for(i=0; i< _tmp_loop.length ; i++){
        var __tmp_dict = {};
        __tmp_dict["name"] = _tmp_loop[i].getAttribute("name");
        __tmp_dict["data"] = _tmp_loop[i].getAttribute("data");
        __tmp_dict;
        upload_data.push(__tmp_dict);
    };
    $.post("/post", upload_data, function(result){
        console.log(result);
        console.log(result);
    });
    ;
};

```

### HTML Server Build

```golang
package main

import (
	"fmt"
	"log"
	"net/http"

	E "github.com/Qingluan/FrameUtils/engine"

	W "github.com/Qingluan/FrameUtils/web"
)

func main() {

	fmt.Println("//----- test1 html ------- table (xlsx/ csv) data => html ")
	obj, err := E.OpenObj("test.xlsx")
	if err != nil {
		log.Fatal(err)
	}
	// header := obj.Header()
	// fmt.Println(header)
	table := obj.ToHTML()
	fmt.Println(table)
	obj, err = E.OpenObj("test.csv")
	if err != nil {
		log.Fatal(err)
	}
	// header := obj.Header()
	// fmt.Println(header)
	table = obj.ToHTML()
	fmt.Println(table)
}
```

> 运行结果


```html
<table  class="table" >
    <thead class="thead-dark">
        <tr>
            <th scope="col">UserName</th>
            <th scope="col">ID</th>
            <th scope="col">Birth</th>
            <th scope="col">Phone</th>
        </tr>
    </thead><tbody>
        <tr onclick="click_tr(this);" ><td data="UserName" >UserName</td><td data="ID" >ID</td><td data="Birth" >Birth</td><td data="Phone" >Phone</td></tr>
        <tr onclick="click_tr(this);" ><td data="Zhang" >Zhang</td><td data="Shenfen" >Shenfen</td><td data="1995" >1995</td><td data=" 13000000001242100" > 13000000001242100</td></tr>
        <tr onclick="click_tr(this);" ><td data="Li" >Li</td><td data="Shenfen" >Shenfen</td><td data="1994" >1994</td><td data=" 130000000001240" > 130000000001240</td></tr>
        <tr onclick="click_tr(this);" ><td data="HZaosf" >HZaosf</td><td data="Shenfen" >Shenfen</td><td data="1991" >1991</td><td data=" 13000012300000" > 13000012300000</td></tr>
        <tr onclick="click_tr(this);" ><td data="Wang" >Wang</td><td data="Shen" >Shen</td><td data="1992" >1992</td><td data=" 130000000012400" > 130000000012400</td></tr>
        <tr onclick="click_tr(this);" ><td data="Zhong" >Zhong</td><td data="She" >She</td><td data="1993" >1993</td><td data=" 13000000000124120" > 13000000000124120</td></tr>
    </tbody>
</table>
```

### Console Tools

1. AppController

> this console cmd can build app by react . this will take some time to download react dev

2. ConsoleSearcher

> this tools will search some xlsx/ sql /csv in current dir .
