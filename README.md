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
