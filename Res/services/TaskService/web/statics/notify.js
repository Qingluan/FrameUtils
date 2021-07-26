var JSCONTENT = ``

var goodblue = "#00A8FC";
// var serverAddress = `http://${CONTROLL_IP}:51081/`
var notify_stack = null;
var ColorCls = "blue-me"
var New = function(name, attrs, extend){
    var e = document.createElement(name);
    if (extend != null){
        for (var key in extend) {
        　　var item = extend[key];
            e.setAttribute(key,item)
        }
    }
    if (attrs.hasOwnProperty("id")){
        e.id = attrs.id;
    }
    if (attrs.hasOwnProperty("class")){
        e.className = attrs.class;
    }
    if (attrs.hasOwnProperty("text")){
        e.textContent = attrs.text;
    }
    if (attrs.hasOwnProperty("html")){
        e.innerHTML = attrs.html;
    }

    if (attrs.hasOwnProperty("style")){
        e.style = attrs.style;
    }
    if (attrs.hasOwnProperty("value")){
        e.value = attrs.value;
    }
    if (attrs.hasOwnProperty("name")){
        e.name = attrs.name;
    }
    if (attrs.hasOwnProperty("title")){
        e.title = attrs.title;
    }

    if (attrs.hasOwnProperty("timeout")){
        e.timeout = attrs.timeout;
        // alert(e.timeout);
    }

    return e
}


var PluginName = "SyncAll"
var UseIframeSelect = false
var NewHTML = New("iframe",{
    name:PluginName,
    frameborder:"0",
    style:`
    position: fixed;
    top: 0%;
    right: 3%;
    height: 100%;
    `
})
// var new_style_element = New("style",{
//     text:CSSSTYLE
// });

// if (UseIframeSelect){
//     document.getElementsByTagName("html")[0].appendChild(NewHTML);
//     window.frames[PluginName].document.getElementsByTagName("head")[0].appendChild(new_style_element);
// }




notify_stack = New("div",{
    id:"notify-stack",
    style:`
    position:fixed;
right:1%;
top: 3%;
max-width:400px;

    `
})


var InjectHtml = function(e){
    // document.getElementsByTagName("html")[0].appendChild(e);
    // NewHTML.appendChild(e)
    if(UseIframeSelect){
        document.getElementsByName(PluginName)[0].contentDocument.getElementsByTagName("body")[0].appendChild(e);
    }else{
        document.getElementsByTagName("html")[0].appendChild(e);
    }
}
var DeleteInjectHtml = function(e){
    if (UseIframeSelect){
        document.getElementsByName(PluginName)[0].contentDocument.getElementsByTagName("body")[0].removeChild(e);
    }else{
        document.getElementsByTagName("html")[0].removeChild(e);
    }
}


console.log("Use Iframe:",UseIframeSelect);
function Q(nameOrId,iframeName){
    var _select = document;
    
    if (iframeName != null){
        _select = document.getElementsByName(iframeName)[0].contentDocument;
    }else if(UseIframeSelect){
        console.log("use iframe mode");
        _select = document.getElementsByName(PluginName)[0].contentDocument;
        if (_select == null){
            console.log("no iframe found!");    
        }
    }
    if (nameOrId.trim()[0] == "#"){
        let id = nameOrId.trim().substr(1,nameOrId.trim().length)
        // console.log("id:",id,_select.getElementById(id)[0])
        return _select.getElementById(id);
    }else if (nameOrId.trim()[0] == "."){
        let cls = nameOrId.trim().substr(1,nameOrId.trim().length)
        // console.log("id:",id)
        return _select.getElementsByClassName(cls)
    }else{
        return _select.getElementsByTagName(nameOrId)[0]
    }
}

InjectHtml(notify_stack)

var Dismiss = function(ele, callback){
    console.log("this is: ",ele);
    if (ele !=null){
        $(ele).transition({
            // "top":"-10%;",
            // "margin-left":"100px",
            "opacity":0.1,
        },500,function(){
            ele.parentElement.removeChild(ele)
            if (callback != null){
                callback()
            }
        })
    }
    
}

Q("#notify-stack").onclick = function(e){
    var ele = e.target||e.srcElement;
    switch(ele.className){
        case "dissmiss":
            Dismiss(ele.parentElement.parentElement);
            break;
    }
}

