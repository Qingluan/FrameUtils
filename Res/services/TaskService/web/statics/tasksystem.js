/* global bootstrap: false */
(function () {
    'use strict'
    var tooltipTriggerList = [].slice.call(document.querySelectorAll('[data-bs-toggle="tooltip"]'))
    tooltipTriggerList.forEach(function (tooltipTriggerEl) {
      new bootstrap.Tooltip(tooltipTriggerEl)
    })
})()

var websocket = new MyWebSocket("/task/v1/websocket")

function LoadJS(id,content, location){
    if ($("#"+id).length == 0){
        var script = document.createElement('script');
        script.type = 'text/javascript';
        script.text = content;
        script.id = id;
        location.appendChild(script);
    }
}


function api(data,after){
    $.post("/task/v1/api", JSON.stringify(data),function(data){
        console.log(data);
        var data = JSON.parse(data);
        after(data);
    })
}
function setPopupContent(content){
    $("#modal-body-detail")[0].innerHTML = "<iframe id=\"detailFrame\"  width=\"750\" height=\"600\" frameborder=\"no\" border=\"0\"></iframe>";
    $("#detailFrame").contents().find("body").append("<pre>" + content +"</pre>")
    $("#showModalDetail").modal();
}
function showDetail(id){
    $.post("/task/v1/api", JSON.stringify({
        oper:"pull",
        id:id,
    }),function(data){
        console.log(data);
        var data = JSON.parse(data);
        setPopupContent(data.log);
    })
}

function newTask(){ 
    $("#showModalForm").modal();
}
function newFileTask(){ 
    $("#showModal").modal();
}

function newSetting(){ 
    $("#showModalSetting").modal();
}


$("#taskPanel>button").click(function(){
    var v = $("#taskPanel>input").val();
    api({
        oper:"clear",
        id:v
    },function(data){
        if (data.state != "ok"){
            notifymsg(data.log,true)
        }else{
            location.reload();
        }
    })
})



var SERVER_LOGS = {};

function UpdateOneLineLog(log ){
    var tr = $("#task-logs>tbody>#" + log.id);
    // console.log(log);
    if (tr.length > 0){
        console.log("Update:",log.state,log.id);
        tr[0].innerHTML = `
            <td>${log.id}</td>
            <td>${log.deployed_server}</td>
            <td>${log.state}</td>
            <td>${log.log_last}</td>
            <td>${log.log_size}</td>
            <td><a href="#" onclick="return showDetail('${log.id}')" >預覽結果</a></td>
    `
    }else{
        var tr = document.createElement("tr");
        tr.id = log.id
        tr.className = "task-item"
        tr.innerHTML = `
            <td>${log.id}</td>
            <td>${log.deployed_server}</td>
            <td>${log.state}</td>
            <td>${log.log_last}</td>
            <td>${log.log_size}</td>
            <td><a href="#" onclick="return showDetail('${log.id}')" >預覽結果</a></td>
    `
        $("#task-logs>tbody")[0].appendChild(tr);
    }
}

websocket.AddCallback("updatelog",function(data){
    UpdateOneLineLog(data.value)
})


