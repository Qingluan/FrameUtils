package web

var baseFunctionJS = `

var ws = new WebSocket("ws://" + window.location.host + "/api");
var SendAction;
// ws.onopen = function() {
// 	SendAction("hello","hello","hello world!");
// }

Date.prototype.format = function(format) {
    /*
     * eg:format="yyyy-MM-dd hh:mm:ss";
     */
    var o = {
        "M+" : this.getMonth() + 1, // month
        "d+" : this.getDate(), // day
        "h+" : this.getHours(), // hour
        "m+" : this.getMinutes(), // minute
        "s+" : this.getSeconds(), // second
        "q+" : Math.floor((this.getMonth() + 3) / 3), // quarter
        "S+" : this.getMilliseconds()
        // millisecond
    }

    if (/(y+)/.test(format)) {
        format = format.replace(RegExp.$1, (this.getFullYear() + "").substr(4
        - RegExp.$1.length));
    }

    for (var k in o) {
        if (new RegExp("(" + k + ")").test(format)) {
        var formatStr="";
        for(var i=1;i<=RegExp.$1.length;i++){
            formatStr+="0";
        }

        var replaceStr="";
        if(RegExp.$1.length == 1){
            replaceStr=o[k];
        }else{
            formatStr=formatStr+o[k];
            var index=("" + o[k]).length;
            formatStr=formatStr.substr(index);
            replaceStr=formatStr;
        }
        format = format.replace(RegExp.$1, replaceStr);
        }
    }
    return format;
}


var updateview = function(id, inner, oper){
	var ele ;
	if (id.trim()[0] == "#"){
		id = id.trim().substr(1,id.trim().length)
	}
	if(document.getElementById(id) == null){
		if ($("show-data-area")[0] == null){
			var tag = "<div id=\"show-data-area\"></div>"
			document.getElementsByTagName("body")[0].insertAdjacentHTML("beforeEnd",tag);		
			ele = document.getElementById("show-data-area");
		}else{
			ele = document.getElementById("show-data-area");
		}
	}else{
		ele = document.getElementById(id);
	}

	if (oper == "update"){
		ele.innerHTML = inner;
	}else{
		ele.insertAdjacentHTML("beforeEnd",inner);
	}
}

var showPage = function(ele,value, offsetnum){
	num = parseInt($(ele).parent().parent().find("#data-num").val())
	if (offsetnum != null){
		num += parseInt(offsetnum);
		$(ele).parent().parent().find("#data-num").val(num);
	}
	size = parseInt($(ele).parent().parent().find("#data-size").val())
	console.log(num,size)
	SendAction("db","show-data",JSON.stringify({
		size:size,
		num:num,
		db: value
	}));
}

var actions = {
	AddView: function(data){
		id = data.id
		value = data.value
		if (id == ""){
			$("body").append(value) 
		}else{
			ele = document.getElementById(id)
			ele.innerHTML = ele.innerHTML + value 
		}
		
	},
	Notify:function(data){
		id = data.id
		
		subtitle = new Date().format("yyyy-MM-dd hh:mm:ss");
		content = ""
		if (id == "show-data"){
			
			// TODO
			tp = "show-data"
			id = "db"
			content = '<div class="row g-3"><div class="col-4"><label for="pagesize" class="visually-hidden">Page Size</label><input type="number" class="form-control" id="data-size" value="100"></div><div class="col-3"><label for="page" class="visually-hidden">Page</label><input type="number" class="form-control" id="data-num" placeholder="page num" value="0" ></div><div class="col-2"><button class="btn btn-primary mb-3" style="margin-top:5px" onclick="return showPage(this,\''+ data.value+'\')">Go To</button></div><div class="col-2"><button class="btn btn-primary mb-3" style="margin-top:5px" onclick="return showPage(this,\''+ data.value+'\',1)">Next Page</button></div></div>'
		}else{
			content = value
		}
		$.toast({type: 'info',
			title: 'Notice!',
			subtitle: subtitle,
			content:  content,
		})
	},
	SetView: function(data){
		id = data.id
		value = data.value
		updateview(id, value, "update")
		//ele = document.getElementById(id)
		//ele.innerHTML = value 
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
	Err:function(data){
		alert(data.value);
	},
	UpdateTable:function(data){
		id = data.id
		value = JSON.parse(data.value)
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
		}else{
			console.log("Err:",m.tp);
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