var notifymsg = function(res, onlyJsonOrFunction){
    var timeout = -1;
    var notifyframe = New("div",{
        class:"notify",
    })
    // notifyframe.className="notify"
    var title  = "Notify"
    var info = "blue"
    var callback = null
    var onlyJson = false
    if (typeof(onlyJsonOrFunction)== "function"){
        callback = onlyJsonOrFunction;
    }else if (typeof(onlyJsonOrFunction) == "boolean"){
        onlyJson = onlyJsonOrFunction;
    }
    // console.log(typeof(res))
    var notifybody = New("div",{
        class:"notify-body",
    })
    if (res != null && res.hasOwnProperty("timeout")){
        timeout = parseInt(res.timeout);
        if (timeout == NaN){
            timeout = -1;
        }
        // console.log("t:",timeout)     
    }
    if (res != null && res.hasOwnProperty("title")){
        title = res.title
    }
    if (res != null && res.hasOwnProperty("url")){
        res.url = `<sup><a href="${res.url}" >[Url]</a></sup>`
    }
    if (res != null && res.hasOwnProperty("info")){
        
        if(res.info == "ok"){
            info = "green";
        }else if (res.info == "fail"){
            info = "red"
        }else{
            info = "blue"
        }
        console.log(info)        
        
    }

    

    if (onlyJson != null && onlyJson ==true){
        var text = JSON.stringify(res);
        notifybody.innerHTML = `<p class="notify-text">${text}</p>`
        if (timeout == -1){
            timeout = 10000;
        }
        
    }else if (res.outerHTML != null){
        if (res.hasAttribute("title")){
            title = res.getAttribute("title");
        }
        if (res.hasAttribute("timeout") && res.getAttribute("timeout") != null){
            timeout = parseInt(res.getAttribute("timeout"));
            // console.log("T2:",res.getAttribute("wait"));
            // alert(timeout,res.getAttribute("wait"));
            
        }
        notifybody.innerHTML = res.outerHTML
    }else if (res.hasOwnProperty("html")){
        
        notifybody.innerHTML = res.html        
    }
    var left = "∞"
    if (timeout != -1){
        left = timeout / 1000;
        left = left + " s";
    }
    console.log("T1:",timeout)
    var notihead = New("div",{
        
        class:"notify-head",
        html:`
        <span class="color-${info}"></span><span class="notify-title" >${title}</span><span class="notify-info">${left}</span><span class="dissmiss">×</span>
        `
    })

    notifyframe.appendChild(notihead)
    notifyframe.appendChild(notifybody)

    var notifyheight = notify_stack.childElementCount * 62 + 6;
    console.log(notifyheight,res)
    notifyframe.setAttribute("style",notifyframe.getAttribute("style") + `;right:-300px;`)
    Q("#notify-stack").appendChild(notifyframe)
    $(notifyframe).transition({
        right:"1%",
        "margin-top": "10px",
        // "top":notifyheight,
    },500)
    
    if (callback != null){
        callback(notifyframe);
    }
    if (timeout != -1){
        console.log("T:",timeout)
        var timer = setInterval(function(){
            if (left != "∞"){
                var _text = notihead.getElementsByClassName("notify-info")[0].textContent;
                var left = parseInt(_text.substr(0,_text.length -1)) - 1; 
                notihead.getElementsByClassName("notify-info")[0].textContent = `${left} s`
            
            }
        },1000)
        setTimeout(function(){
            $(notifyframe).transition({
                "right":"-30%;",
            },500,function(){
                Q("#notify-stack").removeChild(notifyframe)  
        
            })
            clearInterval(timer);
        },timeout)
        
    }
    
}


var dispose = function(){
    var eleO = Q("#inputsync");
        var ele1 = Q("#inputsync-btn");
    if ($("#inputsync")!= null){
        $("#inputsync").transition({
            right:"-30%"
        },500,function(){
            DeleteInjectHtml(eleO)
        });
        $("#inputsync-btn").transition({
            right:"-30%"
        },500,function(){
            DeleteInjectHtml(ele1)
        });
    }else{
        DeleteInjectHtml(eleO)
        
        DeleteInjectHtml(ele1)
    }
}



