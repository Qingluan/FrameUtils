package web

var baseFunctionJS = `

var ws = new WebSocket("ws://" + window.location.host + "/api");
var SendAction;
ws.onopen = function() {
	SendAction("hello","hello","hello world!");
}
var actions = {
	AddView: function(data){
		id = data.id
		value = data.value
		ele = document.getElementById(id)
		ele.innerHTML = ele.innerHTML + value 
	},
	SetView: function(data){
		id = data.id
		value = data.value
		ele = document.getElementById(id)
		ele.innerHTML = value 
	},
	SetAttr: function(data){
		id = data.id
		value = data.value
		ele = document.getElementById(id)
		kv = value.split("=")
		ele.setAttribute(kv[0],kv[1])  
	},
	AddAction: function(data){
		id = data.id
		value = data.value
		ele = document.getElementById("layout-body")
		var newScript = document.createElement('script');
		newScript.type = 'text/javascript';
		newScript.innerHTML = value;
		ele.appendChild(newScript);
	},
	GetAttr: function(data){
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
			SendAction(data.backid, "GetAttr",v)
		}else{
			SendAction(id, "GetAttr",v)
		}
	},
	GetContent: function(data){
		id = data.id
		value = data.value
		ele = document.getElementById(id)
		v = ele.textContent
		if (data.hasOwnProperty("backid") == true){
			SendAction(data.backid, "GetContent",v)
		}else{
			SendAction(id, "GetContent",v)
		}
	},
	GetHtml: function(data){
		id = data.id
		value = data.value
		ele = document.getElementById(id)
		v = ele.innerHTML
		if (data.hasOwnProperty("backid") == true){
			SendAction(data.backid, "GetHtml",v)
		}else{
			SendAction(id, "GetHtml",v)
		}
	},
	OnDo:function(data) {
		console.log(data.value);
		eval(data.value);
	},
	UpdateTable:function(data){
		id = data.id
		value = data.value
		for ( i = 0; i < value.length; i ++){
			locTd = $("tr[data-row="+ value[i].row +"]>td[data-col=" + value[i].col + "]")
			locTd.text(value[i].data)
			locTd.attr("data",value[i].data)
		}
	}
}

window.addEventListener("load", function(evt) {
	ws.onmessage = function(event) {
		var m = JSON.parse(event.data);
		console.debug("Received message", m.id, m.tp, m.value);
		callback = actions[m.tp];
		if (callback != null){
			callback(m);
		}
	}
	ws.onerror = function(event) {
		console.debug(event)
	}
	
})

SendAction = function(id, tp , value){
	ws.send(JSON.stringify({
		id:id,
		tp:tp,
		value:value
	}))
}


`
