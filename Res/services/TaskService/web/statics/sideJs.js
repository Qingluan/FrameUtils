/* global bootstrap: false */
(function () {
    'use strict'
    var tooltipTriggerList = [].slice.call(document.querySelectorAll('[data-bs-toggle="tooltip"]'))
    tooltipTriggerList.forEach(function (tooltipTriggerEl) {
      new bootstrap.Tooltip(tooltipTriggerEl)
    })
  })()
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