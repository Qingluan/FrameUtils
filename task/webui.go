package task

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/Qingluan/FrameUtils/asset"
	"github.com/Qingluan/FrameUtils/utils"
	"github.com/Qingluan/FrameUtils/web"
)

var (
	TEMP = `

	<!doctype html>
	<html lang="en">
	  <head>
		<meta charset="utf-8">
		<meta name="viewport" content="width=device-width, initial-scale=1">
		<meta name="description" content="a task dist system">
		<meta name="author" content="Qingluan">
		<meta name="generator" content="task-service">
		<title>🌏task-services</title>
		
		<style name="sidebars">
		  {{ .SideCss }}
		</style>
		
	<!-- 
		<link rel="canonical" href="https://getbootstrap.com/docs/5.0/examples/sidebars/">
	
		 -->
	
		<!-- Bootstrap core CSS -->
	  <style name="bootstrap" rel="stylesheet">
		{{ .BCss }}
	  </style>
		<style name="upload.css" rel="stylesheet">{{ .UploadCSS }}</style>
		<!-- Favicons -->
	<!-- <link rel="apple-touch-icon" href="/docs/5.0/assets/img/favicons/apple-touch-icon.png" sizes="180x180">
	<link rel="icon" href="/docs/5.0/assets/img/favicons/favicon-32x32.png" sizes="32x32" type="image/png">
	<link rel="icon" href="/docs/5.0/assets/img/favicons/favicon-16x16.png" sizes="16x16" type="image/png">
	<link rel="manifest" href="/docs/5.0/assets/img/favicons/manifest.json">
	<link rel="mask-icon" href="/docs/5.0/assets/img/favicons/safari-pinned-tab.svg" color="#7952b3">
	<link rel="icon" href="/docs/5.0/assets/img/favicons/favicon.ico"> -->
	<meta name="theme-color" content="#7952b3">
		<style>
		  .bd-placeholder-img {
			font-size: 1.125rem;
			text-anchor: middle;
			-webkit-user-select: none;
			-moz-user-select: none;
			user-select: none;
		  }
	
		  @media (min-width: 768px) {
			.bd-placeholder-img-lg {
			  font-size: 3.5rem;
			}
		  }
		  
		  #showModalDetail{
			  min-width:900px;
		  }
		  
		  #showModalDetail>div{
			min-width:900px;
		  }
		
		</style>
	
		
		<!-- Custom styles for this template -->
	</head>
	  <body>
	  {{ .Svgs }}
	
	<div class="d-flex flex-column p-3 text-white bg-dark" style="width: 280px;">
	  <a href="/task/v1/ui" class="d-flex align-items-center mb-3 mb-md-0 me-md-auto text-white text-decoration-none">
		<svg class="bi me-2" width="40" height="32"><use xlink:href="#bootstrap"/></svg>
		<span class="fs-4">任务系统</span>
	  </a>
	  <hr>
	  <ul class="nav nav-pills flex-column mb-auto">
		<li class="nav-item">
		  <a href="#" class="nav-link active">
			<svg class="bi me-2" width="16" height="16"><use xlink:href="#home"/></svg>
			服务器IP: {{ .ServerIP}}
		  </a>
		</li>
		<li>
		  <a href="#" class="nav-link text-white">
			<svg class="bi me-2" width="16" height="16"><use xlink:href="#speedometer2"/></svg>
			正在进行: {{ .RunningNum}}
		  </a>
		</li>
		<li>
		  <a href="#" class="nav-link text-white">
			<svg class="bi me-2" width="16" height="16"><use xlink:href="#table"/></svg>
			等待: {{ .ReadyNum }}
		  </a>
		</li>
		<li>
		  <a href="#" class="nav-link text-white">
			<svg class="bi me-2" width="16" height="16"><use xlink:href="#grid"/></svg>
			日志数: {{.LogsNum}}
		  </a>
		</li>
		<li>
		  <a href="#" class="nav-link text-white">
			<svg class="bi me-2" width="16" height="16"><use xlink:href="#people-circle"/></svg>
			服务器数 : {{ .ServerNum }}
		  </a>
		</li>
		<li>
		  <a href="#" onclick="return newTask();" style="margin-top:100px" class="nav-link text-white">
			<svg class="bi me-2" width="16" height="16"><use xlink:href="#cpu-fill"/></svg>
			新任务
		  </a>
		</li>
		
	  </ul>
	  <hr>
	  <div class="dropdown">
		<a href="#" class="d-flex align-items-center text-white text-decoration-none dropdown-toggle" id="dropdownUser1" data-bs-toggle="dropdown" aria-expanded="false">
		  <strong>mdo</strong>
		</a>
		<ul class="dropdown-menu dropdown-menu-dark text-small shadow" aria-labelledby="dropdownUser1">
		  <li><a class="dropdown-item" href="#">New project...</a></li>
		  <li><a class="dropdown-item" href="#">Settings</a></li>
		  <li><a class="dropdown-item" href="#">Profile</a></li>
		  <li><hr class="dropdown-divider"></li>
		  <li><a class="dropdown-item" href="#">Sign out</a></li>
		</ul>
	  </div>
	</div>
	
	<main class="col-md-9 ms-sm-auto col-lg-10 px-md-4">
	  <div class="d-flex justify-content-between flex-wrap flex-md-nowrap align-items-center pt-3 pb-2 mb-3 border-bottom">
		<h1 class="h2">Dashboard</h1>
		<div class="btn-toolbar mb-2 mb-md-0">
		  <div class="btn-group me-2">
			<button type="button" class="btn btn-sm btn-outline-secondary">Share</button>
			<button type="button" class="btn btn-sm btn-outline-secondary">Export</button>
		  </div>
		  <button type="button" class="btn btn-sm btn-outline-secondary dropdown-toggle">
			<span data-feather="calendar"></span>
			This week
		  </button>
		</div>
	  </div>
	
	
	  <h2><em>Log Root: {{ .LogRoot}}</em> </h2>
	  <div class="table-responsive">
		<h2>任务日志</h2>
	  
		<table class="table table-striped table-sm">
		  <thead>
			<tr>
			  <th>Task ID</th>
			  <th>最后更改时间</th>
			  <th>文件大小</th>
			  <th>查看</th>
			</tr>
		  </thead>
		  <tbody>
		  	{{ range .Logs }}
			<tr>
			  <td>{{ .ID }}</td>
			  <td>{{ .ModiTime }}</td>
			  <td>{{ .Size }}</td>
			  <td><a href="#" onclick="return showDetail('{{ .ID }}')" >click</a></td>
			</tr>
			{{ end }}
		  </tbody>
		</table>
	  </div>
	  <canvas class="my-4 w-100" id="myChart" width="900" height="380"></canvas>
	<div class="modal fade" id="showModal" tabindex="-1" role="dialog" aria-labelledby="showModalTitle" aria-hidden="true">
	  <div class="modal-dialog" role="document">
		<div class="modal-content">
		  <div class="modal-header">
			<h5 class="modal-title" id="showModalTitle">Modal title</h5>
			<button type="button" class="close" data-dismiss="modal" aria-label="Close">
			  <span aria-hidden="true">&times;</span>
			</button>
		  </div>
		  <div id="modal-body" style="margin-left:10px">
			{{.UploadHTML}}
		  </div>
		  <div class="modal-footer">
			<button type="button" class="btn btn-secondary" data-dismiss="modal">Close</button>
			<button type="button" class="btn btn-primary">Save changes</button>
		  </div>
		</div>
	  </div>
	</div>
	<div class="modal  fade" id="showModalDetail" tabindex="-1" role="dialog" aria-labelledby="showModalTitle" aria-hidden="true">
	  <div class="modal-dialog" role="document">
		<div class="modal-content">
		  <div class="modal-header">
			<h5 class="modal-title" id="showModalTitle">Modal title</h5>
			<button type="button" class="close" data-dismiss="modal" aria-label="Close">
			  <span aria-hidden="true">&times;</span>
			</button>
		  </div>
		  <div id="modal-body-detail" style="margin-left:10px">
		  </div>
		  <div class="modal-footer">
			<button type="button" class="btn btn-secondary" data-dismiss="modal">Close</button>
			<button type="button" class="btn btn-primary">Save changes</button>
		  </div>
		</div>
	  </div>
	</div>
	</main>
	
	<script name="bjquery" >{{ .BJQuery }}</script>
	<script name="jquery" >{{ .JQuery }}</script>
	
	<script name="bootstrap-js">{{ .BJs }}</script>
	
	<script name="upload.js">{{ .UploadJS }}</script>
	<script name="sidebars.js">{{ .SideJs }}</script>
	
	</body>
</html>
	
	`
)

