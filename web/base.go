package web

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/Qingluan/FrameUtils/asset"
	"github.com/Qingluan/FrameUtils/utils"
)

var (
	RandomLoginSession = GenSession()
)

type TData map[string]interface{}

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
			if data["password"] == RandomLoginSession || data["password"] == "?" {
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
				if data["password"] == RandomLoginSession || data["password"] == "?" {
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
	b, _ := asset.Asset(WEBROOT + "login.html")
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

func BuildStatic() (err error) {
	temp := os.TempDir()
	rootDir := filepath.Join(temp, "Res/templates/statics")
	for _, file := range asset.AssetNames() {
		if e, err := asset.AssetAsFile(file); err != nil {
			log.Printf("Release: %s \n", utils.Red(file, "->", e, err))
		} else {
			fmt.Printf("Release: %s \r", utils.Green(file))
		}

	}
	fmt.Println(utils.Green("Statics in :", rootDir))
	http.Handle("/statics/", http.StripPrefix("/statics/", http.FileServer(http.Dir(rootDir))))
	return err
}

func BuildAuth(uri string) {
	http.HandleFunc(filepath.Join(uri, "login"), WebAuthLogin)
	http.HandleFunc(filepath.Join(uri, "logout"), WebAuthLogout)
}

type WebTemp struct {
	t *template.Template
}

func FromBase() (tlp *WebTemp) {
	tlp = new(WebTemp)
	tlp.t, _ = template.ParseFiles(BASE_INDEX_TEMPLATE)
	return
}

func (tpl *WebTemp) AddExistsBlock(name string) *WebTemp {
	tpl.t.ParseFiles(name)
	return tpl
}

func (tpl *WebTemp) NewBlockf(name string, format string, data interface{}) *WebTemp {
	tn := template.New(name)
	var t *template.Template
	if f, err := os.Stat(format); err == nil && !f.IsDir() {
		t, err = tn.ParseFiles(format)
		if err != nil {
			log.Println("New Block printf err:", err)
			return tpl
		}
		buffer := bytes.NewBuffer([]byte{})
		switch data.(type) {
		case string:
			t.Execute(buffer, template.HTML(data.(string)))
		default:
			t.Execute(buffer, data)
		}
		tpl.NewBlock(name, buffer.String())

	} else {
		buffer := bytes.NewBuffer([]byte{})

		t, _ = tn.Parse(format)
		switch data.(type) {
		case string:
			t.Execute(buffer, template.HTML(data.(string)))
		default:
			t.Execute(buffer, data)
		}
		// fmt.Println("res:", buffer.String())
		tpl.NewBlock(name, buffer.String())
	}
	return tpl
}

func (tpl *WebTemp) NewBlock(name string, html string) *WebTemp {
	tpl.t.New(name).Parse(html)
	return tpl
}

func (tpl *WebTemp) InheritTo(name string, htmlPath string) *WebTemp {
	tpl.t.New(name).ParseFiles(htmlPath)
	return tpl
}

func (tpl *WebTemp) Execute(w io.Writer, name string, data interface{}) *WebTemp {
	tpl.t.ExecuteTemplate(w, name, data)
	return tpl
}

func (tpl *WebTemp) ExecuteFinish(w io.Writer) *WebTemp {
	tpl.t.ExecuteTemplate(w, "base", "")
	return tpl
}
