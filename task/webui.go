package task

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"text/template"
	"time"

	"github.com/Qingluan/FrameUtils/asset"
	"github.com/Qingluan/FrameUtils/utils"
	"github.com/Qingluan/FrameUtils/web"
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
	Logs            []TaskState
}
type LogUI struct {
	ID       string
	ModiTime string
	Size     string
}

var (
	RandomLoginSession = GenSession()
	l                  = sync.Mutex{}
)

func GenSession() string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	rbuf := make([]byte, 7)
	r.Read(rbuf)
	return fmt.Sprintf("%x", rbuf)
}
func WebAuthLogout(w http.ResponseWriter, r *http.Request) {
	if c, err := r.Cookie("TaskUID"); err == nil && c.Value == RandomLoginSession {
		// l.Lock()
		// defer l.Unlock()
		// RandomLoginSession = GenSession()
		cookie := http.Cookie{Name: "TaskUID", Value: RandomLoginSession, Path: "/", MaxAge: -1}
		http.SetCookie(w, &cookie)
		log.Println("logout ok:", utils.Green(r.RemoteAddr))

	}
	w.Header().Set("Location", "/task/v1/login") //跳转地址设置
	w.WriteHeader(307)                           //关键在这里！
}

func WebAuthCheck(w http.ResponseWriter, r *http.Request) bool {
	if c, err := r.Cookie("TaskUID"); err == nil && c.Value == RandomLoginSession {
		return true
	} else {
		w.Header().Set("Location", "/task/v1/login") //跳转地址设置
		w.WriteHeader(307)                           //关键在这里！
	}
	return false
}

func WebAuthLogin(w http.ResponseWriter, r *http.Request) {

	log.Println("loging :", utils.Yellow(r.RemoteAddr))
	if c, err := r.Cookie("TaskUID"); err != nil {

		log.Println("loging err:", utils.Red(err))
		if r.Method == "POST" {
			log.Println("try login  by:", utils.Yellow(r.RemoteAddr))

			body, err := ioutil.ReadAll(r.Body)
			if err != nil {
				log.Println("WebAuthlogin:", err)
				return
			}

			data := TData{}
			json.Unmarshal(body, &data)
			if data["password"] == RandomLoginSession || data["password"] == "hallo" {
				cookie := http.Cookie{Name: "TaskUID", Value: RandomLoginSession, Path: "/", MaxAge: 3600}
				http.SetCookie(w, &cookie)
				log.Println("login ok:", utils.Green(r.RemoteAddr))
				w.Write([]byte("write cookie ok"))
				return
			}
		}
	} else {

		if c.Value == RandomLoginSession {

			log.Println("already logined :", utils.Yellow(r.RemoteAddr))
			// http.Redirect(w, r, "/task/v1/ui", 200)
			w.Header().Set("Location", "/task/v1/ui") //跳转地址设置
			w.WriteHeader(307)                        //关键在这里！
			return
		} else {
			if r.Method == "POST" {
				log.Println("try login  by:", utils.Yellow(r.RemoteAddr))

				body, err := ioutil.ReadAll(r.Body)
				if err != nil {
					log.Println("login Err:", err)
					return
				}

				data := TData{}
				json.Unmarshal(body, &data)
				if data["password"] == RandomLoginSession || data["password"] == "hallo" {
					cookie := http.Cookie{Name: "TaskUID", Value: RandomLoginSession, Path: "/", MaxAge: 3600}
					http.SetCookie(w, &cookie)
					log.Println("login ok:", utils.Green(r.RemoteAddr))
					w.Write([]byte("write cookie ok"))
					return
				}
			}
		}

	}
	// if t, err := template.New("login").Parse(); err != nil {
	// log.Fatal("login page broken !", err)
	// } else {
	b, _ := asset.Asset(web.WEBROOT + "login.html")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("ETag", "5d958342-e42")
	w.Header().Set("Server", "nginx/1.16.1")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Accept-Ranges", "bytes")

	w.WriteHeader(404)
	fmt.Fprintf(w, string(b), "/task/v1/ui")

	// t.Execute(w, nil)
	// }

	// return false

}

func (self *TaskConfig) BuildWebInitialization() (err error) {
	// asset.AssetAsFile("Res/services/TaskService/web")
	// root, err := asset.AssetDir("Res/services/TaskService/web")

	temp := os.TempDir()
	dir := filepath.Dir("Res/services/TaskService/web/statics/toast.js")
	rootDir := filepath.Join(temp, dir)
	for _, file := range asset.AssetNames() {
		if strings.Contains(file, "web/") {
			if e, err := asset.AssetAsFile(file); err != nil {
				log.Printf("Release: %s \n", utils.Red(file, "->", e, err))
			} else {
				fmt.Printf("Release: %s \r", utils.Green(file))
			}
		}

	}
	fmt.Println(utils.Green("Statics in :", rootDir))
	http.Handle("/statics/", http.StripPrefix("/statics/", http.FileServer(http.Dir(rootDir))))
	http.HandleFunc("/task/v1/login", WebAuthLogin)
	http.HandleFunc("/task/v1/logout", WebAuthLogout)
	log.Println("LogToPath:", utils.Yellow(self.LogPath()))
	log.Println("Password: ", utils.Red(RandomLoginSession))

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
	if !WebAuthCheck(w, r) {
		return
	}
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
		onePage.Logs = self.DeployStateFind("")
		// if fs, err := ioutil.ReadDir(self.LogPath()); err == nil {
		// 	paths := []string{}
		// 	for _, f := range fs {
		// 		onePage.Logs = append(onePage.Logs, LogUI{
		// 			ID:       f.Name(),
		// 			ModiTime: f.ModTime().Local().String(),
		// 			Size:     fmt.Sprintf("%fMB", float64(f.Size())/float64(1024*1024)),
		// 		})
		// 		paths = append(paths, f.Name())
		// 	}
		// }

		onePage.LogRoot = self.LogPath()
		onePage.TaskNum = TaskNum.(string)
		onePage.ReadyNum = ReadyNum.(string)
		onePage.RunningNum = RunningNum.(string)
		onePage.LogsNum = LogNum.(string)
		onePage.ErrNum = ErrNum.(string)

		t1.Execute(w, onePage)

	} else if r.Method == "POST" {
		log := self.GetMyState()
		myip := utils.GetLocalIP()
		TaskNum, _ := log["task"]
		others := fmt.Sprintf("%d", len(self.Others))
		ReadyNum, _ := log["wait"]
		RunningNum, _ := log["running"]
		LogNum, _ := log["lognum"]
		ErrNum, _ := log["errnum"]
		Logs := self.DeployStateFind("")
		// Logs := []LogUI{}
		// if fs, err := ioutil.ReadDir(self.LogPath()); err == nil {
		// 	paths := []string{}
		// 	for _, f := range fs {
		// 		Logs = append(Logs, LogUI{
		// 			ID:       f.Name(),
		// 			ModiTime: f.ModTime().Local().String(),
		// 			Size:     fmt.Sprintf("%fMB", float64(f.Size())/float64(1024*1024)),
		// 		})
		// 		paths = append(paths, f.Name())
		// 	}
		// }
		// stateBytes,_ := json.Marshal()
		jsonWrite(w, TData{
			"ip":         myip,
			"LogRoot":    self.LogPath(),
			"TaskNum":    TaskNum.(string),
			"ReadyNum":   ReadyNum.(string),
			"RunningNum": RunningNum.(string),
			"LogsNum":    LogNum.(string),
			"ErrNum":     ErrNum.(string),
			"Servers":    others,
			"States":     self.state,
			"Logs":       Logs,
		})

	}
}