type TEMPStruct struct {
	Svgs            string
	SideJs          string
	SideCss         string
	UploadJS        string
	UploadCSS       string
	UploadHTML      string
	BJQuery         string
	BJs             string
	BCss            string
	ServerNum       string
	ReadyNum        string
	RunningNum      string
	TaskNum         string
	LogsNum         string
	FailNum         string
	ErrNum          string
	ServerIP        string
	JQuery          string
	LogRoot         string
	TaskPanel       string
	TaskCreateHTML  string
	TaskSettingHTML string
	Logs            []LogUI
}
type LogUI struct {
	ID       string
	ModiTime string
	Size     string
}

func (self *TaskConfig) BuildWebInitialization() (err error) {
	// asset.AssetAsFile("Res/services/TaskService/web")
	// root, err := asset.AssetDir("Res/services/TaskService/web")
	if err != nil {
		return
	}
	temp := os.TempDir()
	dir := filepath.Dir("Res/services/TaskService/web/statics/toast.js")
	rootDir := filepath.Join(temp, dir)
	for _, file := range asset.AssetNames() {
		if strings.Contains(file, "web/") {
			if e, err := asset.AssetAsFile(file); err != nil {
				log.Println("Release:", utils.Red(file, "->", e, err))
			} else {
				log.Println("Release:", utils.Green(file, " => ", e))
			}
		}

	}
	fmt.Println(utils.Green("Statics in :", rootDir))
	http.Handle("/statics/", http.StripPrefix("/statics/", http.FileServer(http.Dir(rootDir))))
	return nil

}

