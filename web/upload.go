package web

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/Qingluan/FrameUtils/asset"
	"github.com/Qingluan/FrameUtils/utils"
)

type WebUpload struct {
	Action  string
	Session map[string]string
}

var (
	UPLOAD_TEMPLATE, _ = asset.AssetAsFile("Res/templates/upload.html")
)

func NewWebUpload(action string) (u *WebUpload) {
	return &WebUpload{
		Action: action,
	}
}

func (u *WebUpload) Parse(name string) string {
	buffer := bytes.NewBuffer([]byte{})
	t, err := template.New(name).ParseFiles(UPLOAD_TEMPLATE)
	if err != nil {
		log.Println("parse upload err :", err)
	}
	t.Execute(buffer, u)
	return buffer.String()
}

func (u *WebUpload) BuildUploadFunc(call func(id, filePath string)) {
	uploadFile := func(w http.ResponseWriter, r *http.Request) {
		r.ParseMultipartForm(1024 << 20)
		file, handler, err := r.FormFile("uploadFile")
		if err != nil {
			log.Println("Error Retrieving the File", "|", err, "|")
			return
		}
		defer file.Close()
		tmpDir := os.TempDir()
		tempFile := filepath.Join(tmpDir, "Res", "templates", "statics", handler.Filename)
		log.Println("upload :", handler.Filename)
		sessId := utils.NewSessionID()
		u.Session[sessId] = tempFile
		// w.WriteHeader(201)
		cookie := http.Cookie{Name: "session-id", Value: sessId, Path: "/", MaxAge: -1}
		http.SetCookie(w, &cookie)
		// w.Header().Set("session-id", sessId)

		if _, err := os.Stat(tempFile); err != nil {
			if err != nil {
				fmt.Println(err)
			}
			// defer tempFile.Close()
			fp, err := os.Create(tempFile)
			if err != nil {
				return
			}
			io.Copy(fp, file)
		}
		call(sessId, tempFile)
		w.Write([]byte("{\"msg\":\"upload ok\"}"))
	}
	http.HandleFunc(u.Action, uploadFile)
}
