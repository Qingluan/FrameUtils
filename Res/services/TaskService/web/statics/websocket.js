class MyWebSocket {
    constructor(uri){
        var ws = null
        try {
            ws = new WebSocket("wss://" + window.location.host + uri);
        } catch (error) {
            try{
                ws = new WebSocket("ws://" + window.location.host + uri);
            }catch (error){
                
            } 
        }
        this.ws = ws;
        var sockH = this;
        this.ws.onopen = function() {
            sockH.SendAction("hello","hello","connecting");
        }
        ws.onmessage = function(event) {
            var m = JSON.parse(event.data);
            console.log(m);
            console.debug("Received message", m.id, m.tp, m.value);
            var callback = sockH.actions[m.tp];
            
            if (callback != null){
                callback(m);
            }else if (Object.getPrototypeOf(sockH)[m.tp] != null && typeof(Object.getPrototypeOf(sockH)[m.tp]) == "function"){
                callback = Object.getPrototypeOf(sockH)[m.tp];
                callback(m);
            }
        }
        ws.onerror = function(event) {
            console.debug(event)
        }
        this.ws = ws
        
        this.actions = {
            
        }

    
    }
    SendAction(id, tp , value){
        this.ws.send(JSON.stringify({
            id:id,
            tp:tp,
            value:value
        }))
    }
    hello(data){
        console.log("Connected !")
    }
    
    AddCallback(tp, func){
        if (typeof(func) == "function"){
            this.actions[tp] = func
        }
    }

    
    
    AddView(data){
        id = data.id
        value = data.value
        ele = document.getElementById(id)
        ele.innerHTML = ele.innerHTML + value 
    }
    SetView(data){
        id = data.id
        value = data.value
        ele = document.getElementById(id)
        ele.innerHTML = value 
    }
    SetAttr(data){
        id = data.id
        value = data.value
        ele = document.getElementById(id)
        kv = value.split("=")
        ele.setAttribute(kv[0],kv[1])  
    }
    AddAction(data){
        id = data.id
        value = data.value
        ele = document.getElementById("layout-body")
        var newScript = document.createElement('script');
        newScript.type = 'text/javascript';
        newScript.innerHTML = value;
        ele.appendChild(newScript);
    }
    GetAttr(data){
        id = data.id
        value = data.value
        ele = document.getElementById(id)
        kv = value.split("=")
        var v;
        if (value == "value"){
            v = ele.value;
        }else{
            v = ele.getAttribute(value);
        }
        if (data.hasOwnProperty("backid") == true){
            this.SendAction(data.backid, "GetAttr",v)
        }else{
            this.SendAction(id, "GetAttr",v)
        }
    }
    GetContent(data){
        id = data.id
        value = data.value
        ele = document.getElementById(id)
        v = ele.textContent
        if (data.hasOwnProperty("backid") == true){
            this.SendAction(data.backid, "GetContent",v)
        }else{
            this.SendAction(id, "GetContent",v)
        }
    }
    GetHtml(data){
        id = data.id
        value = data.value
        ele = document.getElementById(id)
        v = ele.innerHTML
        if (data.hasOwnProperty("backid") == true){
            this.SendAction(data.backid, "GetHtml",v)
        }else{
            this.SendAction(id, "GetHtml",v)
        }
    }
    OnDo(data){
        console.log(data.value);
        eval(data.value);
    }
}