func (self *TaskConfig) GetAsset(name string) string {
	f := ""
	if strings.HasSuffix(name, ".html") {
		f = "Res/services/TaskService/web/" + name
	} else {
		f = "Res/services/TaskService/web/statics/" + name
	}
	if e, err := asset.Asset(f); err == nil {
		return string(e)
	}
	log.Fatal("Not found " + name)
	return "{{ Not Found !! " + name + " }}"
}

func (self *TaskConfig) SimeplUI(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		onePage := TEMPStruct{
			UploadHTML:      self.GetAsset("upload.html"),
			TaskPanel:       web.NewSearchUI("taskPanel", "onclick=\" return taskClear()\"", "清理任务").String(),
			TaskCreateHTML:  self.GetAsset("singleTask.html"),
			TaskSettingHTML: self.GetAsset("settingTask.html"),
		}
		t1, _ := template.New("base").Parse(self.GetAsset("index.html"))
		log := self.GetMyState()
		onePage.UploadHTML = fmt.Sprintf(onePage.UploadHTML, "/task/v1/taskfile")
		onePage.ServerIP = utils.GetLocalIP()
		TaskNum, _ := log["task"]
		onePage.ServerNum = fmt.Sprintf("%d", len(self.Others))
		ReadyNum, _ := log["wait"]
		RunningNum, _ := log["running"]
		LogNum, _ := log["lognum"]
		ErrNum, _ := log["errnum"]
		if fs, err := ioutil.ReadDir(self.LogPath()); err == nil {
			paths := []string{}
			for _, f := range fs {
				onePage.Logs = append(onePage.Logs, LogUI{
					ID:       f.Name(),
					ModiTime: f.ModTime().Local().String(),
					Size:     fmt.Sprintf("%fMB", float64(f.Size())/float64(1024*1024)),
				})
				paths = append(paths, f.Name())
			}
		}

		onePage.LogRoot = self.LogPath()
		onePage.TaskNum = TaskNum.(string)
		onePage.ReadyNum = ReadyNum.(string)
		onePage.RunningNum = RunningNum.(string)
		onePage.LogsNum = LogNum.(string)
		onePage.ErrNum = ErrNum.(string)

		t1.Execute(w, onePage)

	}
}