function StateUpdate(){
    $.post("/task/v1/", JSON.stringify(),function(data){
        // console.log(data);
        var data = JSON.parse(data);
        // console.log(data);
        if (data.Logs != null && data.Logs.length > 0){
            for (var no in data.Logs){
                // console.log(log);
                var log = data.Logs[no];
                if (log.id == null){
                    continue
                }
                if (SERVER_LOGS[log.id] == null){
                    /// 添加一行
                    UpdateOneLineLog(log);
                    SERVER_LOGS[log.id] = log
                }else{
                    /// 更新一行
                    var f = SERVER_LOGS[log.id];
                    var if_update = false 
                    if (log.log_last != f.log_last ){
                        if_update = true
                    }
                    if (log.log_size != f.log_size ){
                        if_update = true
                    }
                    if (log.state != f.state ){
                        if_update = true
                    }
                    if (if_update == true){
                        UpdateOneLineLog(log);
                        SERVER_LOGS[log.id] = log
                    }
                    
                }
            }
        }
        /*
    ErrNum: "0"
LogRoot: "/tmp/my-task"
Logs: (4) [{…}, {…}, {…}, {…}]
LogsNum: "6"
ReadyNum: "0"
RunningNum: "1"
Servers: "0"
States: {cmd-66471eb060dd0f7d1ceacbc5c1999456: {…}, http-698062f2e84bd33cb6f9bfd66c1d3de1: {…}, http-9752101d4b0656477917f97d40cfe1ad: {…}, http-c8f2bda12e58ba6c3251b92e9d49ae68: {…}}
TaskNum: "{\"cmd-66471eb060dd0f7d1ceacbc5c1999456\":{\"deployed_server\":\"localhost:5002\",\"args\":[\"sleep 10 \\u0026\\u0026 ls -lha /tmp/my-task\"],\"kargs\":{\"logTo\":\"192.168.1.180:4099\"},\"id\":\"cmd-66471eb060dd0f7d1ceacbc5c1999456\",\"state\":\"Finished\",\"pid\":\"\",\"log_size\":\"\",\"log_last\":\"\"},\"http-698062f2e84bd33cb6f9bfd66c1d3de1\":{\"deployed_server\":\"localhost:5002\",\"args\":[\"http\",\"http://www.yybnet.net/xianyang/binxian/\"],\"kargs\":{\"logTo\":\"192.168.1.180:4099\"},\"id\":\"http-698062f2e84bd33cb6f9bfd66c1d3de1\",\"state\":\"Failed\",\"pid\":\"\",\"log_size\":\"\",\"log_last\":\"\"},\"http-9752101d4b0656477917f97d40cfe1ad\":{\"deployed_server\":\"localhost:5002\",\"args\":[\"http://www.aynews.net.cn/\"],\"kargs\":{\"logTo\":\"192.168.1.180:4099\"},\"id\":\"http-9752101d4b0656477917f97d40cfe1ad\",\"state\":\"Finished\",\"pid\":\"\",\"log_size\":\"\",\"log_last\":\"\"},\"http-c8f2bda12e58ba6c3251b92e9d49ae68\":{\"deployed_server\":\"localhost:5002\",\"args\":[\"https://www.baidu.com\"],\"kargs\":{\"logTo\":\"192.168.1.180:4099\"},\"id\":\"http-c8f2bda12e58ba6c3251b92e9d49ae68\",\"state\":\"Finished\",\"pid\":\"\",\"log_size\":\"\",\"log_last\":\"\"}}"
ip: "192.168.1.180"
        */
        $("#running").text(data.RunningNum);
        $("#proxy").text(data.Proxy);
        // $("#log-root").text(data.LogRoot);
        if (data.Servers == null){
            $("#server-num").text("0");
        }else{
            $("#server-num").text(data.Servers.length);
            $("#others")[0].innerHTML = "<li>" + data.Servers.join("</li><li>")+"</li>"
        }
    })
}



$("#settingSubmit").click(function(){
    var vproxy = $("#settingProxy").val()
    var vothers = $("#settingOthers").val().split("\n")
    console.log(vothers.join(","),vproxy)
    api({
        oper:"config",
        proxy:vproxy,
        others:vothers,
    },data =>{
        data["timeout"] = 10000
        notifymsg(data,true)
        $("#showModalSetting").modal("hide");
        StateUpdate();
    })
})

$("#taskSubmit").click(function(){
    var tp = $("#taskOper").val();
    var input = $("#taskInput").val();
    api({
        oper:"push",
        tp:tp,
        input:input,
    },data =>{
        notifymsg(data,true)
        $("#showModalForm").modal("hide");
    })
})




$(".task-item").click(function(){
    console.log(this.id);
})